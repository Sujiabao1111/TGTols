package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models/dtos"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func configuredStarsPerUSD() int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(os.Getenv("TELEGRAM_STARS_PER_USD")), 10, 64)
	if err == nil && value > 0 {
		return value
	}
	if cfg := telegramConfig(); cfg != nil && cfg.StarsPerUSD > 0 {
		return cfg.StarsPerUSD
	}
	return 50
}

type TelegramStarsUpdate struct {
	UpdateID         int64 `json:"update_id"`
	PreCheckoutQuery *struct {
		ID   string `json:"id"`
		From struct {
			ID int64 `json:"id"`
		} `json:"from"`
		Currency       string `json:"currency"`
		TotalAmount    int64  `json:"total_amount"`
		InvoicePayload string `json:"invoice_payload"`
	} `json:"pre_checkout_query,omitempty"`
	Message *struct {
		From struct {
			ID int64 `json:"id"`
		} `json:"from"`
		SuccessfulPayment *struct {
			Currency                string `json:"currency"`
			TotalAmount             int64  `json:"total_amount"`
			InvoicePayload          string `json:"invoice_payload"`
			TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
			ProviderPaymentChargeID string `json:"provider_payment_charge_id"`
		} `json:"successful_payment,omitempty"`
	} `json:"message,omitempty"`
}

func telegramConfig() *helpers.TelegramConfig {
	holder := helpers.GetCfgInstance()
	if holder == nil || holder.Conf == nil {
		return nil
	}
	return &holder.Conf.Payment.Telegram
}

func telegramBotToken() string {
	if value := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")); value != "" {
		return value
	}
	if cfg := telegramConfig(); cfg != nil {
		return strings.TrimSpace(cfg.BotToken)
	}
	return ""
}

func TelegramWebhookSecret() string {
	if value := strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_SECRET")); value != "" {
		return value
	}
	if cfg := telegramConfig(); cfg != nil {
		return strings.TrimSpace(cfg.WebhookSecret)
	}
	return ""
}

func telegramStarsEnabled() bool {
	if value := strings.TrimSpace(os.Getenv("TELEGRAM_STARS_ENABLED")); value != "" {
		return !strings.EqualFold(value, "false")
	}
	if cfg := telegramConfig(); cfg != nil {
		return cfg.Enabled
	}
	return false
}

func (s *PaymentService) CreateTelegramStarsOrder(ctx context.Context, userID uint64, req dtos.CreateTelegramStarsRequest) (*dtos.CreateTelegramStarsResponse, error) {
	if telegramBotToken() == "" {
		return nil, errors.New("Telegram Stars is not configured")
	}
	if !telegramStarsEnabled() {
		return nil, errors.New("Telegram Stars is disabled")
	}
	if req.Amount < 1 || req.Amount > 100000 || req.Amount != float64(int64(req.Amount)) {
		return nil, errors.New("amount must be a whole number between 1 and 100000 U")
	}
	telegramUser, _, err := ValidateTelegramInitData(req.InitData)
	if err != nil {
		return nil, err
	}
	telegramUserID := strconv.FormatInt(telegramUser.ID, 10)
	starsPerUSD := configuredStarsPerUSD()
	stars := int64(req.Amount) * starsPerUSD
	orderID := common.GenerateOrderID(userID)
	payload, err := securePayload(orderID)
	if err != nil {
		return nil, err
	}
	rate, _ := json.Marshal(map[string]interface{}{"stars_per_usd": starsPerUSD, "credited_usd": req.Amount})
	order := dtos.PaymentOrder{UserID: userID, OrderID: orderID, Amount: req.Amount, Cost: req.Amount, Type: "telegram_stars", DstCode: "TG_STARS", Status: 0, ProductInfo: "Telegram Stars recharge", TelegramUserID: telegramUserID, StarsAmount: stars, InvoicePayload: payload, RateSnapshot: string(rate)}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, fmt.Errorf("save Stars order: %w", err)
	}
	invoiceURL, err := telegramCreateInvoiceLink(ctx, payload, stars, orderID)
	if err != nil {
		s.db.Model(&order).Updates(map[string]interface{}{"status": 2, "ref_msg": err.Error()})
		return nil, err
	}
	if err := s.db.Model(&order).Update("pay_url", invoiceURL).Error; err != nil {
		return nil, err
	}
	return &dtos.CreateTelegramStarsResponse{OrderID: orderID, InvoiceURL: invoiceURL, Amount: req.Amount, StarsAmount: stars, Status: 0}, nil
}

type TelegramWebAppUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoURL  string `json:"photo_url"`
}

func ValidateTelegramInitData(initData string) (*TelegramWebAppUser, url.Values, error) {
	botToken := telegramBotToken()
	if botToken == "" {
		return nil, nil, errors.New("Telegram login is not configured")
	}
	values, err := url.ParseQuery(initData)
	if err != nil || values.Get("hash") == "" {
		return nil, nil, errors.New("open this page from the Telegram Mini App")
	}
	providedHash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, nil, errors.New("invalid Telegram init data")
	}
	values.Del("hash")
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+"="+values.Get(k))
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	signature := hmac.New(sha256.New, secret.Sum(nil))
	signature.Write([]byte(strings.Join(lines, "\n")))
	if !hmac.Equal(providedHash, signature.Sum(nil)) {
		return nil, nil, errors.New("invalid Telegram init data signature")
	}
	authDate, _ := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if authDate == 0 || time.Since(time.Unix(authDate, 0)) > 24*time.Hour {
		return nil, nil, errors.New("Telegram session expired")
	}
	var user TelegramWebAppUser
	if json.Unmarshal([]byte(values.Get("user")), &user) != nil || user.ID == 0 {
		return nil, nil, errors.New("Telegram user is missing")
	}
	return &user, values, nil
}

func securePayload(orderID string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return orderID + "_" + hex.EncodeToString(b), nil
}

func telegramAPI(ctx context.Context, method string, payload interface{}, result interface{}) error {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+telegramBotToken()+"/"+method, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("Telegram API request: %w", err)
	}
	defer resp.Body.Close()
	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		Description string          `json:"description"`
	}
	if json.NewDecoder(resp.Body).Decode(&envelope) != nil {
		return errors.New("invalid Telegram API response")
	}
	if !envelope.OK {
		return errors.New(envelope.Description)
	}
	if result != nil {
		return json.Unmarshal(envelope.Result, result)
	}
	return nil
}

func telegramCreateInvoiceLink(ctx context.Context, payload string, stars int64, orderID string) (string, error) {
	var link string
	err := telegramAPI(ctx, "createInvoiceLink", map[string]interface{}{"title": "PPNet balance recharge", "description": fmt.Sprintf("Recharge order %s", orderID), "payload": payload, "currency": "XTR", "prices": []map[string]interface{}{{"label": "PPNet balance", "amount": stars}}}, &link)
	return link, err
}

func (s *PaymentService) HandleTelegramStarsUpdate(ctx context.Context, update TelegramStarsUpdate) error {
	if q := update.PreCheckoutQuery; q != nil {
		ok, message := s.validateStarsPayment(ctx, q.InvoicePayload, strconv.FormatInt(q.From.ID, 10), q.Currency, q.TotalAmount)
		payload := map[string]interface{}{"pre_checkout_query_id": q.ID, "ok": ok}
		if !ok {
			payload["error_message"] = message
		}
		return telegramAPI(ctx, "answerPreCheckoutQuery", payload, nil)
	}
	if update.Message == nil || update.Message.SuccessfulPayment == nil {
		return nil
	}
	p := update.Message.SuccessfulPayment
	if ok, msg := s.validateStarsPayment(ctx, p.InvoicePayload, strconv.FormatInt(update.Message.From.ID, 10), p.Currency, p.TotalAmount); !ok {
		return errors.New(msg)
	}
	var order dtos.PaymentOrder
	if err := s.db.WithContext(ctx).Where("invoice_payload = ?", p.InvoicePayload).First(&order).Error; err != nil {
		return err
	}
	if order.Status == 1 {
		return nil
	}
	if err := s.db.Model(&order).Updates(map[string]interface{}{"telegram_charge_id": p.TelegramPaymentChargeID, "provider_charge_id": p.ProviderPaymentChargeID}).Error; err != nil {
		return err
	}
	return s.processPaymentSuccess(ctx, &order, p.TelegramPaymentChargeID, time.Now(), order.Cost)
}

func (s *PaymentService) validateStarsPayment(ctx context.Context, payload, userID, currency string, total int64) (bool, string) {
	var order dtos.PaymentOrder
	if s.db.WithContext(ctx).Where("invoice_payload = ?", payload).First(&order).Error != nil {
		return false, "Order not found"
	}
	if order.Status != 0 {
		return false, "Order is no longer payable"
	}
	if order.TelegramUserID != userID {
		return false, "Telegram account mismatch"
	}
	if currency != "XTR" || order.StarsAmount != total {
		return false, "Payment amount mismatch"
	}
	return true, ""
}

func TelegramRefundStarPayment(ctx context.Context, userID, chargeID string) error {
	return telegramAPI(ctx, "refundStarPayment", map[string]string{"user_id": userID, "telegram_payment_charge_id": chargeID}, nil)
}
func TelegramGetStarTransactions(ctx context.Context, offset, limit int) (json.RawMessage, error) {
	var result json.RawMessage
	err := telegramAPI(ctx, "getStarTransactions", map[string]int{"offset": offset, "limit": limit}, &result)
	return result, err
}

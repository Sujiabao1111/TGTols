package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gogogo/helpers"
	"gogogo/models/dtos"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func currentTONUSD(ctx context.Context, cfg helpers.TONConfig) float64 {
	if cfg.TestMode {
		if cfg.TestUSDPerTON > 0 {
			return cfg.TestUSDPerTON
		}
		return 500 // 5 USD = 0.01 TON, for payment testing.
	}
	fallback := cfg.USDPerTON
	if fallback <= 0 {
		fallback = 5
	}
	rateURL := cfg.RateURL
	if strings.TrimSpace(rateURL) == "" {
		rateURL = "https://api.coingecko.com/api/v3/simple/price?ids=the-open-network&vs_currencies=usd"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rateURL, nil)
	if err != nil {
		return fallback
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return fallback
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fallback
	}
	var data struct {
		TON struct {
			USD float64 `json:"usd"`
		} `json:"the-open-network"`
	}
	if json.NewDecoder(resp.Body).Decode(&data) == nil && data.TON.USD > 0 {
		return data.TON.USD
	}
	return fallback
}

func (s *PaymentService) CurrentTONRate(ctx context.Context) float64 {
	cfg := helpers.GetCfgInstance().Conf.Payment.TON
	return currentTONUSD(ctx, cfg)
}

// CreateTonOrder creates a server-owned invoice. The recipient and rate are never
// accepted from the client, preventing users from redirecting funds or changing credit.
func (s *PaymentService) CreateTonOrder(ctx context.Context, userID uint64, amountUSD float64, walletAddr string) (*dtos.TonCreateOrderResponse, error) {
	if math.IsNaN(amountUSD) || math.IsInf(amountUSD, 0) || amountUSD < 1 || amountUSD > 100000 {
		return nil, errors.New("TON payment amount must be between 1 and 100000 USD")
	}
	canonicalSender, err := canonicalTONAddress(walletAddr)
	if err != nil {
		return nil, errors.New("invalid TON wallet address")
	}
	cfg := helpers.GetCfgInstance().Conf.Payment.TON
	addr := strings.TrimSpace(cfg.WalletAddress)
	if addr == "" {
		return nil, errors.New("TON payments are not configured")
	}
	if _, err := canonicalTONAddress(addr); err != nil {
		return nil, errors.New("configured TON payment address is invalid")
	}
	rate := currentTONUSD(ctx, cfg)
	ton := amountUSD / rate
	nano := int64(math.Round(ton * 1e9))
	orderID := "TON-" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	order := &dtos.PaymentOrder{UserID: userID, OrderID: orderID, Amount: amountUSD, Type: "ton", DstCode: "TON", Status: 0, TonAmount: ton, TonRate: rate, TonSender: canonicalSender, ProductInfo: fmt.Sprintf("%.9f TON", ton), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, err
	}
	exp := time.Now().Add(30 * time.Minute)
	return &dtos.TonCreateOrderResponse{OrderID: orderID, WalletAddr: addr, AmountUSD: amountUSD, AmountTON: ton, NanoTON: nano, Comment: orderID, ExpiresAt: exp.Format(time.RFC3339)}, nil
}

// ConfirmTonOrder is intentionally an idempotent server-side settlement hook.
// A blockchain indexer should call it only after validating destination, value and
// comment; txHash is retained as the provider reference for auditability.
func (s *PaymentService) ConfirmTonOrder(ctx context.Context, userID uint64, orderID, txHash, walletAddr string) error {
	if strings.TrimSpace(txHash) == "" {
		return errors.New("transaction hash is required")
	}
	var o dtos.PaymentOrder
	if err := s.db.WithContext(ctx).Where("order_id=? AND user_id=?", orderID, userID).First(&o).Error; err != nil {
		return err
	}
	if o.Status == 1 {
		return nil
	}
	canonicalWallet, err := canonicalTONAddress(walletAddr)
	if err != nil || canonicalWallet != o.TonSender {
		return errors.New("wallet address does not match order")
	}
	if err := verifyTonTransaction(ctx, txHash, o); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Model(&o).Updates(map[string]interface{}{"ton_tx_hash": txHash}).Error; err != nil {
		return err
	}
	return s.processPaymentSuccess(ctx, &o, txHash, time.Now(), o.Amount)
}

func verifyTonTransaction(ctx context.Context, hash string, o dtos.PaymentOrder) error {
	cfg := helpers.GetCfgInstance().Conf.Payment.TON
	endpoint := cfg.APIURL
	if endpoint == "" {
		endpoint = "https://toncenter.com/api/v2/getTransactions"
	}
	u, _ := url.Parse(endpoint)
	q := u.Query()
	q.Set("address", cfg.WalletAddress)
	q.Set("limit", "100")
	// toncenter's getTransactions endpoint does not consistently support a
	// hash query filter; fetch recent transactions and match the hash locally.
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if cfg.APIKey != "" {
		req.Header.Set("X-API-Key", cfg.APIKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ton verification unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("ton verification failed: http %d", resp.StatusCode)
	}
	var body struct {
		Ok     bool `json:"ok"`
		Result []struct {
			TransactionID struct {
				Hash string `json:"hash"`
			} `json:"transaction_id"`
			InMsg struct {
				Value       string `json:"value"`
				Message     string `json:"message"`
				Destination string `json:"destination"`
				Source      string `json:"source"`
			} `json:"in_msg"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	if !body.Ok {
		return errors.New("ton api rejected transaction")
	}
	for _, tx := range body.Result {
		source, sourceErr := canonicalTONAddress(tx.InMsg.Source)
		destination, destinationErr := canonicalTONAddress(tx.InMsg.Destination)
		merchant, merchantErr := canonicalTONAddress(cfg.WalletAddress)
		if sameTONHash(tx.TransactionID.Hash, hash) && sourceErr == nil && destinationErr == nil && merchantErr == nil && tx.InMsg.Message == o.OrderID && destination == merchant && source == o.TonSender {
			if v, e := strconv.ParseInt(tx.InMsg.Value, 10, 64); e == nil && math.Abs(float64(v)-o.TonAmount*1e9) <= 2 {
				return nil
			}
		}
	}
	return errors.New("transaction does not match order")
}

// ScanTonPendingOrders checks recent chain activity for every pending TON order.
// It is safe to run repeatedly because settlement is idempotent.
func (s *PaymentService) ScanTonPendingOrders(ctx context.Context) error {
	var orders []dtos.PaymentOrder
	if err := s.db.WithContext(ctx).Where("type = ? AND dst_code = ? AND status = ? AND created_at > ?", "ton", "TON", 0, time.Now().Add(-2*time.Hour)).Limit(100).Find(&orders).Error; err != nil {
		return err
	}
	for _, order := range orders {
		cfg := helpers.GetCfgInstance().Conf.Payment.TON
		endpoint := cfg.APIURL
		if endpoint == "" {
			endpoint = "https://toncenter.com/api/v2/getTransactions"
		}
		u, _ := url.Parse(endpoint)
		q := u.Query()
		q.Set("address", cfg.WalletAddress)
		q.Set("limit", "100")
		u.RawQuery = q.Encode()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if cfg.APIKey != "" {
			req.Header.Set("X-API-Key", cfg.APIKey)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		var body struct {
			Ok     bool `json:"ok"`
			Result []struct {
				TransactionID struct {
					Hash string `json:"hash"`
				} `json:"transaction_id"`
				InMsg struct {
					Value       string `json:"value"`
					Message     string `json:"message"`
					Destination string `json:"destination"`
				} `json:"in_msg"`
			} `json:"result"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if !body.Ok {
			continue
		}
		for _, tx := range body.Result {
			if tx.InMsg.Message == order.OrderID && tx.TransactionID.Hash != "" {
				if err := s.ConfirmTonOrder(ctx, order.UserID, order.OrderID, tx.TransactionID.Hash, order.TonSender); err == nil {
					break
				}
			}
		}
	}
	return nil
}

func canonicalTONAddress(value string) (string, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ":")
	if len(parts) == 2 {
		wc, err := strconv.ParseInt(parts[0], 10, 8)
		hash := strings.ToLower(parts[1])
		if err != nil || len(hash) != 64 {
			return "", errors.New("invalid raw TON address")
		}
		if _, err = hex.DecodeString(hash); err != nil {
			return "", errors.New("invalid raw TON address")
		}
		return fmt.Sprintf("%d:%s", wc, hash), nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(value, "="))
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(value, "="))
	}
	if err != nil || len(raw) != 36 || crc16Xmodem(raw[:34]) != uint16(raw[34])<<8|uint16(raw[35]) {
		return "", errors.New("invalid friendly TON address")
	}
	return fmt.Sprintf("%d:%x", int8(raw[1]), raw[2:34]), nil
}

func crc16Xmodem(data []byte) uint16 {
	var crc uint16
	for _, b := range data {
		crc ^= uint16(b) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func sameTONHash(a, b string) bool {
	normalize := func(value string) string {
		value = strings.TrimSpace(value)
		if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
			return fmt.Sprintf("%x", decoded)
		}
		if decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(value, "=")); err == nil {
			return fmt.Sprintf("%x", decoded)
		}
		return strings.ToLower(value)
	}
	return normalize(a) != "" && normalize(a) == normalize(b)
}

func (s *PaymentService) ExpireTonOrders(ctx context.Context) error {
	return s.db.WithContext(ctx).Model(&dtos.PaymentOrder{}).Where("type = ? AND status = ? AND created_at < ?", "ton", 0, time.Now().Add(-30*time.Minute)).Updates(map[string]interface{}{"status": 2, "ref_msg": "expired", "updated_at": time.Now()}).Error
}

package services

import (
	"context"
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

func currentTONUSD(ctx context.Context, fallback float64) float64 {
	if fallback <= 0 {
		fallback = 5
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.coingecko.com/api/v3/simple/price?ids=the-open-network&vs_currencies=usd", nil)
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
	return currentTONUSD(ctx, cfg.USDPerTON)
}

// CreateTonOrder creates a server-owned invoice. The recipient and rate are never
// accepted from the client, preventing users from redirecting funds or changing credit.
func (s *PaymentService) CreateTonOrder(ctx context.Context, userID uint64, amountUSD float64) (*dtos.TonCreateOrderResponse, error) {
	if amountUSD <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	cfg := helpers.GetCfgInstance().Conf.Payment.TON
	addr := strings.TrimSpace(cfg.WalletAddress)
	if addr == "" {
		return nil, errors.New("TON payments are not configured")
	}
	rate := currentTONUSD(ctx, cfg.USDPerTON)
	ton := amountUSD / rate
	nano := int64(math.Round(ton * 1e9))
	orderID := "TON-" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	order := &dtos.PaymentOrder{UserID: userID, OrderID: orderID, Amount: amountUSD, Type: "ton", DstCode: "TON", Status: 0, TonAmount: ton, TonRate: rate, ProductInfo: fmt.Sprintf("%.9f TON", ton), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, err
	}
	exp := time.Now().Add(30 * time.Minute)
	return &dtos.TonCreateOrderResponse{OrderID: orderID, WalletAddr: addr, AmountUSD: amountUSD, AmountTON: ton, NanoTON: nano, Comment: orderID, ExpiresAt: exp.Format(time.RFC3339)}, nil
}

// ConfirmTonOrder is intentionally an idempotent server-side settlement hook.
// A blockchain indexer should call it only after validating destination, value and
// comment; txHash is retained as the provider reference for auditability.
func (s *PaymentService) ConfirmTonOrder(ctx context.Context, userID uint64, orderID, txHash string) error {
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
	q.Set("hash", hash)
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
			InMsg struct {
				Value       string `json:"value"`
				Message     string `json:"message"`
				Destination string `json:"destination"`
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
		if tx.InMsg.Message == o.OrderID && tx.InMsg.Destination == cfg.WalletAddress {
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
				if err := s.ConfirmTonOrder(ctx, order.UserID, order.OrderID, tx.TransactionID.Hash); err == nil {
					break
				}
			}
		}
	}
	return nil
}

func (s *PaymentService) ExpireTonOrders(ctx context.Context) error {
	return s.db.WithContext(ctx).Model(&dtos.PaymentOrder{}).Where("type = ? AND status = ? AND created_at < ?", "ton", 0, time.Now().Add(-30*time.Minute)).Updates(map[string]interface{}{"status": 2, "ref_msg": "expired", "updated_at": time.Now()}).Error
}

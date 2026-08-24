package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type M7PPConfig struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	VendorCode string
	Currency   string
	ReturnURL  string
}

type M7PPClient struct {
	client *resty.Client
	cfg    M7PPConfig
}

type M7PPError struct {
	Op      string
	Status  string
	Message string
}

func (e M7PPError) Error() string {
	return fmt.Sprintf("m7pp %s failed: status=%s message=%s", e.Op, e.Status, e.Message)
}

type m7ppBaseResponse struct {
	TraceID string `json:"traceId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type M7PPGame struct {
	GameCode       string
	GameName       string
	CategoryCode   string
	ImageSquare    string
	ImageLandscape string
	LanguageCode   string
	PlatformCode   string
	CurrencyCode   string
}

type M7PPTransaction struct {
	BetID                 string
	RoundID               string
	ExternalTransactionID string
	Username              string
	CurrencyCode          string
	GameCode              string
	VendorCode            string
	GameCategoryCode      string
	BetAmount             float64
	WinAmount             float64
	WinLoss               float64
	EffectiveTurnover     float64
	JackpotAmount         float64
	Status                int
	VendorBetTime         int64
	VendorSettleTime      int64
	CreateTime            int64
	UpdateTime            int64
}

type m7ppGameListResponse struct {
	m7ppBaseResponse
	Data struct {
		Headers     map[string]int `json:"headers"`
		Games       [][]any        `json:"games"`
		CurrentPage int            `json:"currentPage"`
		TotalPages  int            `json:"totalPages"`
		TotalItems  int            `json:"totalItems"`
	} `json:"data"`
}

type m7ppGameURLResponse struct {
	m7ppBaseResponse
	Data struct {
		GameURL string `json:"gameUrl"`
		Token   string `json:"token"`
	} `json:"data"`
}

type m7ppTransactionListResponse struct {
	m7ppBaseResponse
	Data struct {
		Header       map[string]int `json:"header"`
		Headers      map[string]int `json:"headers"`
		Transactions [][]any        `json:"transactions"`
		CurrentPage  int            `json:"currentPage"`
		TotalPages   int            `json:"totalPages"`
		TotalItems   int            `json:"totalItems"`
	} `json:"data"`
}

type m7ppCashResponse struct {
	m7ppBaseResponse
	Data struct {
		ReferenceID    string  `json:"referenceId"`
		TransactionID  string  `json:"transactionId"`
		Username       string  `json:"username"`
		Currency       string  `json:"currency"`
		CurrencyCode   string  `json:"currencyCode"`
		Amount         float64 `json:"amount"`
		TransferAmount float64 `json:"transferAmount"`
		Timestamp      int64   `json:"timestamp"`
	} `json:"data"`
}

func NewM7PPClient(cfg M7PPConfig) (*M7PPClient, error) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://gw.boan.games"
	}
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.APISecret = strings.TrimSpace(cfg.APISecret)
	if cfg.APIKey == "" || cfg.APISecret == "" || strings.Contains(cfg.APIKey, "PLACEHOLDER") || strings.Contains(cfg.APISecret, "PLACEHOLDER") {
		return nil, fmt.Errorf("m7pp api_key/api_secret is not configured")
	}
	cfg.VendorCode = strings.ToUpper(strings.TrimSpace(cfg.VendorCode))
	if cfg.VendorCode == "" {
		cfg.VendorCode = "BOANZZPP"
	}
	cfg.Currency = strings.ToUpper(strings.TrimSpace(cfg.Currency))
	if cfg.Currency == "" {
		cfg.Currency = "USD"
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		// BOAN validates the caller's network path. Do not inherit container
		// HTTP(S)_PROXY settings, otherwise the request may leave through a
		// different proxy IP even though the socket itself is IPv4.
		Proxy: nil,
		DialContext: func(ctx context.Context, _ string, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", address)
		},
		ForceAttemptHTTP2: true,
	}
	return &M7PPClient{
		cfg: cfg,
		client: resty.New().
			SetTimeout(10 * time.Second).
			SetTransport(transport),
	}, nil
}

func SignM7PPBody(rawBody []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *M7PPClient) post(path string, payload any, result any) error {
	rawBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("m7pp marshal %s request: %w", path, err)
	}
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-API-Key", c.cfg.APIKey).
		SetHeader("X-Signature", SignM7PPBody(rawBody, c.cfg.APISecret)).
		SetBody(rawBody).
		SetResult(result).
		Post(c.cfg.BaseURL + path)
	if err != nil {
		return fmt.Errorf("m7pp %s network error: %w", path, err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("m7pp %s http status %d", path, resp.StatusCode())
	}
	return nil
}

func m7ppCheck(op string, resp m7ppBaseResponse) error {
	if resp.Status == "SC_OK" {
		return nil
	}
	return M7PPError{Op: op, Status: resp.Status, Message: resp.Message}
}

// IsM7PPWalletNotInitializedError recognizes BOAN's test-environment response
// for a player who has never had a transfer wallet created. Although the API
// document lists SC_USER_NOT_EXISTS, the gateway currently returns
// SC_INTERNAL_ERROR from /cash/balance for this case.
func IsM7PPWalletNotInitializedError(err error) bool {
	var providerErr M7PPError
	if !errors.As(err, &providerErr) {
		return false
	}
	if providerErr.Op != "cash/balance" {
		return false
	}
	switch providerErr.Status {
	case "SC_USER_NOT_EXISTS", "SC_TRANSACTION_DOES_NOT_EXIST", "SC_INTERNAL_ERROR":
		return true
	default:
		return false
	}
}

func (c *M7PPClient) FetchGames() ([]M7PPGame, error) {
	all := make([]M7PPGame, 0, 100)
	for page := 1; ; page++ {
		var resp m7ppGameListResponse
		err := c.post("/game/list", map[string]any{
			"traceId": uuid.NewString(), "vendorCode": c.cfg.VendorCode, "pageNo": page,
			"pageSize": 100, "displayLanguage": "en", "currency": c.cfg.Currency,
		}, &resp)
		if err != nil {
			return nil, err
		}
		if err := m7ppCheck("game/list", resp.m7ppBaseResponse); err != nil {
			return nil, err
		}
		index := func(name string) int {
			if value, ok := resp.Data.Headers[name]; ok {
				return value
			}
			return -1
		}
		for _, row := range resp.Data.Games {
			all = append(all, M7PPGame{
				GameCode: valueAt(row, index("gameCode")), GameName: valueAt(row, index("gameName")),
				CategoryCode: valueAt(row, index("categoryCode")), ImageSquare: valueAt(row, index("imageSquare")),
				ImageLandscape: valueAt(row, index("imageLandscape")), LanguageCode: valueAt(row, index("languageCode")),
				PlatformCode: valueAt(row, index("platformCode")), CurrencyCode: valueAt(row, index("currencyCode")),
			})
		}
		if resp.Data.TotalPages <= page || len(resp.Data.Games) == 0 {
			break
		}
	}
	return all, nil
}

func valueAt(row []any, index int) string {
	if index < 0 || index >= len(row) || row[index] == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(row[index]))
}

func floatAt(row []any, index int) float64 {
	value := valueAt(row, index)
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func int64At(row []any, index int) int64 {
	if index < 0 || index >= len(row) || row[index] == nil {
		return 0
	}
	switch value := row[index].(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case float32:
		return int64(value)
	case json.Number:
		if parsed, err := value.Int64(); err == nil {
			return parsed
		}
		parsed, _ := strconv.ParseFloat(value.String(), 64)
		return int64(parsed)
	case string:
		if parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
			return parsed
		}
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return int64(parsed)
	default:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
		return int64(parsed)
	}
}

func (c *M7PPClient) FetchTransactions(fromTime, toTime int64, pageNo, pageSize int) ([]M7PPTransaction, int, error) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 5000 {
		pageSize = 2000
	}
	var resp m7ppTransactionListResponse
	err := c.post("/transaction/list", map[string]any{
		"traceId": uuid.NewString(), "fromTime": fromTime, "toTime": toTime,
		"pageNo": pageNo, "pageSize": pageSize,
	}, &resp)
	if err != nil {
		return nil, 0, err
	}
	if err := m7ppCheck("transaction/list", resp.m7ppBaseResponse); err != nil {
		return nil, 0, err
	}
	header := resp.Data.Headers
	if len(header) == 0 {
		header = resp.Data.Header
	}
	index := func(name string) int {
		if value, ok := header[name]; ok {
			return value
		}
		return -1
	}
	items := make([]M7PPTransaction, 0, len(resp.Data.Transactions))
	for _, row := range resp.Data.Transactions {
		items = append(items, M7PPTransaction{
			BetID: valueAt(row, index("betId")), RoundID: valueAt(row, index("roundId")),
			ExternalTransactionID: valueAt(row, index("externalTransactionId")), Username: valueAt(row, index("username")),
			CurrencyCode: valueAt(row, index("currencyCode")), GameCode: valueAt(row, index("gameCode")),
			VendorCode: valueAt(row, index("vendorCode")), GameCategoryCode: valueAt(row, index("gameCategoryCode")),
			BetAmount: floatAt(row, index("betAmount")), WinAmount: floatAt(row, index("winAmount")),
			WinLoss: floatAt(row, index("winLoss")), EffectiveTurnover: floatAt(row, index("effectiveTurnover")),
			JackpotAmount: floatAt(row, index("jackpotAmount")), Status: int(int64At(row, index("status"))),
			VendorBetTime: int64At(row, index("vendorBetTime")), VendorSettleTime: int64At(row, index("vendorSettleTime")),
			CreateTime: int64At(row, index("createTime")), UpdateTime: int64At(row, index("updateTime")),
		})
	}
	return items, resp.Data.TotalPages, nil
}

func (c *M7PPClient) LaunchGame(username, gameCode, language, platform, ipAddress string) (string, error) {
	var resp m7ppGameURLResponse
	err := c.post("/game/url", map[string]any{
		"traceId": uuid.NewString(), "username": username, "gameCode": gameCode,
		"language": normalizeM7PPLanguage(language), "platform": platform,
		"currency": c.cfg.Currency, "lobbyUrl": c.cfg.ReturnURL, "ipAddress": ipAddress,
	}, &resp)
	if err != nil {
		return "", err
	}
	if err := m7ppCheck("game/url", resp.m7ppBaseResponse); err != nil {
		return "", err
	}
	if strings.TrimSpace(resp.Data.GameURL) == "" {
		return "", fmt.Errorf("m7pp game/url returned empty gameUrl")
	}
	return resp.Data.GameURL, nil
}

func normalizeM7PPLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "cn", "zh-cn", "zh":
		return "zh"
	case "tw", "zh-tw", "hk":
		return "hk"
	case "id", "vi", "hi", "pt", "th", "es", "tl":
		return strings.ToLower(strings.TrimSpace(language))
	default:
		return "en"
	}
}

func (c *M7PPClient) Balance(username string) (float64, error) {
	var resp m7ppCashResponse
	err := c.post("/cash/balance", map[string]any{
		"username": username, "vendorCode": c.cfg.VendorCode,
		"traceId": uuid.NewString(), "currency": c.cfg.Currency,
	}, &resp)
	if err != nil {
		return 0, err
	}
	if err := m7ppCheck("cash/balance", resp.m7ppBaseResponse); err != nil {
		return 0, err
	}
	return resp.Data.Amount, nil
}

func (c *M7PPClient) Deposit(username string, amount float64, referenceID string) (float64, error) {
	return c.transfer("/cash/deposit", "cash/deposit", username, amount, referenceID)
}

func (c *M7PPClient) Withdraw(username string, amount float64, referenceID string) (float64, error) {
	return c.transfer("/cash/withdraw", "cash/withdraw", username, amount, referenceID)
}

func (c *M7PPClient) transfer(path, op, username string, amount float64, referenceID string) (float64, error) {
	var resp m7ppCashResponse
	err := c.post(path, map[string]any{
		"username": username, "vendorCode": c.cfg.VendorCode, "traceId": uuid.NewString(),
		"transferAmount": fmt.Sprintf("%.2f", amount), "currency": c.cfg.Currency, "referenceId": referenceID,
	}, &resp)
	if err != nil {
		return 0, err
	}
	if err := m7ppCheck(op, resp.m7ppBaseResponse); err != nil {
		return 0, err
	}
	return c.Balance(username)
}

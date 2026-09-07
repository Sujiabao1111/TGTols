package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentService struct {
	db         *gorm.DB
	httpClient *http.Client
}

const (
	withdrawOrderStatusPending    = 0
	withdrawOrderStatusSuccess    = 1
	withdrawOrderStatusFailed     = 2
	withdrawOrderStatusProcessing = 3
	minWithdrawAmountUSD          = 10.0
	// Withdrawals below this USD amount are submitted automatically.
	autoWithdrawLimitUSD           = 20.0
	dailyWithdrawLimit             = int64(5)
	defaultWithdrawFeeConfigKey    = "default"
	defaultWithdrawPlatformFeeRate = 0.003
)

func shouldAutoSubmitWithdraw(amountUSD float64) bool {
	return amountUSD < autoWithdrawLimitUSD
}

func tonWithdrawConfig() helpers.TONConfig {
	cfg := helpers.GetCfgInstance()
	if cfg == nil || cfg.Conf == nil {
		return helpers.TONConfig{}
	}
	return cfg.Conf.Payment.TON
}

type withdrawFeeConfigValues struct {
	Enabled bool
	FeeRate float64
}

type paymentMethodConfig struct {
	Code        string
	Name        string
	Type        string
	Currency    string
	Channel     string
	AccountType string
	MinAmount   float64
	MaxAmount   float64
	Enabled     bool
	Description string
}

type paymentMethodOption struct {
	Code string
	Name string
}

var depositMethodConfigs = []paymentMethodConfig{
	{Code: "TON", Name: "TON", Type: "crypto", Currency: "TON", Channel: "ton", MinAmount: 1, MaxAmount: 100000, Enabled: true, Description: "Native TON wallet transfer"},
	{Code: "TG_STARS", Name: "Telegram Stars", Type: "channel", Currency: "XTR", Channel: "telegram", MinAmount: 1, MaxAmount: 100000, Enabled: true, Description: "Telegram Stars"},
	{Code: "GCASH_QR", Name: "GCash QR", Type: "channel", Currency: "PHP", Channel: "quantixcore", MinAmount: 100, MaxAmount: 50000, Enabled: true, Description: "Philippines GCash QR payment"},
	{Code: "GCASH_APP", Name: "GCash App", Type: "channel", Currency: "PHP", Channel: "quantixcore", MinAmount: 100, MaxAmount: 50000, Enabled: true, Description: "Philippines GCash app payment"},
	{Code: "DANA", Name: "DANA", Type: "ewallet", Currency: "IDR", Channel: "gateway", MinAmount: 10000, MaxAmount: 20000000, Enabled: true, Description: "Indonesia DANA wallet"},
	{Code: "OVO", Name: "OVO", Type: "ewallet", Currency: "IDR", Channel: "gateway", MinAmount: 10000, MaxAmount: 20000000, Enabled: true, Description: "Indonesia OVO wallet"},
	{Code: "LINKAJA", Name: "LinkAja", Type: "ewallet", Currency: "IDR", Channel: "gateway", MinAmount: 10000, MaxAmount: 20000000, Enabled: true, Description: "Indonesia LinkAja wallet"},
	{Code: "IDR_QRIS", Name: "QRIS", Type: "qris", Currency: "IDR", Channel: "gateway", MinAmount: 10000, MaxAmount: 10000000, Enabled: true, Description: "Indonesia QRIS payment"},
	{Code: "IDR_VA", Name: "Virtual Account", Type: "bank", Currency: "IDR", Channel: "gateway", MinAmount: 10000, MaxAmount: 50000000, Enabled: true, Description: "Indonesia virtual account payment"},
	{Code: "USDT", Name: "USDT", Type: "channel", Currency: "USDT", Channel: "tokenpay", MinAmount: 77500, MaxAmount: 15500000, Enabled: true, Description: "USDT via TokenPay"},
	{Code: "PAYPAL", Name: "PayPal", Type: "channel", Currency: "USD", Channel: "paypal", MinAmount: 10000, MaxAmount: 100000000, Enabled: false, Description: "Coming soon"},
	{Code: "199VOUCHER", Name: "199 Voucher", Type: "voucher", Currency: "IDR", Channel: "voucher", MinAmount: 10000, MaxAmount: 1000000, Enabled: true, Description: "Voucher recharge"},
}

var legacyGatewayWithdrawWalletOptions = []paymentMethodOption{
	{Code: "OVO", Name: "OVO"},
	{Code: "DANA", Name: "DANA"},
	{Code: "GOPAY", Name: "GoPay"},
	{Code: "SHOPEEPAY", Name: "ShopeePay"},
	{Code: "LINKAJA", Name: "LinkAja"},
}

func buildLegacyGatewayWithdrawMethodConfigs() []paymentMethodConfig {
	methods := make([]paymentMethodConfig, 0, len(legacyGatewayWithdrawWalletOptions))

	for _, option := range legacyGatewayWithdrawWalletOptions {
		methods = append(methods, paymentMethodConfig{
			Code:        option.Code,
			Name:        option.Name,
			Type:        "ewallet",
			Currency:    "IDR",
			Channel:     "gateway",
			AccountType: option.Code,
			MinAmount:   10000,
			MaxAmount:   25000000,
			Enabled:     true,
			Description: "Indonesia e-wallet payout",
		})
	}

	return methods
}

var withdrawMethodConfigs = func() []paymentMethodConfig {
	methods := []paymentMethodConfig{
		{Code: "GCASH_QR", Name: "GCash", Type: "ewallet", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: true, Description: "Philippines GCash wallet"},
		{Code: "TON", Name: "TON Wallet", Type: "ewallet", Currency: "USD", Channel: "manual", AccountType: "TON", MinAmount: 10, MaxAmount: 100000, Enabled: true, Description: "TON wallet withdrawal pending review"},
	}

	methods = append(methods, buildLegacyGatewayWithdrawMethodConfigs()...)
	methods = append(methods,
		paymentMethodConfig{Code: "TON", Name: "TON Wallet", Type: "crypto", Currency: "TON", Channel: "ton", MinAmount: 0, MaxAmount: 0, Enabled: false, Description: "TON wallet withdrawal (coming soon)"},
		paymentMethodConfig{Code: "BDO", Name: "Banco de Oro", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "BPI", Name: "Bank of the Philippine Islands", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "METROBANK", Name: "Metrobank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "LANDBANK", Name: "LandBank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "PNB", Name: "Philippine National Bank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "SECURITY", Name: "Security Bank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "UNIONBANK", Name: "UnionBank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "CHINABANK", Name: "China Bank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "RCBC", Name: "RCBC", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
		paymentMethodConfig{Code: "EASTWEST", Name: "EastWest Bank", Type: "bankcard", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH", MinAmount: 100, MaxAmount: 50000, Enabled: false, Description: "Philippines bank transfer"},
	)

	return methods
}()

var quantixCoreWalletAccountTypes = map[string]string{
	"GCASH":     "GCASH",
	"GCASH_QR":  "GCASH",
	"GCASH_APP": "GCASH",
	"GCASH_H5":  "GCASH",
	"OVO":       "OVO",
	"DANA":      "DANA",
	"GOPAY":     "GOPAY",
	"SHOPEEPAY": "SHOPEEPAY",
	"LINKAJA":   "LINKAJA",
	"KASPRO":    "KASPRO",
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *PaymentService) getConfig() *helpers.PaymentConfig {
	instance := helpers.GetCfgInstance()
	if instance == nil || instance.Conf == nil {
		return nil
	}

	cfg := instance.Conf.Payment
	cfg.MerchantID = strings.TrimSpace(cfg.MerchantID)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.NotifyURL = strings.TrimSpace(cfg.NotifyURL)
	cfg.ReturnURL = strings.TrimSpace(cfg.ReturnURL)
	cfg.QuantixCore.BaseURL = strings.TrimSpace(cfg.QuantixCore.BaseURL)
	cfg.QuantixCore.ID.BaseURL = strings.TrimSpace(cfg.QuantixCore.ID.BaseURL)
	cfg.QuantixCore.ID.MerchantID = strings.TrimSpace(cfg.QuantixCore.ID.MerchantID)
	cfg.QuantixCore.ID.SecretKey = strings.TrimSpace(cfg.QuantixCore.ID.SecretKey)
	cfg.QuantixCore.ID.NotifyURL = strings.TrimSpace(cfg.QuantixCore.ID.NotifyURL)
	cfg.QuantixCore.ID.ReturnURL = strings.TrimSpace(cfg.QuantixCore.ID.ReturnURL)
	cfg.QuantixCore.PH.BaseURL = strings.TrimSpace(cfg.QuantixCore.PH.BaseURL)
	cfg.QuantixCore.PH.MerchantID = strings.TrimSpace(cfg.QuantixCore.PH.MerchantID)
	cfg.QuantixCore.PH.SecretKey = strings.TrimSpace(cfg.QuantixCore.PH.SecretKey)
	cfg.QuantixCore.PH.NotifyURL = strings.TrimSpace(cfg.QuantixCore.PH.NotifyURL)
	cfg.QuantixCore.PH.ReturnURL = strings.TrimSpace(cfg.QuantixCore.PH.ReturnURL)
	cfg.TokenPay.BaseURL = strings.TrimSpace(cfg.TokenPay.BaseURL)
	cfg.TokenPay.SecretKey = strings.TrimSpace(cfg.TokenPay.SecretKey)
	cfg.TokenPay.NotifyURL = strings.TrimSpace(cfg.TokenPay.NotifyURL)
	cfg.TokenPay.ReturnURL = strings.TrimSpace(cfg.TokenPay.ReturnURL)
	return &cfg
}

func (s *PaymentService) getQuantixCoreConfig(cfg *helpers.PaymentConfig, currency string) helpers.QuantixCoreRegionConfig {
	defaultCfg := helpers.QuantixCoreRegionConfig{
		BaseURL:    firstNonEmpty(cfg.QuantixCore.BaseURL, cfg.BaseURL),
		MerchantID: cfg.MerchantID,
		SecretKey:  cfg.SecretKey,
		NotifyURL:  cfg.NotifyURL,
		ReturnURL:  cfg.ReturnURL,
	}

	regionCfg := helpers.QuantixCoreRegionConfig{}
	switch normalizeCurrency(currency) {
	case "PHP":
		regionCfg = cfg.QuantixCore.PH
	case "IDR":
		regionCfg = cfg.QuantixCore.ID
	}

	return helpers.QuantixCoreRegionConfig{
		BaseURL:    firstNonEmpty(regionCfg.BaseURL, defaultCfg.BaseURL),
		MerchantID: firstNonEmpty(regionCfg.MerchantID, defaultCfg.MerchantID),
		SecretKey:  firstNonEmpty(regionCfg.SecretKey, defaultCfg.SecretKey),
		NotifyURL:  firstNonEmpty(regionCfg.NotifyURL, defaultCfg.NotifyURL),
		ReturnURL:  firstNonEmpty(regionCfg.ReturnURL, defaultCfg.ReturnURL),
	}
}

func (s *PaymentService) getQuantixCoreConfigByMerchant(cfg *helpers.PaymentConfig, merchantID, currency string) helpers.QuantixCoreRegionConfig {
	merchantID = strings.TrimSpace(merchantID)
	phCfg := s.getQuantixCoreConfig(cfg, "PHP")
	if merchantID != "" && phCfg.MerchantID != "" && merchantID == phCfg.MerchantID {
		return phCfg
	}

	idCfg := s.getQuantixCoreConfig(cfg, "IDR")
	if merchantID != "" && idCfg.MerchantID != "" && merchantID == idCfg.MerchantID {
		return idCfg
	}

	return s.getQuantixCoreConfig(cfg, currency)
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func quantixCoreMinorUnitFactor(currency string) float64 {
	switch normalizeCurrency(currency) {
	case "PHP":
		return 100
	default:
		return 1
	}
}

func formatAmountForSign(amount float64, currency string) string {
	factor := quantixCoreMinorUnitFactor(currency)
	return strconv.Itoa(int(math.Round(amount * factor)))
}

func amountToMinorUnits(amount float64, currency string) int {
	return int(math.Round(amount * quantixCoreMinorUnitFactor(currency)))
}

func amountFromMinorUnits(amount int, currency string) float64 {
	return float64(amount) / quantixCoreMinorUnitFactor(currency)
}

func normalizeWithdrawFeeConfig(config dtos.WithdrawFeeConfig) withdrawFeeConfigValues {
	values := withdrawFeeConfigValues{
		Enabled: config.Enabled,
		FeeRate: config.FeeRate,
	}
	if values.FeeRate < 0 || values.FeeRate > 1 {
		values.FeeRate = defaultWithdrawPlatformFeeRate
	}
	return values
}

func (s *PaymentService) getWithdrawFeeConfig(ctx context.Context) withdrawFeeConfigValues {
	if s == nil || s.db == nil {
		return withdrawFeeConfigValues{Enabled: true, FeeRate: defaultWithdrawPlatformFeeRate}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var config dtos.WithdrawFeeConfig
	if err := s.db.WithContext(ctx).Where("config_key = ?", defaultWithdrawFeeConfigKey).Take(&config).Error; err != nil {
		return withdrawFeeConfigValues{Enabled: true, FeeRate: defaultWithdrawPlatformFeeRate}
	}
	return normalizeWithdrawFeeConfig(config)
}

func withdrawGatewayAmount(amount float64, currency string, feeConfig withdrawFeeConfigValues) float64 {
	feeRate := 0.0
	if feeConfig.Enabled {
		feeRate = feeConfig.FeeRate
	}
	if feeRate < 0 {
		feeRate = 0
	}
	if feeRate > 1 {
		feeRate = 1
	}

	netAmount := amount * (1 - feeRate)
	factor := quantixCoreMinorUnitFactor(currency)
	if factor <= 0 {
		return netAmount
	}
	return math.Round(netAmount*factor) / factor
}

func normalizeLegacyGatewayDstCode(dstCode string) string {
	switch strings.ToUpper(strings.TrimSpace(dstCode)) {
	case "IDR_QRIS":
		return "QRIS"
	case "IDR_VA":
		// Generic VA should open the cashier landing page instead of forcing a specific bank.
		return ""
	default:
		return strings.TrimSpace(dstCode)
	}
}

func (s *PaymentService) getExchangeRateForCurrency(currency string) float64 {
	switch normalizeCurrency(currency) {
	case "PHP":
		exchangeRate, err := models.GetInstance().GetExchangeRate("PHP")
		if err != nil || exchangeRate <= 0 {
			return 56
		}
		return exchangeRate
	case "IDR":
		exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
		if err != nil || exchangeRate <= 0 {
			return 16000
		}
		return exchangeRate
	default:
		return 1
	}
}

func (s *PaymentService) convertLocalAmountToUSD(amount float64, currency string) float64 {
	currency = normalizeCurrency(currency)
	if currency == "" || currency == "USD" {
		return amount
	}

	exchangeRate := s.getExchangeRateForCurrency(currency)
	if exchangeRate <= 0 {
		return amount
	}
	return amount / exchangeRate
}

func (s *PaymentService) findDepositMethod(code string) (paymentMethodConfig, bool) {
	for _, cfg := range depositMethodConfigs {
		if cfg.Code == code {
			return cfg, true
		}
	}
	return paymentMethodConfig{}, false
}

func (s *PaymentService) findWithdrawMethod(code string) (paymentMethodConfig, bool) {
	for _, cfg := range withdrawMethodConfigs {
		if cfg.Code == code {
			return cfg, true
		}
	}
	return paymentMethodConfig{}, false
}

func (s *PaymentService) CreatePaymentOrder(ctx context.Context, userID uint64, req dtos.CreatePaymentRequest) (*dtos.CreatePaymentResponse, error) {
	cfg := s.getConfig()
	if cfg == nil {
		return nil, errors.New("payment config is not initialized")
	}

	methodCfg, ok := s.findDepositMethod(req.DstCode)
	if !ok {
		methodCfg = paymentMethodConfig{
			Code:      req.DstCode,
			Name:      "recharge",
			Type:      req.Type,
			Currency:  "IDR",
			Channel:   "gateway",
			MinAmount: 1,
			MaxAmount: 999999999,
			Enabled:   true,
		}
	}

	if !methodCfg.Enabled {
		return nil, errors.New("payment method is unavailable")
	}
	if req.Amount < methodCfg.MinAmount || req.Amount > methodCfg.MaxAmount {
		return nil, fmt.Errorf("amount is out of range: %.0f - %.0f %s", methodCfg.MinAmount, methodCfg.MaxAmount, methodCfg.Currency)
	}

	if req.Currency == "" {
		req.Currency = methodCfg.Currency
	}
	if req.Channel == "" {
		req.Channel = methodCfg.Channel
	}

	if req.Channel == "quantixcore" {
		return s.createQuantixCorePaymentOrder(ctx, userID, req, methodCfg)
	}

	if cfg.MerchantID == "" || cfg.SecretKey == "" {
		return nil, errors.New("payment config is not initialized")
	}

	orderID := common.GenerateOrderID(userID)
	now := time.Now().Format("2006-01-02 15:04:05")

	callbackURL := req.CallbackURL
	if callbackURL == "" {
		callbackURL = cfg.ReturnURL
	}

	params := map[string]string{
		"memberId":    cfg.MerchantID,
		"orderId":     orderID,
		"amount":      fmt.Sprintf("%.0f", req.Amount),
		"dateTime":    now,
		"notifyUrl":   cfg.NotifyURL,
		"name":        fmt.Sprintf("USER%d", userID),
		"phone":       "088291448849",
		"email":       "user@example.com",
		"lang":        "id",
		"productInfo": "recharge",
		"callbackUrl": callbackURL,
		"version":     "1",
	}

	params["sign"] = common.GenerateSign(params, cfg.SecretKey)

	apiURL := strings.TrimSuffix(cfg.BaseURL, "/") + "/pay_counter.html"
	respData, err := s.postMultipartFormData(ctx, apiURL, params)
	if err != nil {
		return nil, fmt.Errorf("payment gateway request failed: %w", err)
	}

	var gatewayResp struct {
		Status   string `json:"status"`
		Msg      string `json:"msg"`
		ErrCode  int    `json:"errCode"`
		Amount   string `json:"amount"`
		PayURL   string `json:"payUrl"`
		DateTime string `json:"dateTime"`
	}

	if err := json.Unmarshal(respData, &gatewayResp); err != nil {
		return nil, fmt.Errorf("parse gateway response failed: %w", err)
	}
	if gatewayResp.Status != "success" {
		return nil, fmt.Errorf("payment gateway error: %s (errCode: %d)", gatewayResp.Msg, gatewayResp.ErrCode)
	}

	order := dtos.PaymentOrder{
		UserID:      userID,
		OrderID:     orderID,
		PlatOrderID: "",
		Amount:      req.Amount,
		Type:        req.Type,
		DstCode:     req.DstCode,
		Status:      0,
		PayURL:      gatewayResp.PayURL,
		ProductInfo: "recharge",
	}

	if err := s.db.Create(&order).Error; err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	return &dtos.CreatePaymentResponse{
		OrderID:   orderID,
		PayURL:    gatewayResp.PayURL,
		Amount:    req.Amount,
		Status:    0,
		ExpiredAt: "",
	}, nil
}

func (s *PaymentService) createQuantixCorePaymentOrder(ctx context.Context, userID uint64, req dtos.CreatePaymentRequest, methodCfg paymentMethodConfig) (*dtos.CreatePaymentResponse, error) {
	cfg := s.getConfig()
	if cfg == nil {
		return nil, errors.New("payment config is not initialized")
	}

	quantixCfg := s.getQuantixCoreConfig(cfg, req.Currency)
	if quantixCfg.MerchantID == "" || quantixCfg.SecretKey == "" {
		return nil, errors.New("QuantixCore config is not initialized")
	}

	orderID := common.GenerateOrderID(userID)
	amountInMinorUnits := amountToMinorUnits(req.Amount, req.Currency)
	if amountInMinorUnits <= 0 {
		return nil, errors.New("invalid amount")
	}

	returnURL := req.CallbackURL
	if returnURL == "" {
		returnURL = quantixCfg.ReturnURL
	}

	payload := map[string]string{
		"merchantNo":      quantixCfg.MerchantID,
		"merchantOrderNo": orderID,
		"amount":          strconv.Itoa(amountInMinorUnits),
		"code":            req.DstCode,
		"currency":        req.Currency,
		"content":         methodCfg.Name,
		"uid":             strconv.FormatUint(userID, 10),
		"clientIp":        req.ClientIP,
		"callback":        quantixCfg.NotifyURL,
		"return":          returnURL,
	}
	payload["sign"] = generateQuantixCoreSign(payload, quantixCfg.SecretKey)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal QuantixCore request failed: %w", err)
	}

	apiURL := strings.TrimSuffix(s.getQuantixCoreBaseURL(cfg, quantixCfg), "/") + "/api/open/flex/order/trade/add"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create QuantixCore request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("QuantixCore request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read QuantixCore response failed: %w", err)
	}

	var gatewayResp struct {
		Success   bool   `json:"success"`
		ErrorCode string `json:"errorCode"`
		Message   string `json:"message"`
		Data      struct {
			OrderNo         string `json:"orderNo"`
			MerchantOrderNo string `json:"merchantOrderNo"`
			PayInfo         string `json:"payInfo"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respData, &gatewayResp); err != nil {
		return nil, fmt.Errorf("parse QuantixCore response failed: %w", err)
	}
	if !gatewayResp.Success {
		if gatewayResp.Message == "" {
			gatewayResp.Message = "create order failed"
		}
		if gatewayResp.ErrorCode != "" {
			return nil, fmt.Errorf("QuantixCore error: %s (%s)", gatewayResp.Message, gatewayResp.ErrorCode)
		}
		return nil, fmt.Errorf("QuantixCore error: %s", gatewayResp.Message)
	}
	if gatewayResp.Data.PayInfo == "" {
		return nil, errors.New("QuantixCore pay url is empty")
	}

	order := dtos.PaymentOrder{
		UserID:      userID,
		OrderID:     orderID,
		PlatOrderID: gatewayResp.Data.OrderNo,
		Amount:      req.Amount,
		Type:        req.Type,
		DstCode:     req.DstCode,
		Status:      0,
		PayURL:      gatewayResp.Data.PayInfo,
		ProductInfo: methodCfg.Name,
	}

	if err := s.db.Create(&order).Error; err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	return &dtos.CreatePaymentResponse{
		OrderID:   orderID,
		PayURL:    gatewayResp.Data.PayInfo,
		Amount:    req.Amount,
		Status:    0,
		ExpiredAt: "",
	}, nil
}

func (s *PaymentService) HandlePaymentNotify(ctx context.Context, notify dtos.PaymentNotifyRequest) (string, error) {
	notify = normalizePaymentNotify(notify)
	if notify.OrderID == "" {
		return "FAIL", errors.New("order id is empty")
	}

	currency := notify.Currency
	code := notify.Code
	isQuantixCoreCode := currency == "PHP" || strings.HasPrefix(code, "GCASH")
	if isQuantixCoreCode || strings.EqualFold(notify.Status, "PAID") || strings.EqualFold(notify.Status, "PAY_FAILED") || strings.EqualFold(notify.Status, "REFUND") {
		return s.handleQuantixCorePaymentNotify(ctx, notify)
	}

	cfg := s.getConfig()
	if cfg == nil {
		return "FAIL", errors.New("payment config is not initialized")
	}

	legacyParams := map[string]string{
		"merchant_id":   notify.MerchantID,
		"order_id":      notify.OrderID,
		"plat_order_id": notify.PlatOrderID,
		"amount":        fmt.Sprintf("%.2f", notify.Amount),
		"status":        notify.Status,
		"ref_code":      strconv.Itoa(notify.RefCode),
		"ref_msg":       notify.RefMsg,
	}

	camelCaseParams := map[string]string{
		"memberId":    notify.MerchantID,
		"orderId":     notify.OrderID,
		"platOrderId": notify.PlatOrderID,
		"amount":      formatLegacyGatewayAmount(notify.Amount),
		"cost":        strings.TrimSpace(notify.Cost),
		"status":      notify.Status,
		"refCode":     strconv.Itoa(notify.RefCode),
		"refMsg":      notify.RefMsg,
	}

	if !common.VerifySign(legacyParams, cfg.SecretKey, notify.Sign) && !common.VerifySign(camelCaseParams, cfg.SecretKey, notify.Sign) {
		return "FAIL", errors.New("invalid signature")
	}

	var order dtos.PaymentOrder
	if err := s.db.Where("order_id = ?", notify.OrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "FAIL", errors.New("order not found")
		}
		return "FAIL", err
	}

	if order.Status == 1 {
		return "OK", nil
	}

	if math.Abs(order.Amount-notify.Amount) > 0.000001 {
		return "FAIL", errors.New("order amount mismatch")
	}

	now := time.Now()
	switch notify.Status {
	case "success", "SUCCESS":
		exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
		if err != nil || exchangeRate <= 0 {
			exchangeRate = 16000
		}
		usdAmount := order.Amount / exchangeRate
		if err := s.processPaymentSuccess(ctx, &order, notify.PlatOrderID, now, usdAmount); err != nil {
			return "FAIL", err
		}
	case "failed", "FAILED", "fail", "FAIL":
		order.Status = 2
		order.RefCode = notify.RefCode
		order.RefMsg = notify.RefMsg
		order.UpdatedAt = now
		if err := s.db.Save(&order).Error; err != nil {
			return "FAIL", err
		}
	}

	return "OK", nil
}

func formatLegacyGatewayAmount(amount float64) string {
	if amount == math.Trunc(amount) {
		return strconv.FormatInt(int64(amount), 10)
	}
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

func normalizePaymentNotify(notify dtos.PaymentNotifyRequest) dtos.PaymentNotifyRequest {
	notify.MerchantID = firstNonEmpty(notify.MerchantNo, notify.MerchantID)
	notify.OrderID = firstNonEmpty(notify.MerchantOrderNo, notify.OrderID)
	notify.PlatOrderID = firstNonEmpty(notify.OrderNo, notify.PlatOrderID)
	notify.RefMsg = firstNonEmpty(notify.RefMsg, notify.RefMsgAlt)
	if notify.RefCode == 0 && notify.RefCodeAlt != 0 {
		notify.RefCode = notify.RefCodeAlt
	}
	if notify.RawAmount == "" && notify.Amount > 0 {
		notify.RawAmount = formatLegacyGatewayAmount(notify.Amount)
	}
	notify.Currency = normalizeCurrency(notify.Currency)
	notify.Code = strings.ToUpper(strings.TrimSpace(notify.Code))
	notify.Status = strings.TrimSpace(notify.Status)
	return notify
}

func (s *PaymentService) handleQuantixCorePaymentNotify(ctx context.Context, notify dtos.PaymentNotifyRequest) (string, error) {
	notify = normalizePaymentNotify(notify)
	if notify.OrderID == "" {
		return "FAIL", errors.New("QuantixCore order id is empty")
	}

	cfg := s.getConfig()
	if cfg == nil {
		return "FAIL", errors.New("payment config is not initialized")
	}

	quantixCfg := s.getQuantixCoreConfigByMerchant(cfg, notify.MerchantID, notify.Currency)
	if quantixCfg.SecretKey == "" {
		return "FAIL", errors.New("QuantixCore config is not initialized")
	}

	amountForSign := strings.TrimSpace(notify.RawAmount)
	if amountForSign == "" {
		amountForSign = formatLegacyGatewayAmount(notify.Amount)
	}

	signParams := map[string]string{
		"merchantNo":      notify.MerchantID,
		"merchantOrderNo": notify.OrderID,
		"orderNo":         notify.PlatOrderID,
		"amount":          amountForSign,
		"status":          notify.Status,
		"currency":        notify.Currency,
		"code":            notify.Code,
	}
	if !verifyQuantixCoreSign(signParams, quantixCfg.SecretKey, notify.Sign) {
		return "FAIL", errors.New("invalid QuantixCore signature")
	}

	var order dtos.PaymentOrder
	if err := s.db.Where("order_id = ?", notify.OrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "FAIL", errors.New("order not found")
		}
		return "FAIL", err
	}
	if order.Status == 1 {
		return "SUCCESS", nil
	}

	currency := normalizeCurrency(notify.Currency)
	if currency == "" {
		currency = s.getOrderCurrency(order.DstCode)
	}

	rawAmount := notify.RawAmount
	if rawAmount == "" {
		rawAmount = formatLegacyGatewayAmount(notify.Amount)
	}
	amountInCents, err := strconv.Atoi(rawAmount)
	if err != nil {
		return "FAIL", fmt.Errorf("invalid QuantixCore amount: %w", err)
	}
	paidAmount := amountFromMinorUnits(amountInCents, currency)
	if math.Abs(order.Amount-paidAmount) > 0.000001 {
		return "FAIL", errors.New("order amount mismatch")
	}

	now := time.Now()
	switch strings.ToUpper(notify.Status) {
	case "PAID":
		usdAmount := s.convertLocalAmountToUSD(paidAmount, currency)
		if err := s.processPaymentSuccess(ctx, &order, notify.PlatOrderID, now, usdAmount); err != nil {
			return "FAIL", err
		}
		return "SUCCESS", nil
	case "PAY_FAILED", "REFUND":
		order.Status = 2
		order.RefCode = notify.RefCode
		order.RefMsg = firstNonEmpty(notify.RefMsg, notify.Status)
		order.UpdatedAt = now
		if notify.PlatOrderID != "" {
			order.PlatOrderID = notify.PlatOrderID
		}
		if err := s.db.Save(&order).Error; err != nil {
			return "FAIL", err
		}
		return "SUCCESS", nil
	default:
		return "SUCCESS", nil
	}
}

func (s *PaymentService) processPaymentSuccess(ctx context.Context, order *dtos.PaymentOrder, platOrderID string, now time.Time, usdAmount float64) error {
	credited := false
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&dtos.PaymentOrder{}).
			Where("id = ? AND status = ?", order.ID, 0).
			Updates(map[string]interface{}{"status": 1, "plat_order_id": platOrderID, "paid_at": now, "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		credited = true

		var user dtos.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := beforeBalance + usdAmount

		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", usdAmount)).Error; err != nil {
			return err
		}

		currency := s.getOrderCurrency(order.DstCode)
		transaction := dtos.Transaction{
			UserID:        order.UserID,
			Type:          1,
			Amount:        usdAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   order.OrderID,
			Remark:        fmt.Sprintf("%s recharge (%s %.2f)", s.getPaymentTypeName(order.Type), currency, order.Amount),
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		if err := tx.Exec("UPDATE users SET total_deposit = total_deposit + ? WHERE id = ?", usdAmount, order.UserID).Error; err != nil {
			return err
		}

		if err := models.UpsertDailyUserStatsWithDB(tx, order.UserID, now.Format("2006-01-02"), 0, 0, usdAmount, 0, 0); err != nil {
			return err
		}

		if err := GetActivityService().GrantDepositWagerLockTx(tx, order.UserID, order.OrderID, usdAmount, now); err != nil {
			return err
		}

		GetFacebookPixelService().TrackPurchase(order.UserID, user.ParentID, order.Amount, currency, order.OrderID, EventContext{})
		return nil
	}); err != nil {
		return err
	}
	if !credited {
		return nil
	}

	if bonusErr := GetActivityService().ProcessNewUserRecharge(ctx, order.UserID, usdAmount); bonusErr != nil {
		fmt.Printf("new user recharge bonus processing failed for user %d: %v\n", order.UserID, bonusErr)
	}
	if _, vipErr := RecalculateAndPersistUserVIPLevel(ctx, order.UserID); vipErr != nil {
		fmt.Printf("vip level recalculation failed for user %d: %v\n", order.UserID, vipErr)
	}

	return nil
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, userID uint64, orderID string) (*dtos.PaymentStatusResponse, error) {
	var order dtos.PaymentOrder
	if err := s.db.Where("order_id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	resp := &dtos.PaymentStatusResponse{
		OrderID:     order.OrderID,
		PlatOrderID: order.PlatOrderID,
		Amount:      order.Amount,
		Status:      order.Status,
		Type:        order.Type,
		DstCode:     order.DstCode,
		CreatedAt:   order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if order.PaidAt != nil {
		paidAt := order.PaidAt.Format("2006-01-02 15:04:05")
		resp.PaidAt = &paidAt
	}
	return resp, nil
}

func (s *PaymentService) GetUserPayments(ctx context.Context, userID uint64, limit int) ([]dtos.WalletRecord, error) {
	var paymentOrders []dtos.PaymentOrder
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&paymentOrders).Error; err != nil {
		return nil, err
	}

	var withdrawOrders []dtos.WithdrawOrder
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&withdrawOrders).Error; err != nil {
		return nil, err
	}

	var rewardTransactions []dtos.Transaction
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND amount > ?", userID, 6, 0).
		Order("created_at DESC").
		Limit(limit).
		Find(&rewardTransactions).Error; err != nil {
		return nil, err
	}

	records := make([]dtos.WalletRecord, 0, len(paymentOrders)+len(withdrawOrders)+len(rewardTransactions))
	for _, order := range paymentOrders {
		records = append(records, dtos.WalletRecord{
			ID:         fmt.Sprintf("deposit-%d", order.ID),
			RecordType: "deposit",
			Title:      "Deposit",
			OrderID:    order.OrderID,
			Amount:     order.Amount,
			Currency:   walletRecordCurrencyFromCode(order.DstCode),
			Status:     order.Status,
			DstCode:    order.DstCode,
			CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, order := range withdrawOrders {
		records = append(records, dtos.WalletRecord{
			ID:         fmt.Sprintf("withdraw-%d", order.ID),
			RecordType: "withdraw",
			Title:      "Withdraw",
			OrderID:    order.OrderID,
			Amount:     order.Amount,
			Currency:   walletRecordCurrencyFromCode(order.DstCode),
			Status:     order.Status,
			DstCode:    order.DstCode,
			CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, transaction := range rewardTransactions {
		records = append(records, dtos.WalletRecord{
			ID:          fmt.Sprintf("bonus-%d", transaction.ID),
			RecordType:  "bonus",
			Title:       walletRecordBonusTitle(transaction.ReferenceID, transaction.Remark),
			ReferenceID: transaction.ReferenceID,
			Amount:      transaction.Amount,
			Currency:    "U",
			Status:      1,
			Remark:      transaction.Remark,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
		})
	}

	sort.Slice(records, func(i, j int) bool {
		left, leftErr := time.Parse(time.RFC3339, records[i].CreatedAt)
		right, rightErr := time.Parse(time.RFC3339, records[j].CreatedAt)
		if leftErr != nil || rightErr != nil {
			return records[i].CreatedAt > records[j].CreatedAt
		}
		return left.After(right)
	})

	if len(records) > limit {
		records = records[:limit]
	}

	return records, nil
}

func walletRecordCurrencyFromCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	switch {
	case strings.HasPrefix(normalized, "GCASH"),
		strings.HasPrefix(normalized, "BDO"),
		strings.HasPrefix(normalized, "BPI"),
		strings.HasPrefix(normalized, "METROBANK"),
		strings.HasPrefix(normalized, "LANDBANK"),
		strings.HasPrefix(normalized, "PNB"),
		strings.HasPrefix(normalized, "SECURITY"),
		strings.HasPrefix(normalized, "UNIONBANK"),
		strings.HasPrefix(normalized, "CHINABANK"),
		strings.HasPrefix(normalized, "RCBC"),
		strings.HasPrefix(normalized, "EASTWEST"):
		return "PHP"
	case strings.HasPrefix(normalized, "USDT"):
		return "USDT"
	default:
		return "IDR"
	}
}

func walletRecordBonusTitle(referenceID string, remark string) string {
	text := strings.ToLower(strings.TrimSpace(referenceID + " " + remark))
	switch {
	case strings.Contains(text, "add-desktop-insurance") || strings.Contains(text, "desktop insurance"):
		return "Add Desktop insurance compensation"
	case strings.Contains(text, "add_desktop") || strings.Contains(text, "desktop"):
		return "Add Desktop reward"
	case strings.Contains(text, "daily-weekly-challenge") || strings.Contains(text, "daily weekly challenge"):
		return "Activity task reward"
	case strings.Contains(text, "new-user-recharge") || strings.Contains(text, "new user recharge"):
		return "New user recharge reward"
	case strings.Contains(text, "recharge-rebate") || strings.Contains(text, "recharge rebate"):
		return "Recharge rebate reward"
	case strings.Contains(text, "vip-monthly-bonus") || strings.Contains(text, "vip monthly"):
		return "VIP monthly reward"
	case strings.Contains(text, "world_cup") || strings.Contains(text, "world cup"):
		return "World Cup reward"
	case strings.Contains(text, "mail"):
		return "Mail reward"
	case strings.Contains(text, "manual_reward") || strings.Contains(text, "manual reward"):
		return "Manual reward"
	case strings.Contains(text, "withdraw refund"):
		return "Withdraw refund"
	}

	if trimmedRemark := strings.TrimSpace(remark); trimmedRemark != "" && !walletRecordContainsCJK(trimmedRemark) {
		return remark
	}
	return "Reward"
}

func walletRecordContainsCJK(value string) bool {
	for _, char := range value {
		if char >= 0x3400 && char <= 0x9fff {
			return true
		}
	}
	return false
}

func (s *PaymentService) GetPaymentMethods(ctx context.Context) ([]dtos.PaymentMethod, error) {
	// Keep existing methods and expose Telegram Stars as a separate deposit tab/card.
	// It remains disabled until a Bot Token is configured, but must still be returned
	// so the frontend can display the dedicated Telegram payment interface.
	methodCodes := []string{"TG_STARS", "GCASH_QR", "GCASH_APP", "DANA", "USDT", "PAYPAL"}
	methods := make([]dtos.PaymentMethod, 0, len(methodCodes))
	for _, code := range methodCodes {
		cfg, ok := s.findDepositMethod(code)
		if !ok {
			continue
		}

		enabled := cfg.Enabled
		if code == "TG_STARS" {
			enabled = telegramBotToken() != ""
		}
		methods = append(methods, dtos.PaymentMethod{
			Code:        cfg.Code,
			Name:        cfg.Name,
			Type:        cfg.Type,
			MinAmount:   cfg.MinAmount,
			MaxAmount:   cfg.MaxAmount,
			Enabled:     enabled,
			Currency:    cfg.Currency,
			Channel:     cfg.Channel,
			Description: cfg.Description,
		})
	}
	return methods, nil
}

func (s *PaymentService) GetWithdrawMethods(ctx context.Context) ([]dtos.WithdrawMethod, error) {
	methods := make([]dtos.WithdrawMethod, 0, len(withdrawMethodConfigs))
	for _, cfg := range withdrawMethodConfigs {
		methods = append(methods, dtos.WithdrawMethod{
			Code:        cfg.Code,
			Name:        cfg.Name,
			Type:        cfg.Type,
			Currency:    cfg.Currency,
			Channel:     cfg.Channel,
			MinAmount:   cfg.MinAmount,
			MaxAmount:   cfg.MaxAmount,
			Enabled:     cfg.Enabled,
			Description: cfg.Description,
		})
	}
	return methods, nil
}

func (s *PaymentService) CreateTokenPayOrder(ctx context.Context, userID uint64, req dtos.CreateTokenPayOrderRequest) (*dtos.TokenPayCreateOrderResponse, error) {
	cfg := s.getConfig()
	if cfg == nil || cfg.TokenPay.BaseURL == "" || cfg.TokenPay.SecretKey == "" {
		return nil, errors.New("TokenPay config is not initialized")
	}

	chainMinAmount := map[string]float64{
		"ETH": 5,
		"TRX": 10,
	}

	minUSDT, ok := chainMinAmount[req.ChainType]
	if !ok {
		return nil, errors.New("invalid chain type, please choose ETH or TRX")
	}

	localCurrency := normalizeCurrency(req.Currency)
	if localCurrency != "USD" && localCurrency != "PHP" {
		localCurrency = "IDR"
	}
	exchangeRate, err := models.GetInstance().GetExchangeRate(localCurrency)
	if localCurrency == "USD" {
		exchangeRate = 1
		err = nil
	}
	if err != nil || exchangeRate <= 0 {
		if localCurrency == "PHP" {
			exchangeRate = 57
		} else {
			exchangeRate = 15500
		}
	}
	if localCurrency == "PHP" && req.Amount < 500 {
		return nil, errors.New("minimum USDT deposit is 500 PHP")
	}
	usdtAmount := req.Amount / exchangeRate
	if usdtAmount < minUSDT {
		return nil, fmt.Errorf("minimum deposit for this chain is %.0f USDT (about %.0f %s)", minUSDT, minUSDT*exchangeRate, localCurrency)
	}

	orderID := common.GenerateOrderID(userID)

	currency := "USDT_TRC20"
	if req.ChainType == "ETH" {
		currency = "EVM_ETH_USDT_ERC20"
	}

	orderUserKey := fmt.Sprintf("user_%d", userID)
	actualAmount := fmt.Sprintf("%.2f", usdtAmount)

	tokenPayReq := map[string]interface{}{
		"OutOrderId":   orderID,
		"OrderUserKey": orderUserKey,
		"ActualAmount": actualAmount,
		"Currency":     currency,
		"NotifyUrl":    cfg.TokenPay.NotifyURL,
		"RedirectUrl":  req.CallbackURL,
	}

	signature := helpers.GenerateTokenPaySignature(tokenPayReq, cfg.TokenPay.SecretKey)
	tokenPayReq["Signature"] = signature

	apiURL := strings.TrimSuffix(cfg.TokenPay.BaseURL, "/") + "/CreateOrder"
	jsonBody, _ := json.Marshal(tokenPayReq)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("TokenPay request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var tokenPayResp dtos.TokenPayCreateOrderResponse
	if err := json.Unmarshal(respData, &tokenPayResp); err != nil {
		return nil, fmt.Errorf("parse TokenPay response failed: %w", err)
	}
	if !tokenPayResp.Success {
		message := strings.TrimSpace(tokenPayResp.Message)
		if message == "" {
			message = fmt.Sprintf("gateway response (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respData)))
		}
		return nil, fmt.Errorf("TokenPay create order failed: %s", message)
	}

	tokenPayResp.OrderID = orderID
	tokenPayResp.PayAddress = tokenPayResp.Info.ToAddress
	tokenPayResp.PayAmount = tokenPayResp.Info.Amount
	tokenPayResp.QRCode = tokenPayResp.Info.QrCodeLink
	tokenPayResp.Status = "pending"
	if expireTime, err := time.Parse("2006-01-02 15:04:05", tokenPayResp.Info.ExpireTime); err == nil {
		tokenPayResp.ExpiredAt = expireTime.Unix()
	} else {
		tokenPayResp.ExpiredAt = time.Now().Add(30 * time.Minute).Unix()
	}

	order := dtos.PaymentOrder{
		UserID:      userID,
		OrderID:     orderID,
		Amount:      req.Amount,
		Type:        req.Type,
		DstCode:     "USDT_" + localCurrency + "_" + req.ChainType,
		Status:      0,
		ProductInfo: fmt.Sprintf("USDT recharge (%s)", req.ChainType),
	}

	if err := s.db.Create(&order).Error; err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	return &tokenPayResp, nil
}

func (s *PaymentService) HandleTokenPayNotify(ctx context.Context, notifyData map[string]interface{}) (string, error) {
	cfg := s.getConfig()
	if cfg == nil || cfg.TokenPay.SecretKey == "" {
		return "FAIL", errors.New("TokenPay config is not initialized")
	}

	signature, ok := notifyData["Signature"].(string)
	if !ok || signature == "" {
		return "FAIL", errors.New("missing signature")
	}
	if !helpers.VerifyTokenPayCallbackSignature(notifyData, signature, cfg.TokenPay.SecretKey) {
		return "FAIL", errors.New("invalid signature")
	}

	notify := dtos.TokenPayNotifyRequest{
		Id:                 getStringFromMap(notifyData, "Id"),
		OutOrderId:         getStringFromMap(notifyData, "OutOrderId"),
		OrderUserKey:       getStringFromMap(notifyData, "OrderUserKey"),
		BlockTransactionId: getStringFromMap(notifyData, "BlockTransactionId"),
		PayTime:            getStringFromMap(notifyData, "PayTime"),
		BlockchainName:     getStringFromMap(notifyData, "BlockchainName"),
		Currency:           getStringFromMap(notifyData, "Currency"),
		CurrencyName:       getStringFromMap(notifyData, "CurrencyName"),
		BaseCurrency:       getStringFromMap(notifyData, "BaseCurrency"),
		Amount:             getStringFromMap(notifyData, "Amount"),
		ActualAmount:       getStringFromMap(notifyData, "ActualAmount"),
		FromAddress:        getStringFromMap(notifyData, "FromAddress"),
		ToAddress:          getStringFromMap(notifyData, "ToAddress"),
		PassThroughInfo:    getStringFromMap(notifyData, "PassThroughInfo"),
		Signature:          signature,
	}
	if status, ok := notifyData["Status"].(float64); ok {
		notify.Status = int(status)
	} else if status, ok := notifyData["Status"].(int); ok {
		notify.Status = status
	}

	var order dtos.PaymentOrder
	if err := s.db.Where("order_id = ?", notify.OutOrderId).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "FAIL", errors.New("order not found")
		}
		return "FAIL", err
	}
	if order.Status == 1 {
		return "OK", nil
	}

	now := time.Now()
	switch notify.Status {
	case 1:
		cryptoAmount := s.resolveTokenPayPaidAmount(&order, notify)
		if cryptoAmount <= 0 {
			return "FAIL", errors.New("invalid payment amount")
		}
		if err := s.processTokenPaySuccess(ctx, &order, notify.BlockTransactionId, now, cryptoAmount); err != nil {
			return "FAIL", err
		}
	case 2:
		order.Status = 2
		order.UpdatedAt = now
		if err := s.db.Save(&order).Error; err != nil {
			return "FAIL", err
		}
	}

	return "OK", nil
}

func (s *PaymentService) resolveTokenPayPaidAmount(order *dtos.PaymentOrder, notify dtos.TokenPayNotifyRequest) float64 {
	expectedAmount := 0.0
	localCurrency := "IDR"
	if order != nil && strings.HasPrefix(strings.ToUpper(order.DstCode), "USDT_PHP_") {
		localCurrency = "PHP"
	}
	if order != nil && strings.HasPrefix(strings.ToUpper(order.DstCode), "USDT_USD_") {
		localCurrency = "USD"
	}
	if localCurrency == "USD" {
		return pickClosestPositiveAmount(order.Amount, notify.ActualAmount, notify.Amount)
	}
	exchangeRate, err := models.GetInstance().GetExchangeRate(localCurrency)
	if err != nil || exchangeRate <= 0 {
		if localCurrency == "PHP" {
			exchangeRate = 57
		} else {
			exchangeRate = 15500
		}
	}
	if order != nil && order.Amount > 0 {
		expectedAmount = order.Amount / exchangeRate
	}

	return pickClosestPositiveAmount(expectedAmount, notify.ActualAmount, notify.Amount)
}

func pickClosestPositiveAmount(expectedAmount float64, rawAmounts ...string) float64 {
	bestAmount := 0.0
	bestDistance := math.MaxFloat64

	for _, rawAmount := range rawAmounts {
		amount, err := strconv.ParseFloat(strings.TrimSpace(rawAmount), 64)
		if err != nil || amount <= 0 {
			continue
		}

		if expectedAmount <= 0 {
			return amount
		}

		distance := math.Abs(amount - expectedAmount)
		if distance < bestDistance {
			bestAmount = amount
			bestDistance = distance
		}
	}

	return bestAmount
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (s *PaymentService) processTokenPaySuccess(ctx context.Context, order *dtos.PaymentOrder, txHash string, now time.Time, usdtAmount float64) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		order.Status = 1
		order.PlatOrderID = txHash
		order.PaidAt = &now
		order.UpdatedAt = now
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		var user dtos.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := beforeBalance + usdtAmount

		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", usdtAmount)).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        order.UserID,
			Type:          1,
			Amount:        usdtAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   order.OrderID,
			Remark:        fmt.Sprintf("USDT recharge %s", order.DstCode),
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		if err := tx.Exec("UPDATE users SET total_deposit = total_deposit + ? WHERE id = ?", usdtAmount, order.UserID).Error; err != nil {
			return err
		}

		if err := models.UpsertDailyUserStatsWithDB(tx, order.UserID, now.Format("2006-01-02"), 0, 0, usdtAmount, 0, 0); err != nil {
			return err
		}

		if err := GetActivityService().GrantDepositWagerLockTx(tx, order.UserID, order.OrderID, usdtAmount, now); err != nil {
			return err
		}

		GetFacebookPixelService().TrackPurchase(order.UserID, user.ParentID, order.Amount, "IDR", order.OrderID, EventContext{})
		return nil
	}); err != nil {
		return err
	}

	if bonusErr := GetActivityService().ProcessNewUserRecharge(ctx, order.UserID, usdtAmount); bonusErr != nil {
		fmt.Printf("new user recharge bonus processing failed for user %d: %v\n", order.UserID, bonusErr)
	}
	if _, vipErr := RecalculateAndPersistUserVIPLevel(ctx, order.UserID); vipErr != nil {
		fmt.Printf("vip level recalculation failed for user %d: %v\n", order.UserID, vipErr)
	}

	return nil
}

func (s *PaymentService) CreateWithdrawOrder(ctx context.Context, userID uint64, req dtos.CreateWithdrawRequest) (*dtos.WithdrawOrder, error) {
	var methodCfg paymentMethodConfig
	if req.DstCode != "" {
		if foundCfg, ok := s.findWithdrawMethod(req.DstCode); ok {
			methodCfg = foundCfg
			if req.Currency == "" {
				req.Currency = foundCfg.Currency
			}
			if req.Channel == "" {
				req.Channel = foundCfg.Channel
			}
		}
	}
	if methodCfg.Code == "" {
		return nil, errors.New("unsupported withdraw method")
	}
	if !methodCfg.Enabled {
		return nil, errors.New("withdraw method is unavailable")
	}
	isTONWithdraw := strings.EqualFold(req.DstCode, "TON")
	minMethodAmount := methodCfg.MinAmount
	if isTONWithdraw {
		if configuredMin := tonWithdrawConfig().MinWithdrawAmountUSD; configuredMin > 0 {
			minMethodAmount = configuredMin
		}
	}
	if req.Amount < minMethodAmount || req.Amount > methodCfg.MaxAmount {
		return nil, fmt.Errorf("amount is out of range: %.0f - %.0f %s", minMethodAmount, methodCfg.MaxAmount, methodCfg.Currency)
	}

	balanceDeductionUSD := s.convertLocalAmountToUSD(req.Amount, req.Currency)
	minimumWithdrawUSD := minWithdrawAmountUSD
	if isTONWithdraw && tonWithdrawConfig().MinWithdrawAmountUSD > 0 {
		minimumWithdrawUSD = tonWithdrawConfig().MinWithdrawAmountUSD
	}
	if balanceDeductionUSD < minimumWithdrawUSD {
		return nil, fmt.Errorf("minimum withdraw amount is %.0f USDT", minimumWithdrawUSD)
	}

	orderID := common.GenerateWithdrawOrderID(userID)

	var withdrawOrder dtos.WithdrawOrder
	autoSubmit := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		requireConditions := !isTONWithdraw || tonWithdrawConfig().WithdrawConditionsEnabled
		if requireConditions && user.TotalDeposit <= 0 {
			return fmt.Errorf("withdraw requires matching deposit %.2f USDT", balanceDeductionUSD)
		}

		if requireConditions {
			now := time.Now()
			dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			dayEnd := dayStart.AddDate(0, 0, 1)
			var todayWithdrawCount int64
			if err := tx.Model(&dtos.WithdrawOrder{}).
				Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, dayStart, dayEnd).
				Count(&todayWithdrawCount).Error; err != nil {
				return err
			}
			if todayWithdrawCount >= dailyWithdrawLimit {
				return fmt.Errorf("daily withdraw limit reached: %d times", dailyWithdrawLimit)
			}

			remainingWager, withdrawableBalance, err := GetActivityService().GetWithdrawWagerStatus(ctx, userID, user.Balance)
			if err != nil {
				return err
			}
			if balanceDeductionUSD > withdrawableBalance {
				if remainingWager > 0 {
					return fmt.Errorf("reward funds still require %.2f wager turnover before withdrawal", remainingWager)
				}
				return errors.New("insufficient withdrawable balance")
			}
		}

		if user.Balance < balanceDeductionUSD {
			return errors.New("insufficient balance")
		}
		// TON uses the server-owned hot wallet for automatic payouts. Other
		// methods still require their configured gateway channel.
		autoSubmit = shouldAutoSubmitWithdraw(balanceDeductionUSD) &&
			(methodCfg.Channel != "manual" || strings.EqualFold(req.DstCode, "TON"))

		updateResult := tx.Model(&user).
			Where("balance >= ?", balanceDeductionUSD).
			Update("balance", gorm.Expr("balance - ?", balanceDeductionUSD))
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			return errors.New("insufficient balance")
		}

		beforeBalance := user.Balance
		afterBalance := beforeBalance - balanceDeductionUSD

		if err := tx.Model(&user).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		orderStatus := withdrawOrderStatusPending
		orderRefMsg := "pending admin review"
		if autoSubmit {
			orderStatus = withdrawOrderStatusProcessing
			orderRefMsg = "auto submitting to gateway"
		}

		withdrawOrder = dtos.WithdrawOrder{
			UserID:      userID,
			OrderID:     orderID,
			Amount:      req.Amount,
			Cost:        balanceDeductionUSD,
			Type:        req.Type,
			DstCode:     req.DstCode,
			Account:     req.Account,
			AccountName: req.AccountName,
			Phone:       req.Phone,
			Email:       req.Email,
			Address:     req.Address,
			Status:      orderStatus,
			RefMsg:      orderRefMsg,
		}
		if err := tx.Create(&withdrawOrder).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          2,
			Amount:        -balanceDeductionUSD,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   orderID,
			Remark:        fmt.Sprintf("%s withdraw", s.getPaymentTypeName(req.Type)),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	if !autoSubmit {
		return &withdrawOrder, nil
	}

	platOrderID, refMsg, submitErr := s.submitApprovedWithdrawToGateway(ctx, &withdrawOrder)
	if submitErr != nil {
		failureMsg := truncateWithdrawRefMsg("gateway submit failed: " + submitErr.Error())
		if refundErr := s.processWithdrawFailure(ctx, &withdrawOrder, "", 2, failureMsg, time.Now()); refundErr != nil {
			return nil, fmt.Errorf("auto withdraw submit failed: %v; refund failed: %w", submitErr, refundErr)
		}
		return nil, fmt.Errorf("auto withdraw submit failed and funds were refunded: %w", submitErr)
	}

	updateNow := time.Now()
	updateResult := s.db.WithContext(ctx).Model(&dtos.WithdrawOrder{}).
		Where("order_id = ? AND status = ?", orderID, withdrawOrderStatusProcessing).
		Updates(map[string]interface{}{
			"plat_order_id": platOrderID,
			"ref_msg":       truncateWithdrawRefMsg(firstNonEmpty(refMsg, "submitted to gateway")),
			"updated_at":    updateNow,
		})
	if updateResult.Error != nil {
		return nil, updateResult.Error
	}
	if err := s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&withdrawOrder).Error; err != nil {
		return nil, err
	}
	if strings.EqualFold(withdrawOrder.DstCode, "TON") && withdrawOrder.Status == withdrawOrderStatusProcessing {
		if err := s.processWithdrawSuccess(ctx, &withdrawOrder, platOrderID, 1, firstNonEmpty(refMsg, "TON payout confirmed"), time.Now()); err != nil {
			return nil, err
		}
		if err := s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&withdrawOrder).Error; err != nil {
			return nil, err
		}
	}

	return &withdrawOrder, nil
}

func (s *PaymentService) ApproveWithdrawOrder(ctx context.Context, req dtos.AdminApproveWithdrawRequest) (*dtos.WithdrawOrder, error) {
	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		return nil, errors.New("order id is required")
	}

	var order dtos.WithdrawOrder
	if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("withdraw order not found")
		}
		return nil, err
	}
	if order.Status != withdrawOrderStatusPending {
		return nil, fmt.Errorf("withdraw order status %d cannot be approved", order.Status)
	}

	now := time.Now()
	claimResult := s.db.WithContext(ctx).Model(&dtos.WithdrawOrder{}).
		Where("order_id = ? AND status = ?", orderID, withdrawOrderStatusPending).
		Updates(map[string]interface{}{
			"status":        withdrawOrderStatusProcessing,
			"reviewed_by":   req.ReviewerID,
			"reviewer":      strings.TrimSpace(req.Reviewer),
			"review_remark": strings.TrimSpace(req.Remark),
			"reviewed_at":   now,
			"ref_msg":       "submitting to gateway",
			"updated_at":    now,
		})
	if claimResult.Error != nil {
		return nil, claimResult.Error
	}
	if claimResult.RowsAffected == 0 {
		return nil, errors.New("withdraw order is being processed or already reviewed")
	}

	platOrderID, refMsg, err := s.submitApprovedWithdrawToGateway(ctx, &order)
	if err != nil {
		revertNow := time.Now()
		_ = s.db.WithContext(ctx).Model(&dtos.WithdrawOrder{}).
			Where("order_id = ? AND status = ?", orderID, withdrawOrderStatusProcessing).
			Updates(map[string]interface{}{
				"status":     withdrawOrderStatusPending,
				"ref_msg":    truncateWithdrawRefMsg("gateway submit failed: " + err.Error()),
				"updated_at": revertNow,
			}).Error
		return nil, err
	}

	updateNow := time.Now()
	updateResult := s.db.WithContext(ctx).Model(&dtos.WithdrawOrder{}).
		Where("order_id = ? AND status = ?", orderID, withdrawOrderStatusProcessing).
		Updates(map[string]interface{}{
			"plat_order_id": platOrderID,
			"ref_msg":       truncateWithdrawRefMsg(firstNonEmpty(refMsg, "submitted to gateway")),
			"updated_at":    updateNow,
		})
	if updateResult.Error != nil {
		return nil, updateResult.Error
	}

	if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	if strings.EqualFold(order.DstCode, "TON") && order.Status == withdrawOrderStatusProcessing {
		if err := s.processWithdrawSuccess(ctx, &order, platOrderID, 1, firstNonEmpty(refMsg, "TON payout confirmed"), time.Now()); err != nil {
			return nil, err
		}
		if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
			return nil, err
		}
	}

	return &order, nil
}

func (s *PaymentService) submitApprovedWithdrawToGateway(ctx context.Context, order *dtos.WithdrawOrder) (string, string, error) {
	// TON orders must never fall through to legacy/Quantix gateways, regardless
	// of the configured display type or channel.
	if strings.EqualFold(strings.TrimSpace(order.DstCode), "TON") {
		return s.submitTONWithdraw(ctx, order)
	}
	methodCfg, ok := s.findWithdrawMethod(order.DstCode)
	if !ok {
		return "", "", errors.New("unsupported withdraw method")
	}

	currency := methodCfg.Currency
	if currency == "" {
		currency = s.getWithdrawCurrency(order.DstCode)
	}

	if strings.EqualFold(methodCfg.Channel, "quantixcore") {
		return s.submitQuantixCoreWithdraw(ctx, order, methodCfg, currency)
	}
	return s.submitLegacyGatewayWithdraw(ctx, order)
}

func (s *PaymentService) submitLegacyGatewayWithdraw(ctx context.Context, order *dtos.WithdrawOrder) (string, string, error) {
	cfg := s.getConfig()
	if cfg == nil {
		return "", "", errors.New("payment config is not initialized")
	}
	if cfg.MerchantID == "" || cfg.SecretKey == "" {
		return "", "", errors.New("payment config is not initialized")
	}

	params := buildLegacyGatewayWithdrawParams(order, cfg, s.getWithdrawFeeConfig(ctx))
	params["sign"] = common.GenerateSign(params, cfg.SecretKey)

	apiURL := strings.TrimSuffix(cfg.BaseURL, "/") + "/payment_dfpay_add.html"
	respData, err := s.postMultipartFormData(ctx, apiURL, params)
	if err != nil {
		return "", "", fmt.Errorf("withdraw gateway request failed: %w", err)
	}

	var gatewayResp struct {
		Status      string `json:"status"`
		Msg         string `json:"msg"`
		ErrCode     int    `json:"errCode"`
		PlatOrderID string `json:"platOrderId"`
		OrderID     string `json:"orderId"`
	}

	if err := json.Unmarshal(respData, &gatewayResp); err != nil {
		return "", "", fmt.Errorf("parse gateway response failed: %w", err)
	}
	if gatewayResp.Status != "success" {
		return "", "", fmt.Errorf("withdraw gateway error: %s (errCode: %d)", gatewayResp.Msg, gatewayResp.ErrCode)
	}

	return gatewayResp.PlatOrderID, gatewayResp.Msg, nil
}

func buildLegacyGatewayWithdrawParams(order *dtos.WithdrawOrder, cfg *helpers.PaymentConfig, feeConfig withdrawFeeConfigValues) map[string]string {
	gatewayAmount := withdrawGatewayAmount(order.Amount, "IDR", feeConfig)
	return map[string]string{
		"memberId":  strings.TrimSpace(cfg.MerchantID),
		"orderId":   strings.TrimSpace(order.OrderID),
		"amount":    fmt.Sprintf("%.0f", gatewayAmount),
		"type":      normalizeLegacyWithdrawType(order.Type),
		"dstCode":   strings.TrimSpace(order.DstCode),
		"account":   strings.TrimSpace(order.Account),
		"name":      strings.TrimSpace(order.AccountName),
		"phone":     strings.TrimSpace(order.Phone),
		"email":     strings.TrimSpace(order.Email),
		"address":   strings.TrimSpace(order.Address),
		"notifyUrl": getWithdrawNotifyURL(cfg.NotifyURL, cfg.NotifyURL),
	}
}

func buildQuantixCoreWithdrawPayload(
	order *dtos.WithdrawOrder,
	methodCfg paymentMethodConfig,
	currency string,
	quantixCfg helpers.QuantixCoreRegionConfig,
	cfg *helpers.PaymentConfig,
	feeConfig withdrawFeeConfigValues,
) (map[string]string, error) {
	currency = normalizeCurrency(currency)
	amountInMinorUnits := amountToMinorUnits(withdrawGatewayAmount(order.Amount, currency, feeConfig), currency)
	if amountInMinorUnits <= 0 {
		return nil, errors.New("invalid amount")
	}

	accountType := resolveQuantixCoreAccountType(order, methodCfg)
	if accountType == "" {
		return nil, errors.New("QuantixCore account type is required")
	}

	payload := map[string]string{
		"merchantNo":      strings.TrimSpace(quantixCfg.MerchantID),
		"merchantOrderNo": strings.TrimSpace(order.OrderID),
		"uid":             strconv.FormatUint(order.UserID, 10),
		"amount":          strconv.Itoa(amountInMinorUnits),
		"currency":        currency,
		"accountType":     accountType,
		"account":         strings.TrimSpace(order.Account),
		"accountName":     strings.TrimSpace(order.AccountName),
		"phone":           strings.TrimSpace(order.Phone),
		"email":           strings.TrimSpace(order.Email),
		"callback":        getWithdrawNotifyURL(quantixCfg.NotifyURL, cfg.NotifyURL),
	}

	if bankCode := resolveQuantixCoreBankCode(order, methodCfg); bankCode != "" {
		payload["bankCode"] = bankCode
	}

	return payload, nil
}

func resolveQuantixCoreAccountType(order *dtos.WithdrawOrder, methodCfg paymentMethodConfig) string {
	if accountType := strings.TrimSpace(methodCfg.AccountType); accountType != "" {
		return accountType
	}

	normalizedCode := strings.ToUpper(strings.TrimSpace(order.DstCode))
	if accountType := quantixCoreWalletAccountTypes[normalizedCode]; accountType != "" {
		return accountType
	}

	if strings.EqualFold(order.Type, "bankcard") {
		return "PERSONAL_BANK"
	}

	return ""
}

func resolveQuantixCoreBankCode(order *dtos.WithdrawOrder, methodCfg paymentMethodConfig) string {
	if !strings.EqualFold(order.Type, "bankcard") {
		return ""
	}

	normalizedCode := strings.ToUpper(strings.TrimSpace(order.DstCode))
	if _, isWallet := quantixCoreWalletAccountTypes[normalizedCode]; isWallet {
		return ""
	}

	return normalizedCode
}

func normalizeLegacyWithdrawType(withdrawType string) string {
	switch strings.ToLower(strings.TrimSpace(withdrawType)) {
	case "bank", "bankcard":
		return "bankcard"
	case "wallet", "ewallet":
		return "ewallet"
	default:
		return strings.TrimSpace(withdrawType)
	}
}

func (s *PaymentService) submitQuantixCoreWithdraw(ctx context.Context, order *dtos.WithdrawOrder, methodCfg paymentMethodConfig, currency string) (string, string, error) {
	cfg := s.getConfig()
	if cfg == nil {
		return "", "", errors.New("payment config is not initialized")
	}

	quantixCfg := s.getQuantixCoreConfig(cfg, currency)
	if quantixCfg.MerchantID == "" || quantixCfg.SecretKey == "" {
		return "", "", errors.New("QuantixCore config is not initialized")
	}

	payload, err := buildQuantixCoreWithdrawPayload(order, methodCfg, currency, quantixCfg, cfg, s.getWithdrawFeeConfig(ctx))
	if err != nil {
		return "", "", err
	}
	payload["sign"] = generateQuantixCoreSign(payload, quantixCfg.SecretKey)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("marshal QuantixCore withdraw request failed: %w", err)
	}

	apiURL := strings.TrimSuffix(s.getQuantixCoreBaseURL(cfg, quantixCfg), "/") + "/api/open/flex/order/payment/add"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("create QuantixCore withdraw request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("QuantixCore withdraw request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read QuantixCore withdraw response failed: %w", err)
	}

	var gatewayResp struct {
		Success   bool   `json:"success"`
		ErrorCode string `json:"errorCode"`
		Message   string `json:"message"`
		Data      struct {
			OrderNo string `json:"orderNo"`
			Status  string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respData, &gatewayResp); err != nil {
		return "", "", fmt.Errorf("parse QuantixCore withdraw response failed: %w", err)
	}
	if !gatewayResp.Success {
		if gatewayResp.Message == "" {
			gatewayResp.Message = "create withdraw order failed"
		}
		if gatewayResp.ErrorCode != "" {
			return "", "", fmt.Errorf("QuantixCore withdraw error: %s (%s)", gatewayResp.Message, gatewayResp.ErrorCode)
		}
		return "", "", fmt.Errorf("QuantixCore withdraw error: %s", gatewayResp.Message)
	}

	return gatewayResp.Data.OrderNo, gatewayResp.Message, nil
}

func (s *PaymentService) HandleWithdrawNotify(ctx context.Context, notify dtos.WithdrawNotifyRequest) (string, error) {
	cfg := s.getConfig()
	if cfg == nil {
		return "FAIL", errors.New("payment config is not initialized")
	}

	isQuantixCoreNotify := notify.MerchantNo != "" || notify.MerchantOrderNo != "" || notify.OrderNo != "" || notify.Code != "" || notify.ErrorMsg != "" || notify.ErrorMsgAlt != ""
	successAck := "SUCCESS"
	if isQuantixCoreNotify {
		successAck = "success"
	}
	merchantID := firstNonEmpty(notify.MerchantNo, notify.MerchantID)
	orderID := firstNonEmpty(notify.MerchantOrderNo, notify.OrderID)
	platOrderID := firstNonEmpty(notify.OrderNo, notify.PlatOrderID)
	status := strings.TrimSpace(notify.Status)
	refCodeStr := firstNonEmpty(notify.RefCode, notify.RefCodeAlt)
	refMsg := firstNonEmpty(notify.RefMsg, notify.RefMsgAlt, notify.ErrorMsg, notify.ErrorMsgAlt)
	amountRaw := strings.TrimSpace(notify.Amount)
	currency := normalizeCurrency(notify.Currency)

	if orderID == "" {
		return "FAIL", errors.New("withdraw order id is empty")
	}

	if isQuantixCoreNotify {
		quantixCfg := s.getQuantixCoreConfigByMerchant(cfg, merchantID, currency)
		if quantixCfg.SecretKey == "" {
			return "FAIL", errors.New("QuantixCore config is not initialized")
		}

		signParams := map[string]string{
			"merchantNo":      merchantID,
			"merchantOrderNo": orderID,
			"orderNo":         platOrderID,
			"amount":          amountRaw,
			"status":          status,
			"currency":        currency,
			"code":            notify.Code,
			"errorMsg":        firstNonEmpty(notify.ErrorMsg, notify.ErrorMsgAlt),
		}
		if !verifyQuantixCoreSign(signParams, quantixCfg.SecretKey, notify.Sign) {
			return "FAIL", errors.New("invalid QuantixCore withdraw signature")
		}
	} else {
		signParams := map[string]string{
			"memberId":    merchantID,
			"orderId":     orderID,
			"platOrderId": platOrderID,
			"amount":      amountRaw,
			"cost":        strings.TrimSpace(notify.Cost),
			"status":      status,
			"refCode":     refCodeStr,
			"refMsg":      refMsg,
		}
		if !common.VerifySign(signParams, cfg.SecretKey, notify.Sign) {
			return "FAIL", errors.New("invalid withdraw signature")
		}
	}

	var order dtos.WithdrawOrder
	if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "FAIL", errors.New("withdraw order not found")
		}
		return "FAIL", err
	}

	if order.Status == withdrawOrderStatusSuccess || order.Status == withdrawOrderStatusFailed {
		return successAck, nil
	}

	if currency == "" {
		currency = s.getWithdrawCurrency(order.DstCode)
	}
	expectedGatewayAmount := withdrawGatewayAmount(order.Amount, currency, s.getWithdrawFeeConfig(ctx))

	var notifyAmount float64
	if amountRaw != "" {
		if isQuantixCoreNotify {
			amountMinorUnits, err := strconv.Atoi(amountRaw)
			if err != nil {
				return "FAIL", fmt.Errorf("invalid withdraw amount: %w", err)
			}
			notifyAmount = amountFromMinorUnits(amountMinorUnits, currency)
		} else {
			parsedAmount, err := strconv.ParseFloat(amountRaw, 64)
			if err != nil {
				return "FAIL", fmt.Errorf("invalid withdraw amount: %w", err)
			}
			notifyAmount = parsedAmount
		}
	}

	if notifyAmount > 0 && math.Abs(expectedGatewayAmount-notifyAmount) > 0.000001 {
		return "FAIL", errors.New("withdraw amount mismatch")
	}

	refCode, _ := strconv.Atoi(refCodeStr)
	statusUpper := strings.ToUpper(status)
	success := refCode == 1 || statusUpper == "SUCCESS" || statusUpper == "PAID" || statusUpper == "COMPLETED" || statusUpper == "1"
	failed := refCode == 2 || statusUpper == "FAIL" || statusUpper == "FAILED" || statusUpper == "ERROR" || statusUpper == "PAY_FAILED" || statusUpper == "REFUND" || statusUpper == "2"

	now := time.Now()
	switch {
	case success:
		if err := s.processWithdrawSuccess(ctx, &order, platOrderID, refCode, refMsg, now); err != nil {
			return "FAIL", err
		}
	case failed:
		if err := s.processWithdrawFailure(ctx, &order, platOrderID, refCode, refMsg, now); err != nil {
			return "FAIL", err
		}
	}

	return successAck, nil
}

func (s *PaymentService) processWithdrawSuccess(ctx context.Context, order *dtos.WithdrawOrder, platOrderID string, refCode int, refMsg string, now time.Time) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		order.Status = withdrawOrderStatusSuccess
		order.RefCode = refCode
		order.RefMsg = truncateWithdrawRefMsg(refMsg)
		order.UpdatedAt = now
		order.CompletedAt = &now
		if platOrderID != "" {
			order.PlatOrderID = platOrderID
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		withdrawCost := order.Cost
		if withdrawCost <= 0 {
			withdrawCost = s.convertLocalAmountToUSD(order.Amount, s.getWithdrawCurrency(order.DstCode))
		}
		if withdrawCost > 0 {
			if err := tx.Exec("UPDATE users SET total_withdraw = total_withdraw + ? WHERE id = ?", withdrawCost, order.UserID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *PaymentService) processWithdrawFailure(ctx context.Context, order *dtos.WithdrawOrder, platOrderID string, refCode int, refMsg string, now time.Time) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		order.Status = withdrawOrderStatusFailed
		order.RefCode = refCode
		order.RefMsg = truncateWithdrawRefMsg(refMsg)
		order.UpdatedAt = now
		order.CompletedAt = &now
		if platOrderID != "" {
			order.PlatOrderID = platOrderID
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		refundAmount := order.Cost
		if refundAmount <= 0 {
			refundAmount = s.convertLocalAmountToUSD(order.Amount, s.getWithdrawCurrency(order.DstCode))
		}
		if refundAmount <= 0 {
			return nil
		}

		var user dtos.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			return err
		}

		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", refundAmount)).Error; err != nil {
			return err
		}

		refundTransaction := dtos.Transaction{
			UserID:        order.UserID,
			Type:          6,
			Amount:        refundAmount,
			BeforeBalance: user.Balance,
			AfterBalance:  user.Balance + refundAmount,
			ReferenceID:   order.OrderID,
			Remark:        "Withdraw refund",
			CreatedAt:     now,
		}
		if err := tx.Create(&refundTransaction).Error; err != nil {
			return err
		}

		return nil
	})
}

func truncateWithdrawRefMsg(message string) string {
	message = strings.TrimSpace(message)
	if len(message) <= 255 {
		return message
	}
	return message[:255]
}

func (s *PaymentService) postFormData(ctx context.Context, apiURL string, params map[string]string) ([]byte, error) {
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *PaymentService) postMultipartFormData(ctx context.Context, apiURL string, params map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for k, v := range params {
		if err := writer.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *PaymentService) getPaymentTypeName(paymentType string) string {
	switch paymentType {
	case "bank":
		return "Bank transfer"
	case "ewallet":
		return "E-wallet"
	case "qris":
		return "QRIS"
	case "bankcard":
		return "Bank card"
	case "voucher":
		return "Voucher"
	default:
		return paymentType
	}
}

func (s *PaymentService) getOrderCurrency(dstCode string) string {
	if methodCfg, ok := s.findDepositMethod(dstCode); ok && methodCfg.Currency != "" {
		return methodCfg.Currency
	}
	if strings.HasPrefix(dstCode, "USDT") {
		return "USDT"
	}
	return "IDR"
}

func (s *PaymentService) getWithdrawCurrency(dstCode string) string {
	if methodCfg, ok := s.findWithdrawMethod(dstCode); ok && methodCfg.Currency != "" {
		return methodCfg.Currency
	}
	return "IDR"
}

func (s *PaymentService) getQuantixCoreBaseURL(cfg *helpers.PaymentConfig, quantixCfg helpers.QuantixCoreRegionConfig) string {
	baseURL := firstNonEmpty(quantixCfg.BaseURL, cfg.QuantixCore.BaseURL, cfg.BaseURL)
	if baseURL == "" || !strings.Contains(strings.ToLower(baseURL), "quantixcore") {
		return "https://api.quantixcore.com"
	}
	return baseURL
}

func getWithdrawNotifyURL(primaryNotifyURL, fallbackNotifyURL string) string {
	notifyURL := firstNonEmpty(primaryNotifyURL, fallbackNotifyURL)
	if notifyURL == "" {
		return ""
	}
	if strings.Contains(notifyURL, "/api/payments/notify") {
		return strings.Replace(notifyURL, "/api/payments/notify", "/api/withdraw/notify", 1)
	}
	if strings.Contains(notifyURL, "/payments/notify") {
		return strings.Replace(notifyURL, "/payments/notify", "/withdraw/notify", 1)
	}
	return strings.TrimSuffix(notifyURL, "/") + "/withdraw/notify"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func generateQuantixCoreSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || strings.TrimSpace(params[k]) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, params[key]))
	}

	signStr := strings.Join(parts, "&") + "&secret=" + secret
	sum := md5.Sum([]byte(signStr))
	return hex.EncodeToString(sum[:])
}

func verifyQuantixCoreSign(params map[string]string, secret, sign string) bool {
	return strings.EqualFold(generateQuantixCoreSign(params, secret), sign)
}

func (s *PaymentService) GetVoucherConfig() *dtos.VoucherConfigResponse {
	instance := helpers.GetCfgInstance()
	purchaseURL := ""
	if instance != nil && instance.Conf != nil {
		purchaseURL = instance.Conf.VocherURL
	}
	return &dtos.VoucherConfigResponse{PurchaseURL: purchaseURL}
}

func (s *PaymentService) RedeemVoucher(ctx context.Context, userID uint64, req dtos.VoucherRedeemRequest) (*dtos.VoucherRedeemResponse, error) {
	if req.VoucherCode == "" {
		return nil, errors.New("please enter voucher code")
	}

	exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
	if err != nil || exchangeRate <= 0 {
		exchangeRate = 16000
	}
	usdAmount := req.Amount / exchangeRate

	orderID := common.GenerateOrderID(userID)
	now := time.Now()

	var newBalance float64
	var inviterUserID uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		order := dtos.PaymentOrder{
			UserID:      userID,
			OrderID:     orderID,
			PlatOrderID: "VOUCHER-" + req.VoucherCode[:min(8, len(req.VoucherCode))],
			Amount:      req.Amount,
			Type:        "voucher",
			DstCode:     "199VOUCHER",
			Status:      1,
			ProductInfo: "199 Voucher recharge",
			PayURL:      "",
			PaidAt:      &now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		var user dtos.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		inviterUserID = user.ParentID
		beforeBalance := user.Balance
		newBalance = beforeBalance + usdAmount

		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          1,
			Amount:        usdAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  newBalance,
			ReferenceID:   order.OrderID,
			Remark:        fmt.Sprintf("199 Voucher recharge (IDR %.0f)", req.Amount),
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		if err := tx.Exec("UPDATE users SET total_deposit = total_deposit + ? WHERE id = ?", usdAmount, userID).Error; err != nil {
			return err
		}

		if err := models.UpsertDailyUserStatsWithDB(tx, userID, now.Format("2006-01-02"), 0, 0, usdAmount, 0, 0); err != nil {
			return err
		}

		if err := GetActivityService().GrantDepositWagerLockTx(tx, userID, order.OrderID, usdAmount, now); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("redeem failed: %w", err)
	}

	if bonusErr := GetActivityService().ProcessNewUserRecharge(ctx, userID, usdAmount); bonusErr != nil {
		fmt.Printf("new user recharge bonus processing failed for user %d: %v\n", userID, bonusErr)
	}
	if _, vipErr := RecalculateAndPersistUserVIPLevel(ctx, userID); vipErr != nil {
		fmt.Printf("vip level recalculation failed for user %d: %v\n", userID, vipErr)
	}

	GetFacebookPixelService().TrackPurchase(userID, inviterUserID, req.Amount, "IDR", orderID, EventContext{})

	return &dtos.VoucherRedeemResponse{
		OrderID:   orderID,
		Amount:    req.Amount,
		UsdAmount: usdAmount,
		Balance:   newBalance,
		Status:    1,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	paymentService     *PaymentService
	paymentServiceOnce sync.Once
)

func GetPaymentService() *PaymentService {
	paymentServiceOnce.Do(func() {
		db := models.GetInstance()
		if db == nil {
			panic("models.GetInstance() returned nil - database not initialized")
		}
		paymentService = NewPaymentService(db.DbInstance)
	})
	return paymentService
}

func SetPaymentServiceForTest(service *PaymentService) {
	paymentService = service
}

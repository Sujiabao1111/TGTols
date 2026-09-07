package example

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WithdrawAuditService struct{}

const (
	withdrawStatusPendingReview = 0
	withdrawStatusSuccess       = 1
	withdrawStatusRejected      = 2
	withdrawStatusProcessing    = 3
)

type withdrawMethodConfig struct {
	Code        string
	Type        string
	Currency    string
	Channel     string
	AccountType string
}

var withdrawMethodConfigs = []withdrawMethodConfig{
	{Code: "GCASH_QR", Type: "ewallet", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH"},
	{Code: "GCASH_APP", Type: "ewallet", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH"},
	{Code: "GCASH_H5", Type: "ewallet", Currency: "PHP", Channel: "quantixcore", AccountType: "GCASH"},
	{Code: "OVO", Type: "ewallet", Currency: "IDR", Channel: "gateway", AccountType: "OVO"},
	{Code: "DANA", Type: "ewallet", Currency: "IDR", Channel: "gateway", AccountType: "DANA"},
	{Code: "GOPAY", Type: "ewallet", Currency: "IDR", Channel: "gateway", AccountType: "GOPAY"},
	{Code: "SHOPEEPAY", Type: "ewallet", Currency: "IDR", Channel: "gateway", AccountType: "SHOPEEPAY"},
	{Code: "LINKAJA", Type: "ewallet", Currency: "IDR", Channel: "gateway", AccountType: "LINKAJA"},
}

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

type withdrawOrderRow struct {
	ID           uint64     `gorm:"column:id"`
	UserID       uint64     `gorm:"column:user_id"`
	OrderID      string     `gorm:"column:order_id"`
	PlatOrderID  string     `gorm:"column:plat_order_id"`
	Amount       float64    `gorm:"column:amount"`
	Cost         float64    `gorm:"column:cost"`
	Type         string     `gorm:"column:type"`
	DstCode      string     `gorm:"column:dst_code"`
	Account      string     `gorm:"column:account"`
	AccountName  string     `gorm:"column:account_name"`
	Phone        string     `gorm:"column:phone"`
	Email        string     `gorm:"column:email"`
	Address      string     `gorm:"column:address"`
	Status       int        `gorm:"column:status"`
	RefCode      int        `gorm:"column:ref_code"`
	RefMsg       string     `gorm:"column:ref_msg"`
	Reviewer     string     `gorm:"column:reviewer"`
	ReviewRemark string     `gorm:"column:review_remark"`
	ReviewedAt   *time.Time `gorm:"column:reviewed_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
}

type withdrawUserRow struct {
	ID       uint64  `gorm:"column:id"`
	Username string  `gorm:"column:username"`
	Balance  float64 `gorm:"column:balance"`
}

func (withdrawAuditService *WithdrawAuditService) GetWithdrawAuditList(ctx context.Context, info exampleReq.WithdrawAuditSearch) (list []model.WithdrawAuditRecord, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Table("withdraw_orders wo").
		Select(`
			wo.id,
			wo.user_id,
			u.username,
			wo.order_id,
			wo.plat_order_id,
			wo.amount,
			wo.cost,
			wo.type,
			wo.dst_code,
			wo.account,
			wo.account_name,
			wo.phone,
			wo.email,
			wo.status,
			wo.ref_code,
			wo.ref_msg,
			wo.reviewer,
			wo.review_remark,
			wo.reviewed_at,
			wo.completed_at,
			wo.created_at
		`).
		Joins("LEFT JOIN users u ON u.id = wo.user_id")

	if info.UserID != nil && *info.UserID > 0 {
		db = db.Where("wo.user_id = ?", *info.UserID)
	}
	if orderID := strings.TrimSpace(info.OrderID); orderID != "" {
		db = db.Where("wo.order_id LIKE ?", "%"+orderID+"%")
	}
	if info.Status != nil && *info.Status >= 0 {
		db = db.Where("wo.status = ?", *info.Status)
	}

	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err = db.Scopes(info.Paginate()).
		Order("CASE WHEN wo.status = 0 THEN 0 ELSE 1 END ASC").
		Order("wo.created_at DESC").
		Scan(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (withdrawAuditService *WithdrawAuditService) ApproveWithdrawOrder(ctx context.Context, orderID string, reviewerID uint, reviewer string, remark string) error {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return errors.New("order id is required")
	}

	var order withdrawOrderRow
	// TON payouts are executed by the payment backend, which owns the hot
	// wallet key. Do not run the legacy gateway flow from the admin service.
	var tonOrder withdrawOrderRow
	if err := global.GVA_DB.WithContext(ctx).Table("withdraw_orders").Where("order_id = ?", orderID).Take(&tonOrder).Error; err == nil && strings.EqualFold(strings.TrimSpace(tonOrder.DstCode), "TON") {
		return approveTONViaPaymentBackend(ctx, orderID, reviewerID, reviewer, remark)
	}
	now := time.Now()
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("withdraw_orders").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_id = ?", orderID).
			Take(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("withdraw order not found")
			}
			return err
		}

		if order.Status != withdrawStatusPendingReview {
			return fmt.Errorf("withdraw order status %d cannot be approved", order.Status)
		}

		return tx.Table("withdraw_orders").Where("id = ?", order.ID).Updates(map[string]interface{}{
			"status":        withdrawStatusProcessing,
			"reviewed_by":   reviewerID,
			"reviewer":      strings.TrimSpace(reviewer),
			"review_remark": strings.TrimSpace(remark),
			"reviewed_at":   now,
			"ref_msg":       "submitting to gateway",
			"updated_at":    now,
		}).Error
	})
	if err != nil {
		return err
	}

	platOrderID, refMsg, err := withdrawAuditService.submitApprovedWithdrawToGateway(ctx, &order)
	if err != nil {
		revertNow := time.Now()
		_ = global.GVA_DB.WithContext(ctx).Table("withdraw_orders").
			Where("order_id = ? AND status = ?", orderID, withdrawStatusProcessing).
			Updates(map[string]interface{}{
				"status":     withdrawStatusPendingReview,
				"ref_msg":    truncateWithdrawRefMsg("gateway submit failed: " + err.Error()),
				"updated_at": revertNow,
			}).Error
		return err
	}

	updateNow := time.Now()
	return global.GVA_DB.WithContext(ctx).Table("withdraw_orders").
		Where("order_id = ? AND status = ?", orderID, withdrawStatusProcessing).
		Updates(map[string]interface{}{
			"plat_order_id": platOrderID,
			"ref_msg":       truncateWithdrawRefMsg(firstNonEmpty(refMsg, "submitted to gateway")),
			"updated_at":    updateNow,
		}).Error
}

func approveTONViaPaymentBackend(ctx context.Context, orderID string, reviewerID uint, reviewer, remark string) error {
	cfg := global.GVA_CONFIG.CustomCfg
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.Commonsrv), "/")
	if baseURL == "" {
		return errors.New("payment backend URL is not configured")
	}
	body, _ := json.Marshal(map[string]interface{}{"order_id": orderID, "reviewer_id": reviewerID, "reviewer": reviewer, "remark": remark})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/withdraw/admin/approve", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(cfg.CommonsrvToken); token != "" {
		req.Header.Set("X-Internal-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("payment backend request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var result map[string]interface{}
		if json.Unmarshal(data, &result) == nil {
			if msg, ok := result["error"].(string); ok && msg != "" {
				return errors.New(msg)
			}
		}
		return fmt.Errorf("payment backend returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (withdrawAuditService *WithdrawAuditService) RejectWithdrawOrder(ctx context.Context, orderID string, reviewerID uint, reviewer string, remark string) error {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return errors.New("order id is required")
	}

	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order withdrawOrderRow
		if err := tx.Table("withdraw_orders").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_id = ?", orderID).
			Take(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("withdraw order not found")
			}
			return err
		}

		if order.Status != withdrawStatusPendingReview {
			return fmt.Errorf("withdraw order status %d cannot be rejected", order.Status)
		}

		var user withdrawUserRow
		if err := tx.Table("users").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id, username, balance").
			Where("id = ?", order.UserID).
			Take(&user).Error; err != nil {
			return err
		}

		now := time.Now()
		refundAmount := order.Cost
		updates := map[string]interface{}{
			"status":        withdrawStatusRejected,
			"ref_code":      2,
			"ref_msg":       "rejected",
			"reviewed_by":   reviewerID,
			"reviewer":      reviewer,
			"review_remark": strings.TrimSpace(remark),
			"reviewed_at":   now,
			"completed_at":  now,
			"updated_at":    now,
		}
		if err := tx.Table("withdraw_orders").Where("id = ?", order.ID).Updates(updates).Error; err != nil {
			return err
		}

		if refundAmount > 0 {
			if err := tx.Table("users").
				Where("id = ?", user.ID).
				Updates(map[string]interface{}{
					"balance":    gorm.Expr("balance + ?", refundAmount),
					"updated_at": now,
				}).Error; err != nil {
				return err
			}

			reviewRemark := strings.TrimSpace(remark)
			if reviewRemark == "" {
				reviewRemark = "admin rejected"
			}
			if err := tx.Table("transactions").Create(map[string]interface{}{
				"user_id":        user.ID,
				"type":           6,
				"amount":         refundAmount,
				"before_balance": user.Balance,
				"after_balance":  user.Balance + refundAmount,
				"reference_id":   order.OrderID,
				"remark":         "withdraw rejected refund: " + reviewRemark,
				"created_at":     now,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (withdrawAuditService *WithdrawAuditService) submitApprovedWithdrawToGateway(ctx context.Context, order *withdrawOrderRow) (string, string, error) {
	methodCfg := findWithdrawMethod(order.DstCode, order.Type)
	currency := methodCfg.Currency
	if currency == "" {
		currency = getWithdrawCurrency(order.DstCode)
	}

	if strings.EqualFold(methodCfg.Channel, "quantixcore") {
		return submitQuantixCoreWithdraw(ctx, order, methodCfg, currency)
	}

	return submitLegacyGatewayWithdraw(ctx, order)
}

func submitLegacyGatewayWithdraw(ctx context.Context, order *withdrawOrderRow) (string, string, error) {
	paymentCfg := getPaymentConfig()
	if paymentCfg.MerchantID == "" || paymentCfg.SecretKey == "" || paymentCfg.BaseURL == "" {
		return "", "", errors.New("payment config is not initialized")
	}

	params := buildLegacyGatewayWithdrawParams(order, paymentCfg, withdrawFeeValues(getWithdrawFeeConfig(ctx, global.GVA_DB)))
	params["sign"] = generateLegacyGatewaySign(params, paymentCfg.SecretKey)

	apiURL := strings.TrimSuffix(paymentCfg.BaseURL, "/") + "/payment_dfpay_add.html"
	respData, err := postMultipartFormData(ctx, apiURL, params)
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
	if !strings.EqualFold(gatewayResp.Status, "success") {
		return "", "", fmt.Errorf("withdraw gateway error: %s (errCode: %d)", gatewayResp.Msg, gatewayResp.ErrCode)
	}

	return gatewayResp.PlatOrderID, gatewayResp.Msg, nil
}

func submitQuantixCoreWithdraw(ctx context.Context, order *withdrawOrderRow, methodCfg withdrawMethodConfig, currency string) (string, string, error) {
	paymentCfg := getPaymentConfig()
	quantixCfg := getQuantixCoreConfig(paymentCfg, currency)
	if quantixCfg.MerchantID == "" || quantixCfg.SecretKey == "" {
		return "", "", errors.New("QuantixCore config is not initialized")
	}

	payload, err := buildQuantixCoreWithdrawPayload(order, methodCfg, currency, quantixCfg, paymentCfg, withdrawFeeValues(getWithdrawFeeConfig(ctx, global.GVA_DB)))
	if err != nil {
		return "", "", err
	}
	payload["sign"] = generateQuantixCoreSign(payload, quantixCfg.SecretKey)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("marshal QuantixCore withdraw request failed: %w", err)
	}

	apiURL := strings.TrimSuffix(getQuantixCoreBaseURL(paymentCfg, quantixCfg), "/") + "/api/open/flex/order/payment/add"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("create QuantixCore withdraw request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
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

func buildLegacyGatewayWithdrawParams(order *withdrawOrderRow, cfg config.PaymentConfig, feeConfig withdrawFeeConfigValues) map[string]string {
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
	order *withdrawOrderRow,
	methodCfg withdrawMethodConfig,
	currency string,
	quantixCfg config.QuantixCoreRegionConfig,
	cfg config.PaymentConfig,
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

	if bankCode := resolveQuantixCoreBankCode(order); bankCode != "" {
		payload["bankCode"] = bankCode
	}

	return payload, nil
}

func findWithdrawMethod(dstCode string, withdrawType string) withdrawMethodConfig {
	normalizedCode := strings.ToUpper(strings.TrimSpace(dstCode))
	for _, cfg := range withdrawMethodConfigs {
		if cfg.Code == normalizedCode {
			return cfg
		}
	}

	currency := getWithdrawCurrency(normalizedCode)
	channel := "gateway"
	accountType := ""
	if currency == "PHP" {
		channel = "quantixcore"
		accountType = quantixCoreWalletAccountTypes[normalizedCode]
		if accountType == "" && strings.EqualFold(withdrawType, "bankcard") {
			accountType = "PERSONAL_BANK"
		}
	}

	return withdrawMethodConfig{
		Code:        normalizedCode,
		Type:        normalizeLegacyWithdrawType(withdrawType),
		Currency:    currency,
		Channel:     channel,
		AccountType: accountType,
	}
}

func getPaymentConfig() config.PaymentConfig {
	cfg := global.GVA_CONFIG.CustomCfg.Payment
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.MerchantID = strings.TrimSpace(cfg.MerchantID)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.NotifyURL = strings.TrimSpace(cfg.NotifyURL)
	cfg.QuantixCore.BaseURL = strings.TrimSpace(cfg.QuantixCore.BaseURL)
	cfg.QuantixCore.ID.BaseURL = strings.TrimSpace(cfg.QuantixCore.ID.BaseURL)
	cfg.QuantixCore.ID.MerchantID = strings.TrimSpace(cfg.QuantixCore.ID.MerchantID)
	cfg.QuantixCore.ID.SecretKey = strings.TrimSpace(cfg.QuantixCore.ID.SecretKey)
	cfg.QuantixCore.ID.NotifyURL = strings.TrimSpace(cfg.QuantixCore.ID.NotifyURL)
	cfg.QuantixCore.PH.BaseURL = strings.TrimSpace(cfg.QuantixCore.PH.BaseURL)
	cfg.QuantixCore.PH.MerchantID = strings.TrimSpace(cfg.QuantixCore.PH.MerchantID)
	cfg.QuantixCore.PH.SecretKey = strings.TrimSpace(cfg.QuantixCore.PH.SecretKey)
	cfg.QuantixCore.PH.NotifyURL = strings.TrimSpace(cfg.QuantixCore.PH.NotifyURL)
	return cfg
}

func getQuantixCoreConfig(cfg config.PaymentConfig, currency string) config.QuantixCoreRegionConfig {
	defaultCfg := config.QuantixCoreRegionConfig{
		BaseURL:    firstNonEmpty(cfg.QuantixCore.BaseURL, cfg.BaseURL),
		MerchantID: cfg.MerchantID,
		SecretKey:  cfg.SecretKey,
		NotifyURL:  cfg.NotifyURL,
	}

	regionCfg := config.QuantixCoreRegionConfig{}
	switch normalizeCurrency(currency) {
	case "PHP":
		regionCfg = cfg.QuantixCore.PH
	case "IDR":
		regionCfg = cfg.QuantixCore.ID
	}

	return config.QuantixCoreRegionConfig{
		BaseURL:    firstNonEmpty(regionCfg.BaseURL, defaultCfg.BaseURL),
		MerchantID: firstNonEmpty(regionCfg.MerchantID, defaultCfg.MerchantID),
		SecretKey:  firstNonEmpty(regionCfg.SecretKey, defaultCfg.SecretKey),
		NotifyURL:  firstNonEmpty(regionCfg.NotifyURL, defaultCfg.NotifyURL),
	}
}

func resolveQuantixCoreAccountType(order *withdrawOrderRow, methodCfg withdrawMethodConfig) string {
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

func resolveQuantixCoreBankCode(order *withdrawOrderRow) string {
	if !strings.EqualFold(order.Type, "bankcard") {
		return ""
	}
	normalizedCode := strings.ToUpper(strings.TrimSpace(order.DstCode))
	if _, isWallet := quantixCoreWalletAccountTypes[normalizedCode]; isWallet {
		return ""
	}
	return normalizedCode
}

func getWithdrawCurrency(dstCode string) string {
	normalizedCode := strings.ToUpper(strings.TrimSpace(dstCode))
	if strings.HasPrefix(normalizedCode, "GCASH") || normalizedCode == "PHP" {
		return "PHP"
	}
	return "IDR"
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
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
	factor := minorUnitFactor(currency)
	if factor <= 0 {
		return netAmount
	}
	return math.Round(netAmount*factor) / factor
}

func minorUnitFactor(currency string) float64 {
	if normalizeCurrency(currency) == "PHP" {
		return 100
	}
	return 1
}

func amountToMinorUnits(amount float64, currency string) int {
	return int(math.Round(amount * minorUnitFactor(currency)))
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

func postMultipartFormData(ctx context.Context, apiURL string, params map[string]string) ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range params {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(respData))
	}
	return respData, nil
}

func generateLegacyGatewaySign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || isSignEmptyValue(value) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, params[key]))
	}
	signStr := strings.Join(parts, "&") + "&key=" + secret
	sum := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func generateQuantixCoreSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
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

func isSignEmptyValue(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "0", "null", "false":
		return true
	default:
		return false
	}
}

func getQuantixCoreBaseURL(cfg config.PaymentConfig, quantixCfg config.QuantixCoreRegionConfig) string {
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

func truncateWithdrawRefMsg(message string) string {
	message = strings.TrimSpace(message)
	if len(message) <= 255 {
		return message
	}
	return message[:255]
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

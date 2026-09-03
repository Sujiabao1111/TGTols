package controllers

import (
	"fmt"
	"gogogo/models/dtos"
	"gogogo/services"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gogogo/helpers"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService *services.PaymentService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		paymentService: services.GetPaymentService(),
	}
}

// ============================================
// 支付方式 API
// ============================================

// GetPaymentMethods 获取支持的支付方式列表
func (h *PaymentHandler) GetPaymentMethods(c *fiber.Ctx) error {
	methods, err := h.paymentService.GetPaymentMethods(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(dtos.PaymentMethodsResponse{Methods: methods})
}

func (h *PaymentHandler) CreateTonOrder(c *fiber.Ctx) error {
	uid := getUserIDFromJWT(c)
	if uid == 0 {
		return c.Status(401).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}
	var req dtos.TonCreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	resp, err := h.paymentService.CreateTonOrder(c.Context(), uid, req.Amount)
	if err != nil {
		return c.Status(400).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(resp)
}

func (h *PaymentHandler) GetTonRate(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"usd_per_ton": h.paymentService.CurrentTONRate(c.Context())})
}

func (h *PaymentHandler) ConfirmTonOrder(c *fiber.Ctx) error {
	uid := getUserIDFromJWT(c)
	if uid == 0 {
		return c.Status(401).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}
	var req struct {
		OrderID string `json:"order_id"`
		TxHash  string `json:"tx_hash"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	if err := h.paymentService.ConfirmTonOrder(c.Context(), uid, req.OrderID, req.TxHash); err != nil {
		return c.Status(400).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(fiber.Map{"status": "paid", "order_id": req.OrderID})
}

func (h *PaymentHandler) GetWithdrawMethods(c *fiber.Ctx) error {
	methods, err := h.paymentService.GetWithdrawMethods(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(dtos.WithdrawMethodsResponse{Methods: methods})
}

// ============================================
// 代收 (充值) API
// ============================================

// CreatePaymentOrder 创建支付订单
func (h *PaymentHandler) CreatePaymentOrder(c *fiber.Ctx) error {
	fmt.Println("[CreatePaymentOrder] ========== 开始创建支付订单 ==========")

	userID := getUserIDFromJWT(c)
	fmt.Println("[CreatePaymentOrder] userID:", userID)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
	}

	fmt.Println("[CreatePaymentOrder] Raw body:", string(c.Body()))

	var req dtos.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("[CreatePaymentOrder] BodyParser error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请求参数错误: " + err.Error()})
	}

	fmt.Printf("[CreatePaymentOrder] Parsed request: %+v\n", req)
	req.ClientIP = QueryClientIP(c)

	// 参数验证
	if req.Amount <= 0 {
		fmt.Println("[CreatePaymentOrder] 参数错误: 金额必须大于0")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "金额必须大于0"})
	}
	if req.Type == "" || req.DstCode == "" {
		fmt.Println("[CreatePaymentOrder] 参数错误: 支付类型和支付方式不能为空")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "支付类型和支付方式不能为空"})
	}

	fmt.Println("[CreatePaymentOrder] 调用 paymentService.CreatePaymentOrder...")
	resp, err := h.paymentService.CreatePaymentOrder(c.Context(), userID, req)
	if err != nil {
		fmt.Println("[CreatePaymentOrder] 创建订单失败:", err)
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	fmt.Printf("[CreatePaymentOrder] 订单创建成功: %+v\n", resp)
	fmt.Println("[CreatePaymentOrder] ========== 结束 ==========")
	return c.JSON(resp)
}

// GetPaymentStatus 查询支付订单状态
func (h *PaymentHandler) GetPaymentStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
	}

	orderID := c.Params("orderId")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "订单号不能为空"})
	}

	status, err := h.paymentService.GetPaymentStatus(c.Context(), userID, orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

// GetUserPayments 获取用户支付记录
func (h *PaymentHandler) GetUserPayments(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
	}

	// 默认限制20条
	limit := 20
	if l := c.QueryInt("limit"); l > 0 && l <= 100 {
		limit = l
	}

	orders, err := h.paymentService.GetUserPayments(c.Context(), userID, limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(orders)
}

func (h *PaymentHandler) CreateTelegramStarsOrder(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "not logged in"})
	}
	var req dtos.CreateTelegramStarsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid request"})
	}
	resp, err := h.paymentService.CreateTelegramStarsOrder(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(resp)
}

func (h *PaymentHandler) HandleTelegramStarsWebhook(c *fiber.Ctx) error {
	secret := services.TelegramWebhookSecret()
	if secret == "" || c.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var update services.TelegramStarsUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid update"})
	}
	if err := h.paymentService.HandleTelegramStarsUpdate(c.Context(), update); err != nil {
		fmt.Printf("[TelegramStars] webhook error: %v\n", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func requireInternalToken(c *fiber.Ctx) bool {
	cfg := helpers.GetCfgInstance()
	return cfg != nil && cfg.Conf != nil && strings.TrimSpace(cfg.Conf.InternalToken) != "" && c.Get("X-Internal-Token") == strings.TrimSpace(cfg.Conf.InternalToken)
}

func (h *PaymentHandler) RefundTelegramStarsPayment(c *fiber.Ctx) error {
	if !requireInternalToken(c) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req dtos.TelegramRefundRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid request"})
	}
	if err := services.TelegramRefundStarPayment(c.Context(), req.TelegramUserID, req.ChargeID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *PaymentHandler) GetTelegramStarsTransactions(c *fiber.Ctx) error {
	if !requireInternalToken(c) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	result, err := services.TelegramGetStarTransactions(c.Context(), c.QueryInt("offset", 0), c.QueryInt("limit", 100))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}
	c.Type("json")
	return c.Send(result)
}

// HandlePaymentNotify 处理支付回调通知 (公开接口，无需JWT)
func (h *PaymentHandler) HandlePaymentNotify(c *fiber.Ctx) error {
	var notify dtos.PaymentNotifyRequest

	// 打印调试信息
	fmt.Println("[PaymentNotify] Content-Type:", c.Get("Content-Type"))
	fmt.Println("[PaymentNotify] Raw body:", string(c.Body()))
	fmt.Println("[PaymentNotify] Query params:", c.Request().URI().QueryArgs().String())

	// 解析form-data格式的回调参数
	if err := c.BodyParser(&notify); err != nil {
		fmt.Println("[PaymentNotify] BodyParser error:", err)
		return c.Status(fiber.StatusBadRequest).SendString("FAIL")
	}

	fmt.Printf("[PaymentNotify] After BodyParser: %+v\n", notify)

	// 如果BodyParser解析失败，手动解析form（支持对方驼峰命名格式）
	if notify.OrderID == "" && notify.MerchantOrderNo == "" {
		// 优先尝试驼峰格式参数名
		notify.MerchantID = c.FormValue("memberId")
		if notify.MerchantID == "" {
			notify.MerchantID = c.FormValue("merchant_id")
		}
		notify.OrderID = c.FormValue("orderId")
		if notify.OrderID == "" {
			notify.OrderID = c.FormValue("order_id")
		}
		notify.PlatOrderID = c.FormValue("platOrderId")
		if notify.PlatOrderID == "" {
			notify.PlatOrderID = c.FormValue("plat_order_id")
		}
		amountStr := c.FormValue("amount")
		if amountStr != "" {
			if amountFloat, err := strconv.ParseFloat(amountStr, 64); err == nil {
				notify.Amount = amountFloat
			}
		}
		notify.Cost = c.FormValue("cost")
		notify.Status = c.FormValue("status")
		refCodeStr := c.FormValue("refCode")
		if refCodeStr == "" {
			refCodeStr = c.FormValue("ref_code")
		}
		notify.RefCode, _ = strconv.Atoi(refCodeStr)
		notify.RefMsg = c.FormValue("refMsg")
		if notify.RefMsg == "" {
			notify.RefMsg = c.FormValue("ref_msg")
		}
		notify.Sign = c.FormValue("sign")
		// dstCode 在需要时可用于路由或记录
		_ = c.FormValue("dstCode")
	}

	fmt.Println("[PaymentNotify] After fallback:", notify)

	if notify.MerchantNo == "" {
		notify.MerchantNo = c.FormValue("merchantNo")
	}
	if notify.MerchantID == "" {
		notify.MerchantID = c.FormValue("memberId")
		if notify.MerchantID == "" {
			notify.MerchantID = c.FormValue("merchant_id")
		}
	}
	if notify.MerchantOrderNo == "" {
		notify.MerchantOrderNo = c.FormValue("merchantOrderNo")
	}
	if notify.OrderID == "" {
		notify.OrderID = c.FormValue("orderId")
		if notify.OrderID == "" {
			notify.OrderID = c.FormValue("order_id")
		}
	}
	if notify.OrderNo == "" {
		notify.OrderNo = c.FormValue("orderNo")
	}
	if notify.PlatOrderID == "" {
		notify.PlatOrderID = c.FormValue("platOrderId")
		if notify.PlatOrderID == "" {
			notify.PlatOrderID = c.FormValue("plat_order_id")
		}
	}

	amountStr := c.FormValue("amount")
	if notify.RawAmount == "" {
		notify.RawAmount = amountStr
	}
	if notify.Amount == 0 && amountStr != "" {
		if amountFloat, err := strconv.ParseFloat(amountStr, 64); err == nil {
			notify.Amount = amountFloat
		}
	}

	if notify.Cost == "" {
		notify.Cost = c.FormValue("cost")
	}
	if notify.Status == "" {
		notify.Status = c.FormValue("status")
	}
	if notify.RefCode == 0 {
		refCodeStr := c.FormValue("refCode")
		if refCodeStr == "" {
			refCodeStr = c.FormValue("ref_code")
		}
		if refCodeStr != "" {
			notify.RefCode, _ = strconv.Atoi(refCodeStr)
		}
	}
	if notify.RefMsg == "" {
		notify.RefMsg = c.FormValue("refMsg")
		if notify.RefMsg == "" {
			notify.RefMsg = c.FormValue("ref_msg")
		}
	}
	if notify.Sign == "" {
		notify.Sign = c.FormValue("sign")
	}
	if notify.Currency == "" {
		notify.Currency = c.FormValue("currency")
	}
	if notify.Code == "" {
		notify.Code = c.FormValue("code")
	}

	fmt.Printf("[PaymentNotify] After normalize: %+v\n", notify)

	result, err := h.paymentService.HandlePaymentNotify(c.Context(), notify)
	fmt.Println("[PaymentNotify] Result:", result, "Error:", err)
	if err != nil {
		// 记录错误日志
		return c.Status(fiber.StatusOK).SendString(result)
	}

	return c.Status(fiber.StatusOK).SendString(result)
}

// ============================================
// 代付 (提现) API
// ============================================

// CreateWithdrawOrder 创建代付订单 (提现)
func (h *PaymentHandler) CreateWithdrawOrder(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
	}

	var req dtos.CreateWithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请求参数错误: " + err.Error()})
	}

	// 参数验证
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "金额必须大于0"})
	}
	if req.Account == "" || (req.Type != "crypto" && req.AccountName == "") {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "账户信息不能为空"})
	}
	if req.Type != "crypto" && (req.Phone == "" || req.Email == "") {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "联系信息不能为空"})
	}

	req.ClientIP = QueryClientIP(c)
	order, err := h.paymentService.CreateWithdrawOrder(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(order)
}

// HandleWithdrawNotify 处理代付回调通知 (公开接口，无需JWT)
func (h *PaymentHandler) HandleWithdrawNotify(c *fiber.Ctx) error {
	var notify dtos.WithdrawNotifyRequest
	if err := c.BodyParser(&notify); err != nil {
		if !parseWithdrawNotifyJSONBody(c.Body(), &notify) {
			return c.Status(fiber.StatusBadRequest).SendString("FAIL")
		}
	}

	normalizeWithdrawNotifyForm(c, &notify)

	result, err := h.paymentService.HandleWithdrawNotify(c.Context(), notify)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString(result)
	}

	return c.Status(fiber.StatusOK).SendString(result)
}

func normalizeWithdrawNotifyForm(c *fiber.Ctx, notify *dtos.WithdrawNotifyRequest) {
	setStringFromFormIfEmpty(c, &notify.MerchantNo, "merchantNo")
	setStringFromFormIfEmpty(c, &notify.MerchantID, "merchant_id", "memberId")
	setStringFromFormIfEmpty(c, &notify.MerchantOrderNo, "merchantOrderNo")
	setStringFromFormIfEmpty(c, &notify.OrderID, "order_id", "orderId")
	setStringFromFormIfEmpty(c, &notify.OrderNo, "orderNo")
	setStringFromFormIfEmpty(c, &notify.PlatOrderID, "plat_order_id", "platOrderId")
	setStringFromFormIfEmpty(c, &notify.Amount, "amount")
	setStringFromFormIfEmpty(c, &notify.Cost, "cost")
	setStringFromFormIfEmpty(c, &notify.Status, "status")
	setStringFromFormIfEmpty(c, &notify.RefCode, "ref_code")
	setStringFromFormIfEmpty(c, &notify.RefCodeAlt, "refCode")
	setStringFromFormIfEmpty(c, &notify.RefMsg, "ref_msg")
	setStringFromFormIfEmpty(c, &notify.RefMsgAlt, "refMsg")
	setStringFromFormIfEmpty(c, &notify.ErrorMsg, "errorMsg")
	setStringFromFormIfEmpty(c, &notify.ErrorMsgAlt, "error_msg")
	setStringFromFormIfEmpty(c, &notify.Currency, "currency")
	setStringFromFormIfEmpty(c, &notify.Code, "code")
	setStringFromFormIfEmpty(c, &notify.Sign, "sign")
}

func parseWithdrawNotifyJSONBody(body []byte, notify *dtos.WithdrawNotifyRequest) bool {
	if len(body) == 0 {
		return false
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return false
	}

	setStringFromMapIfEmpty(data, &notify.MerchantNo, "merchantNo")
	setStringFromMapIfEmpty(data, &notify.MerchantID, "merchant_id", "memberId")
	setStringFromMapIfEmpty(data, &notify.MerchantOrderNo, "merchantOrderNo")
	setStringFromMapIfEmpty(data, &notify.OrderID, "order_id", "orderId")
	setStringFromMapIfEmpty(data, &notify.OrderNo, "orderNo")
	setStringFromMapIfEmpty(data, &notify.PlatOrderID, "plat_order_id", "platOrderId")
	setStringFromMapIfEmpty(data, &notify.Amount, "amount")
	setStringFromMapIfEmpty(data, &notify.Cost, "cost")
	setStringFromMapIfEmpty(data, &notify.Status, "status")
	setStringFromMapIfEmpty(data, &notify.RefCode, "ref_code")
	setStringFromMapIfEmpty(data, &notify.RefCodeAlt, "refCode")
	setStringFromMapIfEmpty(data, &notify.RefMsg, "ref_msg")
	setStringFromMapIfEmpty(data, &notify.RefMsgAlt, "refMsg")
	setStringFromMapIfEmpty(data, &notify.ErrorMsg, "errorMsg")
	setStringFromMapIfEmpty(data, &notify.ErrorMsgAlt, "error_msg")
	setStringFromMapIfEmpty(data, &notify.Currency, "currency")
	setStringFromMapIfEmpty(data, &notify.Code, "code")
	setStringFromMapIfEmpty(data, &notify.Sign, "sign")

	return notify.OrderID != "" || notify.MerchantOrderNo != ""
}

func setStringFromFormIfEmpty(c *fiber.Ctx, target *string, keys ...string) {
	if strings.TrimSpace(*target) != "" {
		return
	}
	for _, key := range keys {
		if value := strings.TrimSpace(c.FormValue(key)); value != "" {
			*target = value
			return
		}
	}
}

func setStringFromMapIfEmpty(data map[string]interface{}, target *string, keys ...string) {
	if strings.TrimSpace(*target) != "" {
		return
	}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if text := strings.TrimSpace(jsonValueToString(value)); text != "" {
				*target = text
				return
			}
		}
	}
}

func jsonValueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprint(v)
	}
}

func (h *PaymentHandler) AdminApproveWithdrawOrder(c *fiber.Ctx) error {
	internalToken := strings.TrimSpace(c.Get("X-Internal-Token"))
	cfg := helpers.GetCfgInstance()
	if cfg == nil || cfg.Conf == nil || strings.TrimSpace(cfg.Conf.InternalToken) == "" || internalToken != strings.TrimSpace(cfg.Conf.InternalToken) {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	var req dtos.AdminApproveWithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid request: " + err.Error()})
	}

	order, err := h.paymentService.ApproveWithdrawOrder(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(order)
}

// ============================================
// 199 Voucher 卡密支付 API
// ============================================

// GetVoucherConfig 获取卡密配置
func (h *PaymentHandler) GetVoucherConfig(c *fiber.Ctx) error {
	config := h.paymentService.GetVoucherConfig()
	return c.JSON(config)
}

// RedeemVoucher 卡密兑换
func (h *PaymentHandler) RedeemVoucher(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
	}

	var req dtos.VoucherRedeemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请求参数错误: " + err.Error()})
	}

	// 参数验证
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "金额必须大于0"})
	}
	if req.VoucherCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "卡密不能为空"})
	}

	resp, err := h.paymentService.RedeemVoucher(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

// ============================================
// TokenPay USDT 支付 API
// ============================================

// CreateTokenPayOrder 创建 TokenPay 订单
func (h *PaymentHandler) CreateTokenPayOrder(c *fiber.Ctx) error {
	fmt.Println("[CreateTokenPayOrder] ========== 开始创建 TokenPay 订单 ==========")

	userID := getUserIDFromJWT(c)
	fmt.Println("[CreateTokenPayOrder] userID:", userID)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
	}

	fmt.Println("[CreateTokenPayOrder] Raw body:", string(c.Body()))

	var req dtos.CreateTokenPayOrderRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("[CreateTokenPayOrder] BodyParser error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请求参数错误: " + err.Error()})
	}

	fmt.Printf("[CreateTokenPayOrder] Parsed request: %+v\n", req)
	req.ClientIP = QueryClientIP(c)

	// 参数验证
	if req.Amount <= 0 {
		fmt.Println("[CreateTokenPayOrder] 参数错误: 金额必须大于0")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "金额必须大于0"})
	}
	if req.ChainType != "ETH" && req.ChainType != "TRX" {
		fmt.Println("[CreateTokenPayOrder] 参数错误: 无效的链类型")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请选择有效的链类型（ETH 或 TRX）"})
	}

	fmt.Println("[CreateTokenPayOrder] 调用 paymentService.CreateTokenPayOrder...")
	resp, err := h.paymentService.CreateTokenPayOrder(c.Context(), userID, req)
	if err != nil {
		fmt.Println("[CreateTokenPayOrder] 创建订单失败:", err)
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	fmt.Printf("[CreateTokenPayOrder] 订单创建成功: %+v\n", resp)
	fmt.Println("[CreateTokenPayOrder] ========== 结束 ==========")
	return c.JSON(resp)
}

// HandleTokenPayNotify 处理 TokenPay 回调通知 (公开接口，无需JWT)
func (h *PaymentHandler) HandleTokenPayNotify(c *fiber.Ctx) error {
	fmt.Println("[TokenPayNotify] Content-Type:", c.Get("Content-Type"))
	fmt.Println("[TokenPayNotify] Raw body:", string(c.Body()))

	// 解析为 map 以便进行签名验证
	var notifyData map[string]interface{}
	if err := c.BodyParser(&notifyData); err != nil {
		fmt.Println("[TokenPayNotify] BodyParser error:", err)
		return c.Status(fiber.StatusBadRequest).SendString("FAIL")
	}

	fmt.Printf("[TokenPayNotify] Parsed data: %+v\n", notifyData)

	result, err := h.paymentService.HandleTokenPayNotify(c.Context(), notifyData)
	if err != nil {
		fmt.Println("[TokenPayNotify] Handle error:", err)
		return c.Status(fiber.StatusOK).SendString(result)
	}

	return c.Status(fiber.StatusOK).SendString(result)
}

# TokenPay USDT 支付渠道集成实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 集成 TokenPay 支付网关，支持 ETH 链(ERC20)和 TRX 链(TRC20)的 USDT 充值

**Architecture:**
- 后端新增 TokenPay 专属接口和回调处理
- 前端增加链选择弹窗和支付地址展示
- 使用现有汇率表进行 IDR 到 USDT 的转换
- ETH链最低 5 USDT，TRX链最低 10 USDT

**Tech Stack:** Go (Fiber), TypeScript (Next.js), MySQL, TokenPay API

---

## 前置条件

- [ ] TokenPay 服务已部署并获取 API Key
- [ ] 确认 TokenPay 回调 URL 可公网访问
- [ ] 更新 `be/config.json` 添加 tokenpay 配置段

---

## Task 1: 更新配置文件结构

**Files:**
- Modify: `be/helpers/config.go`

**Step 1: 添加 TokenPayConfig 结构体**

在 `PaymentConfig` 结构体中添加 TokenPay 配置：

```go
type PaymentConfig struct {
    BaseURL    string         `json:"base_url"`
    MerchantID string         `json:"merchant_id"`
    SecretKey  string         `json:"secret_key"`
    NotifyURL  string         `json:"notify_url"`
    ReturnURL  string         `json:"return_url"`
    TokenPay   TokenPayConfig `json:"tokenpay"` // 新增
}

type TokenPayConfig struct {
    BaseURL   string `json:"base_url"`
    APIKey    string `json:"api_key"`
    NotifyURL string `json:"notify_url"`
}
```

**Step 2: 验证编译通过**

```bash
cd be && go build ./...
```

Expected: 编译成功

**Step 3: Commit**

```bash
git add be/helpers/config.go
git commit -m "config: add TokenPay configuration structure"
```

---

## Task 2: 添加 TokenPay 数据模型

**Files:**
- Modify: `be/models/dtos/payment_models.go`

**Step 1: 在文件末尾添加 TokenPay 相关结构体**

```go
// ============================================
// TokenPay 支付相关结构体
// ============================================

// TokenPayCreateOrderRequest 创建 TokenPay 订单请求
type TokenPayCreateOrderRequest struct {
	OrderID   string  `json:"order_id"`   // 商户订单号
	Amount    float64 `json:"amount"`     // USDT 金额
	ChainType string  `json:"chain_type"` // ETH 或 TRX
	NotifyURL string  `json:"notify_url"`
	ReturnURL string  `json:"return_url"`
}

// TokenPayCreateOrderResponse TokenPay 创建订单响应
type TokenPayCreateOrderResponse struct {
	OrderID    string `json:"order_id"`
	PayAddress string `json:"pay_address"` // 支付地址
	PayAmount  string `json:"pay_amount"`  // 支付金额
	QRCode     string `json:"qr_code"`     // 二维码数据
	ExpiredAt  int64  `json:"expired_at"`  // 过期时间戳
	Status     string `json:"status"`
}

// TokenPayNotifyRequest TokenPay 回调通知
type TokenPayNotifyRequest struct {
	OrderID       string `json:"order_id"`
	TxHash        string `json:"tx_hash"`       // 区块链交易哈希
	Amount        string `json:"amount"`        // 实际支付金额
	Status        string `json:"status"`        // paid / pending / failed
	Chain         string `json:"chain"`         // ETH / TRX
	Confirmations int    `json:"confirmations"` // 确认数
	Timestamp     int64  `json:"timestamp"`
}

// CreateTokenPayOrderRequest 前端创建 TokenPay 订单请求
type CreateTokenPayOrderRequest struct {
	CreatePaymentRequest
	ChainType string `json:"chain_type"` // ETH 或 TRX
}
```

**Step 2: 验证编译通过**

```bash
cd be && go build ./...
```

**Step 3: Commit**

```bash
git add be/models/dtos/payment_models.go
git commit -m "feat(payment): add TokenPay DTOs"
```

---

## Task 3: 更新支付方式列表

**Files:**
- Modify: `be/services/payment_service.go`

**Step 1: 修改 GetPaymentMethods 启用 USDT**

```go
func (s *PaymentService) GetPaymentMethods(ctx context.Context) ([]dtos.PaymentMethod, error) {
	methods := []dtos.PaymentMethod{
		{Code: "DANA", Name: "DANA", Type: "channel", MinAmount: 10000, MaxAmount: 10000000, Enabled: true},
		// 启用 USDT，调整金额范围（以 IDR 计价，约 5-1000 USDT）
		{Code: "USDT", Name: "USDT", Type: "channel", MinAmount: 77500, MaxAmount: 15500000, Enabled: true},
		{Code: "PAYPAL", Name: "PayPal", Type: "channel", MinAmount: 10000, MaxAmount: 100000000, Enabled: false},
		{Code: "199VOUCHER", Name: "199 Voucher", Type: "voucher", MinAmount: 10000, MaxAmount: 1000000, Enabled: true},
	}

	return methods, nil
}
```

**Step 2: Commit**

```bash
git add be/services/payment_service.go
git commit -m "feat(payment): enable USDT payment method"
```

---

## Task 4: 实现 TokenPay 创建订单服务

**Files:**
- Modify: `be/services/payment_service.go`

**Step 1: 添加 CreateTokenPayOrder 方法**

在 `payment_service.go` 中 `GetPaymentMethods` 方法后添加：

```go
// CreateTokenPayOrder 创建 TokenPay 订单
func (s *PaymentService) CreateTokenPayOrder(ctx context.Context, userID uint64, req dtos.CreateTokenPayOrderRequest) (*dtos.TokenPayCreateOrderResponse, error) {
	// 1. 验证 TokenPay 配置
	cfg := s.getConfig()
	if cfg == nil || cfg.TokenPay.BaseURL == "" || cfg.TokenPay.APIKey == "" {
		return nil, errors.New("TokenPay 配置未初始化")
	}

	// 2. 验证链类型和最低金额
	chainMinAmount := map[string]float64{
		"ETH": 5,   // ETH链最低 5 USDT
		"TRX": 10,  // TRX链最低 10 USDT
	}

	minUSDT, ok := chainMinAmount[req.ChainType]
	if !ok {
		return nil, errors.New("无效的链类型，请选择 ETH 或 TRX")
	}

	// 3. IDR 转 USDT
	exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
	if err != nil {
		exchangeRate = 15500 // 默认汇率 1 USD = 15500 IDR
	}
	usdtAmount := req.Amount / exchangeRate

	// 验证最低金额
	if usdtAmount < minUSDT {
		return nil, fmt.Errorf("该链最低充值 %.0f USDT (约 %.0f IDR)", minUSDT, minUSDT*exchangeRate)
	}

	// 4. 生成订单号
	orderID := common.GenerateOrderID(userID)

	// 5. 构建 TokenPay 请求
	tokenPayReq := map[string]interface{}{
		"order_id":   orderID,
		"amount":     fmt.Sprintf("%.2f", usdtAmount),
		"currency":   "USDT",
		"chain":      req.ChainType,
		"notify_url": cfg.TokenPay.NotifyURL,
		"return_url": req.CallbackURL,
	}

	// 6. 调用 TokenPay API
	apiURL := strings.TrimSuffix(cfg.TokenPay.BaseURL, "/") + "/api/order"

	jsonBody, _ := json.Marshal(tokenPayReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.TokenPay.APIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("TokenPay 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 7. 解析响应
	var tokenPayResp dtos.TokenPayCreateOrderResponse
	if err := json.Unmarshal(respData, &tokenPayResp); err != nil {
		return nil, fmt.Errorf("解析 TokenPay 响应失败: %w", err)
	}

	if tokenPayResp.Status != "success" && tokenPayResp.Status != "pending" {
		return nil, fmt.Errorf("TokenPay 创建订单失败: %s", tokenPayResp.Status)
	}

	// 8. 保存订单到数据库
	order := dtos.PaymentOrder{
		UserID:      userID,
		OrderID:     orderID,
		Amount:      req.Amount,      // 原始 IDR 金额
		Type:        req.Type,
		DstCode:     "USDT_" + req.ChainType,
		Status:      0, // 待支付
		ProductInfo: fmt.Sprintf("USDT充值(%s)", req.ChainType),
	}

	if err := s.db.Create(&order).Error; err != nil {
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}

	// 填充订单号到响应
	tokenPayResp.OrderID = orderID

	return &tokenPayResp, nil
}
```

**Step 2: Commit**

```bash
git add be/services/payment_service.go
git commit -m "feat(payment): implement CreateTokenPayOrder service"
```

---

## Task 5: 实现 TokenPay 回调处理服务

**Files:**
- Modify: `be/services/payment_service.go`

**Step 1: 添加 HandleTokenPayNotify 方法**

```go
// HandleTokenPayNotify 处理 TokenPay 回调通知
func (s *PaymentService) HandleTokenPayNotify(ctx context.Context, notify dtos.TokenPayNotifyRequest) (string, error) {
	// 1. 查询订单
	var order dtos.PaymentOrder
	if err := s.db.Where("order_id = ?", notify.OrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "FAIL", errors.New("订单不存在")
		}
		return "FAIL", err
	}

	// 2. 检查订单状态
	if order.Status == 1 {
		return "OK", nil // 已处理过
	}

	// 3. 根据状态处理
	now := time.Now()
	switch notify.Status {
	case "paid": // 支付成功
		// 解析实际支付金额
		actualAmount, _ := strconv.ParseFloat(notify.Amount, 64)
		if actualAmount <= 0 {
			return "FAIL", errors.New("支付金额无效")
		}

		// 更新订单和余额（使用实际支付的 USDT 金额）
		if err := s.processTokenPaySuccess(ctx, &order, notify.TxHash, now, actualAmount); err != nil {
			return "FAIL", err
		}

	case "failed": // 支付失败
		order.Status = 2
		order.UpdatedAt = now
		if err := s.db.Save(&order).Error; err != nil {
			return "FAIL", err
		}
	}

	return "OK", nil
}

// processTokenPaySuccess 处理 TokenPay 支付成功
func (s *PaymentService) processTokenPaySuccess(ctx context.Context, order *dtos.PaymentOrder, txHash string, now time.Time, usdtAmount float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新订单状态
		order.Status = 1
		order.PlatOrderID = txHash // 使用交易哈希作为平台订单号
		order.PaidAt = &now
		order.UpdatedAt = now
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		// 2. 获取用户当前余额
		var user dtos.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := beforeBalance + usdtAmount

		// 3. 更新用户余额（直接加 USDT 金额）
		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", usdtAmount)).Error; err != nil {
			return err
		}

		// 4. 写入资金流水（使用 USDT 金额）
		transaction := dtos.Transaction{
			UserID:        order.UserID,
			Type:          1, // 充值
			Amount:        usdtAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   order.OrderID,
			Remark:        fmt.Sprintf("USDT充值 %s", order.DstCode),
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// 5. 更新用户累计充值
		if err := tx.Exec("UPDATE users SET total_deposit = total_deposit + ? WHERE id = ?", usdtAmount, order.UserID).Error; err != nil {
			return err
		}

		return nil
	})
}
```

**Step 2: Commit**

```bash
git add be/services/payment_service.go
git commit -m "feat(payment): implement HandleTokenPayNotify service"
```

---

## Task 6: 添加 TokenPay HTTP 接口

**Files:**
- Modify: `be/controllers/payment_handler.go`

**Step 1: 添加 CreateTokenPayOrder 接口**

在文件中找到 `CreatePaymentOrder` 方法，在其后添加：

```go
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
```

**Step 2: 添加 HandleTokenPayNotify 接口**

在文件末尾 `RedeemVoucher` 方法后添加：

```go
// HandleTokenPayNotify 处理 TokenPay 回调通知 (公开接口，无需JWT)
func (h *PaymentHandler) HandleTokenPayNotify(c *fiber.Ctx) error {
	fmt.Println("[TokenPayNotify] Content-Type:", c.Get("Content-Type"))
	fmt.Println("[TokenPayNotify] Raw body:", string(c.Body()))

	var notify dtos.TokenPayNotifyRequest
	if err := c.BodyParser(&notify); err != nil {
		fmt.Println("[TokenPayNotify] BodyParser error:", err)
		return c.Status(fiber.StatusBadRequest).SendString("FAIL")
	}

	fmt.Printf("[TokenPayNotify] Parsed: %+v\n", notify)

	result, err := h.paymentService.HandleTokenPayNotify(c.Context(), notify)
	if err != nil {
		fmt.Println("[TokenPayNotify] Handle error:", err)
		return c.Status(fiber.StatusOK).SendString(result)
	}

	return c.Status(fiber.StatusOK).SendString(result)
}
```

**Step 3: Commit**

```bash
git add be/controllers/payment_handler.go
git commit -m "feat(payment): add TokenPay HTTP handlers"
```

---

## Task 7: 注册 TokenPay 路由

**Files:**
- Modify: `be/controllers/router.go`

**Step 1: 在路由注册中添加 TokenPay 接口**

找到支付相关路由注册处，添加：

```go
// TokenPay USDT 支付接口
paymentGroup.Post("/tokenpay/order", paymentHandler.CreateTokenPayOrder)
paymentGroup.Post("/tokenpay/notify", paymentHandler.HandleTokenPayNotify) // 公开接口，无需JWT
```

**Step 2: 验证编译通过**

```bash
cd be && go build ./...
```

**Step 3: Commit**

```bash
git add be/controllers/router.go
git commit -m "feat(payment): register TokenPay routes"
```

---

## Task 8: 更新前端支付服务

**Files:**
- Modify: `fe/services/payment.ts`

**Step 1: 添加 TokenPay 类型定义**

在文件中找到类型定义区域，添加：

```typescript
// TokenPay 创建订单响应
export interface TokenPayCreateOrderResponse {
  order_id: string
  pay_address: string      // 支付地址
  pay_amount: string       // 支付金额（USDT）
  qr_code: string          // 二维码数据
  expired_at: number       // 过期时间戳
  status: string
}

// TokenPay 创建订单请求
export interface CreateTokenPayOrderRequest {
  amount: number
  type: string
  dst_code: string
  callback_url?: string
  chain_type: "ETH" | "TRX"  // 链类型
}
```

**Step 2: 添加 createTokenPayOrder 方法**

在 `paymentService` 对象中添加：

```typescript
  // TokenPay 创建订单
  async createTokenPayOrder(data: CreateTokenPayOrderRequest): Promise<TokenPayCreateOrderResponse> {
    try {
      const response = await fetchWithAuth(`${API_BASE_URL}/payments/tokenpay/order`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })

      const responseData = await response.json().catch(() => ({}))

      if (!response.ok) {
        const errorMessage = responseData.error || responseData.message || "创建订单失败"
        throw new Error(errorMessage)
      }

      return responseData
    } catch (error) {
      console.error("Create TokenPay order failed:", error)
      throw error
    }
  },
```

**Step 3: Commit**

```bash
git add fe/services/payment.ts
git commit -m "feat(payment): add TokenPay frontend service"
```

---

## Task 9: 创建 TokenPay 支付弹窗组件

**Files:**
- Create: `fe/components/payment/TokenPayModal.tsx`

**Step 1: 创建组件文件**

```tsx
"use client"

import { QRCodeSVG } from "qrcode.react"
import { Copy, Check, X } from "lucide-react"
import { useState, useEffect } from "react"

interface TokenPayModalProps {
  isOpen: boolean
  onClose: () => void
  payAddress: string
  payAmount: string
  chainType: "ETH" | "TRX"
  expiredAt: number
}

export function TokenPayModal({
  isOpen,
  onClose,
  payAddress,
  payAmount,
  chainType,
  expiredAt,
}: TokenPayModalProps) {
  const [copied, setCopied] = useState(false)
  const [timeLeft, setTimeLeft] = useState(0)

  useEffect(() => {
    if (!isOpen) return

    const updateTimer = () => {
      const seconds = Math.max(0, Math.floor((expiredAt * 1000 - Date.now()) / 1000))
      setTimeLeft(seconds)
    }

    updateTimer()
    const timer = setInterval(updateTimer, 1000)

    return () => clearInterval(timer)
  }, [isOpen, expiredAt])

  if (!isOpen) return null

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(payAddress)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // 复制失败静默处理
    }
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, "0")}`
  }

  const chainNames: Record<string, string> = {
    ETH: "Ethereum (ERC20)",
    TRX: "Tron (TRC20)",
  }

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-lucky-dark rounded-2xl p-6 max-w-md w-full relative">
        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 text-gray-400 hover:text-white"
        >
          <X size={20} />
        </button>

        <h3 className="text-xl font-bold text-white mb-2 text-center">
          USDT 充值
        </h3>
        <p className="text-gray-400 text-sm text-center mb-4">
          {chainNames[chainType]}
        </p>

        {/* 倒计时 */}
        <div className="bg-yellow-500/20 rounded-lg p-3 mb-4 text-center">
          <p className="text-yellow-500 text-sm">
            剩余时间: {formatTime(timeLeft)}
          </p>
        </div>

        {/* 二维码 */}
        <div className="bg-white p-4 rounded-xl flex justify-center mb-4">
          <QRCodeSVG value={payAddress} size={180} />
        </div>

        {/* 支付金额 */}
        <div className="bg-lucky-gold/20 rounded-lg p-3 mb-4 text-center">
          <p className="text-gray-400 text-xs mb-1">支付金额</p>
          <p className="text-2xl font-bold text-lucky-gold">{payAmount} USDT</p>
        </div>

        {/* 支付地址 */}
        <div className="bg-white/10 rounded-lg p-3 mb-4">
          <p className="text-gray-400 text-xs mb-1">支付地址</p>
          <div className="flex items-center gap-2">
            <p className="text-white text-xs font-mono break-all flex-1">
              {payAddress}
            </p>
            <button
              onClick={handleCopy}
              className="p-2 bg-lucky-gold/20 rounded-lg hover:bg-lucky-gold/30 transition-colors shrink-0"
            >
              {copied ? (
                <Check size={16} className="text-green-400" />
              ) : (
                <Copy size={16} className="text-lucky-gold" />
              )}
            </button>
          </div>
        </div>

        {/* 提示 */}
        <p className="text-gray-400 text-xs text-center">
          请向以上地址转账指定金额，转账完成后系统将自动确认
        </p>
      </div>
    </div>
  )
}
```

**Step 2: Commit**

```bash
git add fe/components/payment/TokenPayModal.tsx
git commit -m "feat(payment): add TokenPay payment modal component"
```

---

## Task 10: 创建链选择弹窗组件

**Files:**
- Create: `fe/components/payment/ChainSelectorModal.tsx`

**Step 1: 创建组件文件**

```tsx
"use client"

import { X } from "lucide-react"

interface ChainSelectorModalProps {
  isOpen: boolean
  onClose: () => void
  onSelect: (chain: "ETH" | "TRX") => void
  amount: number
  idrRate?: number // 1 USD = ? IDR
}

export function ChainSelectorModal({
  isOpen,
  onClose,
  onSelect,
  amount,
  idrRate = 15500,
}: ChainSelectorModalProps) {
  if (!isOpen) return null

  const usdtAmount = (amount / idrRate).toFixed(2)

  const chains = [
    {
      code: "ETH" as const,
      name: "Ethereum",
      standard: "ERC20",
      minAmount: 5,
      color: "from-blue-500 to-purple-500",
      icon: "⟠",
    },
    {
      code: "TRX" as const,
      name: "Tron",
      standard: "TRC20",
      minAmount: 10,
      color: "from-red-500 to-orange-500",
      icon: "🔴",
    },
  ]

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-lucky-dark rounded-2xl p-6 max-w-sm w-full relative">
        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 text-gray-400 hover:text-white"
        >
          <X size={20} />
        </button>

        <h3 className="text-xl font-bold text-white mb-2 text-center">
          选择链类型
        </h3>
        <p className="text-gray-400 text-sm text-center mb-6">
          约 {usdtAmount} USDT
        </p>

        <div className="space-y-3">
          {chains.map((chain) => (
            <button
              key={chain.code}
              onClick={() => onSelect(chain.code)}
              className={`w-full py-4 bg-gradient-to-r ${chain.color} rounded-xl flex items-center justify-between px-4 hover:opacity-90 transition-opacity`}
            >
              <div className="flex items-center gap-3">
                <span className="text-2xl">{chain.icon}</span>
                <div className="text-left">
                  <span className="text-white font-bold block">
                    {chain.name}
                  </span>
                  <span className="text-white/70 text-xs">
                    {chain.standard}
                  </span>
                </div>
              </div>
              <span className="text-white/80 text-xs">
                最低 {chain.minAmount} USDT
              </span>
            </button>
          ))}
        </div>

        <button
          onClick={onClose}
          className="w-full mt-4 py-3 text-gray-400 hover:text-white transition-colors"
        >
          取消
        </button>
      </div>
    </div>
  )
}
```

**Step 3: Commit**

```bash
git add fe/components/payment/ChainSelectorModal.tsx
git commit -m "feat(payment): add chain selector modal component"
```

---

## Task 11: 更新 DepositForm 组件

**Files:**
- Modify: `fe/components/payment/DepositForm.tsx`

**Step 1: 导入新组件和类型**

在文件顶部添加导入：

```tsx
import { TokenPayModal } from "./TokenPayModal"
import { ChainSelectorModal } from "./ChainSelectorModal"
```

**Step 2: 添加状态变量**

在组件中找到 `useState` 声明处，添加：

```tsx
const [showChainSelector, setShowChainSelector] = useState(false)
const [showTokenPayModal, setShowTokenPayModal] = useState(false)
const [tokenPayData, setTokenPayData] = useState<TokenPayCreateOrderResponse | null>(null)
const [selectedChain, setSelectedChain] = useState<"ETH" | "TRX">("ETH")
```

**Step 3: 添加 USDT 特殊处理逻辑**

在 `handleSubmit` 函数中找到验证金额的部分，添加 USDT 特殊处理：

```tsx
// USDT 特殊处理：显示链选择弹窗
if (selectedMethod.code === "USDT") {
  setShowChainSelector(true)
  return
}
```

**Step 4: 添加 handleChainSelect 函数**

在 `handleSubmit` 函数后添加：

```tsx
const handleChainSelect = async (chain: "ETH" | "TRX") => {
  setSelectedChain(chain)
  setShowChainSelector(false)

  const numAmount = parseFloat(amount)
  setLoading(true)

  try {
    const currentUrl = window.location.origin + "/wallet"

    const response = await paymentService.createTokenPayOrder({
      amount: numAmount,
      type: "channel",
      dst_code: "USDT",
      callback_url: currentUrl,
      chain_type: chain,
    })

    setTokenPayData(response)
    setShowTokenPayModal(true)

    toast({
      title: "订单创建成功",
      description: "请完成转账",
    })
  } catch (error: any) {
    toast({
      title: "创建订单失败",
      description: error.message || t("common.error"),
      variant: "destructive",
    })
  } finally {
    setLoading(false)
  }
}
```

**Step 5: 添加 UI 组件**

在 `return` 语句中找到 `VoucherModal` 的部分，在其后添加：

```tsx
{/* 链选择弹窗 */}
<ChainSelectorModal
  isOpen={showChainSelector}
  onClose={() => setShowChainSelector(false)}
  onSelect={handleChainSelect}
  amount={parseFloat(amount) || 0}
/>

{/* TokenPay 支付弹窗 */}
{tokenPayData && (
  <TokenPayModal
    isOpen={showTokenPayModal}
    onClose={() => {
      setShowTokenPayModal(false)
      onSuccess?.()
    }}
    payAddress={tokenPayData.pay_address}
    payAmount={tokenPayData.pay_amount}
    chainType={selectedChain}
    expiredAt={tokenPayData.expired_at}
  />
)}
```

**Step 6: 显示 USDT 预估金额**

在金额输入区域添加预估 USDT 显示：

```tsx
{selectedMethod?.code === "USDT" && amount && (
  <p className="text-xs text-lucky-gold mt-2">
    预估: ~{(parseFloat(amount) / 15500).toFixed(2)} USDT
  </p>
)}
```

**Step 7: Commit**

```bash
git add fe/components/payment/DepositForm.tsx
git commit -m "feat(payment): integrate TokenPay into DepositForm"
```

---

## Task 12: 安装 QRCode 依赖

**Files:**
- Modify: `fe/package.json`

**Step 1: 安装 qrcode.react**

```bash
cd fe && npm install qrcode.react
```

**Step 2: Commit**

```bash
git add fe/package.json fe/package-lock.json
git commit -m "chore: add qrcode.react dependency"
```

---

## Task 13: 更新配置文件

**Files:**
- Modify: `be/config.json`

**Step 1: 添加 TokenPay 配置**

```json
{
    "payment": {
        "base_url": "https://pay-test.targeted.work/",
        "merchant_id": "10036",
        "secret_key": "xxx",
        "notify_url": "https://gg.ppnet55.com/api/payments/notify",
        "return_url": "https://www.yourdomain.com/wallet",
        "tokenpay": {
            "base_url": "http://your-tokenpay-server:8080",
            "api_key": "your-api-key",
            "notify_url": "https://your-domain.com/api/payments/tokenpay/notify"
        }
    }
}
```

**注意：** 这不是代码文件，只是示例，不提交到 git。

---

## Task 14: 测试与验证

### 14.1 后端测试

```bash
cd be
go build ./...
```

### 14.2 前端测试

```bash
cd fe
npm run build
```

### 14.3 功能测试清单

- [ ] 获取支付方式列表包含 USDT
- [ ] 选择 USDT 显示预估金额
- [ ] 选择 ETH 链创建订单
- [ ] 选择 TRX 链创建订单
- [ ] 低于最低金额返回错误
- [ ] 显示支付地址和二维码
- [ ] TokenPay 回调更新余额
- [ ] 余额不足拒绝处理

---

## 部署检查清单

### 后端
- [ ] 更新 `config.json` 添加 tokenpay 配置
- [ ] 确保 `/api/payments/tokenpay/notify` 可公网访问
- [ ] 防火墙开放 TokenPay 服务器访问

### 前端
- [ ] 部署新版本前端代码
- [ ] 验证 qrcode.react 正确加载

### TokenPay 服务
- [ ] 配置回调 URL 指向 `/api/payments/tokenpay/notify`
- [ ] 测试 API Key 有效性

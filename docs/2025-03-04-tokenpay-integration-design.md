# TokenPay 支付渠道集成设计方案

## 一、概述

### 1.1 TokenPay 简介
TokenPay 是一个开源的加密货币支付网关，支持自建部署。本项目将集成 TokenPay 以支持 ETH 链（ERC20）和 TRX 链（TRC20）的 USDT 支付。

GitHub: https://github.com/LightCountry/TokenPay

### 1.2 核心特点
- **货币单位**：USDT（美元），与现有 IDR 系统需要汇率转换
- **链支持**：ETH（ERC20）、TRX（TRC20）
- **支付方式**：用户向指定地址转账，TokenPay 检测链上确认后回调
- **自建部署**：通过配置文件接入自建的 TokenPay 服务

## 二、实施方案

### 2.1 方案选型：独立 TokenPay 渠道（推荐）

新增 `TOKENPAY` 支付方式，前端增加链选择（ETH/TRX），后端单独处理 TokenPay 订单创建和回调。

**优势：**
- 代码清晰，不混淆现有逻辑
- 便于后续扩展其他 USDT 支付渠道
- TokenPay 有特殊的回调格式，独立处理更安全

### 2.2 用户支付流程

```
1. 用户选择 TokenPay 支付方式
2. 输入金额（IDR）→ 前端显示预估 USDT 金额
3. 选择链类型（ETH/TRX）
4. 提交订单 → 后端创建 TokenPay 订单
5. 后端调用 TokenPay API，获取支付地址
6. 前端显示支付地址和二维码
7. 用户向地址转账 USDT
8. TokenPay 检测到链上确认，发送回调通知
9. 后端处理回调，更新用户余额（USDT 直接入账）
```

## 三、货币转换逻辑

由于 TokenPay 使用 USDT（美元），而前端输入为 IDR：

```go
// 1. 获取汇率（1 USD = ? IDR）
exchangeRate := models.GetExchangeRate("IDR") // 如 16000

// 2. IDR 金额转换为 USDT
usdtAmount := idrAmount / exchangeRate

// 3. 调用 TokenPay 创建订单（使用 USDT 金额）

// 4. 保存订单时记录两种金额
order.Amount = idrAmount      // 原始 IDR 金额
order.UsdAmount = usdtAmount  // USDT 金额用于显示
order.DstCode = "TOKENPAY_ETH" // 或 "TOKENPAY_TRX"
```

## 四、配置文件变更

### 4.1 be/config.json

```json
{
    "payment": {
        "base_url": "https://pay-test.targeted.work/",
        "merchant_id": "10036",
        "secret_key": "xxx",
        "notify_url": "https://gg.vazhenina.com/api/payments/notify",
        "return_url": "https://www.yourdomain.com/wallet",
        // 新增 TokenPay 配置
        "tokenpay": {
            "base_url": "http://your-tokenpay-server:8080",
            "api_key": "your-api-key",
            "notify_url": "https://your-domain.com/api/payments/tokenpay/notify"
        }
    }
}
```

## 五、后端修改清单

### 5.1 配置文件结构 (helpers/config.go)

新增 TokenPayConfig 结构体：

```go
type PaymentConfig struct {
    BaseURL    string          `json:"base_url"`
    MerchantID string          `json:"merchant_id"`
    SecretKey  string          `json:"secret_key"`
    NotifyURL  string          `json:"notify_url"`
    ReturnURL  string          `json:"return_url"`
    TokenPay   TokenPayConfig  `json:"tokenpay"`  // 新增
}

type TokenPayConfig struct {
    BaseURL   string `json:"base_url"`
    APIKey    string `json:"api_key"`
    NotifyURL string `json:"notify_url"`
}
```

### 5.2 数据模型 (models/dtos/payment_models.go)

新增 TokenPay 相关结构体：

```go
// TokenPayCreateOrderRequest 创建 TokenPay 订单请求
type TokenPayCreateOrderRequest struct {
    OrderID     string  `json:"order_id"`     // 商户订单号
    Amount      float64 `json:"amount"`       // USDT 金额
    ChainType   string  `json:"chain_type"`   // ETH 或 TRX
    NotifyURL   string  `json:"notify_url"`
    ReturnURL   string  `json:"return_url"`
}

// TokenPayCreateOrderResponse TokenPay 创建订单响应
type TokenPayCreateOrderResponse struct {
    OrderID       string `json:"order_id"`
    PayAddress    string `json:"pay_address"`     // 支付地址
    PayAmount     string `json:"pay_amount"`      // 支付金额
    QRCode        string `json:"qr_code"`         // 二维码数据
    ExpiredAt     int64  `json:"expired_at"`      // 过期时间戳
    Status        string `json:"status"`
}

// TokenPayNotifyRequest TokenPay 回调通知
type TokenPayNotifyRequest struct {
    OrderID       string `json:"order_id"`
    TxHash        string `json:"tx_hash"`         // 区块链交易哈希
    Amount        string `json:"amount"`          // 实际支付金额
    Status        string `json:"status"`          // paid / pending / failed
    Chain         string `json:"chain"`           // ETH / TRX
    Confirmations int    `json:"confirmations"`   // 确认数
    Timestamp     int64  `json:"timestamp"`
}
```

### 5.3 支付服务 (services/payment_service.go)

#### 修改 GetPaymentMethods

```go
func (s *PaymentService) GetPaymentMethods(ctx context.Context) ([]dtos.PaymentMethod, error) {
    methods := []dtos.PaymentMethod{
        {Code: "DANA", Name: "DANA", Type: "channel", MinAmount: 10000, MaxAmount: 10000000, Enabled: true},
        // 修改 USDT 为可用状态
        {Code: "USDT", Name: "USDT", Type: "channel", MinAmount: 10, MaxAmount: 100000, Enabled: true}, // USDT 以美元计
        {Code: "PAYPAL", Name: "PayPal", Type: "channel", MinAmount: 10000, MaxAmount: 100000000, Enabled: false},
        {Code: "199VOUCHER", Name: "199 Voucher", Type: "voucher", MinAmount: 10000, MaxAmount: 1000000, Enabled: true},
    }
    return methods, nil
}
```

#### 新增 CreateTokenPayOrder

```go
// CreateTokenPayOrder 创建 TokenPay 订单
func (s *PaymentService) CreateTokenPayOrder(ctx context.Context, userID uint64, req dtos.CreatePaymentRequest, chainType string) (*dtos.TokenPayCreateOrderResponse, error) {
    // 1. 验证 TokenPay 配置
    cfg := s.getConfig()
    if cfg == nil || cfg.TokenPay.BaseURL == "" || cfg.TokenPay.APIKey == "" {
        return nil, errors.New("TokenPay 配置未初始化")
    }

    // 2. 生成订单号
    orderID := common.GenerateOrderID(userID)

    // 3. IDR 转 USDT
    exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
    if err != nil {
        exchangeRate = 16000 // 默认汇率
    }
    usdtAmount := req.Amount / exchangeRate

    // 4. 构建 TokenPay 请求
    tokenPayReq := map[string]interface{}{
        "order_id":   orderID,
        "amount":     fmt.Sprintf("%.2f", usdtAmount),
        "currency":   "USDT",
        "chain":      chainType, // ETH 或 TRX
        "notify_url": cfg.TokenPay.NotifyURL,
        "return_url": req.CallbackURL,
    }

    // 5. 调用 TokenPay API
    apiURL := strings.TrimSuffix(cfg.TokenPay.BaseURL, "/") + "/api/order"
    respData, err := s.postJSON(ctx, apiURL, tokenPayReq, cfg.TokenPay.APIKey)
    if err != nil {
        return nil, fmt.Errorf("TokenPay 请求失败: %w", err)
    }

    // 6. 解析响应
    var tokenPayResp dtos.TokenPayCreateOrderResponse
    if err := json.Unmarshal(respData, &tokenPayResp); err != nil {
        return nil, fmt.Errorf("解析 TokenPay 响应失败: %w", err)
    }

    // 7. 保存订单到数据库
    order := dtos.PaymentOrder{
        UserID:      userID,
        OrderID:     orderID,
        PlatOrderID: "", // TokenPay 在回调时提供
        Amount:      req.Amount,      // 原始 IDR 金额
        UsdAmount:   usdtAmount,      // USDT 金额
        Type:        req.Type,
        DstCode:     "TOKENPAY_" + chainType,
        Status:      0, // 待支付
        PayURL:      "", // TokenPay 使用地址而非 URL
        ProductInfo: "USDT充值",
    }

    if err := s.db.Create(&order).Error; err != nil {
        return nil, fmt.Errorf("保存订单失败: %w", err)
    }

    return &tokenPayResp, nil
}
```

#### 新增 HandleTokenPayNotify

```go
// HandleTokenPayNotify 处理 TokenPay 回调通知
func (s *PaymentService) HandleTokenPayNotify(ctx context.Context, notify dtos.TokenPayNotifyRequest) (string, error) {
    // 1. 验证签名（TokenPay 签名方式需根据实际文档实现）
    // TODO: 根据 TokenPay 文档实现签名验证

    // 2. 查询订单
    var order dtos.PaymentOrder
    if err := s.db.Where("order_id = ?", notify.OrderID).First(&order).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return "FAIL", errors.New("订单不存在")
        }
        return "FAIL", err
    }

    // 3. 检查订单状态
    if order.Status == 1 {
        return "OK", nil // 已处理过
    }

    // 4. 根据状态处理
    now := time.Now()
    switch notify.Status {
    case "paid": // 支付成功
        // TokenPay 直接返回 USDT 金额，无需汇率转换
        actualAmount, _ := strconv.ParseFloat(notify.Amount, 64)

        // 使用订单创建时的 USDT 金额（或实际支付金额）
        usdAmount := order.UsdAmount
        if actualAmount > 0 {
            usdAmount = actualAmount
        }

        // 更新订单和余额
        if err := s.db.Transaction(func(tx *gorm.DB) error {
            // 更新订单
            order.Status = 1
            order.PlatOrderID = notify.TxHash // 使用交易哈希作为平台订单号
            order.PaidAt = &now
            order.UpdatedAt = now
            if err := tx.Save(&order).Error; err != nil {
                return err
            }

            // 更新用户余额（直接加 USDT 金额）
            if err := tx.Model(&dtos.User{}).Where("id = ?", order.UserID).
                Update("balance", gorm.Expr("balance + ?", usdAmount)).Error; err != nil {
                return err
            }

            // 写入资金流水
            transaction := dtos.Transaction{
                UserID:      order.UserID,
                Type:        1, // 充值
                Amount:      usdAmount,
                ReferenceID: order.OrderID,
                Remark:      fmt.Sprintf("USDT充值 (%s) - %s", notify.Chain, notify.TxHash[:min(16, len(notify.TxHash))]),
                CreatedAt:   now,
            }
            if err := tx.Create(&transaction).Error; err != nil {
                return err
            }

            return nil
        }); err != nil {
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
```

### 5.4 控制器 (controllers/payment_handler.go)

#### 新增 CreateTokenPayOrder 接口

```go
// CreateTokenPayOrder 创建 TokenPay 订单
func (h *PaymentHandler) CreateTokenPayOrder(c *fiber.Ctx) error {
    userID := getUserIDFromJWT(c)
    if userID == 0 {
        return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "未登录"})
    }

    var req struct {
        dtos.CreatePaymentRequest
        ChainType string `json:"chain_type"` // ETH 或 TRX
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请求参数错误"})
    }

    if req.ChainType != "ETH" && req.ChainType != "TRX" {
        return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "请选择有效的链类型（ETH 或 TRX）"})
    }

    resp, err := h.paymentService.CreateTokenPayOrder(c.Context(), userID, req.CreatePaymentRequest, req.ChainType)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
    }

    return c.JSON(resp)
}
```

#### 新增 HandleTokenPayNotify 接口

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

    fmt.Println("[TokenPayNotify] Parsed:", notify)

    result, err := h.paymentService.HandleTokenPayNotify(c.Context(), notify)
    if err != nil {
        fmt.Println("[TokenPayNotify] Handle error:", err)
        return c.Status(fiber.StatusOK).SendString(result)
    }

    return c.Status(fiber.StatusOK).SendString(result)
}
```

### 5.5 路由注册 (controllers/router.go)

```go
// TokenPay 支付接口
paymentGroup.Post("/tokenpay/order", paymentHandler.CreateTokenPayOrder)
paymentGroup.Post("/tokenpay/notify", paymentHandler.HandleTokenPayNotify) // 公开接口
```

## 六、前端修改清单

### 6.1 支付服务 (fe/services/payment.ts)

#### 新增 TokenPay 相关类型

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
export interface CreateTokenPayRequest {
  amount: number
  type: string
  dst_code: string
  callback_url?: string
  chain_type: "ETH" | "TRX"  // 链类型
}
```

#### 新增 createTokenPayOrder 方法

```typescript
export const paymentService = {
  // ... 现有方法

  // TokenPay 创建订单
  async createTokenPayOrder(data: CreateTokenPayRequest): Promise<TokenPayCreateOrderResponse> {
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
  },
}
```

### 6.2 存款表单 (fe/components/payment/DepositForm.tsx)

#### 修改 handleSubmit 处理 USDT

```typescript
const handleSubmit = async () => {
  // ... 现有验证逻辑

  // USDT (TokenPay) 特殊处理
  if (selectedMethod.code === "USDT") {
    setShowChainSelector(true) // 显示链选择弹窗
    return
  }

  // ... 其他支付方式处理
}
```

#### 修改金额输入显示

当选择 USDT 时，实时显示预估的 USDT 金额：

```typescript
// 在 DepositForm 中添加汇率转换显示
{selectedMethod?.code === "USDT" && amount && (
  <p className="text-xs text-lucky-gold mt-2">
    预估支付: ~{(parseFloat(amount) / 16000).toFixed(2)} USDT
  </p>
)}
```

### 6.3 新增 TokenPay 弹窗组件 (fe/components/payment/TokenPayModal.tsx)

```typescript
"use client"

import { QRCodeSVG } from "qrcode.react"
import { Copy, Check } from "lucide-react"
import { useState } from "react"

interface TokenPayModalProps {
  isOpen: boolean
  onClose: () => void
  payAddress: string
  payAmount: string
  qrCode: string
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

  if (!isOpen) return null

  const handleCopy = () => {
    navigator.clipboard.writeText(payAddress)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const timeLeft = Math.max(0, Math.floor((expiredAt * 1000 - Date.now()) / 1000 / 60))

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-lucky-dark rounded-2xl p-6 max-w-md w-full">
        <h3 className="text-xl font-bold text-white mb-4 text-center">
          USDT 充值 ({chainType}链)
        </h3>

        {/* 二维码 */}
        <div className="bg-white p-4 rounded-xl flex justify-center mb-4">
          <QRCodeSVG value={payAddress} size={200} />
        </div>

        {/* 支付地址 */}
        <div className="bg-white/10 rounded-lg p-3 mb-4">
          <p className="text-gray-400 text-xs mb-1">支付地址</p>
          <div className="flex items-center gap-2">
            <p className="text-white text-sm font-mono break-all flex-1">{payAddress}</p>
            <button
              onClick={handleCopy}
              className="p-2 bg-lucky-gold/20 rounded-lg hover:bg-lucky-gold/30 transition-colors"
            >
              {copied ? <Check size={18} className="text-green-400" /> : <Copy size={18} className="text-lucky-gold" />}
            </button>
          </div>
        </div>

        {/* 支付金额 */}
        <div className="bg-lucky-gold/20 rounded-lg p-3 mb-4 text-center">
          <p className="text-gray-400 text-xs mb-1">支付金额</p>
          <p className="text-2xl font-bold text-lucky-gold">{payAmount} USDT</p>
        </div>

        {/* 倒计时和提示 */}
        <p className="text-yellow-500 text-sm text-center mb-4">
          请在 {timeLeft} 分钟内完成转账
        </p>
        <p className="text-gray-400 text-xs text-center">
          转账完成后系统将自动确认，请勿关闭此页面
        </p>

        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="w-full mt-4 py-3 bg-white/10 rounded-xl text-white font-bold hover:bg-white/20 transition-colors"
        >
          我已转账
        </button>
      </div>
    </div>
  )
}
```

### 6.4 新增链选择弹窗 (fe/components/payment/ChainSelectorModal.tsx)

```typescript
"use client"

interface ChainSelectorModalProps {
  isOpen: boolean
  onClose: () => void
  onSelect: (chain: "ETH" | "TRX") => void
  amount: number
}

export function ChainSelectorModal({ isOpen, onClose, onSelect, amount }: ChainSelectorModalProps) {
  if (!isOpen) return null

  const usdtAmount = (amount / 16000).toFixed(2)

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-lucky-dark rounded-2xl p-6 max-w-sm w-full">
        <h3 className="text-xl font-bold text-white mb-2 text-center">选择链类型</h3>
        <p className="text-gray-400 text-sm text-center mb-6">
          需支付约 {usdtAmount} USDT
        </p>

        <div className="space-y-3">
          <button
            onClick={() => onSelect("ETH")}
            className="w-full py-4 bg-white/10 rounded-xl flex items-center justify-center gap-3 hover:bg-white/20 transition-colors"
          >
            <span className="text-2xl">⟠</span>
            <span className="text-white font-bold">Ethereum (ERC20)</span>
          </button>

          <button
            onClick={() => onSelect("TRX")}
            className="w-full py-4 bg-white/10 rounded-xl flex items-center justify-center gap-3 hover:bg-white/20 transition-colors"
          >
            <span className="text-2xl">🔴</span>
            <span className="text-white font-bold">Tron (TRC20)</span>
          </button>
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

## 七、数据库变更

### 7.1 payment_orders 表

可能需要添加 `usd_amount` 字段存储 USDT 金额：

```sql
ALTER TABLE payment_orders ADD COLUMN usd_amount DECIMAL(18, 8) DEFAULT 0 AFTER amount;
```

## 八、已确认事项

### 8.1 TokenPay API 格式
参考文档：https://github.com/LightCountry/TokenPay/blob/master/Wiki/docs.md

### 8.2 签名验证方式
TokenPay 使用 API Key + 签名验证，具体实现参考文档。

### 8.3 汇率来源
使用现有 `exchange_rates` 表，默认 1 USD = 15500 IDR。

### 8.4 链选择策略
**用户自选链**：前端提供 ETH 和 TRX 选项。

### 8.5 最低充值金额
| 链类型 | 最低金额 |
|-------|---------|
| ETH (ERC20) | 5 USDT |
| TRX (TRC20) | 10 USDT |

前端需要根据汇率计算对应的最低 IDR 金额显示。

## 九、实施优先级

1. **P0** - 配置文件和基础结构
2. **P0** - TokenPay 创建订单接口（后端）
3. **P0** - TokenPay 回调处理接口（后端）
4. **P1** - 前端链选择和支付弹窗
5. **P1** - 数据库字段调整
6. **P2** - 完善签名验证和错误处理

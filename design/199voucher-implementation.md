# 199 Voucher 支付渠道实施文档

## 一、需求概述

**渠道名称**: 199 Voucher
**渠道类型**: 卡密兑换（演示渠道）
**特殊功能**: 弹窗显示卡密购买网站链接

**业务流程**:
1. 用户选择 "199 Voucher" 支付方式
2. 输入金额，点击充值
3. 弹出卡密输入框，**显示购买卡密网站链接**
4. 用户点击链接跳转购买卡密
5. 返回输入卡密，提交验证
6. 验证成功立即到账（演示模式总是成功）

---

## 二、配置更新

### 2.1 后端配置 (config.dev.json / config.json)

配置文件中已存在（注意拼写为 `vocher`）:

```json
{
    "payment": {
        "base_url": "https://pay-test.targeted.work/",
        "merchant_id": "10036",
        "secret_key": "150kz9esh1s2f793abb8l0e4mjot3r2c",
        "notify_url": "https://gg.vazhenina.com/api/payments/notify",
        "return_url": "https://www.yourdomain.com/wallet"
    },
    "vocher": "https://www.188topup.com"
}
```

### 2.2 后端配置结构 (helpers/config.go)

```go
type Config struct {
    // ... 其他配置 ...
    Payment   PaymentConfig `mapstructure:"payment"`
    VocherURL string        `mapstructure:"vocher"` // 卡密购买网站
}
```

---

## 三、后端实现

### 3.1 DTO 定义 (models/dtos/payment_models.go)

**新增请求/响应结构体**:

```go
// VoucherRedeemRequest 卡密兑换请求
type VoucherRedeemRequest struct {
    Amount      float64 `json:"amount" binding:"required,gt=0"`  // 充值金额(IDR)
    VoucherCode string  `json:"voucher_code" binding:"required"` // 卡密
    DstCode     string  `json:"dst_code" binding:"required"`     // "199VOUCHER"
}

// VoucherRedeemResponse 卡密兑换响应
type VoucherRedeemResponse struct {
    OrderID   string  `json:"order_id"`    // 订单号
    Amount    float64 `json:"amount"`      // 充值金额(IDR)
    UsdAmount float64 `json:"usd_amount"`  // 实际到账(USD)
    Balance   float64 `json:"balance"`     // 当前余额
    Status    int     `json:"status"`      // 1:成功
}

// VoucherConfigResponse 卡密配置响应
type VoucherConfigResponse struct {
    PurchaseURL string `json:"purchase_url"` // 卡密购买网站
}
```

### 3.2 PaymentService 新增方法 (services/payment_service.go)

```go
// GetVoucherConfig 获取卡密配置
func (s *PaymentService) GetVoucherConfig() *dtos.VoucherConfigResponse {
    cfg := s.getConfig()
    return &dtos.VoucherConfigResponse{
        PurchaseURL: cfg.VocherURL,
    }
}

// RedeemVoucher 卡密兑换
func (s *PaymentService) RedeemVoucher(ctx context.Context, userID uint64, req dtos.VoucherRedeemRequest) (*dtos.VoucherRedeemResponse, error) {
    // 1. 参数校验
    if req.VoucherCode == "" {
        return nil, errors.New("请输入卡密")
    }

    // 2. 【演示模式】卡密验证 - 总是成功
    // 生产环境可扩展为真实验证逻辑

    // 3. 汇率转换 IDR -> USD
    exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
    if err != nil {
        exchangeRate = 16000 // 默认汇率
    }
    usdAmount := req.Amount / exchangeRate

    // 4. 生成订单号
    orderID := common.GenerateOrderID(userID)
    now := time.Now()

    // 5. 事务处理
    var newBalance float64
    err = s.db.Transaction(func(tx *gorm.DB) error {
        // 5.1 创建支付订单（直接标记为成功）
        order := dtos.PaymentOrder{
            UserID:      userID,
            OrderID:     orderID,
            PlatOrderID: "VOUCHER-" + req.VoucherCode[:min(8, len(req.VoucherCode))],
            Amount:      req.Amount,
            Type:        "voucher",
            DstCode:     "199VOUCHER",
            Status:      1, // 直接标记为成功
            ProductInfo: "199 Voucher充值",
            PayURL:      "",
            PaidAt:      &now,
            CreatedAt:   now,
            UpdatedAt:   now,
        }
        if err := tx.Create(&order).Error; err != nil {
            return err
        }

        // 5.2 获取用户当前余额
        var user dtos.User
        if err := tx.First(&user, userID).Error; err != nil {
            return err
        }
        beforeBalance := user.Balance
        newBalance = beforeBalance + usdAmount

        // 5.3 更新用户余额
        if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
            return err
        }

        // 5.4 写入资金流水
        transaction := dtos.Transaction{
            UserID:        userID,
            Type:          1, // 充值
            Amount:        usdAmount,
            BeforeBalance: beforeBalance,
            AfterBalance:  newBalance,
            ReferenceID:   orderID,
            Remark:        fmt.Sprintf("199 Voucher充值 (IDR %.0f)", req.Amount),
            CreatedAt:     now,
        }
        if err := tx.Create(&transaction).Error; err != nil {
            return err
        }

        // 5.5 更新用户累计充值
        if err := tx.Exec("UPDATE users SET total_deposit = total_deposit + ? WHERE id = ?", usdAmount, userID).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("兑换失败: %w", err)
    }

    return &dtos.VoucherRedeemResponse{
        OrderID:   orderID,
        Amount:    req.Amount,
        UsdAmount: usdAmount,
        Balance:   newBalance,
        Status:    1,
    }, nil
}
```

### 3.3 PaymentHandler 新增路由 (controllers/payment_handler.go)

```go
// RegisterPaymentRoutes 注册支付路由
func (h *PaymentHandler) RegisterPaymentRoutes(app *fiber.App) {
    payments := app.Group("/api/payments")
    payments.Use(middleware.Auth())

    // ... 原有路由 ...

    // 199 Voucher 路由
    voucher := payments.Group("/voucher")
    {
        voucher.GET("/config", h.GetVoucherConfig)    // 获取卡密配置
        voucher.POST("/redeem", h.RedeemVoucher)      // 卡密兑换
    }
}

// GetVoucherConfig 获取卡密配置
func (h *PaymentHandler) GetVoucherConfig(c *fiber.Ctx) error {
    config := h.service.GetVoucherConfig()
    return c.JSON(config)
}

// RedeemVoucher 卡密兑换
func (h *PaymentHandler) RedeemVoucher(c *fiber.Ctx) error {
    userID := c.Locals("userID").(uint64)

    var req dtos.VoucherRedeemRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
    }

    resp, err := h.service.RedeemVoucher(c.Context(), userID, req)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(resp)
}
```

### 3.4 更新支付方式列表 (services/payment_service.go)

```go
// GetPaymentMethods 获取支持的支付方式列表
func (s *PaymentService) GetPaymentMethods(ctx context.Context) ([]dtos.PaymentMethod, error) {
    methods := []dtos.PaymentMethod{
        {Code: "DANA", Name: "DANA", Type: "channel", MinAmount: 10000, MaxAmount: 10000000, Enabled: true},
        {Code: "USDT", Name: "USDT", Type: "channel", MinAmount: 10000, MaxAmount: 100000000, Enabled: false},
        {Code: "PAYPAL", Name: "PayPal", Type: "channel", MinAmount: 10000, MaxAmount: 100000000, Enabled: false},
        {Code: "199VOUCHER", Name: "199 Voucher", Type: "voucher", MinAmount: 10000, MaxAmount: 1000000, Enabled: true},
    }
    return methods, nil
}
```

---

## 四、前端实现

### 4.1 新增 API 接口 (services/payment.ts)

```typescript
// 卡密兑换请求
export interface RedeemVoucherRequest {
  amount: number
  voucher_code: string
  dst_code: string
}

// 卡密兑换响应
export interface RedeemVoucherResponse {
  order_id: string
  amount: number
  usd_amount: number
  balance: number
  status: number
}

// 卡密配置响应
export interface VoucherConfigResponse {
  purchase_url: string
}

// paymentService 新增方法
export const paymentService = {
  // ... 原有方法 ...

  // 获取卡密配置
  async getVoucherConfig(): Promise<VoucherConfigResponse> {
    const response = await fetchWithAuth(`${API_BASE_URL}/payments/voucher/config`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    })
    if (!response.ok) {
      throw new Error("获取配置失败")
    }
    return response.json()
  },

  // 卡密兑换
  async redeemVoucher(data: RedeemVoucherRequest): Promise<RedeemVoucherResponse> {
    const response = await fetchWithAuth(`${API_BASE_URL}/payments/voucher/redeem`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.error || "兑换失败")
    }

    return response.json()
  }
}
```

### 4.2 新建 VoucherModal 组件 (components/payment/VoucherModal.tsx)

```typescript
"use client"

import { useState, useEffect } from "react"
import { X, Loader2, Gift, ExternalLink, Info } from "lucide-react"
import { paymentService } from "@/services/payment"

interface VoucherModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (voucherCode: string) => Promise<void>
  amount: number
}

export function VoucherModal({ isOpen, onClose, onSubmit, amount }: VoucherModalProps) {
  const [voucherCode, setVoucherCode] = useState("")
  const [loading, setLoading] = useState(false)
  const [purchaseUrl, setPurchaseUrl] = useState("")
  const [configLoading, setConfigLoading] = useState(true)

  // 获取卡密购买网站配置
  useEffect(() => {
    if (isOpen) {
      loadConfig()
    }
  }, [isOpen])

  const loadConfig = async () => {
    setConfigLoading(true)
    try {
      const config = await paymentService.getVoucherConfig()
      setPurchaseUrl(config.purchase_url)
    } catch (error) {
      console.error("获取配置失败:", error)
      // 使用默认链接
      setPurchaseUrl("https://www.188topup.com")
    } finally {
      setConfigLoading(false)
    }
  }

  if (!isOpen) return null

  const handleSubmit = async () => {
    if (!voucherCode.trim()) return
    setLoading(true)
    try {
      await onSubmit(voucherCode.trim())
      setVoucherCode("")
    } finally {
      setLoading(false)
    }
  }

  // 格式化卡密输入：XXXX-XXXX-XXXX-XXXX
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let value = e.target.value.replace(/[^a-zA-Z0-9]/g, "").toUpperCase()
    if (value.length > 16) value = value.slice(0, 16)

    // 添加分隔符
    const parts = []
    for (let i = 0; i < value.length; i += 4) {
      parts.push(value.slice(i, i + 4))
    }
    setVoucherCode(parts.join("-"))
  }

  // 打开购买网站
  const openPurchaseSite = () => {
    if (purchaseUrl) {
      window.open(purchaseUrl, "_blank")
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
      <div className="bg-lucky-dark border border-white/10 rounded-2xl w-full max-w-md mx-4 p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-lucky-gold/20 flex items-center justify-center">
              <Gift className="text-lucky-gold" size={20} />
            </div>
            <div>
              <h3 className="font-bold text-white">199 Voucher</h3>
              <p className="text-xs text-gray-400">卡密充值</p>
            </div>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <X size={20} />
          </button>
        </div>

        {/* 购买提示区域 */}
        <div className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/20 rounded-xl p-4 mb-6">
          <div className="flex items-start gap-3">
            <Info className="text-blue-400 shrink-0 mt-0.5" size={18} />
            <div className="flex-1">
              <p className="text-sm text-gray-300 mb-2">
                没有卡密？点击下方的按钮购买充值卡密
              </p>
              <button
                onClick={openPurchaseSite}
                disabled={configLoading || !purchaseUrl}
                className="flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300 transition-colors"
              >
                <ExternalLink size={14} />
                {configLoading ? "加载中..." : "前往购买卡密"}
              </button>
            </div>
          </div>
        </div>

        {/* Amount Display */}
        <div className="bg-white/5 rounded-xl p-4 mb-6 text-center">
          <div className="text-sm text-gray-400 mb-1">充值金额</div>
          <div className="text-2xl font-bold text-lucky-gold">
            {amount.toLocaleString()} IDR
          </div>
        </div>

        {/* Voucher Code Input */}
        <div className="mb-6">
          <label className="block text-sm text-gray-400 mb-2">卡密</label>
          <input
            type="text"
            value={voucherCode}
            onChange={handleInputChange}
            placeholder="XXXX-XXXX-XXXX-XXXX"
            className="w-full bg-black/30 border border-white/20 rounded-xl px-4 py-4 text-white text-center text-lg tracking-wider font-mono focus:border-lucky-gold focus:outline-none uppercase"
            maxLength={19}
            disabled={loading}
          />
          <p className="text-xs text-gray-500 mt-2 text-center">
            请输入16位卡密，支持自动格式化
          </p>
        </div>

        {/* Submit Button */}
        <button
          onClick={handleSubmit}
          disabled={loading || voucherCode.length < 19}
          className="w-full py-4 rounded-full bg-gradient-to-r from-lucky-gold to-orange-500 text-lucky-dark font-bold text-lg shadow-xl shadow-lucky-gold/20 hover:scale-[1.02] transition-transform disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {loading ? (
            <>
              <Loader2 className="animate-spin" size={20} />
              验证中...
            </>
          ) : (
            "立即兑换"
          )}
        </button>

        {/* Demo Notice */}
        <div className="mt-4 p-3 bg-green-500/10 border border-green-500/20 rounded-lg">
          <p className="text-xs text-green-400 text-center">
            ✅ 演示模式：任意卡密均可兑换成功
          </p>
        </div>
      </div>
    </div>
  )
}
```

### 4.3 修改 DepositForm.tsx

```typescript
// 新增状态
const [showVoucherModal, setShowVoucherModal] = useState(false)

// 处理199 Voucher提交
const handleVoucherSubmit = async (voucherCode: string) => {
  try {
    const response = await paymentService.redeemVoucher({
      amount: numAmount,
      voucher_code: voucherCode,
      dst_code: "199VOUCHER"
    })

    toast({
      title: "兑换成功",
      description: `到账 $${response.usd_amount.toFixed(2)} USD`,
    })

    setShowVoucherModal(false)
    onSuccess?.()
  } catch (error: any) {
    toast({
      title: "兑换失败",
      description: error.message,
      variant: "destructive",
    })
  }
}

// 修改 handleSubmit 函数
const handleSubmit = async () => {
  // ... 原有校验逻辑 ...

  // 199 Voucher 特殊处理
  if (selectedMethod?.code === "199VOUCHER") {
    setShowVoucherModal(true)
    return
  }

  // ... 原有支付流程 ...
}

// 在 return 的 JSX 末尾添加弹窗
return (
  <div className="space-y-6">
    {/* ... 原有 JSX ... */}

    {/* 199 Voucher 弹窗 */}
    {selectedMethod?.code === "199VOUCHER" && (
      <VoucherModal
        isOpen={showVoucherModal}
        onClose={() => setShowVoucherModal(false)}
        onSubmit={handleVoucherSubmit}
        amount={parseFloat(amount) || 0}
      />
    )}
  </div>
)
```

### 4.4 修改 PaymentMethods.tsx（可选，添加 Voucher 图标）

```typescript
// 在渲染方法图标时添加特殊处理
const getMethodIcon = (code: string) => {
  if (code === "199VOUCHER") {
    return <Gift className="text-lucky-gold" size={20} />
  }
  // ... 原有逻辑 ...
}
```

---

## 五、文件变更清单

### 后端修改
| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `be/helpers/config.go` | 修改 | Config 结构体添加 VocherURL 字段 |
| `be/models/dtos/payment_models.go` | 修改 | 新增 VoucherRedeemRequest、VoucherRedeemResponse、VoucherConfigResponse |
| `be/services/payment_service.go` | 修改 | 新增 GetVoucherConfig、RedeemVoucher 方法；更新 GetPaymentMethods |
| `be/controllers/payment_handler.go` | 修改 | 新增 /voucher/config 和 /voucher/redeem 路由 |

### 前端修改
| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `fe/services/payment.ts` | 修改 | 新增 getVoucherConfig、redeemVoucher 方法和类型定义 |
| `fe/components/payment/VoucherModal.tsx` | 新增 | 卡密输入弹窗组件（含购买链接） |
| `fe/components/payment/DepositForm.tsx` | 修改 | 集成 VoucherModal，处理 voucher 类型充值 |
| `fe/components/payment/PaymentMethods.tsx` | 可选修改 | 添加 Voucher 图标样式 |

---

## 六、测试流程

### 6.1 功能测试

1. **选择支付方式**
   - 进入钱包页面
   - 选择 "199 Voucher"
   - 确认显示正确的限额信息

2. **输入金额并点击充值**
   - 输入有效金额（10000 - 1000000 IDR）
   - 点击 "立即充值"
   - 确认弹出 VoucherModal

3. **验证弹窗内容**
   - 显示充值金额
   - 显示 "前往购买卡密" 链接
   - 点击链接跳转到 `https://www.188topup.com`

4. **卡密兑换**
   - 输入任意16位卡密（如：ABCD-EFGH-IJKL-MNOP）
   - 点击 "立即兑换"
   - 确认显示 "兑换成功" 提示
   - 确认余额增加（USD）

5. **验证数据库**
   ```sql
   -- 检查订单记录
   SELECT * FROM payment_orders WHERE dst_code = '199VOUCHER' ORDER BY id DESC LIMIT 1;

   -- 检查交易流水
   SELECT * FROM transactions WHERE type = 1 ORDER BY id DESC LIMIT 1;

   -- 检查用户余额
   SELECT balance FROM users WHERE id = {user_id};
   ```

### 6.2 配置测试

1. **修改配置**
   ```json
   // config.dev.json
   "vocher": "https://www.example-voucher-site.com"
   ```

2. **重启后端**

3. **验证新链接**
   - 重新打开 VoucherModal
   - 点击 "前往购买卡密"
   - 确认跳转到新配置的网址

---

## 七、注意事项

1. **配置字段拼写**
   - 配置文件中字段名为 `vocher`（不是 voucher）
   - 代码中映射使用 `mapstructure:"vocher"`

2. **汇率转换**
   - 用户输入的是 IDR 金额
   - 系统按实时汇率转换为 USD 存入余额
   - 转换失败时使用默认汇率 16000

3. **演示模式**
   - 当前实现任意卡密都验证成功
   - 生产环境可扩展为真实验证逻辑

4. **卡密格式**
   - 支持自动格式化为 XXXX-XXXX-XXXX-XXXX
   - 实际存储时只取前8位作为参考

---

## 八、扩展建议（生产环境）

1. **卡密管理系统**
   - 后台生成/导入卡密
   - 设置卡密面额、有效期、使用次数
   - 追踪卡密使用状态

2. **安全增强**
   - 卡密错误次数限制（如：5次错误锁定1小时）
   - IP/设备绑定
   - 大额充值人工审核

3. **批量兑换**
   - 支持一次输入多个卡密
   - 批量验证和到账

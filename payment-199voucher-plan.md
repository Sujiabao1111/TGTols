# 199 Voucher 演示支付渠道实施方案

## 一、需求概述

**渠道名称**: 199 Voucher
**渠道类型**: 卡密兑换（演示渠道）
**流程特点**:
1. 用户选择渠道 + 输入金额
2. 弹出卡密输入框
3. 提交卡密到后端验证
4. 验证成功立即到账（演示时总是成功）

---

## 二、整体流程图

```
┌─────────────────────────────────────────────────────────────┐
│ 前端 (DepositForm)                                          │
├─────────────────────────────────────────────────────────────┤
│ 1. 选择 "199 Voucher" 支付方式                              │
│ 2. 输入金额                                                 │
│ 3. 点击 "充值"                                              │
│ 4. 弹出卡密输入对话框 (VoucherInputModal)                   │
│ 5. 用户输入卡密 (格式: XXXX-XXXX-XXXX-XXXX)                 │
│ 6. 调用后端 /payments/voucher/redeem                        │
│ 7. 显示处理中 → 成功/失败提示                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 后端 (PaymentService)                                       │
├─────────────────────────────────────────────────────────────┤
│ 1. 接收 voucher_code + amount + user_id                     │
│ 2. 【演示模式】卡密验证直接返回成功                           │
│ 3. 生成订单并立即标记为已支付                                │
│ 4. 更新用户余额（USD）                                       │
│ 5. 记录交易流水                                              │
│ 6. 返回成功响应                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、数据库设计

### 3.1 复用现有表 payment_orders
无需新建表，使用现有 `payment_orders` 表，扩展类型支持：

```sql
-- 在 payment_orders 表中，type 字段支持 "voucher"
-- dst_code 存储 "199VOUCHER"
-- 备注字段记录卡密（可选，演示渠道可不存）
```

### 3.2 可选：创建卡密使用记录表（演示可省略）

```sql
-- 演示阶段可不创建，生产环境如需追踪卡密使用可添加
CREATE TABLE voucher_codes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL UNIQUE COMMENT '卡密',
  amount DECIMAL(15,2) NOT NULL COMMENT '面额(IDR)',
  status TINYINT DEFAULT 0 COMMENT '0:未使用 1:已使用',
  used_by BIGINT DEFAULT NULL COMMENT '使用用户ID',
  used_at DATETIME DEFAULT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 四、后端实现方案

### 4.1 DTO 定义 (payment_models.go)

```go
// VoucherRedeemRequest 卡密兑换请求
type VoucherRedeemRequest struct {
    Amount      float64 `json:"amount" binding:"required,gt=0"`      // 充值金额(IDR)
    VoucherCode string  `json:"voucher_code" binding:"required"`     // 卡密
    DstCode     string  `json:"dst_code" binding:"required"`         // "199VOUCHER"
}

// VoucherRedeemResponse 卡密兑换响应
type VoucherRedeemResponse struct {
    OrderID   string  `json:"order_id"`    // 订单号
    Amount    float64 `json:"amount"`      // 充值金额(IDR)
    UsdAmount float64 `json:"usd_amount"`  // 实际到账(USD)
    Balance   float64 `json:"balance"`     // 当前余额
    Status    int     `json:"status"`      // 1:成功
}
```

### 4.2 接口路由 (payment_handler.go)

```go
// 新增路由
voucher := payments.Group("/voucher")
{
    voucher.POST("/redeem", handler.RedeemVoucher)  // 卡密兑换
}
```

### 4.3 业务逻辑 (payment_service.go)

```go
// RedeemVoucher 卡密兑换
func (s *PaymentService) RedeemVoucher(ctx context.Context, userID uint64, req dtos.VoucherRedeemRequest) (*dtos.VoucherRedeemResponse, error) {
    // 1. 参数校验
    if req.VoucherCode == "" {
        return nil, errors.New("请输入卡密")
    }

    // 2. 【演示模式】卡密验证 - 总是成功
    // 生产环境：调用第三方验证接口或查询数据库
    // if !validateVoucherCode(req.VoucherCode) {
    //     return nil, errors.New("卡密无效或已使用")
    // }

    // 3. 汇率转换 IDR -> USD
    exchangeRate, err := models.GetInstance().GetExchangeRate("IDR")
    if err != nil {
        exchangeRate = 16000  // 默认汇率
    }
    usdAmount := req.Amount / exchangeRate

    // 4. 生成订单号
    orderID := common.GenerateOrderID(userID)
    now := time.Now()

    // 5. 事务处理：创建订单 + 更新余额
    var newBalance float64
    err = s.db.Transaction(func(tx *gorm.DB) error {
        // 5.1 创建支付订单（直接标记为成功）
        order := dtos.PaymentOrder{
            UserID:      userID,
            OrderID:     orderID,
            PlatOrderID: "VOUCHER-" + req.VoucherCode[:8],  // 记录卡密前8位
            Amount:      req.Amount,
            Type:        "voucher",
            DstCode:     "199VOUCHER",
            Status:      1,  // 直接标记为成功
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
            Type:          1,  // 充值
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

---

## 五、前端实现方案

### 5.1 新增支付方式

在 `PaymentMethods.tsx` 和 `paymentService.getPaymentMethods()` 中添加：

```typescript
{
  code: "199VOUCHER",
  name: "199 Voucher",
  type: "voucher",
  min_amount: 10000,
  max_amount: 10000000,
  enabled: true
}
```

### 5.2 新增卡密输入弹窗组件 (VoucherModal.tsx)

```typescript
"use client"

import { useState } from "react"
import { X, Loader2, Gift } from "lucide-react"

interface VoucherModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (voucherCode: string) => Promise<void>
  amount: number
}

export function VoucherModal({ isOpen, onClose, onSubmit, amount }: VoucherModalProps) {
  const [voucherCode, setVoucherCode] = useState("")
  const [loading, setLoading] = useState(false)

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
              <p className="text-xs text-gray-400">输入卡密完成充值</p>
            </div>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <X size={20} />
          </button>
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
        <div className="mt-4 p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg">
          <p className="text-xs text-blue-400 text-center">
            💡 演示模式：任意卡密均可兑换成功
          </p>
        </div>
      </div>
    </div>
  )
}
```

### 5.3 修改 DepositForm.tsx

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
  if (selectedMethod.code === "199VOUCHER") {
    setShowVoucherModal(true)
    return
  }

  // ... 原有支付流程 ...
}

// 在 return 中添加弹窗
{
  selectedMethod?.code === "199VOUCHER" && (
    <VoucherModal
      isOpen={showVoucherModal}
      onClose={() => setShowVoucherModal(false)}
      onSubmit={handleVoucherSubmit}
      amount={parseFloat(amount)}
    />
  )
}
```

### 5.4 payment.ts 新增接口

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

// paymentService 新增方法
export const paymentService = {
  // ... 原有方法 ...

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

---

## 六、文件修改清单

### 后端
| 文件 | 修改内容 |
|------|----------|
| `be/models/dtos/payment_models.go` | 新增 VoucherRedeemRequest/VoucherRedeemResponse |
| `be/services/payment_service.go` | 新增 RedeemVoucher 方法 |
| `be/controllers/payment_handler.go` | 新增 /voucher/redeem 路由处理 |
| `be/services/payment_service.go` | GetPaymentMethods 添加 199VOUCHER |

### 前端
| 文件 | 修改内容 |
|------|----------|
| `fe/services/payment.ts` | 新增 redeemVoucher 方法和类型定义 |
| `fe/components/payment/VoucherModal.tsx` | 新建卡密输入弹窗组件 |
| `fe/components/payment/DepositForm.tsx` | 集成 VoucherModal，处理 voucher 类型 |
| `fe/components/payment/PaymentMethods.tsx` | 199VOUCHER 添加图标样式 |

---

## 七、测试流程

1. **前端选择 199 Voucher** → 显示卡密输入框
2. **输入任意16位卡密** → 点击兑换
3. **后端验证** → 【演示】总是返回成功
4. **余额立即更新** → 显示到账 USD 金额
5. **查看订单记录** → 显示 199 Voucher 充值记录

---

## 八、后续扩展（生产环境）

1. **卡密管理系统**
   - 后台生成/导入卡密
   - 设置卡密面额、有效期
   - 追踪卡密使用状态

2. **真实验证逻辑**
   ```go
   func validateVoucherCode(code string) (bool, float64, error) {
       var voucher VoucherCode
       err := db.Where("code = ? AND status = 0", code).First(&voucher).Error
       if err != nil {
           return false, 0, errors.New("卡密无效或已使用")
       }
       return true, voucher.Amount, nil
   }
   ```

3. **安全增强**
   - 卡密输入错误次数限制
   - 卡密绑定IP/设备
   - 延迟到账风控审核

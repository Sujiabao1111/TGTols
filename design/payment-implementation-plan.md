# 支付对接实施计划

## 1. 项目概述

### 1.1 对接平台
印尼第三方支付平台，提供代收（收银台支付）、代付（提款）服务

### 1.2 支付方式
| 类型 | 编码 | 说明 | 金额范围 |
|------|------|------|----------|
| bank | PERMATA, BNI, CIMB, BRI, MANDIRI等 | 银行VA | 10,000 ~ 50,000,000 IDR |
| ewallet | DANA, OVO, LINKAJA | 电子钱包 | 10,000 ~ 20,000,000 IDR |
| qris | QRIS | 扫码支付 | 10,000 ~ 10,000,000 IDR |

### 1.3 核心流程
1. **收银台支付（代收）**：用户选择金额 → 后端创建订单 → 返回支付链接 → 用户支付 → 平台回调通知 → 更新余额
2. **代付（提款）**：用户发起提款 → 后端创建代付订单 → 平台处理 → 回调通知（待定需求）

---

## 2. 技术方案

### 2.1 签名算法（MD5）
```
1. 过滤空值 ("", 0, "0", null, false)
2. 按键名ASCII码排序
3. 拼接成 key1=value1&key2=value2 格式
4. 末尾追加 &key=商户密钥
5. MD5加密并转大写
```

### 2.2 请求格式
- **Content-Type**: `multipart/form-data`
- **Method**: POST
- **编码**: UTF-8

### 2.3 回调响应
- 必须返回纯文本: `OK` 或 `SUCCESS`
- 平台会在12小时内重试5次

---

## 实施状态

### 已完成
- [x] 数据库迁移脚本 (`be/migrations/migration_0204.sql`)
- [x] Config 结构体扩展 (`be/helpers/config.go`)
- [x] 签名工具 (`be/utils/sign.go`)
- [x] 支付模型 (`be/models/dtos/payment_models.go`)
- [x] User 模型扩展 (`be/models/dtos/models.gen.go`)
- [x] 支付服务 (`be/services/payment_service.go`)
- [x] 支付处理器 (`be/controllers/payment_handler.go`)
- [x] 路由配置 (`be/controllers/router.go`)

### 待完成
- [ ] config.json 配置更新
- [ ] 前端支付组件
- [ ] 联调测试

---

## 3. 后端实施计划（Go）

### 3.1 项目结构

基于现有项目结构，支付模块添加以下文件：

```
be/
├── helpers/
│   ├── config.go               # 扩展 Config 结构体，添加 PaymentConfig
│   └── configParser.go         # 现有配置解析器（无需修改）
├── services/
│   └── payment_service.go      # 支付服务核心逻辑（新增）
├── controllers/
│   └── payment_controller.go   # HTTP接口处理器（新增）
├── models/
│   ├── payment_order.go        # 支付订单模型（新增）
│   └── dtos/
│       └── payment_dto.go      # 支付相关DTO（新增）
├── utils/
│   ├── sign.go                 # MD5签名工具（新增）
│   └── setup.go                # 现有初始化逻辑（无需修改）
├── config.json                 # 生产环境配置（扩展）
├── config.dev.json             # 开发环境配置（扩展）
└── main.go                     # 现有入口（无需修改）
```

### 3.2 配置扩展

#### 3.2.1 扩展 Config 结构体 (`helpers/config.go`)

```go
package helpers

// Config 全局配置结构体
type Config struct {
    // 现有配置
    LocalURL  string
    Port      int
    DbDsn     string
    PoolIdle  int
    PoolMax   int
    Jwt       string
    RedisNode string
    RedisDb   int
    Agentid   string
    Agentapi  string
    Prefix    string

    // 新增支付配置
    Payment PaymentConfig
}

// PaymentConfig 支付相关配置
type PaymentConfig struct {
    MchID           string `json:"mchId"`           // 商户号
    MchKey          string `json:"mchKey"`          // 商户密钥
    BaseURL         string `json:"baseUrl"`         // 支付平台基础URL
    NotifyBaseURL   string `json:"notifyBaseUrl"`   // 我方回调基础URL
    CallbackBaseURL string `json:"callbackBaseUrl"` // 前端回调基础URL
    Timeout         int    `json:"timeout"`         // 请求超时(秒)
}
```

#### 3.2.2 数据库表设计

**支付订单表 (payment_orders)**
```sql
CREATE TABLE `payment_orders` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `order_id` VARCHAR(64) NOT NULL UNIQUE COMMENT '商户订单号',
    `plat_order_id` VARCHAR(64) DEFAULT NULL COMMENT '平台订单号',
    `amount` BIGINT NOT NULL COMMENT '金额（印尼盾）',
    `cost` BIGINT DEFAULT 0 COMMENT '手续费',
    `type` VARCHAR(20) NOT NULL COMMENT '支付类型: bank/ewallet/qris',
    `dst_code` VARCHAR(50) NOT NULL COMMENT '支付方式编码',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0:待支付 1:成功 2:失败 3:处理中',
    `ref_code` INT DEFAULT NULL COMMENT '业务状态码',
    `ref_msg` VARCHAR(255) DEFAULT NULL COMMENT '业务描述',
    `pay_url` VARCHAR(500) DEFAULT NULL COMMENT '支付链接',
    `product_info` VARCHAR(255) DEFAULT NULL COMMENT '产品信息',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `paid_at` DATETIME DEFAULT NULL COMMENT '支付完成时间',
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单表';
```

**用户余额表扩展 (user_balances)**
```sql
-- 已有表，需确保有以下字段
ALTER TABLE `user_balances`
ADD COLUMN `currency` VARCHAR(10) DEFAULT 'IDR' COMMENT '币种',
ADD COLUMN `total_deposit` BIGINT DEFAULT 0 COMMENT '累计充值',
ADD COLUMN `total_withdraw` BIGINT DEFAULT 0 COMMENT '累计提现';
```

### 3.3 Go代码实现

#### 3.3.1 签名工具 (utils/sign.go)
```go
package utils

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "sort"
    "strings"
)

// GenerateSign 生成MD5签名
func GenerateSign(params map[string]string, key string) string {
    // 1. 过滤空值并排序
    var keys []string
    for k, v := range params {
        if isEmptyValue(v) {
            continue
        }
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // 2. 拼接字符串
    var parts []string
    for _, k := range keys {
        parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
    }
    stringA := strings.Join(parts, "&")

    // 3. 追加key
    stringSignTemp := stringA + "&key=" + key

    // 4. MD5并转大写
    hash := md5.Sum([]byte(stringSignTemp))
    return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// VerifySign 验证签名
func VerifySign(params map[string]string, key, sign string) bool {
    return GenerateSign(params, key) == sign
}

func isEmptyValue(v string) bool {
    return v == "" || v == "0" || v == "null" || v == "false"
}
```

#### 3.3.2 支付服务 (services/payment_service.go)
```go
package services

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "time"

    "gorm.io/gorm"
    "your-project/helpers"
    "your-project/models"
    "your-project/utils"
)

type PaymentService struct {
    db     *gorm.DB
    client *http.Client
}

func NewPaymentService(db *gorm.DB) *PaymentService {
    return &PaymentService{
        db:     db,
        client: &http.Client{Timeout: time.Duration(helpers.GetCfgInstance().Conf.Payment.Timeout) * time.Second},
    }
}

// CreatePaymentOrder 创建收银台支付订单
func (s *PaymentService) CreatePaymentOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
    cfg := helpers.GetCfgInstance().Conf.Payment

    // 1. 生成订单号
    orderID := generateOrderID()

    // 2. 构建请求参数
    params := map[string]string{
        "memberId":    cfg.MchID,
        "orderId":     orderID,
        "amount":      fmt.Sprintf("%d", req.Amount),
        "dstCode":     req.DstCode,
        "type":        req.Type,
        "dateTime":    time.Now().Format("2006-01-02 15:04:05"),
        "name":        req.Name,
        "phone":       req.Phone,
        "email":       req.Email,
        "notifyUrl":   cfg.NotifyBaseURL + "/api/payment/notify",
        "callbackUrl": cfg.CallbackBaseURL + "/wallet",
        "productInfo": "Game Recharge",
        "version":     "2", // 返回dstCode
    }

    // 3. 生成签名
    params["sign"] = utils.GenerateSign(params, cfg.MchKey)

    // 4. 发送请求到支付平台
    payURL := cfg.BaseURL + "/pay_counter.html"
    resp, err := s.postForm(payURL, params)
    if err != nil {
        return nil, err
    }

    // 5. 解析响应
    var payResp PaymentResponse
    if err := json.Unmarshal(resp, &payResp); err != nil {
        return nil, err
    }

    if payResp.Status != "success" {
        return nil, errors.New(payResp.Msg)
    }

    // 6. 保存订单到数据库
    order := &models.PaymentOrder{
        UserID:      req.UserID,
        OrderID:     orderID,
        Amount:      req.Amount,
        Type:        req.Type,
        DstCode:     req.DstCode,
        Status:      0, // 待支付
        PayURL:      payResp.PayURL,
        ProductInfo: params["productInfo"],
    }

    if err := s.db.Create(order).Error; err != nil {
        return nil, err
    }

    return &CreateOrderResponse{
        OrderID: orderID,
        PayURL:  payResp.PayURL,
        Amount:  req.Amount,
    }, nil
}

// HandlePaymentNotify 处理支付回调通知
func (s *PaymentService) HandlePaymentNotify(ctx context.Context, params map[string]string) error {
    cfg := helpers.GetCfgInstance().Conf.Payment

    // 1. 验证签名
    if !utils.VerifySign(params, cfg.MchKey, params["sign"]) {
        return errors.New("签名验证失败")
    }

    // 2. 查询订单
    var order models.PaymentOrder
    if err := s.db.Where("order_id = ?", params["orderId"]).First(&order).Error; err != nil {
        return err
    }

    // 3. 幂等性检查
    if order.Status == 1 {
        return nil // 已处理过
    }

    // 4. 更新订单状态
    refCode := parseInt(params["refCode"])
    status := parseStatus(refCode)

    updates := map[string]interface{}{
        "status":        status,
        "plat_order_id": params["platOrderId"],
        "ref_code":      refCode,
        "ref_msg":       params["refMsg"],
        "cost":          parseInt(params["cost"]),
    }

    if status == 1 {
        updates["paid_at"] = time.Now()
    }

    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 更新订单
    if err := tx.Model(&order).Updates(updates).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 5. 更新用户余额（仅在支付成功时）
    if status == 1 {
        if err := s.updateUserBalance(tx, order.UserID, order.Amount); err != nil {
            tx.Rollback()
            return err
        }

        // 记录余额变动
        if err := s.createBalanceLog(tx, order); err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit().Error
}

// postForm 发送form-data请求
func (s *PaymentService) postForm(url string, params map[string]string) ([]byte, error) {
    var b bytes.Buffer
    w := multipart.NewWriter(&b)

    for key, val := range params {
        if err := w.WriteField(key, val); err != nil {
            return nil, err
        }
    }
    w.Close()

    req, err := http.NewRequest("POST", url, &b)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", w.FormDataContentType())

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

### 3.4 路由配置

已在 `be/controllers/router.go` 中添加支付路由：

**公开路由（无需JWT）：**
```go
// 支付回调通知
r.Post("/payments/notify", paymentHandler.HandlePaymentNotify)
```

**JWT保护路由：**
```go
// 支付方式
r.Get("/payments/methods", paymentHandler.GetPaymentMethods)

// 代收 (充值)
r.Post("/payments/order", paymentHandler.CreatePaymentOrder)
r.Get("/payments/order/:orderId", paymentHandler.GetPaymentStatus)
r.Get("/payments/orders", paymentHandler.GetUserPayments)

// 代付 (提现)
r.Post("/withdraw", paymentHandler.CreateWithdrawOrder)
```

### 3.5 API接口定义

#### 3.5.1 获取支付方式列表
```
GET /payments/methods

Response:
{
    "methods": [
        {
            "type": "ewallet",
            "name": "DANA",
            "code": "DANA",
            "icon": "/icons/dana.png",
            "min_amount": 10000,
            "max_amount": 10000000
        }
    ]
}
```

#### 3.5.2 创建支付订单（充值）
```
POST /payments/order
Authorization: Bearer {token}

Request:
{
    "amount": 100000,        // 金额（印尼盾）
    "type": "ewallet",       // bank/ewallet/qris
    "dst_code": "DANA"       // 支付方式编码
}

Response:
{
    "order_id": "ORD202402041200001234",
    "pay_url": "http://pay.xxx.com/pay_cashier.html?sn=xxx",
    "amount": 100000,
    "status": 0,
    "expired_at": "2024-02-04 12:30:00"
}
```

#### 3.5.3 查询订单状态
```
GET /payments/order/:orderId
Authorization: Bearer {token}

Response:
{
    "order_id": "ORD202402041200001234",
    "plat_order_id": "PLAT123456",
    "amount": 100000,
    "status": 1,              // 0:待支付 1:成功 2:失败 3:处理中
    "type": "ewallet",
    "dst_code": "DANA",
    "paid_at": "2024-02-04 12:05:00",
    "created_at": "2024-02-04 12:00:00"
}
```

#### 3.5.4 获取用户支付记录
```
GET /payments/orders?limit=20
Authorization: Bearer {token}

Response:
[
    {
        "id": 1,
        "order_id": "ORD202402041200001234",
        "amount": 100000,
        "status": 1,
        "type": "ewallet",
        "dst_code": "DANA",
        "created_at": "2024-02-04T12:00:00Z"
    }
]
```

#### 3.5.5 创建代付订单（提现）
```
POST /withdraw
Authorization: Bearer {token}

Request:
{
    "amount": 100000,
    "type": "bankcard",      // bankcard/ewallet
    "dst_code": "BCA",
    "account": "1234567890",
    "account_name": "John Doe",
    "phone": "8123456789",
    "email": "user@example.com",
    "address": "Jakarta"
}

Response:
{
    "id": 1,
    "order_id": "WDR202402041200001234",
    "amount": 100000,
    "status": 0
}
```

#### 3.5.6 支付回调通知（由支付平台调用）
```
POST /payments/notify
Content-Type: application/x-www-form-urlencoded

Parameters:
- merchant_id: 商户号
- order_id: 商户订单号
- plat_order_id: 平台订单号
- amount: 订单金额
- status: 状态 (1成功 2失败 3处理中)
- ref_code: 业务状态码
- ref_msg: 业务描述
- sign: MD5签名

Response:
OK 或 SUCCESS
```

---

## 4. 前端实施计划（Next.js）

### 4.1 项目结构
```
fe/
├── app/
│   └── (protected)/
│       └── wallet/
│           ├── page.tsx              # 钱包页面
│           └── components/
│               ├── PaymentMethods.tsx    # 支付方式选择
│               ├── DepositForm.tsx       # 充值表单
│               ├── PaymentModal.tsx      # 支付确认弹框
│               └── PaymentHistory.tsx    # 支付记录
├── services/
│   └── payment.ts                    # 支付API服务
├── hooks/
│   └── usePayment.ts                 # 支付相关hooks
└── types/
    └── payment.ts                    # 支付类型定义
```

### 4.2 核心组件实现

#### 4.2.1 支付方式选择 (PaymentMethods.tsx)
```tsx
"use client"

import { useState } from 'react'
import Image from 'next/image'

interface PaymentMethod {
    type: string
    code: string
    name: string
    icon: string
    minAmount: number
    maxAmount: number
}

const PAYMENT_METHODS: PaymentMethod[] = [
    { type: 'ewallet', code: 'DANA', name: 'DANA', icon: '/icons/dana.svg', minAmount: 10000, maxAmount: 20000000 },
    { type: 'ewallet', code: 'OVO', name: 'OVO', icon: '/icons/ovo.svg', minAmount: 10000, maxAmount: 20000000 },
    { type: 'ewallet', code: 'LINKAJA', name: 'LinkAja', icon: '/icons/linkaja.svg', minAmount: 10000, maxAmount: 20000000 },
    { type: 'bank', code: 'BCA', name: 'BCA', icon: '/icons/bca.svg', minAmount: 10000, maxAmount: 50000000 },
    { type: 'bank', code: 'BNI', name: 'BNI', icon: '/icons/bni.svg', minAmount: 10000, maxAmount: 50000000 },
    { type: 'qris', code: 'QRIS', name: 'QRIS', icon: '/icons/qris.svg', minAmount: 10000, maxAmount: 10000000 },
]

interface PaymentMethodsProps {
    selectedMethod: PaymentMethod | null
    onSelect: (method: PaymentMethod) => void
}

export function PaymentMethods({ selectedMethod, onSelect }: PaymentMethodsProps) {
    return (
        <div className="grid grid-cols-3 gap-3">
            {PAYMENT_METHODS.map((method) => (
                <button
                    key={method.code}
                    onClick={() => onSelect(method)}
                    className={`flex flex-col items-center p-4 rounded-xl border transition-all ${
                        selectedMethod?.code === method.code
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-gray-300'
                    }`}
                >
                    <Image
                        src={method.icon}
                        alt={method.name}
                        width={48}
                        height={48}
                        className="mb-2"
                    />
                    <span className="text-sm font-medium">{method.name}</span>
                    <span className="text-xs text-gray-500">
                        {method.minAmount.toLocaleString()} - {method.maxAmount.toLocaleString()}
                    </span>
                </button>
            ))}
        </div>
    )
}
```

#### 4.2.2 充值表单 (DepositForm.tsx)
```tsx
"use client"

import { useState } from 'react'
import { useAuth } from '@/app/contexts/AuthContext'
import { paymentService } from '@/services/payment'
import { PaymentMethods } from './PaymentMethods'

const QUICK_AMOUNTS = [50000, 100000, 200000, 500000, 1000000]

export function DepositForm() {
    const { user } = useAuth()
    const [amount, setAmount] = useState('')
    const [selectedMethod, setSelectedMethod] = useState(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')

    const handleSubmit = async () => {
        if (!selectedMethod || !amount) return

        setLoading(true)
        setError('')

        try {
            const numAmount = parseInt(amount)

            // 验证金额范围
            if (numAmount < selectedMethod.minAmount || numAmount > selectedMethod.maxAmount) {
                setError(`金额范围: ${selectedMethod.minAmount.toLocaleString()} - ${selectedMethod.maxAmount.toLocaleString()}`)
                return
            }

            // 创建订单
            const response = await paymentService.createOrder({
                amount: numAmount,
                type: selectedMethod.type,
                dstCode: selectedMethod.code,
                name: user?.name || 'Guest',
                phone: user?.phone?.replace(/^0/, '8') || '',
                email: user?.email || ''
            })

            // 跳转到支付页面
            if (response.data?.payUrl) {
                window.location.href = response.data.payUrl
            }
        } catch (err: any) {
            setError(err.message || '创建订单失败')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="space-y-6">
            <div>
                <label className="block text-sm font-medium mb-2">选择支付方式</label>
                <PaymentMethods
                    selectedMethod={selectedMethod}
                    onSelect={setSelectedMethod}
                />
            </div>

            <div>
                <label className="block text-sm font-medium mb-2">充值金额 (IDR)</label>
                <div className="grid grid-cols-5 gap-2 mb-3">
                    {QUICK_AMOUNTS.map((amt) => (
                        <button
                            key={amt}
                            onClick={() => setAmount(amt.toString())}
                            className={`py-2 rounded-lg border text-sm ${
                                amount === amt.toString()
                                    ? 'border-blue-500 bg-blue-50'
                                    : 'border-gray-200 hover:border-gray-300'
                            }`}
                        >
                            {amt.toLocaleString()}
                        </button>
                    ))}
                </div>
                <input
                    type="number"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    placeholder="输入金额"
                    className="w-full px-4 py-3 border rounded-lg"
                />
            </div>

            {error && (
                <div className="text-red-500 text-sm">{error}</div>
            )}

            <button
                onClick={handleSubmit}
                disabled={!selectedMethod || !amount || loading}
                className="w-full py-3 bg-blue-500 text-white rounded-lg disabled:opacity-50"
            >
                {loading ? '处理中...' : '立即充值'}
            </button>
        </div>
    )
}
```

#### 4.2.3 支付服务 (services/payment.ts)
```typescript
import axios from 'axios'

const API_BASE = process.env.NEXT_PUBLIC_API_URL

export interface CreateOrderRequest {
    amount: number
    type: string
    dstCode: string
    name: string
    phone: string
    email: string
}

export interface CreateOrderResponse {
    orderId: string
    payUrl: string
    amount: number
}

export const paymentService = {
    // 创建支付订单
    async createOrder(data: CreateOrderRequest) {
        const response = await axios.post(`${API_BASE}/api/payment/create`, data, {
            headers: {
                Authorization: `Bearer ${localStorage.getItem('token')}`
            }
        })
        return response.data
    },

    // 查询订单状态
    async getOrderStatus(orderId: string) {
        const response = await axios.get(`${API_BASE}/api/payment/order/${orderId}`, {
            headers: {
                Authorization: `Bearer ${localStorage.getItem('token')}`
            }
        })
        return response.data
    },

    // 获取支付方式列表
    async getPaymentMethods() {
        const response = await axios.get(`${API_BASE}/api/payment/methods`)
        return response.data
    }
}
```

### 4.3 钱包页面更新

更新 `fe/app/(protected)/wallet/page.tsx`，集成支付功能：

```tsx
"use client"

import { useState } from 'react'
import { DepositForm } from './components/DepositForm'
import { WithdrawForm } from './components/WithdrawForm'
import { PaymentHistory } from './components/PaymentHistory'

type TabType = 'deposit' | 'withdraw' | 'history'

export default function WalletPage() {
    const [activeTab, setActiveTab] = useState<TabType>('deposit')

    return (
        <div className="pt-16 pb-24 px-4 max-w-4xl mx-auto">
            <h1 className="text-2xl font-bold mb-6">钱包</h1>

            {/* Tab切换 */}
            <div className="flex gap-4 mb-6 border-b">
                <button
                    onClick={() => setActiveTab('deposit')}
                    className={`pb-2 px-4 ${activeTab === 'deposit' ? 'border-b-2 border-blue-500' : ''}`}
                >
                    充值
                </button>
                <button
                    onClick={() => setActiveTab('withdraw')}
                    className={`pb-2 px-4 ${activeTab === 'withdraw' ? 'border-b-2 border-blue-500' : ''}`}
                >
                    提现
                </button>
                <button
                    onClick={() => setActiveTab('history')}
                    className={`pb-2 px-4 ${activeTab === 'history' ? 'border-b-2 border-blue-500' : ''}`}
                >
                    记录
                </button>
            </div>

            {/* 内容区域 */}
            {activeTab === 'deposit' && <DepositForm />}
            {activeTab === 'withdraw' && <WithdrawForm />}
            {activeTab === 'history' && <PaymentHistory />}
        </div>
    )
}
```

---

## 5. 配置管理

### 5.1 后端配置 (config.json)

根据现有配置系统，在 `config.json` 和 `config.dev.json` 中添加支付配置：

```json
{
    "localURL": "127.0.0.1",
    "port": 3555,
    "dbDsn": "root:password@tcp(127.0.0.1:3306)/star?charset=utf8mb4&parseTime=True&loc=Local",
    "poolIdle": 10,
    "poolMax": 100,
    "jwt": "abcedfghijk",
    "redisNode": "127.0.0.1:6379",
    "redisDb": 0,
    "agentid": "xxx",
    "agentapi": "https://api.xxx.xyz/",
    "prefix": "plat",

    "payment": {
        "base_url": "https://pay.xxxxxx.com",
        "merchant_id": "10085",
        "secret_key": "slrkogmvo13hb92d8i5111bxm8q8euc7",
        "notify_url": "https://api.yourdomain.com/payments/notify",
        "return_url": "https://www.yourdomain.com/wallet"
    }
}
```

**字段说明：**
| 字段 | 说明 |
|------|------|
| `base_url` | 支付网关基础URL |
| `merchant_id` | 商户号 (mch_id) |
| `secret_key` | 商户密钥 (用于签名) |
| `notify_url` | 回调通知地址 (后端接口) |
| `return_url` | 支付完成跳转地址 (前端页面) |

### 5.2 配置读取方式

使用现有配置系统读取支付配置：

```go
import "your-project/helpers"

// 在任何服务中获取支付配置
func SomeFunction() {
    cfg := helpers.GetCfgInstance().Conf.Payment

    mchID := cfg.MchID
    mchKey := cfg.MchKey
    baseURL := cfg.BaseURL
    // ...
}
```

### 5.3 环境切换

通过 `.env` 文件切换开发和生产环境：

```bash
# .env 文件
# dev=1 使用 config.dev.json
dev=1

# dev=0 或不存在 使用 config.json
dev=0
```

### 5.4 配置设计说明

**为什么选择 config.json 而不是数据库存储？**

1. **与现有系统一致**：项目已使用 viper + config.json 管理配置，支付配置遵循相同模式
2. **环境隔离**：通过 `config.json` 和 `config.dev.json` 天然支持多环境
3. **无需额外查询**：配置随应用启动加载，无需数据库查询
4. **敏感信息安全**：生产环境密钥仅保存在服务器配置文件中，不进入版本控制
5. **热重载支持**：可通过 `helpers.GetCfgInstance().Reload()` 动态重载配置

---

## 6. 安全考虑

### 6.1 签名验证
- 所有回调必须验证签名
- 密钥不传输，仅用于服务端签名计算
- 空值不参与签名

### 6.2 订单安全
- 订单号全局唯一，格式: `PAY` + 时间 + 随机数
- 金额验证：防止篡改
- 用户ID绑定：防止越权

### 6.3 回调安全
- IP白名单验证（如有）
- 幂等性处理：同一订单多次回调只处理一次
- 响应时间控制：避免超时导致重复通知

### 6.4 数据安全
- 敏感信息加密存储
- HTTPS 传输
- 日志脱敏：日志中不记录完整密钥

---

## 7. 测试计划

### 7.1 单元测试
- 签名生成/验证测试
- 订单状态流转测试
- 回调处理测试

### 7.2 集成测试
- 创建订单 → 查询订单 → 回调通知完整流程
- 并发订单测试
- 异常场景测试（签名错误、金额篡改等）

### 7.3 沙箱测试
使用支付平台提供的测试环境：
```bash
# 测试配置
PAYMENT_BASE_URL=https://test-pay.xxxxxx.com
PAYMENT_MCH_ID=test_mch_id
PAYMENT_MCH_KEY=test_key
```

### 7.4 验收测试用例
| 用例 | 预期结果 |
|------|----------|
| 正常创建DANA订单 | 返回支付链接，可正常跳转 |
| 支付成功后回调 | 余额增加，订单状态更新为成功 |
| 重复回调 | 只处理一次，返回OK |
| 签名错误 | 拒绝处理，返回错误 |
| 金额超出范围 | 创建订单失败，返回错误提示 |

---

## 8. 部署计划

### 8.1 部署清单
1. 数据库迁移脚本执行
2. 后端服务部署
3. 回调接口域名配置
4. SSL证书检查
5. 环境变量配置
6. 前端重新构建部署

### 8.2 监控告警
- 订单创建成功率监控
- 回调处理延迟监控
- 支付成功率统计
- 异常订单告警

### 8.3 回滚方案
- 数据库备份
- 配置开关：可快速关闭支付功能
- 版本回滚机制

---

## 9. 时间计划

| 阶段 | 任务 | 状态 | 预计时间 |
|------|------|------|----------|
| 第1天 | 后端：数据库表、签名工具、基础服务 | ✅ 完成 | 1天 |
| 第2天 | 后端：API接口、回调处理、配置更新 | ✅ 完成 | 1天 |
| 第3天 | 前端：支付组件、钱包页面集成 | 🔄 待开始 | 1天 |
| 第4天 | 联调测试、沙箱验证 | ⏳ 待开始 | 1天 |
| 第5天 | 部署上线、监控配置 | ⏳ 待开始 | 1天 |

**已完成：**
- 数据库迁移脚本 (`migration_0204.sql`)
- 后端支付服务完整实现
- API路由配置

**下一步：**
1. 更新 `config.json` 添加支付配置
2. 前端支付组件开发
3. 联调测试

---

## 10. 相关文档

- [支付平台对接文档](./payment.txt)
- [Portal弹框使用指南](../devlog/portal-modal-usage.md)
- [Layout路由问题总结](../devlog/layout-routing-issue.md)

---

## 11. 风险与注意事项

1. **签名问题**：确保空值过滤逻辑与支付平台一致
2. **时区问题**：使用印尼时间（WIB, UTC+7）
3. **金额精度**：印尼盾无小数，全部使用整数
4. **手机号格式**：去掉0前缀，使用8开头
5. **回调可靠性**：必须返回OK/SUCCESS，否则平台会重试

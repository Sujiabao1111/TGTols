package dtos

import (
	"time"
)

// PaymentOrder 支付订单表
type PaymentOrder struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64     `gorm:"not null;index:idx_user_id" json:"user_id"`
	OrderID     string     `gorm:"size:64;not null;uniqueIndex:idx_order_id" json:"order_id"`
	PlatOrderID string     `gorm:"size:64;default:''" json:"plat_order_id"`
	Amount      float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Cost        float64    `gorm:"type:decimal(15,2);default:0.00" json:"cost"`
	Type        string     `gorm:"size:20;not null" json:"type"`
	DstCode     string     `gorm:"size:50;not null" json:"dst_code"`
	Status      int        `gorm:"not null;default:0;index:idx_status" json:"status"`
	RefCode     int        `gorm:"default:0" json:"ref_code"`
	RefMsg      string     `gorm:"size:255;default:''" json:"ref_msg"`
	PayURL      string     `gorm:"type:text" json:"pay_url"`
	ProductInfo string     `gorm:"size:255;default:''" json:"product_info"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (PaymentOrder) TableName() string {
	return "payment_orders"
}

// WithdrawOrder 提现订单表
type WithdrawOrder struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64     `gorm:"not null;index:idx_user_id" json:"user_id"`
	OrderID      string     `gorm:"size:64;not null;uniqueIndex:idx_order_id" json:"order_id"`
	PlatOrderID  string     `gorm:"size:64;default:''" json:"plat_order_id"`
	Amount       float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Cost         float64    `gorm:"type:decimal(15,2);default:0.00" json:"cost"`
	Type         string     `gorm:"size:20;not null" json:"type"`
	DstCode      string     `gorm:"size:50;not null" json:"dst_code"`
	Account      string     `gorm:"size:50;not null" json:"account"`
	AccountName  string     `gorm:"size:100;not null" json:"account_name"`
	Phone        string     `gorm:"size:20;not null" json:"phone"`
	Email        string     `gorm:"size:100;not null" json:"email"`
	Address      string     `gorm:"size:255;default:''" json:"address"`
	Status       int        `gorm:"not null;default:0;index:idx_status" json:"status"`
	RefCode      int        `gorm:"default:0" json:"ref_code"`
	RefMsg       string     `gorm:"size:255;default:''" json:"ref_msg"`
	ReviewedBy   uint64     `gorm:"default:0" json:"reviewed_by"`
	Reviewer     string     `gorm:"size:100;default:''" json:"reviewer"`
	ReviewRemark string     `gorm:"size:255;default:''" json:"review_remark"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (WithdrawOrder) TableName() string {
	return "withdraw_orders"
}

// CreatePaymentRequest 创建支付订单请求
type CreatePaymentRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required"`
	DstCode     string  `json:"dst_code"`
	CallbackURL string  `json:"callback_url"`
	Currency    string  `json:"currency"`
	Channel     string  `json:"channel"`
	ClientIP    string  `json:"client_ip"`
}

// CreatePaymentResponse 创建支付订单响应
type CreatePaymentResponse struct {
	OrderID   string  `json:"order_id"`
	PayURL    string  `json:"pay_url"`
	Amount    float64 `json:"amount"`
	Status    int     `json:"status"`
	ExpiredAt string  `json:"expired_at"`
}

// PaymentStatusResponse 支付状态
type PaymentStatusResponse struct {
	OrderID     string  `json:"order_id"`
	PlatOrderID string  `json:"plat_order_id"`
	Amount      float64 `json:"amount"`
	Status      int     `json:"status"`
	Type        string  `json:"type"`
	DstCode     string  `json:"dst_code"`
	PaidAt      *string `json:"paid_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type WalletRecord struct {
	ID          string  `json:"id"`
	RecordType  string  `json:"record_type"`
	Title       string  `json:"title"`
	OrderID     string  `json:"order_id"`
	ReferenceID string  `json:"reference_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      int     `json:"status"`
	DstCode     string  `json:"dst_code"`
	Remark      string  `json:"remark"`
	CreatedAt   string  `json:"created_at"`
}

// PaymentMethod 支付方式
type PaymentMethod struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Icon        string  `json:"icon"`
	MinAmount   float64 `json:"min_amount"`
	MaxAmount   float64 `json:"max_amount"`
	Enabled     bool    `json:"enabled"`
	Currency    string  `json:"currency"`
	Channel     string  `json:"channel"`
	Description string  `json:"description"`
}

type PaymentMethodsResponse struct {
	Methods []PaymentMethod `json:"methods"`
}

// PaymentNotifyRequest 支付回调
type PaymentNotifyRequest struct {
	MerchantID      string  `form:"merchant_id" json:"merchant_id"`
	MerchantNo      string  `form:"merchantNo" json:"merchantNo"`
	OrderID         string  `form:"order_id" json:"order_id"`
	MerchantOrderNo string  `form:"merchantOrderNo" json:"merchantOrderNo"`
	PlatOrderID     string  `form:"plat_order_id" json:"plat_order_id"`
	OrderNo         string  `form:"orderNo" json:"orderNo"`
	Amount          float64 `form:"amount" json:"amount"`
	Cost            string  `form:"cost" json:"cost"`
	Status          string  `form:"status" json:"status"`
	RefCode         int     `form:"ref_code" json:"ref_code"`
	RefCodeAlt      int     `form:"refCode" json:"refCode"`
	RefMsg          string  `form:"ref_msg" json:"ref_msg"`
	RefMsgAlt       string  `form:"refMsg" json:"refMsg"`
	Sign            string  `form:"sign" json:"sign"`
	Currency        string  `form:"currency" json:"currency"`
	Code            string  `form:"code" json:"code"`
	RawAmount       string  `form:"raw_amount" json:"raw_amount"`
}

func (p *PaymentNotifyRequest) IsSuccess() bool {
	return p.Status == "success" || p.Status == "SUCCESS" || p.Status == "1"
}

// CreateWithdrawRequest 创建提现订单请求
type CreateWithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=bankcard ewallet"`
	DstCode     string  `json:"dst_code" binding:"required"`
	Account     string  `json:"account" binding:"required"`
	AccountName string  `json:"account_name" binding:"required"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Address     string  `json:"address"`
	Currency    string  `json:"currency"`
	Channel     string  `json:"channel"`
	BankCode    string  `json:"bank_code"`
	ClientIP    string  `json:"client_ip"`
}

type AdminApproveWithdrawRequest struct {
	OrderID    string `json:"order_id"`
	ReviewerID uint64 `json:"reviewer_id"`
	Reviewer   string `json:"reviewer"`
	Remark     string `json:"remark"`
}

// WithdrawMethod 提现方式
type WithdrawMethod struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Currency    string  `json:"currency"`
	Channel     string  `json:"channel"`
	MinAmount   float64 `json:"min_amount"`
	MaxAmount   float64 `json:"max_amount"`
	Enabled     bool    `json:"enabled"`
	Description string  `json:"description"`
}

type WithdrawMethodsResponse struct {
	Methods []WithdrawMethod `json:"methods"`
}

type WithdrawNotifyRequest struct {
	MerchantID      string `form:"merchant_id" json:"merchant_id"`
	MerchantNo      string `form:"merchantNo" json:"merchantNo"`
	OrderID         string `form:"order_id" json:"order_id"`
	MerchantOrderNo string `form:"merchantOrderNo" json:"merchantOrderNo"`
	PlatOrderID     string `form:"plat_order_id" json:"plat_order_id"`
	OrderNo         string `form:"orderNo" json:"orderNo"`
	Amount          string `form:"amount" json:"amount"`
	Cost            string `form:"cost" json:"cost"`
	Status          string `form:"status" json:"status"`
	RefCode         string `form:"ref_code" json:"ref_code"`
	RefCodeAlt      string `form:"refCode" json:"refCode"`
	RefMsg          string `form:"ref_msg" json:"ref_msg"`
	RefMsgAlt       string `form:"refMsg" json:"refMsg"`
	ErrorMsg        string `form:"errorMsg" json:"errorMsg"`
	ErrorMsgAlt     string `form:"error_msg" json:"error_msg"`
	Currency        string `form:"currency" json:"currency"`
	Code            string `form:"code" json:"code"`
	Sign            string `form:"sign" json:"sign"`
}

// UserBalanceResponse 用户余额响应
type UserBalanceResponse struct {
	Balance       float64 `json:"balance"`
	Currency      string  `json:"currency"`
	TotalDeposit  float64 `json:"total_deposit"`
	TotalWithdraw float64 `json:"total_withdraw"`
}

// VoucherRedeemRequest 卡密兑换请求
type VoucherRedeemRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	VoucherCode string  `json:"voucher_code" binding:"required"`
	DstCode     string  `json:"dst_code" binding:"required"`
}

// VoucherRedeemResponse 卡密兑换响应
type VoucherRedeemResponse struct {
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	UsdAmount float64 `json:"usd_amount"`
	Balance   float64 `json:"balance"`
	Status    int     `json:"status"`
}

type VoucherConfigResponse struct {
	PurchaseURL string `json:"purchase_url"`
}

// TokenPayCreateOrderRequest TokenPay 订单请求
type TokenPayCreateOrderRequest struct {
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	ChainType string  `json:"chain_type"`
	NotifyURL string  `json:"notify_url"`
	ReturnURL string  `json:"return_url"`
}

// TokenPayCreateOrderResponse TokenPay 订单响应
type TokenPayCreateOrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
	Info    struct {
		ActualAmount   string `json:"ActualAmount"`
		Amount         string `json:"Amount"`
		BaseCurrency   string `json:"BaseCurrency"`
		BlockChainName string `json:"BlockChainName"`
		CurrencyName   string `json:"CurrencyName"`
		ExpireTime     string `json:"ExpireTime"`
		Id             string `json:"Id"`
		OrderUserKey   string `json:"OrderUserKey"`
		OutOrderId     string `json:"OutOrderId"`
		QrCodeBase64   string `json:"QrCodeBase64"`
		QrCodeLink     string `json:"QrCodeLink"`
		ToAddress      string `json:"ToAddress"`
	} `json:"info"`

	OrderID    string `json:"order_id"`
	PayAddress string `json:"pay_address"`
	PayAmount  string `json:"pay_amount"`
	QRCode     string `json:"qr_code"`
	ExpiredAt  int64  `json:"expired_at"`
	Status     string `json:"status"`
}

// TokenPayNotifyRequest TokenPay 回调
type TokenPayNotifyRequest struct {
	Id                 string `json:"Id"`
	OutOrderId         string `json:"OutOrderId"`
	OrderUserKey       string `json:"OrderUserKey"`
	BlockTransactionId string `json:"BlockTransactionId"`
	PayTime            string `json:"PayTime"`
	BlockchainName     string `json:"BlockchainName"`
	Currency           string `json:"Currency"`
	CurrencyName       string `json:"CurrencyName"`
	BaseCurrency       string `json:"BaseCurrency"`
	Amount             string `json:"Amount"`
	ActualAmount       string `json:"ActualAmount"`
	FromAddress        string `json:"FromAddress"`
	ToAddress          string `json:"ToAddress"`
	Status             int    `json:"Status"`
	Signature          string `json:"Signature"`
	PassThroughInfo    string `json:"PassThroughInfo"`
}

func (t *TokenPayNotifyRequest) IsPaid() bool {
	return t.Status == 1
}

// CreateTokenPayOrderRequest 前端 TokenPay 下单请求
type CreateTokenPayOrderRequest struct {
	CreatePaymentRequest
	ChainType string `json:"chain_type"`
}

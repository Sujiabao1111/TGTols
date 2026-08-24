package example

import "time"

type RechargeQueryRecord struct {
	ID                  uint64     `json:"id"`
	UserID              uint64     `json:"userId"`
	Username            string     `json:"username"`
	OrderID             string     `json:"orderId"`
	PlatOrderID         string     `json:"platOrderId"`
	LocalAmount         float64    `json:"localAmount"`
	Currency            string     `json:"currency"`
	RechargeAmountU     float64    `json:"rechargeAmountU"`
	TotalRechargeAmount float64    `json:"totalRechargeAmount"`
	Type                string     `json:"type"`
	DstCode             string     `json:"dstCode"`
	PaidAt              *time.Time `json:"paidAt"`
	CreatedAt           time.Time  `json:"createdAt"`
}

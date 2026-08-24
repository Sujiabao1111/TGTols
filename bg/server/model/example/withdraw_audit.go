package example

import "time"

// WithdrawAuditRecord is the admin-facing withdraw order row.
type WithdrawAuditRecord struct {
	ID           uint64     `json:"id"`
	UserID       uint64     `json:"userId"`
	Username     string     `json:"username"`
	OrderID      string     `json:"orderId"`
	PlatOrderID  string     `json:"platOrderId"`
	Amount       float64    `json:"amount"`
	Cost         float64    `json:"cost"`
	Type         string     `json:"type"`
	DstCode      string     `json:"dstCode"`
	Account      string     `json:"account"`
	AccountName  string     `json:"accountName"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	Status       int        `json:"status"`
	RefCode      int        `json:"refCode"`
	RefMsg       string     `json:"refMsg"`
	Reviewer     string     `json:"reviewer"`
	ReviewRemark string     `json:"reviewRemark"`
	ReviewedAt   *time.Time `json:"reviewedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

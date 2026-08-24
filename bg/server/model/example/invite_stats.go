package example

import "time"

// InviteStats is the admin-facing invitee detail row.
type InviteStats struct {
	InviterUserID       uint64     `json:"inviterUserId"`
	UserID              uint64     `json:"userId" gorm:"column:user_id"`
	Username            string     `json:"username" gorm:"column:username"`
	Balance             float64    `json:"balance" gorm:"column:balance"`
	RechargeCount       int64      `json:"rechargeCount" gorm:"column:recharge_count"`
	RechargeAmount      float64    `json:"rechargeAmount" gorm:"column:recharge_amount"`
	CreatedAt           time.Time  `json:"createdAt" gorm:"column:created_at"`
	RegisterIP          string     `json:"registerIp" gorm:"column:register_ip"`
	RegisterDomain      string     `json:"registerDomain" gorm:"column:register_domain"`
	SameRegisterIPCount int64      `json:"sameRegisterIpCount"`
	LastLoginAt         *time.Time `json:"lastLoginAt" gorm:"column:last_login_at"`
	LastLoginIP         string     `json:"lastLoginIp" gorm:"column:last_login_ip"`
	IPCountryNote       string     `json:"ipCountryNote"`
}

type InviteStatsSummary struct {
	RegisterCount     int64   `json:"registerCount"`
	RechargeUserCount int64   `json:"rechargeUserCount"`
	RechargeAmountU   float64 `json:"rechargeAmountU"`
}

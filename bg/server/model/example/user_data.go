package example

import "time"

// UserDataRecord is the admin-facing single-user data snapshot.
type UserDataRecord struct {
	UserID              uint64     `json:"userId"`
	Username            string     `json:"username"`
	Balance             float64    `json:"balance"`
	TotalTurnover       float64    `json:"totalTurnover"`
	TotalWinlose        float64    `json:"totalWinlose"`
	RechargeCount       int64      `json:"rechargeCount"`
	RechargeAmount      float64    `json:"rechargeAmount"`
	WithdrawAmount      float64    `json:"withdrawAmount"`
	TotalGameCount      int64      `json:"totalGameCount"`
	SlotGameCount       int64      `json:"slotGameCount"`
	CasinoGameCount     int64      `json:"casinoGameCount"`
	SportbookGameCount  int64      `json:"sportbookGameCount"`
	OtherGameCount      int64      `json:"otherGameCount"`
	CreatedAt           time.Time  `json:"createdAt"`
	RegisterIP          string     `json:"registerIp"`
	RegisterDomain      string     `json:"registerDomain"`
	SameRegisterIPCount int64      `json:"sameRegisterIpCount"`
	LastLoginAt         *time.Time `json:"lastLoginAt"`
	LastLoginIP         string     `json:"lastLoginIp"`
	IPCountryNote       string     `json:"ipCountryNote"`
}

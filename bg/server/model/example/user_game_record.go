package example

import "time"

// UserGameRecord is the admin-facing single game transaction row.
type UserGameRecord struct {
	ID           uint64     `json:"id"`
	UserID       uint64     `json:"userId"`
	Username     string     `json:"username"`
	PlatformCode string     `json:"platformCode"`
	ProviderID   int        `json:"providerId"`
	ProviderCode string     `json:"providerCode"`
	ProviderName string     `json:"providerName"`
	GameCode     string     `json:"gameCode"`
	GameNameCN   string     `json:"gameNameCn"`
	GameNameEN   string     `json:"gameNameEn"`
	ExternalID   string     `json:"externalId"`
	Turnover     float64    `json:"turnover"`
	Bet          float64    `json:"bet"`
	Win          float64    `json:"win"`
	WinLose      float64    `json:"winLose"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	StatDate     string     `json:"statDate"`
	CreatedAt    time.Time  `json:"createdAt"`
}

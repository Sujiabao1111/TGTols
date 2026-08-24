package example

import "time"

// GameLaunchErrorRecord is the admin-facing third-party game launch error row.
type GameLaunchErrorRecord struct {
	ID           uint64    `json:"id"`
	UserID       uint64    `json:"userId"`
	Username     string    `json:"username"`
	GameCode     string    `json:"gameCode"`
	GameName     string    `json:"gameName"`
	ProviderCode string    `json:"providerCode"`
	ProviderName string    `json:"providerName"`
	IsLobby      bool      `json:"isLobby"`
	IsMobile     bool      `json:"isMobile"`
	Language     string    `json:"language"`
	ErrorType    string    `json:"errorType"`
	ErrorMessage string    `json:"errorMessage"`
	PageURL      string    `json:"pageUrl"`
	GameURLHost  string    `json:"gameUrlHost"`
	ClientIP     string    `json:"clientIp"`
	UserAgent    string    `json:"userAgent"`
	CreatedAt    time.Time `json:"createdAt"`
}

package dtos

import "time"

// UserGameLaunchError stores failed third-party game launch attempts reported by players.
type UserGameLaunchError struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"not null;index:idx_ugle_user_created;index:idx_ugle_user_type_created" json:"user_id"`
	Username     string    `gorm:"size:100;default:''" json:"username"`
	GameCode     string    `gorm:"size:100;default:''" json:"game_code"`
	GameName     string    `gorm:"size:255;default:''" json:"game_name"`
	ProviderCode string    `gorm:"size:50;default:''" json:"provider_code"`
	ProviderName string    `gorm:"size:100;default:''" json:"provider_name"`
	IsLobby      bool      `gorm:"not null;default:false" json:"is_lobby"`
	IsMobile     bool      `gorm:"not null;default:false" json:"is_mobile"`
	Language     string    `gorm:"size:20;default:''" json:"language"`
	ErrorType    string    `gorm:"size:64;not null;index:idx_ugle_user_type_created" json:"error_type"`
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	PageURL      string    `gorm:"type:text" json:"page_url"`
	GameURLHost  string    `gorm:"size:255;default:''" json:"game_url_host"`
	ClientIP     string    `gorm:"size:64;default:''" json:"client_ip"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	CreatedAt    time.Time `gorm:"index:idx_ugle_user_created;index:idx_ugle_user_type_created" json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (UserGameLaunchError) TableName() string {
	return "user_game_launch_errors"
}

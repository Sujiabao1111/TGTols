package dtos

import "time"

type UserGameTransactionStat struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"not null;uniqueIndex:uk_user_period_key;index:idx_user_game_transaction_stats_user" json:"user_id"`
	PeriodType  string    `gorm:"size:16;not null;uniqueIndex:uk_user_period_key;index:idx_user_game_transaction_stats_period_type" json:"period_type"`
	PeriodKey   string    `gorm:"size:32;not null;uniqueIndex:uk_user_period_key" json:"period_key"`
	PeriodStart time.Time `gorm:"not null;index:idx_user_game_transaction_stats_period_start" json:"period_start"`
	PeriodEnd   time.Time `gorm:"not null" json:"period_end"`
	LoginID     string    `gorm:"size:100;default:''" json:"login_id"`
	AmountUnit  string    `gorm:"size:8;default:''" json:"amount_unit"`
	Count       int       `gorm:"default:0" json:"count"`
	Turnover    float64   `gorm:"type:decimal(15,2);default:0" json:"turnover"`
	Bet         float64   `gorm:"type:decimal(15,2);default:0" json:"bet"`
	Win         float64   `gorm:"type:decimal(15,2);default:0" json:"win"`
	Winlose     float64   `gorm:"type:decimal(15,2);default:0" json:"winlose"`
	JPShare     float64   `gorm:"type:decimal(15,2);default:0" json:"jp_share"`
	JPWin       float64   `gorm:"type:decimal(15,2);default:0" json:"jp_win"`
	SyncedAt    time.Time `json:"synced_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (UserGameTransactionStat) TableName() string {
	return "user_game_transaction_stats"
}

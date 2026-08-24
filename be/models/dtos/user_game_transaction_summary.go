package dtos

import "time"

// UserGameTransactionSummary 用户第三方日汇总
type UserGameTransactionSummary struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_user_stat_date" json:"user_id"`
	StatDate  string    `gorm:"type:date;not null;uniqueIndex:uk_user_stat_date" json:"stat_date"`
	LoginID   string    `gorm:"size:100;default:''" json:"login_id"`
	Count     int       `gorm:"default:0" json:"count"`
	Turnover  float64   `gorm:"type:decimal(15,2);default:0" json:"turnover"`
	Bet       float64   `gorm:"type:decimal(15,2);default:0" json:"bet"`
	Win       float64   `gorm:"type:decimal(15,2);default:0" json:"win"`
	Winlose   float64   `gorm:"type:decimal(15,2);default:0" json:"winlose"`
	JPShare   float64   `gorm:"type:decimal(15,2);default:0" json:"jp_share"`
	JPWin     float64   `gorm:"type:decimal(15,2);default:0" json:"jp_win"`
	SyncedAt  time.Time `json:"synced_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserGameTransactionSummary) TableName() string {
	return "user_game_transaction_summaries"
}

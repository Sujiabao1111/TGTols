package dtos

import "time"

// UserGameTransactionDetail 用户游戏逐笔流水
type UserGameTransactionDetail struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64     `gorm:"not null;index:idx_user_stat_date_start" json:"user_id"`
	StatDate         string     `gorm:"type:date;not null;index:idx_user_stat_date_start" json:"stat_date"`
	AgentID          string     `gorm:"size:50;default:''" json:"agent_id"`
	ProviderUserID   string     `gorm:"size:100;default:''" json:"provider_user_id"`
	LoginID          string     `gorm:"size:100;default:'';index:idx_login_stat_date" json:"login_id"`
	GameProviderCode int        `gorm:"not null;index:idx_provider_external" json:"game_provider_code"`
	GameCode         string     `gorm:"size:100;default:'';index:idx_user_stat_date_start" json:"game_code"`
	ExternalID       string     `gorm:"size:100;default:'';index:idx_provider_external" json:"external_id"`
	RoundID          string     `gorm:"size:100;default:''" json:"round_id"`
	Type             int        `gorm:"default:0" json:"type"`
	StartDate        *time.Time `gorm:"index:idx_user_stat_date_start" json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	StartBalance     float64    `gorm:"type:decimal(15,2);default:0" json:"start_balance"`
	EndBalance       float64    `gorm:"type:decimal(15,2);default:0" json:"end_balance"`
	Deposit          float64    `gorm:"type:decimal(15,2);default:0" json:"deposit"`
	Turnover         float64    `gorm:"type:decimal(15,2);default:0" json:"turnover"`
	Bet              float64    `gorm:"type:decimal(15,2);default:0" json:"bet"`
	Win              float64    `gorm:"type:decimal(15,2);default:0" json:"win"`
	Winlose          float64    `gorm:"type:decimal(15,2);default:0" json:"winlose"`
	JPShare          float64    `gorm:"type:decimal(15,2);default:0" json:"jp_share"`
	JPWin            float64    `gorm:"type:decimal(15,2);default:0" json:"jp_win"`
	Remark           string     `gorm:"type:text" json:"remark"`
	RecordHash       string     `gorm:"size:64;not null;uniqueIndex:uk_record_hash" json:"record_hash"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (UserGameTransactionDetail) TableName() string {
	return "user_game_transaction_details"
}

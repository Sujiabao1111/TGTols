package dtos

import "time"

type DailyWeeklyChallengeProgress struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64     `gorm:"not null;uniqueIndex:uk_user_challenge_period_task;index:idx_daily_weekly_challenge_user" json:"user_id"`
	CycleType       string     `gorm:"size:16;not null;uniqueIndex:uk_user_challenge_period_task;index:idx_daily_weekly_challenge_cycle" json:"cycle_type"`
	TaskIndex       int        `gorm:"not null;uniqueIndex:uk_user_challenge_period_task" json:"task_index"`
	PeriodStart     time.Time  `gorm:"not null;uniqueIndex:uk_user_challenge_period_task;index:idx_daily_weekly_challenge_period" json:"period_start"`
	PeriodEnd       time.Time  `gorm:"not null" json:"period_end"`
	RequiredBetU    float64    `gorm:"type:decimal(15,2);default:0.00" json:"required_bet_u"`
	RewardAmountU   float64    `gorm:"type:decimal(15,2);default:0.00" json:"reward_amount_u"`
	RequiredBetIDR  float64    `gorm:"type:decimal(15,2);default:0.00" json:"required_bet_idr"`
	RewardAmountIDR float64    `gorm:"type:decimal(15,2);default:0.00" json:"reward_amount_idr"`
	RequiredBetPHP  float64    `gorm:"type:decimal(15,2);default:0.00" json:"required_bet_php"`
	RewardAmountPHP float64    `gorm:"type:decimal(15,2);default:0.00" json:"reward_amount_php"`
	IDRRate         float64    `gorm:"type:decimal(15,4);default:0.00" json:"idr_rate"`
	PHPRate         float64    `gorm:"type:decimal(15,4);default:0.00" json:"php_rate"`
	Claimed         bool       `gorm:"default:false;index:idx_daily_weekly_challenge_claimed" json:"claimed"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (DailyWeeklyChallengeProgress) TableName() string {
	return "daily_weekly_challenge_progress"
}

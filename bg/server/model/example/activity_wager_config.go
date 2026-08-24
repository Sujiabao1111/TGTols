package example

import "time"

type ActivityWagerConfig struct {
	ID                     uint64    `json:"id" form:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ConfigKey              string    `json:"configKey" form:"configKey" gorm:"size:50;not null;uniqueIndex:idx_activity_wager_config_key;column:config_key"`
	DepositWagerMultiplier float64   `json:"depositWagerMultiplier" form:"depositWagerMultiplier" gorm:"type:decimal(10,2);not null;default:2.00;column:deposit_wager_multiplier"`
	RewardWagerMultiplier  float64   `json:"rewardWagerMultiplier" form:"rewardWagerMultiplier" gorm:"type:decimal(10,2);not null;default:20.00;column:reward_wager_multiplier"`
	CreatedAt              time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at"`
	UpdatedAt              time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at"`
}

func (ActivityWagerConfig) TableName() string {
	return "activity_wager_configs"
}

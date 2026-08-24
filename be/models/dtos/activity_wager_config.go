package dtos

import "time"

// ActivityWagerConfig stores configurable wager multipliers for deposit and reward funds.
type ActivityWagerConfig struct {
	ID                     uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey              string    `gorm:"size:50;not null;uniqueIndex:idx_activity_wager_config_key" json:"config_key"`
	DepositWagerMultiplier float64   `gorm:"type:decimal(10,2);not null;default:2.00" json:"deposit_wager_multiplier"`
	RewardWagerMultiplier  float64   `gorm:"type:decimal(10,2);not null;default:20.00" json:"reward_wager_multiplier"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (ActivityWagerConfig) TableName() string {
	return "activity_wager_configs"
}

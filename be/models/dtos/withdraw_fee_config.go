package dtos

import "time"

type WithdrawFeeConfig struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey string    `gorm:"size:50;not null;uniqueIndex:idx_withdraw_fee_config_key" json:"config_key"`
	Enabled   bool      `gorm:"not null;default:true" json:"enabled"`
	FeeRate   float64   `gorm:"type:decimal(10,6);not null;default:0.003000" json:"fee_rate"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WithdrawFeeConfig) TableName() string {
	return "withdraw_fee_configs"
}

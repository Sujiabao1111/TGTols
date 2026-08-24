package example

import "time"

type WithdrawFeeConfig struct {
	ID        uint64    `json:"id" form:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ConfigKey string    `json:"configKey" form:"configKey" gorm:"size:50;not null;uniqueIndex:idx_withdraw_fee_config_key;column:config_key"`
	Enabled   bool      `json:"enabled" form:"enabled" gorm:"not null;default:true;column:enabled"`
	FeeRate   float64   `json:"feeRate" form:"feeRate" gorm:"type:decimal(10,6);not null;default:0.003000;column:fee_rate"`
	CreatedAt time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at"`
}

func (WithdrawFeeConfig) TableName() string {
	return "withdraw_fee_configs"
}

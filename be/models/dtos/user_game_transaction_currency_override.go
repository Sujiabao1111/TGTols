package dtos

import "time"

type UserGameTransactionCurrencyOverride struct {
	UserID    uint64    `gorm:"primaryKey" json:"user_id"`
	Currency  string    `gorm:"size:8;not null" json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserGameTransactionCurrencyOverride) TableName() string {
	return "user_game_transaction_currency_overrides"
}

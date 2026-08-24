package dtos

import "time"

// UserWagerLock records reward-related wager requirements that must be cleared before withdrawal.
type UserWagerLock struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64    `gorm:"not null;index:idx_user_wager_lock_user" json:"user_id"`
	SourceType      string    `gorm:"size:50;not null;index:idx_user_wager_lock_source" json:"source_type"`
	ReferenceID     string    `gorm:"size:100;not null;uniqueIndex:idx_user_wager_lock_reference" json:"reference_id"`
	RewardAmountU   float64   `gorm:"type:decimal(15,2);default:0.00" json:"reward_amount_u"`
	WagerMultiplier float64   `gorm:"type:decimal(10,2);default:0.00" json:"wager_multiplier"`
	WagerRequired   float64   `gorm:"type:decimal(15,2);default:0.00" json:"wager_required"`
	GrantedAt       time.Time `gorm:"index:idx_user_wager_lock_granted" json:"granted_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (UserWagerLock) TableName() string {
	return "user_wager_locks"
}

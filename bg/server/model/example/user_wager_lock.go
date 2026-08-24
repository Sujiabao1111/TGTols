package example

import "time"

type UserWagerLock struct {
	ID              uint64    `json:"id" gorm:"primarykey;column:id"`
	UserID          uint64    `json:"userId" gorm:"column:user_id;not null;index"`
	SourceType      string    `json:"sourceType" gorm:"column:source_type;size:50;not null;index"`
	ReferenceID     string    `json:"referenceId" gorm:"column:reference_id;size:100;not null;uniqueIndex"`
	RewardAmountU   float64   `json:"rewardAmountU" gorm:"column:reward_amount_u;type:decimal(15,2);default:0.00"`
	WagerMultiplier float64   `json:"wagerMultiplier" gorm:"column:wager_multiplier;type:decimal(10,2);default:0.00"`
	WagerRequired   float64   `json:"wagerRequired" gorm:"column:wager_required;type:decimal(15,2);default:0.00"`
	GrantedAt       time.Time `json:"grantedAt" gorm:"column:granted_at;index"`
	CreatedAt       time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (UserWagerLock) TableName() string {
	return "user_wager_locks"
}

package example

import "time"

type DesktopRewardClaim struct {
	ID        uint64    `json:"id" gorm:"primarykey;column:id"`
	UserID    uint64    `json:"userId" gorm:"column:user_id;not null;uniqueIndex"`
	ClaimIP   *string   `json:"claimIp" gorm:"column:claim_ip;size:64;uniqueIndex"`
	ClaimedAt time.Time `json:"claimedAt" gorm:"column:claimed_at;not null;index"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (DesktopRewardClaim) TableName() string {
	return "desktop_reward_claims"
}

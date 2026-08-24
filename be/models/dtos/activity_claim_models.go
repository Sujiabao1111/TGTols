package dtos

import "time"

// AddDesktopRewardClaim 记录用户是否领取过 Add Desktop 活动奖励。
type AddDesktopRewardClaim struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_add_desktop_user" json:"user_id"`
	ClaimIP   *string   `gorm:"size:64;uniqueIndex:uk_add_desktop_claim_ip" json:"claim_ip,omitempty"`
	ClaimedAt time.Time `gorm:"not null" json:"claimed_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AddDesktopRewardClaim) TableName() string {
	return "add_desktop_reward_claims"
}

// AddDesktopGameEntry records each game entry amount while an add-desktop insurance coupon may be used.
type AddDesktopGameEntry struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID              uint64     `gorm:"not null;index:idx_add_desktop_entry_user_status" json:"user_id"`
	PlatformCode        string     `gorm:"size:32;not null;default:HEDOC;index:idx_add_desktop_entry_user_status" json:"platform_code"`
	ReferenceID         string     `gorm:"size:100;not null;uniqueIndex:uk_add_desktop_entry_reference" json:"reference_id"`
	EntryAmountU        float64    `gorm:"type:decimal(15,2);not null;default:0.00" json:"entry_amount_u"`
	ExitAmountU         float64    `gorm:"type:decimal(15,2);not null;default:0.00" json:"exit_amount_u"`
	LossAmountU         float64    `gorm:"type:decimal(15,2);not null;default:0.00" json:"loss_amount_u"`
	CompensationAmountU float64    `gorm:"type:decimal(15,2);not null;default:0.00" json:"compensation_amount_u"`
	CouponID            *uint64    `gorm:"index:idx_add_desktop_entry_coupon" json:"coupon_id,omitempty"`
	Status              int        `gorm:"not null;default:0;index:idx_add_desktop_entry_user_status;comment:0:pending 1:settled 2:insured" json:"status"`
	EnteredAt           time.Time  `gorm:"not null;index:idx_add_desktop_entry_entered_at" json:"entered_at"`
	SettledAt           *time.Time `json:"settled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (AddDesktopGameEntry) TableName() string {
	return "add_desktop_game_entries"
}

// WorldCupRewardClaim 记录用户是否领取过 World Cup 活动奖励。
type WorldCupRewardClaim struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_world_cup_user" json:"user_id"`
	ClaimedAt time.Time `gorm:"not null" json:"claimed_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WorldCupRewardClaim) TableName() string {
	return "world_cup_reward_claims"
}

// VIPMonthlyBonusClaim 记录用户每个月是否已领取 VIP 月度奖金。
type VIPMonthlyBonusClaim struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"not null;uniqueIndex:uk_vip_monthly_user_month,priority:1" json:"user_id"`
	ClaimMonth   string    `gorm:"type:varchar(7);not null;uniqueIndex:uk_vip_monthly_user_month,priority:2" json:"claim_month"`
	VIPLevel     int       `gorm:"not null" json:"vip_level"`
	RewardAmount float64   `gorm:"not null;default:0" json:"reward_amount"`
	ClaimedAt    time.Time `gorm:"not null" json:"claimed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (VIPMonthlyBonusClaim) TableName() string {
	return "vip_monthly_bonus_claims"
}

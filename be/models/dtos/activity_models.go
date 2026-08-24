package dtos

import (
	"time"

	"gorm.io/datatypes"
)

// Activity 活动基础表定义在 models.gen.go 中
// 本文件仅补充关联表和配置结构体

// ActivityConfig 活动详细配置表
type ActivityConfig struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID  uint           `gorm:"not null;uniqueIndex:idx_activity_key" json:"activity_id"`
	ConfigKey   string         `gorm:"size:100;not null;uniqueIndex:idx_activity_key" json:"config_key"`
	ConfigValue datatypes.JSON `gorm:"type:json" json:"config_value"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (ActivityConfig) TableName() string {
	return "activity_config"
}

// UserActivityProgress 用户活动进度表
type UserActivityProgress struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64         `gorm:"not null;uniqueIndex:idx_user_activity_date" json:"user_id"`
	ActivityType string         `gorm:"size:50;not null;uniqueIndex:idx_user_activity_date" json:"activity_type"`
	DayNumber    int            `gorm:"default:1;uniqueIndex:idx_user_activity_date" json:"day_number"`
	ProgressData datatypes.JSON `gorm:"type:json" json:"progress_data"`
	Status       int            `gorm:"default:0;comment:0:未开始 1:进行中 2:已完成 3:已领取" json:"status"`
	ProgressDate time.Time      `gorm:"type:date;not null;uniqueIndex:idx_user_activity_date" json:"progress_date"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (UserActivityProgress) TableName() string {
	return "user_activity_progress"
}

// UserCoupon 用户优惠券表
type UserCoupon struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64     `gorm:"not null;index:idx_user_id" json:"user_id"`
	CouponCode   string     `gorm:"size:100;not null;uniqueIndex:idx_coupon_code" json:"coupon_code"`
	ActivityType string     `gorm:"size:50;not null" json:"activity_type"`
	CouponValue  float64    `gorm:"type:decimal(15,2);default:0.00" json:"coupon_value"`
	MinDeposit   float64    `gorm:"type:decimal(15,2);default:0.00" json:"min_deposit"`
	Status       int        `gorm:"default:0;comment:0:未激活 1:已激活 2:已使用 3:已过期" json:"status"`
	ValidStart   time.Time  `json:"valid_start"`
	ValidEnd     time.Time  `json:"valid_end"`
	ActivatedAt  *time.Time `json:"activated_at,omitempty"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (UserCoupon) TableName() string {
	return "user_coupons"
}

// ============================================
// 活动配置结构体（用于解析JSON配置）
// ============================================

// ActivityBaseConfig 活动基础配置
type ActivityBaseConfig struct {
	MinDeposit     float64 `json:"min_deposit"`
	Currency       string  `json:"currency"`
	Repeatable     bool    `json:"repeatable,omitempty"`
	CycleDays      int     `json:"cycle_days,omitempty"`
	DailyLimit     int     `json:"daily_limit,omitempty"`
	RebateRate     float64 `json:"rebate_rate,omitempty"`
	MinLoss        float64 `json:"min_loss,omitempty"`
	SettlementType string  `json:"settlement_type,omitempty"`
}

// DayRewardConfig 每日返利配置（储值返利活动）
type DayRewardConfig struct {
	Day         int     `json:"day"`
	Rate        float64 `json:"rate"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
}

// WheelRewardConfig 轮盘奖项配置
type WheelRewardConfig struct {
	Value       float64 `json:"value"`
	Probability float64 `json:"probability"`
	Label       string  `json:"label"`
}

// LossRebateConfig 输返活动规则配置
type LossRebateConfig struct {
	RebateRate       float64 `json:"rebate_rate"`
	MinLossThreshold float64 `json:"min_loss_threshold"`
	MaxRebateAmount  float64 `json:"max_rebate_amount"`
	SettlementCycle  string  `json:"settlement_cycle"`
	SettlementTime   string  `json:"settlement_time"`
	Description      string  `json:"description"`
}

// ============================================
// 进度数据结构（用于解析progress_data JSON）
// ============================================

// RechargeRebateProgressData 储值返利进度数据
type RechargeRebateProgressData struct {
	ActivityID      uint64  `json:"activity_id"`
	DayNumber       int     `json:"day_number"`
	DepositAmount   float64 `json:"deposit_amount"`
	RewardAmount    float64 `json:"reward_amount"`
	RewardRate      float64 `json:"reward_rate"`
	MinDeposit      float64 `json:"min_deposit"`
	Claimed         bool    `json:"claimed"`
	WagerRequired   float64 `json:"wager_required"`
	WagerCompleted  float64 `json:"wager_completed"`
	WagerUnlocked   bool    `json:"wager_unlocked"`
	RewardGrantedAt string  `json:"reward_granted_at"`
	ActivityStart   string  `json:"activity_start"`
}

// WheelProgressData 轮盘活动进度数据
type WheelProgressData struct {
	CouponCode           string  `json:"coupon_code"`
	CouponValue          float64 `json:"coupon_value"`
	MinDepositToActivate float64 `json:"min_deposit_to_activate"`
	SpunAt               string  `json:"spun_at"`
	Activated            bool    `json:"activated"`
	ActivatedAt          *string `json:"activated_at,omitempty"`
}

// LossRebateProgressData 输返活动进度数据
type LossRebateProgressData struct {
	ActivityID       uint64  `json:"activity_id"`
	BetAmount        float64 `json:"bet_amount"`
	WinAmount        float64 `json:"win_amount"`
	NetLoss          float64 `json:"net_loss"`
	RebateRate       float64 `json:"rebate_rate"`
	RebateAmount     float64 `json:"rebate_amount"`
	MinLossThreshold float64 `json:"min_loss_threshold"`
	Claimed          bool    `json:"claimed"`
	CalculatedAt     string  `json:"calculated_at"`
	ClaimedAt        *string `json:"claimed_at,omitempty"`
}

// NewUserRechargeProgressData 新人充值活动进度数据
type NewUserRechargeProgressData struct {
	ActivityID      uint64  `json:"activity_id"`
	DayNumber       int     `json:"day_number"`
	DepositAmount   float64 `json:"deposit_amount"`
	RewardAmount    float64 `json:"reward_amount"`
	RewardRate      float64 `json:"reward_rate"`
	MinDeposit      float64 `json:"min_deposit"`
	WagerMultiplier float64 `json:"wager_multiplier"`
	WagerRequired   float64 `json:"wager_required"`
	WagerCompleted  float64 `json:"wager_completed"`
	WagerUnlocked   bool    `json:"wager_unlocked"`
	RewardGrantedAt string  `json:"reward_granted_at"`
	ActivityStart   string  `json:"activity_start"`
}

type WeeklySpinWheelTicketProgressData struct {
	TicketCode       string  `json:"ticket_code"`
	AwardedAt        string  `json:"awarded_at"`
	BetAmountU       float64 `json:"bet_amount_u"`
	DailyTicketIndex int     `json:"daily_ticket_index"`
	IsBaseTicket     bool    `json:"is_base_ticket"`
	CycleStart       string  `json:"cycle_start"`
	CycleEnd         string  `json:"cycle_end"`
	DrawnAt          *string `json:"drawn_at,omitempty"`
	IsWinner         bool    `json:"is_winner"`
	PrizeTier        string  `json:"prize_tier,omitempty"`
	PrizeRate        float64 `json:"prize_rate,omitempty"`
	PrizeAmountU     float64 `json:"prize_amount_u,omitempty"`
	InvalidatedAt    *string `json:"invalidated_at,omitempty"`
}

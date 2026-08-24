package dtos

// ============================================
// 通用响应结构体
// ============================================

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}

// ============================================
// 储值返利活动 API
// ============================================

// RechargeRebateDayInfo 每日返利信息
type RechargeRebateDayInfo struct {
	Day           int     `json:"day"`
	Rate          float64 `json:"rate"`
	Label         string  `json:"label"`
	Description   string  `json:"description"`
	Status        string  `json:"status"` // upcoming, active, completed, claimed
	DepositAmount float64 `json:"deposit_amount"`
	RewardAmount  float64 `json:"reward_amount"`
	MinDeposit    float64 `json:"min_deposit"`
}

// RechargeRebateStatusResponse 储值返利状态响应（基于充值次数）
// 1st: 50%, 2nd: 75%, 3rd: 100%, 4th: 150%
type RechargeRebateStatusResponse struct {
	ActivityID     uint                    `json:"activity_id"`
	ActivityName   string                  `json:"activity_name"`
	StartTime      string                  `json:"start_time"`
	EndTime        string                  `json:"end_time"`
	CurrentTier    int                     `json:"current_tier"`    // 当前是第几档（下一次充值享受的返利档次）
	CompletedCount int                     `json:"completed_count"` // 已完成充值次数
	MinDeposit     float64                 `json:"min_deposit"`
	Currency       string                  `json:"currency"`
	Tiers          []RechargeRebateDayInfo `json:"tiers"` // 4档返利配置
}

// ClaimRechargeRebateRequest 领取储值返利请求
type ClaimRechargeRebateRequest struct {
	DayNumber int `json:"day_number" binding:"required,min=1,max=4"`
}

// ============================================
// 轮盘活动 API
// ============================================

// WheelItem 轮盘奖项
type WheelItem struct {
	Value       float64 `json:"value"`
	Label       string  `json:"label"`
	Probability float64 `json:"probability"`
}

// WheelCurrentCoupon 当前优惠券状态
type WheelCurrentCoupon struct {
	CouponCode           string  `json:"coupon_code"`
	CouponValue          float64 `json:"coupon_value"`
	MinDepositToActivate float64 `json:"min_deposit_to_activate"`
	Activated            bool    `json:"activated"`
	ActivatedAt          *string `json:"activated_at,omitempty"`
	ExpiresAt            string  `json:"expires_at"`
}

// WheelStatusResponse 轮盘状态响应
type WheelStatusResponse struct {
	ActivityID    uint                `json:"activity_id"`
	ActivityName  string              `json:"activity_name"`
	MinDeposit    float64             `json:"min_deposit"`
	CanSpin       bool                `json:"can_spin"`
	DailyLimit    int                 `json:"daily_limit"`
	Items         []WheelItem         `json:"items"`
	CurrentCoupon *WheelCurrentCoupon `json:"current_coupon,omitempty"`
}

// SpinResponse 轮盘抽奖响应
type SpinResponse struct {
	Success    bool    `json:"success"`
	Value      float64 `json:"value"`
	Label      string  `json:"label"`
	CouponCode string  `json:"coupon_code"`
	MinDeposit float64 `json:"min_deposit"`
	ValidUntil string  `json:"valid_until"`
}

// ============================================
// 输返活动 API
// ============================================

// LossRebateStats 输返统计
type LossRebateStats struct {
	BetAmount        float64 `json:"bet_amount"`
	WinAmount        float64 `json:"win_amount"`
	NetLoss          float64 `json:"net_loss"`
	RebateRate       float64 `json:"rebate_rate"`
	RebateAmount     float64 `json:"rebate_amount"`
	MinLossThreshold float64 `json:"min_loss_threshold"`
	Claimed          bool    `json:"claimed"`
}

// LossRebateStatusResponse 输返状态响应
type LossRebateStatusResponse struct {
	ActivityID       uint            `json:"activity_id"`
	ActivityName     string          `json:"activity_name"`
	StartTime        string          `json:"start_time"`
	EndTime          string          `json:"end_time"`
	RebateRate       float64         `json:"rebate_rate"`
	MinLossThreshold float64         `json:"min_loss_threshold"`
	Currency         string          `json:"currency"`
	Stats            LossRebateStats `json:"stats"`
	CanClaim         bool            `json:"can_claim"`
}

// ============================================
// 通用领取响应
// ============================================

// ClaimResponse 领取奖励响应
type ClaimResponse struct {
	Success      bool    `json:"success"`
	RewardType   string  `json:"reward_type"` // balance, coupon, etc.
	RewardAmount float64 `json:"reward_amount"`
	Message      string  `json:"message"`
}

// ============================================
// 优惠券 API
// ============================================

// ActivateCouponRequest 激活优惠券请求
type ActivateCouponRequest struct {
	CouponCode    string  `json:"coupon_code" binding:"required"`
	DepositAmount float64 `json:"deposit_amount" binding:"required,gt=0"`
}

// CouponListResponse 优惠券列表响应
type CouponListResponse struct {
	Total   int          `json:"total"`
	Coupons []UserCoupon `json:"coupons"`
}

// ============================================
// 活动状态（包含未开启/倒计时）
// ============================================

// ActivityStatus 活动状态（通用）
type ActivityStatus struct {
	ActivityID   uint   `json:"activity_id"`
	ActivityName string `json:"activity_name"`
	Type         string `json:"type"`
	Status       int    `json:"status"` // 0:未开始 1:进行中 2:已结束
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Description  string `json:"description"`
}

// ActivityWithCountdown 带倒计时的活动信息
type ActivityWithCountdown struct {
	ActivityStatus
	Countdown     int64       `json:"countdown"`        // 倒计时秒数（-1表示未开始，0表示进行中，正数表示距离开始还有多久）
	CountdownText string      `json:"countdown_text"`   // 倒计时显示文本
	Config        interface{} `json:"config,omitempty"` // 活动配置
	IsActive      bool        `json:"is_active"`        // 是否进行中
}

// ActivityListResponse 活动列表响应
type ActivityListResponse struct {
	Activities []ActivityWithCountdown `json:"activities"`
}

// ============================================
// 单次领取活动 API
// ============================================

type ActivityClaimStatusResponse struct {
	ActivityType string  `json:"activity_type"`
	Claimed      bool    `json:"claimed"`
	ClaimedAt    *string `json:"claimed_at,omitempty"`
}

type VIPMonthlyBonusStatusResponse struct {
	ActivityType    string  `json:"activity_type"`
	CurrentVIPLevel int     `json:"current_vip_level"`
	ClaimMonth      string  `json:"claim_month"`
	RewardAmount    float64 `json:"reward_amount"`
	CanClaim        bool    `json:"can_claim"`
	Claimed         bool    `json:"claimed"`
	ClaimedAt       *string `json:"claimed_at,omitempty"`
	Balance         float64 `json:"balance"`
}

type ActivityClaimRecordResponse struct {
	Success        bool    `json:"success"`
	ActivityType   string  `json:"activity_type"`
	Claimed        bool    `json:"claimed"`
	AlreadyClaimed bool    `json:"already_claimed"`
	ClaimedAt      string  `json:"claimed_at"`
	RewardAmount   float64 `json:"reward_amount"`
	Balance        float64 `json:"balance"`
	RewardType     string  `json:"reward_type,omitempty"`
	CouponID       uint64  `json:"coupon_id,omitempty"`
	Message        string  `json:"message"`
}

type AddDesktopInsuranceSettlementResponse struct {
	Triggered           bool    `json:"triggered"`
	EntryAmountU        float64 `json:"entry_amount_u"`
	ExitAmountU         float64 `json:"exit_amount_u"`
	LossAmountU         float64 `json:"loss_amount_u"`
	CompensationAmountU float64 `json:"compensation_amount_u"`
	CurrentBalance      float64 `json:"current_balance"`
	CouponCode          string  `json:"coupon_code,omitempty"`
	Message             string  `json:"message,omitempty"`
}

// BettingRankValueResponse 投注排行页面展示数值
type BettingRankValueResponse struct {
	CurrentValue       int64  `json:"current_value"`
	WeekStart          string `json:"week_start"`
	NextResetAt        string `json:"next_reset_at"`
	RefreshedAt        string `json:"refreshed_at"`
	CompletedIntervals int    `json:"completed_intervals"`
}

type WeeklySpinWheelTicket struct {
	TicketCode       string  `json:"ticket_code"`
	AwardedAt        string  `json:"awarded_at"`
	BetAmountU       float64 `json:"bet_amount_u"`
	IsDailyQualified bool    `json:"is_daily_qualified"`
	IsToday          bool    `json:"is_today"`
	IsWinner         bool    `json:"is_winner"`
	PrizeTier        string  `json:"prize_tier,omitempty"`
	PrizeRate        float64 `json:"prize_rate,omitempty"`
	PrizeAmountU     float64 `json:"prize_amount_u,omitempty"`
}

type WeeklySpinWheelDrawWinner struct {
	TicketCode   string  `json:"ticket_code"`
	PrizeTier    string  `json:"prize_tier"`
	PrizeRate    float64 `json:"prize_rate"`
	PrizeAmountU float64 `json:"prize_amount_u"`
}

type WeeklySpinWheelLatestDraw struct {
	CycleStart   string                      `json:"cycle_start"`
	CycleEnd     string                      `json:"cycle_end"`
	DrawnAt      string                      `json:"drawn_at"`
	TotalTickets int                         `json:"total_tickets"`
	Winners      []WeeklySpinWheelDrawWinner `json:"winners"`
}

type WeeklySpinWheelStatusResponse struct {
	ActivityType          string                     `json:"activity_type"`
	ServerTimeUTC         string                     `json:"server_time_utc"`
	CurrentCycleStart     string                     `json:"current_cycle_start"`
	CurrentCycleEnd       string                     `json:"current_cycle_end"`
	NextDailyResetAt      string                     `json:"next_daily_reset_at"`
	NextWeeklyDrawAt      string                     `json:"next_weekly_draw_at"`
	DisplayCurrency       string                     `json:"display_currency"`
	TodayBetAmountU       float64                    `json:"today_bet_amount_u"`
	TodayBetAmountIDR     float64                    `json:"today_bet_amount_idr"`
	TodayBetAmountPHP     float64                    `json:"today_bet_amount_php"`
	TodayBetAmount        float64                    `json:"today_bet_amount"`
	BaseTicketThresholdU  float64                    `json:"base_ticket_threshold_u"`
	ExtraTicketThresholdU float64                    `json:"extra_ticket_threshold_u"`
	BaseTicketThreshold   float64                    `json:"base_ticket_threshold"`
	ExtraTicketThreshold  float64                    `json:"extra_ticket_threshold"`
	BaseTicketQualified   bool                       `json:"base_ticket_qualified"`
	ExtraTicketsToday     int                        `json:"extra_tickets_today"`
	CurrentCycleTickets   int                        `json:"current_cycle_tickets"`
	TotalPrizePoolU       float64                    `json:"total_prize_pool_u"`
	TotalPrizePool        float64                    `json:"total_prize_pool"`
	TotalPrizePoolIDR     float64                    `json:"total_prize_pool_idr"`
	TotalPrizePoolPHP     float64                    `json:"total_prize_pool_php"`
	HundredUIDR           float64                    `json:"hundred_u_idr"`
	HundredUPHP           float64                    `json:"hundred_u_php"`
	ThousandUIDR          float64                    `json:"thousand_u_idr"`
	ThousandUPHP          float64                    `json:"thousand_u_php"`
	Tickets               []WeeklySpinWheelTicket    `json:"tickets"`
	LatestDraw            *WeeklySpinWheelLatestDraw `json:"latest_draw,omitempty"`
}

type SevenDayTopupPendingMakeup struct {
	MissedDate   string `json:"missed_date"`
	RechargeDate string `json:"recharge_date"`
}

type SevenDayTopupStatusResponse struct {
	ActivityType      string                      `json:"activity_type"`
	Currency          string                      `json:"currency"`
	DailyThreshold    float64                     `json:"daily_threshold"`
	ProgressDates     []string                    `json:"progress_dates"`
	ProgressCount     int                         `json:"progress_count"`
	UsedMakeup        bool                        `json:"used_makeup"`
	CompletedRound    bool                        `json:"completed_round"`
	PendingMakeup     *SevenDayTopupPendingMakeup `json:"pending_makeup,omitempty"`
	DirectInviteCount int64                       `json:"direct_invite_count"`
}

type NewUserRechargeTierInfo struct {
	Day             int     `json:"day"`
	Rate            float64 `json:"rate"`
	Status          string  `json:"status"`
	DepositAmount   float64 `json:"deposit_amount"`
	RewardAmount    float64 `json:"reward_amount"`
	MinDeposit      float64 `json:"min_deposit"`
	WagerRequired   float64 `json:"wager_required"`
	WagerCompleted  float64 `json:"wager_completed"`
	DepositAmountU  float64 `json:"deposit_amount_u"`
	RewardAmountU   float64 `json:"reward_amount_u"`
	MinDepositU     float64 `json:"min_deposit_u"`
	WagerRequiredU  float64 `json:"wager_required_u"`
	WagerCompletedU float64 `json:"wager_completed_u"`
	WagerUnlocked   bool    `json:"wager_unlocked"`
	RewardGrantedAt string  `json:"reward_granted_at,omitempty"`
}

type NewUserRechargeStatusResponse struct {
	ActivityType        string                    `json:"activity_type"`
	MinDeposit          float64                   `json:"min_deposit"`
	WagerMultiplier     float64                   `json:"wager_multiplier"`
	Currency            string                    `json:"currency"`
	ProgressCount       int                       `json:"progress_count"`
	NextTier            int                       `json:"next_tier"`
	Hidden              bool                      `json:"hidden"`
	AllWagerUnlocked    bool                      `json:"all_wager_unlocked"`
	RemainingWager      float64                   `json:"remaining_wager"`
	RemainingWagerU     float64                   `json:"remaining_wager_u"`
	LatestRewardAmount  float64                   `json:"latest_reward_amount"`
	LatestRewardAmountU float64                   `json:"latest_reward_amount_u"`
	Tiers               []NewUserRechargeTierInfo `json:"tiers"`
}

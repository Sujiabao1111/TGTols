package dtos

type DailyWeeklyChallengeTask struct {
	CycleType       string  `json:"cycle_type"`
	TaskIndex       int     `json:"task_index"`
	RequiredBetU    float64 `json:"required_bet_u"`
	RewardAmountU   float64 `json:"reward_amount_u"`
	RequiredBetIDR  float64 `json:"required_bet_idr"`
	RewardAmountIDR float64 `json:"reward_amount_idr"`
	RequiredBetPHP  float64 `json:"required_bet_php"`
	RewardAmountPHP float64 `json:"reward_amount_php"`
	CurrentBetU     float64 `json:"current_bet_u"`
	CurrentBetIDR   float64 `json:"current_bet_idr"`
	CurrentBetPHP   float64 `json:"current_bet_php"`
	DisplayCurrency string  `json:"display_currency"`
	RequiredBet     float64 `json:"required_bet"`
	RewardAmount    float64 `json:"reward_amount"`
	CurrentBet      float64 `json:"current_bet"`
	ProgressRatio   float64 `json:"progress_ratio"`
	Claimed         bool    `json:"claimed"`
	Claimable       bool    `json:"claimable"`
	ClaimedAt       string  `json:"claimed_at,omitempty"`
	PeriodStartUTC  string  `json:"period_start_utc"`
	PeriodEndUTC    string  `json:"period_end_utc"`
}

type DailyWeeklyChallengeStatusResponse struct {
	ActivityType      string                     `json:"activity_type"`
	ServerTimeUTC     string                     `json:"server_time_utc"`
	DailyPeriodStart  string                     `json:"daily_period_start"`
	DailyPeriodEnd    string                     `json:"daily_period_end"`
	WeeklyPeriodStart string                     `json:"weekly_period_start"`
	WeeklyPeriodEnd   string                     `json:"weekly_period_end"`
	NextDailyResetAt  string                     `json:"next_daily_reset_at"`
	NextWeeklyResetAt string                     `json:"next_weekly_reset_at"`
	DailyTasks        []DailyWeeklyChallengeTask `json:"daily_tasks"`
	WeeklyTasks       []DailyWeeklyChallengeTask `json:"weekly_tasks"`
}

type DailyWeeklyChallengeClaimRequest struct {
	CycleType string `json:"cycle_type"`
	TaskIndex int    `json:"task_index"`
}

type DailyWeeklyChallengeClaimResponse struct {
	Success        bool    `json:"success"`
	ActivityType   string  `json:"activity_type"`
	CycleType      string  `json:"cycle_type"`
	TaskIndex      int     `json:"task_index"`
	Claimed        bool    `json:"claimed"`
	AlreadyClaimed bool    `json:"already_claimed"`
	ClaimedAt      string  `json:"claimed_at"`
	RewardAmountU  float64 `json:"reward_amount_u"`
	Balance        float64 `json:"balance"`
	Message        string  `json:"message"`
}

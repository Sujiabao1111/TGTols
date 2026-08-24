package request

type GrantDesktopRewardRequest struct {
	UserID uint64 `json:"userId" binding:"required,gt=0"`
}

type GrantRewardRequest struct {
	UserID        uint64  `json:"userId" binding:"required,gt=0"`
	RewardAmountU float64 `json:"rewardAmountU" binding:"required,gt=0"`
}

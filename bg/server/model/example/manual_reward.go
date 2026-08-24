package example

// ManualRewardResult describes the reward grant outcome returned to the admin UI.
type ManualRewardResult struct {
	UserID          uint64  `json:"userId"`
	Username        string  `json:"username"`
	RewardAmountU   float64 `json:"rewardAmountU"`
	RewardType      string  `json:"rewardType"`
	CouponID        uint64  `json:"couponId"`
	CouponCode      string  `json:"couponCode"`
	Currency        string  `json:"currency"`
	LocalAmount     float64 `json:"localAmount"`
	ExchangeRate    float64 `json:"exchangeRate"`
	WagerMultiplier float64 `json:"wagerMultiplier"`
	WagerRequired   float64 `json:"wagerRequired"`
	BeforeBalance   float64 `json:"beforeBalance"`
	AfterBalance    float64 `json:"afterBalance"`
	ReferenceID     string  `json:"referenceId"`
	RewardCategory  string  `json:"rewardCategory"`
}

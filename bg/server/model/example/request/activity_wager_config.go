package request

type UpdateActivityWagerConfigRequest struct {
	DepositWagerMultiplier float64 `json:"depositWagerMultiplier" binding:"required,gt=0"`
	RewardWagerMultiplier  float64 `json:"rewardWagerMultiplier" binding:"required,gt=0"`
}

package request

type UpdateWithdrawFeeConfigRequest struct {
	Enabled bool    `json:"enabled"`
	FeeRate float64 `json:"feeRate" binding:"gte=0,lte=1"`
}

package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type WithdrawAuditSearch struct {
	UserID  *uint64 `json:"userId" form:"userId"`
	OrderID string  `json:"orderId" form:"orderId"`
	Status  *int    `json:"status" form:"status"`
	request.PageInfo
}

type ReviewWithdrawOrderRequest struct {
	OrderID string `json:"orderId"`
	Remark  string `json:"remark"`
}

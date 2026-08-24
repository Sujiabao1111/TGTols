package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type RechargeQuerySearch struct {
	UserID    *uint64  `json:"userId" form:"userId"`
	Date      string   `json:"date" form:"date"`
	DateRange []string `json:"dateRange" form:"dateRange[]"`
	request.PageInfo
}

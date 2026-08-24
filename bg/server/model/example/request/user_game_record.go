package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type UserGameRecordSearch struct {
	UserID       *uint64  `json:"userId" form:"userId"`
	Date         string   `json:"date" form:"date"`
	DateRange    []string `json:"dateRange" form:"dateRange[]"`
	PlatformCode string   `json:"platformCode" form:"platformCode"`
	request.PageInfo
}

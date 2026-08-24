package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type GameLaunchErrorSearch struct {
	UserID    *uint64 `json:"userId" form:"userId"`
	ErrorType string  `json:"errorType" form:"errorType"`
	request.PageInfo
}

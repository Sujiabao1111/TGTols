package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type GameProvidersSearch struct {
	PlatformCode *string `json:"platformCode" form:"platformCode"`
	request.PageInfo
}

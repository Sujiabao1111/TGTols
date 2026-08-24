package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type GamesSearch struct {
	PlatformCode *string `json:"platformCode" form:"platformCode"`
	ProviderId   *int    `json:"providerId" form:"providerId"`
	GameTypeId   *int    `json:"gameTypeId" form:"gameTypeId"`
	request.PageInfo
}

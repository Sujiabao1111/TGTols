package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type InviteStatsSearch struct {
	InviterUserId  *uint64  `json:"inviterUserId" form:"inviterUserId"`
	InviteCode     string   `json:"inviteCode" form:"inviteCode"`
	RegisterDomain string   `json:"registerDomain" form:"registerDomain"`
	StatDateRange  []string `json:"statDateRange" form:"statDateRange[]"`
	request.PageInfo
}

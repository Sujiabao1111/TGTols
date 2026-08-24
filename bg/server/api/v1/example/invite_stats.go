package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type InviteStatsApi struct{}

type inviteStatsPageResult struct {
	response.PageResult
	Summary example.InviteStatsSummary `json:"summary"`
}

// GetInviteStatsList returns invite statistics by inviter.
func (inviteStatsApi *InviteStatsApi) GetInviteStatsList(c *gin.Context) {
	ctx := c.Request.Context()

	var pageInfo exampleReq.InviteStatsSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, total, summary, err := inviteStatsService.GetInviteStatsInfoList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("get invite stats failed", zap.Error(err))
		response.FailWithMessage("\u83b7\u53d6\u9080\u8bf7\u7edf\u8ba1\u5931\u8d25: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(inviteStatsPageResult{
		PageResult: response.PageResult{
			List:     list,
			Total:    total,
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
		},
		Summary: summary,
	}, "\u83b7\u53d6\u6210\u529f", c)
}

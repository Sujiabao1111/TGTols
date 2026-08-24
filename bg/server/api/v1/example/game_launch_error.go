package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GameLaunchErrorApi struct{}

func (gameLaunchErrorApi *GameLaunchErrorApi) GetGameLaunchErrorList(c *gin.Context) {
	ctx := c.Request.Context()

	var pageInfo exampleReq.GameLaunchErrorSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	list, total, err := gameLaunchErrorService.GetGameLaunchErrorList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("get game launch error list failed", zap.Error(err))
		response.FailWithMessage("获取玩家报错信息失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

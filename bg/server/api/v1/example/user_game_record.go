package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserGameRecordApi struct{}

func (userGameRecordApi *UserGameRecordApi) GetUserGameRecordList(c *gin.Context) {
	ctx := c.Request.Context()

	var pageInfo exampleReq.UserGameRecordSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, total, err := userGameRecordService.GetUserGameRecordList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("get user game records failed", zap.Error(err))
		response.FailWithMessage("获取用户游戏记录失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

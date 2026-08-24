package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserDataApi struct{}

// GetUserDataList returns a single user data snapshot by user id.
func (userDataApi *UserDataApi) GetUserDataList(c *gin.Context) {
	ctx := c.Request.Context()

	var pageInfo exampleReq.UserDataSearch
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

	list, total, err := userDataService.GetUserDataList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("get user data failed", zap.Error(err))
		response.FailWithMessage("获取用户数据失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

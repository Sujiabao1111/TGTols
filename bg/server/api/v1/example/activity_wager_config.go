package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ActivityWagerConfigApi struct{}

func (api *ActivityWagerConfigApi) GetActivityWagerConfig(c *gin.Context) {
	config, err := activityWagerConfigService.GetActivityWagerConfig(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("get activity wager config failed", zap.Error(err))
		response.FailWithMessage("获取打码配置失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(config, "获取成功", c)
}

func (api *ActivityWagerConfigApi) UpdateActivityWagerConfig(c *gin.Context) {
	var req exampleReq.UpdateActivityWagerConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	config, err := activityWagerConfigService.UpdateActivityWagerConfig(
		c.Request.Context(),
		req.DepositWagerMultiplier,
		req.RewardWagerMultiplier,
	)
	if err != nil {
		global.GVA_LOG.Error("update activity wager config failed", zap.Error(err))
		response.FailWithMessage("保存打码配置失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(config, "保存成功", c)
}

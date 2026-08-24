package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WithdrawFeeConfigApi struct{}

func (api *WithdrawFeeConfigApi) GetWithdrawFeeConfig(c *gin.Context) {
	config, err := withdrawFeeConfigService.GetWithdrawFeeConfig(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("get withdraw fee config failed", zap.Error(err))
		response.FailWithMessage("获取提现手续费配置失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(config, "获取成功", c)
}

func (api *WithdrawFeeConfigApi) UpdateWithdrawFeeConfig(c *gin.Context) {
	var req exampleReq.UpdateWithdrawFeeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	config, err := withdrawFeeConfigService.UpdateWithdrawFeeConfig(c.Request.Context(), req.Enabled, req.FeeRate)
	if err != nil {
		global.GVA_LOG.Error("update withdraw fee config failed", zap.Error(err))
		response.FailWithMessage("保存提现手续费配置失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(config, "保存成功", c)
}

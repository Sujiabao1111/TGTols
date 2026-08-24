package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ManualRewardApi struct{}

// GrantDesktopReward grants a desktop insurance coupon.
func (manualRewardApi *ManualRewardApi) GrantDesktopReward(c *gin.Context) {
	ctx := c.Request.Context()

	var req exampleReq.GrantDesktopRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := manualRewardService.GrantDesktopReward(ctx, req.UserID)
	if err != nil {
		global.GVA_LOG.Error("grant desktop reward failed", zap.Error(err))
		response.FailWithMessage("\u53d1\u653e\u684c\u9762\u8865\u507f\u5238\u5931\u8d25: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(result, "\u53d1\u653e\u6210\u529f", c)
}

// GrantReward grants a custom U-denominated reward.
func (manualRewardApi *ManualRewardApi) GrantReward(c *gin.Context) {
	ctx := c.Request.Context()

	var req exampleReq.GrantRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := manualRewardService.GrantReward(ctx, req.UserID, req.RewardAmountU)
	if err != nil {
		global.GVA_LOG.Error("grant reward failed", zap.Error(err))
		response.FailWithMessage("\u53d1\u653e\u5956\u52b1\u5931\u8d25: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(result, "\u53d1\u653e\u6210\u529f", c)
}

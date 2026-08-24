package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ManualRewardRouter struct{}

func (r *ManualRewardRouter) InitManualRewardRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	manualRewardRouter := Router.Group("manualReward").Use(middleware.OperationRecord())
	_ = PublicRouter

	manualRewardRouter.POST("grantDesktopReward", manualRewardApi.GrantDesktopReward)
	manualRewardRouter.POST("grantReward", manualRewardApi.GrantReward)
}

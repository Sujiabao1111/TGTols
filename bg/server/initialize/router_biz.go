package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/gin-gonic/gin"
)

func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}
func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]
	holder(publicGroup, privateGroup)
	{
		exampleRouter := router.RouterGroupApp.Example
		exampleRouter.InitGamesRouter(privateGroup, publicGroup)
		exampleRouter.InitAgentRebateConfigsRouter(privateGroup, publicGroup)
		exampleRouter.InitBannersRouter(privateGroup, publicGroup)
		exampleRouter.InitGameTypesRouter(privateGroup, publicGroup)
		exampleRouter.InitGameProvidersRouter(privateGroup, publicGroup)
		exampleRouter.InitDailyUserStatsRouter(privateGroup, publicGroup)
		exampleRouter.InitStatsRetentionRouter(privateGroup, publicGroup)
		exampleRouter.InitInviteStatsRouter(privateGroup, publicGroup)
		exampleRouter.InitManualRewardRouter(privateGroup, publicGroup)
		exampleRouter.InitActivityWagerConfigRouter(privateGroup, publicGroup)
		exampleRouter.InitWithdrawFeeConfigRouter(privateGroup, publicGroup)
		exampleRouter.InitGameLaunchErrorRouter(privateGroup, publicGroup)
		exampleRouter.InitUserDataRouter(privateGroup, publicGroup)
		exampleRouter.InitRechargeQueryRouter(privateGroup, publicGroup)
		exampleRouter.InitUserGameRecordRouter(privateGroup, publicGroup)
		exampleRouter.InitWithdrawAuditRouter(privateGroup, publicGroup)
		exampleRouter.InitDataStatisticsRouter(privateGroup, publicGroup)
	}
}

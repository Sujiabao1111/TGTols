package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type StatsRetentionRouter struct {}

// InitStatsRetentionRouter 初始化 statsRetention表 路由信息
func (s *StatsRetentionRouter) InitStatsRetentionRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	statsRetentionRouter := Router.Group("statsRetention").Use(middleware.OperationRecord())
	statsRetentionRouterWithoutRecord := Router.Group("statsRetention")
	statsRetentionRouterWithoutAuth := PublicRouter.Group("statsRetention")
	{
		statsRetentionRouter.POST("createStatsRetention", statsRetentionApi.CreateStatsRetention)   // 新建statsRetention表
		statsRetentionRouter.DELETE("deleteStatsRetention", statsRetentionApi.DeleteStatsRetention) // 删除statsRetention表
		statsRetentionRouter.DELETE("deleteStatsRetentionByIds", statsRetentionApi.DeleteStatsRetentionByIds) // 批量删除statsRetention表
		statsRetentionRouter.PUT("updateStatsRetention", statsRetentionApi.UpdateStatsRetention)    // 更新statsRetention表
	}
	{
		statsRetentionRouterWithoutRecord.GET("findStatsRetention", statsRetentionApi.FindStatsRetention)        // 根据ID获取statsRetention表
		statsRetentionRouterWithoutRecord.GET("getStatsRetentionList", statsRetentionApi.GetStatsRetentionList)  // 获取statsRetention表列表
	}
	{
	    statsRetentionRouterWithoutAuth.GET("getStatsRetentionPublic", statsRetentionApi.GetStatsRetentionPublic)  // statsRetention表开放接口
	}
}

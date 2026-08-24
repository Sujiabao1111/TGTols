package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DailyUserStatsRouter struct {}

// InitDailyUserStatsRouter 初始化 dailyUserStats表 路由信息
func (s *DailyUserStatsRouter) InitDailyUserStatsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	dailyUserStatsRouter := Router.Group("dailyUserStats").Use(middleware.OperationRecord())
	dailyUserStatsRouterWithoutRecord := Router.Group("dailyUserStats")
	dailyUserStatsRouterWithoutAuth := PublicRouter.Group("dailyUserStats")
	{
		dailyUserStatsRouter.POST("createDailyUserStats", dailyUserStatsApi.CreateDailyUserStats)   // 新建dailyUserStats表
		dailyUserStatsRouter.DELETE("deleteDailyUserStats", dailyUserStatsApi.DeleteDailyUserStats) // 删除dailyUserStats表
		dailyUserStatsRouter.DELETE("deleteDailyUserStatsByIds", dailyUserStatsApi.DeleteDailyUserStatsByIds) // 批量删除dailyUserStats表
		dailyUserStatsRouter.PUT("updateDailyUserStats", dailyUserStatsApi.UpdateDailyUserStats)    // 更新dailyUserStats表
	}
	{
		dailyUserStatsRouterWithoutRecord.GET("findDailyUserStats", dailyUserStatsApi.FindDailyUserStats)        // 根据ID获取dailyUserStats表
		dailyUserStatsRouterWithoutRecord.GET("getDailyUserStatsList", dailyUserStatsApi.GetDailyUserStatsList)  // 获取dailyUserStats表列表
	}
	{
	    dailyUserStatsRouterWithoutAuth.GET("getDailyUserStatsPublic", dailyUserStatsApi.GetDailyUserStatsPublic)  // dailyUserStats表开放接口
	}
}

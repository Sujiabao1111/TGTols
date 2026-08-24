package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type GameTypesRouter struct {}

// InitGameTypesRouter 初始化 gameTypes表 路由信息
func (s *GameTypesRouter) InitGameTypesRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	gameTypesRouter := Router.Group("gameTypes").Use(middleware.OperationRecord())
	gameTypesRouterWithoutRecord := Router.Group("gameTypes")
	gameTypesRouterWithoutAuth := PublicRouter.Group("gameTypes")
	{
		gameTypesRouter.POST("createGameTypes", gameTypesApi.CreateGameTypes)   // 新建gameTypes表
		gameTypesRouter.DELETE("deleteGameTypes", gameTypesApi.DeleteGameTypes) // 删除gameTypes表
		gameTypesRouter.DELETE("deleteGameTypesByIds", gameTypesApi.DeleteGameTypesByIds) // 批量删除gameTypes表
		gameTypesRouter.PUT("updateGameTypes", gameTypesApi.UpdateGameTypes)    // 更新gameTypes表
	}
	{
		gameTypesRouterWithoutRecord.GET("findGameTypes", gameTypesApi.FindGameTypes)        // 根据ID获取gameTypes表
		gameTypesRouterWithoutRecord.GET("getGameTypesList", gameTypesApi.GetGameTypesList)  // 获取gameTypes表列表
	}
	{
	    gameTypesRouterWithoutAuth.GET("getGameTypesPublic", gameTypesApi.GetGameTypesPublic)  // gameTypes表开放接口
	}
}

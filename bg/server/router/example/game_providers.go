package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type GameProvidersRouter struct {}

// InitGameProvidersRouter 初始化 gameProviders表 路由信息
func (s *GameProvidersRouter) InitGameProvidersRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	gameProvidersRouter := Router.Group("gameProviders").Use(middleware.OperationRecord())
	gameProvidersRouterWithoutRecord := Router.Group("gameProviders")
	gameProvidersRouterWithoutAuth := PublicRouter.Group("gameProviders")
	{
		gameProvidersRouter.POST("createGameProviders", gameProvidersApi.CreateGameProviders)   // 新建gameProviders表
		gameProvidersRouter.DELETE("deleteGameProviders", gameProvidersApi.DeleteGameProviders) // 删除gameProviders表
		gameProvidersRouter.DELETE("deleteGameProvidersByIds", gameProvidersApi.DeleteGameProvidersByIds) // 批量删除gameProviders表
		gameProvidersRouter.PUT("updateGameProviders", gameProvidersApi.UpdateGameProviders)    // 更新gameProviders表
	}
	{
		gameProvidersRouterWithoutRecord.GET("findGameProviders", gameProvidersApi.FindGameProviders)        // 根据ID获取gameProviders表
		gameProvidersRouterWithoutRecord.GET("getGameProvidersList", gameProvidersApi.GetGameProvidersList)  // 获取gameProviders表列表
	}
	{
	    gameProvidersRouterWithoutAuth.GET("getGameProvidersPublic", gameProvidersApi.GetGameProvidersPublic)  // gameProviders表开放接口
	}
}

package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type GamesRouter struct {}

// InitGamesRouter 初始化 games表 路由信息
func (s *GamesRouter) InitGamesRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	gamesRouter := Router.Group("games").Use(middleware.OperationRecord())
	gamesRouterWithoutRecord := Router.Group("games")
	gamesRouterWithoutAuth := PublicRouter.Group("games")
	{
		gamesRouter.POST("createGames", gamesApi.CreateGames)   // 新建games表
		gamesRouter.DELETE("deleteGames", gamesApi.DeleteGames) // 删除games表
		gamesRouter.DELETE("deleteGamesByIds", gamesApi.DeleteGamesByIds) // 批量删除games表
		gamesRouter.PUT("updateGames", gamesApi.UpdateGames)    // 更新games表
	}
	{
		gamesRouterWithoutRecord.GET("findGames", gamesApi.FindGames)        // 根据ID获取games表
		gamesRouterWithoutRecord.GET("getGamesList", gamesApi.GetGamesList)  // 获取games表列表
	}
	{
	    gamesRouterWithoutAuth.GET("getGamesPublic", gamesApi.GetGamesPublic)  // games表开放接口
	}
}

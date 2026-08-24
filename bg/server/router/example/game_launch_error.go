package example

import "github.com/gin-gonic/gin"

type GameLaunchErrorRouter struct{}

func (r *GameLaunchErrorRouter) InitGameLaunchErrorRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	gameLaunchErrorRouterWithoutRecord := Router.Group("gameLaunchError")
	_ = PublicRouter

	gameLaunchErrorRouterWithoutRecord.GET("getGameLaunchErrorList", gameLaunchErrorApi.GetGameLaunchErrorList)
}

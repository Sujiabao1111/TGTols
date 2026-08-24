package example

import "github.com/gin-gonic/gin"

type UserGameRecordRouter struct{}

func (r *UserGameRecordRouter) InitUserGameRecordRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	userGameRecordRouterWithoutRecord := Router.Group("userGameRecord")
	_ = PublicRouter

	userGameRecordRouterWithoutRecord.GET("getUserGameRecordList", userGameRecordApi.GetUserGameRecordList)
}

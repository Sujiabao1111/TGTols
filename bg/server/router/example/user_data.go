package example

import "github.com/gin-gonic/gin"

type UserDataRouter struct{}

// InitUserDataRouter initializes user data routes.
func (r *UserDataRouter) InitUserDataRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	userDataRouterWithoutRecord := Router.Group("userData")
	_ = PublicRouter

	userDataRouterWithoutRecord.GET("getUserDataList", userDataApi.GetUserDataList)
}

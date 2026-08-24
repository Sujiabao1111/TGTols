package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ActivityWagerConfigRouter struct{}

func (r *ActivityWagerConfigRouter) InitActivityWagerConfigRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	activityWagerConfigRouter := Router.Group("activityWagerConfig").Use(middleware.OperationRecord())
	activityWagerConfigRouterWithoutRecord := Router.Group("activityWagerConfig")
	_ = PublicRouter

	activityWagerConfigRouterWithoutRecord.GET("getActivityWagerConfig", activityWagerConfigApi.GetActivityWagerConfig)
	activityWagerConfigRouter.PUT("updateActivityWagerConfig", activityWagerConfigApi.UpdateActivityWagerConfig)
}

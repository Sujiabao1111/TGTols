package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type WithdrawFeeConfigRouter struct{}

func (r *WithdrawFeeConfigRouter) InitWithdrawFeeConfigRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	withdrawFeeConfigRouter := Router.Group("withdrawFeeConfig").Use(middleware.OperationRecord())
	withdrawFeeConfigRouterWithoutRecord := Router.Group("withdrawFeeConfig")
	_ = PublicRouter

	withdrawFeeConfigRouterWithoutRecord.GET("getWithdrawFeeConfig", withdrawFeeConfigApi.GetWithdrawFeeConfig)
	withdrawFeeConfigRouter.PUT("updateWithdrawFeeConfig", withdrawFeeConfigApi.UpdateWithdrawFeeConfig)
}

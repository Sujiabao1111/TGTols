package example

import "github.com/gin-gonic/gin"

type RechargeQueryRouter struct{}

func (r *RechargeQueryRouter) InitRechargeQueryRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	rechargeQueryRouterWithoutRecord := Router.Group("rechargeQuery")
	_ = PublicRouter

	rechargeQueryRouterWithoutRecord.GET("getRechargeQueryList", rechargeQueryApi.GetRechargeQueryList)
}

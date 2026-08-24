package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type WithdrawAuditRouter struct{}

func (r *WithdrawAuditRouter) InitWithdrawAuditRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	withdrawAuditRouter := Router.Group("withdrawAudit").Use(middleware.OperationRecord())
	withdrawAuditRouterWithoutRecord := Router.Group("withdrawAudit")
	_ = PublicRouter

	withdrawAuditRouterWithoutRecord.GET("getWithdrawAuditList", withdrawAuditApi.GetWithdrawAuditList)
	withdrawAuditRouter.POST("approve", withdrawAuditApi.ApproveWithdrawOrder)
	withdrawAuditRouter.POST("reject", withdrawAuditApi.RejectWithdrawOrder)
}

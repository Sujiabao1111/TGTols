package example

import "github.com/gin-gonic/gin"

type InviteStatsRouter struct{}

// InitInviteStatsRouter initializes invite stats routes.
func (s *InviteStatsRouter) InitInviteStatsRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	inviteStatsRouterWithoutRecord := Router.Group("inviteStats")
	_ = PublicRouter

	inviteStatsRouterWithoutRecord.GET("getInviteStatsList", inviteStatsApi.GetInviteStatsList)
}

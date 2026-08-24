package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AgentRebateConfigsRouter struct {}

// InitAgentRebateConfigsRouter 初始化 agentRebateConfigs表 路由信息
func (s *AgentRebateConfigsRouter) InitAgentRebateConfigsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	agentRebateConfigsRouter := Router.Group("agentRebateConfigs").Use(middleware.OperationRecord())
	agentRebateConfigsRouterWithoutRecord := Router.Group("agentRebateConfigs")
	agentRebateConfigsRouterWithoutAuth := PublicRouter.Group("agentRebateConfigs")
	{
		agentRebateConfigsRouter.POST("createAgentRebateConfigs", agentRebateConfigsApi.CreateAgentRebateConfigs)   // 新建agentRebateConfigs表
		agentRebateConfigsRouter.DELETE("deleteAgentRebateConfigs", agentRebateConfigsApi.DeleteAgentRebateConfigs) // 删除agentRebateConfigs表
		agentRebateConfigsRouter.DELETE("deleteAgentRebateConfigsByIds", agentRebateConfigsApi.DeleteAgentRebateConfigsByIds) // 批量删除agentRebateConfigs表
		agentRebateConfigsRouter.PUT("updateAgentRebateConfigs", agentRebateConfigsApi.UpdateAgentRebateConfigs)    // 更新agentRebateConfigs表
	}
	{
		agentRebateConfigsRouterWithoutRecord.GET("findAgentRebateConfigs", agentRebateConfigsApi.FindAgentRebateConfigs)        // 根据ID获取agentRebateConfigs表
		agentRebateConfigsRouterWithoutRecord.GET("getAgentRebateConfigsList", agentRebateConfigsApi.GetAgentRebateConfigsList)  // 获取agentRebateConfigs表列表
	}
	{
	    agentRebateConfigsRouterWithoutAuth.GET("getAgentRebateConfigsPublic", agentRebateConfigsApi.GetAgentRebateConfigsPublic)  // agentRebateConfigs表开放接口
	}
}

import service from '@/utils/request'
// @Tags AgentRebateConfigs
// @Summary 创建agentRebateConfigs表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AgentRebateConfigs true "创建agentRebateConfigs表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /agentRebateConfigs/createAgentRebateConfigs [post]
export const createAgentRebateConfigs = (data) => {
  return service({
    url: '/agentRebateConfigs/createAgentRebateConfigs',
    method: 'post',
    data
  })
}

// @Tags AgentRebateConfigs
// @Summary 删除agentRebateConfigs表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AgentRebateConfigs true "删除agentRebateConfigs表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /agentRebateConfigs/deleteAgentRebateConfigs [delete]
export const deleteAgentRebateConfigs = (params) => {
  return service({
    url: '/agentRebateConfigs/deleteAgentRebateConfigs',
    method: 'delete',
    params
  })
}

// @Tags AgentRebateConfigs
// @Summary 批量删除agentRebateConfigs表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除agentRebateConfigs表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /agentRebateConfigs/deleteAgentRebateConfigs [delete]
export const deleteAgentRebateConfigsByIds = (params) => {
  return service({
    url: '/agentRebateConfigs/deleteAgentRebateConfigsByIds',
    method: 'delete',
    params
  })
}

// @Tags AgentRebateConfigs
// @Summary 更新agentRebateConfigs表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AgentRebateConfigs true "更新agentRebateConfigs表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /agentRebateConfigs/updateAgentRebateConfigs [put]
export const updateAgentRebateConfigs = (data) => {
  return service({
    url: '/agentRebateConfigs/updateAgentRebateConfigs',
    method: 'put',
    data
  })
}

// @Tags AgentRebateConfigs
// @Summary 用id查询agentRebateConfigs表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.AgentRebateConfigs true "用id查询agentRebateConfigs表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /agentRebateConfigs/findAgentRebateConfigs [get]
export const findAgentRebateConfigs = (params) => {
  return service({
    url: '/agentRebateConfigs/findAgentRebateConfigs',
    method: 'get',
    params
  })
}

// @Tags AgentRebateConfigs
// @Summary 分页获取agentRebateConfigs表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取agentRebateConfigs表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /agentRebateConfigs/getAgentRebateConfigsList [get]
export const getAgentRebateConfigsList = (params) => {
  return service({
    url: '/agentRebateConfigs/getAgentRebateConfigsList',
    method: 'get',
    params
  })
}

// @Tags AgentRebateConfigs
// @Summary 不需要鉴权的agentRebateConfigs表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.AgentRebateConfigsSearch true "分页获取agentRebateConfigs表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /agentRebateConfigs/getAgentRebateConfigsPublic [get]
export const getAgentRebateConfigsPublic = () => {
  return service({
    url: '/agentRebateConfigs/getAgentRebateConfigsPublic',
    method: 'get',
  })
}

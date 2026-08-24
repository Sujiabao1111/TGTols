import service from '@/utils/request'
// @Tags GameProviders
// @Summary 创建gameProviders表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.GameProviders true "创建gameProviders表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /gameProviders/createGameProviders [post]
export const createGameProviders = (data) => {
  return service({
    url: '/gameProviders/createGameProviders',
    method: 'post',
    data
  })
}

// @Tags GameProviders
// @Summary 删除gameProviders表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.GameProviders true "删除gameProviders表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /gameProviders/deleteGameProviders [delete]
export const deleteGameProviders = (params) => {
  return service({
    url: '/gameProviders/deleteGameProviders',
    method: 'delete',
    params
  })
}

// @Tags GameProviders
// @Summary 批量删除gameProviders表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除gameProviders表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /gameProviders/deleteGameProviders [delete]
export const deleteGameProvidersByIds = (params) => {
  return service({
    url: '/gameProviders/deleteGameProvidersByIds',
    method: 'delete',
    params
  })
}

// @Tags GameProviders
// @Summary 更新gameProviders表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.GameProviders true "更新gameProviders表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /gameProviders/updateGameProviders [put]
export const updateGameProviders = (data) => {
  return service({
    url: '/gameProviders/updateGameProviders',
    method: 'put',
    data
  })
}

// @Tags GameProviders
// @Summary 用id查询gameProviders表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.GameProviders true "用id查询gameProviders表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /gameProviders/findGameProviders [get]
export const findGameProviders = (params) => {
  return service({
    url: '/gameProviders/findGameProviders',
    method: 'get',
    params
  })
}

// @Tags GameProviders
// @Summary 分页获取gameProviders表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取gameProviders表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /gameProviders/getGameProvidersList [get]
export const getGameProvidersList = (params) => {
  return service({
    url: '/gameProviders/getGameProvidersList',
    method: 'get',
    params
  })
}

// @Tags GameProviders
// @Summary 不需要鉴权的gameProviders表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.GameProvidersSearch true "分页获取gameProviders表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /gameProviders/getGameProvidersPublic [get]
export const getGameProvidersPublic = () => {
  return service({
    url: '/gameProviders/getGameProvidersPublic',
    method: 'get',
  })
}

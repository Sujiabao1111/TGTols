import service from '@/utils/request'
// @Tags GameTypes
// @Summary 创建gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.GameTypes true "创建gameTypes表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /gameTypes/createGameTypes [post]
export const createGameTypes = (data) => {
  return service({
    url: '/gameTypes/createGameTypes',
    method: 'post',
    data
  })
}

// @Tags GameTypes
// @Summary 删除gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.GameTypes true "删除gameTypes表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /gameTypes/deleteGameTypes [delete]
export const deleteGameTypes = (params) => {
  return service({
    url: '/gameTypes/deleteGameTypes',
    method: 'delete',
    params
  })
}

// @Tags GameTypes
// @Summary 批量删除gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除gameTypes表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /gameTypes/deleteGameTypes [delete]
export const deleteGameTypesByIds = (params) => {
  return service({
    url: '/gameTypes/deleteGameTypesByIds',
    method: 'delete',
    params
  })
}

// @Tags GameTypes
// @Summary 更新gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.GameTypes true "更新gameTypes表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /gameTypes/updateGameTypes [put]
export const updateGameTypes = (data) => {
  return service({
    url: '/gameTypes/updateGameTypes',
    method: 'put',
    data
  })
}

// @Tags GameTypes
// @Summary 用id查询gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.GameTypes true "用id查询gameTypes表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /gameTypes/findGameTypes [get]
export const findGameTypes = (params) => {
  return service({
    url: '/gameTypes/findGameTypes',
    method: 'get',
    params
  })
}

// @Tags GameTypes
// @Summary 分页获取gameTypes表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取gameTypes表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /gameTypes/getGameTypesList [get]
export const getGameTypesList = (params) => {
  return service({
    url: '/gameTypes/getGameTypesList',
    method: 'get',
    params
  })
}

// @Tags GameTypes
// @Summary 不需要鉴权的gameTypes表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.GameTypesSearch true "分页获取gameTypes表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /gameTypes/getGameTypesPublic [get]
export const getGameTypesPublic = () => {
  return service({
    url: '/gameTypes/getGameTypesPublic',
    method: 'get',
  })
}

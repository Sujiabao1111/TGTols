import service from '@/utils/request'
// @Tags Games
// @Summary 创建games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Games true "创建games表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /games/createGames [post]
export const createGames = (data) => {
  return service({
    url: '/games/createGames',
    method: 'post',
    data
  })
}

// @Tags Games
// @Summary 删除games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Games true "删除games表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /games/deleteGames [delete]
export const deleteGames = (params) => {
  return service({
    url: '/games/deleteGames',
    method: 'delete',
    params
  })
}

// @Tags Games
// @Summary 批量删除games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除games表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /games/deleteGames [delete]
export const deleteGamesByIds = (params) => {
  return service({
    url: '/games/deleteGamesByIds',
    method: 'delete',
    params
  })
}

// @Tags Games
// @Summary 更新games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Games true "更新games表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /games/updateGames [put]
export const updateGames = (data) => {
  return service({
    url: '/games/updateGames',
    method: 'put',
    data
  })
}

// @Tags Games
// @Summary 用id查询games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.Games true "用id查询games表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /games/findGames [get]
export const findGames = (params) => {
  return service({
    url: '/games/findGames',
    method: 'get',
    params
  })
}

// @Tags Games
// @Summary 分页获取games表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取games表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /games/getGamesList [get]
export const getGamesList = (params) => {
  return service({
    url: '/games/getGamesList',
    method: 'get',
    params
  })
}

// @Tags Games
// @Summary 不需要鉴权的games表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.GamesSearch true "分页获取games表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /games/getGamesPublic [get]
export const getGamesPublic = () => {
  return service({
    url: '/games/getGamesPublic',
    method: 'get',
  })
}

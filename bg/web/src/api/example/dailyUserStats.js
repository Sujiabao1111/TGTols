import service from '@/utils/request'
// @Tags DailyUserStats
// @Summary 创建dailyUserStats表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.DailyUserStats true "创建dailyUserStats表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /dailyUserStats/createDailyUserStats [post]
export const createDailyUserStats = (data) => {
  return service({
    url: '/dailyUserStats/createDailyUserStats',
    method: 'post',
    data
  })
}

// @Tags DailyUserStats
// @Summary 删除dailyUserStats表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.DailyUserStats true "删除dailyUserStats表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /dailyUserStats/deleteDailyUserStats [delete]
export const deleteDailyUserStats = (params) => {
  return service({
    url: '/dailyUserStats/deleteDailyUserStats',
    method: 'delete',
    params
  })
}

// @Tags DailyUserStats
// @Summary 批量删除dailyUserStats表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除dailyUserStats表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /dailyUserStats/deleteDailyUserStats [delete]
export const deleteDailyUserStatsByIds = (params) => {
  return service({
    url: '/dailyUserStats/deleteDailyUserStatsByIds',
    method: 'delete',
    params
  })
}

// @Tags DailyUserStats
// @Summary 更新dailyUserStats表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.DailyUserStats true "更新dailyUserStats表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /dailyUserStats/updateDailyUserStats [put]
export const updateDailyUserStats = (data) => {
  return service({
    url: '/dailyUserStats/updateDailyUserStats',
    method: 'put',
    data
  })
}

// @Tags DailyUserStats
// @Summary 用id查询dailyUserStats表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.DailyUserStats true "用id查询dailyUserStats表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /dailyUserStats/findDailyUserStats [get]
export const findDailyUserStats = (params) => {
  return service({
    url: '/dailyUserStats/findDailyUserStats',
    method: 'get',
    params
  })
}

// @Tags DailyUserStats
// @Summary 分页获取dailyUserStats表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取dailyUserStats表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /dailyUserStats/getDailyUserStatsList [get]
export const getDailyUserStatsList = (params) => {
  return service({
    url: '/dailyUserStats/getDailyUserStatsList',
    method: 'get',
    params
  })
}

// @Tags DailyUserStats
// @Summary 不需要鉴权的dailyUserStats表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.DailyUserStatsSearch true "分页获取dailyUserStats表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /dailyUserStats/getDailyUserStatsPublic [get]
export const getDailyUserStatsPublic = () => {
  return service({
    url: '/dailyUserStats/getDailyUserStatsPublic',
    method: 'get',
  })
}

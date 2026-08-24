import service from '@/utils/request'
// @Tags StatsRetention
// @Summary 创建statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.StatsRetention true "创建statsRetention表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /statsRetention/createStatsRetention [post]
export const createStatsRetention = (data) => {
  return service({
    url: '/statsRetention/createStatsRetention',
    method: 'post',
    data
  })
}

// @Tags StatsRetention
// @Summary 删除statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.StatsRetention true "删除statsRetention表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /statsRetention/deleteStatsRetention [delete]
export const deleteStatsRetention = (params) => {
  return service({
    url: '/statsRetention/deleteStatsRetention',
    method: 'delete',
    params
  })
}

// @Tags StatsRetention
// @Summary 批量删除statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除statsRetention表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /statsRetention/deleteStatsRetention [delete]
export const deleteStatsRetentionByIds = (params) => {
  return service({
    url: '/statsRetention/deleteStatsRetentionByIds',
    method: 'delete',
    params
  })
}

// @Tags StatsRetention
// @Summary 更新statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.StatsRetention true "更新statsRetention表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /statsRetention/updateStatsRetention [put]
export const updateStatsRetention = (data) => {
  return service({
    url: '/statsRetention/updateStatsRetention',
    method: 'put',
    data
  })
}

// @Tags StatsRetention
// @Summary 用id查询statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.StatsRetention true "用id查询statsRetention表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /statsRetention/findStatsRetention [get]
export const findStatsRetention = (params) => {
  return service({
    url: '/statsRetention/findStatsRetention',
    method: 'get',
    params
  })
}

// @Tags StatsRetention
// @Summary 分页获取statsRetention表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取statsRetention表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /statsRetention/getStatsRetentionList [get]
export const getStatsRetentionList = (params) => {
  return service({
    url: '/statsRetention/getStatsRetentionList',
    method: 'get',
    params
  })
}

// @Tags StatsRetention
// @Summary 不需要鉴权的statsRetention表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.StatsRetentionSearch true "分页获取statsRetention表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /statsRetention/getStatsRetentionPublic [get]
export const getStatsRetentionPublic = () => {
  return service({
    url: '/statsRetention/getStatsRetentionPublic',
    method: 'get',
  })
}

import service from '@/utils/request'
// @Tags Banners
// @Summary 创建banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Banners true "创建banners表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /banners/createBanners [post]
export const createBanners = (data) => {
  return service({
    url: '/banners/createBanners',
    method: 'post',
    data
  })
}

// @Tags Banners
// @Summary 删除banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Banners true "删除banners表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /banners/deleteBanners [delete]
export const deleteBanners = (params) => {
  return service({
    url: '/banners/deleteBanners',
    method: 'delete',
    params
  })
}

// @Tags Banners
// @Summary 批量删除banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除banners表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /banners/deleteBanners [delete]
export const deleteBannersByIds = (params) => {
  return service({
    url: '/banners/deleteBannersByIds',
    method: 'delete',
    params
  })
}

// @Tags Banners
// @Summary 更新banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Banners true "更新banners表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /banners/updateBanners [put]
export const updateBanners = (data) => {
  return service({
    url: '/banners/updateBanners',
    method: 'put',
    data
  })
}

// @Tags Banners
// @Summary 用id查询banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.Banners true "用id查询banners表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /banners/findBanners [get]
export const findBanners = (params) => {
  return service({
    url: '/banners/findBanners',
    method: 'get',
    params
  })
}

// @Tags Banners
// @Summary 分页获取banners表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取banners表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /banners/getBannersList [get]
export const getBannersList = (params) => {
  return service({
    url: '/banners/getBannersList',
    method: 'get',
    params
  })
}

// @Tags Banners
// @Summary 不需要鉴权的banners表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.BannersSearch true "分页获取banners表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /banners/getBannersPublic [get]
export const getBannersPublic = () => {
  return service({
    url: '/banners/getBannersPublic',
    method: 'get',
  })
}

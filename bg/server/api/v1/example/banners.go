package example

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type BannersApi struct {}



// CreateBanners 创建banners表
// @Tags Banners
// @Summary 创建banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.Banners true "创建banners表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /banners/createBanners [post]
func (bannersApi *BannersApi) CreateBanners(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var banners example.Banners
	err := c.ShouldBindJSON(&banners)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = bannersService.CreateBanners(ctx,&banners)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteBanners 删除banners表
// @Tags Banners
// @Summary 删除banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.Banners true "删除banners表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /banners/deleteBanners [delete]
func (bannersApi *BannersApi) DeleteBanners(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := bannersService.DeleteBanners(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteBannersByIds 批量删除banners表
// @Tags Banners
// @Summary 批量删除banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /banners/deleteBannersByIds [delete]
func (bannersApi *BannersApi) DeleteBannersByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := bannersService.DeleteBannersByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateBanners 更新banners表
// @Tags Banners
// @Summary 更新banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.Banners true "更新banners表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /banners/updateBanners [put]
func (bannersApi *BannersApi) UpdateBanners(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var banners example.Banners
	err := c.ShouldBindJSON(&banners)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = bannersService.UpdateBanners(ctx,banners)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindBanners 用id查询banners表
// @Tags Banners
// @Summary 用id查询banners表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询banners表"
// @Success 200 {object} response.Response{data=example.Banners,msg=string} "查询成功"
// @Router /banners/findBanners [get]
func (bannersApi *BannersApi) FindBanners(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	rebanners, err := bannersService.GetBanners(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(rebanners, c)
}
// GetBannersList 分页获取banners表列表
// @Tags Banners
// @Summary 分页获取banners表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.BannersSearch true "分页获取banners表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /banners/getBannersList [get]
func (bannersApi *BannersApi) GetBannersList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo exampleReq.BannersSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := bannersService.GetBannersInfoList(ctx,pageInfo)
	if err != nil {
	    global.GVA_LOG.Error("获取失败!", zap.Error(err))
        response.FailWithMessage("获取失败:" + err.Error(), c)
        return
    }
    response.OkWithDetailed(response.PageResult{
        List:     list,
        Total:    total,
        Page:     pageInfo.Page,
        PageSize: pageInfo.PageSize,
    }, "获取成功", c)
}

// GetBannersPublic 不需要鉴权的banners表接口
// @Tags Banners
// @Summary 不需要鉴权的banners表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /banners/getBannersPublic [get]
func (bannersApi *BannersApi) GetBannersPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    bannersService.GetBannersPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的banners表接口信息",
    }, "获取成功", c)
}

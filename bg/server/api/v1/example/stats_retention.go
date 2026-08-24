package example

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type StatsRetentionApi struct {}



// CreateStatsRetention 创建statsRetention表
// @Tags StatsRetention
// @Summary 创建statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.StatsRetention true "创建statsRetention表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /statsRetention/createStatsRetention [post]
func (statsRetentionApi *StatsRetentionApi) CreateStatsRetention(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var statsRetention example.StatsRetention
	err := c.ShouldBindJSON(&statsRetention)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = statsRetentionService.CreateStatsRetention(ctx,&statsRetention)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteStatsRetention 删除statsRetention表
// @Tags StatsRetention
// @Summary 删除statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.StatsRetention true "删除statsRetention表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /statsRetention/deleteStatsRetention [delete]
func (statsRetentionApi *StatsRetentionApi) DeleteStatsRetention(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := statsRetentionService.DeleteStatsRetention(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteStatsRetentionByIds 批量删除statsRetention表
// @Tags StatsRetention
// @Summary 批量删除statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /statsRetention/deleteStatsRetentionByIds [delete]
func (statsRetentionApi *StatsRetentionApi) DeleteStatsRetentionByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := statsRetentionService.DeleteStatsRetentionByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateStatsRetention 更新statsRetention表
// @Tags StatsRetention
// @Summary 更新statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.StatsRetention true "更新statsRetention表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /statsRetention/updateStatsRetention [put]
func (statsRetentionApi *StatsRetentionApi) UpdateStatsRetention(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var statsRetention example.StatsRetention
	err := c.ShouldBindJSON(&statsRetention)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = statsRetentionService.UpdateStatsRetention(ctx,statsRetention)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindStatsRetention 用id查询statsRetention表
// @Tags StatsRetention
// @Summary 用id查询statsRetention表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询statsRetention表"
// @Success 200 {object} response.Response{data=example.StatsRetention,msg=string} "查询成功"
// @Router /statsRetention/findStatsRetention [get]
func (statsRetentionApi *StatsRetentionApi) FindStatsRetention(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	restatsRetention, err := statsRetentionService.GetStatsRetention(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(restatsRetention, c)
}
// GetStatsRetentionList 分页获取statsRetention表列表
// @Tags StatsRetention
// @Summary 分页获取statsRetention表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.StatsRetentionSearch true "分页获取statsRetention表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /statsRetention/getStatsRetentionList [get]
func (statsRetentionApi *StatsRetentionApi) GetStatsRetentionList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo exampleReq.StatsRetentionSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := statsRetentionService.GetStatsRetentionInfoList(ctx,pageInfo)
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

// GetStatsRetentionPublic 不需要鉴权的statsRetention表接口
// @Tags StatsRetention
// @Summary 不需要鉴权的statsRetention表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /statsRetention/getStatsRetentionPublic [get]
func (statsRetentionApi *StatsRetentionApi) GetStatsRetentionPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    statsRetentionService.GetStatsRetentionPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的statsRetention表接口信息",
    }, "获取成功", c)
}

package example

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type GameTypesApi struct {}



// CreateGameTypes 创建gameTypes表
// @Tags GameTypes
// @Summary 创建gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.GameTypes true "创建gameTypes表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /gameTypes/createGameTypes [post]
func (gameTypesApi *GameTypesApi) CreateGameTypes(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var gameTypes example.GameTypes
	err := c.ShouldBindJSON(&gameTypes)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = gameTypesService.CreateGameTypes(ctx,&gameTypes)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteGameTypes 删除gameTypes表
// @Tags GameTypes
// @Summary 删除gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.GameTypes true "删除gameTypes表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /gameTypes/deleteGameTypes [delete]
func (gameTypesApi *GameTypesApi) DeleteGameTypes(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := gameTypesService.DeleteGameTypes(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteGameTypesByIds 批量删除gameTypes表
// @Tags GameTypes
// @Summary 批量删除gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /gameTypes/deleteGameTypesByIds [delete]
func (gameTypesApi *GameTypesApi) DeleteGameTypesByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := gameTypesService.DeleteGameTypesByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateGameTypes 更新gameTypes表
// @Tags GameTypes
// @Summary 更新gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.GameTypes true "更新gameTypes表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /gameTypes/updateGameTypes [put]
func (gameTypesApi *GameTypesApi) UpdateGameTypes(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var gameTypes example.GameTypes
	err := c.ShouldBindJSON(&gameTypes)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = gameTypesService.UpdateGameTypes(ctx,gameTypes)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindGameTypes 用id查询gameTypes表
// @Tags GameTypes
// @Summary 用id查询gameTypes表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询gameTypes表"
// @Success 200 {object} response.Response{data=example.GameTypes,msg=string} "查询成功"
// @Router /gameTypes/findGameTypes [get]
func (gameTypesApi *GameTypesApi) FindGameTypes(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	regameTypes, err := gameTypesService.GetGameTypes(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(regameTypes, c)
}
// GetGameTypesList 分页获取gameTypes表列表
// @Tags GameTypes
// @Summary 分页获取gameTypes表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.GameTypesSearch true "分页获取gameTypes表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /gameTypes/getGameTypesList [get]
func (gameTypesApi *GameTypesApi) GetGameTypesList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo exampleReq.GameTypesSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := gameTypesService.GetGameTypesInfoList(ctx,pageInfo)
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

// GetGameTypesPublic 不需要鉴权的gameTypes表接口
// @Tags GameTypes
// @Summary 不需要鉴权的gameTypes表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /gameTypes/getGameTypesPublic [get]
func (gameTypesApi *GameTypesApi) GetGameTypesPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    gameTypesService.GetGameTypesPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的gameTypes表接口信息",
    }, "获取成功", c)
}

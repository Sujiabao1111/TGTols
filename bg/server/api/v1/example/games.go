package example

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type GamesApi struct {}



// CreateGames 创建games表
// @Tags Games
// @Summary 创建games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.Games true "创建games表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /games/createGames [post]
func (gamesApi *GamesApi) CreateGames(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var games example.Games
	err := c.ShouldBindJSON(&games)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = gamesService.CreateGames(ctx,&games)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteGames 删除games表
// @Tags Games
// @Summary 删除games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.Games true "删除games表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /games/deleteGames [delete]
func (gamesApi *GamesApi) DeleteGames(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := gamesService.DeleteGames(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteGamesByIds 批量删除games表
// @Tags Games
// @Summary 批量删除games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /games/deleteGamesByIds [delete]
func (gamesApi *GamesApi) DeleteGamesByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := gamesService.DeleteGamesByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateGames 更新games表
// @Tags Games
// @Summary 更新games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.Games true "更新games表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /games/updateGames [put]
func (gamesApi *GamesApi) UpdateGames(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var games example.Games
	err := c.ShouldBindJSON(&games)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = gamesService.UpdateGames(ctx,games)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindGames 用id查询games表
// @Tags Games
// @Summary 用id查询games表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询games表"
// @Success 200 {object} response.Response{data=example.Games,msg=string} "查询成功"
// @Router /games/findGames [get]
func (gamesApi *GamesApi) FindGames(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	regames, err := gamesService.GetGames(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(regames, c)
}
// GetGamesList 分页获取games表列表
// @Tags Games
// @Summary 分页获取games表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.GamesSearch true "分页获取games表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /games/getGamesList [get]
func (gamesApi *GamesApi) GetGamesList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo exampleReq.GamesSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := gamesService.GetGamesInfoList(ctx,pageInfo)
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

// GetGamesPublic 不需要鉴权的games表接口
// @Tags Games
// @Summary 不需要鉴权的games表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /games/getGamesPublic [get]
func (gamesApi *GamesApi) GetGamesPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    gamesService.GetGamesPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的games表接口信息",
    }, "获取成功", c)
}

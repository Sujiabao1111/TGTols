package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WithdrawAuditApi struct{}

func (withdrawAuditApi *WithdrawAuditApi) GetWithdrawAuditList(c *gin.Context) {
	ctx := c.Request.Context()

	var pageInfo exampleReq.WithdrawAuditSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, total, err := withdrawAuditService.GetWithdrawAuditList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("get withdraw audit list failed", zap.Error(err))
		response.FailWithMessage("获取提现审核列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

func (withdrawAuditApi *WithdrawAuditApi) ApproveWithdrawOrder(c *gin.Context) {
	ctx := c.Request.Context()

	var req exampleReq.ReviewWithdrawOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := withdrawAuditService.ApproveWithdrawOrder(ctx, req.OrderID, utils.GetUserID(c), utils.GetUserName(c), req.Remark); err != nil {
		global.GVA_LOG.Error("approve withdraw order failed", zap.Error(err))
		response.FailWithMessage("审核通过失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("审核通过成功", c)
}

func (withdrawAuditApi *WithdrawAuditApi) RejectWithdrawOrder(c *gin.Context) {
	ctx := c.Request.Context()

	var req exampleReq.ReviewWithdrawOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := withdrawAuditService.RejectWithdrawOrder(ctx, req.OrderID, utils.GetUserID(c), utils.GetUserName(c), req.Remark); err != nil {
		global.GVA_LOG.Error("reject withdraw order failed", zap.Error(err))
		response.FailWithMessage("驳回提现失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("驳回成功，余额已退回", c)
}

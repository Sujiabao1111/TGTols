import service from '@/utils/request'

// @Tags InviteStats
// @Summary 分页获取邀请统计列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取邀请统计列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /inviteStats/getInviteStatsList [get]
export const getInviteStatsList = (params) => {
  return service({
    url: '/inviteStats/getInviteStatsList',
    method: 'get',
    params
  })
}

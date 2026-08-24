import service from '@/utils/request'

export const getWithdrawAuditList = (params) => {
  return service({
    url: '/withdrawAudit/getWithdrawAuditList',
    method: 'get',
    params
  })
}

export const approveWithdrawOrder = (data) => {
  return service({
    url: '/withdrawAudit/approve',
    method: 'post',
    data
  })
}

export const rejectWithdrawOrder = (data) => {
  return service({
    url: '/withdrawAudit/reject',
    method: 'post',
    data
  })
}

import service from '@/utils/request'

export const getRechargeQueryList = (params) => {
  return service({
    url: '/rechargeQuery/getRechargeQueryList',
    method: 'get',
    params
  })
}

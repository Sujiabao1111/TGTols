import service from '@/utils/request'

export const getWithdrawFeeConfig = () => {
  return service({
    url: '/withdrawFeeConfig/getWithdrawFeeConfig',
    method: 'get'
  })
}

export const updateWithdrawFeeConfig = (data) => {
  return service({
    url: '/withdrawFeeConfig/updateWithdrawFeeConfig',
    method: 'put',
    data
  })
}

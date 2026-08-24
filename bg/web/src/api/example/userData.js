import service from '@/utils/request'

export const getUserDataList = (params) => {
  return service({
    url: '/userData/getUserDataList',
    method: 'get',
    params
  })
}

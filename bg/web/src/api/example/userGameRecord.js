import service from '@/utils/request'

export const getUserGameRecordList = (params) => {
  return service({
    url: '/userGameRecord/getUserGameRecordList',
    method: 'get',
    params
  })
}

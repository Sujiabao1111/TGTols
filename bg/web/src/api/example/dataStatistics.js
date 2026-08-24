import service from '@/utils/request'

export const getDataStatistics = (params) => service({
  url: '/dataStatistics/getDataStatistics',
  method: 'get',
  params
})

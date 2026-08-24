import service from '@/utils/request'

export const getDomainStats = (params) => service({
  url: '/admin/domain-stats',
  method: 'get',
  params
})

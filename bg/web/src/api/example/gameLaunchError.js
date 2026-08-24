import service from '@/utils/request'

export const getGameLaunchErrorList = (params) => {
  return service({
    url: '/gameLaunchError/getGameLaunchErrorList',
    method: 'get',
    params
  })
}

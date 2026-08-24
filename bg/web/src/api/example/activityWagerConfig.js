import service from '@/utils/request'

export const getActivityWagerConfig = () => {
  return service({
    url: '/activityWagerConfig/getActivityWagerConfig',
    method: 'get'
  })
}

export const updateActivityWagerConfig = (data) => {
  return service({
    url: '/activityWagerConfig/updateActivityWagerConfig',
    method: 'put',
    data
  })
}

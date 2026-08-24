import service from '@/utils/request'

export const grantDesktopReward = (data) => {
  return service({
    url: '/manualReward/grantDesktopReward',
    method: 'post',
    data
  })
}

export const grantReward = (data) => {
  return service({
    url: '/manualReward/grantReward',
    method: 'post',
    data
  })
}

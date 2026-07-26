import { orderFoodRequest } from './request'

export const getOrderFoodModerationWorkspace = () => {
  return orderFoodRequest({
    path: '/moderation-config',
    method: 'get'
  })
}

/**
 * 保存图片审核配置并立即生效。
 */
export const updateOrderFoodModerationConfig = (data) => {
  return orderFoodRequest({
    path: '/moderation-config',
    method: 'put',
    data,
    mutation: true
  })
}

export const testOrderFoodModerationConnection = (data) => {
  return orderFoodRequest({
    path: '/moderation-config/connection-tests',
    method: 'post',
    data,
    mutation: true
  })
}

export const testOrderFoodModerationImage = ({ expectedVersion, file }) => {
  const data = new FormData()
  if (expectedVersion) data.append('expectedVersion', String(expectedVersion))
  data.append('file', file, file.name || 'moderation-test-image')

  return orderFoodRequest({
    path: '/moderation-config/moderation-tests',
    method: 'post',
    headers: { 'Content-Type': 'multipart/form-data' },
    data,
    mutation: true
  })
}

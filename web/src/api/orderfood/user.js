import { orderFoodRequest } from './request'

/**
 * 分页查询小程序用户。
 */
export const getOrderFoodUserList = (params) => {
  return orderFoodRequest({
    path: '/users',
    method: 'get',
    params
  })
}

/**
 * 获取小程序用户基础详情。
 */
export const getOrderFoodUserDetail = (userId) => {
  return orderFoodRequest({
    path: `/users/${encodeURIComponent(userId)}`,
    method: 'get'
  })
}

/**
 * 禁用或恢复小程序用户。
 */
export const updateOrderFoodUserStatus = (userId, data) => {
  return orderFoodRequest({
    path: `/users/${encodeURIComponent(userId)}/status`,
    method: 'put',
    data,
    mutation: true
  })
}

/**
 * 获取聚合后的用户偏好画像。调用方必须先检查独立按钮权限。
 */
export const getOrderFoodUserPreferenceProfile = (userId) => {
  return orderFoodRequest({
    path: `/users/${encodeURIComponent(userId)}/preference-profile`,
    method: 'get'
  })
}

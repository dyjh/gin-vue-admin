import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

/**
 * 分页查询站内通知投递记录。
 */
export const getOrderFoodNotificationList = (params) =>
  orderFoodRequest({
    path: '/notifications',
    method: 'get',
    params
  })

/**
 * 获取站内通知投递详情。
 */
export const getOrderFoodNotificationDetail = (notificationId) =>
  orderFoodRequest({
    path: `/notifications/${encodePath(notificationId)}`,
    method: 'get'
  })

/**
 * 查询代码内定义的固定订阅消息场景及当前绑定。
 */
export const getSubscribeSceneList = () =>
  orderFoodRequest({
    path: '/subscribe-scenes',
    method: 'get'
  })

/**
 * 获取固定订阅消息场景及完整模板绑定。
 */
export const getSubscribeSceneDetail = (scene) =>
  orderFoodRequest({
    path: `/subscribe-scenes/${encodePath(scene)}`,
    method: 'get'
  })

/**
 * 创建或更新固定场景的唯一微信模板绑定。
 */
export const configureSubscribeScene = (scene, data) =>
  orderFoodRequest({
    path: `/subscribe-scenes/${encodePath(scene)}`,
    method: 'put',
    data,
    mutation: true
  })

/**
 * 独立启用或停用固定订阅场景。
 */
export const updateSubscribeSceneStatus = (scene, data) =>
  orderFoodRequest({
    path: `/subscribe-scenes/${encodePath(scene)}/status`,
    method: 'put',
    data,
    mutation: true
  })
/**
 * 分页查询订阅消息发送记录。
 */
export const getSubscribeLogList = (params) =>
  orderFoodRequest({
    path: '/subscribe-logs',
    method: 'get',
    params
  })

/**
 * 获取订阅消息发送详情。
 */
export const getSubscribeLogDetail = (logId) =>
  orderFoodRequest({
    path: `/subscribe-logs/${encodePath(logId)}`,
    method: 'get'
  })

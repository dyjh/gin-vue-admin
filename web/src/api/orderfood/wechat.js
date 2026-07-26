import { orderFoodRequest } from './request'

/**
 * 获取微信小程序当前配置。
 */
export const getOrderFoodWeChatConfig = () => {
  return orderFoodRequest({
    path: '/wechat-config',
    method: 'get'
  })
}

/**
 * 保存微信小程序配置并立即生效。
 */
export const updateOrderFoodWeChatConfig = (data) => {
  return orderFoodRequest({
    path: '/wechat-config',
    method: 'put',
    data,
    mutation: true
  })
}

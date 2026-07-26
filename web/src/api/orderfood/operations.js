import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

/**
 * 获取运营概览。
 */
export const getOrderFoodDashboard = (params) =>
  orderFoodRequest({
    path: '/dashboard',
    method: 'get',
    params
  })

/**
 * 分页查询AI调用记录。
 */
export const getAIUsageList = (params) =>
  orderFoodRequest({
    path: '/ai-usages',
    method: 'get',
    params
  })

/**
 * 获取AI调用记录详情。
 */
export const getAIUsageDetail = (usageId, params) =>
  orderFoodRequest({
    path: `/ai-usages/${encodePath(usageId)}`,
    method: 'get',
    params
  })

/**
 * 永久清除AI调用记录中的敏感内容。
 */
export const deleteAIUsageSensitiveContent = (usageId, data) =>
  orderFoodRequest({
    path: `/ai-usages/${encodePath(usageId)}/sensitive-content`,
    method: 'delete',
    data,
    mutation: true
  })

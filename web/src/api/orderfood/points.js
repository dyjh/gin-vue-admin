import { orderFoodRequest } from './request'

/**
 * 分页查询积分流水。
 */
export const getPointEntryList = (params) =>
  orderFoodRequest({
    path: '/point-entries',
    method: 'get',
    params
  })

/**
 * 预览人工积分调整。
 */
export const previewPointAdjustment = (data) =>
  orderFoodRequest({
    path: '/point-adjustments/preview',
    method: 'post',
    data
  })

/**
 * 提交人工积分调整。
 */
export const createPointAdjustment = (data) =>
  orderFoodRequest({
    path: '/point-adjustments',
    method: 'post',
    data,
    mutation: true
  })

export const getPointRule = () =>
  orderFoodRequest({
    path: '/point-rules',
    method: 'get'
  })

/**
 * 保存积分规则并立即生效。
 */
export const updatePointRule = (data) =>
  orderFoodRequest({
    path: '/point-rules',
    method: 'put',
    data,
    mutation: true
  })

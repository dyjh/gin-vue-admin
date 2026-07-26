import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

/**
 * 分页查询图片资源。
 */
export const getMediaList = (params) =>
  orderFoodRequest({
    path: '/media',
    method: 'get',
    params
  })

/**
 * 获取图片资源详情。
 */
export const getMediaDetail = (fileId) =>
  orderFoodRequest({
    path: `/media/${encodePath(fileId)}`,
    method: 'get'
  })

/**
 * 分页查询图片审核记录。
 */
export const getModerationRecordList = (params) =>
  orderFoodRequest({
    path: '/moderation-records',
    method: 'get',
    params
  })

/**
 * 获取图片审核记录详情。
 */
export const getModerationRecordDetail = (recordId) =>
  orderFoodRequest({
    path: `/moderation-records/${encodePath(recordId)}`,
    method: 'get'
  })

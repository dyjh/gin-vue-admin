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
 * 分页查询订阅消息模板。
 */
export const getSubscribeTemplateList = (params) =>
  orderFoodRequest({
    path: '/subscribe-templates',
    method: 'get',
    params
  })

/**
 * 获取订阅消息模板详情。
 */
export const getSubscribeTemplateDetail = (templateId) =>
  orderFoodRequest({
    path: `/subscribe-templates/${encodePath(templateId)}`,
    method: 'get'
  })

/**
 * 创建默认停用的订阅消息模板。
 */
export const createSubscribeTemplate = (data) =>
  orderFoodRequest({
    path: '/subscribe-templates',
    method: 'post',
    data,
    mutation: true
  })

/**
 * 编辑订阅消息模板但不改变启用状态。
 */
export const updateSubscribeTemplate = (templateId, data) =>
  orderFoodRequest({
    path: `/subscribe-templates/${encodePath(templateId)}`,
    method: 'put',
    data,
    mutation: true
  })

/**
 * 独立启用或停用订阅消息模板。
 */
export const updateSubscribeTemplateStatus = (templateId, data) =>
  orderFoodRequest({
    path: `/subscribe-templates/${encodePath(templateId)}/status`,
    method: 'put',
    data,
    mutation: true
  })

/**
 * 删除未启用且没有发送记录引用的订阅消息模板。
 */
export const deleteSubscribeTemplate = (templateId, data) =>
  orderFoodRequest({
    path: `/subscribe-templates/${encodePath(templateId)}`,
    method: 'delete',
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

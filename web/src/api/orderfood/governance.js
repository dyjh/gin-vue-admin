import { orderFoodRequest } from './request'

/**
 * 分页查询违规处理记录。
 */
export const getOrderFoodGovernanceRecordList = (params) => {
  return orderFoodRequest({
    path: '/governance-records',
    method: 'get',
    params
  })
}

/**
 * 获取违规处理记录详情。
 */
export const getOrderFoodGovernanceRecordDetail = (recordId) => {
  return orderFoodRequest({
    path: `/governance-records/${recordId}`,
    method: 'get'
  })
}

/**
 * 获取违规处理异步任务进度。
 */
export const getOrderFoodGovernanceJob = (jobId) => {
  return orderFoodRequest({
    path: `/governance-jobs/${jobId}`,
    method: 'get'
  })
}

/**
 * 重试违规处理异步任务失败项。
 */
export const retryOrderFoodGovernanceJob = (jobId, data) => {
  return orderFoodRequest({
    path: `/governance-jobs/${jobId}/retry`,
    method: 'post',
    data,
    mutation: true
  })
}

/**
 * 分页查询关联的管理员审计记录。
 */
export const getOrderFoodAuditLogList = (params) => {
  return orderFoodRequest({
    path: '/audit-logs',
    method: 'get',
    params
  })
}

/**
 * 获取管理员审计记录详情。
 */
export const getOrderFoodAuditLogDetail = (auditLogId) => {
  return orderFoodRequest({
    path: `/audit-logs/${encodeURIComponent(auditLogId)}`,
    method: 'get'
  })
}

/**
 * 预览违规处理影响。预览不改变业务状态。
 */
export const previewOrderFoodGovernanceAction = (data) => {
  return orderFoodRequest({
    path: '/governance-actions/preview',
    method: 'post',
    data
  })
}

/**
 * 执行已经预览确认的违规处理。
 */
export const executeOrderFoodGovernanceAction = (data) => {
  return orderFoodRequest({
    path: '/governance-actions',
    method: 'post',
    data,
    mutation: true
  })
}

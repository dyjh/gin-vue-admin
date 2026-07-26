import service from '@/utils/request'
import { CreateUUID } from '@/utils/format'

/**
 * 后台管理端对外路径。项目的 request 实例已通过 VITE_BASE_API 注入 `/api`，
 * 因此实际请求 path 需要去掉这一层，最终浏览器请求仍为该完整地址。
 */
export const ORDER_FOOD_ADMIN_BASE = '/api/orderfood'
const ORDER_FOOD_ADMIN_PATH = ORDER_FOOD_ADMIN_BASE.replace(/^\/api/, '')

const CONFLICT_CODES = new Set([20901, 20902, 20904])

export const orderFoodRequest = ({
  path,
  mutation = false,
  headers = {},
  ...config
}) => {
  return service({
    ...config,
    url: `${ORDER_FOOD_ADMIN_PATH}${path}`,
    headers: {
      'X-Request-Id': CreateUUID(),
      ...(mutation ? { 'X-Idempotency-Key': CreateUUID() } : {}),
      ...headers
    }
  })
}

export const isOrderFoodSuccess = (response) => response?.code === 0

// getOrderFoodErrorCode 从业务响应或网络错误中提取统一错误码。
export const getOrderFoodErrorCode = (errorOrResponse) =>
  Number(
    errorOrResponse?.orderFoodResponse?.code ??
      errorOrResponse?.response?.data?.code ??
      errorOrResponse?.code
  )

export const unwrapOrderFoodResponse = (response, fallback = '请求失败') => {
  if (isOrderFoodSuccess(response)) {
    return response.data
  }

  const error = new Error(response?.msg || fallback)
  error.orderFoodResponse = response
  throw error
}

export const isOrderFoodConflict = (errorOrResponse) => {
  const status = errorOrResponse?.response?.status
  const code = getOrderFoodErrorCode(errorOrResponse)
  return status === 409 || CONFLICT_CODES.has(code)
}

export const getOrderFoodErrorMessage = (
  errorOrResponse,
  fallback = '请求失败，请稍后重试'
) => {
  return (
    errorOrResponse?.orderFoodResponse?.msg ||
    errorOrResponse?.response?.data?.msg ||
    errorOrResponse?.msg ||
    errorOrResponse?.message ||
    fallback
  )
}

import { orderFoodRequest } from './request'

/**
 * 分页查询全部用户菜品摘要。
 */
export const getOrderFoodUserDishList = (params) => {
  return orderFoodRequest({
    path: '/user-dishes',
    method: 'get',
    params
  })
}

/**
 * 获取用户菜品私有详情。调用方必须先检查 private-read 按钮权限。
 */
export const getOrderFoodUserDishDetail = (dishId) => {
  return orderFoodRequest({
    path: `/user-dishes/${encodeURIComponent(dishId)}`,
    method: 'get'
  })
}

/**
 * 分页查询用户菜品引用位置。
 */
export const getOrderFoodUserDishReferences = (dishId, params) => {
  return orderFoodRequest({
    path: `/user-dishes/${encodeURIComponent(dishId)}/references`,
    method: 'get',
    params
  })
}

import { orderFoodRequest } from './request'

/**
 * 分页查询全部用户菜谱摘要。
 */
export const getOrderFoodUserRecipeList = (params) => {
  return orderFoodRequest({
    path: '/user-recipes',
    method: 'get',
    params
  })
}

/**
 * 获取用户菜谱私有详情。调用方必须先检查 private-read 按钮权限。
 */
export const getOrderFoodUserRecipeDetail = (recipeId) => {
  return orderFoodRequest({
    path: `/user-recipes/${encodeURIComponent(recipeId)}`,
    method: 'get'
  })
}

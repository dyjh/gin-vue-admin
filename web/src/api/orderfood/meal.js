import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

/**
 * 分页查询饭局。
 */
export const getMealAdminList = (params) =>
  orderFoodRequest({
    path: '/meals',
    method: 'get',
    params
  })

/**
 * 获取饭局详情。
 */
export const getMealAdminDetail = (mealId) =>
  orderFoodRequest({
    path: `/meals/${encodePath(mealId)}`,
    method: 'get'
  })

/**
 * 分页查询采购清单。
 */
export const getShoppingListAdminList = (params) =>
  orderFoodRequest({
    path: '/shopping-lists',
    method: 'get',
    params
  })

/**
 * 获取采购清单详情。
 */
export const getShoppingListAdminDetail = (shoppingListId) =>
  orderFoodRequest({
    path: `/shopping-lists/${encodePath(shoppingListId)}`,
    method: 'get'
  })

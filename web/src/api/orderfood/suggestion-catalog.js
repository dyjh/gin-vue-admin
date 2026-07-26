import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

export const getSuggestionCatalogWorkspace = () =>
  orderFoodRequest({
    path: '/suggestion-catalog',
    method: 'get'
  })

/**
 * 保存生成结果校验策略并立即生效。
 */
export const updateSuggestionValidationPolicy = (data) =>
  orderFoodRequest({
    path: '/suggestion-catalog/policy',
    method: 'put',
    data,
    mutation: true
  })

export const getStandardIngredients = (params) =>
  orderFoodRequest({
    path: '/suggestion-catalog/ingredients',
    method: 'get',
    params
  })

export const createStandardIngredient = (data) =>
  orderFoodRequest({
    path: '/suggestion-catalog/ingredients',
    method: 'post',
    data,
    mutation: true
  })

export const updateStandardIngredient = (id, data) =>
  orderFoodRequest({
    path: `/suggestion-catalog/ingredients/${encodePath(id)}`,
    method: 'put',
    data,
    mutation: true
  })

export const getStandardDishes = (params) =>
  orderFoodRequest({
    path: '/suggestion-catalog/dishes',
    method: 'get',
    params
  })

export const createStandardDish = (data) =>
  orderFoodRequest({
    path: '/suggestion-catalog/dishes',
    method: 'post',
    data,
    mutation: true
  })

export const updateStandardDish = (id, data) =>
  orderFoodRequest({
    path: `/suggestion-catalog/dishes/${encodePath(id)}`,
    method: 'put',
    data,
    mutation: true
  })

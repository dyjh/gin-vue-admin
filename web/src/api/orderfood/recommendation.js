import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

/**
 * 分页查询可发现候选菜品。
 */
export const getDiscoverableDishList = (params) =>
  orderFoodRequest({
    path: '/discoverable-dishes',
    method: 'get',
    params
  })

/**
 * 获取可发现候选菜品详情。
 */
export const getDiscoverableDishDetail = (dishId, params) =>
  orderFoodRequest({
    path: `/discoverable-dishes/${encodePath(dishId)}`,
    method: 'get',
    params
  })

/**
 * 分页查询推荐精选。
 */
export const getRecommendationList = (params) =>
  orderFoodRequest({
    path: '/recommendations',
    method: 'get',
    params
  })

/**
 * 获取推荐精选详情。
 */
export const getRecommendationDetail = (recommendationId) =>
  orderFoodRequest({
    path: `/recommendations/${encodePath(recommendationId)}`,
    method: 'get'
  })

/**
 * 创建推荐草稿。
 */
export const createRecommendation = (data) =>
  orderFoodRequest({
    path: '/recommendations',
    method: 'post',
    data,
    mutation: true
  })

/**
 * 编辑推荐排序和展示说明。
 */
export const updateRecommendation = (recommendationId, data) =>
  orderFoodRequest({
    path: `/recommendations/${encodePath(recommendationId)}`,
    method: 'put',
    data,
    mutation: true
  })

/**
 * 删除推荐草稿。
 */
export const deleteRecommendation = (recommendationId, data) =>
  orderFoodRequest({
    path: `/recommendations/${encodePath(recommendationId)}`,
    method: 'delete',
    data,
    mutation: true
  })

/**
 * 发布或重新发布推荐。
 */
export const publishRecommendation = (recommendationId, data) =>
  orderFoodRequest({
    path: `/recommendations/${encodePath(recommendationId)}/publish`,
    method: 'post',
    data,
    mutation: true
  })

/**
 * 下线推荐。
 */
export const offlineRecommendation = (recommendationId, data) =>
  orderFoodRequest({
    path: `/recommendations/${encodePath(recommendationId)}/offline`,
    method: 'post',
    data,
    mutation: true
  })

/**
 * 批量调整推荐排序。
 */
export const updateRecommendationSortOrder = (data) =>
  orderFoodRequest({
    path: '/recommendations/sort-order',
    method: 'put',
    data,
    mutation: true
  })

import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

/**
 * 分页查询官方菜品。
 */
export const getOfficialDishList = (params) =>
  orderFoodRequest({
    path: '/official-dishes',
    method: 'get',
    params
  })

/**
 * 获取官方菜品详情。
 */
export const getOfficialDishDetail = (dishId) =>
  orderFoodRequest({
    path: `/official-dishes/${encodePath(dishId)}`,
    method: 'get'
  })

/**
 * 创建官方菜品。
 */
export const createOfficialDish = (data) =>
  orderFoodRequest({
    path: '/official-dishes',
    method: 'post',
    data,
    mutation: true
  })

/**
 * 编辑官方菜品。
 */
export const updateOfficialDish = (dishId, data) =>
  orderFoodRequest({
    path: `/official-dishes/${encodePath(dishId)}`,
    method: 'put',
    data,
    mutation: true
  })

/**
 * 软删除官方菜品。
 */
export const deleteOfficialDish = (dishId, data) =>
  orderFoodRequest({
    path: `/official-dishes/${encodePath(dishId)}`,
    method: 'delete',
    data,
    mutation: true
  })

/**
 * 分页查询官方菜品封面。
 */
export const getOfficialDishCoverList = (params) =>
  orderFoodRequest({
    path: '/official-dish-covers',
    method: 'get',
    params
  })

/**
 * 上传官方菜品封面。该接口只做文件安全校验，不调用外部图片审核。
 */
export const uploadOfficialDishCover = (file) => {
  const data = new FormData()
  data.append('file', file)
  return orderFoodRequest({
    path: '/official-dish-covers',
    method: 'post',
    data,
    mutation: true,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

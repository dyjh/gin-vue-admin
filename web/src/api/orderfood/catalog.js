import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)
const resourcePaths = {
  category: 'categories',
  tag: 'tags',
  unit: 'units'
}

const getResourcePath = (resource) => {
  const path = resourcePaths[resource]
  if (!path) throw new Error('不支持的基础数据类型')
  return path
}

/**
 * 分页查询分类、标签或单位。
 */
export const getCatalogItemList = (resource, params) =>
  orderFoodRequest({
    path: `/${getResourcePath(resource)}`,
    method: 'get',
    params
  })

/**
 * 新增分类、标签或单位。
 */
export const createCatalogItem = (resource, data) =>
  orderFoodRequest({
    path: `/${getResourcePath(resource)}`,
    method: 'post',
    data,
    mutation: true
  })

/**
 * 编辑分类、标签或单位。
 */
export const updateCatalogItem = (resource, id, data) =>
  orderFoodRequest({
    path: `/${getResourcePath(resource)}/${encodePath(id)}`,
    method: 'put',
    data,
    mutation: true
  })

/**
 * 删除未被引用的分类、标签或单位。
 */
export const deleteCatalogItem = (resource, id, data) =>
  orderFoodRequest({
    path: `/${getResourcePath(resource)}/${encodePath(id)}`,
    method: 'delete',
    data,
    mutation: true
  })

/**
 * 批量调整分类、标签或单位的排序。
 */
export const updateCatalogSortOrder = (resource, data) =>
  orderFoodRequest({
    path: `/${getResourcePath(resource)}/sort-order`,
    method: 'put',
    data,
    mutation: true
  })

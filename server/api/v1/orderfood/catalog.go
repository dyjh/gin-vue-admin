package orderfood

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/gin-gonic/gin"
)

// CatalogApi 提供分类、标签和单位的管理端接口处理能力。
type CatalogApi struct {
	service *orderfoodService.CatalogService // 基础数据业务服务
}

// NewCatalogApi 创建基础数据API实例。
func NewCatalogApi(service *orderfoodService.CatalogService) *CatalogApi {
	return &CatalogApi{service: service}
}

// ListCategories 分页查询分类
// @Tags OrderFoodCatalog
// @Summary 分页查询分类
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.CatalogItemListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.CatalogItem},msg=string} "获取成功"
// @Router /orderfood/categories [get]
func (api *CatalogApi) ListCategories(c *gin.Context) {
	var query orderfoodRequest.CatalogItemListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionCategoryRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.CatalogItem]
	result, err := api.service.ListCategories(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateCategory 新增分类
// @Tags OrderFoodCatalog
// @Summary 新增分类
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.CatalogItemCreateInput true "分类参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogItem,msg=string} "操作成功"
// @Router /orderfood/categories [post]
func (api *CatalogApi) CreateCategory(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogItemCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionCategoryCreate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateCategory(c.Request.Context(), actor, header.Key, body)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateCategory 编辑分类
// @Tags OrderFoodCatalog
// @Summary 编辑分类
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param categoryId path string true "分类ID"
// @Param data body orderfoodRequest.CatalogItemUpdateInput true "分类参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogItem,msg=string} "操作成功"
// @Router /orderfood/categories/{categoryId} [put]
func (api *CatalogApi) UpdateCategory(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindCategoryIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogItemUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionCategoryUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateCategory(
		c.Request.Context(), path.CategoryID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// DeleteCategory 删除分类
// @Tags OrderFoodCatalog
// @Summary 删除未被业务引用的分类
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param categoryId path string true "分类ID"
// @Param data body orderfoodRequest.CatalogDeleteInput true "删除参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogDeleteResult,msg=string} "操作成功"
// @Router /orderfood/categories/{categoryId} [delete]
func (api *CatalogApi) DeleteCategory(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindCategoryIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionCategoryDelete) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteCategory(
		c.Request.Context(), path.CategoryID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateCategorySortOrder 批量调整分类排序
// @Tags OrderFoodCatalog
// @Summary 批量调整分类排序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.CatalogSortOrderInput true "分类排序参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogSortOrderResult,msg=string} "操作成功"
// @Router /orderfood/categories/sort-order [put]
func (api *CatalogApi) UpdateCategorySortOrder(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogSortOrderInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionCategorySort) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateCategorySortOrder(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListTags 分页查询标签
// @Tags OrderFoodCatalog
// @Summary 分页查询标签
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.CatalogItemListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.CatalogItem},msg=string} "获取成功"
// @Router /orderfood/tags [get]
func (api *CatalogApi) ListTags(c *gin.Context) {
	var query orderfoodRequest.CatalogItemListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionTagRead) {
		return
	}
	result, err := api.service.ListTags(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateTag 新增标签
// @Tags OrderFoodCatalog
// @Summary 新增标签
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.CatalogItemCreateInput true "标签参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogItem,msg=string} "操作成功"
// @Router /orderfood/tags [post]
func (api *CatalogApi) CreateTag(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogItemCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionTagCreate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateTag(c.Request.Context(), actor, header.Key, body)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateTag 编辑标签
// @Tags OrderFoodCatalog
// @Summary 编辑标签
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param tagId path string true "标签ID"
// @Param data body orderfoodRequest.CatalogItemUpdateInput true "标签参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogItem,msg=string} "操作成功"
// @Router /orderfood/tags/{tagId} [put]
func (api *CatalogApi) UpdateTag(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindTagIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogItemUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionTagUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateTag(
		c.Request.Context(), path.TagID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// DeleteTag 删除标签
// @Tags OrderFoodCatalog
// @Summary 删除未被业务引用的标签
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param tagId path string true "标签ID"
// @Param data body orderfoodRequest.CatalogDeleteInput true "删除参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogDeleteResult,msg=string} "操作成功"
// @Router /orderfood/tags/{tagId} [delete]
func (api *CatalogApi) DeleteTag(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindTagIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionTagDelete) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteTag(
		c.Request.Context(), path.TagID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateTagSortOrder 批量调整标签排序
// @Tags OrderFoodCatalog
// @Summary 批量调整标签排序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.CatalogSortOrderInput true "标签排序参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogSortOrderResult,msg=string} "操作成功"
// @Router /orderfood/tags/sort-order [put]
func (api *CatalogApi) UpdateTagSortOrder(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogSortOrderInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionTagSort) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateTagSortOrder(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListUnits 分页查询单位
// @Tags OrderFoodCatalog
// @Summary 分页查询单位
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.CatalogItemListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.CatalogItem},msg=string} "获取成功"
// @Router /orderfood/units [get]
func (api *CatalogApi) ListUnits(c *gin.Context) {
	var query orderfoodRequest.CatalogItemListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionUnitRead) {
		return
	}
	result, err := api.service.ListUnits(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateUnit 新增单位
// @Tags OrderFoodCatalog
// @Summary 新增单位
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.CatalogItemCreateInput true "单位参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogItem,msg=string} "操作成功"
// @Router /orderfood/units [post]
func (api *CatalogApi) CreateUnit(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogItemCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionUnitCreate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateUnit(c.Request.Context(), actor, header.Key, body)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateUnit 编辑单位
// @Tags OrderFoodCatalog
// @Summary 编辑单位
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param unitId path string true "单位ID"
// @Param data body orderfoodRequest.CatalogItemUpdateInput true "单位参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogItem,msg=string} "操作成功"
// @Router /orderfood/units/{unitId} [put]
func (api *CatalogApi) UpdateUnit(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindUnitIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogItemUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionUnitUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateUnit(
		c.Request.Context(), path.UnitID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// DeleteUnit 删除单位
// @Tags OrderFoodCatalog
// @Summary 删除未被业务引用的单位
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param unitId path string true "单位ID"
// @Param data body orderfoodRequest.CatalogDeleteInput true "删除参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogDeleteResult,msg=string} "操作成功"
// @Router /orderfood/units/{unitId} [delete]
func (api *CatalogApi) DeleteUnit(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindUnitIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionUnitDelete) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteUnit(
		c.Request.Context(), path.UnitID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateUnitSortOrder 批量调整单位排序
// @Tags OrderFoodCatalog
// @Summary 批量调整单位排序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.CatalogSortOrderInput true "单位排序参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.CatalogSortOrderResult,msg=string} "操作成功"
// @Router /orderfood/units/sort-order [put]
func (api *CatalogApi) UpdateUnitSortOrder(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.CatalogSortOrderInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionUnitSort) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateUnitSortOrder(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindCategoryIDPath 绑定并规范化分类ID路径参数。
func bindCategoryIDPath(c *gin.Context) (orderfoodRequest.CategoryIDPath, bool) {
	var path orderfoodRequest.CategoryIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.CategoryIDPath{}, false
	}
	path.CategoryID = strings.TrimSpace(path.CategoryID)
	if path.CategoryID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.CategoryIDPath{}, false
	}
	return path, true
}

// bindTagIDPath 绑定并规范化标签ID路径参数。
func bindTagIDPath(c *gin.Context) (orderfoodRequest.TagIDPath, bool) {
	var path orderfoodRequest.TagIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.TagIDPath{}, false
	}
	path.TagID = strings.TrimSpace(path.TagID)
	if path.TagID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.TagIDPath{}, false
	}
	return path, true
}

// bindUnitIDPath 绑定并规范化单位ID路径参数。
func bindUnitIDPath(c *gin.Context) (orderfoodRequest.UnitIDPath, bool) {
	var path orderfoodRequest.UnitIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.UnitIDPath{}, false
	}
	path.UnitID = strings.TrimSpace(path.UnitID)
	if path.UnitID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.UnitIDPath{}, false
	}
	return path, true
}

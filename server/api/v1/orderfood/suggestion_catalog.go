package orderfood

import (
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/gin-gonic/gin"
)

// SuggestionCatalogApi 提供推荐菜索引接口处理能力。
type SuggestionCatalogApi struct {
	service *orderfoodService.SuggestionCatalogService
}

// NewSuggestionCatalogApi 创建推荐菜索引API实例。
func NewSuggestionCatalogApi(service *orderfoodService.SuggestionCatalogService) *SuggestionCatalogApi {
	return &SuggestionCatalogApi{service: service}
}

// GetWorkspace 获取标准菜品索引、食材词库和生成校验策略概况
// @Tags OrderFoodSuggestionCatalog
// @Summary 获取标准菜品索引、食材词库和生成校验策略概况
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=orderfoodResponse.SuggestionCatalogWorkspace,msg=string} "获取成功"
// @Router /orderfood/suggestion-catalog [get]
func (api *SuggestionCatalogApi) GetWorkspace(c *gin.Context) {
	var query orderfoodRequest.SuggestionCatalogEmptyQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogRead) {
		return
	}
	var result orderfoodResponse.SuggestionCatalogWorkspace
	result, err := api.service.Workspace(c.Request.Context())
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdatePolicy 保存并立即应用生成结果索引校验策略
// @Tags OrderFoodSuggestionCatalog
// @Summary 保存并立即应用生成结果索引校验策略
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.SuggestionValidationPolicyUpdateInput true "校验策略"
// @Success 200 {object} response.Response{data=orderfoodResponse.SuggestionValidationPolicyConfig,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/policy [put]
func (api *SuggestionCatalogApi) UpdatePolicy(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.SuggestionValidationPolicyUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdatePolicy(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListIngredients 分页查询标准食材词库
// @Tags OrderFoodSuggestionCatalog
// @Summary 分页查询标准食材词库
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.StandardIngredient},msg=string} "获取成功"
// @Router /orderfood/suggestion-catalog/ingredients [get]
func (api *SuggestionCatalogApi) ListIngredients(c *gin.Context) {
	var query orderfoodRequest.StandardCatalogListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogRead) {
		return
	}
	result, err := api.service.ListIngredients(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateIngredient 新增标准食材
// @Tags OrderFoodSuggestionCatalog
// @Summary 新增标准食材
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.StandardIngredientInput true "标准食材"
// @Success 200 {object} response.Response{data=orderfoodResponse.StandardIngredient,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/ingredients [post]
func (api *SuggestionCatalogApi) CreateIngredient(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.StandardIngredientInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateIngredient(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateIngredient 编辑标准食材
// @Tags OrderFoodSuggestionCatalog
// @Summary 编辑标准食材
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param id path string true "标准食材 ID"
// @Param data body orderfoodRequest.StandardIngredientInput true "标准食材"
// @Success 200 {object} response.Response{data=orderfoodResponse.StandardIngredient,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/ingredients/{id} [put]
func (api *SuggestionCatalogApi) UpdateIngredient(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path orderfoodRequest.StandardCatalogIDPath
	var body orderfoodRequest.StandardIngredientInput
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateIngredient(
		c.Request.Context(), actor, header.Key, path.ID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListDishes 分页查询标准菜品索引
// @Tags OrderFoodSuggestionCatalog
// @Summary 分页查询标准菜品索引
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.StandardDish},msg=string} "获取成功"
// @Router /orderfood/suggestion-catalog/dishes [get]
func (api *SuggestionCatalogApi) ListDishes(c *gin.Context) {
	var query orderfoodRequest.StandardCatalogListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogRead) {
		return
	}
	result, err := api.service.ListDishes(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateDish 新增标准菜品索引
// @Tags OrderFoodSuggestionCatalog
// @Summary 新增标准菜品索引
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.StandardDishInput true "标准菜品索引"
// @Success 200 {object} response.Response{data=orderfoodResponse.StandardDish,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/dishes [post]
func (api *SuggestionCatalogApi) CreateDish(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.StandardDishInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateDish(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateDish 编辑标准菜品索引
// @Tags OrderFoodSuggestionCatalog
// @Summary 编辑标准菜品索引
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param id path string true "标准菜品索引 ID"
// @Param data body orderfoodRequest.StandardDishInput true "标准菜品索引"
// @Success 200 {object} response.Response{data=orderfoodResponse.StandardDish,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/dishes/{id} [put]
func (api *SuggestionCatalogApi) UpdateDish(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path orderfoodRequest.StandardCatalogIDPath
	var body orderfoodRequest.StandardDishInput
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateDish(
		c.Request.Context(), actor, header.Key, path.ID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

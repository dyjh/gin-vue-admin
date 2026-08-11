package dish

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	dishResponse "github.com/dyjh/order-food-mini-app/server/model/dish/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/dish"
	"github.com/gin-gonic/gin"
)

// SuggestionCatalogApi 提供推荐菜索引接口处理能力。
type SuggestionCatalogApi struct{}

// GetWorkspace 获取标准菜品索引、食材词库和生成校验策略概况
// @Tags OrderFoodSuggestionCatalog
// @Summary 获取标准菜品索引、食材词库和生成校验策略概况
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=dishResponse.SuggestionCatalogWorkspace,msg=string} "获取成功"
// @Router /orderfood/suggestion-catalog [get]
func (api *SuggestionCatalogApi) GetWorkspace(c *gin.Context) {
	var query dishRequest.SuggestionCatalogEmptyQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogRead) {
		return
	}
	var result dishResponse.SuggestionCatalogWorkspace
	result, err := suggestionCatalogService.Workspace(c.Request.Context())
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
// @Param data body dishRequest.SuggestionValidationPolicyUpdateInput true "校验策略"
// @Success 200 {object} response.Response{data=dishResponse.SuggestionValidationPolicyConfig,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/policy [put]
func (api *SuggestionCatalogApi) UpdatePolicy(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body dishRequest.SuggestionValidationPolicyUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := suggestionCatalogService.UpdatePolicy(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListIngredients 分页查询标准食材词库
// @Tags OrderFoodSuggestionCatalog
// @Summary 分页查询标准食材词库
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=response.PageResult{list=[]dishResponse.StandardIngredient},msg=string} "获取成功"
// @Router /orderfood/suggestion-catalog/ingredients [get]
func (api *SuggestionCatalogApi) ListIngredients(c *gin.Context) {
	var query dishRequest.StandardCatalogListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogRead) {
		return
	}
	result, err := suggestionCatalogService.ListIngredients(c.Request.Context(), query)
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
// @Param data body dishRequest.StandardIngredientInput true "标准食材"
// @Success 200 {object} response.Response{data=dishResponse.StandardIngredient,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/ingredients [post]
func (api *SuggestionCatalogApi) CreateIngredient(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body dishRequest.StandardIngredientInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := suggestionCatalogService.CreateIngredient(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
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
// @Param data body dishRequest.StandardIngredientInput true "标准食材"
// @Success 200 {object} response.Response{data=dishResponse.StandardIngredient,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/ingredients/{id} [put]
func (api *SuggestionCatalogApi) UpdateIngredient(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path dishRequest.StandardCatalogIDPath
	var body dishRequest.StandardIngredientInput
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := suggestionCatalogService.UpdateIngredient(
		c.Request.Context(), actor, header.Key, path.ID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListDishes 分页查询标准菜品索引
// @Tags OrderFoodSuggestionCatalog
// @Summary 分页查询标准菜品索引
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=response.PageResult{list=[]dishResponse.StandardDish},msg=string} "获取成功"
// @Router /orderfood/suggestion-catalog/dishes [get]
func (api *SuggestionCatalogApi) ListDishes(c *gin.Context) {
	var query dishRequest.StandardCatalogListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogRead) {
		return
	}
	result, err := suggestionCatalogService.ListDishes(c.Request.Context(), query)
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
// @Param data body dishRequest.StandardDishInput true "标准菜品索引"
// @Success 200 {object} response.Response{data=dishResponse.StandardDish,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/dishes [post]
func (api *SuggestionCatalogApi) CreateDish(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body dishRequest.StandardDishInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := suggestionCatalogService.CreateDish(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
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
// @Param data body dishRequest.StandardDishInput true "标准菜品索引"
// @Success 200 {object} response.Response{data=dishResponse.StandardDish,msg=string} "操作成功"
// @Router /orderfood/suggestion-catalog/dishes/{id} [put]
func (api *SuggestionCatalogApi) UpdateDish(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path dishRequest.StandardCatalogIDPath
	var body dishRequest.StandardDishInput
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSuggestionCatalogUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := suggestionCatalogService.UpdateDish(
		c.Request.Context(), actor, header.Key, path.ID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

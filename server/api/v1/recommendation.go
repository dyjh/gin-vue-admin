package v1

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/gin-gonic/gin"
)

// RecommendationApi 提供可发现候选菜品和推荐精选接口处理能力。
type RecommendationApi struct {
	service *orderfoodService.RecommendationService // 推荐管理业务服务
}

// NewRecommendationApi 创建推荐管理API实例。
func NewRecommendationApi(service *orderfoodService.RecommendationService) *RecommendationApi {
	return &RecommendationApi{service: service}
}

// ListDiscoverableDishes 分页查询可发现候选菜品
// @Tags OrderFoodRecommendation
// @Summary 分页查询可发现候选菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.DiscoverableDishListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.DiscoverableDishSummary},msg=string} "获取成功"
// @Router /orderfood/discoverable-dishes [get]
func (api *RecommendationApi) ListDiscoverableDishes(c *gin.Context) {
	var query orderfoodRequest.DiscoverableDishListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionDiscoverableDishRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.DiscoverableDishSummary]
	result, err := api.service.ListDiscoverableDishes(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetDiscoverableDish 获取可发现候选菜品详情
// @Tags OrderFoodRecommendation
// @Summary 获取可发现候选菜品详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param dishId path string true "菜品ID"
// @Param data query orderfoodRequest.DiscoverableDishDetailQuery false "查询条件"
// @Success 200 {object} response.Response{data=orderfoodResponse.DiscoverableDishDetail,msg=string} "获取成功"
// @Router /orderfood/discoverable-dishes/{dishId} [get]
func (api *RecommendationApi) GetDiscoverableDish(c *gin.Context) {
	path, ok := bindDishIDPath(c)
	if !ok {
		return
	}
	var query orderfoodRequest.DiscoverableDishDetailQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionDiscoverableDishRead) {
		return
	}
	if query.IncludeModerationSummary &&
		!requirePermission(c, orderfoodService.PermissionModerationRead) {
		return
	}
	result, err := api.service.GetDiscoverableDish(
		c.Request.Context(), path.DishID, query.IncludeModerationSummary,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListRecommendations 分页查询推荐精选
// @Tags OrderFoodRecommendation
// @Summary 分页查询推荐精选
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.RecommendationListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.RecommendationSummary},msg=string} "获取成功"
// @Router /orderfood/recommendations [get]
func (api *RecommendationApi) ListRecommendations(c *gin.Context) {
	var query orderfoodRequest.RecommendationListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.RecommendationSummary]
	result, err := api.service.ListRecommendations(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetRecommendation 获取推荐精选详情
// @Tags OrderFoodRecommendation
// @Summary 获取推荐精选详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recommendationId path string true "推荐ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationDetail,msg=string} "获取成功"
// @Router /orderfood/recommendations/{recommendationId} [get]
func (api *RecommendationApi) GetRecommendation(c *gin.Context) {
	path, ok := bindRecommendationIDPath(c)
	if !ok || !requirePermission(c, orderfoodService.PermissionRecommendationRead) {
		return
	}
	result, err := api.service.GetRecommendation(c.Request.Context(), path.RecommendationID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateRecommendation 创建推荐草稿
// @Tags OrderFoodRecommendation
// @Summary 创建推荐草稿
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.RecommendationCreateInput true "推荐草稿参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations [post]
func (api *RecommendationApi) CreateRecommendation(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.RecommendationCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationCreate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateRecommendation(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateRecommendation 编辑推荐排序和展示说明
// @Tags OrderFoodRecommendation
// @Summary 编辑推荐排序和展示说明
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recommendationId path string true "推荐ID"
// @Param data body orderfoodRequest.RecommendationUpdateInput true "推荐编辑参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId} [put]
func (api *RecommendationApi) UpdateRecommendation(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.RecommendationUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// DeleteRecommendation 删除推荐草稿
// @Tags OrderFoodRecommendation
// @Summary 删除推荐草稿
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recommendationId path string true "推荐ID"
// @Param data body orderfoodRequest.RecommendationVersionInput true "推荐版本参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationDeleteResult,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId} [delete]
func (api *RecommendationApi) DeleteRecommendation(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.RecommendationVersionInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationDelete) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// PublishRecommendation 发布或重新发布推荐
// @Tags OrderFoodRecommendation
// @Summary 发布或重新发布推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recommendationId path string true "推荐ID"
// @Param data body orderfoodRequest.RecommendationVersionInput true "推荐版本参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId}/publish [post]
func (api *RecommendationApi) PublishRecommendation(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.RecommendationVersionInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationPublish) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.PublishRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// OfflineRecommendation 下线推荐
// @Tags OrderFoodRecommendation
// @Summary 下线推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recommendationId path string true "推荐ID"
// @Param data body orderfoodRequest.RecommendationOfflineInput true "推荐下线参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId}/offline [post]
func (api *RecommendationApi) OfflineRecommendation(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.RecommendationOfflineInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationOffline) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.OfflineRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateRecommendationSortOrder 批量调整推荐排序
// @Tags OrderFoodRecommendation
// @Summary 批量调整推荐排序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.RecommendationSortOrderInput true "推荐排序参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.RecommendationSortOrderResult,msg=string} "操作成功"
// @Router /orderfood/recommendations/sort-order [put]
func (api *RecommendationApi) UpdateRecommendationSortOrder(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.RecommendationSortOrderInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionRecommendationSort) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateRecommendationSortOrder(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindRecommendationIDPath 绑定并规范化推荐ID路径参数。
func bindRecommendationIDPath(c *gin.Context) (orderfoodRequest.RecommendationIDPath, bool) {
	var path orderfoodRequest.RecommendationIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.RecommendationIDPath{}, false
	}
	path.RecommendationID = strings.TrimSpace(path.RecommendationID)
	if path.RecommendationID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.RecommendationIDPath{}, false
	}
	return path, true
}

package dish

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	dishResponse "github.com/dyjh/order-food-mini-app/server/model/dish/response"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/dish"
	"github.com/gin-gonic/gin"
)

// RecommendationApi 提供可发现候选菜品和推荐精选接口处理能力。
type RecommendationApi struct{}

// ListDiscoverableDishes 分页查询可发现候选菜品
// @Tags OrderFoodRecommendation
// @Summary 分页查询可发现候选菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query dishRequest.DiscoverableDishListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]dishResponse.DiscoverableDishSummary},msg=string} "获取成功"
// @Router /orderfood/discoverable-dishes [get]
func (api *RecommendationApi) ListDiscoverableDishes(c *gin.Context) {
	var query dishRequest.DiscoverableDishListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionDiscoverableDishRead) {
		return
	}
	var result response.Page[dishResponse.DiscoverableDishSummary]
	result, err := recommendationService.ListDiscoverableDishes(c.Request.Context(), query)
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
// @Param data query dishRequest.DiscoverableDishDetailQuery false "查询条件"
// @Success 200 {object} response.Response{data=dishResponse.DiscoverableDishDetail,msg=string} "获取成功"
// @Router /orderfood/discoverable-dishes/{dishId} [get]
func (api *RecommendationApi) GetDiscoverableDish(c *gin.Context) {
	path, ok := apiCommon.BindDishIDPath(c)
	if !ok {
		return
	}
	var query dishRequest.DiscoverableDishDetailQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionDiscoverableDishRead) {
		return
	}
	if query.IncludeModerationSummary &&
		!apiCommon.RequirePermission(c, serviceCommon.PermissionModerationRead) {
		return
	}
	result, err := recommendationService.GetDiscoverableDish(
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
// @Param data query dishRequest.RecommendationListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]dishResponse.RecommendationSummary},msg=string} "获取成功"
// @Router /orderfood/recommendations [get]
func (api *RecommendationApi) ListRecommendations(c *gin.Context) {
	var query dishRequest.RecommendationListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationRead) {
		return
	}
	var result response.Page[dishResponse.RecommendationSummary]
	result, err := recommendationService.ListRecommendations(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=dishResponse.RecommendationDetail,msg=string} "获取成功"
// @Router /orderfood/recommendations/{recommendationId} [get]
func (api *RecommendationApi) GetRecommendation(c *gin.Context) {
	path, ok := bindRecommendationIDPath(c)
	if !ok || !apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationRead) {
		return
	}
	result, err := recommendationService.GetRecommendation(c.Request.Context(), path.RecommendationID)
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
// @Param data body dishRequest.RecommendationCreateInput true "推荐草稿参数"
// @Success 200 {object} response.Response{data=dishResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations [post]
func (api *RecommendationApi) CreateRecommendation(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body dishRequest.RecommendationCreateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationCreate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := recommendationService.CreateRecommendation(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
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
// @Param data body dishRequest.RecommendationUpdateInput true "推荐编辑参数"
// @Success 200 {object} response.Response{data=dishResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId} [put]
func (api *RecommendationApi) UpdateRecommendation(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body dishRequest.RecommendationUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := recommendationService.UpdateRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
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
// @Param data body dishRequest.RecommendationVersionInput true "推荐版本参数"
// @Success 200 {object} response.Response{data=dishResponse.RecommendationDeleteResult,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId} [delete]
func (api *RecommendationApi) DeleteRecommendation(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body dishRequest.RecommendationVersionInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationDelete) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := recommendationService.DeleteRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
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
// @Param data body dishRequest.RecommendationVersionInput true "推荐版本参数"
// @Success 200 {object} response.Response{data=dishResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId}/publish [post]
func (api *RecommendationApi) PublishRecommendation(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body dishRequest.RecommendationVersionInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationPublish) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := recommendationService.PublishRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
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
// @Param data body dishRequest.RecommendationOfflineInput true "推荐下线参数"
// @Success 200 {object} response.Response{data=dishResponse.RecommendationDetail,msg=string} "操作成功"
// @Router /orderfood/recommendations/{recommendationId}/offline [post]
func (api *RecommendationApi) OfflineRecommendation(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindRecommendationIDPath(c)
	if !ok {
		return
	}
	var body dishRequest.RecommendationOfflineInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationOffline) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := recommendationService.OfflineRecommendation(
		c.Request.Context(), path.RecommendationID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateRecommendationSortOrder 批量调整推荐排序
// @Tags OrderFoodRecommendation
// @Summary 批量调整推荐排序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body dishRequest.RecommendationSortOrderInput true "推荐排序参数"
// @Success 200 {object} response.Response{data=dishResponse.RecommendationSortOrderResult,msg=string} "操作成功"
// @Router /orderfood/recommendations/sort-order [put]
func (api *RecommendationApi) UpdateRecommendationSortOrder(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body dishRequest.RecommendationSortOrderInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionRecommendationSort) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := recommendationService.UpdateRecommendationSortOrder(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindRecommendationIDPath 绑定并规范化推荐ID路径参数。
func bindRecommendationIDPath(c *gin.Context) (dishRequest.RecommendationIDPath, bool) {
	var path dishRequest.RecommendationIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return dishRequest.RecommendationIDPath{}, false
	}
	path.RecommendationID = strings.TrimSpace(path.RecommendationID)
	if path.RecommendationID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return dishRequest.RecommendationIDPath{}, false
	}
	return path, true
}

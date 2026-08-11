package engagement

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	engagementRequest "github.com/dyjh/order-food-mini-app/server/model/engagement/request"
	engagementResponse "github.com/dyjh/order-food-mini-app/server/model/engagement/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/engagement"
	"github.com/gin-gonic/gin"
)

// PointsApi 提供积分规则接口处理能力。
type PointsApi struct{}

// ListPointEntries 分页查询积分流水
// @Tags OrderFoodPoints
// @Summary 分页查询积分流水
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query engagementRequest.PointEntryListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]engagementResponse.PointEntry},msg=string} "获取成功"
// @Router /orderfood/point-entries [get]
func (api *PointsApi) ListPointEntries(c *gin.Context) {
	var query engagementRequest.PointEntryListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionPointsRead) {
		return
	}
	result, err := pointsService.ListPointEntries(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// PreviewPointAdjustment 预览人工积分调整
// @Tags OrderFoodPoints
// @Summary 预览人工积分调整
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body engagementRequest.PointAdjustmentPreviewInput true "积分调整参数"
// @Success 200 {object} response.Response{data=engagementResponse.PointAdjustmentPreview,msg=string} "获取成功"
// @Router /orderfood/point-adjustments/preview [post]
func (api *PointsApi) PreviewPointAdjustment(c *gin.Context) {
	var body engagementRequest.PointAdjustmentPreviewInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionPointsAdjust) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, err := pointsService.PreviewPointAdjustment(
		c.Request.Context(), actor, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreatePointAdjustment 提交人工积分调整
// @Tags OrderFoodPoints
// @Summary 提交人工积分调整
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body engagementRequest.PointAdjustmentCreateInput true "积分调整确认参数"
// @Success 200 {object} response.Response{data=engagementResponse.PointAdjustmentResult,msg=string} "操作成功"
// @Router /orderfood/point-adjustments [post]
func (api *PointsApi) CreatePointAdjustment(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body engagementRequest.PointAdjustmentCreateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionPointsAdjust) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := pointsService.CreatePointAdjustment(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// GetPointRule 获取当前积分规则
// @Tags OrderFoodPointRule
// @Summary 获取当前积分规则
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=engagementResponse.PointRuleConfig,msg=string} "获取成功"
// @Router /orderfood/point-rules [get]
func (api *PointsApi) GetPointRule(c *gin.Context) {
	var query engagementRequest.PointRuleEmptyQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionPointRuleRead) {
		return
	}
	var result engagementResponse.PointRuleConfig
	result, err := pointsService.CurrentPointRule(c.Request.Context())
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdatePointRule 保存并立即应用积分规则
// @Tags OrderFoodPointRule
// @Summary 保存并立即应用积分规则
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body engagementRequest.PointRuleUpdateInput true "积分规则"
// @Success 200 {object} response.Response{data=engagementResponse.PointRuleConfig,msg=string} "操作成功"
// @Router /orderfood/point-rules [put]
func (api *PointsApi) UpdatePointRule(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body engagementRequest.PointRuleUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionPointRuleUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := pointsService.UpdatePointRule(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

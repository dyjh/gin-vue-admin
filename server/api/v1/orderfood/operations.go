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

// DashboardApi 提供管理端运营概览接口处理能力。
type DashboardApi struct {
	service *orderfoodService.DashboardService // 运营概览业务服务
}

// NewDashboardApi 创建运营概览API实例。
func NewDashboardApi(service *orderfoodService.DashboardService) *DashboardApi {
	return &DashboardApi{service: service}
}

// GetDashboard 获取运营概览
// @Tags OrderFoodDashboard
// @Summary 获取运营概览
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.DashboardQuery false "统计范围"
// @Success 200 {object} response.Response{data=orderfoodResponse.DashboardData,msg=string} "获取成功"
// @Router /orderfood/dashboard [get]
func (api *DashboardApi) GetDashboard(c *gin.Context) {
	var query orderfoodRequest.DashboardQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.DashboardData
	result, err := api.service.GetDashboard(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// available 检查运营概览API是否已注入业务服务。
func (api *DashboardApi) available(c *gin.Context) bool {
	if api != nil && api.service != nil {
		return true
	}
	response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
	return false
}

// AIUsageApi 提供管理端AI调用记录接口处理能力。
type AIUsageApi struct {
	service *orderfoodService.AIUsageService // AI调用记录业务服务
}

// NewAIUsageApi 创建AI调用记录API实例。
func NewAIUsageApi(service *orderfoodService.AIUsageService) *AIUsageApi {
	return &AIUsageApi{service: service}
}

// ListAIUsages 分页查询AI调用记录
// @Tags OrderFoodAIUsage
// @Summary 分页查询AI调用记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.AIUsageListQuery false "分页和筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.AIUsageSummary},msg=string} "获取成功"
// @Router /orderfood/ai-usages [get]
func (api *AIUsageApi) ListAIUsages(c *gin.Context) {
	var query orderfoodRequest.AIUsageListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]
	result, err := api.service.List(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetAIUsage 获取AI调用记录详情
// @Tags OrderFoodAIUsage
// @Summary 获取AI调用记录详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param usageId path string true "AI调用记录ID"
// @Param data query orderfoodRequest.AIUsageDetailQuery false "详情读取选项"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIUsageDetail,msg=string} "获取成功"
// @Router /orderfood/ai-usages/{usageId} [get]
func (api *AIUsageApi) GetAIUsage(c *gin.Context) {
	path, ok := bindAIUsageIDPath(c)
	if !ok {
		return
	}
	var query orderfoodRequest.AIUsageDetailQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.AIUsageDetail
	result, err := api.service.Detail(
		c.Request.Context(), path.UsageID, query.IncludeSensitiveContent, actor,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// DeleteAIUsageSensitiveContent 清除AI调用敏感内容
// @Tags OrderFoodAIUsage
// @Summary 清除AI调用敏感内容
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param usageId path string true "AI调用记录ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.AIUsageSensitiveDeleteInput true "清除原因和预期版本"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIUsageSensitiveDeleteResult,msg=string} "操作成功"
// @Router /orderfood/ai-usages/{usageId}/sensitive-content [delete]
func (api *AIUsageApi) DeleteAIUsageSensitiveContent(c *gin.Context) {
	path, ok := bindAIUsageIDPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIUsageSensitiveDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.AIUsageSensitiveDeleteResult
	result, replayed, err := api.service.ClearSensitiveContent(
		c.Request.Context(), path.UsageID, body, actor, header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// available 检查AI调用记录API是否已注入业务服务。
func (api *AIUsageApi) available(c *gin.Context) bool {
	if api != nil && api.service != nil {
		return true
	}
	response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
	return false
}

// bindAIUsageIDPath 绑定并规范化AI调用记录ID路径参数。
func bindAIUsageIDPath(c *gin.Context) (orderfoodRequest.AIUsageIDPath, bool) {
	var path orderfoodRequest.AIUsageIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.AIUsageIDPath{}, false
	}
	path.UsageID = strings.TrimSpace(path.UsageID)
	if path.UsageID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.AIUsageIDPath{}, false
	}
	return path, true
}

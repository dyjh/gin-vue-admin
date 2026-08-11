package ai

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	aiRequest "github.com/dyjh/order-food-mini-app/server/model/ai/request"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// AIUsageApi 提供管理端AI调用记录接口处理能力。
type AIUsageApi struct{}

// ListAIUsages 分页查询AI调用记录
// @Tags OrderFoodAIUsage
// @Summary 分页查询AI调用记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query aiRequest.AIUsageListQuery false "分页和筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]aiResponse.AIUsageSummary},msg=string} "获取成功"
// @Router /orderfood/ai-usages [get]
func (api *AIUsageApi) ListAIUsages(c *gin.Context) {
	var query aiRequest.AIUsageListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result response.Page[aiResponse.AIUsageSummary]
	result, err := aiUsageService.List(c.Request.Context(), query, actor)
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
// @Param data query aiRequest.AIUsageDetailQuery false "详情读取选项"
// @Success 200 {object} response.Response{data=aiResponse.AIUsageDetail,msg=string} "获取成功"
// @Router /orderfood/ai-usages/{usageId} [get]
func (api *AIUsageApi) GetAIUsage(c *gin.Context) {
	path, ok := bindAIUsageIDPath(c)
	if !ok {
		return
	}
	var query aiRequest.AIUsageDetailQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result aiResponse.AIUsageDetail
	result, err := aiUsageService.Detail(
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
// @Param data body aiRequest.AIUsageSensitiveDeleteInput true "清除原因和预期版本"
// @Success 200 {object} response.Response{data=aiResponse.AIUsageSensitiveDeleteResult,msg=string} "操作成功"
// @Router /orderfood/ai-usages/{usageId}/sensitive-content [delete]
func (api *AIUsageApi) DeleteAIUsageSensitiveContent(c *gin.Context) {
	path, ok := bindAIUsageIDPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIUsageSensitiveDeleteInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result aiResponse.AIUsageSensitiveDeleteResult
	result, replayed, err := aiUsageService.ClearSensitiveContent(
		c.Request.Context(), path.UsageID, body, actor, header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindAIUsageIDPath 绑定并规范化AI调用记录ID路径参数。
func bindAIUsageIDPath(c *gin.Context) (aiRequest.AIUsageIDPath, bool) {
	var path aiRequest.AIUsageIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return aiRequest.AIUsageIDPath{}, false
	}
	path.UsageID = strings.TrimSpace(path.UsageID)
	if path.UsageID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return aiRequest.AIUsageIDPath{}, false
	}
	return path, true
}

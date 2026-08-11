package ai

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	aiRequest "github.com/dyjh/order-food-mini-app/server/model/ai/request"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/ai"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// AIApi 提供AI 配置管理接口处理能力。
type AIApi struct{}

func (api *AIApi) require(c *gin.Context, permission string) bool {
	if err := aiService.Permission.Require(
		c.Request.Context(),
		utils.GetUserAuthorityId(c),
		permission,
	); err != nil {
		response.FailWithBusinessError(err, c)
		return false
	}
	return true
}

func (api *AIApi) has(c *gin.Context, permission string) (bool, bool) {
	allowed, err := aiService.Permission.HasPermission(
		c.Request.Context(),
		utils.GetUserAuthorityId(c),
		permission,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return false, false
	}
	return allowed, true
}

func aiAdminActor(c *gin.Context) (aiRequest.AIAdminActor, bool) {
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return aiRequest.AIAdminActor{}, false
	}
	return aiRequest.AIAdminActor{
		AdministratorID:  actor.AdministratorID,
		AuthorityID:      actor.AuthorityID,
		Username:         actor.Username,
		Nickname:         actor.Nickname,
		RequestID:        actor.RequestID,
		SourceIPMasked:   actor.SourceIPMasked,
		UserAgentSummary: actor.UserAgentSummary,
	}, true
}

func bindAIProviderPath(c *gin.Context) (aiRequest.AIProviderPath, bool) {
	var path aiRequest.AIProviderPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return path, false
	}
	path.ProviderID = strings.TrimSpace(path.ProviderID)
	return path, true
}

func bindAIModelPath(c *gin.Context) (aiRequest.AIModelPath, bool) {
	var path aiRequest.AIModelPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return path, false
	}
	path.ModelID = strings.TrimSpace(path.ModelID)
	return path, true
}

func bindAICapabilityPath(c *gin.Context) (aiRequest.AICapabilityPath, bool) {
	var path aiRequest.AICapabilityPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return path, false
	}
	path.CapabilityCode = strings.TrimSpace(path.CapabilityCode)
	return path, true
}

// GetPlatformPolicy 获取当前平台整体能力策略
// @Tags OrderFoodAIPlatformPolicy
// @Summary 获取当前平台整体能力策略
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=aiResponse.PlatformPolicyWorkspace,msg=string} "获取成功"
// @Router /orderfood/platform-capability-policy [get]
func (api *AIApi) GetPlatformPolicy(c *gin.Context) {
	var query aiRequest.AIEmptyQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionPlatformPolicyRead) {
		return
	}
	result, err := aiService.PlatformPolicyWorkspace(c.Request.Context())
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdatePlatformPolicy 保存并立即应用平台整体能力策略
// @Tags OrderFoodAIPlatformPolicy
// @Summary 保存并立即应用平台整体能力策略
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.PlatformPolicyUpdateInput true "平台策略"
// @Success 200 {object} response.Response{data=aiResponse.PlatformPolicyConfig,msg=string} "操作成功"
// @Router /orderfood/platform-capability-policy [put]
func (api *AIApi) UpdatePlatformPolicy(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.PlatformPolicyUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionPlatformPolicyUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdatePlatformPolicy(
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

// SetPlatformEmergencyStatus 立即启用或解除平台紧急停用
// @Tags OrderFoodAIPlatformPolicy
// @Summary 立即启用或解除平台紧急停用
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.PlatformPolicyEmergencyInput true "紧急状态"
// @Success 200 {object} response.Response{data=aiResponse.PlatformPolicyConfig,msg=string} "操作成功"
// @Router /orderfood/platform-capability-policy/emergency-status [put]
func (api *AIApi) SetPlatformEmergencyStatus(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.PlatformPolicyEmergencyInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionPlatformPolicyUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.SetPlatformEmergencyStatus(
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

// ListAIProviders 分页查询 AI 供应商
// @Tags OrderFoodAIProvider
// @Summary 分页查询 AI 供应商
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query aiRequest.AIProviderListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]aiResponse.AIProviderSummary},msg=string} "获取成功"
// @Router /orderfood/ai-providers [get]
func (api *AIApi) ListAIProviders(c *gin.Context) {
	var query aiRequest.AIProviderListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionAIProviderRead) {
		return
	}
	result, err := aiService.ListAIProviders(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetAIProvider 获取 AI 供应商详情
// @Tags OrderFoodAIProvider
// @Summary 获取 AI 供应商详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param providerId path string true "供应商 ID"
// @Success 200 {object} response.Response{data=aiResponse.AIProviderDetail,msg=string} "获取成功"
// @Router /orderfood/ai-providers/{providerId} [get]
func (api *AIApi) GetAIProvider(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAIProviderRead) {
		return
	}
	result, err := aiService.AIProviderDetail(c.Request.Context(), path.ProviderID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListAIProviderModels 查询供应商实时可选模型。
// @Tags OrderFoodAIProvider
// @Summary 查询供应商实时可选模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param providerId path string true "供应商 ID"
// @Success 200 {object} response.Response{data=[]aiResponse.AIProviderModelOption,msg=string} "获取成功"
// @Router /orderfood/ai-providers/{providerId}/models [get]
func (api *AIApi) ListAIProviderModels(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAIModelRead) {
		return
	}
	result, err := aiService.AIProviderModelOptions(c.Request.Context(), path.ProviderID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListAIModelProviders 查询模型页面供应商选项。
// @Tags OrderFoodAIModel
// @Summary 查询模型页面供应商选项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=[]aiResponse.AIModelProviderOption,msg=string} "获取成功"
// @Router /orderfood/ai-model-provider-options [get]
func (api *AIApi) ListAIModelProviders(c *gin.Context) {
	if !api.require(c, orderfoodService.PermissionAIModelRead) {
		return
	}
	result, err := aiService.AIModelProviderOptions(c.Request.Context())
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateAIProvider 创建 AI 供应商
// @Tags OrderFoodAIProvider
// @Summary 创建 AI 供应商
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIProviderCreateInput true "供应商配置"
// @Success 200 {object} response.Response{data=aiResponse.AIProviderDetail,msg=string} "操作成功"
// @Router /orderfood/ai-providers [post]
func (api *AIApi) CreateAIProvider(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIProviderCreateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderCreate) ||
		!api.require(c, orderfoodService.PermissionAIProviderCredentialWrite) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.CreateAIProvider(
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

// UpdateAIProvider 编辑 AI 供应商
// @Tags OrderFoodAIProvider
// @Summary 编辑 AI 供应商
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param providerId path string true "供应商 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIProviderUpdateInput true "供应商配置"
// @Success 200 {object} response.Response{data=aiResponse.AIProviderDetail,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId} [put]
func (api *AIApi) UpdateAIProvider(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIProviderUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderUpdate) {
		return
	}
	if body.APIKey != nil &&
		!api.require(c, orderfoodService.PermissionAIProviderCredentialWrite) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdateAIProvider(
		c.Request.Context(),
		actor,
		path.ProviderID,
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

// UpdateAIProviderStatus 启用或停用 AI 供应商
// @Tags OrderFoodAIProvider
// @Summary 启用或停用 AI 供应商
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param providerId path string true "供应商 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIStatusUpdateInput true "状态变更"
// @Success 200 {object} response.Response{data=aiResponse.AIProviderDetail,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId}/status [put]
func (api *AIApi) UpdateAIProviderStatus(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIStatusUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderStatus) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdateAIProviderStatus(
		c.Request.Context(),
		actor,
		path.ProviderID,
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

// DeleteAIProvider 删除未使用的 AI 供应商
// @Tags OrderFoodAIProvider
// @Summary 删除未使用的 AI 供应商
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param providerId path string true "供应商 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIDeleteInput true "删除原因和版本"
// @Success 200 {object} response.Response{data=aiResponse.DeletedResult,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId} [delete]
func (api *AIApi) DeleteAIProvider(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIDeleteInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderDelete) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.DeleteAIProvider(
		c.Request.Context(),
		actor,
		path.ProviderID,
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

// TestAIProviderConnection 测试 AI 供应商连接
// @Tags OrderFoodAIProvider
// @Summary 测试 AI 供应商连接
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param providerId path string true "供应商 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIProviderConnectionTestInput true "测试参数"
// @Success 200 {object} response.Response{data=aiResponse.ProviderConnectionTestResult,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId}/connection-tests [post]
func (api *AIApi) TestAIProviderConnection(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIProviderConnectionTestInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderTest) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.TestAIProviderConnection(
		c.Request.Context(),
		actor,
		path.ProviderID,
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

// ListAIModels 分页查询 AI 模型
// @Tags OrderFoodAIModel
// @Summary 分页查询 AI 模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query aiRequest.AIModelListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]aiResponse.AIModelSummary},msg=string} "获取成功"
// @Router /orderfood/ai-models [get]
func (api *AIApi) ListAIModels(c *gin.Context) {
	var query aiRequest.AIModelListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionAIModelRead) {
		return
	}
	result, err := aiService.ListAIModels(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetAIModel 获取 AI 模型详情
// @Tags OrderFoodAIModel
// @Summary 获取 AI 模型详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param modelId path string true "模型 ID"
// @Success 200 {object} response.Response{data=aiResponse.AIModelDetail,msg=string} "获取成功"
// @Router /orderfood/ai-models/{modelId} [get]
func (api *AIApi) GetAIModel(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAIModelRead) {
		return
	}
	result, err := aiService.AIModelDetail(c.Request.Context(), path.ModelID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateAIModel 创建 AI 模型
// @Tags OrderFoodAIModel
// @Summary 创建 AI 模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIModelCreateInput true "模型配置"
// @Success 200 {object} response.Response{data=aiResponse.AIModelDetail,msg=string} "操作成功"
// @Router /orderfood/ai-models [post]
func (api *AIApi) CreateAIModel(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIModelCreateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelCreate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.CreateAIModel(
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

// UpdateAIModel 编辑 AI 模型
// @Tags OrderFoodAIModel
// @Summary 编辑 AI 模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param modelId path string true "模型 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIModelUpdateInput true "模型配置"
// @Success 200 {object} response.Response{data=aiResponse.AIModelDetail,msg=string} "操作成功"
// @Router /orderfood/ai-models/{modelId} [put]
func (api *AIApi) UpdateAIModel(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIModelUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdateAIModel(
		c.Request.Context(),
		actor,
		path.ModelID,
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

// UpdateAIModelStatus 启用或停用 AI 模型
// @Tags OrderFoodAIModel
// @Summary 启用或停用 AI 模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param modelId path string true "模型 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIStatusUpdateInput true "状态变更"
// @Success 200 {object} response.Response{data=aiResponse.AIModelDetail,msg=string} "操作成功"
// @Router /orderfood/ai-models/{modelId}/status [put]
func (api *AIApi) UpdateAIModelStatus(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIStatusUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelStatus) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdateAIModelStatus(
		c.Request.Context(),
		actor,
		path.ModelID,
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

// DeleteAIModel 删除未使用的 AI 模型
// @Tags OrderFoodAIModel
// @Summary 删除未使用的 AI 模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param modelId path string true "模型 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIDeleteInput true "删除原因和版本"
// @Success 200 {object} response.Response{data=aiResponse.DeletedResult,msg=string} "操作成功"
// @Router /orderfood/ai-models/{modelId} [delete]
func (api *AIApi) DeleteAIModel(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIDeleteInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelDelete) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.DeleteAIModel(
		c.Request.Context(),
		actor,
		path.ModelID,
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

// ListAICapabilities 查询六项 AI 能力配置
// @Tags OrderFoodAICapability
// @Summary 查询六项 AI 能力配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query aiRequest.AICapabilityListQuery true "查询条件"
// @Success 200 {object} response.Response{data=aiResponse.AICapabilityListResult,msg=string} "获取成功"
// @Router /orderfood/ai-capabilities [get]
func (api *AIApi) ListAICapabilities(c *gin.Context) {
	var query aiRequest.AICapabilityListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionAICapabilityRead) {
		return
	}
	result, err := aiService.ListAICapabilities(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(aiResponse.AICapabilityListResult{List: result}, c)
}

// GetAICapability 获取 AI 能力配置工作区
// @Tags OrderFoodAICapability
// @Summary 获取 AI 能力配置工作区
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Success 200 {object} response.Response{data=aiResponse.AICapabilityWorkspace,msg=string} "获取成功"
// @Router /orderfood/ai-capabilities/{capabilityCode} [get]
func (api *AIApi) GetAICapability(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAICapabilityRead) {
		return
	}
	result, err := aiService.AICapabilityDetail(
		c.Request.Context(),
		aiRequest.AIAdminActor{},
		path.CapabilityCode,
		false,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdateAICapability 更新 AI 能力配置并立即生效
// @Tags OrderFoodAICapability
// @Summary 更新 AI 能力配置并立即生效
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AICapabilityUpdateInput true "能力配置"
// @Success 200 {object} response.Response{data=aiResponse.AICapabilityConfig,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode} [put]
func (api *AIApi) UpdateAICapability(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AICapabilityUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdateAICapability(
		c.Request.Context(),
		actor,
		path.CapabilityCode,
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

// GetAIPromptWorkspace 获取平台预设和当前生效提示词正文
// @Tags OrderFoodAICapabilityPrompt
// @Summary 获取平台预设和当前生效提示词正文
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Success 200 {object} response.Response{data=aiResponse.AIPromptWorkspace,msg=string} "获取成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompts [get]
func (api *AIApi) GetAIPromptWorkspace(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAICapabilityPromptRead) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, err := aiService.AIPromptWorkspace(
		c.Request.Context(),
		actor,
		path.CapabilityCode,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdateAIPrompt 更新 AI 提示词并立即生效
// @Tags OrderFoodAICapabilityPrompt
// @Summary 更新 AI 提示词并立即生效
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIPromptUpdateInput true "提示词配置"
// @Success 200 {object} response.Response{data=aiResponse.AIPromptConfig,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompts [put]
func (api *AIApi) UpdateAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIPromptUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.UpdateAIPrompt(
		c.Request.Context(), actor, path.CapabilityCode, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ValidateAIPrompt 校验提示词模板及变量白名单
// @Tags OrderFoodAICapabilityPrompt
// @Summary 校验提示词模板及变量白名单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIPromptValidationInput true "待校验提示词"
// @Success 200 {object} response.Response{data=aiResponse.AIPromptValidationResult,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompt-validation [post]
func (api *AIApi) ValidateAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIPromptValidationInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.ValidateAICapabilityPrompt(
		c.Request.Context(),
		actor,
		path.CapabilityCode,
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

// RenderAIPrompt 预览当前填写提示词的变量渲染结果
// @Tags OrderFoodAICapabilityPrompt
// @Summary 预览当前填写提示词的变量渲染结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIPromptRenderPreviewInput true "变量测试数据"
// @Success 200 {object} response.Response{data=aiResponse.AIPromptRenderPreviewResult,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompt-render-preview [post]
func (api *AIApi) RenderAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIPromptRenderPreviewInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptRead) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.RenderAICapabilityPrompt(
		c.Request.Context(),
		actor,
		path.CapabilityCode,
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

// TestAIPrompt 使用管理员测试数据运行当前填写的提示词
// @Tags OrderFoodAICapabilityPrompt
// @Summary 使用管理员测试数据运行当前填写的提示词
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body aiRequest.AIPromptTestInput true "提示词测试数据"
// @Success 200 {object} response.Response{data=aiResponse.AIPromptTestResult,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompt-tests [post]
func (api *AIApi) TestAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body aiRequest.AIPromptTestInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptTest) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := aiService.TestAICapabilityPrompt(
		c.Request.Context(),
		actor,
		path.CapabilityCode,
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

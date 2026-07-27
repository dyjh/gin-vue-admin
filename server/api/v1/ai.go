package v1

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// AIApi 提供AI 配置管理接口处理能力。
type AIApi struct {
	service *orderfoodService.AIService
}

// NewAIApi 创建AIAPI实例。
func NewAIApi(service *orderfoodService.AIService) *AIApi {
	return &AIApi{service: service}
}

func (api *AIApi) available(c *gin.Context) bool {
	if api == nil || api.service == nil || api.service.Permission == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return false
	}
	return true
}

func (api *AIApi) require(c *gin.Context, permission string) bool {
	if !api.available(c) {
		return false
	}
	if err := api.service.Permission.Require(
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
	if !api.available(c) {
		return false, false
	}
	allowed, err := api.service.Permission.HasPermission(
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

func aiAdminActor(c *gin.Context) (orderfoodRequest.AIAdminActor, bool) {
	actor, ok := contentAdminActor(c)
	if !ok {
		return orderfoodRequest.AIAdminActor{}, false
	}
	return orderfoodRequest.AIAdminActor{
		AdministratorID:  actor.AdministratorID,
		AuthorityID:      actor.AuthorityID,
		Username:         actor.Username,
		Nickname:         actor.Nickname,
		RequestID:        actor.RequestID,
		SourceIPMasked:   actor.SourceIPMasked,
		UserAgentSummary: actor.UserAgentSummary,
	}, true
}

func setAIReplayHeader(c *gin.Context, replayed bool) {
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
}

func bindAIProviderPath(c *gin.Context) (orderfoodRequest.AIProviderPath, bool) {
	var path orderfoodRequest.AIProviderPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return path, false
	}
	path.ProviderID = strings.TrimSpace(path.ProviderID)
	return path, true
}

func bindAIModelPath(c *gin.Context) (orderfoodRequest.AIModelPath, bool) {
	var path orderfoodRequest.AIModelPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return path, false
	}
	path.ModelID = strings.TrimSpace(path.ModelID)
	return path, true
}

func bindAICapabilityPath(c *gin.Context) (orderfoodRequest.AICapabilityPath, bool) {
	var path orderfoodRequest.AICapabilityPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
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
// @Success 200 {object} response.Response{data=orderfoodResponse.PlatformPolicyWorkspace,msg=string} "获取成功"
// @Router /orderfood/platform-capability-policy [get]
func (api *AIApi) GetPlatformPolicy(c *gin.Context) {
	var query orderfoodRequest.AIEmptyQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionPlatformPolicyRead) {
		return
	}
	result, err := api.service.PlatformPolicyWorkspace(c.Request.Context())
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
// @Param data body orderfoodRequest.PlatformPolicyUpdateInput true "平台策略"
// @Success 200 {object} response.Response{data=orderfoodResponse.PlatformPolicyConfig,msg=string} "操作成功"
// @Router /orderfood/platform-capability-policy [put]
func (api *AIApi) UpdatePlatformPolicy(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.PlatformPolicyUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionPlatformPolicyUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdatePlatformPolicy(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// SetPlatformEmergencyStatus 立即启用或解除平台紧急停用
// @Tags OrderFoodAIPlatformPolicy
// @Summary 立即启用或解除平台紧急停用
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.PlatformPolicyEmergencyInput true "紧急状态"
// @Success 200 {object} response.Response{data=orderfoodResponse.PlatformPolicyConfig,msg=string} "操作成功"
// @Router /orderfood/platform-capability-policy/emergency-status [put]
func (api *AIApi) SetPlatformEmergencyStatus(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.PlatformPolicyEmergencyInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionPlatformPolicyUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.SetPlatformEmergencyStatus(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListAIProviders 分页查询 AI 供应商
// @Tags OrderFoodAIProvider
// @Summary 分页查询 AI 供应商
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.AIProviderListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.AIProviderSummary},msg=string} "获取成功"
// @Router /orderfood/ai-providers [get]
func (api *AIApi) ListAIProviders(c *gin.Context) {
	var query orderfoodRequest.AIProviderListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionAIProviderRead) {
		return
	}
	result, err := api.service.ListAIProviders(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.AIProviderDetail,msg=string} "获取成功"
// @Router /orderfood/ai-providers/{providerId} [get]
func (api *AIApi) GetAIProvider(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAIProviderRead) {
		return
	}
	result, err := api.service.AIProviderDetail(c.Request.Context(), path.ProviderID)
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
// @Param data body orderfoodRequest.AIProviderCreateInput true "供应商配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIProviderDetail,msg=string} "操作成功"
// @Router /orderfood/ai-providers [post]
func (api *AIApi) CreateAIProvider(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIProviderCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderCreate) ||
		!api.require(c, orderfoodService.PermissionAIProviderCredentialWrite) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateAIProvider(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIProviderUpdateInput true "供应商配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIProviderDetail,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId} [put]
func (api *AIApi) UpdateAIProvider(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIProviderUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
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
	result, replayed, err := api.service.UpdateAIProvider(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIStatusUpdateInput true "状态变更"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIProviderDetail,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId}/status [put]
func (api *AIApi) UpdateAIProviderStatus(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIStatusUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderStatus) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateAIProviderStatus(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIDeleteInput true "删除原因和版本"
// @Success 200 {object} response.Response{data=orderfoodResponse.DeletedResult,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId} [delete]
func (api *AIApi) DeleteAIProvider(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderDelete) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteAIProvider(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIProviderConnectionTestInput true "测试参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.ProviderConnectionTestResult,msg=string} "操作成功"
// @Router /orderfood/ai-providers/{providerId}/connection-tests [post]
func (api *AIApi) TestAIProviderConnection(c *gin.Context) {
	path, ok := bindAIProviderPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIProviderConnectionTestInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIProviderTest) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.TestAIProviderConnection(
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
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListAIModels 分页查询 AI 模型
// @Tags OrderFoodAIModel
// @Summary 分页查询 AI 模型
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.AIModelListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.AIModelSummary},msg=string} "获取成功"
// @Router /orderfood/ai-models [get]
func (api *AIApi) ListAIModels(c *gin.Context) {
	var query orderfoodRequest.AIModelListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionAIModelRead) {
		return
	}
	result, err := api.service.ListAIModels(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.AIModelDetail,msg=string} "获取成功"
// @Router /orderfood/ai-models/{modelId} [get]
func (api *AIApi) GetAIModel(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAIModelRead) {
		return
	}
	result, err := api.service.AIModelDetail(c.Request.Context(), path.ModelID)
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
// @Param data body orderfoodRequest.AIModelCreateInput true "模型配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIModelDetail,msg=string} "操作成功"
// @Router /orderfood/ai-models [post]
func (api *AIApi) CreateAIModel(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIModelCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelCreate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateAIModel(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIModelUpdateInput true "模型配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIModelDetail,msg=string} "操作成功"
// @Router /orderfood/ai-models/{modelId} [put]
func (api *AIApi) UpdateAIModel(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIModelUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateAIModel(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIStatusUpdateInput true "状态变更"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIModelDetail,msg=string} "操作成功"
// @Router /orderfood/ai-models/{modelId}/status [put]
func (api *AIApi) UpdateAIModelStatus(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIStatusUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelStatus) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateAIModelStatus(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIDeleteInput true "删除原因和版本"
// @Success 200 {object} response.Response{data=orderfoodResponse.DeletedResult,msg=string} "操作成功"
// @Router /orderfood/ai-models/{modelId} [delete]
func (api *AIApi) DeleteAIModel(c *gin.Context) {
	path, ok := bindAIModelPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAIModelDelete) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteAIModel(
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
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListAICapabilities 查询六项 AI 能力配置
// @Tags OrderFoodAICapability
// @Summary 查询六项 AI 能力配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.AICapabilityListQuery true "查询条件"
// @Success 200 {object} response.Response{data=orderfoodResponse.AICapabilityListResult,msg=string} "获取成功"
// @Router /orderfood/ai-capabilities [get]
func (api *AIApi) ListAICapabilities(c *gin.Context) {
	var query orderfoodRequest.AICapabilityListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!api.require(c, orderfoodService.PermissionAICapabilityRead) {
		return
	}
	result, err := api.service.ListAICapabilities(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(orderfoodResponse.AICapabilityListResult{List: result}, c)
}

// GetAICapability 获取 AI 能力配置工作区
// @Tags OrderFoodAICapability
// @Summary 获取 AI 能力配置工作区
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Success 200 {object} response.Response{data=orderfoodResponse.AICapabilityWorkspace,msg=string} "获取成功"
// @Router /orderfood/ai-capabilities/{capabilityCode} [get]
func (api *AIApi) GetAICapability(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok || !api.require(c, orderfoodService.PermissionAICapabilityRead) {
		return
	}
	result, err := api.service.AICapabilityDetail(
		c.Request.Context(),
		orderfoodRequest.AIAdminActor{},
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
// @Param data body orderfoodRequest.AICapabilityUpdateInput true "能力配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.AICapabilityConfig,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode} [put]
func (api *AIApi) UpdateAICapability(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AICapabilityUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateAICapability(
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
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// GetAIPromptWorkspace 获取平台预设和当前生效提示词正文
// @Tags OrderFoodAICapabilityPrompt
// @Summary 获取平台预设和当前生效提示词正文
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param capabilityCode path string true "能力编码"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIPromptWorkspace,msg=string} "获取成功"
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
	result, err := api.service.AIPromptWorkspace(
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
// @Param data body orderfoodRequest.AIPromptUpdateInput true "提示词配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIPromptConfig,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompts [put]
func (api *AIApi) UpdateAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIPromptUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateAIPrompt(
		c.Request.Context(), actor, path.CapabilityCode, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIPromptValidationInput true "待校验提示词"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIPromptValidationResult,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompt-validation [post]
func (api *AIApi) ValidateAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIPromptValidationInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptUpdate) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.ValidateAICapabilityPrompt(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIPromptRenderPreviewInput true "变量测试数据"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIPromptRenderPreviewResult,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompt-render-preview [post]
func (api *AIApi) RenderAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIPromptRenderPreviewInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptRead) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.RenderAICapabilityPrompt(
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
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.AIPromptTestInput true "提示词测试数据"
// @Success 200 {object} response.Response{data=orderfoodResponse.AIPromptTestResult,msg=string} "操作成功"
// @Router /orderfood/ai-capabilities/{capabilityCode}/prompt-tests [post]
func (api *AIApi) TestAIPrompt(c *gin.Context) {
	path, ok := bindAICapabilityPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.AIPromptTestInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!api.require(c, orderfoodService.PermissionAICapabilityPromptTest) {
		return
	}
	actor, ok := aiAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.TestAICapabilityPrompt(
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
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

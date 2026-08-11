package ai

import (
	"context"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	aiRequest "github.com/dyjh/order-food-mini-app/server/model/ai/request"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	"gorm.io/gorm"
)

const aiCapabilityMutationTimeout = 5 * time.Second

func capabilityDefinition(
	db *gorm.DB,
	capabilityCode string,
) (aiModel.AICapabilityDefinition, error) {
	var definition aiModel.AICapabilityDefinition
	if err := db.First(&definition, "code = ?", capabilityCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return definition, appErrors.AdminNotFound.DefaultMsg()
		}
		return definition, appErrors.AdminInternal.Wrap(err, "get AI capability definition")
	}
	return definition, nil
}

func capabilityCurrent(
	db *gorm.DB,
	definition aiModel.AICapabilityDefinition,
) (*aiModel.AICapabilityConfig, error) {
	var config aiModel.AICapabilityConfig
	if err := db.First(&config, "capability_code = ?", definition.Code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "get current AI capability")
	}
	return &config, nil
}

func capabilityConfigWriteError(err error, operation string) error {
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(lower, "duplicate entry") ||
		strings.Contains(lower, "lock wait timeout") ||
		strings.Contains(lower, "deadlock") {
		return appErrors.AdminStateConflict.New("能力配置正在被其他操作修改，请稍后重试")
	}
	return appErrors.AdminInternal.Wrap(err, operation)
}

func capabilityConfigUpdateValues(
	next aiModel.AICapabilityConfig,
) map[string]interface{} {
	return map[string]interface{}{
		"version":                        next.Version,
		"primary_model_id":               next.PrimaryModelID,
		"auxiliary_model_id":             next.AuxiliaryModelID,
		"point_cost":                     next.PointCost,
		"daily_limit_per_user":           next.DailyLimitPerUser,
		"timeout_ms":                     next.TimeoutMS,
		"free_quota_per_day":             next.FreeQuotaPerDay,
		"prompt_mode":                    next.PromptMode,
		"prompt_preset_version":          next.PromptPresetVersion,
		"system_prompt":                  next.SystemPrompt,
		"user_prompt_template":           next.UserPromptTemplate,
		"prompt_allowed_variables_json":  next.PromptAllowedVariablesJSON,
		"prompt_required_variables_json": next.PromptRequiredVariablesJSON,
		"prompt_output_schema_version":   next.PromptOutputSchemaVersion,
		"prompt_hash":                    next.PromptHash,
		"applied_by_id":                  next.AppliedByID,
		"applied_by_username":            next.AppliedByUsername,
		"applied_by_nickname":            next.AppliedByNickname,
		"applied_at":                     next.AppliedAt,
		"reason":                         next.Reason,
	}
}

func persistCapabilityConfig(
	tx *gorm.DB,
	current *aiModel.AICapabilityConfig,
	next aiModel.AICapabilityConfig,
	operation string,
) error {
	if current == nil {
		return capabilityConfigWriteError(tx.Create(&next).Error, operation)
	}
	update := tx.Model(&aiModel.AICapabilityConfig{}).
		Where(
			"capability_code = ? AND version = ?",
			current.CapabilityCode,
			current.Version,
		).
		Updates(capabilityConfigUpdateValues(next))
	if update.Error != nil {
		return capabilityConfigWriteError(update.Error, operation)
	}
	if update.RowsAffected != 1 {
		return appErrors.AdminStateConflict.DefaultMsg()
	}
	return nil
}

// validateAICapabilityUpdateInput 校验服务层直接调用也必须遵守的能力配置边界。
func validateAICapabilityUpdateInput(
	input aiRequest.AICapabilityUpdateInput,
) error {
	reason := strings.TrimSpace(input.Reason)
	primaryModelID := strings.TrimSpace(input.PrimaryModelID)
	auxiliaryModelID := ""
	if input.AuxiliaryModelID != nil {
		auxiliaryModelID = strings.TrimSpace(*input.AuxiliaryModelID)
	}
	if primaryModelID == "" || len([]rune(primaryModelID)) > 64 || len([]rune(auxiliaryModelID)) > 64 ||
		input.PointCost < 0 || input.PointCost > 100000 || input.DailyLimitPerUser < 0 || input.DailyLimitPerUser > 100000 ||
		input.FreeQuotaPerDay < 0 || input.FreeQuotaPerDay > 100000 || input.TimeoutMS < 1000 || input.TimeoutMS > 120000 ||
		input.ExpectedVersion < 0 || len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// fixedCapabilityValidationRules 返回与实际运行时代码一致的只读校验说明。
func fixedCapabilityValidationRules(capabilityCode string) []string {
	rules := map[string][]string{
		aiModel.AICapabilityDishTextExtract: {
			"返回内容必须能解析为服务端固定的菜品草稿结构",
			"结果只回填可编辑草稿，不直接创建、公开或推荐菜品",
		},
		aiModel.AICapabilityRecipeImageExtract: {
			"返回内容必须能解析为服务端固定的菜品草稿结构",
			"只处理已通过上传审核的图片，结果只回填可编辑草稿",
		},
		aiModel.AICapabilityDishCoverCreate: {
			"只接受合法 Base64 图片或不含用户凭据的 HTTPS 图片地址",
			"图片必须通过非空、文件大小和文件类型校验后才能保存",
		},
		aiModel.AICapabilityCheckinImageAnalyze: {
			"返回内容必须能解析为服务端固定的偏好事实结构",
			"置信度、词条清洗、去重和数量上限由服务端固定执行",
		},
		aiModel.AICapabilityPreferenceSummarize: {
			"输入仅包含已清洗的结构化证据，不向模型发送打卡原图",
			"用户明确设置与系统归纳内容必须分别保留来源标识",
		},
		aiModel.AICapabilityMealSuggest: {
			"固定校验菜品数量、字段边界、分类、标签、单位、配料与步骤一致性",
			"固定拦截重复菜名和违反用户硬条件的结果",
			"标准菜品索引校验及失败重试只由“标准菜品索引”页面的当前开关控制",
		},
		aiModel.AICapabilityPrepSequence: {
			"结果必须完整覆盖原饭局步骤，不允许新增、遗漏或重复步骤",
			"并行步骤必须引用有效步骤，且不能复现用户明确排除的旧方案",
			"预计时长必须处于服务端固定范围",
		},
	}
	result := append([]string(nil), rules[capabilityCode]...)
	if capabilityCode == aiModel.AICapabilityCheckinImageAnalyze {
		result = append(result, rules[aiModel.AICapabilityPreferenceSummarize]...)
	}
	return result
}

func capabilityConfigResponse(
	version aiModel.AICapabilityConfig,
	latestPresetVersion int64,
	includePromptBodies bool,
) (aiResponse.AICapabilityConfig, error) {
	prompt, err := promptMetadata(
		version.PromptMode,
		version.PromptPresetVersion,
		version.SystemPrompt,
		version.UserPromptTemplate,
		version.PromptAllowedVariablesJSON,
		version.PromptRequiredVariablesJSON,
		version.PromptOutputSchemaVersion,
		version.PromptHash,
		latestPresetVersion,
		includePromptBodies,
	)
	if err != nil {
		return aiResponse.AICapabilityConfig{}, err
	}
	return aiResponse.AICapabilityConfig{
		CapabilityCode:       version.CapabilityCode,
		PrimaryModelID:       version.PrimaryModelID,
		AuxiliaryModelID:     version.AuxiliaryModelID,
		PointCost:            version.PointCost,
		DailyLimitPerUser:    version.DailyLimitPerUser,
		TimeoutMS:            version.TimeoutMS,
		FreeQuotaPerDay:      version.FreeQuotaPerDay,
		FixedValidationRules: fixedCapabilityValidationRules(version.CapabilityCode),
		Prompt:               prompt,
		PromptMode:           externalPromptMode(version.PromptMode),
		PromptHash:           version.PromptHash,
		Version:              version.Version,
		UpdatedBy: serviceCommon.AdministratorSummary(
			version.AppliedByID,
			version.AppliedByUsername,
			version.AppliedByNickname,
		),
		UpdatedAt: version.AppliedAt,
		Reason:    version.Reason,
	}, nil
}

func modelName(db *gorm.DB, modelID string) (string, error) {
	if modelID == "" {
		return "", nil
	}
	var model aiModel.AIModel
	if err := db.Select("name").First(&model, "id = ?", modelID).Error; err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "get AI capability model name")
	}
	return model.Name, nil
}

func capabilityPointerString(value *aiModel.AIModelCapability) *string {
	if value == nil {
		return nil
	}
	result := string(*value)
	return &result
}

func auxiliaryModelReady(db *gorm.DB, config *aiModel.AICapabilityConfig, definition aiModel.AICapabilityDefinition) (bool, error) {
	if definition.RequiredAuxiliaryModelCapability == nil {
		return config == nil || config.AuxiliaryModelID == nil, nil
	}
	if config == nil || config.AuxiliaryModelID == nil || strings.TrimSpace(*config.AuxiliaryModelID) == "" {
		return false, nil
	}
	_, _, ready, err := availableModelWithCapability(db, strings.TrimSpace(*config.AuxiliaryModelID), *definition.RequiredAuxiliaryModelCapability)
	return ready, err
}

func capabilityConfigReady(
	db *gorm.DB,
	definition aiModel.AICapabilityDefinition,
) (bool, error) {
	current, err := capabilityCurrent(db, definition)
	if err != nil || current == nil {
		return false, err
	}
	_, _, primaryReady, err := availableCapabilityModel(db, current.PrimaryModelID, definition)
	if err != nil {
		return false, err
	}
	auxiliaryReady, err := auxiliaryModelReady(db, current, definition)
	if err != nil {
		return false, err
	}
	return primaryReady && auxiliaryReady && strings.TrimSpace(current.PromptHash) != "", nil
}

func capabilitySummary(
	db *gorm.DB,
	definition aiModel.AICapabilityDefinition,
) (aiResponse.AICapabilitySummary, error) {
	preset, err := promptPreset(db, definition.Code)
	if err != nil {
		return aiResponse.AICapabilitySummary{}, err
	}
	displayName := definition.Name
	if definition.Code == aiModel.AICapabilityCheckinImageAnalyze {
		displayName = "打卡智能分析"
	}
	summary := aiResponse.AICapabilitySummary{
		Code:                             definition.Code,
		Name:                             displayName,
		ClientFeatureCode:                definition.ClientFeatureCode,
		RequiredModelCapability:          string(definition.RequiredModelCapability),
		RequiredAuxiliaryModelCapability: capabilityPointerString(definition.RequiredAuxiliaryModelCapability),
		Version:                          0,
		PromptMode:                       "default",
		PromptPresetVersion:              preset.Version,
		PromptHash:                       preset.ContentHash,
		SortOrder:                        definition.SortOrder,
		UpdatedAt:                        definition.UpdatedAt,
	}
	current, err := capabilityCurrent(db, definition)
	if err != nil {
		return summary, err
	}
	if current == nil {
		return summary, nil
	}
	summary.Version = current.Version
	summary.Ready, err = capabilityConfigReady(db, definition)
	if err != nil {
		return summary, err
	}
	if definition.Code == aiModel.AICapabilityCheckinImageAnalyze {
		var preferenceDefinition aiModel.AICapabilityDefinition
		definitionErr := db.First(
			&preferenceDefinition,
			"code = ?",
			aiModel.AICapabilityPreferenceSummarize,
		).Error
		if errors.Is(definitionErr, gorm.ErrRecordNotFound) {
			summary.Ready = false
		} else if definitionErr != nil {
			return summary, appErrors.AdminInternal.Wrap(
				definitionErr,
				"get preference summarize definition",
			)
		} else {
			preferenceReady, readyErr := capabilityConfigReady(db, preferenceDefinition)
			if readyErr != nil {
				return summary, readyErr
			}
			summary.Ready = summary.Ready && preferenceReady
		}
	}
	summary.PointCost = current.PointCost
	summary.DailyLimitPerUser = current.DailyLimitPerUser
	summary.FreeQuotaPerDay = current.FreeQuotaPerDay
	summary.PromptMode = externalPromptMode(current.PromptMode)
	summary.PromptPresetVersion = current.PromptPresetVersion
	summary.PromptHash = current.PromptHash
	summary.UpdatedAt = current.AppliedAt
	summary.PrimaryModelName, err = modelName(db, current.PrimaryModelID)
	if err != nil {
		return summary, err
	}
	if current.AuxiliaryModelID != nil {
		summary.AuxiliaryModelName, err = modelName(db, *current.AuxiliaryModelID)
		if err != nil {
			return summary, err
		}
	}
	return summary, nil
}

// ListAICapabilities 查询 AI 能力列表。
func (service *AIService) ListAICapabilities(
	ctx context.Context,
	query aiRequest.AICapabilityListQuery,
) ([]aiResponse.AICapabilitySummary, error) {
	query.ApplyDefaults()
	sortColumns := map[string]string{
		"sortOrder": "sort_order",
		"code":      "code",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return nil, appErrors.AdminBadRequest.DefaultMsg()
	}
	order := strings.ToLower(query.SortOrder)
	if order != "asc" && order != "desc" {
		return nil, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database().WithContext(ctx)
	statement := db.Model(&aiModel.AICapabilityDefinition{})
	if query.ClientFeatureCode != "" {
		statement = statement.Where("client_feature_code = ?", query.ClientFeatureCode)
	}
	var definitions []aiModel.AICapabilityDefinition
	if err := statement.Order(column + " " + order).Find(&definitions).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list AI capabilities")
	}
	list := make([]aiResponse.AICapabilitySummary, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Code == aiModel.AICapabilityPreferenceSummarize {
			continue
		}
		summary, err := capabilitySummary(db, definition)
		if err != nil {
			return nil, err
		}
		list = append(list, summary)
	}
	return list, nil
}

// AICapabilityDetail 获取指定 AI 能力的当前配置。
func (service *AIService) AICapabilityDetail(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
	includePromptBodies bool,
) (aiResponse.AICapabilityWorkspace, error) {
	db := service.database().WithContext(ctx)
	definition, err := capabilityDefinition(db, capabilityCode)
	if err != nil {
		return aiResponse.AICapabilityWorkspace{}, err
	}
	if includePromptBodies {
		if err := validateAIActor(actor); err != nil {
			return aiResponse.AICapabilityWorkspace{}, err
		}
		if _, err := service.recordPromptAccess(
			ctx,
			actor,
			"read_ai_capability_prompt",
			capabilityCode,
		); err != nil {
			return aiResponse.AICapabilityWorkspace{}, err
		}
	}
	summary, err := capabilitySummary(db, definition)
	if err != nil {
		return aiResponse.AICapabilityWorkspace{}, err
	}
	result := aiResponse.AICapabilityWorkspace{Summary: summary}
	current, err := capabilityCurrent(db, definition)
	if err != nil {
		return result, err
	}
	if current == nil {
		return result, nil
	}
	latest, err := promptPreset(db, capabilityCode)
	if err != nil {
		return result, err
	}
	config, err := capabilityConfigResponse(*current, latest.Version, includePromptBodies)
	if err != nil {
		return result, err
	}
	result.Config = &config
	return result, nil
}

func availableModelWithCapability(db *gorm.DB, modelID string, required aiModel.AIModelCapability) (aiModel.AIModel, aiModel.AIProvider, bool, error) {
	var model aiModel.AIModel
	if err := db.First(&model, "id = ?", modelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model, aiModel.AIProvider{}, false, nil
		}
		return model, aiModel.AIProvider{}, false, appErrors.AdminInternal.Wrap(err, "get AI capability model")
	}
	var provider aiModel.AIProvider
	if err := db.First(&provider, "id = ?", model.ProviderID).Error; err != nil {
		return model, provider, false, appErrors.AdminInternal.Wrap(err, "get AI capability provider")
	}
	supports, err := modelSupports(model, required)
	if err != nil {
		return model, provider, false, err
	}
	available := model.Enabled && provider.Enabled && strings.TrimSpace(provider.APIKey) != "" && supports
	return model, provider, available, nil
}
func availableCapabilityModel(db *gorm.DB, modelID string, definition aiModel.AICapabilityDefinition) (aiModel.AIModel, aiModel.AIProvider, bool, error) {
	return availableModelWithCapability(db, modelID, definition.RequiredModelCapability)
}

func aiCapabilitiesReadiness(
	db *gorm.DB,
) (aiResponse.AICapabilityReadiness, error) {
	var definitions []aiModel.AICapabilityDefinition
	if err := db.Order("sort_order asc").Find(&definitions).Error; err != nil {
		return aiResponse.AICapabilityReadiness{},
			appErrors.AdminInternal.Wrap(err, "list AI capability readiness")
	}
	definitionByCode := make(map[string]aiModel.AICapabilityDefinition, len(definitions))
	for _, definition := range definitions {
		definitionByCode[definition.Code] = definition
	}
	expected := defaultCapabilities()
	result := aiResponse.AICapabilityReadiness{
		TotalCount:          len(expected) - 1,
		UnreadyCapabilities: make([]string, 0, len(expected)-1),
	}
	for _, item := range expected {
		if item.Code == aiModel.AICapabilityPreferenceSummarize {
			continue
		}
		definition, exists := definitionByCode[item.Code]
		if !exists {
			result.UnreadyCapabilities = append(result.UnreadyCapabilities, item.Code)
			continue
		}
		ready, err := capabilityConfigReady(db, definition)
		if err != nil {
			return result, err
		}
		if item.Code == aiModel.AICapabilityCheckinImageAnalyze {
			preferenceDefinition, preferenceExists := definitionByCode[aiModel.AICapabilityPreferenceSummarize]
			if !preferenceExists {
				ready = false
			} else {
				preferenceReady, readyErr := capabilityConfigReady(db, preferenceDefinition)
				if readyErr != nil {
					return result, readyErr
				}
				ready = ready && preferenceReady
			}
		}
		if !ready {
			result.UnreadyCapabilities = append(result.UnreadyCapabilities, item.Code)
			continue
		}
		result.ReadyCount++
	}
	result.AllReady = result.ReadyCount == result.TotalCount
	return result, nil
}
func promptFromCurrentCapability(
	service *AIService,
	tx *gorm.DB,
	capabilityCode string,
	current *aiModel.AICapabilityConfig,
) (resolvedPrompt, error) {
	if current == nil {
		resolved, validation, err := service.resolvePromptInput(
			tx,
			capabilityCode,
			aiRequest.AIPromptConfigInput{Mode: aiModel.AIPromptPreset},
		)
		if err != nil {
			return resolvedPrompt{}, err
		}
		if !validation.Valid {
			return resolvedPrompt{}, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		return resolved, nil
	}
	var allowed []string
	if err := decodeJSON(current.PromptAllowedVariablesJSON, &allowed); err != nil {
		return resolvedPrompt{}, err
	}
	var required []string
	if err := decodeJSON(current.PromptRequiredVariablesJSON, &required); err != nil {
		return resolvedPrompt{}, err
	}
	latest, err := promptPreset(tx, capabilityCode)
	if err != nil {
		return resolvedPrompt{}, err
	}
	return resolvedPrompt{
		Mode:                current.PromptMode,
		PresetVersion:       current.PromptPresetVersion,
		SystemPrompt:        current.SystemPrompt,
		UserPromptTemplate:  current.UserPromptTemplate,
		AllowedVariables:    allowed,
		RequiredVariables:   required,
		OutputSchemaVersion: current.PromptOutputSchemaVersion,
		ContentHash:         current.PromptHash,
		LatestPresetVersion: latest.Version,
	}, nil
}

func syncPreferenceSummarizeConfig(
	service *AIService,
	tx *gorm.DB,
	actor aiRequest.AIAdminActor,
	primaryModelID string,
	timeoutMS int,
	reason string,
	now time.Time,
) error {
	definition, err := capabilityDefinition(
		tx,
		aiModel.AICapabilityPreferenceSummarize,
	)
	if err != nil {
		return err
	}
	current, err := capabilityCurrent(tx, definition)
	if err != nil {
		return err
	}
	model, _, available, err := availableCapabilityModel(tx, primaryModelID, definition)
	if err != nil {
		return err
	}
	if !available {
		return appErrors.AdminInvalidConfig.DefaultMsg()
	}
	prompt, err := promptFromCurrentCapability(
		service,
		tx,
		aiModel.AICapabilityPreferenceSummarize,
		current,
	)
	if err != nil {
		return err
	}
	if err := validatePromptContext(model, prompt); err != nil {
		return err
	}
	allowedJSON, err := encodeJSON(prompt.AllowedVariables)
	if err != nil {
		return err
	}
	requiredJSON, err := encodeJSON(prompt.RequiredVariables)
	if err != nil {
		return err
	}
	currentVersion := int64(0)
	if current != nil {
		currentVersion = current.Version
	}
	next := aiModel.AICapabilityConfig{
		CapabilityCode:              aiModel.AICapabilityPreferenceSummarize,
		Version:                     currentVersion + 1,
		PrimaryModelID:              primaryModelID,
		PointCost:                   0,
		DailyLimitPerUser:           0,
		TimeoutMS:                   timeoutMS,
		FreeQuotaPerDay:             0,
		PromptMode:                  prompt.Mode,
		PromptPresetVersion:         prompt.PresetVersion,
		SystemPrompt:                prompt.SystemPrompt,
		UserPromptTemplate:          prompt.UserPromptTemplate,
		PromptAllowedVariablesJSON:  allowedJSON,
		PromptRequiredVariablesJSON: requiredJSON,
		PromptOutputSchemaVersion:   prompt.OutputSchemaVersion,
		PromptHash:                  prompt.ContentHash,
		AppliedByID:                 actor.AdministratorID,
		AppliedByUsername:           actor.Username,
		AppliedByNickname:           actor.Nickname,
		AppliedAt:                   now,
		Reason:                      reason,
	}
	if err := persistCapabilityConfig(
		tx,
		current,
		next,
		"sync preference summarize config",
	); err != nil {
		return err
	}
	return recordAIChange(
		tx,
		now,
		"ai_capability",
		aiModel.AICapabilityPreferenceSummarize,
		"sync",
		next.Version,
		actor,
		&reason,
	)
}

// UpdateAICapability 保存并立即应用 AI 能力配置。
func (service *AIService) UpdateAICapability(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input aiRequest.AICapabilityUpdateInput,
) (aiResponse.AICapabilityConfig, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.AICapabilityConfig{}, false, err
	}
	if service == nil || service.Permission == nil || service.Idempotency == nil {
		return aiResponse.AICapabilityConfig{}, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(
		ctx, actor.AuthorityID, PermissionAICapabilityUpdate,
	); err != nil {
		return aiResponse.AICapabilityConfig{}, false, err
	}
	if err := validateAICapabilityUpdateInput(input); err != nil {
		return aiResponse.AICapabilityConfig{}, false, err
	}
	reason := strings.TrimSpace(input.Reason)
	input.PrimaryModelID = strings.TrimSpace(input.PrimaryModelID)
	if input.AuxiliaryModelID != nil {
		value := strings.TrimSpace(*input.AuxiliaryModelID)
		if value == "" {
			input.AuxiliaryModelID = nil
		} else {
			input.AuxiliaryModelID = &value
		}
	}
	if capabilityCode == aiModel.AICapabilityPreferenceSummarize {
		return aiResponse.AICapabilityConfig{}, false, appErrors.AdminInvalidConfig.DefaultMsg()
	}
	if capabilityCode == aiModel.AICapabilityCheckinImageAnalyze {
		input.PointCost = 0
		input.FreeQuotaPerDay = 0
	}

	payload := struct {
		CapabilityCode string                            `json:"capabilityCode"`
		Input          aiRequest.AICapabilityUpdateInput `json:"input"`
	}{CapabilityCode: capabilityCode, Input: input}
	mutationContext, cancelMutation := context.WithTimeout(
		ctx,
		aiCapabilityMutationTimeout,
	)
	defer cancelMutation()
	raw, replayed, err := service.Idempotency.Execute(
		mutationContext,
		actor.AdministratorID,
		"ai_capability_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			definition, err := capabilityDefinition(
				tx,
				capabilityCode,
			)
			if err != nil {
				return nil, err
			}

			current, err := capabilityCurrent(tx, definition)
			if err != nil {
				return nil, err
			}
			currentVersion := int64(0)
			if current != nil {
				currentVersion = current.Version
			}
			if currentVersion != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			primaryModel, _, available, err := availableCapabilityModel(
				tx,
				input.PrimaryModelID,
				definition,
			)
			if err != nil {
				return nil, err
			}
			if !available {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			if definition.RequiredAuxiliaryModelCapability == nil {
				if input.AuxiliaryModelID != nil {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			} else {
				if input.AuxiliaryModelID == nil {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
				_, _, auxiliaryAvailable, auxiliaryErr := availableModelWithCapability(tx, *input.AuxiliaryModelID, *definition.RequiredAuxiliaryModelCapability)
				if auxiliaryErr != nil {
					return nil, auxiliaryErr
				}
				if !auxiliaryAvailable {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			}
			prompt, err := promptFromCurrentCapability(service, tx, capabilityCode, current)
			if err != nil {
				return nil, err
			}
			if err := validatePromptContext(primaryModel, prompt); err != nil {
				return nil, err
			}
			allowedJSON, err := encodeJSON(prompt.AllowedVariables)
			if err != nil {
				return nil, err
			}
			requiredJSON, err := encodeJSON(prompt.RequiredVariables)
			if err != nil {
				return nil, err
			}
			now := service.now()
			next := aiModel.AICapabilityConfig{
				CapabilityCode:              capabilityCode,
				Version:                     currentVersion + 1,
				PrimaryModelID:              input.PrimaryModelID,
				AuxiliaryModelID:            input.AuxiliaryModelID,
				PointCost:                   input.PointCost,
				DailyLimitPerUser:           input.DailyLimitPerUser,
				TimeoutMS:                   input.TimeoutMS,
				FreeQuotaPerDay:             input.FreeQuotaPerDay,
				PromptMode:                  prompt.Mode,
				PromptPresetVersion:         prompt.PresetVersion,
				SystemPrompt:                prompt.SystemPrompt,
				UserPromptTemplate:          prompt.UserPromptTemplate,
				PromptAllowedVariablesJSON:  allowedJSON,
				PromptRequiredVariablesJSON: requiredJSON,
				PromptOutputSchemaVersion:   prompt.OutputSchemaVersion,
				PromptHash:                  prompt.ContentHash,
				AppliedByID:                 actor.AdministratorID,
				AppliedByUsername:           actor.Username,
				AppliedByNickname:           actor.Nickname,
				AppliedAt:                   now,
				Reason:                      reason,
			}
			if err := persistCapabilityConfig(
				tx,
				current,
				next,
				"save AI capability config",
			); err != nil {
				return nil, err
			}
			if capabilityCode == aiModel.AICapabilityCheckinImageAnalyze {
				if err := syncPreferenceSummarizeConfig(
					service,
					tx,
					actor,
					input.PrimaryModelID,
					input.TimeoutMS,
					reason,
					now,
				); err != nil {
					return nil, err
				}
			}
			if err := recordAIChange(
				tx,
				now,
				"ai_capability",
				capabilityCode,
				"update",
				next.Version,
				actor,
				&reason,
			); err != nil {
				return nil, err
			}
			latest, err := promptPreset(tx, capabilityCode)
			if err != nil {
				return nil, err
			}
			after, err := capabilityConfigResponse(next, latest.Version, false)
			if err != nil {
				return nil, err
			}
			var before interface{}
			if current != nil {
				before, err = capabilityConfigResponse(*current, latest.Version, false)
				if err != nil {
					return nil, err
				}
			}
			if err := service.writeMutationAudit(
				mutationContext,
				tx,
				actor,
				"update_ai_capability",
				"ai_capability",
				capabilityCode,
				reason,
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[aiResponse.AICapabilityConfig](raw, replayed, err)
}

// TestAICapabilityPrompt 测试当前主模型和管理员提交的提示词。
func (service *AIService) TestAICapabilityPrompt(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input aiRequest.AIPromptTestInput,
) (aiResponse.AIPromptTestResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.AIPromptTestResult{}, false, err
	}
	if strings.TrimSpace(actor.RequestID) == "" {
		return aiResponse.AIPromptTestResult{}, false,
			appErrors.AdminBadRequest.DefaultMsg()
	}

	payload := struct {
		CapabilityCode string                      `json:"capabilityCode"`
		Input          aiRequest.AIPromptTestInput `json:"input"`
	}{CapabilityCode: capabilityCode, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_capability_prompt_test",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			definition, err := capabilityDefinition(tx, capabilityCode)
			if err != nil {
				return nil, err
			}
			current, err := capabilityCurrent(tx, definition)
			if err != nil {
				return nil, err
			}
			if current == nil {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			prompt, validation, err := service.resolvePromptInput(
				tx,
				capabilityCode,
				input.Prompt,
			)
			if err != nil {
				return nil, err
			}
			if !validation.Valid {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			systemPrompt, err := renderPromptTemplate(
				prompt.SystemPrompt,
				prompt.AllowedVariables,
				prompt.RequiredVariables,
				input.Variables,
			)
			if err != nil {
				return nil, err
			}
			userPrompt, err := renderPromptTemplate(
				prompt.UserPromptTemplate,
				prompt.AllowedVariables,
				prompt.RequiredVariables,
				input.Variables,
			)
			if err != nil {
				return nil, err
			}
			model, provider, available, err := availableCapabilityModel(
				tx,
				current.PrimaryModelID,
				definition,
			)
			if err != nil {
				return nil, err
			}
			if !available {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			if err := validatePromptContext(model, prompt); err != nil {
				return nil, err
			}
			runResult, err := service.Connector.TestPrompt(
				ctx,
				provider,
				model,
				AIPromptRunRequest{
					ModelCapability: definition.RequiredModelCapability,
					SystemPrompt:    systemPrompt,
					UserPrompt:      userPrompt,
				},
				time.Duration(current.TimeoutMS)*time.Millisecond,
			)
			if err != nil {
				return nil, appErrors.AdminProviderFailed.Wrap(err, "test AI capability prompt")
			}
			now := service.now()
			record := aiModel.AIPromptTestRecord{
				ID:                    commonModel.NewID(),
				CapabilityCode:        capabilityCode,
				ConfigVersion:         current.Version,
				ModelID:               model.ID,
				PromptHash:            prompt.ContentHash,
				AdministratorID:       actor.AdministratorID,
				AdministratorUsername: actor.Username,
				RequestID:             actor.RequestID,
				Success:               runResult.Success,
				DurationMS:            runResult.DurationMS,
				InputTokens:           runResult.InputTokens,
				OutputTokens:          runResult.OutputTokens,
				EstimatedCostCNY:      estimatePromptTestCost(model, runResult),
				SafeMessage:           runResult.SafeMessage,
				CreatedAt:             now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "record AI prompt test")
			}
			result, err := promptTestResponse(tx, record)
			if err != nil {
				return nil, err
			}
			result.TestedBy = serviceCommon.AdministratorSummary(
				actor.AdministratorID,
				actor.Username,
				actor.Nickname,
			)
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"test_ai_capability_prompt",
				"ai_capability",
				capabilityCode,
				"测试 AI 能力提示词",
				idempotencyKey,
				map[string]interface{}{
					"configVersion": current.Version,
					"modelId":       model.ID,
					"promptHash":    prompt.ContentHash,
				},
				result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[aiResponse.AIPromptTestResult](raw, replayed, err)
}

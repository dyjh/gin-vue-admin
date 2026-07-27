package service

import (
	"context"
	"errors"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func capabilityDefinition(
	db *gorm.DB,
	capabilityCode string,
) (orderfoodModel.AICapabilityDefinition, error) {
	var definition orderfoodModel.AICapabilityDefinition
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
	definition orderfoodModel.AICapabilityDefinition,
) (*orderfoodModel.AICapabilityConfig, error) {
	var config orderfoodModel.AICapabilityConfig
	if err := db.First(&config, "capability_code = ?", definition.Code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "get current AI capability")
	}
	return &config, nil
}

// validateAICapabilityUpdateInput 校验服务层直接调用也必须遵守的能力配置边界。
func validateAICapabilityUpdateInput(
	input orderfoodRequest.AICapabilityUpdateInput,
) error {
	reason := strings.TrimSpace(input.Reason)
	primaryModelID := strings.TrimSpace(input.PrimaryModelID)
	if primaryModelID == "" ||
		len([]rune(primaryModelID)) > 64 ||
		input.PointCost < 0 || input.PointCost > 100000 ||
		input.DailyLimitPerUser < 0 || input.DailyLimitPerUser > 100000 ||
		input.FreeQuotaPerDay < 0 || input.FreeQuotaPerDay > 100000 ||
		input.TimeoutMS < 1000 || input.TimeoutMS > 120000 ||
		input.ExpectedVersion < 0 ||
		len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// fixedCapabilityValidationRules 返回与实际运行时代码一致的只读校验说明。
func fixedCapabilityValidationRules(capabilityCode string) []string {
	rules := map[string][]string{
		orderfoodModel.AICapabilityDishTextExtract: {
			"返回内容必须能解析为服务端固定的菜品草稿结构",
			"结果只回填可编辑草稿，不直接创建、公开或推荐菜品",
		},
		orderfoodModel.AICapabilityRecipeImageExtract: {
			"返回内容必须能解析为服务端固定的菜品草稿结构",
			"只处理已通过上传审核的图片，结果只回填可编辑草稿",
		},
		orderfoodModel.AICapabilityDishCoverCreate: {
			"只接受合法 Base64 图片或不含用户凭据的 HTTPS 图片地址",
			"图片必须通过非空、文件大小和文件类型校验后才能保存",
		},
		orderfoodModel.AICapabilityCheckinImageAnalyze: {
			"返回内容必须能解析为服务端固定的偏好事实结构",
			"置信度、词条清洗、去重和数量上限由服务端固定执行",
		},
		orderfoodModel.AICapabilityMealSuggest: {
			"固定校验菜品数量、字段边界、分类、标签、单位、配料与步骤一致性",
			"固定拦截重复菜名和违反用户硬条件的结果",
			"标准菜品索引校验及失败重试只由“标准菜品索引”页面的当前开关控制",
		},
		orderfoodModel.AICapabilityPrepSequence: {
			"结果必须完整覆盖原饭局步骤，不允许新增、遗漏或重复步骤",
			"并行步骤必须引用有效步骤，且不能复现用户明确排除的旧方案",
			"预计时长必须处于服务端固定范围",
		},
	}
	result := rules[capabilityCode]
	return append([]string(nil), result...)
}

func capabilityConfigResponse(
	version orderfoodModel.AICapabilityConfig,
	latestPresetVersion int64,
	includePromptBodies bool,
) (orderfoodResponse.AICapabilityConfig, error) {
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
		return orderfoodResponse.AICapabilityConfig{}, err
	}
	return orderfoodResponse.AICapabilityConfig{
		CapabilityCode:       version.CapabilityCode,
		PrimaryModelID:       version.PrimaryModelID,
		PointCost:            version.PointCost,
		DailyLimitPerUser:    version.DailyLimitPerUser,
		TimeoutMS:            version.TimeoutMS,
		FreeQuotaPerDay:      version.FreeQuotaPerDay,
		FixedValidationRules: fixedCapabilityValidationRules(version.CapabilityCode),
		Prompt:               prompt,
		PromptMode:           externalPromptMode(version.PromptMode),
		PromptHash:           version.PromptHash,
		Version:              version.Version,
		UpdatedBy: aiAdministratorSummary(
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
	var model orderfoodModel.AIModel
	if err := db.Select("name").First(&model, "id = ?", modelID).Error; err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "get AI capability model name")
	}
	return model.Name, nil
}

func capabilitySummary(
	db *gorm.DB,
	definition orderfoodModel.AICapabilityDefinition,
) (orderfoodResponse.AICapabilitySummary, error) {
	preset, err := promptPreset(db, definition.Code)
	if err != nil {
		return orderfoodResponse.AICapabilitySummary{}, err
	}
	summary := orderfoodResponse.AICapabilitySummary{
		Code:                    definition.Code,
		Name:                    definition.Name,
		ClientFeatureCode:       definition.ClientFeatureCode,
		RequiredModelCapability: string(definition.RequiredModelCapability),
		Version:                 0,
		PromptMode:              "default",
		PromptPresetVersion:     preset.Version,
		PromptHash:              preset.ContentHash,
		SortOrder:               definition.SortOrder,
		UpdatedAt:               definition.UpdatedAt,
	}
	current, err := capabilityCurrent(db, definition)
	if err != nil {
		return summary, err
	}
	if current == nil {
		return summary, nil
	}
	summary.Version = current.Version
	_, _, summary.Ready, err = availableCapabilityModel(
		db,
		current.PrimaryModelID,
		definition,
	)
	if err != nil {
		return summary, err
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
	return summary, nil
}

// ListAICapabilities 查询 AI 能力列表。
func (service *AIService) ListAICapabilities(
	ctx context.Context,
	query orderfoodRequest.AICapabilityListQuery,
) ([]orderfoodResponse.AICapabilitySummary, error) {
	if err := service.EnsureDefaults(ctx); err != nil {
		return nil, err
	}
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
	statement := db.Model(&orderfoodModel.AICapabilityDefinition{})
	if query.ClientFeatureCode != "" {
		statement = statement.Where("client_feature_code = ?", query.ClientFeatureCode)
	}
	var definitions []orderfoodModel.AICapabilityDefinition
	if err := statement.Order(column + " " + order).Find(&definitions).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list AI capabilities")
	}
	list := make([]orderfoodResponse.AICapabilitySummary, 0, len(definitions))
	for _, definition := range definitions {
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
	actor orderfoodRequest.AIAdminActor,
	capabilityCode string,
	includePromptBodies bool,
) (orderfoodResponse.AICapabilityWorkspace, error) {
	if err := service.EnsureDefaults(ctx); err != nil {
		return orderfoodResponse.AICapabilityWorkspace{}, err
	}
	db := service.database().WithContext(ctx)
	definition, err := capabilityDefinition(db, capabilityCode)
	if err != nil {
		return orderfoodResponse.AICapabilityWorkspace{}, err
	}
	if includePromptBodies {
		if err := validateAIActor(actor); err != nil {
			return orderfoodResponse.AICapabilityWorkspace{}, err
		}
		if _, err := service.recordPromptAccess(
			ctx,
			actor,
			"read_ai_capability_prompt",
			capabilityCode,
		); err != nil {
			return orderfoodResponse.AICapabilityWorkspace{}, err
		}
	}
	summary, err := capabilitySummary(db, definition)
	if err != nil {
		return orderfoodResponse.AICapabilityWorkspace{}, err
	}
	result := orderfoodResponse.AICapabilityWorkspace{Summary: summary}
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

func availableCapabilityModel(
	db *gorm.DB,
	modelID string,
	definition orderfoodModel.AICapabilityDefinition,
) (orderfoodModel.AIModel, orderfoodModel.AIProvider, bool, error) {
	var model orderfoodModel.AIModel
	if err := db.First(&model, "id = ?", modelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model, orderfoodModel.AIProvider{}, false, nil
		}
		return model, orderfoodModel.AIProvider{}, false,
			appErrors.AdminInternal.Wrap(err, "get AI capability model")
	}
	var provider orderfoodModel.AIProvider
	if err := db.First(&provider, "id = ?", model.ProviderID).Error; err != nil {
		return model, provider, false,
			appErrors.AdminInternal.Wrap(err, "get AI capability provider")
	}
	supports, err := modelSupports(model, definition.RequiredModelCapability)
	if err != nil {
		return model, provider, false, err
	}
	available := model.Enabled &&
		provider.Enabled &&
		strings.TrimSpace(provider.APIKey) != "" &&
		supports
	return model, provider, available, nil
}

func aiCapabilitiesReadiness(
	db *gorm.DB,
) (orderfoodResponse.AICapabilityReadiness, error) {
	var definitions []orderfoodModel.AICapabilityDefinition
	if err := db.Order("sort_order asc").Find(&definitions).Error; err != nil {
		return orderfoodResponse.AICapabilityReadiness{},
			appErrors.AdminInternal.Wrap(err, "list AI capability readiness")
	}
	expected := defaultCapabilities()
	result := orderfoodResponse.AICapabilityReadiness{
		TotalCount:          len(expected),
		UnreadyCapabilities: make([]string, 0, len(expected)),
	}
	definitionByCode := make(map[string]orderfoodModel.AICapabilityDefinition, len(definitions))
	for _, definition := range definitions {
		definitionByCode[definition.Code] = definition
	}
	for _, item := range expected {
		definition, exists := definitionByCode[item.Code]
		if !exists {
			result.UnreadyCapabilities = append(result.UnreadyCapabilities, item.Code)
			continue
		}
		current, err := capabilityCurrent(db, definition)
		if err != nil {
			return result, err
		}
		if current == nil {
			result.UnreadyCapabilities = append(result.UnreadyCapabilities, item.Code)
			continue
		}
		_, _, primaryReady, err := availableCapabilityModel(
			db,
			current.PrimaryModelID,
			definition,
		)
		if err != nil {
			return result, err
		}
		if !primaryReady || strings.TrimSpace(current.PromptHash) == "" {
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
	current *orderfoodModel.AICapabilityConfig,
) (resolvedPrompt, error) {
	if current == nil {
		resolved, validation, err := service.resolvePromptInput(
			tx,
			capabilityCode,
			orderfoodRequest.AIPromptConfigInput{Mode: orderfoodModel.AIPromptPreset},
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

// UpdateAICapability 保存并立即应用 AI 能力配置。
func (service *AIService) UpdateAICapability(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input orderfoodRequest.AICapabilityUpdateInput,
) (orderfoodResponse.AICapabilityConfig, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AICapabilityConfig{}, false, err
	}
	if service == nil || service.Permission == nil || service.Idempotency == nil {
		return orderfoodResponse.AICapabilityConfig{}, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(
		ctx, actor.AuthorityID, PermissionAICapabilityUpdate,
	); err != nil {
		return orderfoodResponse.AICapabilityConfig{}, false, err
	}
	if err := validateAICapabilityUpdateInput(input); err != nil {
		return orderfoodResponse.AICapabilityConfig{}, false, err
	}
	reason := strings.TrimSpace(input.Reason)
	input.PrimaryModelID = strings.TrimSpace(input.PrimaryModelID)
	if err := service.EnsureDefaults(ctx); err != nil {
		return orderfoodResponse.AICapabilityConfig{}, false, err
	}
	payload := struct {
		CapabilityCode string                                   `json:"capabilityCode"`
		Input          orderfoodRequest.AICapabilityUpdateInput `json:"input"`
	}{CapabilityCode: capabilityCode, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_capability_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			definition, err := capabilityDefinition(
				tx.Clauses(clause.Locking{Strength: "UPDATE"}),
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
			next := orderfoodModel.AICapabilityConfig{
				CapabilityCode:              capabilityCode,
				Version:                     currentVersion + 1,
				PrimaryModelID:              input.PrimaryModelID,
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
			if current == nil {
				if err := tx.Create(&next).Error; err != nil {
					return nil, appErrors.AdminInternal.Wrap(err, "create AI capability config")
				}
			} else if err := tx.Save(&next).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "save AI capability config")
			}
			if err := tx.Model(&orderfoodModel.AICapabilityDefinition{}).
				Where("code = ?", capabilityCode).
				Update("updated_at", now).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update AI capability timestamp")
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
				ctx,
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
	return decodeIdempotentResult[orderfoodResponse.AICapabilityConfig](raw, replayed, err)
}

// TestAICapabilityPrompt 测试当前主模型和管理员提交的提示词。
func (service *AIService) TestAICapabilityPrompt(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input orderfoodRequest.AIPromptTestInput,
) (orderfoodResponse.AIPromptTestResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIPromptTestResult{}, false, err
	}
	if strings.TrimSpace(actor.RequestID) == "" {
		return orderfoodResponse.AIPromptTestResult{}, false,
			appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := service.EnsureDefaults(ctx); err != nil {
		return orderfoodResponse.AIPromptTestResult{}, false, err
	}
	payload := struct {
		CapabilityCode string                             `json:"capabilityCode"`
		Input          orderfoodRequest.AIPromptTestInput `json:"input"`
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
			record := orderfoodModel.AIPromptTestRecord{
				ID:                    orderfoodModel.NewID(),
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
			result.TestedBy = aiAdministratorSummary(
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
	return decodeIdempotentResult[orderfoodResponse.AIPromptTestResult](raw, replayed, err)
}

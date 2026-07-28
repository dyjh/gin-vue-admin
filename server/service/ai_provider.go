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

func providerModelCount(db *gorm.DB, providerID string) (int64, error) {
	var count int64
	if err := db.Model(&orderfoodModel.AIModel{}).
		Where("provider_id = ?", providerID).
		Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count provider models")
	}
	return count, nil
}

func providerEnabledModelCount(db *gorm.DB, providerID string) (int64, error) {
	var count int64
	if err := db.Model(&orderfoodModel.AIModel{}).
		Where("provider_id = ? AND enabled = ?", providerID, true).
		Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count enabled provider models")
	}
	return count, nil
}

// providerUsageCount 统计供应商历史调用引用数量。
func providerUsageCount(db *gorm.DB, providerID string) (int64, error) {
	return countOptionalTableReference(db, "of_ai_usages", "provider_id", providerID)
}

func providerLastConnectionTest(
	provider orderfoodModel.AIProvider,
) *orderfoodResponse.ProviderConnectionTestResult {
	if provider.LastTestSuccess == nil ||
		provider.LastTestCategory == nil ||
		provider.LastTestDurationMS == nil ||
		provider.LastTestSafeMessage == nil ||
		provider.LastTestedAt == nil {
		return nil
	}
	return &orderfoodResponse.ProviderConnectionTestResult{
		Success:     *provider.LastTestSuccess,
		Category:    *provider.LastTestCategory,
		DurationMS:  *provider.LastTestDurationMS,
		TestedAt:    *provider.LastTestedAt,
		SafeMessage: *provider.LastTestSafeMessage,
	}
}

func providerReferencedCapabilityCount(db *gorm.DB, providerID string) (int64, error) {
	var count int64
	err := db.Table(orderfoodModel.AICapabilityConfig{}.TableName()+" AS configs").
		Joins(
			"JOIN "+orderfoodModel.AIModel{}.TableName()+
				" AS models ON models.id = configs.primary_model_id",
		).
		Where("models.provider_id = ?", providerID).
		Distinct("configs.capability_code").
		Count(&count).Error
	if err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count provider referenced capabilities")
	}
	return count, nil
}

func providerSummary(
	db *gorm.DB,
	provider orderfoodModel.AIProvider,
) (orderfoodResponse.AIProviderSummary, error) {
	modelCount, err := providerModelCount(db, provider.ID)
	if err != nil {
		return orderfoodResponse.AIProviderSummary{}, err
	}
	enabledModelCount, err := providerEnabledModelCount(db, provider.ID)
	if err != nil {
		return orderfoodResponse.AIProviderSummary{}, err
	}
	capabilityCount, err := providerReferencedCapabilityCount(db, provider.ID)
	if err != nil {
		return orderfoodResponse.AIProviderSummary{}, err
	}
	usageCount, err := providerUsageCount(db, provider.ID)
	if err != nil {
		return orderfoodResponse.AIProviderSummary{}, err
	}
	var maskedLabel *string
	configured := strings.TrimSpace(provider.APIKey) != ""
	if configured {
		label := "已配置（密钥不回显）"
		maskedLabel = &label
	}
	return orderfoodResponse.AIProviderSummary{
		ID:                     provider.ID,
		Name:                   provider.Name,
		Type:                   string(provider.Type),
		BaseURL:                provider.BaseURL,
		CredentialConfigured:   configured,
		CredentialMaskedLabel:  maskedLabel,
		TimeoutMS:              provider.TimeoutMS,
		Enabled:                provider.Enabled,
		ModelCount:             modelCount,
		EnabledModelCount:      enabledModelCount,
		EnabledCapabilityCount: capabilityCount,
		UsageCount:             usageCount,
		LastConnectionTest:     providerLastConnectionTest(provider),
		Version:                provider.Version,
		UpdatedAt:              provider.UpdatedAt,
	}, nil
}

// providerReferenceDetails 查询供应商关联的模型和能力配置明细。
func providerReferenceDetails(
	db *gorm.DB,
	providerID string,
) ([]orderfoodResponse.AIProviderModelReference, []orderfoodResponse.AIProviderCapabilityReference, error) {
	var models []orderfoodModel.AIModel
	if err := db.Where("provider_id = ?", providerID).
		Order("enabled DESC, name ASC, id ASC").
		Find(&models).Error; err != nil {
		return nil, nil, appErrors.AdminInternal.Wrap(err, "list provider referenced models")
	}
	modelReferences := make([]orderfoodResponse.AIProviderModelReference, 0, len(models))
	for index := range models {
		modelReferences = append(modelReferences, orderfoodResponse.AIProviderModelReference{
			ID:       models[index].ID,
			Name:     models[index].Name,
			ModelKey: models[index].ModelKey,
			Enabled:  models[index].Enabled,
		})
	}
	capabilityReferences := make([]orderfoodResponse.AIProviderCapabilityReference, 0)
	if len(models) == 0 {
		return modelReferences, capabilityReferences, nil
	}
	if err := db.Table(orderfoodModel.AICapabilityConfig{}.TableName()+" AS configs").
		Select(
			"configs.capability_code, COALESCE(definitions.name, configs.capability_code) AS capability_name, "+
				"models.id AS model_id, models.name AS model_name",
		).
		Joins(
			"JOIN "+orderfoodModel.AIModel{}.TableName()+
				" AS models ON models.id = configs.primary_model_id",
		).
		Joins(
			"LEFT JOIN "+orderfoodModel.AICapabilityDefinition{}.TableName()+
				" AS definitions ON definitions.code = configs.capability_code",
		).
		Where("models.provider_id = ?", providerID).
		Order("definitions.sort_order ASC, configs.capability_code ASC").
		Scan(&capabilityReferences).Error; err != nil {
		return nil, nil, appErrors.AdminInternal.Wrap(err, "list provider referenced capabilities")
	}
	return modelReferences, capabilityReferences, nil
}

func providerDetail(
	db *gorm.DB,
	provider orderfoodModel.AIProvider,
) (orderfoodResponse.AIProviderDetail, error) {
	summary, err := providerSummary(db, provider)
	if err != nil {
		return orderfoodResponse.AIProviderDetail{}, err
	}
	modelReferences, capabilityReferences, err := providerReferenceDetails(db, provider.ID)
	if err != nil {
		return orderfoodResponse.AIProviderDetail{}, err
	}
	return orderfoodResponse.AIProviderDetail{
		AIProviderSummary:      summary,
		ReferencedModels:       modelReferences,
		ReferencedCapabilities: capabilityReferences,
		Remark:                 provider.Remark,
		CreatedAt:              provider.CreatedAt,
	}, nil
}

// validateAIProviderListQuery 校验供应商列表筛选、分页和排序参数。
func validateAIProviderListQuery(query orderfoodRequest.AIProviderListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.Keyword))) > 60 ||
		(query.SortBy != "createdAt" && query.SortBy != "updatedAt" && query.SortBy != "name") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.Type != "" {
		return validateProviderType(query.Type)
	}
	return nil
}

// validateAIProviderIdentity 校验供应商公开ID。
func validateAIProviderIdentity(providerID string) error {
	length := len([]rune(strings.TrimSpace(providerID)))
	if length < 1 || length > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// validateAIProviderCommonInput 校验供应商名称、地址、超时和备注。
func validateAIProviderCommonInput(name string, baseURL string, timeoutMS int, remark *string) error {
	nameLength := len([]rune(strings.TrimSpace(name)))
	if nameLength < 1 || nameLength > 60 ||
		len([]rune(strings.TrimSpace(baseURL))) > 300 ||
		timeoutMS < 1000 || timeoutMS > 120000 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if remark != nil && len([]rune(strings.TrimSpace(*remark))) > 300 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return validateAIBaseURL(baseURL)
}

// ListAIProviders 分页查询AI供应商列表。
func (service *AIService) ListAIProviders(
	ctx context.Context,
	query orderfoodRequest.AIProviderListQuery,
) (orderfoodResponse.Page[orderfoodResponse.AIProviderSummary], error) {
	query.ApplyDefaults()
	if err := validateAIProviderListQuery(query); err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{}, err
	}
	sortColumns := map[string]string{
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		"name":      "name",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	order := strings.ToLower(query.SortOrder)
	if order != "asc" && order != "desc" {
		return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database().WithContext(ctx)
	statement := db.Model(&orderfoodModel.AIProvider{}).Where(
		"type IN ?",
		[]orderfoodModel.AIProviderType{
			orderfoodModel.AIProviderBailian,
			orderfoodModel.AIProviderDeepSeek,
			orderfoodModel.AIProviderOpenAI,
		},
	)
	if query.Keyword != "" {
		search := "%" + strings.TrimSpace(query.Keyword) + "%"
		statement = statement.Where("name LIKE ?", search)
	}
	if query.Type != "" {
		statement = statement.Where("type = ?", query.Type)
	}
	if query.Enabled != nil {
		statement = statement.Where("enabled = ?", *query.Enabled)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{}, appErrors.AdminInternal.Wrap(err, "count AI providers")
	}
	var providers []orderfoodModel.AIProvider
	if err := statement.Order(column + " " + order).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&providers).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{}, appErrors.AdminInternal.Wrap(err, "list AI providers")
	}
	list := make([]orderfoodResponse.AIProviderSummary, 0, len(providers))
	for _, provider := range providers {
		item, err := providerSummary(db, provider)
		if err != nil {
			return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{}, err
		}
		list = append(list, item)
	}
	return orderfoodResponse.Page[orderfoodResponse.AIProviderSummary]{
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
		List:     list,
	}, nil
}

// AIProviderDetail 获取 AI 供应商详情及最近一次连接测试结果。
func (service *AIService) AIProviderDetail(
	ctx context.Context,
	providerID string,
) (orderfoodResponse.AIProviderDetail, error) {
	if err := validateAIProviderIdentity(providerID); err != nil {
		return orderfoodResponse.AIProviderDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	providerID = strings.TrimSpace(providerID)
	db := service.database().WithContext(ctx)
	var provider orderfoodModel.AIProvider
	if err := db.First(&provider, "id = ?", providerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodResponse.AIProviderDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return orderfoodResponse.AIProviderDetail{}, appErrors.AdminInternal.Wrap(err, "get AI provider")
	}
	return providerDetail(db, provider)
}

// AIProviderModelOptions 获取供应商实时模型列表并标记已配置项。
func (service *AIService) AIProviderModelOptions(
	ctx context.Context,
	providerID string,
) ([]orderfoodResponse.AIProviderModelOption, error) {
	if err := validateAIProviderIdentity(providerID); err != nil {
		return nil, appErrors.AdminBadRequest.DefaultMsg()
	}
	providerID = strings.TrimSpace(providerID)
	db := service.database().WithContext(ctx)
	var provider orderfoodModel.AIProvider
	if err := db.First(&provider, "id = ?", providerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.AdminNotFound.DefaultMsg()
		}
		return nil, appErrors.AdminInternal.Wrap(err, "get AI provider for model options")
	}
	modelKeys, err := service.Connector.ListModels(
		ctx,
		provider,
		time.Duration(provider.TimeoutMS)*time.Millisecond,
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, appErrors.AdminProviderTimeout.Wrap(err, "list AI provider models")
		}
		return nil, appErrors.AdminProviderFailed.Wrap(err, "list AI provider models")
	}
	var configuredModelKeys []string
	if err := db.Model(&orderfoodModel.AIModel{}).
		Where("provider_id = ?", provider.ID).
		Pluck("model_key", &configuredModelKeys).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list configured AI provider models")
	}
	configured := make(map[string]struct{}, len(configuredModelKeys))
	for _, modelKey := range configuredModelKeys {
		configured[modelKey] = struct{}{}
	}
	options := make([]orderfoodResponse.AIProviderModelOption, 0, len(modelKeys))
	for _, modelKey := range modelKeys {
		_, exists := configured[modelKey]
		options = append(options, orderfoodResponse.AIProviderModelOption{
			Name:       modelKey,
			ModelKey:   modelKey,
			Configured: exists,
		})
	}
	return options, nil
}

// AIModelProviderOptions 获取模型页面可用的供应商基础选项。
func (service *AIService) AIModelProviderOptions(
	ctx context.Context,
) ([]orderfoodResponse.AIModelProviderOption, error) {
	var providers []orderfoodModel.AIProvider
	if err := service.database().WithContext(ctx).
		Select("id", "name", "type", "enabled").
		Where(
			"type IN ?",
			[]orderfoodModel.AIProviderType{
				orderfoodModel.AIProviderBailian,
				orderfoodModel.AIProviderDeepSeek,
				orderfoodModel.AIProviderOpenAI,
			},
		).
		Order("enabled DESC, name ASC, id ASC").
		Find(&providers).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list AI model provider options")
	}
	options := make([]orderfoodResponse.AIModelProviderOption, 0, len(providers))
	for _, provider := range providers {
		options = append(options, orderfoodResponse.AIModelProviderOption{
			ID:      provider.ID,
			Name:    provider.Name,
			Type:    string(provider.Type),
			Enabled: provider.Enabled,
		})
	}
	return options, nil
}

// CreateAIProvider 创建AI供应商。
func (service *AIService) CreateAIProvider(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	idempotencyKey string,
	input orderfoodRequest.AIProviderCreateInput,
) (orderfoodResponse.AIProviderDetail, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	if err := validateProviderType(input.Type); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	if err := validateAIProviderCommonInput(input.Name, input.BaseURL, input.TimeoutMS, input.Remark); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	if len([]rune(strings.TrimSpace(input.APIKey))) > 500 {
		return orderfoodResponse.AIProviderDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateProviderAPIKey(input.APIKey); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_provider_create",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			var existing int64
			if err := tx.Model(&orderfoodModel.AIProvider{}).
				Where("type = ? OR name = ?", input.Type, strings.TrimSpace(input.Name)).
				Count(&existing).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check AI provider uniqueness")
			}
			if existing > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			now := service.now()
			provider := orderfoodModel.AIProvider{
				ID:                orderfoodModel.NewID(),
				Name:              strings.TrimSpace(input.Name),
				Type:              input.Type,
				BaseURL:           strings.TrimRight(strings.TrimSpace(input.BaseURL), "/"),
				APIKey:            strings.TrimSpace(input.APIKey),
				TimeoutMS:         input.TimeoutMS,
				RetryCount:        0,
				Enabled:           false,
				Remark:            normalizeOptionalString(input.Remark),
				Version:           1,
				UpdatedByID:       actor.AdministratorID,
				UpdatedByUsername: actor.Username,
				UpdatedByNickname: actor.Nickname,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			if err := tx.Create(&provider).Error; err != nil {
				return nil, appErrors.AdminAlreadyExists.Wrap(err, "create AI provider")
			}
			if err := recordAIChange(
				tx,
				now,
				"ai_provider",
				provider.ID,
				"create",
				provider.Version,
				actor,
				nil,
			); err != nil {
				return nil, err
			}
			after, err := providerDetail(tx, provider)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"create_ai_provider",
				"ai_provider",
				provider.ID,
				"创建 AI 供应商",
				idempotencyKey,
				nil,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.AIProviderDetail](raw, replayed, err)
}

// UpdateAIProvider 更新AI供应商。
func (service *AIService) UpdateAIProvider(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	providerID string,
	idempotencyKey string,
	input orderfoodRequest.AIProviderUpdateInput,
) (orderfoodResponse.AIProviderDetail, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	if err := validateAIProviderIdentity(providerID); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	providerID = strings.TrimSpace(providerID)
	if input.ExpectedVersion < 1 {
		return orderfoodResponse.AIProviderDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateAIProviderCommonInput(input.Name, input.BaseURL, input.TimeoutMS, input.Remark); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	if input.APIKey != nil {
		if len([]rune(strings.TrimSpace(*input.APIKey))) > 500 {
			return orderfoodResponse.AIProviderDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
		}
		if err := validateProviderAPIKey(*input.APIKey); err != nil {
			return orderfoodResponse.AIProviderDetail{}, false, err
		}
	}
	payload := struct {
		ProviderID string                                 `json:"providerId"`
		Input      orderfoodRequest.AIProviderUpdateInput `json:"input"`
	}{ProviderID: providerID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_provider_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var provider orderfoodModel.AIProvider
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&provider, "id = ?", providerID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI provider for update")
			}
			if provider.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before := provider
			var duplicate int64
			if err := tx.Model(&orderfoodModel.AIProvider{}).
				Where("name = ? AND id <> ?", strings.TrimSpace(input.Name), provider.ID).
				Count(&duplicate).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check AI provider name")
			}
			if duplicate > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			provider.Name = strings.TrimSpace(input.Name)
			nextBaseURL := strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")
			connectionConfigChanged := provider.BaseURL != nextBaseURL ||
				provider.TimeoutMS != input.TimeoutMS
			provider.BaseURL = nextBaseURL
			if input.APIKey != nil {
				nextAPIKey := strings.TrimSpace(*input.APIKey)
				connectionConfigChanged = connectionConfigChanged ||
					provider.APIKey != nextAPIKey
				provider.APIKey = nextAPIKey
			}
			provider.TimeoutMS = input.TimeoutMS
			provider.Remark = normalizeOptionalString(input.Remark)
			if connectionConfigChanged {
				// 连接参数改变后，旧测试结果不再代表当前配置，必须重新测试才能再次启用。
				provider.LastTestSuccess = nil
				provider.LastTestCategory = nil
				provider.LastTestDurationMS = nil
				provider.LastTestSafeMessage = nil
				provider.LastTestedAt = nil
			}
			provider.Version++
			provider.UpdatedByID = actor.AdministratorID
			provider.UpdatedByUsername = actor.Username
			provider.UpdatedByNickname = actor.Nickname
			provider.UpdatedAt = service.now()
			if err := tx.Save(&provider).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update AI provider")
			}
			if err := recordAIChange(
				tx,
				provider.UpdatedAt,
				"ai_provider",
				provider.ID,
				"update",
				provider.Version,
				actor,
				nil,
			); err != nil {
				return nil, err
			}
			after, err := providerDetail(tx, provider)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_ai_provider",
				"ai_provider",
				provider.ID,
				"编辑 AI 供应商",
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.AIProviderDetail](raw, replayed, err)
}

// UpdateAIProviderStatus 更新AI供应商状态。
func (service *AIService) UpdateAIProviderStatus(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	providerID string,
	idempotencyKey string,
	input orderfoodRequest.AIStatusUpdateInput,
) (orderfoodResponse.AIProviderDetail, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	if err := validateAIProviderIdentity(providerID); err != nil || input.ExpectedVersion < 1 {
		return orderfoodResponse.AIProviderDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	providerID = strings.TrimSpace(providerID)
	if input.Enabled == nil {
		return orderfoodResponse.AIProviderDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateAIReason(input.Reason); err != nil {
		return orderfoodResponse.AIProviderDetail{}, false, err
	}
	payload := struct {
		ProviderID string                               `json:"providerId"`
		Input      orderfoodRequest.AIStatusUpdateInput `json:"input"`
	}{ProviderID: providerID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_provider_status_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var provider orderfoodModel.AIProvider
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&provider, "id = ?", providerID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI provider for status update")
			}
			if provider.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if provider.Enabled == *input.Enabled {
				after, err := providerDetail(tx, provider)
				if err != nil {
					return nil, err
				}
				if err := service.writeMutationAudit(
					ctx,
					tx,
					actor,
					"update_ai_provider_status",
					"ai_provider",
					provider.ID,
					input.Reason,
					idempotencyKey,
					provider,
					after,
				); err != nil {
					return nil, err
				}
				return after, nil
			}
			before := provider
			if !*input.Enabled {
				var enabledModels int64
				if err := tx.Model(&orderfoodModel.AIModel{}).
					Where("provider_id = ? AND enabled = ?", provider.ID, true).
					Count(&enabledModels).Error; err != nil {
					return nil, appErrors.AdminInternal.Wrap(err, "count enabled provider models")
				}
				referencedCapabilities, err := providerReferencedCapabilityCount(tx, provider.ID)
				if err != nil {
					return nil, err
				}
				if enabledModels > 0 || referencedCapabilities > 0 {
					return nil, appErrors.AdminResourceInUse.DefaultMsg()
				}
			} else {
				if strings.TrimSpace(provider.APIKey) == "" ||
					provider.LastTestSuccess == nil ||
					!*provider.LastTestSuccess {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			}
			provider.Enabled = *input.Enabled
			provider.Version++
			provider.UpdatedByID = actor.AdministratorID
			provider.UpdatedByUsername = actor.Username
			provider.UpdatedByNickname = actor.Nickname
			provider.UpdatedAt = service.now()
			reason := strings.TrimSpace(input.Reason)
			provider.LastChangeReason = &reason
			if err := tx.Save(&provider).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update AI provider status")
			}
			if err := recordAIChange(
				tx,
				provider.UpdatedAt,
				"ai_provider",
				provider.ID,
				"status",
				provider.Version,
				actor,
				&reason,
			); err != nil {
				return nil, err
			}
			after, err := providerDetail(tx, provider)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_ai_provider_status",
				"ai_provider",
				provider.ID,
				input.Reason,
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.AIProviderDetail](raw, replayed, err)
}

func countOptionalTableReference(
	db *gorm.DB,
	table string,
	column string,
	value string,
) (int64, error) {
	if !db.Migrator().HasTable(table) {
		return 0, nil
	}
	var count int64
	if err := db.Table(table).Where(column+" = ?", value).Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count referenced AI records")
	}
	return count, nil
}

// DeleteAIProvider 删除AI供应商。
func (service *AIService) DeleteAIProvider(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	providerID string,
	idempotencyKey string,
	input orderfoodRequest.AIDeleteInput,
) (orderfoodResponse.DeletedResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.DeletedResult{}, false, err
	}
	if err := validateAIProviderIdentity(providerID); err != nil || input.ExpectedVersion < 1 {
		return orderfoodResponse.DeletedResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	providerID = strings.TrimSpace(providerID)
	if err := validateAIReason(input.Reason); err != nil {
		return orderfoodResponse.DeletedResult{}, false, err
	}
	payload := struct {
		ProviderID string                         `json:"providerId"`
		Input      orderfoodRequest.AIDeleteInput `json:"input"`
	}{ProviderID: providerID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_provider_delete",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var provider orderfoodModel.AIProvider
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&provider, "id = ?", providerID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI provider for delete")
			}
			if provider.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			modelCount, err := providerModelCount(tx, provider.ID)
			if err != nil {
				return nil, err
			}
			usageCount, err := countOptionalTableReference(
				tx,
				"of_ai_usages",
				"provider_id",
				provider.ID,
			)
			if err != nil {
				return nil, err
			}
			if modelCount > 0 || usageCount > 0 {
				return nil, appErrors.AdminResourceInUse.DefaultMsg()
			}
			if err := tx.Delete(&provider).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "delete AI provider")
			}
			reason := strings.TrimSpace(input.Reason)
			if err := recordAIChange(
				tx,
				service.now(),
				"ai_provider",
				provider.ID,
				"delete",
				provider.Version,
				actor,
				&reason,
			); err != nil {
				return nil, err
			}
			after := orderfoodResponse.DeletedResult{Deleted: true}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"delete_ai_provider",
				"ai_provider",
				provider.ID,
				input.Reason,
				idempotencyKey,
				provider,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.DeletedResult](raw, replayed, err)
}

// TestAIProviderConnection 测试AI供应商连接。
func (service *AIService) TestAIProviderConnection(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	providerID string,
	idempotencyKey string,
	input orderfoodRequest.AIProviderConnectionTestInput,
) (orderfoodResponse.ProviderConnectionTestResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.ProviderConnectionTestResult{}, false, err
	}
	if err := validateAIProviderIdentity(providerID); err != nil ||
		input.ExpectedVersion < 1 ||
		(input.TimeoutMS != nil && (*input.TimeoutMS < 1000 || *input.TimeoutMS > 30000)) {
		return orderfoodResponse.ProviderConnectionTestResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	providerID = strings.TrimSpace(providerID)
	payload := struct {
		ProviderID string                                         `json:"providerId"`
		Input      orderfoodRequest.AIProviderConnectionTestInput `json:"input"`
	}{ProviderID: providerID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_provider_connection_test",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var provider orderfoodModel.AIProvider
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&provider, "id = ?", providerID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI provider for connection test")
			}
			if provider.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before := map[string]interface{}{
				"version":              provider.Version,
				"lastConnectionTestAt": provider.LastTestedAt,
			}
			timeoutMS := provider.TimeoutMS
			if input.TimeoutMS != nil {
				timeoutMS = *input.TimeoutMS
			}
			connectionResult, err := service.Connector.TestConnection(
				ctx,
				provider,
				time.Duration(timeoutMS)*time.Millisecond,
			)
			if err != nil {
				return nil, appErrors.AdminProviderFailed.Wrap(err, "test AI provider connection")
			}
			testedAt := service.now()
			provider.LastTestSuccess = &connectionResult.Success
			provider.LastTestCategory = &connectionResult.Category
			provider.LastTestDurationMS = &connectionResult.DurationMS
			provider.LastTestSafeMessage = &connectionResult.SafeMessage
			provider.LastTestedAt = &testedAt
			provider.UpdatedAt = testedAt
			if err := tx.Save(&provider).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "save AI provider connection test")
			}
			if err := recordAIChange(
				tx,
				testedAt,
				"ai_provider",
				provider.ID,
				"connection_test",
				provider.Version,
				actor,
				nil,
			); err != nil {
				return nil, err
			}
			after := orderfoodResponse.ProviderConnectionTestResult{
				Success:     connectionResult.Success,
				Category:    connectionResult.Category,
				DurationMS:  connectionResult.DurationMS,
				TestedAt:    testedAt,
				SafeMessage: connectionResult.SafeMessage,
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"test_ai_provider_connection",
				"ai_provider",
				provider.ID,
				"测试 AI 供应商连接",
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.ProviderConnectionTestResult](raw, replayed, err)
}

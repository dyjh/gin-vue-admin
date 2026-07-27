package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var nonNegativeDecimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,8})?$`)

func normalizePrice(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if !nonNegativeDecimalPattern.MatchString(trimmed) {
		return nil, appErrors.AdminInvalidConfig.DefaultMsg()
	}
	return &trimmed, nil
}

func validateModelCapabilities(
	capabilities []orderfoodModel.AIModelCapability,
) ([]string, error) {
	if len(capabilities) == 0 {
		return nil, appErrors.AdminInvalidConfig.DefaultMsg()
	}
	seen := make(map[orderfoodModel.AIModelCapability]struct{}, len(capabilities))
	result := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		switch capability {
		case orderfoodModel.AIModelText,
			orderfoodModel.AIModelVision,
			orderfoodModel.AIModelImageGeneration:
		default:
			return nil, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		if _, exists := seen[capability]; exists {
			return nil, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		seen[capability] = struct{}{}
		result = append(result, string(capability))
	}
	return result, nil
}

func decodeModelCapabilities(
	model orderfoodModel.AIModel,
) ([]string, error) {
	var capabilities []string
	if err := decodeJSON(model.CapabilitiesJSON, &capabilities); err != nil {
		return nil, err
	}
	return capabilities, nil
}

func modelSupports(
	model orderfoodModel.AIModel,
	required orderfoodModel.AIModelCapability,
) (bool, error) {
	capabilities, err := decodeModelCapabilities(model)
	if err != nil {
		return false, err
	}
	for _, capability := range capabilities {
		if capability == string(required) {
			return true, nil
		}
	}
	return false, nil
}

func modelReferencedCapabilityCount(db *gorm.DB, modelID string) (int64, error) {
	var count int64
	err := db.Model(&orderfoodModel.AICapabilityConfig{}).
		Where("primary_model_id = ?", modelID).
		Count(&count).Error
	if err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count model referenced capabilities")
	}
	return count, nil
}

// modelCapabilityReferences 查询引用指定模型的能力明细。
func modelCapabilityReferences(
	db *gorm.DB,
	modelID string,
) ([]orderfoodResponse.AIModelCapabilityReference, error) {
	references := make([]orderfoodResponse.AIModelCapabilityReference, 0)
	if err := db.Table(orderfoodModel.AICapabilityConfig{}.TableName()+" AS configs").
		Select(
			"configs.capability_code, COALESCE(definitions.name, configs.capability_code) AS capability_name",
		).
		Joins(
			"LEFT JOIN "+orderfoodModel.AICapabilityDefinition{}.TableName()+
				" AS definitions ON definitions.code = configs.capability_code",
		).
		Where("configs.primary_model_id = ?", modelID).
		Order("definitions.sort_order ASC, configs.capability_code ASC").
		Scan(&references).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list model referenced capabilities")
	}
	return references, nil
}

func modelSummary(
	db *gorm.DB,
	model orderfoodModel.AIModel,
) (orderfoodResponse.AIModelSummary, error) {
	var provider orderfoodModel.AIProvider
	if err := db.First(&provider, "id = ?", model.ProviderID).Error; err != nil {
		return orderfoodResponse.AIModelSummary{}, appErrors.AdminInternal.Wrap(err, "get model provider")
	}
	providerResponse, err := providerSummary(db, provider)
	if err != nil {
		return orderfoodResponse.AIModelSummary{}, err
	}
	capabilities, err := decodeModelCapabilities(model)
	if err != nil {
		return orderfoodResponse.AIModelSummary{}, err
	}
	referencedCapabilityCount, err := modelReferencedCapabilityCount(db, model.ID)
	if err != nil {
		return orderfoodResponse.AIModelSummary{}, err
	}
	usageCount, err := countOptionalTableReference(
		db,
		"of_ai_usages",
		"model_id",
		model.ID,
	)
	if err != nil {
		return orderfoodResponse.AIModelSummary{}, err
	}
	return orderfoodResponse.AIModelSummary{
		ID:                             model.ID,
		Provider:                       providerResponse,
		Name:                           model.Name,
		ModelKey:                       model.ModelKey,
		Capabilities:                   capabilities,
		ContextLength:                  model.ContextLength,
		InputPricePerMillionTokensCNY:  model.InputPricePerMillionTokensCNY,
		OutputPricePerMillionTokensCNY: model.OutputPricePerMillionTokensCNY,
		ImagePricePerUnitCNY:           model.ImagePricePerUnitCNY,
		Enabled:                        model.Enabled,
		EnabledCapabilityCount:         referencedCapabilityCount,
		UsageCount:                     usageCount,
		ReferenceCount:                 referencedCapabilityCount + usageCount,
		Version:                        model.Version,
		UpdatedAt:                      model.UpdatedAt,
	}, nil
}

func modelDetail(
	db *gorm.DB,
	model orderfoodModel.AIModel,
) (orderfoodResponse.AIModelDetail, error) {
	summary, err := modelSummary(db, model)
	if err != nil {
		return orderfoodResponse.AIModelDetail{}, err
	}
	references, err := modelCapabilityReferences(db, model.ID)
	if err != nil {
		return orderfoodResponse.AIModelDetail{}, err
	}
	return orderfoodResponse.AIModelDetail{
		AIModelSummary:         summary,
		ReferencedCapabilities: references,
		Currency:               "CNY",
		Remark:                 model.Remark,
		CreatedAt:              model.CreatedAt,
	}, nil
}

// validateAIModelListQuery 校验模型列表筛选、分页和排序参数。
func validateAIModelListQuery(query orderfoodRequest.AIModelListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.Keyword))) > 60 ||
		len([]rune(strings.TrimSpace(query.ProviderID))) > 64 ||
		(query.SortBy != "createdAt" && query.SortBy != "updatedAt" && query.SortBy != "name") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.Capability != "" {
		if _, err := validateModelCapabilities([]orderfoodModel.AIModelCapability{query.Capability}); err != nil {
			return appErrors.AdminBadRequest.DefaultMsg()
		}
	}
	return nil
}

// validateAIModelIdentity 校验模型公开ID。
func validateAIModelIdentity(modelID string) error {
	length := len([]rune(strings.TrimSpace(modelID)))
	if length < 1 || length > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// validateAIModelCommonInput 校验模型名称、标识、能力、价格和备注。
func validateAIModelCommonInput(input orderfoodRequest.AIModelCreateInput) error {
	if err := validateAIModelIdentity(input.ProviderID); err != nil {
		return err
	}
	nameLength := len([]rune(strings.TrimSpace(input.Name)))
	modelKeyLength := len([]rune(strings.TrimSpace(input.ModelKey)))
	if nameLength < 1 || nameLength > 60 ||
		modelKeyLength < 1 || modelKeyLength > 120 ||
		(input.ContextLength != nil && *input.ContextLength < 1) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, err := validateModelCapabilities(input.Capabilities); err != nil {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	for _, price := range []*string{
		input.InputPricePerMillionTokensCNY,
		input.OutputPricePerMillionTokensCNY,
		input.ImagePricePerUnitCNY,
	} {
		if price != nil && len([]rune(strings.TrimSpace(*price))) > 40 {
			return appErrors.AdminBadRequest.DefaultMsg()
		}
	}
	if input.Remark != nil && len([]rune(strings.TrimSpace(*input.Remark))) > 300 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

func validateModelInput(
	db *gorm.DB,
	input orderfoodRequest.AIModelCreateInput,
	providerMustBeEnabled bool,
) (
	orderfoodModel.AIProvider,
	[]string,
	*string,
	*string,
	*string,
	error,
) {
	var provider orderfoodModel.AIProvider
	if err := db.First(&provider, "id = ?", input.ProviderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return provider, nil, nil, nil, nil, appErrors.AdminNotFound.DefaultMsg()
		}
		return provider, nil, nil, nil, nil, appErrors.AdminInternal.Wrap(err, "get AI model provider")
	}
	if providerMustBeEnabled && !provider.Enabled {
		return provider, nil, nil, nil, nil, appErrors.AdminInvalidConfig.DefaultMsg()
	}
	capabilities, err := validateModelCapabilities(input.Capabilities)
	if err != nil {
		return provider, nil, nil, nil, nil, err
	}
	inputPrice, err := normalizePrice(input.InputPricePerMillionTokensCNY)
	if err != nil {
		return provider, nil, nil, nil, nil, err
	}
	outputPrice, err := normalizePrice(input.OutputPricePerMillionTokensCNY)
	if err != nil {
		return provider, nil, nil, nil, nil, err
	}
	imagePrice, err := normalizePrice(input.ImagePricePerUnitCNY)
	if err != nil {
		return provider, nil, nil, nil, nil, err
	}
	return provider, capabilities, inputPrice, outputPrice, imagePrice, nil
}

// ListAIModels 分页查询AI模型列表。
func (service *AIService) ListAIModels(
	ctx context.Context,
	query orderfoodRequest.AIModelListQuery,
) (orderfoodResponse.Page[orderfoodResponse.AIModelSummary], error) {
	query.ApplyDefaults()
	if err := validateAIModelListQuery(query); err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{}, err
	}
	sortColumns := map[string]string{
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		"name":      "name",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	order := strings.ToLower(query.SortOrder)
	if order != "asc" && order != "desc" {
		return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database().WithContext(ctx)
	statement := db.Model(&orderfoodModel.AIModel{})
	if query.Keyword != "" {
		search := "%" + strings.TrimSpace(query.Keyword) + "%"
		statement = statement.Where("name LIKE ? OR model_key LIKE ?", search, search)
	}
	if query.ProviderID != "" {
		statement = statement.Where("provider_id = ?", query.ProviderID)
	}
	if query.Capability != "" {
		statement = statement.Where(
			"JSON_CONTAINS(capabilities_json, JSON_QUOTE(?))",
			string(query.Capability),
		)
	}
	if query.Enabled != nil {
		statement = statement.Where("enabled = ?", *query.Enabled)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{}, appErrors.AdminInternal.Wrap(err, "count AI models")
	}
	var models []orderfoodModel.AIModel
	if err := statement.Order(column + " " + order).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&models).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{}, appErrors.AdminInternal.Wrap(err, "list AI models")
	}
	list := make([]orderfoodResponse.AIModelSummary, 0, len(models))
	for _, model := range models {
		item, err := modelSummary(db, model)
		if err != nil {
			return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{}, err
		}
		list = append(list, item)
	}
	return orderfoodResponse.Page[orderfoodResponse.AIModelSummary]{
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
		List:     list,
	}, nil
}

// AIModelDetail 获取 AI 模型详情。
func (service *AIService) AIModelDetail(
	ctx context.Context,
	modelID string,
) (orderfoodResponse.AIModelDetail, error) {
	if err := validateAIModelIdentity(modelID); err != nil {
		return orderfoodResponse.AIModelDetail{}, err
	}
	modelID = strings.TrimSpace(modelID)
	db := service.database().WithContext(ctx)
	var model orderfoodModel.AIModel
	if err := db.First(&model, "id = ?", modelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodResponse.AIModelDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return orderfoodResponse.AIModelDetail{}, appErrors.AdminInternal.Wrap(err, "get AI model")
	}
	return modelDetail(db, model)
}

// CreateAIModel 创建AI模型。
func (service *AIService) CreateAIModel(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	idempotencyKey string,
	input orderfoodRequest.AIModelCreateInput,
) (orderfoodResponse.AIModelDetail, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIModelDetail{}, false, err
	}
	if err := validateAIModelCommonInput(input); err != nil {
		return orderfoodResponse.AIModelDetail{}, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_model_create",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			_, capabilities, inputPrice, outputPrice, imagePrice, err := validateModelInput(
				tx,
				input,
				false,
			)
			if err != nil {
				return nil, err
			}
			var duplicate int64
			if err := tx.Model(&orderfoodModel.AIModel{}).
				Where(
					"provider_id = ? AND model_key = ?",
					strings.TrimSpace(input.ProviderID),
					strings.TrimSpace(input.ModelKey),
				).
				Count(&duplicate).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check AI model uniqueness")
			}
			if duplicate > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			capabilitiesJSON, err := encodeJSON(capabilities)
			if err != nil {
				return nil, err
			}
			now := service.now()
			model := orderfoodModel.AIModel{
				ID:                             orderfoodModel.NewID(),
				ProviderID:                     strings.TrimSpace(input.ProviderID),
				Name:                           strings.TrimSpace(input.Name),
				ModelKey:                       strings.TrimSpace(input.ModelKey),
				CapabilitiesJSON:               capabilitiesJSON,
				ContextLength:                  input.ContextLength,
				InputPricePerMillionTokensCNY:  inputPrice,
				OutputPricePerMillionTokensCNY: outputPrice,
				ImagePricePerUnitCNY:           imagePrice,
				Enabled:                        false,
				Remark:                         normalizeOptionalString(input.Remark),
				Version:                        1,
				UpdatedByID:                    actor.AdministratorID,
				UpdatedByUsername:              actor.Username,
				UpdatedByNickname:              actor.Nickname,
				CreatedAt:                      now,
				UpdatedAt:                      now,
			}
			if err := tx.Create(&model).Error; err != nil {
				return nil, appErrors.AdminAlreadyExists.Wrap(err, "create AI model")
			}
			if err := recordAIChange(
				tx,
				now,
				"ai_model",
				model.ID,
				"create",
				model.Version,
				actor,
				nil,
			); err != nil {
				return nil, err
			}
			after, err := modelDetail(tx, model)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"create_ai_model",
				"ai_model",
				model.ID,
				"创建 AI 模型",
				idempotencyKey,
				nil,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.AIModelDetail](raw, replayed, err)
}

// UpdateAIModel 更新AI模型。
func (service *AIService) UpdateAIModel(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	modelID string,
	idempotencyKey string,
	input orderfoodRequest.AIModelUpdateInput,
) (orderfoodResponse.AIModelDetail, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIModelDetail{}, false, err
	}
	if err := validateAIModelIdentity(modelID); err != nil || input.ExpectedVersion < 1 {
		return orderfoodResponse.AIModelDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	commonInput := orderfoodRequest.AIModelCreateInput{
		ProviderID:                     input.ProviderID,
		Name:                           input.Name,
		ModelKey:                       input.ModelKey,
		Capabilities:                   input.Capabilities,
		ContextLength:                  input.ContextLength,
		InputPricePerMillionTokensCNY:  input.InputPricePerMillionTokensCNY,
		OutputPricePerMillionTokensCNY: input.OutputPricePerMillionTokensCNY,
		ImagePricePerUnitCNY:           input.ImagePricePerUnitCNY,
		Remark:                         input.Remark,
	}
	if err := validateAIModelCommonInput(commonInput); err != nil {
		return orderfoodResponse.AIModelDetail{}, false, err
	}
	modelID = strings.TrimSpace(modelID)
	payload := struct {
		ModelID string                              `json:"modelId"`
		Input   orderfoodRequest.AIModelUpdateInput `json:"input"`
	}{ModelID: modelID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_model_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var model orderfoodModel.AIModel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&model, "id = ?", modelID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI model for update")
			}
			if model.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before := model
			_, capabilities, inputPrice, outputPrice, imagePrice, err := validateModelInput(
				tx,
				commonInput,
				model.Enabled,
			)
			if err != nil {
				return nil, err
			}
			var duplicate int64
			if err := tx.Model(&orderfoodModel.AIModel{}).
				Where(
					"provider_id = ? AND model_key = ? AND id <> ?",
					strings.TrimSpace(input.ProviderID),
					strings.TrimSpace(input.ModelKey),
					model.ID,
				).
				Count(&duplicate).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check AI model key")
			}
			if duplicate > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			capabilitiesJSON, err := encodeJSON(capabilities)
			if err != nil {
				return nil, err
			}
			model.ProviderID = strings.TrimSpace(input.ProviderID)
			model.Name = strings.TrimSpace(input.Name)
			model.ModelKey = strings.TrimSpace(input.ModelKey)
			model.CapabilitiesJSON = capabilitiesJSON
			model.ContextLength = input.ContextLength
			model.InputPricePerMillionTokensCNY = inputPrice
			model.OutputPricePerMillionTokensCNY = outputPrice
			model.ImagePricePerUnitCNY = imagePrice
			model.Remark = normalizeOptionalString(input.Remark)
			model.Version++
			model.UpdatedByID = actor.AdministratorID
			model.UpdatedByUsername = actor.Username
			model.UpdatedByNickname = actor.Nickname
			model.UpdatedAt = service.now()
			if err := tx.Save(&model).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update AI model")
			}
			if err := recordAIChange(
				tx,
				model.UpdatedAt,
				"ai_model",
				model.ID,
				"update",
				model.Version,
				actor,
				nil,
			); err != nil {
				return nil, err
			}
			after, err := modelDetail(tx, model)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_ai_model",
				"ai_model",
				model.ID,
				"编辑 AI 模型",
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.AIModelDetail](raw, replayed, err)
}

// UpdateAIModelStatus 更新AI模型状态。
func (service *AIService) UpdateAIModelStatus(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	modelID string,
	idempotencyKey string,
	input orderfoodRequest.AIStatusUpdateInput,
) (orderfoodResponse.AIModelDetail, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.AIModelDetail{}, false, err
	}
	if err := validateAIModelIdentity(modelID); err != nil ||
		input.Enabled == nil ||
		input.ExpectedVersion < 1 {
		return orderfoodResponse.AIModelDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateAIReason(input.Reason); err != nil {
		return orderfoodResponse.AIModelDetail{}, false, err
	}
	modelID = strings.TrimSpace(modelID)
	payload := struct {
		ModelID string                               `json:"modelId"`
		Input   orderfoodRequest.AIStatusUpdateInput `json:"input"`
	}{ModelID: modelID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_model_status_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var model orderfoodModel.AIModel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&model, "id = ?", modelID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI model for status update")
			}
			if model.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if model.Enabled == *input.Enabled {
				after, err := modelDetail(tx, model)
				if err != nil {
					return nil, err
				}
				if err := service.writeMutationAudit(
					ctx,
					tx,
					actor,
					"update_ai_model_status",
					"ai_model",
					model.ID,
					input.Reason,
					idempotencyKey,
					model,
					after,
				); err != nil {
					return nil, err
				}
				return after, nil
			}
			before := model
			if !*input.Enabled {
				count, err := modelReferencedCapabilityCount(tx, model.ID)
				if err != nil {
					return nil, err
				}
				if count > 0 {
					return nil, appErrors.AdminResourceInUse.DefaultMsg()
				}
			} else {
				var provider orderfoodModel.AIProvider
				if err := tx.First(&provider, "id = ?", model.ProviderID).Error; err != nil {
					return nil, appErrors.AdminInvalidConfig.Wrap(err, "get model provider for enable")
				}
				if !provider.Enabled ||
					strings.TrimSpace(provider.APIKey) == "" ||
					provider.LastTestSuccess == nil ||
					!*provider.LastTestSuccess {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			}
			model.Enabled = *input.Enabled
			model.Version++
			model.UpdatedByID = actor.AdministratorID
			model.UpdatedByUsername = actor.Username
			model.UpdatedByNickname = actor.Nickname
			model.UpdatedAt = service.now()
			reason := strings.TrimSpace(input.Reason)
			model.LastChangeReason = &reason
			if err := tx.Save(&model).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update AI model status")
			}
			if err := recordAIChange(
				tx,
				model.UpdatedAt,
				"ai_model",
				model.ID,
				"status",
				model.Version,
				actor,
				&reason,
			); err != nil {
				return nil, err
			}
			after, err := modelDetail(tx, model)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_ai_model_status",
				"ai_model",
				model.ID,
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
	return decodeIdempotentResult[orderfoodResponse.AIModelDetail](raw, replayed, err)
}

func modelReferenceCount(db *gorm.DB, modelID string) (int64, error) {
	var configCount int64
	if err := db.Model(&orderfoodModel.AICapabilityConfig{}).
		Where("primary_model_id = ?", modelID).
		Count(&configCount).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count model configuration references")
	}
	return configCount, nil
}

// DeleteAIModel 删除AI模型。
func (service *AIService) DeleteAIModel(
	ctx context.Context,
	actor orderfoodRequest.AIAdminActor,
	modelID string,
	idempotencyKey string,
	input orderfoodRequest.AIDeleteInput,
) (orderfoodResponse.DeletedResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return orderfoodResponse.DeletedResult{}, false, err
	}
	if err := validateAIModelIdentity(modelID); err != nil || input.ExpectedVersion < 1 {
		return orderfoodResponse.DeletedResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateAIReason(input.Reason); err != nil {
		return orderfoodResponse.DeletedResult{}, false, err
	}
	modelID = strings.TrimSpace(modelID)
	payload := struct {
		ModelID string                         `json:"modelId"`
		Input   orderfoodRequest.AIDeleteInput `json:"input"`
	}{ModelID: modelID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_model_delete",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var model orderfoodModel.AIModel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&model, "id = ?", modelID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "get AI model for delete")
			}
			if model.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			referenceCount, err := modelReferenceCount(tx, model.ID)
			if err != nil {
				return nil, err
			}
			usageCount, err := countOptionalTableReference(
				tx,
				"of_ai_usages",
				"model_id",
				model.ID,
			)
			if err != nil {
				return nil, err
			}
			if referenceCount > 0 || usageCount > 0 {
				return nil, appErrors.AdminResourceInUse.DefaultMsg()
			}
			if err := tx.Delete(&model).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "delete AI model")
			}
			reason := strings.TrimSpace(input.Reason)
			if err := recordAIChange(
				tx,
				service.now(),
				"ai_model",
				model.ID,
				"delete",
				model.Version,
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
				"delete_ai_model",
				"ai_model",
				model.ID,
				input.Reason,
				idempotencyKey,
				model,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.DeletedResult](raw, replayed, err)
}

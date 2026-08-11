package ai

import (
	"context"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	aiRequest "github.com/dyjh/order-food-mini-app/server/model/ai/request"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	"gorm.io/gorm"
)

var (
	promptVariablePattern   = regexp.MustCompile(`\{\{\s*([A-Za-z][A-Za-z0-9_]*)\s*\}\}`)
	promptSensitivePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(api[_-]?key|secret|access[_-]?token|bearer)\s*[:=]\s*\S+`),
		regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`),
		regexp.MustCompile(`\bLTAI[A-Za-z0-9]{12,}\b`),
	}
)

type resolvedPrompt struct {
	Mode                aiModel.AIPromptMode
	PresetVersion       int64
	SystemPrompt        string
	UserPromptTemplate  string
	AllowedVariables    []string
	RequiredVariables   []string
	OutputSchemaVersion string
	ContentHash         string
	LatestPresetVersion int64
}

// externalPromptMode 将内部预设模式转换为管理端冻结的 default/custom 契约。
func externalPromptMode(mode aiModel.AIPromptMode) string {
	if mode == aiModel.AIPromptCustom {
		return "custom"
	}
	return "default"
}

func promptPreset(
	db *gorm.DB,
	capabilityCode string,
) (aiModel.AIPromptDefaultConfig, error) {
	var preset aiModel.AIPromptDefaultConfig
	if err := db.First(&preset, "capability_code = ?", capabilityCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return preset, appErrors.AdminNotFound.DefaultMsg()
		}
		return preset, appErrors.AdminInternal.Wrap(err, "get AI prompt preset")
	}
	return preset, nil
}

func promptPresetVariables(
	preset aiModel.AIPromptDefaultConfig,
) ([]string, []string, error) {
	var allowed []string
	if err := decodeJSON(preset.AllowedVariablesJSON, &allowed); err != nil {
		return nil, nil, err
	}
	var required []string
	if err := decodeJSON(preset.RequiredVariablesJSON, &required); err != nil {
		return nil, nil, err
	}
	sort.Strings(allowed)
	sort.Strings(required)
	return allowed, required, nil
}

// promptVariableMetadata 将固定变量白名单转换为管理端可展示的字段说明。
func promptVariableMetadata(
	allowed []string,
	required []string,
) []aiResponse.AIPromptVariable {
	requiredSet := make(map[string]struct{}, len(required))
	for _, name := range required {
		requiredSet[name] = struct{}{}
	}
	result := make([]aiResponse.AIPromptVariable, 0, len(allowed))
	for _, name := range allowed {
		_, isRequired := requiredSet[name]
		description := "由服务端在实际调用时注入"
		if isRequired {
			description = "必填，由服务端在实际调用时注入"
		}
		result = append(result, aiResponse.AIPromptVariable{
			Name:        name,
			Label:       name,
			Description: description,
			Required:    isRequired,
			Example:     "示例" + name,
		})
	}
	return result
}

// fixedCapabilityOutputSchema 返回由服务端固定、管理员不可修改的能力输出结构。
func fixedCapabilityOutputSchema(
	capabilityCode string,
	version string,
) map[string]interface{} {
	stringArray := map[string]interface{}{
		"type":  "array",
		"items": map[string]interface{}{"type": "string"},
	}
	switch capabilityCode {
	case aiModel.AICapabilityDishTextExtract,
		aiModel.AICapabilityRecipeImageExtract:
		return map[string]interface{}{
			"version": version,
			"type":    "object",
			"required": []string{
				"name", "category", "tags", "serving", "ingredients", "steps",
			},
			"properties": map[string]interface{}{
				"name":        map[string]interface{}{"type": "string"},
				"category":    map[string]interface{}{"type": "string"},
				"tags":        stringArray,
				"serving":     map[string]interface{}{"type": "integer"},
				"description": map[string]interface{}{"type": []string{"string", "null"}},
				"ingredients": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type":     "object",
						"required": []string{"name", "amount", "unit", "sortOrder"},
					},
				},
				"steps": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type":     "object",
						"required": []string{"text", "sortOrder"},
					},
				},
			},
		}
	case aiModel.AICapabilityDishCoverCreate:
		return map[string]interface{}{
			"version":    version,
			"type":       "image",
			"mediaTypes": []string{"image/png", "image/jpeg", "image/webp"},
		}
	case aiModel.AICapabilityCheckinImageAnalyze:
		return map[string]interface{}{
			"version": version,
			"type":    "object",
			"required": []string{
				"dishNames", "ingredients", "tastes", "cuisines",
				"cookingMethods", "dietTypes", "confidence",
			},
			"properties": map[string]interface{}{
				"dishNames":      stringArray,
				"ingredients":    stringArray,
				"tastes":         stringArray,
				"cuisines":       stringArray,
				"cookingMethods": stringArray,
				"dietTypes":      stringArray,
				"confidence": map[string]interface{}{
					"type": "number", "minimum": 0, "maximum": 1,
				},
			},
		}
	case aiModel.AICapabilityMealSuggest:
		return map[string]interface{}{
			"version":  version,
			"type":     "object",
			"required": []string{"reason", "dishes"},
			"properties": map[string]interface{}{
				"reason": map[string]interface{}{"type": "string"},
				"dishes": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"required": []string{
							"name", "cuisine", "category", "tags", "serving",
							"description", "ingredients", "steps",
						},
					},
				},
			},
		}
	case aiModel.AICapabilityPrepSequence:
		return map[string]interface{}{
			"version":  version,
			"type":     "object",
			"required": []string{"estimatedMinutes", "steps"},
			"properties": map[string]interface{}{
				"estimatedMinutes": map[string]interface{}{"type": "integer"},
				"steps": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type":     "object",
						"required": []string{"sourceStepId", "parallelSourceStepIds"},
					},
				},
			},
		}
	default:
		return map[string]interface{}{
			"version": version,
			"type":    "object",
		}
	}
}

func extractPromptVariables(values ...string) ([]string, bool) {
	seen := make(map[string]struct{})
	unmatched := false
	for _, value := range values {
		matches := promptVariablePattern.FindAllStringSubmatch(value, -1)
		for _, match := range matches {
			seen[match[1]] = struct{}{}
		}
		stripped := promptVariablePattern.ReplaceAllString(value, "")
		if strings.Contains(stripped, "{{") || strings.Contains(stripped, "}}") {
			unmatched = true
		}
	}
	result := make([]string, 0, len(seen))
	for variable := range seen {
		result = append(result, variable)
	}
	sort.Strings(result)
	return result, unmatched
}

func stringDifference(left []string, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		rightSet[value] = struct{}{}
	}
	result := make([]string, 0)
	for _, value := range left {
		if _, ok := rightSet[value]; !ok {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func containsPromptSensitiveValue(values ...string) bool {
	for _, value := range values {
		for _, pattern := range promptSensitivePatterns {
			if pattern.MatchString(value) {
				return true
			}
		}
	}
	return false
}

// finalizePromptValidation 汇总冻结契约所需的校验错误和可用内容摘要。
func finalizePromptValidation(result *aiResponse.AIPromptValidationResult) {
	if result == nil {
		return
	}
	result.UnknownVariables = uniqueSortedStrings(result.UnknownVariables)
	result.MissingRequiredVariables = uniqueSortedStrings(result.MissingRequiredVariables)
	result.Errors = make([]string, 0)
	for _, variable := range result.UnknownVariables {
		switch variable {
		case "invalid_template_syntax":
			result.Errors = append(result.Errors, "提示词模板变量格式不完整")
		case "sensitive_value":
			result.Errors = append(result.Errors, "提示词疑似包含密钥或令牌，请清理后重试")
		case "model_unavailable":
			result.Errors = append(result.Errors, "当前主模型或所属供应商不可用")
		case "context_limit":
			result.Errors = append(result.Errors, "提示词长度已达到或超过当前主模型上下文上限")
		default:
			result.Errors = append(result.Errors, "不允许或无效的变量："+variable)
		}
	}
	for _, variable := range result.MissingRequiredVariables {
		if variable == "prompt_body" {
			result.Errors = append(result.Errors, "系统提示词和用户提示词模板均不能为空且不能超过 12000 字")
		} else {
			result.Errors = append(result.Errors, "缺少必填内容或变量："+variable)
		}
	}
	result.Valid = len(result.Errors) == 0
	if result.Valid && result.ContentHash != "" {
		hash := result.ContentHash
		result.PromptHash = &hash
	}
}

func (service *AIService) resolvePromptInput(
	db *gorm.DB,
	capabilityCode string,
	input aiRequest.AIPromptConfigInput,
) (resolvedPrompt, aiResponse.AIPromptValidationResult, error) {
	latest, err := promptPreset(db, capabilityCode)
	if err != nil {
		return resolvedPrompt{}, aiResponse.AIPromptValidationResult{}, err
	}
	mode := input.Mode
	if string(mode) == "default" {
		mode = aiModel.AIPromptPreset
	}
	allowed, required, err := promptPresetVariables(latest)
	if err != nil {
		return resolvedPrompt{}, aiResponse.AIPromptValidationResult{}, err
	}
	result := aiResponse.AIPromptValidationResult{
		AllowedVariables:  allowed,
		RequiredVariables: required,
		Errors:            []string{},
		Warnings:          make([]string, 0),
	}
	resolved := resolvedPrompt{
		Mode:                mode,
		PresetVersion:       latest.Version,
		AllowedVariables:    allowed,
		RequiredVariables:   required,
		OutputSchemaVersion: latest.OutputSchemaVersion,
		LatestPresetVersion: latest.Version,
	}
	switch mode {
	case aiModel.AIPromptPreset:
		resolved.SystemPrompt = latest.SystemPrompt
		resolved.UserPromptTemplate = latest.UserPromptTemplate
	case aiModel.AIPromptCustom:
		if input.SystemPrompt == nil || input.UserPromptTemplate == nil {
			result.MissingRequiredVariables = append(result.MissingRequiredVariables, "prompt_body")
			finalizePromptValidation(&result)
			return resolved, result, nil
		}
		resolved.SystemPrompt = strings.TrimSpace(*input.SystemPrompt)
		resolved.UserPromptTemplate = strings.TrimSpace(*input.UserPromptTemplate)
	default:
		return resolvedPrompt{}, result, appErrors.AdminBadRequest.DefaultMsg()
	}

	if resolved.SystemPrompt == "" || resolved.UserPromptTemplate == "" ||
		utf8.RuneCountInString(resolved.SystemPrompt) > 12000 ||
		utf8.RuneCountInString(resolved.UserPromptTemplate) > 12000 {
		result.MissingRequiredVariables = append(result.MissingRequiredVariables, "prompt_body")
		finalizePromptValidation(&result)
		return resolved, result, nil
	}
	referenced, unmatched := extractPromptVariables(
		resolved.SystemPrompt,
		resolved.UserPromptTemplate,
	)
	result.ReferencedVariables = referenced
	result.UnknownVariables = stringDifference(referenced, allowed)
	result.MissingRequiredVariables = append(
		result.MissingRequiredVariables,
		stringDifference(required, referenced)...,
	)
	if unmatched {
		result.UnknownVariables = append(result.UnknownVariables, "invalid_template_syntax")
	}
	if containsPromptSensitiveValue(resolved.SystemPrompt, resolved.UserPromptTemplate) {
		result.UnknownVariables = append(result.UnknownVariables, "sensitive_value")
	}
	contentHash, err := promptContentHash(
		resolved.SystemPrompt,
		resolved.UserPromptTemplate,
		allowed,
		required,
		resolved.OutputSchemaVersion,
	)
	if err != nil {
		return resolvedPrompt{}, result, err
	}
	resolved.ContentHash = contentHash
	result.ContentHash = contentHash
	finalizePromptValidation(&result)
	return resolved, result, nil
}

func uniqueSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func validatePromptContext(
	model aiModel.AIModel,
	prompt resolvedPrompt,
) error {
	if model.ContextLength == nil {
		return nil
	}
	// A conservative baseline estimate; runtime request variables still go
	// through the model-specific context guard before an actual user call.
	estimatedTokens := (utf8.RuneCountInString(prompt.SystemPrompt) +
		utf8.RuneCountInString(prompt.UserPromptTemplate) + 3) / 4
	if estimatedTokens >= *model.ContextLength {
		return appErrors.AdminInvalidConfig.DefaultMsg()
	}
	return nil
}

func renderPromptTemplate(
	template string,
	allowed []string,
	required []string,
	variables map[string]string,
) (string, error) {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, variable := range allowed {
		allowedSet[variable] = struct{}{}
	}
	for key := range variables {
		if _, ok := allowedSet[key]; !ok {
			return "", appErrors.AdminInvalidConfig.DefaultMsg()
		}
	}
	for _, variable := range required {
		if strings.TrimSpace(variables[variable]) == "" {
			return "", appErrors.AdminInvalidConfig.DefaultMsg()
		}
	}
	rendered := promptVariablePattern.ReplaceAllStringFunc(template, func(match string) string {
		submatches := promptVariablePattern.FindStringSubmatch(match)
		return variables[submatches[1]]
	})
	if strings.Contains(rendered, "{{") || strings.Contains(rendered, "}}") {
		return "", appErrors.AdminInvalidConfig.DefaultMsg()
	}
	return rendered, nil
}

func promptMetadata(
	mode aiModel.AIPromptMode,
	presetVersion int64,
	systemPrompt string,
	userPromptTemplate string,
	allowedJSON []byte,
	requiredJSON []byte,
	outputSchemaVersion string,
	contentHash string,
	latestPresetVersion int64,
	includeBodies bool,
) (aiResponse.AIPromptMetadata, error) {
	var allowed []string
	if err := decodeJSON(allowedJSON, &allowed); err != nil {
		return aiResponse.AIPromptMetadata{}, err
	}
	var required []string
	if err := decodeJSON(requiredJSON, &required); err != nil {
		return aiResponse.AIPromptMetadata{}, err
	}
	result := aiResponse.AIPromptMetadata{
		Mode:                  string(mode),
		PresetVersion:         presetVersion,
		LatestPresetVersion:   latestPresetVersion,
		PresetUpdateAvailable: latestPresetVersion > presetVersion,
		ContentHash:           contentHash,
		AllowedVariables:      allowed,
		RequiredVariables:     required,
		OutputSchemaVersion:   outputSchemaVersion,
		BodyReadable:          includeBodies,
	}
	if includeBodies {
		systemCopy := systemPrompt
		userCopy := userPromptTemplate
		result.SystemPrompt = &systemCopy
		result.UserPromptTemplate = &userCopy
	}
	return result, nil
}

// ValidateAICapabilityPrompt 校验AI能力提示词。
func (service *AIService) ValidateAICapabilityPrompt(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input aiRequest.AIPromptValidationInput,
) (aiResponse.AIPromptValidationResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.AIPromptValidationResult{}, false, err
	}
	if service == nil || service.Permission == nil || service.Idempotency == nil {
		return aiResponse.AIPromptValidationResult{}, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(
		ctx, actor.AuthorityID, PermissionAICapabilityPromptUpdate,
	); err != nil {
		return aiResponse.AIPromptValidationResult{}, false, err
	}

	payload := struct {
		CapabilityCode string                            `json:"capabilityCode"` // AI能力编码
		Input          aiRequest.AIPromptValidationInput `json:"input"`          // 提示词校验输入
	}{CapabilityCode: capabilityCode, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_capability_prompt_validation",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			definition, err := capabilityDefinition(tx, capabilityCode)
			if err != nil {
				return nil, err
			}
			prompt, result, err := service.resolvePromptInput(tx, capabilityCode, input.Prompt)
			if err != nil || !result.Valid {
				return result, err
			}
			current, err := capabilityCurrent(tx, definition)
			if err != nil || current == nil {
				return result, err
			}
			model, _, available, err := availableCapabilityModel(
				tx,
				current.PrimaryModelID,
				definition,
			)
			if err != nil {
				return nil, err
			}
			if !available {
				result.UnknownVariables = append(result.UnknownVariables, "model_unavailable")
			} else if err := validatePromptContext(model, prompt); err != nil {
				result.UnknownVariables = append(result.UnknownVariables, "context_limit")
			}
			finalizePromptValidation(&result)
			return result, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[aiResponse.AIPromptValidationResult](raw, replayed, err)
}

// UpdateAIPrompt 保存并立即应用AI能力提示词。
func (service *AIService) UpdateAIPrompt(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input aiRequest.AIPromptUpdateInput,
) (aiResponse.AIPromptConfig, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.AIPromptConfig{}, false, err
	}
	if service == nil || service.Permission == nil || service.Idempotency == nil {
		return aiResponse.AIPromptConfig{}, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(
		ctx, actor.AuthorityID, PermissionAICapabilityPromptUpdate,
	); err != nil {
		return aiResponse.AIPromptConfig{}, false, err
	}
	reason := strings.TrimSpace(input.Reason)
	if input.ExpectedVersion < 1 || len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return aiResponse.AIPromptConfig{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}

	payload := struct {
		CapabilityCode string                        `json:"capabilityCode"` // AI能力编码
		Input          aiRequest.AIPromptUpdateInput `json:"input"`          // 提示词配置
	}{CapabilityCode: capabilityCode, Input: input}
	mutationContext, cancelMutation := context.WithTimeout(
		ctx,
		aiCapabilityMutationTimeout,
	)
	defer cancelMutation()
	raw, replayed, err := service.Idempotency.Execute(
		mutationContext,
		actor.AdministratorID,
		"ai_capability_prompt_update",
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
			if current == nil {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			if current.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			prompt, validation, err := service.resolvePromptInput(
				tx, capabilityCode, input.Prompt,
			)
			if err != nil {
				return nil, err
			}
			if !validation.Valid {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			model, _, available, err := availableCapabilityModel(
				tx, current.PrimaryModelID, definition,
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
			allowedJSON, err := encodeJSON(prompt.AllowedVariables)
			if err != nil {
				return nil, err
			}
			requiredJSON, err := encodeJSON(prompt.RequiredVariables)
			if err != nil {
				return nil, err
			}
			now := service.now()
			before := map[string]interface{}{
				"version":    current.Version,
				"promptMode": current.PromptMode, "promptHash": current.PromptHash,
			}
			next := *current
			next.Version = current.Version + 1
			next.PromptMode = prompt.Mode
			next.PromptPresetVersion = prompt.PresetVersion
			next.SystemPrompt = prompt.SystemPrompt
			next.UserPromptTemplate = prompt.UserPromptTemplate
			next.PromptAllowedVariablesJSON = allowedJSON
			next.PromptRequiredVariablesJSON = requiredJSON
			next.PromptOutputSchemaVersion = prompt.OutputSchemaVersion
			next.PromptHash = prompt.ContentHash
			next.AppliedByID = actor.AdministratorID
			next.AppliedByUsername = actor.Username
			next.AppliedByNickname = actor.Nickname
			next.AppliedAt = now
			next.Reason = reason
			if err := persistCapabilityConfig(
				tx,
				current,
				next,
				"save AI prompt config",
			); err != nil {
				return nil, err
			}
			if err := recordAIChange(
				tx, now, "ai_capability", capabilityCode, "update_prompt",
				next.Version, actor, &reason,
			); err != nil {
				return nil, err
			}
			after := map[string]interface{}{
				"version":    next.Version,
				"promptMode": prompt.Mode, "promptHash": prompt.ContentHash,
			}
			if err := service.writeMutationAudit(
				mutationContext, tx, actor, "update_ai_capability_prompt",
				"ai_capability", capabilityCode, reason,
				idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			mode := string(prompt.Mode)
			if prompt.Mode == aiModel.AIPromptPreset {
				mode = "default"
			}
			return aiResponse.AIPromptConfig{
				CapabilityCode: capabilityCode, Mode: mode,
				SystemPrompt:       prompt.SystemPrompt,
				UserPromptTemplate: prompt.UserPromptTemplate,
				PromptHash:         prompt.ContentHash, Version: next.Version,
				UpdatedBy: serviceCommon.AdministratorSummary(
					actor.AdministratorID, actor.Username, actor.Nickname,
				),
				UpdatedAt: now,
				Reason:    reason,
			}, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[aiResponse.AIPromptConfig](raw, replayed, err)
}

// promptTestResponse 将测试记录转换为不暴露模型ID和Token细节的管理端响应。
func promptTestResponse(
	db *gorm.DB,
	record aiModel.AIPromptTestRecord,
) (aiResponse.AIPromptTestResult, error) {
	var model aiModel.AIModel
	if err := db.First(&model, "id = ?", record.ModelID).Error; err != nil {
		return aiResponse.AIPromptTestResult{},
			appErrors.AdminInternal.Wrap(err, "load AI prompt test model")
	}
	var provider aiModel.AIProvider
	if err := db.First(&provider, "id = ?", model.ProviderID).Error; err != nil {
		return aiResponse.AIPromptTestResult{},
			appErrors.AdminInternal.Wrap(err, "load AI prompt test provider")
	}
	return aiResponse.AIPromptTestResult{
		ID:               record.ID,
		CapabilityCode:   record.CapabilityCode,
		ConfigVersion:    record.ConfigVersion,
		ModelID:          record.ModelID,
		PromptHash:       record.PromptHash,
		ProviderName:     provider.Name,
		ModelName:        model.Name,
		RequestID:        record.RequestID,
		Success:          record.Success,
		DurationMS:       record.DurationMS,
		InputTokens:      record.InputTokens,
		OutputTokens:     record.OutputTokens,
		EstimatedCostCNY: record.EstimatedCostCNY,
		SafeMessage:      record.SafeMessage,
		SafeSummary:      record.SafeMessage,
		TestedBy: serviceCommon.AdministratorSummary(
			record.AdministratorID,
			record.AdministratorUsername,
			nil,
		),
		TestedAt: record.CreatedAt,
	}, nil
}

// AIPromptWorkspace 获取指定 AI 能力的平台预设和当前提示词配置。
func (service *AIService) AIPromptWorkspace(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
) (aiResponse.AIPromptWorkspace, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.AIPromptWorkspace{}, err
	}
	db := service.database().WithContext(ctx)
	var definition aiModel.AICapabilityDefinition
	if err := db.First(&definition, "code = ?", capabilityCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return aiResponse.AIPromptWorkspace{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return aiResponse.AIPromptWorkspace{}, appErrors.AdminInternal.Wrap(err, "get capability for prompt workspace")
	}
	auditID, err := service.recordPromptAccess(
		ctx,
		actor,
		"read_ai_capability_prompt",
		capabilityCode,
	)
	if err != nil {
		return aiResponse.AIPromptWorkspace{}, err
	}
	latest, err := promptPreset(db, capabilityCode)
	if err != nil {
		return aiResponse.AIPromptWorkspace{}, err
	}
	allowed, required, err := promptPresetVariables(latest)
	if err != nil {
		return aiResponse.AIPromptWorkspace{}, err
	}
	result := aiResponse.AIPromptWorkspace{
		CapabilityCode:            capabilityCode,
		DefaultSystemPrompt:       latest.SystemPrompt,
		DefaultUserPromptTemplate: latest.UserPromptTemplate,
		AllowedVariables:          promptVariableMetadata(allowed, required),
		FixedOutputSchema: fixedCapabilityOutputSchema(
			capabilityCode,
			latest.OutputSchemaVersion,
		),
		EffectivePromptHash: latest.ContentHash,
		AccessAuditID:       auditID,
	}
	var current aiModel.AICapabilityConfig
	if err := db.First(&current, "capability_code = ?", capabilityCode).Error; err == nil {
		result.EffectivePromptHash = current.PromptHash
		result.Current = &aiResponse.AIPromptConfig{
			CapabilityCode:     capabilityCode,
			Mode:               externalPromptMode(current.PromptMode),
			SystemPrompt:       current.SystemPrompt,
			UserPromptTemplate: current.UserPromptTemplate,
			PromptHash:         current.PromptHash,
			Version:            current.Version,
			UpdatedBy: serviceCommon.AdministratorSummary(
				current.AppliedByID,
				current.AppliedByUsername,
				current.AppliedByNickname,
			),
			UpdatedAt: current.AppliedAt,
			Reason:    current.Reason,
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return aiResponse.AIPromptWorkspace{},
			appErrors.AdminInternal.Wrap(err, "get current prompt")
	}
	var testRecord aiModel.AIPromptTestRecord
	if err := db.Where(
		"capability_code = ? AND success = ?",
		capabilityCode,
		true,
	).Order("created_at desc").First(&testRecord).Error; err == nil {
		testResult, responseErr := promptTestResponse(db, testRecord)
		if responseErr != nil {
			return aiResponse.AIPromptWorkspace{}, responseErr
		}
		result.LastSuccessfulTest = &testResult
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return aiResponse.AIPromptWorkspace{},
			appErrors.AdminInternal.Wrap(err, "get last successful AI prompt test")
	}
	return result, nil
}

// RenderAICapabilityPrompt 渲染AI能力提示词。
func (service *AIService) RenderAICapabilityPrompt(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	capabilityCode string,
	idempotencyKey string,
	input aiRequest.AIPromptRenderPreviewInput,
) (aiResponse.AIPromptRenderPreviewResult, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.AIPromptRenderPreviewResult{}, false, err
	}
	auditID, err := service.recordPromptAccess(
		ctx,
		actor,
		"render_ai_capability_prompt",
		capabilityCode,
	)
	if err != nil {
		return aiResponse.AIPromptRenderPreviewResult{}, false, err
	}
	payload := struct {
		CapabilityCode string                               `json:"capabilityCode"` // AI能力编码
		Input          aiRequest.AIPromptRenderPreviewInput `json:"input"`          // 提示词渲染输入
	}{CapabilityCode: capabilityCode, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_capability_prompt_render_preview",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			if _, err := capabilityDefinition(tx, capabilityCode); err != nil {
				return nil, err
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
			return aiResponse.AIPromptRenderPreviewResult{
				CapabilityCode:       capabilityCode,
				PromptHash:           prompt.ContentHash,
				RenderedSystemPrompt: systemPrompt,
				RenderedUserPrompt:   userPrompt,
				AccessAuditID:        auditID,
			}, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[aiResponse.AIPromptRenderPreviewResult](raw, replayed, err)
}

func estimatePromptTestCost(
	model aiModel.AIModel,
	result AIPromptRunResult,
) *string {
	total := new(big.Rat)
	addCost := func(tokens *int, price *string) {
		if tokens == nil || price == nil {
			return
		}
		rate, ok := new(big.Rat).SetString(*price)
		if !ok {
			return
		}
		tokenRatio := new(big.Rat).SetFrac64(int64(*tokens), 1_000_000)
		total.Add(total, new(big.Rat).Mul(rate, tokenRatio))
	}
	addCost(result.InputTokens, model.InputPricePerMillionTokensCNY)
	addCost(result.OutputTokens, model.OutputPricePerMillionTokensCNY)
	if total.Sign() == 0 {
		return nil
	}
	value := strings.TrimRight(strings.TrimRight(total.FloatString(6), "0"), ".")
	return &value
}

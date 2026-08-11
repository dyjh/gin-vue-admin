package ai

import (
	"context"
	"encoding/json"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"sort"
	"strconv"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	aiRequest "github.com/dyjh/order-food-mini-app/server/model/ai/request"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	userService "github.com/dyjh/order-food-mini-app/server/service/user"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// platformFeatureLabelConfig 表示平台策略实际持久化的小程序入口文案。
type platformFeatureLabelConfig struct {
	Code        string `json:"code"`        // 客户端功能编码
	Title       string `json:"title"`       // 标题
	ActionLabel string `json:"actionLabel"` // 操作按钮文案
	Description string `json:"description"` // 说明
	SortOrder   int    `json:"sortOrder"`   // 排序值
}

// validateFeatureLabels 校验四个可交互客户端入口文案及其排序并生成持久化配置。
func validateFeatureLabels(
	input []aiRequest.ClientFeatureLabelInput,
) ([]platformFeatureLabelConfig, error) {
	if len(input) != 4 {
		return nil, appErrors.AdminInvalidConfig.DefaultMsg()
	}
	expected := map[string]bool{
		"dish_extract":  false,
		"cover_create":  false,
		"meal_suggest":  false,
		"prep_sequence": false,
	}
	sortOrders := make(map[int]struct{}, len(input))
	result := make([]platformFeatureLabelConfig, 0, len(input))
	for _, label := range input {
		if _, ok := expected[label.Code]; !ok || expected[label.Code] {
			return nil, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		if _, exists := sortOrders[label.SortOrder]; exists {
			return nil, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		if label.SortOrder < 1 || label.SortOrder > 4 ||
			strings.TrimSpace(label.Title) == "" ||
			strings.TrimSpace(label.ActionLabel) == "" ||
			strings.TrimSpace(label.Description) == "" ||
			len([]rune(strings.TrimSpace(label.Title))) > 40 ||
			len([]rune(strings.TrimSpace(label.ActionLabel))) > 20 ||
			len([]rune(strings.TrimSpace(label.Description))) > 160 {
			return nil, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		expected[label.Code] = true
		sortOrders[label.SortOrder] = struct{}{}
		result = append(result, platformFeatureLabelConfig{
			Code:        label.Code,
			Title:       strings.TrimSpace(label.Title),
			ActionLabel: strings.TrimSpace(label.ActionLabel),
			Description: strings.TrimSpace(label.Description),
			SortOrder:   label.SortOrder,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SortOrder < result[j].SortOrder
	})
	return result, nil
}

// capabilityBillingValue 表示单项后端能力的当前计费值。
type capabilityBillingValue struct {
	name         string // 能力名称
	pointCost    int    // 单次积分成本
	freeQuotaDay int    // 每日免费次数
}

// featureBillingHints 表示一个客户端入口实时派生的计费提示。
type featureBillingHints struct {
	costHint      *string // 积分成本提示
	freeQuotaHint *string // 每日免费次数提示
}

// formatCapabilityCost 将单项能力积分成本转换为管理端和小程序可直接展示的文本。
func formatCapabilityCost(pointCost int) string {
	if pointCost <= 0 {
		return "免费"
	}
	return "每次消耗 " + strconv.Itoa(pointCost) + " 积分"
}

// formatCapabilityFreeQuota 将单项能力每日免费额度转换为可直接展示的文本。
func formatCapabilityFreeQuota(freeQuotaDay int) string {
	if freeQuotaDay <= 0 {
		return "无每日免费次数"
	}
	return "每天免费 " + strconv.Itoa(freeQuotaDay) + " 次"
}

// formatFeatureBillingHints 根据一个客户端入口关联的实际能力配置生成费用和免费额度提示。
func formatFeatureBillingHints(values []capabilityBillingValue) featureBillingHints {
	if len(values) == 0 {
		return featureBillingHints{}
	}
	costParts := make([]string, 0, len(values))
	freeParts := make([]string, 0, len(values))
	for _, value := range values {
		costText := formatCapabilityCost(value.pointCost)
		freeText := formatCapabilityFreeQuota(value.freeQuotaDay)
		if len(values) > 1 {
			costText = value.name + "：" + costText
			freeText = value.name + "：" + freeText
		}
		costParts = append(costParts, costText)
		freeParts = append(freeParts, freeText)
	}
	costHint := strings.Join(costParts, "；")
	freeQuotaHint := strings.Join(freeParts, "；")
	return featureBillingHints{
		costHint:      &costHint,
		freeQuotaHint: &freeQuotaHint,
	}
}

// loadFeatureBillingHints 按客户端入口聚合六项当前能力配置，避免平台策略重复保存计费文案。
func loadFeatureBillingHints(
	db *gorm.DB,
) (map[string]featureBillingHints, error) {
	var definitions []aiModel.AICapabilityDefinition
	if err := db.Order("sort_order asc").Find(&definitions).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list capability billing definitions")
	}
	var configs []aiModel.AICapabilityConfig
	if err := db.Find(&configs).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list current capability billing configs")
	}
	configByCode := make(map[string]aiModel.AICapabilityConfig, len(configs))
	for _, config := range configs {
		configByCode[config.CapabilityCode] = config
	}
	valuesByFeature := make(map[string][]capabilityBillingValue)
	for _, definition := range definitions {
		config, exists := configByCode[definition.Code]
		if !exists {
			continue
		}
		valuesByFeature[definition.ClientFeatureCode] = append(
			valuesByFeature[definition.ClientFeatureCode],
			capabilityBillingValue{
				name:         definition.Name,
				pointCost:    config.PointCost,
				freeQuotaDay: config.FreeQuotaPerDay,
			},
		)
	}
	result := make(map[string]featureBillingHints, len(valuesByFeature))
	for featureCode, values := range valuesByFeature {
		result[featureCode] = formatFeatureBillingHints(values)
	}
	return result, nil
}

func latestPlatformPolicy(
	db *gorm.DB,
) (aiModel.PlatformCapabilityPolicy, error) {
	var current aiModel.PlatformCapabilityPolicy
	if err := db.Where(
		"singleton_key = ?",
		serviceCommon.PlatformPolicySingletonKey,
	).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return current, appErrors.AdminNotFound.DefaultMsg()
		}
		return current, appErrors.AdminInternal.Wrap(err, "get current platform policy")
	}
	return current, nil
}

// Snapshot implements CapabilityPolicyProvider for the user administration service.
func (service *AIService) Snapshot(
	ctx context.Context,
) (userService.CapabilityPolicySnapshot, error) {
	db := service.database()
	if db == nil {
		return userService.CapabilityPolicySnapshot{PlatformDefaultEnabled: false}, nil
	}
	var current aiModel.PlatformCapabilityPolicy
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", serviceCommon.PlatformPolicySingletonKey).
		First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return userService.CapabilityPolicySnapshot{PlatformDefaultEnabled: false}, nil
		}
		return userService.CapabilityPolicySnapshot{}, appErrors.AdminInternal.Wrap(
			err,
			"read effective platform capability policy",
		)
	}
	return userService.CapabilityPolicySnapshot{
		PlatformDefaultEnabled: current.PlatformDefaultEnabled,
		EmergencyDisabled:      current.EmergencyDisabled,
	}, nil
}

// platformConfigResponse 组装当前平台策略，并实时附加能力计费提示。
func platformConfigResponse(
	db *gorm.DB,
	model aiModel.PlatformCapabilityPolicy,
) (aiResponse.PlatformPolicyConfig, error) {
	var labels []aiResponse.ClientFeatureLabel
	if err := decodeJSON(model.FeatureLabelsJSON, &labels); err != nil {
		return aiResponse.PlatformPolicyConfig{}, err
	}
	// 打卡智能分析是保存打卡后自动运行的后台能力，不是客户端可点击入口。
	visibleLabels := make([]aiResponse.ClientFeatureLabel, 0, 4)
	for _, label := range labels {
		if label.Code == "taste_profile" || label.Code == "checkin_image_analyze" {
			continue
		}
		visibleLabels = append(visibleLabels, label)
	}
	labels = visibleLabels
	billingHints, err := loadFeatureBillingHints(db)
	if err != nil {
		return aiResponse.PlatformPolicyConfig{}, err
	}
	for index := range labels {
		hints := billingHints[labels[index].Code]
		labels[index].CostHint = hints.costHint
		labels[index].FreeQuotaHint = hints.freeQuotaHint
	}
	return aiResponse.PlatformPolicyConfig{
		PlatformDefaultEnabled: model.PlatformDefaultEnabled,
		EmergencyDisabled:      model.EmergencyDisabled,
		FeatureLabels:          labels,
		PolicyVersion:          model.Version,
		UpdatedBy: serviceCommon.AdministratorSummary(
			model.AppliedByID,
			model.AppliedByUsername,
			model.AppliedByNickname,
		),
		UpdatedAt: model.AppliedAt,
		Reason:    model.Reason,
	}, nil
}

// loadNormalUserCount 统计当前可正常使用小程序的用户数量。
func loadNormalUserCount(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Model(&userModel.MiniAppUser{}).
		Where("status = ?", userModel.UserStatusNormal).
		Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count normal platform users")
	}
	return count, nil
}

// loadIndividuallyDisabledUserCount 统计正常用户中被单独关闭能力的数量。
func loadIndividuallyDisabledUserCount(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Model(&userModel.MiniAppUser{}).
		Where("status = ? AND capability_disabled = ?", userModel.UserStatusNormal, true).
		Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count individually disabled platform users")
	}
	return count, nil
}

// effectiveEnabledCount 根据总开关、紧急状态和用户单独关闭状态计算实际开放数量。
func effectiveEnabledCount(
	defaultEnabled bool,
	emergencyDisabled bool,
	normalUserCount int64,
	individuallyDisabledUserCount int64,
) int64 {
	if emergencyDisabled || !defaultEnabled {
		return 0
	}
	return normalUserCount - individuallyDisabledUserCount
}

// platformUserSnapshot 组装平台整体能力影响用户快照。
func platformUserSnapshot(
	policy aiModel.PlatformCapabilityPolicy,
	normalUserCount int64,
	individuallyDisabledUserCount int64,
) aiResponse.PlatformCapabilitySnapshot {
	return aiResponse.PlatformCapabilitySnapshot{
		PlatformDefaultEnabled:        policy.PlatformDefaultEnabled,
		EmergencyDisabled:             policy.EmergencyDisabled,
		NormalUserCount:               normalUserCount,
		IndividuallyDisabledUserCount: individuallyDisabledUserCount,
		EffectiveEnabledUserCount: effectiveEnabledCount(
			policy.PlatformDefaultEnabled,
			policy.EmergencyDisabled,
			normalUserCount,
			individuallyDisabledUserCount,
		),
		PolicyVersion: policy.Version,
	}
}

// PlatformPolicyWorkspace 获取当前平台总开关策略和能力就绪状态。
func (service *AIService) PlatformPolicyWorkspace(
	ctx context.Context,
) (aiResponse.PlatformPolicyWorkspace, error) {
	db := service.database().WithContext(ctx)
	current, err := latestPlatformPolicy(db)
	if err != nil {
		return aiResponse.PlatformPolicyWorkspace{}, err
	}
	config, err := platformConfigResponse(db, current)
	if err != nil {
		return aiResponse.PlatformPolicyWorkspace{}, err
	}
	normalUserCount, err := loadNormalUserCount(db)
	if err != nil {
		return aiResponse.PlatformPolicyWorkspace{}, err
	}
	individuallyDisabledUserCount, err := loadIndividuallyDisabledUserCount(db)
	if err != nil {
		return aiResponse.PlatformPolicyWorkspace{}, err
	}
	readiness, err := aiCapabilitiesReadiness(db)
	if err != nil {
		return aiResponse.PlatformPolicyWorkspace{}, err
	}
	return aiResponse.PlatformPolicyWorkspace{
		Config:     config,
		UserCounts: platformUserSnapshot(current, normalUserCount, individuallyDisabledUserCount),
		Readiness:  readiness,
	}, nil
}

// UpdatePlatformPolicy 保存并立即应用平台总开关策略。
func (service *AIService) UpdatePlatformPolicy(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	idempotencyKey string,
	input aiRequest.PlatformPolicyUpdateInput,
) (aiResponse.PlatformPolicyConfig, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.PlatformPolicyConfig{}, false, err
	}
	reason := strings.TrimSpace(input.Reason)
	if input.ExpectedVersion < 1 || len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return aiResponse.PlatformPolicyConfig{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	labels, err := validateFeatureLabels(input.FeatureLabels)
	if err != nil {
		return aiResponse.PlatformPolicyConfig{}, false, err
	}

	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"platform_policy_update",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			current, currentErr := latestPlatformPolicy(
				tx,
			)
			if currentErr != nil {
				return nil, currentErr
			}
			if current.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if input.PlatformDefaultEnabled {
				readiness, readinessErr := aiCapabilitiesReadiness(tx)
				if readinessErr != nil {
					return nil, readinessErr
				}
				if !readiness.AllReady {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			}
			labelsJSON, marshalErr := json.Marshal(labels)
			if marshalErr != nil {
				return nil, marshalErr
			}
			now := service.now()
			next := aiModel.PlatformCapabilityPolicy{
				SingletonKey:           serviceCommon.PlatformPolicySingletonKey,
				Version:                current.Version + 1,
				PlatformDefaultEnabled: input.PlatformDefaultEnabled,
				EmergencyDisabled:      current.EmergencyDisabled,
				FeatureLabelsJSON:      datatypes.JSON(labelsJSON),
				AppliedByID:            actor.AdministratorID,
				AppliedByUsername:      actor.Username,
				AppliedByNickname:      actor.Nickname,
				AppliedAt:              now,
				Reason:                 reason,
			}
			update := tx.Model(&aiModel.PlatformCapabilityPolicy{}).
				Where("singleton_key = ? AND version = ?", serviceCommon.PlatformPolicySingletonKey, input.ExpectedVersion).
				Select("*").Omit("singleton_key").Updates(&next)
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "save platform policy")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if err := recordAIChange(
				tx,
				now,
				"platform_policy",
				serviceCommon.PlatformPolicySingletonKey,
				"update",
				next.Version,
				actor,
				&reason,
			); err != nil {
				return nil, err
			}
			before, err := platformConfigResponse(tx, current)
			if err != nil {
				return nil, err
			}
			after, err := platformConfigResponse(tx, next)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_platform_policy",
				"platform_policy",
				serviceCommon.PlatformPolicySingletonKey,
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
	return serviceCommon.DecodeIdempotentResult[aiResponse.PlatformPolicyConfig](raw, replayed, err)
}

// SetPlatformEmergencyStatus 设置全平台 AI 紧急停用状态。
func (service *AIService) SetPlatformEmergencyStatus(
	ctx context.Context,
	actor aiRequest.AIAdminActor,
	idempotencyKey string,
	input aiRequest.PlatformPolicyEmergencyInput,
) (aiResponse.PlatformPolicyConfig, bool, error) {
	if err := validateAIActor(actor); err != nil {
		return aiResponse.PlatformPolicyConfig{}, false, err
	}
	if input.EmergencyDisabled == nil {
		return aiResponse.PlatformPolicyConfig{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateAIReason(input.Reason); err != nil {
		return aiResponse.PlatformPolicyConfig{}, false, err
	}

	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"platform_policy_emergency_status_update",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			current, currentErr := latestPlatformPolicy(
				tx,
			)
			if currentErr != nil {
				return nil, currentErr
			}
			if current.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if current.EmergencyDisabled == *input.EmergencyDisabled {
				return platformConfigResponse(tx, current)
			}
			if !*input.EmergencyDisabled && current.PlatformDefaultEnabled {
				readiness, readinessErr := aiCapabilitiesReadiness(tx)
				if readinessErr != nil {
					return nil, readinessErr
				}
				if !readiness.AllReady {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			}
			now := service.now()
			reason := strings.TrimSpace(input.Reason)
			next := aiModel.PlatformCapabilityPolicy{
				SingletonKey:           serviceCommon.PlatformPolicySingletonKey,
				Version:                current.Version + 1,
				PlatformDefaultEnabled: current.PlatformDefaultEnabled,
				EmergencyDisabled:      *input.EmergencyDisabled,
				FeatureLabelsJSON: datatypes.JSON(
					append([]byte(nil), current.FeatureLabelsJSON...),
				),
				AppliedByID:       actor.AdministratorID,
				AppliedByUsername: actor.Username,
				AppliedByNickname: actor.Nickname,
				AppliedAt:         now,
				Reason:            reason,
			}
			update := tx.Model(&aiModel.PlatformCapabilityPolicy{}).
				Where("singleton_key = ? AND version = ?", serviceCommon.PlatformPolicySingletonKey, input.ExpectedVersion).
				Select("*").Omit("singleton_key").Updates(&next)
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "change platform emergency status")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if err := recordAIChange(
				tx,
				now,
				"platform_policy",
				serviceCommon.PlatformPolicySingletonKey,
				"emergency_status",
				next.Version,
				actor,
				&reason,
			); err != nil {
				return nil, err
			}
			before, err := platformConfigResponse(tx, current)
			if err != nil {
				return nil, err
			}
			after, err := platformConfigResponse(tx, next)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_platform_emergency_status",
				"platform_policy",
				serviceCommon.PlatformPolicySingletonKey,
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
	return serviceCommon.DecodeIdempotentResult[aiResponse.PlatformPolicyConfig](raw, replayed, err)
}

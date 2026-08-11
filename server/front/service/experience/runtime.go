package experience

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/global"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/gorm"
)

// RuntimeService 提供小程序运行时配置业务能力。
type RuntimeService struct {
	DB *gorm.DB // 业务数据库
}

func (service *RuntimeService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// clientFeatureLabel 表示平台策略持久化的小程序入口文案。
type clientFeatureLabel struct {
	Code        string `json:"code"`        // 客户端功能编码
	Title       string `json:"title"`       // 标题
	ActionLabel string `json:"actionLabel"` // 操作按钮文案
	Description string `json:"description"` // 说明
	SortOrder   int    `json:"sortOrder"`   // 排序值
}

// featureCost 表示一个客户端入口聚合后的当前计费信息。
type featureCost struct {
	pointCost     *int                  // 单一积分成本；混合成本时不对外返回
	freeQuota     *int                  // 单一免费额度；混合额度时不对外返回
	costHint      *string               // 实时积分提示
	freeQuotaHint *string               // 实时免费额度提示
	values        []featureBillingValue // 入口关联的后端能力计费值
	mixedCost     bool                  // 是否包含不同积分成本
	mixedFree     bool                  // 是否包含不同免费额度
}

// featureBillingValue 表示单项后端能力的当前计费值。
type featureBillingValue struct {
	name         string // 能力名称
	pointCost    int    // 单次积分成本
	freeQuotaDay int    // 每日免费次数
}

// mealFinalResultSubscriptionConfig 查询当前启用的饭局结果订阅模板。
func (service *RuntimeService) mealFinalResultSubscriptionConfig(
	ctx context.Context,
) (frontResponse.MealFinalResultSubscriptionConfig, error) {
	result := frontResponse.MealFinalResultSubscriptionConfig{
		ButtonText: "接收最终结果提醒",
	}
	var template engagementModel.SubscribeMessageTemplate
	err := service.database().WithContext(ctx).
		Where("scene = ? AND enabled = ?", engagementModel.SubscribeSceneMealStatus, true).
		Order("updated_at desc, id desc").
		First(&template).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
	}
	if err != nil {
		return result, appErrors.FrontInternal.Wrap(err, "load meal result subscription template")
	}
	templateID := template.WechatTemplateID
	result.Enabled = true
	result.TemplateID = &templateID
	return result, nil
}

// Current 获取当前小程序运行时配置。
func (service *RuntimeService) Current(
	ctx context.Context,
	userIDs ...string,
) (frontResponse.RuntimeConfig, error) {
	db := service.database()
	if db == nil {
		return frontResponse.RuntimeConfig{}, appErrors.FrontInternal.DefaultMsg()
	}
	subscriptionConfig, err := service.mealFinalResultSubscriptionConfig(ctx)
	if err != nil {
		return frontResponse.RuntimeConfig{}, err
	}
	var policy aiModel.PlatformCapabilityPolicy
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", "platform").
		First(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.RuntimeConfig{
				Features:                    []frontResponse.ClientFeature{},
				MealFinalResultSubscription: subscriptionConfig,
			}, nil
		}
		return frontResponse.RuntimeConfig{}, appErrors.FrontInternal.Wrap(err, "load miniapp runtime policy")
	}
	userDisabled := false
	if len(userIDs) > 0 && strings.TrimSpace(userIDs[0]) != "" {
		var user userModel.MiniAppUser
		if err := db.WithContext(ctx).
			Select("capability_disabled").
			First(&user, "id = ?", strings.TrimSpace(userIDs[0])).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return frontResponse.RuntimeConfig{}, appErrors.FrontInternal.Wrap(err, "load user capability override")
			}
		} else {
			userDisabled = user.CapabilityDisabled
		}
	}
	enabled := policy.PlatformDefaultEnabled && !policy.EmergencyDisabled && !userDisabled
	result := frontResponse.RuntimeConfig{
		PolicyVersion:               policy.Version,
		EnhancedFeaturesEnabled:     enabled,
		PointsEnabled:               enabled,
		Features:                    []frontResponse.ClientFeature{},
		MealFinalResultSubscription: subscriptionConfig,
	}
	if !enabled {
		return result, nil
	}

	var labels []clientFeatureLabel
	if err := json.Unmarshal(policy.FeatureLabelsJSON, &labels); err != nil {
		return frontResponse.RuntimeConfig{}, appErrors.FrontInternal.Wrap(err, "decode miniapp feature labels")
	}
	costs, err := service.featureCosts(ctx, db)
	if err != nil {
		return frontResponse.RuntimeConfig{}, err
	}
	for _, label := range labels {
		if label.Code == "taste_profile" || label.Code == "checkin_image_analyze" {
			continue
		}
		cost := costs[label.Code]
		var pointCost *int
		var freeQuota *int
		var costHint *string
		var freeQuotaHint *string
		if cost != nil && !cost.mixedCost {
			pointCost = cost.pointCost
		}
		if cost != nil && !cost.mixedFree {
			freeQuota = cost.freeQuota
		}
		if cost != nil {
			costHint = cost.costHint
			freeQuotaHint = cost.freeQuotaHint
		}
		result.Features = append(result.Features, frontResponse.ClientFeature{
			Code:          label.Code,
			Title:         label.Title,
			ActionText:    label.ActionLabel,
			Description:   label.Description,
			Enabled:       true,
			PointCost:     pointCost,
			FreeQuota:     freeQuota,
			CostHint:      costHint,
			FreeQuotaHint: freeQuotaHint,
			SortOrder:     label.SortOrder,
		})
	}
	sort.SliceStable(result.Features, func(left, right int) bool {
		return result.Features[left].SortOrder < result.Features[right].SortOrder
	})
	return result, nil
}

// runtimeCapabilityCostText 将能力积分成本转换为小程序可直接展示的文本。
func runtimeCapabilityCostText(pointCost int) string {
	if pointCost <= 0 {
		return "免费"
	}
	return "每次消耗 " + strconv.Itoa(pointCost) + " 积分"
}

// runtimeCapabilityFreeQuotaText 将能力每日免费额度转换为小程序可直接展示的文本。
func runtimeCapabilityFreeQuotaText(freeQuotaDay int) string {
	if freeQuotaDay <= 0 {
		return "无每日免费次数"
	}
	return "每天免费 " + strconv.Itoa(freeQuotaDay) + " 次"
}

// applyRuntimeBillingHints 根据入口实际关联的能力配置生成实时计费提示。
func applyRuntimeBillingHints(cost *featureCost) {
	if cost == nil || len(cost.values) == 0 {
		return
	}
	costParts := make([]string, 0, len(cost.values))
	freeParts := make([]string, 0, len(cost.values))
	for _, value := range cost.values {
		costText := runtimeCapabilityCostText(value.pointCost)
		freeText := runtimeCapabilityFreeQuotaText(value.freeQuotaDay)
		if len(cost.values) > 1 {
			costText = value.name + "：" + costText
			freeText = value.name + "：" + freeText
		}
		costParts = append(costParts, costText)
		freeParts = append(freeParts, freeText)
	}
	costHint := strings.Join(costParts, "；")
	freeQuotaHint := strings.Join(freeParts, "；")
	cost.costHint = &costHint
	cost.freeQuotaHint = &freeQuotaHint
}

func (service *RuntimeService) featureCosts(
	ctx context.Context,
	db *gorm.DB,
) (map[string]*featureCost, error) {
	var definitions []aiModel.AICapabilityDefinition
	if err := db.WithContext(ctx).
		Find(&definitions).Error; err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "load miniapp capabilities")
	}
	result := make(map[string]*featureCost, len(definitions))
	for _, definition := range definitions {
		if definition.Code == aiModel.AICapabilityCheckinImageAnalyze ||
			definition.Code == aiModel.AICapabilityPreferenceSummarize {
			continue
		}
		var version aiModel.AICapabilityConfig
		if err := db.WithContext(ctx).
			First(&version, "capability_code = ?", definition.Code).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, appErrors.FrontInternal.Wrap(err, "load current miniapp capability")
		}
		entry := result[definition.ClientFeatureCode]
		if entry == nil {
			pointCost := version.PointCost
			freeQuota := version.FreeQuotaPerDay
			result[definition.ClientFeatureCode] = &featureCost{
				pointCost: &pointCost,
				freeQuota: &freeQuota,
				values: []featureBillingValue{{
					name:         definition.Name,
					pointCost:    version.PointCost,
					freeQuotaDay: version.FreeQuotaPerDay,
				}},
			}
			continue
		}
		entry.values = append(entry.values, featureBillingValue{
			name:         definition.Name,
			pointCost:    version.PointCost,
			freeQuotaDay: version.FreeQuotaPerDay,
		})
		if entry.pointCost == nil || *entry.pointCost != version.PointCost {
			entry.mixedCost = true
		}
		if entry.freeQuota == nil || *entry.freeQuota != version.FreeQuotaPerDay {
			entry.mixedFree = true
		}
	}
	for _, cost := range result {
		applyRuntimeBillingHints(cost)
	}
	return result, nil
}

// ProfileForRuntime 根据运行时开关组装可返回的个人资料。
func ProfileForRuntime(
	user userModel.MiniAppUser,
	runtime frontResponse.RuntimeConfig,
) frontResponse.Profile {
	var points *int64
	if runtime.PointsEnabled {
		value := user.Points
		points = &value
	}
	return frontResponse.Profile{
		ID:          user.ID,
		Nickname:    user.Nickname,
		AvatarURL:   user.AvatarURL,
		Points:      points,
		CheckinDays: user.CheckinDayCount,
	}
}

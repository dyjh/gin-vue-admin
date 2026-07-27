package service

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const preferenceEvidenceMaxAttempts = 3

// preferenceSetting 表示用户明确填写的一条偏好或硬条件。
type preferenceSetting struct {
	Type      string    `json:"type"`      // 设置类型
	Value     string    `json:"value"`     // 设置内容
	UpdatedAt time.Time `json:"updatedAt"` // 设置时间
}

// preferenceFacts 表示可参与规则聚合的结构化偏好事实。
type preferenceFacts struct {
	DishNames      []string            `json:"dishNames,omitempty"`      // 菜品名称候选
	Tags           []string            `json:"tags,omitempty"`           // 菜品标签
	Ingredients    []string            `json:"ingredients,omitempty"`    // 食材名称
	Tastes         []string            `json:"tastes,omitempty"`         // 口味事实
	Cuisines       []string            `json:"cuisines,omitempty"`       // 菜系事实
	CookingMethods []string            `json:"cookingMethods,omitempty"` // 烹饪方式
	DietTypes      []string            `json:"dietTypes,omitempty"`      // 荤素等饮食类型
	Settings       []preferenceSetting `json:"settings,omitempty"`       // 用户明确设置
	ImageURL       string              `json:"imageUrl,omitempty"`       // 待分析打卡图片地址
	Caption        string              `json:"caption,omitempty"`        // 待分析用户说明
}

// preferenceAnalysisOutput 表示打卡图片分析能力的固定结构化输出。
type preferenceAnalysisOutput struct {
	DishNames      []string `json:"dishNames"`      // 图片中的菜品名称候选
	Ingredients    []string `json:"ingredients"`    // 图片中可见食材
	Tastes         []string `json:"tastes"`         // 可见或明确说明的口味
	Cuisines       []string `json:"cuisines"`       // 可确认的菜系
	CookingMethods []string `json:"cookingMethods"` // 可确认的烹饪方式
	DietTypes      []string `json:"dietTypes"`      // 荤素等饮食类型
	Confidence     float64  `json:"confidence"`     // 整体置信度
}

// PreferenceService 提供偏好证据采集、图片分析和规则画像聚合能力。
type PreferenceService struct {
	DB      *gorm.DB         // 业务数据库
	Assist  *AssistService   // 打卡图片分析能力
	Runtime *RuntimeService  // 平台整体能力运行时
	Now     func() time.Time // 可注入的当前时间
}

// database 返回当前服务使用的数据库。
func (service *PreferenceService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return ServiceGroupApp.ContentService.database()
}

// assist 返回注入的增强功能服务。
func (service *PreferenceService) assist() *AssistService {
	if service != nil && service.Assist != nil {
		return service.Assist
	}
	return &ServiceGroupApp.AssistService
}

// runtime 返回注入的平台运行时服务。
func (service *PreferenceService) runtime() *RuntimeService {
	if service != nil && service.Runtime != nil {
		return service.Runtime
	}
	return &ServiceGroupApp.RuntimeService
}

// now 返回统一使用的UTC当前时间。
func (service *PreferenceService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// preferenceUpdatesEnabled 判断当前是否允许采集和处理新的偏好证据。
func preferenceUpdatesEnabled(ctx context.Context) bool {
	runtime, err := ServiceGroupApp.RuntimeService.Current(ctx)
	return err == nil && runtime.EnhancedFeaturesEnabled
}

// normalizePreferenceTerms 清理、去重并限制一组结构化词条。
func normalizePreferenceTerms(values []string, limit int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// normalizePreferenceFacts 清理画像聚合前的结构化事实。
func normalizePreferenceFacts(facts preferenceFacts) preferenceFacts {
	facts.DishNames = normalizePreferenceTerms(facts.DishNames, 12)
	facts.Tags = normalizePreferenceTerms(facts.Tags, 20)
	facts.Ingredients = normalizePreferenceTerms(facts.Ingredients, 30)
	facts.Tastes = normalizePreferenceTerms(facts.Tastes, 12)
	facts.Cuisines = normalizePreferenceTerms(facts.Cuisines, 12)
	facts.CookingMethods = normalizePreferenceTerms(facts.CookingMethods, 12)
	facts.DietTypes = normalizePreferenceTerms(facts.DietTypes, 8)
	normalizedSettings := make([]preferenceSetting, 0, len(facts.Settings))
	for _, setting := range facts.Settings {
		setting.Type = strings.TrimSpace(setting.Type)
		setting.Value = strings.TrimSpace(setting.Value)
		if setting.Type == "" || setting.Value == "" {
			continue
		}
		normalizedSettings = append(normalizedSettings, setting)
	}
	facts.Settings = normalizedSettings
	facts.ImageURL = strings.TrimSpace(facts.ImageURL)
	facts.Caption = strings.TrimSpace(facts.Caption)
	return facts
}

// refreshEvidenceAggregateTx 刷新一个用户单一来源类型的管理端计数摘要。
func refreshEvidenceAggregateTx(
	tx *gorm.DB,
	userID string,
	sourceType string,
	now time.Time,
) error {
	var totalCount int64
	var aggregatedCount int64
	var pendingCount int64
	if err := tx.Model(&orderfoodModel.PreferenceEvidence{}).
		Where("user_id = ? AND source_type = ?", userID, sourceType).
		Count(&totalCount).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "count preference evidence")
	}
	if err := tx.Model(&orderfoodModel.PreferenceEvidence{}).
		Where("user_id = ? AND source_type = ? AND status = ?",
			userID, sourceType, orderfoodModel.PreferenceEvidenceAggregated).
		Count(&aggregatedCount).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "count aggregated preference evidence")
	}
	if err := tx.Model(&orderfoodModel.PreferenceEvidence{}).
		Where("user_id = ? AND source_type = ? AND status = ?",
			userID, sourceType, orderfoodModel.PreferenceEvidencePending).
		Count(&pendingCount).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "count pending preference evidence")
	}
	var latest *time.Time
	var latestValue time.Time
	if err := tx.Model(&orderfoodModel.PreferenceEvidence{}).
		Where("user_id = ? AND source_type = ?", userID, sourceType).
		Select("MAX(occurred_at)").Scan(&latestValue).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "load latest preference evidence time")
	}
	if !latestValue.IsZero() {
		latest = &latestValue
	}
	row := orderfoodModel.PreferenceEvidenceAggregate{
		UserID: userID, SourceType: sourceType,
		TotalCount: int(totalCount), AggregatedCount: int(aggregatedCount),
		PendingCount: int(pendingCount), LatestOccurredAt: latest,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "source_type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_count": totalCount, "aggregated_count": aggregatedCount,
			"pending_count": pendingCount, "latest_occurred_at": latest,
			"updated_at": now,
		}),
	}).Create(&row).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "save preference evidence aggregate")
	}
	return nil
}

// queueEvidenceTx 在业务事务内幂等写入一条待处理偏好证据。
func (service *PreferenceService) queueEvidenceTx(
	ctx context.Context,
	tx *gorm.DB,
	userID string,
	sourceType string,
	sourceID string,
	direction string,
	weight float64,
	confidence float64,
	facts preferenceFacts,
	occurredAt time.Time,
) error {
	facts = normalizePreferenceFacts(facts)
	encoded, err := json.Marshal(facts)
	if err != nil {
		return appErrors.FrontInternal.Wrap(err, "encode preference evidence")
	}
	now := service.now()
	if occurredAt.IsZero() {
		occurredAt = now
	}
	row := orderfoodModel.PreferenceEvidence{
		ID: orderfoodModel.NewID(), UserID: userID,
		SourceType: sourceType, SourceID: strings.TrimSpace(sourceID),
		Direction: direction, Confidence: confidence, Weight: weight,
		FactsJSON: datatypes.JSON(encoded), Status: orderfoodModel.PreferenceEvidencePending,
		RequestID: utils.RequestIDFromContext(ctx), OccurredAt: occurredAt.UTC(),
		CreatedAt: now, UpdatedAt: now,
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
	if result.Error != nil {
		return appErrors.FrontInternal.Wrap(result.Error, "queue preference evidence")
	}
	if result.RowsAffected == 0 {
		return nil
	}
	profile := orderfoodModel.UserPreferenceProfile{
		ID: orderfoodModel.NewID(), UserID: userID,
		UpdateState: orderfoodModel.PreferenceUpdatePending, UpdateEnabled: true,
		LastEvidenceAt: &occurredAt, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"update_state":   orderfoodModel.PreferenceUpdatePending,
			"update_enabled": true, "paused_reason_code": nil,
			"paused_reason_summary": nil, "last_evidence_at": occurredAt,
			"version": gorm.Expr("version + 1"), "updated_at": now,
		}),
	}).Create(&profile).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "mark preference profile pending")
	}
	return refreshEvidenceAggregateTx(tx, userID, sourceType, now)
}

// queueDishEvidenceTx 从已有菜品结构中提取事实并写入偏好证据。
func (service *PreferenceService) queueDishEvidenceTx(
	ctx context.Context,
	tx *gorm.DB,
	userID string,
	sourceType string,
	sourceID string,
	direction string,
	weight float64,
	dishID uint,
	occurredAt time.Time,
) error {
	var dish orderfoodModel.UserDish
	if err := preloadDish(tx).First(&dish, "id = ?", dishID).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "load preference evidence dish")
	}
	tags := make([]string, 0, len(dish.Tags)+1)
	for _, tag := range dish.Tags {
		tags = append(tags, tag.Name)
	}
	if dish.Category.Name != "" {
		tags = append(tags, dish.Category.Name)
	}
	ingredients := make([]string, 0, len(dish.Ingredients))
	for _, ingredient := range dish.Ingredients {
		ingredients = append(ingredients, ingredient.Name)
	}
	return service.queueEvidenceTx(
		ctx, tx, userID, sourceType, sourceID, direction, weight, 1,
		preferenceFacts{
			DishNames: []string{dish.Name}, Tags: tags, Ingredients: ingredients,
		},
		occurredAt,
	)
}

type preferenceScore struct {
	Name  string  // 词条原始名称
	Score float64 // 加权并衰减后的分数
}

// topPreferenceTerms 将分值映射转换为稳定排序的正向词条。
func topPreferenceTerms(scores map[string]preferenceScore, limit int) []string {
	values := make([]preferenceScore, 0, len(scores))
	for _, value := range scores {
		if value.Score > 0 {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(left int, right int) bool {
		if values[left].Score == values[right].Score {
			return values[left].Name < values[right].Name
		}
		return values[left].Score > values[right].Score
	})
	if len(values) > limit {
		values = values[:limit]
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}

// addPreferenceScores 将一组事实按证据方向、权重和时间衰减计入分值。
func addPreferenceScores(
	scores map[string]preferenceScore,
	values []string,
	baseScore float64,
) {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		current := scores[key]
		if current.Name == "" {
			current.Name = value
		}
		current.Score += baseScore
		scores[key] = current
	}
}

// preferenceTermSummaries 将画像词条转换为只读响应结构。
func preferenceTermSummaries(values []string) []orderfoodResponse.PreferenceTermSummary {
	result := make([]orderfoodResponse.PreferenceTermSummary, 0, len(values))
	for _, value := range values {
		result = append(result, orderfoodResponse.PreferenceTermSummary{
			Name: value, Origin: "inferred",
		})
	}
	return result
}

// rebuildPreferenceProfileTx 使用所有已聚合证据重新计算用户画像。
func (service *PreferenceService) rebuildPreferenceProfileTx(
	tx *gorm.DB,
	userID string,
	now time.Time,
) error {
	var rows []orderfoodModel.PreferenceEvidence
	if err := tx.Where("user_id = ? AND status = ?",
		userID, orderfoodModel.PreferenceEvidenceAggregated).
		Order("occurred_at asc, id asc").Find(&rows).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "load aggregated preference evidence")
	}
	tagScores := make(map[string]preferenceScore)
	orderedTagScores := make(map[string]preferenceScore)
	ingredientScores := make(map[string]preferenceScore)
	tasteScores := make(map[string]preferenceScore)
	settingsByKey := make(map[string]preferenceSetting)
	for _, row := range rows {
		var facts preferenceFacts
		if err := json.Unmarshal(row.FactsJSON, &facts); err != nil {
			return appErrors.FrontInternal.Wrap(err, "decode aggregated preference evidence")
		}
		direction := 1.0
		if row.Direction == "negative" {
			direction = -1
		}
		ageDays := math.Max(0, now.Sub(row.OccurredAt).Hours()/24)
		decay := 1 / (1 + ageDays/90)
		score := row.Weight * row.Confidence * direction * decay
		addPreferenceScores(tagScores, facts.Tags, score)
		addPreferenceScores(tagScores, facts.Cuisines, score)
		addPreferenceScores(tagScores, facts.CookingMethods, score)
		addPreferenceScores(tagScores, facts.DietTypes, score)
		addPreferenceScores(tasteScores, facts.Tastes, score)
		addPreferenceScores(ingredientScores, facts.Ingredients, score)
		if row.SourceType == orderfoodModel.PreferenceSourceDishOrdered {
			addPreferenceScores(orderedTagScores, facts.Tags, score)
		}
		if row.SourceType == orderfoodModel.PreferenceSourceExplicitSetting {
			for _, setting := range facts.Settings {
				key := setting.Type + "\x00" + strings.ToLower(setting.Value)
				previous, exists := settingsByKey[key]
				if !exists || setting.UpdatedAt.After(previous.UpdatedAt) {
					settingsByKey[key] = setting
				}
			}
		}
	}
	commonTags := topPreferenceTerms(tagScores, 8)
	orderedTags := topPreferenceTerms(orderedTagScores, 8)
	commonIngredients := topPreferenceTerms(ingredientScores, 10)
	tastes := topPreferenceTerms(tasteScores, 5)
	settings := make([]orderfoodResponse.PreferenceUserSettingSummary, 0, len(settingsByKey))
	settingTexts := make([]string, 0, len(settingsByKey))
	for _, setting := range settingsByKey {
		settings = append(settings, orderfoodResponse.PreferenceUserSettingSummary{
			Type: setting.Type, Value: setting.Value, UpdatedAt: setting.UpdatedAt,
		})
		settingTexts = append(settingTexts, setting.Value)
	}
	sort.Slice(settings, func(left int, right int) bool {
		if settings[left].Type == settings[right].Type {
			return settings[left].Value < settings[right].Value
		}
		return settings[left].Type < settings[right].Type
	})
	sort.Strings(settingTexts)
	content := orderfoodResponse.PreferenceProfileContent{
		CommonDishTags:            preferenceTermSummaries(commonTags),
		FrequentlyOrderedDishTags: preferenceTermSummaries(orderedTags),
		CommonIngredients:         preferenceTermSummaries(commonIngredients),
		UserSettings:              settings,
	}
	if len(tastes) > 0 {
		content.TastePreferenceSummary = &orderfoodResponse.PreferenceTextSummary{
			Text: strings.Join(tastes, "、"), Origin: "inferred",
		}
	}
	if len(settingTexts) > 0 {
		content.AvoidanceOrPreferenceSummary = &orderfoodResponse.PreferenceTextSummary{
			Text: strings.Join(settingTexts, "；"), Origin: "user_setting",
		}
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		return appErrors.FrontInternal.Wrap(err, "encode preference profile")
	}
	var pendingCount int64
	if err := tx.Model(&orderfoodModel.PreferenceEvidence{}).
		Where("user_id = ? AND status = ?", userID, orderfoodModel.PreferenceEvidencePending).
		Count(&pendingCount).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "count pending preference profile evidence")
	}
	updateState := orderfoodModel.PreferenceUpdateActive
	if pendingCount > 0 {
		updateState = orderfoodModel.PreferenceUpdatePending
	}
	if err := tx.Model(&orderfoodModel.UserPreferenceProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"has_profile": len(rows) > 0, "update_state": updateState,
			"update_enabled": true, "paused_reason_code": nil, "paused_reason_summary": nil,
			"last_aggregated_at": now, "profile_updated_at": now,
			"profile_json": datatypes.JSON(encoded), "latest_failure_json": nil,
			"version": gorm.Expr("version + 1"), "updated_at": now,
		}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "save preference profile")
	}
	return nil
}

// markPreferenceEvidenceFailure 记录一次安全失败并在达到上限后标记画像失败。
func (service *PreferenceService) markPreferenceEvidenceFailure(
	ctx context.Context,
	row orderfoodModel.PreferenceEvidence,
	processErr error,
) error {
	now := service.now()
	return service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current orderfoodModel.PreferenceEvidence
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current, "id = ?", row.ID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock failed preference evidence")
		}
		if current.Status != orderfoodModel.PreferenceEvidencePending {
			return nil
		}
		retryCount := current.RetryCount + 1
		status := orderfoodModel.PreferenceEvidencePending
		if retryCount >= preferenceEvidenceMaxAttempts {
			status = orderfoodModel.PreferenceEvidenceFailed
		}
		summary := "偏好证据处理失败，系统将自动重试"
		if status == orderfoodModel.PreferenceEvidenceFailed {
			summary = "偏好证据多次处理失败"
		}
		if err := tx.Model(&current).Updates(map[string]interface{}{
			"status": status, "retry_count": retryCount,
			"failure_summary": summary, "updated_at": now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "record preference evidence failure")
		}
		if status == orderfoodModel.PreferenceEvidenceFailed {
			failure, err := json.Marshal(orderfoodResponse.PreferenceProfileFailureSummary{
				FailedAt: now, StableErrorCode: strconv.Itoa(int(appErrors.GetType(processErr))),
				SafeMessage: summary, RequestID: current.RequestID,
			})
			if err != nil {
				return appErrors.FrontInternal.Wrap(err, "encode preference failure summary")
			}
			if err := tx.Model(&orderfoodModel.UserPreferenceProfile{}).
				Where("user_id = ?", current.UserID).
				Updates(map[string]interface{}{
					"update_state":        orderfoodModel.PreferenceUpdateFailed,
					"latest_failure_json": datatypes.JSON(failure),
					"version":             gorm.Expr("version + 1"), "updated_at": now,
				}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "mark preference profile failed")
			}
		}
		return refreshEvidenceAggregateTx(tx, current.UserID, current.SourceType, now)
	})
}

// processPreferenceEvidence 分析单条打卡图片证据并触发规则聚合。
func (service *PreferenceService) processPreferenceEvidence(
	ctx context.Context,
	row orderfoodModel.PreferenceEvidence,
) error {
	var facts preferenceFacts
	if err := json.Unmarshal(row.FactsJSON, &facts); err != nil {
		return service.markPreferenceEvidenceFailure(ctx, row, err)
	}
	confidence := row.Confidence
	if row.SourceType == orderfoodModel.PreferenceSourceCheckinImage {
		var output preferenceAnalysisOutput
		err := service.assist().runJSONCapability(
			ctx,
			orderfoodModel.AICapabilityCheckinImageAnalyze,
			map[string]string{
				"image_content": facts.ImageURL,
				"user_caption":  facts.Caption,
			},
			&output,
			"",
		)
		if err != nil {
			return service.markPreferenceEvidenceFailure(ctx, row, err)
		}
		facts.DishNames = output.DishNames
		facts.Ingredients = output.Ingredients
		facts.Tastes = output.Tastes
		facts.Cuisines = output.Cuisines
		facts.CookingMethods = output.CookingMethods
		facts.DietTypes = output.DietTypes
		facts.ImageURL = ""
		facts.Caption = ""
		confidence = math.Max(0.1, math.Min(1, output.Confidence))
	}
	facts = normalizePreferenceFacts(facts)
	encoded, err := json.Marshal(facts)
	if err != nil {
		return service.markPreferenceEvidenceFailure(ctx, row, err)
	}
	now := service.now()
	return service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current orderfoodModel.PreferenceEvidence
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current, "id = ?", row.ID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock pending preference evidence")
		}
		if current.Status != orderfoodModel.PreferenceEvidencePending {
			return nil
		}
		if err := tx.Model(&current).Updates(map[string]interface{}{
			"facts_json": datatypes.JSON(encoded), "confidence": confidence,
			"status":          orderfoodModel.PreferenceEvidenceAggregated,
			"failure_summary": nil, "aggregated_at": now, "updated_at": now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "aggregate preference evidence")
		}
		if err := refreshEvidenceAggregateTx(tx, current.UserID, current.SourceType, now); err != nil {
			return err
		}
		return service.rebuildPreferenceProfileTx(tx, current.UserID, now)
	})
}

// ProcessPendingEvidence 分批处理待聚合偏好证据，整体增强能力关闭时保留待办并暂停更新。
func (service *PreferenceService) ProcessPendingEvidence(
	ctx context.Context,
	batchSize int,
) error {
	runtime, err := service.runtime().Current(ctx)
	if err != nil {
		return err
	}
	if !runtime.EnhancedFeaturesEnabled {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	var rows []orderfoodModel.PreferenceEvidence
	if err := service.database().WithContext(ctx).
		Where("status = ?", orderfoodModel.PreferenceEvidencePending).
		Order("occurred_at asc, id asc").Limit(batchSize).Find(&rows).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "list pending preference evidence")
	}
	var firstErr error
	for _, row := range rows {
		if err := service.processPreferenceEvidence(ctx, row); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EngagementService 提供打卡、积分与通知业务能力。
type EngagementService struct {
	DB      *gorm.DB         // 业务数据库
	Runtime *RuntimeService  // 平台整体能力运行时服务
	Now     func() time.Time // 可注入的当前时间函数
}

func (service *EngagementService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *EngagementService) runtimeService() *RuntimeService {
	if service != nil && service.Runtime != nil {
		return service.Runtime
	}
	return &ServiceGroupApp.RuntimeService
}

func (service *EngagementService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

func shanghaiDate(value time.Time) string {
	return value.In(shanghaiLocation()).Format("2006-01-02")
}

// shanghaiLocation 返回业务自然日使用的上海时区。
func shanghaiLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return location
}

// shanghaiDayRange 返回指定时刻所在上海自然日对应的 UTC 起止时间。
func shanghaiDayRange(value time.Time) (time.Time, time.Time) {
	location := shanghaiLocation()
	local := value.In(location)
	start := time.Date(
		local.Year(),
		local.Month(),
		local.Day(),
		0,
		0,
		0,
		0,
		location,
	)
	return start.UTC(), start.AddDate(0, 0, 1).UTC()
}

func checkinResponse(row orderfoodModel.FrontCheckin) frontResponse.Checkin {
	return frontResponse.Checkin{
		ID: row.ID, DishName: row.DishName, ImageURL: row.ImageURL,
		Note: row.Note, CheckedAt: row.CheckedAt, Rewarded: row.Rewarded,
	}
}

// CreateCheckin 创建打卡。
func (service *EngagementService) CreateCheckin(
	ctx context.Context,
	userID string,
	input frontRequest.CheckinInput,
) (frontResponse.Checkin, bool, int64, bool, error) {
	db := service.database()
	now := service.now()
	date := shanghaiDate(now)
	runtime, err := service.runtimeService().Current(ctx)
	if err != nil {
		return frontResponse.Checkin{}, false, 0, false, err
	}
	var result orderfoodModel.FrontCheckin
	rewarded := false
	firstDailyCheckin := false
	rewardAmount := int64(0)
	// 锁定用户行后再判断当日首签，确保并发打卡不会重复发放每日奖励。
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user orderfoodModel.MiniAppUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&user, "id = ?", userID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock user for checkin")
		}
		var asset orderfoodModel.FrontMediaAsset
		if err := tx.First(&asset,
			"id = ? AND user_id = ? AND scene = ? AND review_status IN ?",
			input.ImageFileID, userID, "checkin", orderfoodModel.UsableMediaReviewStatuses(),
		).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontInvalidImage.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "resolve checkin image")
		}
		var dailyCount int64
		if err := tx.Model(&orderfoodModel.FrontCheckin{}).
			Where("user_id = ? AND checked_date = ?", userID, date).
			Count(&dailyCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "count daily checkins")
		}
		firstDailyCheckin = dailyCount == 0
		rewarded = firstDailyCheckin && runtime.EnhancedFeaturesEnabled
		result = orderfoodModel.FrontCheckin{
			ID: orderfoodModel.NewID(), UserID: userID, DishName: input.DishName,
			ImageFileID: asset.ID, ImageURL: asset.URL, Note: input.Note,
			CheckedDate: date, Rewarded: rewarded, CheckedAt: now, Version: 1, CreatedAt: now,
		}
		if err := tx.Create(&result).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create checkin")
		}
		updates := map[string]interface{}{
			"checkin_count":     gorm.Expr("checkin_count + 1"),
			"checkin_day_count": gorm.Expr("checkin_day_count + ?", boolInt(firstDailyCheckin)),
		}
		if rewarded {
			configuredReward, err := orderfoodService.CurrentDailyCheckinReward(ctx, tx)
			if err != nil {
				return err
			}
			rewardAmount = configuredReward
			updates["points"] = gorm.Expr("points + ?", rewardAmount)
			scene := "daily_checkin"
			objectType := "checkin"
			entry := orderfoodModel.FrontPointEntry{
				ID: orderfoodModel.NewID(), UserID: userID, Type: orderfoodModel.PointEarned,
				Scene: scene,
				Title: "做菜打卡奖励", Description: "每日首次做菜打卡",
				Amount: rewardAmount, BalanceAfter: user.Points + rewardAmount,
				RelatedObjectType: &objectType, RelatedObjectID: &result.ID,
				CreatedAt: now,
			}
			if err := tx.Create(&entry).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create checkin reward entry")
			}
		}
		if err := tx.Model(&orderfoodModel.MiniAppUser{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "update checkin counters")
		}
		if runtime.EnhancedFeaturesEnabled {
			caption := input.DishName
			if input.Note != nil && strings.TrimSpace(*input.Note) != "" {
				caption += "；" + strings.TrimSpace(*input.Note)
			}
			if err := ServiceGroupApp.PreferenceService.queueEvidenceTx(
				ctx, tx, userID,
				orderfoodModel.PreferenceSourceCheckinImage, result.ID,
				"positive", 1.5, 0.6,
				preferenceFacts{ImageURL: asset.URL, Caption: caption},
				now,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return frontResponse.Checkin{}, false, 0, false, err
	}
	return checkinResponse(result), rewarded, rewardAmount, runtime.EnhancedFeaturesEnabled, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// CheckinCalendar 获取指定月份的做菜打卡日历。
func (service *EngagementService) CheckinCalendar(
	ctx context.Context,
	userID string,
	month string,
) ([]string, int, error) {
	var dates []string
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontCheckin{}).
		Where("user_id = ? AND checked_date LIKE ?", userID, month+"%").
		Distinct("checked_date").
		Order("checked_date asc").
		Pluck("checked_date", &dates).Error; err != nil {
		return nil, 0, appErrors.FrontInternal.Wrap(err, "load checkin calendar")
	}
	var user orderfoodModel.MiniAppUser
	if err := service.database().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, 0, appErrors.FrontInternal.Wrap(err, "load checkin day count")
	}
	return dates, user.CheckinDayCount, nil
}

// Checkins 分页查询当前用户的做菜打卡记录。
func (service *EngagementService) Checkins(
	ctx context.Context,
	userID string,
	query frontRequest.PageQuery,
) (frontResponse.Page[frontResponse.Checkin], error) {
	query.Defaults()
	db := service.database().WithContext(ctx)
	var total int64
	if err := db.Model(&orderfoodModel.FrontCheckin{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.Checkin]{}, appErrors.FrontInternal.Wrap(err, "count checkins")
	}
	var rows []orderfoodModel.FrontCheckin
	if err := db.Where("user_id = ?", userID).Order("checked_at desc, id desc").
		Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error; err != nil {
		return frontResponse.Page[frontResponse.Checkin]{}, appErrors.FrontInternal.Wrap(err, "list checkins")
	}
	list := make([]frontResponse.Checkin, 0, len(rows))
	for _, row := range rows {
		list = append(list, checkinResponse(row))
	}
	return frontResponse.Page[frontResponse.Checkin]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

func (service *EngagementService) requireEnhanced(ctx context.Context) error {
	runtime, err := service.runtimeService().Current(ctx)
	if err != nil {
		return err
	}
	if !runtime.EnhancedFeaturesEnabled {
		return appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	return nil
}

func (service *EngagementService) requirePoints(ctx context.Context) error {
	runtime, err := service.runtimeService().Current(ctx)
	if err != nil {
		return err
	}
	if !runtime.PointsEnabled {
		return appErrors.FrontPointsDisabled.DefaultMsg()
	}
	return nil
}

// PointsSummary 获取当前积分及本月收支汇总。
func (service *EngagementService) PointsSummary(
	ctx context.Context,
	userID string,
) (int64, int64, int64, error) {
	if err := service.requirePoints(ctx); err != nil {
		return 0, 0, 0, err
	}
	var user orderfoodModel.MiniAppUser
	if err := service.database().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return 0, 0, 0, appErrors.FrontInternal.Wrap(err, "load point balance")
	}
	now := service.now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var earned int64
	var spent int64
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontPointEntry{}).
		Where("user_id = ? AND created_at >= ? AND amount > 0", userID, start).
		Select("COALESCE(SUM(amount), 0)").Scan(&earned).Error; err != nil {
		return 0, 0, 0, appErrors.FrontInternal.Wrap(err, "sum monthly earned points")
	}
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontPointEntry{}).
		Where("user_id = ? AND created_at >= ? AND amount < 0", userID, start).
		Select("COALESCE(SUM(-amount), 0)").Scan(&spent).Error; err != nil {
		return 0, 0, 0, appErrors.FrontInternal.Wrap(err, "sum monthly spent points")
	}
	return user.Points, earned, spent, nil
}

// PointEntries 分页查询当前用户的积分变动流水。
func (service *EngagementService) PointEntries(
	ctx context.Context,
	userID string,
	query frontRequest.PointEntriesQuery,
) (frontResponse.Page[frontResponse.PointEntry], error) {
	if err := service.requirePoints(ctx); err != nil {
		return frontResponse.Page[frontResponse.PointEntry]{}, err
	}
	query.PageQuery.Defaults()
	statement := service.database().WithContext(ctx).Model(&orderfoodModel.FrontPointEntry{}).
		Where("user_id = ?", userID)
	if query.Type != "" && query.Type != "all" {
		statement = statement.Where("type = ?", query.Type)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.PointEntry]{}, appErrors.FrontInternal.Wrap(err, "count point entries")
	}
	var rows []orderfoodModel.FrontPointEntry
	if err := statement.Order("created_at desc, id desc").
		Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error; err != nil {
		return frontResponse.Page[frontResponse.PointEntry]{}, appErrors.FrontInternal.Wrap(err, "list point entries")
	}
	list := make([]frontResponse.PointEntry, 0, len(rows))
	for _, row := range rows {
		list = append(list, frontResponse.PointEntry{
			ID: row.ID, Type: string(row.Type), Title: row.Title, Description: row.Description,
			Amount: row.Amount, RelatedEntryID: row.RelatedEntryID, CreatedAt: row.CreatedAt,
		})
	}
	return frontResponse.Page[frontResponse.PointEntry]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// FeatureUsages 分页查询当前用户的增强功能调用记录。
func (service *EngagementService) FeatureUsages(
	ctx context.Context,
	userID string,
	query frontRequest.FeatureUsageQuery,
) (frontResponse.Page[frontResponse.FeatureUsage], error) {
	if err := service.requireEnhanced(ctx); err != nil {
		return frontResponse.Page[frontResponse.FeatureUsage]{}, err
	}
	query.PageQuery.Defaults()
	statement := service.database().WithContext(ctx).Model(&orderfoodModel.FrontFeatureUsage{}).
		Where("user_id = ?", userID)
	if query.Feature != "" {
		statement = statement.Where("feature = ?", query.Feature)
	}
	if query.ExecutionStatus != "" {
		statement = statement.Where("execution_status = ?", query.ExecutionStatus)
	}
	if query.BillingStatus != "" {
		statement = statement.Where("billing_status = ?", query.BillingStatus)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.FeatureUsage]{}, appErrors.FrontInternal.Wrap(err, "count feature usages")
	}
	var rows []orderfoodModel.FrontFeatureUsage
	if err := statement.Order("created_at desc, id desc").
		Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error; err != nil {
		return frontResponse.Page[frontResponse.FeatureUsage]{}, appErrors.FrontInternal.Wrap(err, "list feature usages")
	}
	list := make([]frontResponse.FeatureUsage, 0, len(rows))
	for _, row := range rows {
		var refundPointEntryID *string
		var refundEntry orderfoodModel.FrontPointEntry
		if err := service.database().WithContext(ctx).
			Where(
				"type = ? AND related_object_type = ? AND related_object_id = ?",
				orderfoodModel.PointRefund, "feature_usage", row.ID,
			).
			Order("created_at desc, id desc").
			First(&refundEntry).Error; err == nil {
			value := refundEntry.ID
			refundPointEntryID = &value
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.Page[frontResponse.FeatureUsage]{},
				appErrors.FrontInternal.Wrap(err, "load feature refund entry")
		}
		list = append(list, frontResponse.FeatureUsage{
			ID: row.ID, Feature: row.Feature, PointCost: row.PointCost,
			ExecutionStatus: row.ExecutionStatus, BillingStatus: row.BillingStatus,
			RefundPointEntryID: refundPointEntryID, CreatedAt: row.CreatedAt,
		})
	}
	return frontResponse.Page[frontResponse.FeatureUsage]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

func (service *EngagementService) featurePolicy(
	ctx context.Context,
	capabilityCode string,
) (string, int, int, int, error) {
	runtime, err := service.runtimeService().Current(ctx)
	if err != nil {
		return "", 0, 0, 0, err
	}
	if !runtime.EnhancedFeaturesEnabled {
		return "", 0, 0, 0, appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	var definition orderfoodModel.AICapabilityDefinition
	if err := service.database().WithContext(ctx).
		First(&definition, "code = ?", capabilityCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, 0, 0, appErrors.FrontFeatureDisabled.DefaultMsg()
		}
		return "", 0, 0, 0, appErrors.FrontInternal.Wrap(err, "load feature capability")
	}
	featureEnabled := false
	for _, configured := range runtime.Features {
		if configured.Code == definition.ClientFeatureCode && configured.Enabled {
			featureEnabled = true
			break
		}
	}
	if !featureEnabled {
		return "", 0, 0, 0, appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	var config orderfoodModel.AICapabilityConfig
	if err := service.database().WithContext(ctx).
		First(&config, "capability_code = ?", capabilityCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, 0, 0, appErrors.FrontFeatureDisabled.DefaultMsg()
		}
		return "", 0, 0, 0, appErrors.FrontInternal.Wrap(err, "load feature capability config")
	}
	return definition.ClientFeatureCode,
		config.PointCost,
		config.FreeQuotaPerDay,
		config.DailyLimitPerUser,
		nil
}

// BeginFeatureUsage 开始功能调用记录。
func (service *EngagementService) BeginFeatureUsage(
	ctx context.Context,
	userID string,
	capabilityCode string,
	preferFree bool,
) (orderfoodModel.FrontFeatureUsage, int64, error) {
	feature, pointCost, freeQuota, dailyLimit, err :=
		service.featurePolicy(ctx, capabilityCode)
	if err != nil {
		return orderfoodModel.FrontFeatureUsage{}, 0, err
	}
	now := service.now()
	startOfDay, endOfDay := shanghaiDayRange(now)
	var usage orderfoodModel.FrontFeatureUsage
	var balance int64
	// 免费额度判定、积分扣减、积分流水和调用记录必须原子提交，避免出现已扣分但无调用记录。
	err = service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user orderfoodModel.MiniAppUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", userID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock feature usage user")
		}
		activeExecutionStatuses := []string{
			orderfoodModel.FeatureExecutionProcessing,
			orderfoodModel.FeatureExecutionSucceeded,
		}
		if dailyLimit > 0 {
			var dailyUsed int64
			if err := tx.Model(&orderfoodModel.FrontFeatureUsage{}).
				Where(
					"user_id = ? AND capability_code = ? AND created_at >= ? AND created_at < ? "+
						"AND execution_status IN ?",
					userID,
					capabilityCode,
					startOfDay,
					endOfDay,
					activeExecutionStatuses,
				).
				Count(&dailyUsed).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "count daily feature usages")
			}
			if dailyUsed >= int64(dailyLimit) {
				return appErrors.FrontQuotaExceeded.DefaultMsg()
			}
		}
		var freeUsed int64
		if freeQuota > 0 {
			if err := tx.Model(&orderfoodModel.FrontFeatureUsage{}).
				Where(
					"user_id = ? AND capability_code = ? AND created_at >= ? AND created_at < ? "+
						"AND execution_status IN ? AND point_cost = 0",
					userID,
					capabilityCode,
					startOfDay,
					endOfDay,
					activeExecutionStatuses,
				).
				Count(&freeUsed).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "count free feature usages")
			}
		}
		chargedCost := pointCost
		if preferFree && int(freeUsed) < freeQuota {
			chargedCost = 0
		}
		if chargedCost > 0 && user.Points < int64(chargedCost) {
			return appErrors.FrontInsufficientPoints.DefaultMsg()
		}
		usage = orderfoodModel.FrontFeatureUsage{
			ID: orderfoodModel.NewID(), UserID: userID, Feature: feature,
			CapabilityCode:   capabilityCode,
			RequestID:        utils.RequestIDFromContext(ctx),
			IdempotencyKey:   utils.IdempotencyKeyFromContext(ctx),
			PointCost:        chargedCost,
			ExecutionStatus:  orderfoodModel.FeatureExecutionProcessing,
			BillingStatus:    orderfoodModel.FeatureBillingNotCharged,
			EstimatedCostCNY: "0.000000", Version: 1,
			CreatedAt: now, UpdatedAt: now,
		}
		if chargedCost > 0 {
			usage.BillingStatus = orderfoodModel.FeatureBillingCharged
			user.Points -= int64(chargedCost)
			if err := tx.Model(&orderfoodModel.MiniAppUser{}).Where("id = ?", userID).
				Update("points", user.Points).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "charge feature points")
			}
			scene := "feature_usage"
			objectType := "feature_usage"
			if err := tx.Create(&orderfoodModel.FrontPointEntry{
				ID: orderfoodModel.NewID(), UserID: userID, Type: orderfoodModel.PointSpent,
				Scene: scene,
				Title: "增强功能使用", Description: strings.TrimSpace(feature),
				Amount: -int64(chargedCost), BalanceAfter: user.Points,
				RelatedObjectType: &objectType, RelatedObjectID: &usage.ID,
				CreatedAt: now,
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create feature point charge")
			}
		}
		if err := tx.Create(&usage).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create feature usage")
		}
		balance = user.Points
		return nil
	})
	return usage, balance, err
}

// featureRemainingQuota 返回指定能力在当前上海自然日可继续使用的免费次数。
func (service *EngagementService) featureRemainingQuota(
	ctx context.Context,
	userID string,
	capabilityCode string,
) (int, int, error) {
	_, pointCost, freeQuota, dailyLimit, err :=
		service.featurePolicy(ctx, capabilityCode)
	if err != nil {
		return 0, 0, err
	}
	startOfDay, endOfDay := shanghaiDayRange(service.now())
	activeExecutionStatuses := []string{
		orderfoodModel.FeatureExecutionProcessing,
		orderfoodModel.FeatureExecutionSucceeded,
	}
	var freeUsed int64
	if err := service.database().WithContext(ctx).
		Model(&orderfoodModel.FrontFeatureUsage{}).
		Where(
			"user_id = ? AND capability_code = ? AND created_at >= ? AND created_at < ? "+
				"AND execution_status IN ? AND point_cost = 0",
			userID,
			capabilityCode,
			startOfDay,
			endOfDay,
			activeExecutionStatuses,
		).
		Count(&freeUsed).Error; err != nil {
		return 0, 0, appErrors.FrontInternal.Wrap(err, "count remaining free feature usages")
	}
	remaining := freeQuota - int(freeUsed)
	if remaining < 0 {
		remaining = 0
	}
	if dailyLimit > 0 {
		var dailyUsed int64
		if err := service.database().WithContext(ctx).
			Model(&orderfoodModel.FrontFeatureUsage{}).
			Where(
				"user_id = ? AND capability_code = ? AND created_at >= ? AND created_at < ? "+
					"AND execution_status IN ?",
				userID,
				capabilityCode,
				startOfDay,
				endOfDay,
				activeExecutionStatuses,
			).
			Count(&dailyUsed).Error; err != nil {
			return 0, 0, appErrors.FrontInternal.Wrap(err, "count remaining daily feature usages")
		}
		dailyRemaining := dailyLimit - int(dailyUsed)
		if dailyRemaining < remaining {
			remaining = dailyRemaining
		}
		if remaining < 0 {
			remaining = 0
		}
	}
	return remaining, pointCost, nil
}

// FinishFeatureUsage 完成功能调用记录。
func (service *EngagementService) FinishFeatureUsage(
	ctx context.Context,
	usage orderfoodModel.FrontFeatureUsage,
	success bool,
) (frontResponse.FeatureUsageResult, error) {
	now := service.now()
	current := usage
	var balance int64
	if success {
		err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&current, "id = ?", usage.ID).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "lock successful feature usage")
			}
			if current.ExecutionStatus == orderfoodModel.FeatureExecutionProcessing {
				if err := tx.Model(&current).Updates(map[string]interface{}{
					"execution_status": orderfoodModel.FeatureExecutionSucceeded,
					"duration_ms":      int(now.Sub(current.CreatedAt).Milliseconds()),
					"finished_at":      now,
					"updated_at":       now,
				}).Error; err != nil {
					return appErrors.FrontInternal.Wrap(err, "finish successful feature usage")
				}
				current.ExecutionStatus = orderfoodModel.FeatureExecutionSucceeded
				current.FinishedAt = &now
			}
			if err := tx.Model(&orderfoodModel.MiniAppUser{}).Where("id = ?", current.UserID).
				Select("points").Scan(&balance).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "load feature point balance")
			}
			return nil
		})
		return featureUsageResult(current, current.BillingStatus, balance), err
	}

	billingStatus := current.BillingStatus
	// 先持久化失败终态和待退款状态，保证即时退款事务失败后仍有补偿任务可继续处理。
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current, "id = ?", usage.ID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock failed feature usage")
		}
		billingStatus = current.BillingStatus
		if current.ExecutionStatus == orderfoodModel.FeatureExecutionProcessing {
			if current.PointCost > 0 &&
				current.BillingStatus == orderfoodModel.FeatureBillingCharged {
				billingStatus = orderfoodModel.FeatureBillingRefundPending
			}
			if err := tx.Model(&current).Updates(map[string]interface{}{
				"execution_status": orderfoodModel.FeatureExecutionFailed,
				"billing_status":   billingStatus,
				"duration_ms":      int(now.Sub(current.CreatedAt).Milliseconds()),
				"finished_at":      now,
				"updated_at":       now,
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "mark failed feature usage")
			}
			current.ExecutionStatus = orderfoodModel.FeatureExecutionFailed
			current.BillingStatus = billingStatus
			current.FinishedAt = &now
		}
		if err := tx.Model(&orderfoodModel.MiniAppUser{}).Where("id = ?", current.UserID).
			Select("points").Scan(&balance).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "load failed feature point balance")
		}
		return nil
	})
	if err != nil {
		return featureUsageResult(current, billingStatus, balance), err
	}
	if billingStatus != orderfoodModel.FeatureBillingRefundPending {
		return featureUsageResult(current, billingStatus, balance), nil
	}
	balance, err = service.refundPendingFeatureUsage(ctx, current.ID)
	if err != nil {
		return featureUsageResult(current, orderfoodModel.FeatureBillingRefundPending, balance),
			appErrors.FrontRefundPending.Wrap(err, "refund feature usage later")
	}
	current.BillingStatus = orderfoodModel.FeatureBillingRefunded
	return featureUsageResult(current, orderfoodModel.FeatureBillingRefunded, balance), nil
}

// featureUsageResult 将调用模型转换为符合小程序契约的计费结果。
func featureUsageResult(
	usage orderfoodModel.FrontFeatureUsage,
	billingStatus string,
	balance int64,
) frontResponse.FeatureUsageResult {
	var message *string
	switch billingStatus {
	case orderfoodModel.FeatureBillingRefundPending:
		value := appErrors.Message(appErrors.FrontRefundPending)
		message = &value
	case orderfoodModel.FeatureBillingRefunded:
		value := appErrors.Message(appErrors.FrontExecutionRefunded)
		message = &value
	}
	return frontResponse.FeatureUsageResult{
		UsageID: usage.ID, PointCost: usage.PointCost,
		BillingStatus: billingStatus, PointBalance: balance, UserMessage: message,
	}
}

// refundPendingFeatureUsage 幂等退还一笔待退款调用的积分并写退款流水和站内通知。
func (service *EngagementService) refundPendingFeatureUsage(
	ctx context.Context,
	usageID string,
) (int64, error) {
	now := service.now()
	var balance int64
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var usage orderfoodModel.FrontFeatureUsage
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&usage, "id = ?", usageID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock pending feature refund")
		}
		if usage.BillingStatus == orderfoodModel.FeatureBillingRefunded {
			return tx.Model(&orderfoodModel.MiniAppUser{}).Where("id = ?", usage.UserID).
				Select("points").Scan(&balance).Error
		}
		if usage.BillingStatus != orderfoodModel.FeatureBillingRefundPending || usage.PointCost <= 0 {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		var user orderfoodModel.MiniAppUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&user, "id = ?", usage.UserID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock feature refund user")
		}
		var chargeEntry orderfoodModel.FrontPointEntry
		if err := tx.Where(
			"user_id = ? AND type = ? AND related_object_type = ? AND related_object_id = ?",
			usage.UserID, orderfoodModel.PointSpent, "feature_usage", usage.ID,
		).Order("created_at desc, id desc").First(&chargeEntry).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "load feature charge entry")
		}
		balance = user.Points + int64(usage.PointCost)
		relatedEntryID := chargeEntry.ID
		objectType := "feature_usage"
		if err := tx.Create(&orderfoodModel.FrontPointEntry{
			ID: orderfoodModel.NewID(), UserID: usage.UserID, Type: orderfoodModel.PointRefund,
			Scene: "feature_refund", Title: "处理失败退还", Description: usage.Feature,
			Amount: int64(usage.PointCost), BalanceAfter: balance,
			RelatedObjectType: &objectType, RelatedObjectID: &usage.ID,
			RelatedEntryID: &relatedEntryID, CreatedAt: now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create feature refund entry")
		}
		if err := tx.Model(&user).Update("points", balance).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "refund feature points")
		}
		if err := tx.Model(&usage).Updates(map[string]interface{}{
			"billing_status": orderfoodModel.FeatureBillingRefunded,
			"updated_at":     now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "complete feature refund")
		}
		targetType := "feature_usage"
		targetID := usage.ID
		if err := tx.Create(&orderfoodModel.UserNotification{
			ID: orderfoodModel.NewID(), UserID: usage.UserID,
			Type: "feature_refund", Title: "积分已退还",
			Content:    "本次处理未成功，已退还" + strconv.Itoa(usage.PointCost) + "积分",
			TargetType: &targetType, TargetID: &targetID, CreatedAt: now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create feature refund notification")
		}
		return nil
	})
	return balance, err
}

// ProcessPendingRefunds 分批重试待退积分记录，单笔失败不会阻断同批其他记录。
func (service *EngagementService) ProcessPendingRefunds(ctx context.Context, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	var usageIDs []string
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontFeatureUsage{}).
		Where("billing_status = ?", orderfoodModel.FeatureBillingRefundPending).
		Order("updated_at asc, id asc").Limit(batchSize).Pluck("id", &usageIDs).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "list pending feature refunds")
	}
	var firstErr error
	for _, usageID := range usageIDs {
		if _, err := service.refundPendingFeatureUsage(ctx, usageID); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// NotificationSummary 获取通知未读数和总数。
func (service *EngagementService) NotificationSummary(
	ctx context.Context,
	userID string,
) (int64, int64, error) {
	db := service.database().WithContext(ctx).Model(&orderfoodModel.UserNotification{}).
		Where("user_id = ?", userID)
	var total int64
	var unread int64
	if err := db.Count(&total).Error; err != nil {
		return 0, 0, appErrors.FrontInternal.Wrap(err, "count notifications")
	}
	if err := db.Where("read_at IS NULL").Count(&unread).Error; err != nil {
		return 0, 0, appErrors.FrontInternal.Wrap(err, "count unread notifications")
	}
	return unread, total, nil
}

// Notifications 分页查询当前用户的站内通知。
func (service *EngagementService) Notifications(
	ctx context.Context,
	userID string,
	query frontRequest.NotificationListQuery,
) (frontResponse.Page[frontResponse.Notification], error) {
	query.PageQuery.Defaults()
	if query.Scope == "" {
		query.Scope = "unread"
	}
	statement := service.database().WithContext(ctx).Model(&orderfoodModel.UserNotification{}).
		Where("user_id = ?", userID)
	if query.Scope == "unread" {
		statement = statement.Where("read_at IS NULL")
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.Notification]{}, appErrors.FrontInternal.Wrap(err, "count notification list")
	}
	var rows []orderfoodModel.UserNotification
	if err := statement.Order("created_at desc, id desc").
		Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error; err != nil {
		return frontResponse.Page[frontResponse.Notification]{}, appErrors.FrontInternal.Wrap(err, "list notifications")
	}
	list := make([]frontResponse.Notification, 0, len(rows))
	for _, row := range rows {
		list = append(list, frontResponse.Notification{
			ID: row.ID, Type: row.Type, Title: row.Title, Content: row.Content,
			TargetType: row.TargetType, TargetID: row.TargetID, Read: row.ReadAt != nil, CreatedAt: row.CreatedAt,
		})
	}
	return frontResponse.Page[frontResponse.Notification]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// ReadNotification 读取通知。
func (service *EngagementService) ReadNotification(
	ctx context.Context,
	userID string,
	notificationID string,
) (time.Time, error) {
	now := service.now()
	result := service.database().WithContext(ctx).Model(&orderfoodModel.UserNotification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read_at", now)
	if result.Error != nil {
		return time.Time{}, appErrors.FrontInternal.Wrap(result.Error, "read notification")
	}
	if result.RowsAffected == 0 {
		return time.Time{}, appErrors.FrontNotFound.DefaultMsg()
	}
	return now, nil
}

// ReadAllNotifications 将当前用户的全部通知标记为已读。
func (service *EngagementService) ReadAllNotifications(
	ctx context.Context,
	userID string,
) (int64, error) {
	now := service.now()
	result := service.database().WithContext(ctx).Model(&orderfoodModel.UserNotification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now)
	if result.Error != nil {
		return 0, appErrors.FrontInternal.Wrap(result.Error, "read all notifications")
	}
	return result.RowsAffected, nil
}

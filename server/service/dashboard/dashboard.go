package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"math"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dashboardRequest "github.com/dyjh/order-food-mini-app/server/model/dashboard/request"
	dashboardResponse "github.com/dyjh/order-food-mini-app/server/model/dashboard/response"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/gorm"
)

const (
	// PermissionDashboardRead 表示读取运营概览所需权限。
	PermissionDashboardRead = "orderfood:dashboard:read"
	dashboardTimezone       = "Asia/Shanghai"
)

type dashboardFailureRow struct {
	ProviderID      string // 供应商ID
	ModelID         string // 模型ID
	ExecutionStatus string // 执行状态
}

// DashboardService 提供管理端运营概览统计能力。
type DashboardService struct {
	DB         *gorm.DB                         // 数据库连接
	Permission *serviceCommon.PermissionService // 管理端权限服务
	Now        func() time.Time                 // 当前时间函数
}

// NewDashboardService 创建运营概览服务实例。
func NewDashboardService(db *gorm.DB, permission *serviceCommon.PermissionService) *DashboardService {
	if permission == nil {
		permission = serviceCommon.NewPermissionService(db)
	}
	return &DashboardService{DB: db, Permission: permission, Now: time.Now}
}

// database 返回服务使用的数据库连接。
func (service *DashboardService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回服务使用的UTC时间。
func (service *DashboardService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// GetDashboard 按上海自然日范围生成轻量运营概览。
func (service *DashboardService) GetDashboard(
	ctx context.Context,
	query dashboardRequest.DashboardQuery,
	actor commonRequest.AdminActor,
) (dashboardResponse.DashboardData, error) {
	query.ApplyDefaults()
	if service == nil || service.Permission == nil {
		return dashboardResponse.DashboardData{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionDashboardRead); err != nil {
		return dashboardResponse.DashboardData{}, err
	}
	db := service.database()
	if db == nil {
		return dashboardResponse.DashboardData{}, appErrors.AdminInternal.DefaultMsg()
	}
	rangeStart, rangeEnd, location, err := dashboardRange(query.Range, service.now())
	if err != nil {
		return dashboardResponse.DashboardData{}, err
	}
	metrics, capabilityState, err := service.dashboardMetrics(
		ctx, db, query.Range, rangeStart.UTC(), rangeEnd.UTC(), location,
	)
	if err != nil {
		return dashboardResponse.DashboardData{}, err
	}
	alerts, err := service.dashboardAlerts(
		ctx, db, query.Range, rangeStart.UTC(), rangeEnd.UTC(),
	)
	if err != nil {
		return dashboardResponse.DashboardData{}, err
	}
	return dashboardResponse.DashboardData{
		Range: query.Range, Timezone: dashboardTimezone,
		RangeStart: rangeStart, RangeEnd: rangeEnd,
		GeneratedAt: service.now(), Currency: "CNY",
		Metrics: metrics, CapabilityState: capabilityState, Alerts: alerts,
	}, nil
}

// dashboardRange 将查询范围转换为上海自然日的起止时间。
func dashboardRange(
	rangeCode string,
	now time.Time,
) (time.Time, time.Time, *time.Location, error) {
	location, err := time.LoadLocation(dashboardTimezone)
	if err != nil {
		location = time.FixedZone(dashboardTimezone, 8*60*60)
	}
	localNow := now.In(location)
	start := time.Date(
		localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location,
	)
	switch rangeCode {
	case "today":
	case "last7Days":
		start = start.AddDate(0, 0, -6)
	case "last30Days":
		start = start.AddDate(0, 0, -29)
	default:
		return time.Time{}, time.Time{}, nil, appErrors.AdminBadRequest.DefaultMsg()
	}
	return start, localNow, location, nil
}

// dashboardMetrics 计算运营指标和当前平台能力状态。
func (service *DashboardService) dashboardMetrics(
	ctx context.Context,
	db *gorm.DB,
	rangeCode string,
	start time.Time,
	end time.Time,
	location *time.Location,
) ([]dashboardResponse.DashboardMetric, aiResponse.PlatformCapabilitySnapshot, error) {
	newUsers, err := dashboardCount(
		ctx, db.Model(&userModel.MiniAppUser{}).
			Where("registered_at >= ? AND registered_at <= ?", start, end),
		"count new users",
	)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	activeUsers, err := dashboardCount(
		ctx, db.Model(&engagementModel.FrontUserActivityDay{}).
			Distinct("user_id").
			Where("active_date >= ? AND active_date <= ?",
				start.In(location).Format("2006-01-02"),
				end.In(location).Format("2006-01-02")),
		"count active users",
	)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	firstDishRate, err := dashboardFirstDishRate(ctx, db, start, end, newUsers)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	usableDishes, err := dashboardCount(
		ctx, db.Model(&contentModel.UserDish{}).
			Where("status = ?", contentModel.DishStatusUsable),
		"count usable dishes",
	)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	discoverableDishes, err := dashboardCount(
		ctx, db.Model(&contentModel.UserDish{}).
			Where("status = ? AND discoverable = ? AND source_locked = ?",
				contentModel.DishStatusUsable, true, false),
		"count discoverable dishes",
	)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	recommendedDishes, err := dashboardCount(
		ctx, db.Model(&dishModel.PlatformRecommendation{}).
			Where("status = ?", "published"),
		"count published recommendations",
	)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}

	type rangedMetricSpec struct {
		key         string                 // 指标编码
		label       string                 // 指标名称
		model       interface{}            // GORM模型
		where       string                 // 时间范围条件
		targetPage  string                 // 明细页面路由名
		targetQuery map[string]interface{} // 明细页面筛选
	}
	rangedSpecs := []rangedMetricSpec{
		{
			"meals_created", "创建饭局数", &mealModel.FrontMeal{},
			"created_at >= ? AND created_at <= ?", "OrderFoodMeals",
			map[string]interface{}{"timeField": "created", "range": rangeCode},
		},
		{
			"meal_participants_joined", "加入人数", &mealModel.MealParticipant{},
			"joined_at >= ? AND joined_at <= ?", "OrderFoodMeals",
			map[string]interface{}{"timeField": "joined", "range": rangeCode},
		},
		{
			"meals_confirmed", "确认菜单数", &mealModel.FrontMeal{},
			"confirmed_at >= ? AND confirmed_at <= ?", "OrderFoodMeals",
			map[string]interface{}{"timeField": "confirmed", "range": rangeCode},
		},
		{
			"shopping_lists_created", "采购清单生成数", &mealModel.FrontShoppingList{},
			"created_at >= ? AND created_at <= ?", "OrderFoodShoppingLists",
			map[string]interface{}{"range": rangeCode},
		},
		{
			"checkins", "打卡数", &engagementModel.FrontCheckin{},
			"created_at >= ? AND created_at <= ?", "OrderFoodMedia",
			map[string]interface{}{
				"sourceScene": "checkin", "boundObjectType": "checkin", "range": rangeCode,
			},
		},
	}
	rangedMetrics := make([]dashboardResponse.DashboardMetric, 0, len(rangedSpecs))
	for _, spec := range rangedSpecs {
		statement := db.Model(spec.model).Where(spec.where, start, end)
		if spec.key == "meal_participants_joined" {
			// 创建饭局时生成的创建者参与关系不计入“加入人数”。
			statement = statement.Where(
				"NOT EXISTS (SELECT 1 FROM of_meals WHERE of_meals.id = of_meal_members.meal_id " +
					"AND of_meals.creator_id = of_meal_members.user_id)",
			)
		}
		value, countErr := dashboardCount(ctx, statement, "count "+spec.key)
		if countErr != nil {
			return nil, aiResponse.PlatformCapabilitySnapshot{}, countErr
		}
		rangedMetrics = append(rangedMetrics, dashboardMetric(
			spec.key, spec.label, value, "次", spec.targetPage,
			spec.targetQuery,
		))
	}

	pointsEarned, err := dashboardPointSum(ctx, db, start, end, engagementModel.PointEarned)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	pointsSpent, err := dashboardPointSum(ctx, db, start, end, engagementModel.PointSpent)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	pointsRefunded, err := dashboardPointSum(ctx, db, start, end, engagementModel.PointRefund)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	aiMetrics, err := dashboardAIMetrics(ctx, db, rangeCode, start, end)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}
	capabilityState, err := dashboardCapabilityState(ctx, db)
	if err != nil {
		return nil, aiResponse.PlatformCapabilitySnapshot{}, err
	}

	metrics := []dashboardResponse.DashboardMetric{
		dashboardMetric("new_users", "新增用户", newUsers, "人", "OrderFoodUsers", map[string]interface{}{"registeredRange": rangeCode}),
		dashboardMetric("active_users", "活跃用户", activeUsers, "人", "OrderFoodUsers", map[string]interface{}{"activeRange": rangeCode}),
		dashboardMetric("first_dish_creation_rate", "新用户首次创建菜品率", firstDishRate, "%", "OrderFoodUsers", map[string]interface{}{"registeredRange": rangeCode}),
		dashboardMetric("usable_dishes", "可用菜品数", usableDishes, "道", "OrderFoodUserDishes", map[string]interface{}{"status": contentModel.DishStatusUsable}),
		dashboardMetric("discoverable_dishes", "允许被发现菜品数", discoverableDishes, "道", "OrderFoodDiscoverableDishes", nil),
		dashboardMetric("recommended_dishes", "推荐中菜品数", recommendedDishes, "道", "OrderFoodRecommendations", map[string]interface{}{"status": "published"}),
	}
	metrics = append(metrics, rangedMetrics...)
	metrics = append(metrics,
		dashboardMetric("points_earned", "获得积分", pointsEarned, "积分", "OrderFoodPointEntries", map[string]interface{}{"type": engagementModel.PointEarned, "range": rangeCode}),
		dashboardMetric("points_spent", "消耗积分", pointsSpent, "积分", "OrderFoodPointEntries", map[string]interface{}{"type": engagementModel.PointSpent, "range": rangeCode}),
		dashboardMetric("points_refunded", "退款积分", pointsRefunded, "积分", "OrderFoodPointEntries", map[string]interface{}{"type": engagementModel.PointRefund, "range": rangeCode}),
	)
	metrics = append(metrics, aiMetrics...)
	metrics = append(metrics,
		dashboardMetric("platform_enabled", "平台总开关", capabilityState.PlatformDefaultEnabled, "", "OrderFoodPlatformPolicy", nil),
		dashboardMetric("emergency_disabled", "紧急停用", capabilityState.EmergencyDisabled, "", "OrderFoodPlatformPolicy", nil),
		dashboardMetric("effective_enabled_users", "生效用户数量", capabilityState.EffectiveEnabledUserCount, "人", "OrderFoodUsers", map[string]interface{}{"capabilityEffective": "enabled"}),
	)
	return metrics, capabilityState, nil
}

// dashboardAlerts 统计当前需要人工处理的异常提醒。
func (service *DashboardService) dashboardAlerts(
	ctx context.Context,
	db *gorm.DB,
	rangeCode string,
	start time.Time,
	end time.Time,
) ([]dashboardResponse.DashboardAlert, error) {
	type alertSpec struct {
		alertType   string                 // 异常类型
		level       string                 // 异常级别
		model       interface{}            // GORM模型
		where       string                 // 查询条件
		args        []interface{}          // 查询参数
		targetPage  string                 // 目标页面
		targetQuery map[string]interface{} // 页面筛选条件
	}
	specs := []alertSpec{
		{
			alertType: "governance_job_abnormal", level: "warning",
			model: &contentModel.GovernanceJob{},
			where: "status IN ?", args: []interface{}{[]string{"processing", "partially_succeeded", "failed"}},
			targetPage:  "OrderFoodGovernanceRecords",
			targetQuery: map[string]interface{}{"jobStatus": "abnormal"},
		},
		{
			alertType: "moderation_failed", level: "warning",
			model:       &contentModel.ImageModerationRecord{},
			where:       "created_at >= ? AND created_at <= ? AND status IN ?",
			args:        []interface{}{start, end, []string{"rejected", "failed"}},
			targetPage:  "OrderFoodModerationRecords",
			targetQuery: map[string]interface{}{"status": "rejected_or_failed", "range": rangeCode},
		},
		{
			alertType: "subscribe_failed", level: "warning",
			model:       &engagementModel.SubscribeMessageLog{},
			where:       "created_at >= ? AND created_at <= ? AND status = ?",
			args:        []interface{}{start, end, engagementModel.SubscribeLogFailed},
			targetPage:  "OrderFoodSubscribeLogs",
			targetQuery: map[string]interface{}{"status": engagementModel.SubscribeLogFailed, "range": rangeCode},
		},
		{
			alertType: "refund_pending", level: "critical",
			model: &aiModel.FrontFeatureUsage{},
			where: "billing_status = ?", args: []interface{}{"refund_pending"},
			targetPage:  "OrderFoodAIUsages",
			targetQuery: map[string]interface{}{"billingStatus": "refund_pending"},
		},
	}
	alerts := make([]dashboardResponse.DashboardAlert, 0, len(specs)+2)
	for _, spec := range specs {
		count, err := dashboardCount(
			ctx, db.Model(spec.model).Where(spec.where, spec.args...), "count dashboard alert",
		)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			alerts = append(alerts, dashboardResponse.DashboardAlert{
				Type: spec.alertType, Count: count, Level: spec.level,
				TargetPage: spec.targetPage, TargetQuery: spec.targetQuery,
			})
		}
	}
	providerFailures, modelFailures, err := dashboardConsecutiveFailures(ctx, db)
	if err != nil {
		return nil, err
	}
	if providerFailures > 0 {
		alerts = append(alerts, dashboardResponse.DashboardAlert{
			Type: "ai_provider_consecutive_failure", Count: providerFailures, Level: "critical",
			TargetPage:  "OrderFoodAIUsages",
			TargetQuery: map[string]interface{}{"executionStatus": "failed"},
		})
	}
	if modelFailures > 0 {
		alerts = append(alerts, dashboardResponse.DashboardAlert{
			Type: "ai_model_consecutive_failure", Count: modelFailures, Level: "critical",
			TargetPage:  "OrderFoodAIUsages",
			TargetQuery: map[string]interface{}{"executionStatus": "failed"},
		})
	}
	return alerts, nil
}

// dashboardCount 执行带统一错误包装的计数查询。
func dashboardCount(ctx context.Context, statement *gorm.DB, operation string) (int64, error) {
	var count int64
	if err := statement.WithContext(ctx).Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, operation)
	}
	return count, nil
}

// dashboardFirstDishRate 计算范围内新增用户截至范围结束的首次创建菜品率。
func dashboardFirstDishRate(
	ctx context.Context,
	db *gorm.DB,
	start time.Time,
	end time.Time,
	newUsers int64,
) (interface{}, error) {
	if newUsers == 0 {
		return nil, nil
	}
	var created int64
	if err := db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM of_users AS users "+
			"WHERE users.registered_at >= ? AND users.registered_at <= ? "+
			"AND EXISTS (SELECT 1 FROM of_dishes AS dishes "+
			"WHERE dishes.owner_id = users.id AND dishes.created_at <= ? AND dishes.deleted_at IS NULL)",
		start, end, end,
	).Scan(&created).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "calculate first dish creation rate")
	}
	return math.Round(float64(created)*10000/float64(newUsers)) / 100, nil
}

// dashboardPointSum 汇总指定类型积分流水的绝对值。
func dashboardPointSum(
	ctx context.Context,
	db *gorm.DB,
	start time.Time,
	end time.Time,
	entryType engagementModel.PointEntryType,
) (int64, error) {
	var value sql.NullInt64
	if err := db.WithContext(ctx).Model(&engagementModel.FrontPointEntry{}).
		Select("COALESCE(SUM(ABS(amount)), 0)").
		Where("created_at >= ? AND created_at <= ? AND type = ?", start, end, entryType).
		Scan(&value).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "sum dashboard points")
	}
	if !value.Valid {
		return 0, nil
	}
	return value.Int64, nil
}

// dashboardAIMetrics 计算AI成功率、失败率、平均耗时和估算成本。
func dashboardAIMetrics(
	ctx context.Context,
	db *gorm.DB,
	rangeCode string,
	start time.Time,
	end time.Time,
) ([]dashboardResponse.DashboardMetric, error) {
	var statusRows []struct {
		ExecutionStatus string // 执行状态
		Count           int64  // 调用数量
	}
	if err := db.WithContext(ctx).Model(&aiModel.FrontFeatureUsage{}).
		Select("execution_status, COUNT(*) AS count").
		Where("created_at >= ? AND created_at <= ? AND execution_status IN ?", start, end, []string{"succeeded", "failed"}).
		Group("execution_status").
		Scan(&statusRows).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "count dashboard AI statuses")
	}
	var succeeded int64
	var failed int64
	for _, row := range statusRows {
		switch row.ExecutionStatus {
		case "succeeded":
			succeeded = row.Count
		case "failed":
			failed = row.Count
		}
	}
	var successRate interface{}
	var failureRate interface{}
	if total := succeeded + failed; total > 0 {
		successRate = math.Round(float64(succeeded)*10000/float64(total)) / 100
		failureRate = math.Round(float64(failed)*10000/float64(total)) / 100
	}
	var average sql.NullFloat64
	if err := db.WithContext(ctx).Model(&aiModel.FrontFeatureUsage{}).
		Select("AVG(duration_ms)").
		Where("created_at >= ? AND created_at <= ? AND duration_ms IS NOT NULL AND execution_status IN ?",
			start, end, []string{"succeeded", "failed"}).
		Scan(&average).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "average dashboard AI duration")
	}
	var averageValue interface{}
	if average.Valid {
		averageValue = int64(math.Round(average.Float64))
	}
	var cost sql.NullFloat64
	if err := db.WithContext(ctx).Model(&aiModel.FrontFeatureUsage{}).
		Select("COALESCE(SUM(estimated_cost_cny), 0)").
		Where("created_at >= ? AND created_at <= ?", start, end).
		Scan(&cost).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "sum dashboard AI cost")
	}
	costValue := "0.000000"
	if cost.Valid {
		costValue = fmt.Sprintf("%.6f", cost.Float64)
	}
	query := map[string]interface{}{"range": rangeCode}
	return []dashboardResponse.DashboardMetric{
		dashboardMetric("ai_success_rate", "AI 调用成功率", successRate, "%", "OrderFoodAIUsages", query),
		dashboardMetric("ai_failure_rate", "AI 调用失败率", failureRate, "%", "OrderFoodAIUsages", query),
		dashboardMetric("ai_average_duration_ms", "AI 平均耗时", averageValue, "毫秒", "OrderFoodAIUsages", query),
		dashboardMetric("ai_estimated_cost_cny", "AI 估算成本", costValue, "元", "OrderFoodAIUsages", query),
	}, nil
}

// dashboardCapabilityState 读取当前平台总开关并计算实际生效用户数。
func dashboardCapabilityState(
	ctx context.Context,
	db *gorm.DB,
) (aiResponse.PlatformCapabilitySnapshot, error) {
	var normalUsers int64
	if err := db.WithContext(ctx).Model(&userModel.MiniAppUser{}).
		Where("status = ?", userModel.UserStatusNormal).
		Count(&normalUsers).Error; err != nil {
		return aiResponse.PlatformCapabilitySnapshot{},
			appErrors.AdminInternal.Wrap(err, "count dashboard normal users")
	}
	var policy aiModel.PlatformCapabilityPolicy
	err := db.WithContext(ctx).
		Where("singleton_key = ?", serviceCommon.PlatformPolicySingletonKey).
		First(&policy).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return aiResponse.PlatformCapabilitySnapshot{},
			appErrors.AdminInternal.Wrap(err, "load dashboard platform policy")
	}
	state := aiResponse.PlatformCapabilitySnapshot{
		PlatformDefaultEnabled: policy.PlatformDefaultEnabled,
		EmergencyDisabled:      policy.EmergencyDisabled,
		NormalUserCount:        normalUsers, PolicyVersion: policy.Version,
	}
	if state.PlatformDefaultEnabled && !state.EmergencyDisabled {
		state.EffectiveEnabledUserCount = normalUsers
	}
	return state, nil
}

// dashboardConsecutiveFailures 统计最近连续三次失败的供应商数和模型数。
func dashboardConsecutiveFailures(
	ctx context.Context,
	db *gorm.DB,
) (int64, int64, error) {
	var rows []dashboardFailureRow
	if err := db.WithContext(ctx).Model(&aiModel.FrontFeatureUsage{}).
		Select("provider_id, model_id, execution_status").
		Where("provider_id <> '' AND model_id <> '' AND execution_status IN ?", []string{"succeeded", "failed"}).
		Order("created_at DESC").
		Limit(2000).
		Scan(&rows).Error; err != nil {
		return 0, 0, appErrors.AdminInternal.Wrap(err, "load consecutive AI failures")
	}
	providerCounts := dashboardLeadingFailureCounts(rows, func(row dashboardFailureRow) string {
		return row.ProviderID
	})
	modelCounts := dashboardLeadingFailureCounts(rows, func(row dashboardFailureRow) string {
		return row.ModelID
	})
	return dashboardFailureThresholdCount(providerCounts, 3),
		dashboardFailureThresholdCount(modelCounts, 3), nil
}

// dashboardLeadingFailureCounts 统计每个实体从最新记录开始的连续失败次数。
func dashboardLeadingFailureCounts(
	rows []dashboardFailureRow,
	key func(dashboardFailureRow) string,
) map[string]int {
	counts := make(map[string]int)
	stopped := make(map[string]bool)
	for _, row := range rows {
		value := strings.TrimSpace(key(row))
		if value == "" || stopped[value] {
			continue
		}
		if row.ExecutionStatus == "failed" {
			counts[value]++
			continue
		}
		stopped[value] = true
	}
	return counts
}

// dashboardFailureThresholdCount 计算达到连续失败阈值的实体数量。
func dashboardFailureThresholdCount(counts map[string]int, threshold int) int64 {
	var total int64
	for _, count := range counts {
		if count >= threshold {
			total++
		}
	}
	return total
}

// dashboardMetric 创建可跳转的指标响应。
func dashboardMetric(
	key string,
	label string,
	value interface{},
	unit string,
	targetPage string,
	targetQuery map[string]interface{},
) dashboardResponse.DashboardMetric {
	var target *string
	if strings.TrimSpace(targetPage) != "" {
		value := targetPage
		target = &value
	}
	if targetQuery == nil {
		targetQuery = map[string]interface{}{}
	}
	return dashboardResponse.DashboardMetric{
		Key: key, Label: label, Value: value, Unit: unit,
		TargetPage: target, TargetQuery: targetQuery,
	}
}

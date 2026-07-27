package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// migrateEngagementTestTables 迁移积分调用测试所需的 MySQL 表。
func migrateEngagementTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.PlatformCapabilityPolicy{},
		&orderfoodModel.AICapabilityDefinition{},
		&orderfoodModel.AICapabilityConfig{},
		&orderfoodModel.SubscribeMessageTemplate{},
		&orderfoodModel.FrontFeatureUsage{},
		&orderfoodModel.FrontPointEntry{},
		&orderfoodModel.UserNotification{},
		&orderfoodModel.FrontMediaAsset{},
		&orderfoodModel.FrontCheckin{},
		&orderfoodModel.PointRuleConfig{},
		&orderfoodModel.UserPreferenceProfile{},
		&orderfoodModel.PreferenceEvidence{},
		&orderfoodModel.PreferenceEvidenceAggregate{},
	); err != nil {
		t.Fatalf("migrate engagement test tables: %v", err)
	}
}

// createEngagementTestUser 创建具备固定初始积分的测试用户。
func createEngagementTestUser(t *testing.T, db *gorm.DB, now time.Time) orderfoodModel.MiniAppUser {
	t.Helper()
	user := orderfoodModel.MiniAppUser{
		ID:              "engagement-user",
		OpenIDHash:      "engagement-openid-hash",
		OpenIDEncrypted: "engagement-openid-encrypted",
		Nickname:        "积分测试用户",
		Status:          orderfoodModel.UserStatusNormal,
		Points:          100,
		Version:         1,
		RegisteredAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create engagement test user: %v", err)
	}
	return user
}

// seedEngagementCapability 配置一个每日免费一次、总限额两次的推荐能力。
func seedEngagementCapability(t *testing.T, db *gorm.DB, now time.Time) {
	t.Helper()
	labels, err := json.Marshal([]clientFeatureLabel{{
		Code:        "meal_suggest",
		Title:       "不知道吃什么",
		ActionLabel: "生成建议",
		Description: "测试能力",
		SortOrder:   1,
	}})
	if err != nil {
		t.Fatalf("encode engagement feature labels: %v", err)
	}
	if err = db.Create(&orderfoodModel.PlatformCapabilityPolicy{
		SingletonKey:           "platform",
		Version:                1,
		PlatformDefaultEnabled: true,
		EmergencyDisabled:      false,
		FeatureLabelsJSON:      datatypes.JSON(labels),
		AppliedByID:            1,
		AppliedByUsername:      "root",
		AppliedAt:              now,
		Reason:                 "测试",
	}).Error; err != nil {
		t.Fatalf("create engagement platform policy: %v", err)
	}
	if err = db.Create(&orderfoodModel.AICapabilityDefinition{
		Code:                    orderfoodModel.AICapabilityMealSuggest,
		Name:                    "不知道吃什么",
		ClientFeatureCode:       "meal_suggest",
		RequiredModelCapability: orderfoodModel.AIModelText,
		SortOrder:               1,
		CreatedAt:               now,
		UpdatedAt:               now,
	}).Error; err != nil {
		t.Fatalf("create engagement capability definition: %v", err)
	}
	if err = db.Create(&orderfoodModel.AICapabilityConfig{
		CapabilityCode:              orderfoodModel.AICapabilityMealSuggest,
		Version:                     1,
		PrimaryModelID:              "model-text",
		PointCost:                   5,
		DailyLimitPerUser:           2,
		TimeoutMS:                   30000,
		FreeQuotaPerDay:             1,
		PromptMode:                  orderfoodModel.AIPromptPreset,
		PromptPresetVersion:         1,
		SystemPrompt:                "system",
		UserPromptTemplate:          "{{input}}",
		PromptAllowedVariablesJSON:  datatypes.JSON([]byte(`["input"]`)),
		PromptRequiredVariablesJSON: datatypes.JSON([]byte(`["input"]`)),
		PromptOutputSchemaVersion:   "v1",
		PromptHash:                  "hash",
		AppliedByID:                 1,
		AppliedByUsername:           "root",
		AppliedAt:                   now,
		Reason:                      "测试",
	}).Error; err != nil {
		t.Fatalf("create engagement capability config: %v", err)
	}
}

// TestCheckinRewardFollowsPlatformSwitchAndCurrentPointRule 验证关闭总开关仍保存打卡，开启后按当前规则奖励。
func TestCheckinRewardFollowsPlatformSwitchAndCurrentPointRule(t *testing.T) {
	db := testutil.OpenMySQL(t)
	migrateEngagementTestTables(t, db)
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	disabledUser := createEngagementTestUser(t, db, now)
	if err := db.Create(&orderfoodModel.PlatformCapabilityPolicy{
		SingletonKey: "platform", Version: 1,
		PlatformDefaultEnabled: false, EmergencyDisabled: false,
		FeatureLabelsJSON: datatypes.JSON([]byte(`[]`)),
		AppliedByID:       1, AppliedByUsername: "root",
		AppliedAt: now, Reason: "关闭整体能力测试",
	}).Error; err != nil {
		t.Fatalf("create disabled platform policy: %v", err)
	}
	disabledAsset := orderfoodModel.FrontMediaAsset{
		ID: "checkin-disabled-asset", UserID: disabledUser.ID,
		FileName: "disabled.jpg", UploadSource: "user_upload",
		Scene: "checkin", ResourceStatus: "active",
		URL: "https://example.test/disabled.jpg", StoragePath: "test/disabled.jpg",
		ContentType: "image/jpeg", Width: 100, Height: 100, SizeBytes: 100,
		Checksum: "disabled-checksum", ReviewStatus: orderfoodModel.MediaReviewPassed,
		CreatedAt: now,
	}
	if err := db.Create(&disabledAsset).Error; err != nil {
		t.Fatalf("create disabled checkin asset: %v", err)
	}
	service := &EngagementService{
		DB: db, Runtime: &RuntimeService{DB: db}, Now: func() time.Time { return now },
	}
	checkin, rewarded, rewardAmount, enhancedEnabled, err := service.CreateCheckin(
		context.Background(),
		disabledUser.ID,
		frontRequest.CheckinInput{
			DishName: "关闭时打卡", ImageFileID: disabledAsset.ID,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if checkin.ID == "" || rewarded || rewardAmount != 0 || enhancedEnabled {
		t.Fatalf(
			"disabled checkin = %+v, rewarded=%v amount=%d enabled=%v",
			checkin, rewarded, rewardAmount, enhancedEnabled,
		)
	}
	var disabledUserAfter orderfoodModel.MiniAppUser
	if err := db.First(&disabledUserAfter, "id = ?", disabledUser.ID).Error; err != nil {
		t.Fatal(err)
	}
	if disabledUserAfter.Points != disabledUser.Points ||
		disabledUserAfter.CheckinCount != 1 ||
		disabledUserAfter.CheckinDayCount != 1 {
		t.Fatalf("disabled user after checkin = %+v", disabledUserAfter)
	}
	var disabledEntryCount int64
	if err := db.Model(&orderfoodModel.FrontPointEntry{}).
		Where("user_id = ?", disabledUser.ID).
		Count(&disabledEntryCount).Error; err != nil {
		t.Fatal(err)
	}
	if disabledEntryCount != 0 {
		t.Fatalf("disabled point entry count = %d, want 0", disabledEntryCount)
	}

	if err := db.Model(&orderfoodModel.PlatformCapabilityPolicy{}).
		Where("singleton_key = ?", "platform").
		Updates(map[string]interface{}{
			"version": 2, "platform_default_enabled": true,
			"applied_at": now.Add(time.Minute), "reason": "开启整体能力测试",
		}).Error; err != nil {
		t.Fatalf("enable platform policy: %v", err)
	}
	if err := db.Create(&orderfoodModel.PointRuleConfig{
		SingletonKey: "platform", Version: 1, DailyCheckinReward: 7,
		AppliedByID: 1, AppliedByUsername: "root",
		AppliedAt: now, Reason: "当前打卡奖励规则",
	}).Error; err != nil {
		t.Fatalf("create point rule: %v", err)
	}
	enabledUser := orderfoodModel.MiniAppUser{
		ID:              "engagement-enabled-user",
		OpenIDHash:      "engagement-enabled-openid-hash",
		OpenIDEncrypted: "engagement-enabled-openid-encrypted",
		Nickname:        "开启能力用户", Status: orderfoodModel.UserStatusNormal,
		Points: 100, Version: 1,
		RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&enabledUser).Error; err != nil {
		t.Fatalf("create enabled user: %v", err)
	}
	enabledAsset := disabledAsset
	enabledAsset.ID = "checkin-enabled-asset"
	enabledAsset.UserID = enabledUser.ID
	enabledAsset.FileName = "enabled.jpg"
	enabledAsset.URL = "https://example.test/enabled.jpg"
	enabledAsset.StoragePath = "test/enabled.jpg"
	enabledAsset.Checksum = "enabled-checksum"
	if err := db.Create(&enabledAsset).Error; err != nil {
		t.Fatalf("create enabled checkin asset: %v", err)
	}
	_, rewarded, rewardAmount, enhancedEnabled, err = service.CreateCheckin(
		context.Background(),
		enabledUser.ID,
		frontRequest.CheckinInput{
			DishName: "开启后打卡", ImageFileID: enabledAsset.ID,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !rewarded || rewardAmount != 7 || !enhancedEnabled {
		t.Fatalf(
			"enabled checkin rewarded=%v amount=%d enabled=%v",
			rewarded, rewardAmount, enhancedEnabled,
		)
	}
	var enabledUserAfter orderfoodModel.MiniAppUser
	if err := db.First(&enabledUserAfter, "id = ?", enabledUser.ID).Error; err != nil {
		t.Fatal(err)
	}
	if enabledUserAfter.Points != 107 {
		t.Fatalf("enabled user points = %d, want 107", enabledUserAfter.Points)
	}
	var rewardEntry orderfoodModel.FrontPointEntry
	if err := db.First(
		&rewardEntry,
		"user_id = ? AND scene = ?",
		enabledUser.ID,
		"daily_checkin",
	).Error; err != nil {
		t.Fatal(err)
	}
	if rewardEntry.Amount != 7 || rewardEntry.BalanceAfter != 107 {
		t.Fatalf("reward entry = %+v", rewardEntry)
	}
}

// TestShanghaiDayRangeUsesBusinessTimezone 验证免费额度按上海自然日切换。
func TestShanghaiDayRangeUsesBusinessTimezone(t *testing.T) {
	start, end := shanghaiDayRange(time.Date(2026, time.July, 26, 16, 30, 0, 0, time.UTC))
	if want := time.Date(2026, time.July, 26, 16, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("Shanghai day start = %s, want %s", start, want)
	}
	if want := time.Date(2026, time.July, 27, 16, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("Shanghai day end = %s, want %s", end, want)
	}
}

// TestFeatureQuotaReservesProcessingUsageAndEnforcesDailyLimit 验证处理中调用占位且日限额真实生效。
func TestFeatureQuotaReservesProcessingUsageAndEnforcesDailyLimit(t *testing.T) {
	db := testutil.OpenMySQL(t)
	migrateEngagementTestTables(t, db)
	now := time.Date(2026, time.July, 26, 18, 0, 0, 0, time.UTC)
	user := createEngagementTestUser(t, db, now)
	seedEngagementCapability(t, db, now)
	service := &EngagementService{
		DB:      db,
		Runtime: &RuntimeService{DB: db},
		Now:     func() time.Time { return now },
	}

	first, _, err := service.BeginFeatureUsage(
		context.Background(),
		user.ID,
		orderfoodModel.AICapabilityMealSuggest,
		true,
	)
	if err != nil || first.PointCost != 0 {
		t.Fatalf("first free usage = %+v, err=%v", first, err)
	}
	second, balance, err := service.BeginFeatureUsage(
		context.Background(),
		user.ID,
		orderfoodModel.AICapabilityMealSuggest,
		true,
	)
	if err != nil || second.PointCost != 5 || balance != 95 {
		t.Fatalf("second paid usage = %+v, balance=%d, err=%v", second, balance, err)
	}
	if _, _, err = service.BeginFeatureUsage(
		context.Background(),
		user.ID,
		orderfoodModel.AICapabilityMealSuggest,
		true,
	); appErrors.Code(err) != int(appErrors.FrontQuotaExceeded) {
		t.Fatalf("third usage error code = %d, want %d", appErrors.Code(err), appErrors.FrontQuotaExceeded)
	}

	// 免费调用失败后不占用成功额度，但仍不会改写另一笔处理中的付费调用。
	if _, err = service.FinishFeatureUsage(context.Background(), first, false); err != nil {
		t.Fatalf("finish first usage as failed: %v", err)
	}
	replacement, _, err := service.BeginFeatureUsage(
		context.Background(),
		user.ID,
		orderfoodModel.AICapabilityMealSuggest,
		true,
	)
	if err != nil || replacement.PointCost != 0 {
		t.Fatalf("replacement free usage = %+v, err=%v", replacement, err)
	}
}

// TestFeatureFailureRefundIsIdempotentAndNeverCreatesSubscription 验证重复收尾不会重复退款或创建订阅消息。
func TestFeatureFailureRefundIsIdempotentAndNeverCreatesSubscription(t *testing.T) {
	db := testutil.OpenMySQL(t)
	migrateEngagementTestTables(t, db)
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	user := createEngagementTestUser(t, db, now)
	user.Points = 90
	if err := db.Model(&user).Update("points", user.Points).Error; err != nil {
		t.Fatalf("set charged balance: %v", err)
	}
	objectType := "feature_usage"
	usage := orderfoodModel.FrontFeatureUsage{
		ID:               "usage-refund",
		UserID:           user.ID,
		Feature:          "meal_suggest",
		CapabilityCode:   orderfoodModel.AICapabilityMealSuggest,
		PointCost:        10,
		ExecutionStatus:  orderfoodModel.FeatureExecutionProcessing,
		BillingStatus:    orderfoodModel.FeatureBillingCharged,
		EstimatedCostCNY: "0.000000",
		Version:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(&usage).Error; err != nil {
		t.Fatalf("create charged feature usage: %v", err)
	}
	if err := db.Create(&orderfoodModel.FrontPointEntry{
		ID:                "charge-entry",
		UserID:            user.ID,
		Type:              orderfoodModel.PointSpent,
		Scene:             "feature_usage",
		Title:             "增强功能使用",
		Description:       usage.Feature,
		Amount:            -10,
		BalanceAfter:      90,
		RelatedObjectType: &objectType,
		RelatedObjectID:   &usage.ID,
		CreatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create feature charge entry: %v", err)
	}
	service := &EngagementService{DB: db, Now: func() time.Time { return now.Add(time.Second) }}
	for attempt := 0; attempt < 2; attempt++ {
		result, err := service.FinishFeatureUsage(context.Background(), usage, false)
		if err != nil || result.BillingStatus != orderfoodModel.FeatureBillingRefunded ||
			result.PointBalance != 100 {
			t.Fatalf("refund attempt %d result=%+v err=%v", attempt+1, result, err)
		}
	}

	var refundCount int64
	if err := db.Model(&orderfoodModel.FrontPointEntry{}).
		Where("type = ? AND related_object_id = ?", orderfoodModel.PointRefund, usage.ID).
		Count(&refundCount).Error; err != nil {
		t.Fatalf("count feature refunds: %v", err)
	}
	if refundCount != 1 {
		t.Fatalf("refund entry count = %d, want 1", refundCount)
	}
	var notification orderfoodModel.UserNotification
	if err := db.First(&notification, "type = ? AND target_id = ?", "feature_refund", usage.ID).Error; err != nil {
		t.Fatalf("load feature refund notification: %v", err)
	}
	if notification.SubscribeRequired || notification.SubscribeLogID != nil {
		t.Fatalf("refund notification unexpectedly requires subscription: %+v", notification)
	}
}

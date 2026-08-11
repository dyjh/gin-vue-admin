package engagement

import (
	"context"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	engagementRequest "github.com/dyjh/order-food-mini-app/server/model/engagement/request"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// newPointsTestService 创建积分规则服务及其独立 MySQL 测试数据库。
func newPointsTestService(t *testing.T) (*PointsService, *gorm.DB) {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&engagementModel.PointRuleConfig{},
		&commonModel.AdminIdempotencyRecord{},
		&auditModel.AdminAuditLog{},
		&userModel.MiniAppUser{},
		&engagementModel.FrontPointEntry{},
	); err != nil {
		t.Fatal(err)
	}
	service := NewPointsService(
		db,
		serviceCommon.NewPermissionService(db),
		serviceCommon.NewIdempotencyService(db),
	)
	service.Now = func() time.Time {
		return time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)
	}
	return service, db
}

// TestListPointEntriesFiltersAndReturnsUserMySQL 校验积分流水组合筛选和用户摘要返回。
func TestListPointEntriesFiltersAndReturnsUserMySQL(t *testing.T) {
	service, db := newPointsTestService(t)
	now := service.now()
	users := []userModel.MiniAppUser{
		{
			ID: "user_point_list_target", OpenIDHash: "openid-hash-point-list-target",
			OpenIDEncrypted: "encrypted-openid-target", Nickname: "目标用户",
			Status: userModel.UserStatusNormal, Version: 1,
			RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "user_point_list_other", OpenIDHash: "openid-hash-point-list-other",
			OpenIDEncrypted: "encrypted-openid-other", Nickname: "其他用户",
			Status: userModel.UserStatusNormal, Version: 1,
			RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
		},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	objectType := "feature_usage"
	targetObjectID := "usage_point_list_target"
	otherObjectID := "usage_point_list_other"
	requestID := "request-point-list-target"
	idempotencyKey := "idempotency-point-list-target"
	entries := []engagementModel.FrontPointEntry{
		{
			ID: "point_list_target", UserID: users[0].ID,
			Type: engagementModel.PointSpent, Scene: "feature_usage",
			Title: "增强功能使用", Description: "目标流水",
			Amount: -2, BalanceAfter: 8,
			RelatedObjectType: &objectType, RelatedObjectID: &targetObjectID,
			RequestID: &requestID, IdempotencyKey: &idempotencyKey,
			CreatedAt: now.Add(-time.Hour),
		},
		{
			ID: "point_list_wrong_object", UserID: users[0].ID,
			Type: engagementModel.PointSpent, Scene: "feature_usage",
			Title: "增强功能使用", Description: "其他关联对象",
			Amount: -1, BalanceAfter: 9,
			RelatedObjectType: &objectType, RelatedObjectID: &otherObjectID,
			CreatedAt: now.Add(-30 * time.Minute),
		},
		{
			ID: "point_list_wrong_user", UserID: users[1].ID,
			Type: engagementModel.PointSpent, Scene: "feature_usage",
			Title: "增强功能使用", Description: "其他用户",
			Amount: -2, BalanceAfter: 8,
			RelatedObjectType: &objectType, RelatedObjectID: &targetObjectID,
			RequestID: &requestID, IdempotencyKey: &idempotencyKey,
			CreatedAt: now.Add(-time.Hour),
		},
	}
	if err := db.Create(&entries).Error; err != nil {
		t.Fatal(err)
	}
	from := now.Add(-2 * time.Hour)
	to := now.Add(-time.Minute)
	result, err := service.ListPointEntries(
		context.Background(),
		engagementRequest.PointEntryListQuery{
			UserID: users[0].ID, Type: string(engagementModel.PointSpent),
			Scene: "feature_usage", RelatedObjectType: objectType,
			RelatedObjectID: targetObjectID, RequestID: requestID,
			IdempotencyKey: idempotencyKey, CreatedFrom: &from, CreatedTo: &to,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || len(result.List) != 1 {
		t.Fatalf("point entry result = %+v, want exactly one row", result)
	}
	if result.List[0].ID != entries[0].ID ||
		result.List[0].User.ID != users[0].ID ||
		result.List[0].User.Nickname != users[0].Nickname {
		t.Fatalf("point entry = %+v, want target row and user summary", result.List[0])
	}
}

// TestPointAdjustmentPreviewCreateAndReplay 校验积分调整预览、事务记账和幂等回放。
func TestPointAdjustmentPreviewCreateAndReplay(t *testing.T) {
	service, db := newPointsTestService(t)
	previousSigningKey := global.GVA_CONFIG.JWT.SigningKey
	global.GVA_CONFIG.JWT.SigningKey = "point-adjustment-test-signing-key"
	t.Cleanup(func() {
		global.GVA_CONFIG.JWT.SigningKey = previousSigningKey
	})
	now := service.now()
	user := userModel.MiniAppUser{
		ID:              "user_points_adjustment",
		OpenIDHash:      "openid-hash-points-adjustment",
		OpenIDEncrypted: "encrypted-openid",
		Nickname:        "积分用户",
		Points:          20,
		Status:          userModel.UserStatusNormal,
		Version:         1,
		RegisteredAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     commonModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-point-adjustment",
	}
	previewInput := engagementRequest.PointAdjustmentPreviewInput{
		UserID: user.ID, Direction: "debit", Amount: 8, Reason: "修正重复发放积分",
	}
	preview, err := service.PreviewPointAdjustment(
		context.Background(), actor, previewInput,
	)
	if err != nil {
		t.Fatal(err)
	}
	if preview.BalanceBefore != 20 || preview.BalanceAfter != 12 ||
		preview.ExpectedUserVersion != 1 || preview.PreviewToken == "" {
		t.Fatalf("unexpected point adjustment preview: %+v", preview)
	}
	input := engagementRequest.PointAdjustmentCreateInput{
		PreviewToken: preview.PreviewToken, UserID: user.ID,
		Direction: "debit", Amount: 8, Reason: "修正重复发放积分",
		ExpectedUserVersion: preview.ExpectedUserVersion,
	}
	result, replayed, err := service.CreatePointAdjustment(
		context.Background(), actor, "point-adjustment-idempotency", input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || result.BalanceAfter != 12 || result.PointEntry.Amount != -8 ||
		result.PointEntry.BalanceAfter != 12 || result.UserVersion != 2 {
		t.Fatalf("unexpected point adjustment result: %+v, replayed=%v", result, replayed)
	}
	replayedResult, replayed, err := service.CreatePointAdjustment(
		context.Background(), actor, "point-adjustment-idempotency", input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !replayed || replayedResult.PointEntry.ID != result.PointEntry.ID {
		t.Fatalf("idempotent replay = %+v, replayed=%v", replayedResult, replayed)
	}
	var entryCount int64
	if err := db.Model(&engagementModel.FrontPointEntry{}).
		Where("user_id = ?", user.ID).Count(&entryCount).Error; err != nil {
		t.Fatal(err)
	}
	if entryCount != 1 {
		t.Fatalf("point entry count = %d, want 1", entryCount)
	}
}

// TestPointAdjustmentRejectsStaleOrForgedPreviewState 校验旧余额和伪造版本不能复用预览令牌。
func TestPointAdjustmentRejectsStaleOrForgedPreviewState(t *testing.T) {
	service, db := newPointsTestService(t)
	previousSigningKey := global.GVA_CONFIG.JWT.SigningKey
	global.GVA_CONFIG.JWT.SigningKey = "point-adjustment-state-binding-test-key"
	t.Cleanup(func() {
		global.GVA_CONFIG.JWT.SigningKey = previousSigningKey
	})
	now := service.now()
	users := []userModel.MiniAppUser{
		{
			ID: "user_stale_point_balance", OpenIDHash: "openid-hash-stale-point-balance",
			OpenIDEncrypted: "encrypted-stale-point-balance", Nickname: "余额变化用户",
			Points: 20, Status: userModel.UserStatusNormal, Version: 1,
			RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "user_forged_point_version", OpenIDHash: "openid-hash-forged-point-version",
			OpenIDEncrypted: "encrypted-forged-point-version", Nickname: "版本变化用户",
			Points: 20, Status: userModel.UserStatusNormal, Version: 1,
			RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
		},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     commonModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-point-adjustment-state-binding",
	}
	previewInput := engagementRequest.PointAdjustmentPreviewInput{
		UserID: users[0].ID, Direction: "debit", Amount: 8, Reason: "校验预览状态绑定",
	}
	staleBalancePreview, err := service.PreviewPointAdjustment(
		context.Background(), actor, previewInput,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&userModel.MiniAppUser{}).
		Where("id = ?", users[0].ID).
		Update("points", 21).Error; err != nil {
		t.Fatal(err)
	}
	_, _, err = service.CreatePointAdjustment(
		context.Background(),
		actor,
		"point-adjustment-stale-balance",
		engagementRequest.PointAdjustmentCreateInput{
			PreviewToken: staleBalancePreview.PreviewToken,
			UserID:       users[0].ID, Direction: "debit", Amount: 8,
			Reason:              "校验预览状态绑定",
			ExpectedUserVersion: staleBalancePreview.ExpectedUserVersion,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminStateConflict {
		t.Fatalf("stale balance error = %v, want state conflict", err)
	}

	previewInput.UserID = users[1].ID
	forgedVersionPreview, err := service.PreviewPointAdjustment(
		context.Background(), actor, previewInput,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&userModel.MiniAppUser{}).
		Where("id = ?", users[1].ID).
		Update("version", 2).Error; err != nil {
		t.Fatal(err)
	}
	_, _, err = service.CreatePointAdjustment(
		context.Background(),
		actor,
		"point-adjustment-forged-version",
		engagementRequest.PointAdjustmentCreateInput{
			PreviewToken: forgedVersionPreview.PreviewToken,
			UserID:       users[1].ID, Direction: "debit", Amount: 8,
			Reason:              "校验预览状态绑定",
			ExpectedUserVersion: 2,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminStateConflict {
		t.Fatalf("forged version error = %v, want state conflict", err)
	}

	var entryCount int64
	if err := db.Model(&engagementModel.FrontPointEntry{}).
		Where("user_id IN ?", []string{users[0].ID, users[1].ID}).
		Count(&entryCount).Error; err != nil {
		t.Fatal(err)
	}
	if entryCount != 0 {
		t.Fatalf("point entry count = %d, want 0", entryCount)
	}
}

// TestPointRuleUpdateAppliesImmediatelyAndRuntimeResolution 验证积分规则保存后立即成为运行时规则。
func TestPointRuleUpdateAppliesImmediatelyAndRuntimeResolution(t *testing.T) {
	service, db := newPointsTestService(t)
	ctx := context.Background()
	config, err := service.CurrentPointRule(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.Version != 1 || config.DailyCheckinReward != 1 {
		t.Fatalf("default config = %+v", config)
	}

	actor := commonRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     commonModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-save-point-rule",
	}
	updated, replayed, err := service.UpdatePointRule(
		ctx,
		actor,
		"update-point-rule",
		engagementRequest.PointRuleUpdateInput{
			DailyCheckinReward: 5,
			Reason:             "调整每日打卡奖励",
			ExpectedVersion:    config.Version,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || updated.Version != 2 || updated.DailyCheckinReward != 5 {
		t.Fatalf("updated config = %+v, replayed = %v", updated, replayed)
	}
	current, err := CurrentDailyCheckinReward(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if current != 5 {
		t.Fatalf("runtime reward = %d, want 5", current)
	}
	replayedConfig, replayed, err := service.UpdatePointRule(
		ctx,
		actor,
		"update-point-rule",
		engagementRequest.PointRuleUpdateInput{
			DailyCheckinReward: 5,
			Reason:             "调整每日打卡奖励",
			ExpectedVersion:    config.Version,
		},
	)
	if err != nil || !replayed || replayedConfig.Version != updated.Version {
		t.Fatalf("replay update = %+v, replayed=%v err=%v", replayedConfig, replayed, err)
	}

	var auditCount int64
	if err := db.Model(&auditModel.AdminAuditLog{}).
		Where("target_type = ?", "point_rule").
		Count(&auditCount).Error; err != nil {
		t.Fatal(err)
	}
	if auditCount != 1 {
		t.Fatalf("point rule audit count = %d, want 1", auditCount)
	}
	var configCount int64
	if err := db.Model(&engagementModel.PointRuleConfig{}).Count(&configCount).Error; err != nil {
		t.Fatal(err)
	}
	if configCount != 1 {
		t.Fatalf("point rule config count = %d, want 1", configCount)
	}
}

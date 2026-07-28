package service

import (
	"context"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// openSubscriptionTestDB 创建消息中心服务测试所需的独立MySQL数据库。
func openSubscriptionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.UserNotification{},
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.SubscribeMessageTemplate{},
		&orderfoodModel.SubscribeMessageLog{},
		&orderfoodModel.SubscribeMessageAttempt{},
	); err != nil {
		t.Fatalf("migrate subscription models: %v", err)
	}
	return db
}

// subscriptionTestActor 返回消息中心配置测试使用的管理员上下文。
func subscriptionTestActor() orderfoodRequest.AdminActor {
	return orderfoodRequest.AdminActor{
		AdministratorID:  9,
		AuthorityID:      orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:         "subscription-admin",
		RequestID:        "request-subscription-test",
		SourceIPMasked:   "127.0.0.0",
		UserAgentSummary: "subscription-service-test",
	}
}

// TestNotificationAdminReadOnlyQueryAndDetail 验证站内通知筛选、生成来源和只读详情。
func TestNotificationAdminReadOnlyQueryAndDetail(t *testing.T) {
	db := openSubscriptionTestDB(t)
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	user := orderfoodModel.MiniAppUser{
		ID:              "notification-user",
		OpenIDHash:      strings.Repeat("n", 64),
		OpenIDEncrypted: "encrypted-openid",
		Nickname:        "通知测试用户",
		Status:          orderfoodModel.UserStatusNormal,
		Version:         1,
		RegisteredAt:    now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create notification user: %v", err)
	}
	targetType := "dish"
	targetID := "dish-001"
	rows := []orderfoodModel.UserNotification{
		{
			ID: "notice-governance", UserID: user.ID, Type: "governance",
			Title: "菜品处理通知", Content: "测试正文",
			TargetType: &targetType, TargetID: &targetID, CreatedAt: now,
		},
		{
			ID: "notice-meal", UserID: user.ID, Type: "meal",
			Title: "饭局已完成", Content: "饭局结果",
			ReadAt: &now, CreatedAt: now.Add(time.Minute),
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("create notifications: %v", err)
	}
	service := NewSubscriptionService(db, nil)
	result, err := service.ListNotifications(
		context.Background(),
		orderfoodRequest.NotificationAdminListQuery{
			UserID: user.ID, Type: "governance", Read: boolPointer(false),
		},
	)
	if err != nil || result.Total != 1 || len(result.List) != 1 ||
		result.List[0].ID != rows[0].ID {
		t.Fatalf("list notification result=%+v err=%v", result, err)
	}
	detail, err := service.GetNotification(context.Background(), rows[0].ID)
	if err != nil || detail.GenerationSource != "违规处理" ||
		detail.Content != rows[0].Content || detail.Read {
		t.Fatalf("notification detail=%+v err=%v", detail, err)
	}
	var persisted orderfoodModel.UserNotification
	if err := db.First(&persisted, "id = ?", rows[0].ID).Error; err != nil {
		t.Fatalf("reload notification: %v", err)
	}
	if persisted.ReadAt != nil {
		t.Fatal("admin detail changed the user's read state")
	}
}

// TestNotificationAdminRejectsDirectInvalidInputs 验证服务层拒绝绕过HTTP绑定的非法通知查询参数。
func TestNotificationAdminRejectsDirectInvalidInputs(t *testing.T) {
	service := &SubscriptionService{}
	_, err := service.ListNotifications(
		context.Background(),
		orderfoodRequest.NotificationAdminListQuery{
			SortOrder: "desc; DROP TABLE of_notifications",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe notification sort error = %v", err)
	}
	from := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)
	_, err = service.ListNotifications(
		context.Background(),
		orderfoodRequest.NotificationAdminListQuery{
			CreatedFrom: &from, CreatedTo: &to,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid notification time range error = %v", err)
	}
	_, err = service.GetNotification(context.Background(), "")
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid notification identity error = %v", err)
	}
}

// subscriptionVersionPointer 返回订阅场景配置版本指针。
func subscriptionVersionPointer(value int) *int {
	return &value
}

// TestSubscribeSceneFixedConfiguration 验证场景固定存在且只能维护一个当前模板绑定。
func TestSubscribeSceneFixedConfiguration(t *testing.T) {
	db := openSubscriptionTestDB(t)
	service := NewSubscriptionService(db, NewIdempotencyService(db))
	now := time.Date(2026, time.July, 28, 9, 0, 0, 0, time.UTC)
	service.Now = func() time.Time { return now }
	service.Idempotency.Now = service.Now
	ctx := context.Background()

	scenes, err := service.ListSubscribeScenes(ctx)
	if err != nil || len(scenes) != 1 ||
		scenes[0].Scene != orderfoodModel.SubscribeSceneMealStatus ||
		scenes[0].Configured || scenes[0].Enabled ||
		len(scenes[0].RequiredFields) != 3 {
		t.Fatalf("unconfigured subscription scenes=%+v err=%v", scenes, err)
	}
	unconfigured, err := service.GetSubscribeScene(ctx, orderfoodModel.SubscribeSceneMealStatus)
	if err != nil || unconfigured.Configured || unconfigured.WechatTemplateID != "" ||
		len(unconfigured.FieldMappings) != 0 {
		t.Fatalf("unconfigured subscription scene=%+v err=%v", unconfigured, err)
	}

	configured, replayed, err := service.ConfigureSubscribeScene(
		ctx,
		subscriptionTestActor(),
		"subscription-scene-configure",
		orderfoodModel.SubscribeSceneMealStatus,
		orderfoodRequest.SubscribeSceneConfigureInput{
			WechatTemplateID: "wechat-template-meal-status-001",
			FieldMappings: map[string]string{
				"mealName": "thing1",
				"result":   "phrase2",
				"resultAt": "time3",
			},
		},
	)
	if err != nil || replayed || !configured.Configured || configured.Enabled ||
		configured.TemplateID == nil || configured.Version != 1 ||
		len(configured.FieldMappingSummary) != 3 {
		t.Fatalf("configure subscription scene result=%+v replayed=%v err=%v", configured, replayed, err)
	}

	reconfigured, replayed, err := service.ConfigureSubscribeScene(
		ctx,
		subscriptionTestActor(),
		"subscription-scene-reconfigure",
		orderfoodModel.SubscribeSceneMealStatus,
		orderfoodRequest.SubscribeSceneConfigureInput{
			WechatTemplateID: "wechat-template-meal-status-002",
			FieldMappings: map[string]string{
				"mealName": "thing4",
				"result":   "phrase5",
				"resultAt": "time6",
			},
			ExpectedVersion: subscriptionVersionPointer(configured.Version),
		},
	)
	if err != nil || replayed || reconfigured.TemplateID == nil ||
		*reconfigured.TemplateID != *configured.TemplateID ||
		reconfigured.Version != configured.Version+1 ||
		reconfigured.WechatTemplateID != "wechat-template-meal-status-002" {
		t.Fatalf("reconfigure subscription scene result=%+v replayed=%v err=%v", reconfigured, replayed, err)
	}
	var bindingCount int64
	if err := db.Model(&orderfoodModel.SubscribeMessageTemplate{}).
		Where("scene = ?", orderfoodModel.SubscribeSceneMealStatus).
		Count(&bindingCount).Error; err != nil || bindingCount != 1 {
		t.Fatalf("subscription scene binding count=%d err=%v", bindingCount, err)
	}

	enabled, replayed, err := service.UpdateSubscribeSceneStatus(
		ctx,
		subscriptionTestActor(),
		"subscription-scene-enable",
		orderfoodModel.SubscribeSceneMealStatus,
		orderfoodRequest.SubscribeSceneStatusInput{
			Enabled:         boolPointer(true),
			Reason:          "验证固定订阅场景启用",
			ExpectedVersion: reconfigured.Version,
		},
	)
	if err != nil || replayed || !enabled.Enabled || enabled.Version != reconfigured.Version+1 {
		t.Fatalf("enable subscription scene result=%+v replayed=%v err=%v", enabled, replayed, err)
	}

	_, _, err = service.ConfigureSubscribeScene(
		ctx,
		subscriptionTestActor(),
		"subscription-scene-extra-field",
		orderfoodModel.SubscribeSceneMealStatus,
		orderfoodRequest.SubscribeSceneConfigureInput{
			WechatTemplateID: "wechat-template-invalid",
			FieldMappings: map[string]string{
				"mealName":  "thing1",
				"result":    "phrase2",
				"resultAt":  "time3",
				"extraData": "thing4",
			},
			ExpectedVersion: subscriptionVersionPointer(enabled.Version),
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInvalidConfig {
		t.Fatalf("extra subscription business field error=%v", err)
	}
	if _, err := service.GetSubscribeScene(ctx, "custom_scene"); appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unknown subscription scene error=%v", err)
	}
}

// TestSubscribeTemplateAndLogAdminFlow 验证模板默认停用、列表摘要、发送明细和有引用禁止删除。
func TestSubscribeTemplateAndLogAdminFlow(t *testing.T) {
	db := openSubscriptionTestDB(t)
	service := NewSubscriptionService(db, NewIdempotencyService(db))
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	service.Now = func() time.Time { return now }
	service.Idempotency.Now = service.Now
	ctx := context.Background()
	created, replayed, err := service.CreateSubscribeTemplate(
		ctx,
		subscriptionTestActor(),
		"subscription-template-create",
		orderfoodRequest.SubscribeTemplateCreateInput{
			Name:             "饭局最终结果",
			WechatTemplateID: "wechat-template-meal-status-001",
			Scene:            orderfoodModel.SubscribeSceneMealStatus,
			Purpose:          "通知参与者饭局最终确认或取消结果",
			FieldMappings: map[string]string{
				"mealName": "thing1",
				"result":   "phrase2",
				"resultAt": "time3",
			},
		},
	)
	if err != nil || replayed || created.Enabled ||
		created.LogCount != 0 || len(created.FieldMappingSummary) != 3 {
		t.Fatalf("create subscription template result=%+v replayed=%v err=%v", created, replayed, err)
	}
	user := orderfoodModel.MiniAppUser{
		ID:              "subscribe-log-user",
		OpenIDHash:      strings.Repeat("s", 64),
		OpenIDEncrypted: "encrypted-subscribe-openid",
		Nickname:        "订阅记录用户",
		Status:          orderfoodModel.UserStatusNormal,
		Version:         1,
		RegisteredAt:    now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create subscribe log user: %v", err)
	}
	errorCode := "43101"
	errorSummary := "用户未保留本次授权"
	targetPage := "pages/meal/detail?id=meal-001"
	logRow := orderfoodModel.SubscribeMessageLog{
		ID: "subscribe-log-001", UserID: user.ID, TemplateID: created.ID,
		SubscriptionID: "meal-subscription-001",
		Scene:          orderfoodModel.SubscribeSceneMealStatus,
		Status:         orderfoodModel.SubscribeLogFailed,
		RetryCount:     1, WechatErrorCode: &errorCode,
		RequestID:            "request-subscribe-log-001",
		AuthorizationChecked: true, AuthorizationAvailable: false,
		TargetPage: &targetPage,
		SafePayloadSummary: datatypes.JSON(
			`{"mealId":"meal-001","mealName":"周末饭局","result":"已取消"}`,
		),
		ErrorSummary: &errorSummary,
		CreatedAt:    now.Add(time.Minute),
	}
	if err := db.Create(&logRow).Error; err != nil {
		t.Fatalf("create subscribe log: %v", err)
	}
	attempt := orderfoodModel.SubscribeMessageAttempt{
		LogID: logRow.ID, Attempt: 1, Status: orderfoodModel.SubscribeLogFailed,
		WechatErrorCode: &errorCode, ErrorSummary: &errorSummary,
		StartedAt: now.Add(time.Minute), FinishedAt: now.Add(2 * time.Minute),
	}
	if err := db.Create(&attempt).Error; err != nil {
		t.Fatalf("create subscribe attempt: %v", err)
	}
	templateDetail, err := service.GetSubscribeTemplate(ctx, created.ID)
	if err != nil || templateDetail.LogCount != 1 ||
		templateDetail.Purpose != "通知参与者饭局最终确认或取消结果" {
		t.Fatalf("template detail after log=%+v err=%v", templateDetail, err)
	}
	logs, err := service.ListSubscribeLogs(
		ctx,
		orderfoodRequest.SubscribeLogListQuery{
			TemplateID: created.ID,
			Scene:      orderfoodModel.SubscribeSceneMealStatus,
			Status:     orderfoodModel.SubscribeLogFailed,
		},
	)
	if err != nil || logs.Total != 1 || len(logs.List) != 1 ||
		logs.List[0].RequestID != logRow.RequestID {
		t.Fatalf("subscribe log list=%+v err=%v", logs, err)
	}
	logDetail, err := service.GetSubscribeLog(ctx, logRow.ID)
	if err != nil || len(logDetail.Attempts) != 1 ||
		logDetail.RelatedMealID == nil || *logDetail.RelatedMealID != "meal-001" ||
		logDetail.ErrorSummary == nil {
		t.Fatalf("subscribe log detail=%+v err=%v", logDetail, err)
	}
	_, _, err = service.DeleteSubscribeTemplate(
		ctx,
		subscriptionTestActor(),
		"subscription-template-delete-used",
		created.ID,
		orderfoodRequest.SubscribeTemplateDeleteInput{
			Reason:          "验证已有发送记录禁止删除",
			ExpectedVersion: created.Version,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminResourceInUse {
		t.Fatalf("delete referenced subscription template error = %v", err)
	}
}

// TestSubscriptionAdminRejectsDirectInvalidTemplateAndLogInputs 验证模板与发送记录服务拒绝绕过HTTP绑定的非法参数。
func TestSubscriptionAdminRejectsDirectInvalidTemplateAndLogInputs(t *testing.T) {
	service := &SubscriptionService{}
	_, _, err := service.CreateSubscribeTemplate(
		context.Background(),
		subscriptionTestActor(),
		"invalid-subscription-template",
		orderfoodRequest.SubscribeTemplateCreateInput{
			Name:             "",
			WechatTemplateID: strings.Repeat("w", 101),
			Scene:            orderfoodModel.SubscribeSceneMealStatus,
			Purpose:          "非法输入测试",
			FieldMappings: map[string]string{
				"mealName": "thing1",
				"result":   "thing1",
				"resultAt": "time3",
			},
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid direct subscription template error = %v", err)
	}
	_, _, err = service.UpdateSubscribeTemplateStatus(
		context.Background(),
		subscriptionTestActor(),
		"invalid-subscription-status",
		"template-id",
		orderfoodRequest.SubscribeTemplateStatusInput{
			Reason:          "非法启停输入",
			ExpectedVersion: 1,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("nil subscription template status error = %v", err)
	}
	_, err = service.ListSubscribeTemplates(
		context.Background(),
		orderfoodRequest.SubscribeTemplateListQuery{
			SortOrder: "desc; DROP TABLE of_sub_templates",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe subscription template sort error = %v", err)
	}
	_, err = service.ListSubscribeLogs(
		context.Background(),
		orderfoodRequest.SubscribeLogListQuery{
			Status:    "resent",
			SortOrder: "desc; DROP TABLE of_sub_logs",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe subscription log query error = %v", err)
	}
	_, err = service.GetSubscribeLog(context.Background(), "")
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid subscription log identity error = %v", err)
	}
}

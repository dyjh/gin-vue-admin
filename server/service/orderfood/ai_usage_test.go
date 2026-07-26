package orderfood

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	"gorm.io/datatypes"
)

// TestAIUsageListReturnsActualProviderAndModelIDs 验证调用记录可按实际供应商和模型筛选并返回精确跳转ID。
func TestAIUsageListReturnsActualProviderAndModelIDs(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.FrontFeatureUsage{},
	); err != nil {
		t.Fatalf("migrate AI usage tables: %v", err)
	}

	now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)
	user := orderfoodModel.MiniAppUser{
		ID:              "ai-usage-user",
		OpenIDHash:      "ai-usage-user-hash",
		OpenIDEncrypted: "ai-usage-user-encrypted",
		Nickname:        "调用用户",
		Status:          orderfoodModel.UserStatusNormal,
		Version:         1,
		RegisteredAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create AI usage user: %v", err)
	}
	usage := orderfoodModel.FrontFeatureUsage{
		ID:              "ai-usage-001",
		UserID:          user.ID,
		Feature:         "meal_suggest",
		CapabilityCode:  "meal_suggest",
		ProviderID:      "provider-deepseek",
		ProviderName:    "DeepSeek",
		ModelID:         "model-deepseek-chat",
		ModelName:       "DeepSeek Chat",
		RequestID:       "request-ai-usage-001",
		IdempotencyKey:  "idempotency-ai-usage-001",
		PointCost:       2,
		ExecutionStatus: orderfoodModel.FeatureExecutionSucceeded,
		BillingStatus:   orderfoodModel.FeatureBillingCharged,
		Version:         1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&usage).Error; err != nil {
		t.Fatalf("create AI usage record: %v", err)
	}

	service := NewAIUsageService(db, nil, nil)
	result, err := service.List(
		context.Background(),
		orderfoodRequest.AIUsageListQuery{
			ProviderID: usage.ProviderID,
			ModelID:    usage.ModelID,
		},
		orderfoodRequest.AdminActor{
			AdministratorID: 1,
			AuthorityID:     orderfoodModel.AuthorityOrderFoodSuperAdmin,
			RequestID:       "request-list-ai-usage",
		},
	)
	if err != nil {
		t.Fatalf("list AI usage records: %v", err)
	}
	if result.Total != 1 || len(result.List) != 1 {
		t.Fatalf("unexpected AI usage page: %+v", result)
	}
	if result.List[0].ProviderID != usage.ProviderID ||
		result.List[0].ModelID != usage.ModelID {
		t.Fatalf("actual provider/model IDs are missing: %+v", result.List[0])
	}
}

// TestAIUsageDetailSeparatesSensitiveContentAndBuildsTimeline 验证普通详情不泄露敏感字段，授权读取会写审计并返回完整调用时间线。
func TestAIUsageDetailSeparatesSensitiveContentAndBuildsTimeline(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.FrontFeatureUsage{},
		&orderfoodModel.FrontPointEntry{},
		&orderfoodModel.AdminAccessAudit{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.AdminIdempotencyRecord{},
	); err != nil {
		t.Fatalf("migrate AI usage detail tables: %v", err)
	}

	createdAt := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)
	finishedAt := createdAt.Add(3 * time.Second)
	updatedAt := createdAt.Add(4 * time.Second)
	user := orderfoodModel.MiniAppUser{
		ID: "ai-usage-detail-user", OpenIDHash: "ai-usage-detail-hash",
		OpenIDEncrypted: "ai-usage-detail-encrypted", Nickname: "详情用户",
		Status: orderfoodModel.UserStatusNormal, Version: 1, RegisteredAt: createdAt,
		CreatedAt: createdAt, UpdatedAt: createdAt,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create AI usage detail user: %v", err)
	}
	usage := orderfoodModel.FrontFeatureUsage{
		ID: "ai-usage-detail-001", UserID: user.ID, Feature: "meal_suggest",
		CapabilityCode: "meal_suggest", ProviderID: "provider-deepseek",
		ProviderName: "DeepSeek", ModelID: "model-deepseek-chat",
		ModelName: "DeepSeek Chat", RequestID: "request-ai-usage-detail",
		IdempotencyKey: "idempotency-ai-usage-detail", PromptMode: "custom",
		PromptHash: strings.Repeat("a", 64), PointCost: 2,
		ExecutionStatus: orderfoodModel.FeatureExecutionFailed,
		BillingStatus:   orderfoodModel.FeatureBillingRefunded,
		OriginalInputJSON: datatypes.JSON(
			[]byte(`{"dishName":"测试菜"}`),
		),
		ImageMetadataJSON: datatypes.JSON(
			[]byte(`[{"fileId":"file-001"}]`),
		),
		ModelOutputJSON: datatypes.JSON([]byte(`{"reason":"测试输出"}`)),
		Version:         1,
		FinishedAt:      &finishedAt,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	if err := db.Create(&usage).Error; err != nil {
		t.Fatalf("create AI usage detail record: %v", err)
	}
	objectType := "feature_usage"
	for _, entry := range []orderfoodModel.FrontPointEntry{
		{
			ID: "ai-usage-charge-entry", UserID: user.ID, Type: orderfoodModel.PointSpent,
			Scene: "feature_usage", Title: "扣除积分", Description: "AI调用",
			Amount: -2, RelatedObjectType: &objectType, RelatedObjectID: &usage.ID,
			CreatedAt: createdAt.Add(time.Second),
		},
		{
			ID: "ai-usage-refund-entry", UserID: user.ID, Type: orderfoodModel.PointRefund,
			Scene: "feature_refund", Title: "退还积分", Description: "AI调用失败",
			Amount: 2, RelatedObjectType: &objectType, RelatedObjectID: &usage.ID,
			CreatedAt: updatedAt,
		},
	} {
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("create AI usage point entry %s: %v", entry.ID, err)
		}
	}

	service := NewAIUsageService(db, nil, nil)
	actor := orderfoodRequest.AdminActor{
		AdministratorID: 1, Username: "admin", AuthorityID: orderfoodModel.AuthorityOrderFoodSuperAdmin,
		RequestID: "request-read-ai-usage-detail",
	}
	ordinary, err := service.Detail(context.Background(), usage.ID, false, actor)
	if err != nil {
		t.Fatalf("load ordinary AI usage detail: %v", err)
	}
	encoded, err := json.Marshal(ordinary)
	if err != nil {
		t.Fatalf("marshal ordinary AI usage detail: %v", err)
	}
	for _, forbidden := range []string{"originalInput", "imageMetadata", "modelOutput"} {
		if strings.Contains(string(encoded), `"`+forbidden+`"`) {
			t.Fatalf("ordinary detail leaked sensitive field %q: %s", forbidden, encoded)
		}
	}
	if ordinary.ChargedPointEntryID == nil || *ordinary.ChargedPointEntryID != "ai-usage-charge-entry" ||
		ordinary.RefundPointEntryID == nil || *ordinary.RefundPointEntryID != "ai-usage-refund-entry" ||
		len(ordinary.Timeline) != 4 {
		t.Fatalf("AI usage billing links or timeline mismatch: %+v", ordinary)
	}

	sensitive, err := service.Detail(context.Background(), usage.ID, true, actor)
	if err != nil {
		t.Fatalf("load sensitive AI usage detail: %v", err)
	}
	if sensitive.OriginalInput["dishName"] != "测试菜" ||
		len(sensitive.ImageMetadata) != 1 ||
		sensitive.ModelOutput == nil {
		t.Fatalf("sensitive AI usage content mismatch: %+v", sensitive)
	}
	var accessAuditCount int64
	if err := db.Model(&orderfoodModel.AdminAccessAudit{}).
		Where("action = ? AND target_id = ?", "read_ai_usage_sensitive_content", usage.ID).
		Count(&accessAuditCount).Error; err != nil {
		t.Fatalf("count AI usage sensitive access audits: %v", err)
	}
	if accessAuditCount != 1 {
		t.Fatalf("sensitive access audit count = %d, want 1", accessAuditCount)
	}
}

// TestAIUsageClearSensitiveContentIsPermanentAndIdempotent 验证L3清除会永久置空敏感内容、保留诊断字段并支持成功重放。
func TestAIUsageClearSensitiveContentIsPermanentAndIdempotent(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.FrontFeatureUsage{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.AdminIdempotencyRecord{},
	); err != nil {
		t.Fatalf("migrate AI usage clear tables: %v", err)
	}
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	user := orderfoodModel.MiniAppUser{
		ID: "ai-usage-clear-user", OpenIDHash: "ai-usage-clear-hash",
		OpenIDEncrypted: "ai-usage-clear-encrypted", Nickname: "清除用户",
		Status: orderfoodModel.UserStatusNormal, Version: 1, RegisteredAt: now,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create AI usage clear user: %v", err)
	}
	usage := orderfoodModel.FrontFeatureUsage{
		ID: "ai-usage-clear-001", UserID: user.ID, Feature: "dish_text_extract",
		CapabilityCode: "dish_text_extract", RequestID: "request-ai-usage-clear",
		IdempotencyKey: "idempotency-ai-usage-clear", PointCost: 1,
		ExecutionStatus: orderfoodModel.FeatureExecutionSucceeded,
		BillingStatus:   orderfoodModel.FeatureBillingCharged,
		OriginalInputJSON: datatypes.JSON(
			[]byte(`{"text":"需要清除"}`),
		),
		ImageMetadataJSON: datatypes.JSON([]byte(`[]`)),
		ModelOutputJSON:   datatypes.JSON([]byte(`{"name":"需要清除"}`)),
		Version:           1, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&usage).Error; err != nil {
		t.Fatalf("create AI usage clear record: %v", err)
	}
	service := NewAIUsageService(db, nil, nil)
	service.Now = func() time.Time { return now.Add(time.Minute) }
	actor := orderfoodRequest.AdminActor{
		AdministratorID: 1, Username: "admin", AuthorityID: orderfoodModel.AuthorityOrderFoodSuperAdmin,
		RequestID: "request-clear-ai-usage",
	}
	input := orderfoodRequest.AIUsageSensitiveDeleteInput{
		Reason: "故障排查完成，清除敏感内容", ExpectedVersion: 1,
	}
	result, replayed, err := service.ClearSensitiveContent(
		context.Background(), usage.ID, input, actor, "clear-ai-usage-key",
	)
	if err != nil || replayed || !result.Cleared {
		t.Fatalf("clear AI usage sensitive content: result=%+v replayed=%v err=%v", result, replayed, err)
	}
	replayedResult, replayed, err := service.ClearSensitiveContent(
		context.Background(), usage.ID, input, actor, "clear-ai-usage-key",
	)
	if err != nil || !replayed || !replayedResult.Cleared {
		t.Fatalf("replay AI usage sensitive clear: result=%+v replayed=%v err=%v", replayedResult, replayed, err)
	}
	var stored orderfoodModel.FrontFeatureUsage
	if err := db.First(&stored, "id = ?", usage.ID).Error; err != nil {
		t.Fatalf("reload cleared AI usage: %v", err)
	}
	if !stored.SensitiveContentCleared || stored.Version != 2 ||
		len(stored.OriginalInputJSON) != 0 || len(stored.ImageMetadataJSON) != 0 ||
		len(stored.ModelOutputJSON) != 0 || stored.RequestID != usage.RequestID ||
		stored.BillingStatus != usage.BillingStatus || stored.PointCost != usage.PointCost {
		t.Fatalf("cleared AI usage retained data mismatch: %+v", stored)
	}
	var auditCount int64
	if err := db.Model(&orderfoodModel.AdminAuditLog{}).
		Where("action = ? AND target_id = ?", "clear_ai_usage_sensitive_content", usage.ID).
		Count(&auditCount).Error; err != nil {
		t.Fatalf("count AI usage clear audits: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("clear audit count = %d, want 1", auditCount)
	}
}

// TestAIUsageListRejectsUnsafeDirectSort 验证服务层会拒绝绕过HTTP绑定传入的非法排序值。
func TestAIUsageListRejectsUnsafeDirectSort(t *testing.T) {
	db := newOrderFoodTestDB(t)
	service := NewAIUsageService(db, nil, nil)
	_, err := service.List(
		context.Background(),
		orderfoodRequest.AIUsageListQuery{SortOrder: "desc; DROP TABLE of_ai_usages"},
		orderfoodRequest.AdminActor{AuthorityID: orderfoodModel.AuthorityOrderFoodSuperAdmin},
	)
	if err == nil {
		t.Fatal("unsafe direct AI usage sort was accepted")
	}
}

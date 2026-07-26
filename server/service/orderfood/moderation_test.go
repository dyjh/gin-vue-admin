package orderfood

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

var moderationTinyPNG = func() []byte {
	decoded, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZlV8AAAAASUVORK5CYII=",
	)
	if err != nil {
		panic(err)
	}
	return decoded
}()

// moderationFakeProvider 为图片审核配置测试提供可控的供应商响应。
type moderationFakeProvider struct {
	connectionResult ModerationProviderResult // 连接测试结果
	imageResult      ModerationProviderResult // 图片审核结果
	connectionCalls  int                      // 连接测试调用次数
	imageCalls       int                      // 图片审核调用次数
}

// TestConnection 返回预设的连接测试结果。
func (provider *moderationFakeProvider) TestConnection(
	_ context.Context,
	_ ModerationProviderConfig,
) ModerationProviderResult {
	provider.connectionCalls++
	return provider.connectionResult
}

// ModerateImage 返回预设的图片审核结果。
func (provider *moderationFakeProvider) ModerateImage(
	_ context.Context,
	_ ModerationProviderConfig,
	_ ModerationProviderImage,
) ModerationProviderResult {
	provider.imageCalls++
	return provider.imageResult
}

// moderationAcceptedBusinessRow 表示图片审核通过后写入的测试业务记录。
type moderationAcceptedBusinessRow struct {
	ID                 uint   `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"` // 主键ID
	StagingFileID      string `gorm:"column:staging_file_id;type:varchar(128);not null;comment:暂存文件ID;"`              // 暂存文件ID
	ModerationRecordID string `gorm:"column:moderation_record_id;type:varchar(64);not null;comment:审核记录ID;"`          // 审核记录ID
}

// TableName 指定moderationAcceptedBusinessRow对应的数据表名。
func (moderationAcceptedBusinessRow) TableName() string {
	return "test_mod_accepted_rows"
}

// openModerationTestDB 创建图片审核服务测试所需的独立 MySQL 数据库。
func openModerationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	models := []interface{}{
		&system.SysBaseMenuBtn{},
		&system.SysAuthorityBtn{},
		&orderfoodModel.AdminAccessAudit{},
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.ModerationConfig{},
		&orderfoodModel.ModerationConnectionTest{},
		&orderfoodModel.ModerationImageTest{},
		&orderfoodModel.ModerationConfigHealth{},
		&orderfoodModel.ImageModerationRecord{},
		&moderationAcceptedBusinessRow{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate moderation schema: %v", err)
	}
	seedTestPermissionTemplates(t, db)
	return db
}

// moderationTestActor 返回具备图片审核配置权限的测试管理员。
func moderationTestActor() orderfoodRequest.AdminActor {
	return orderfoodRequest.AdminActor{
		AdministratorID:  7,
		AuthorityID:      orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:         "root",
		RequestID:        "req-moderation-test",
		SourceIPMasked:   "127.0.0.0",
		UserAgentSummary: "moderation-test",
	}
}

// newModerationTestService 创建使用固定时间和供应商的图片审核服务。
func newModerationTestService(
	db *gorm.DB,
	provider ModerationProvider,
	now time.Time,
) *ModerationService {
	idempotency := NewIdempotencyService(db)
	idempotency.Now = func() time.Time { return now }
	service := NewModerationService(
		db,
		provider,
		NewPermissionService(db),
		NewAccessAuditService(db),
		idempotency,
	)
	service.Now = func() time.Time { return now }
	return service
}

// TestModerationConfigSaveAppliesImmediatelyAndTestsCurrentConfig 验证图片审核配置直接生效且测试只使用当前配置。
func TestModerationConfigSaveAppliesImmediatelyAndTestsCurrentConfig(t *testing.T) {
	db := openModerationTestDB(t)
	now := time.Date(2026, 7, 24, 8, 0, 0, 0, time.UTC)
	provider := &moderationFakeProvider{
		connectionResult: ModerationProviderResult{
			Status:      orderfoodModel.ModerationStatusPassed,
			Category:    ModerationCategoryOK,
			RiskLabels:  []string{},
			DurationMS:  18,
			SafeMessage: "连接测试通过",
		},
		imageResult: ModerationProviderResult{
			Status:      orderfoodModel.ModerationStatusPassed,
			Category:    ModerationCategoryPassed,
			RiskLabels:  []string{},
			DurationMS:  23,
			SafeMessage: "图片审核通过",
			ProviderResponseSummary: map[string]interface{}{
				"requestCode":     "200",
				"AccessKeySecret": "must-not-leak",
			},
		},
	}
	service := newModerationTestService(db, provider, now)
	ctx := context.Background()
	actor := moderationTestActor()
	current, err := service.GetConfig(ctx, actor)
	if err != nil {
		t.Fatalf("get default config: %v", err)
	}
	if current.ConfigVersion != 1 || current.Enabled {
		t.Fatalf("unexpected default config: %+v", current)
	}

	credentialRef := "env://ORDERFOOD_MODERATION_TEST_CREDENTIAL"
	updated, replayed, err := service.UpdateConfig(
		ctx,
		orderfoodRequest.ModerationConfigUpdateInput{
			Enabled:         true,
			Region:          "cn-shanghai",
			Endpoint:        "https://green-cip.cn-shanghai.aliyuncs.com",
			ServiceCode:     "baselineCheck",
			TimeoutMS:       3000,
			RetryCount:      1,
			RetryBackoffMS:  200,
			CredentialRef:   &credentialRef,
			Reason:          "启用用户图片审核",
			ExpectedVersion: current.ConfigVersion,
		},
		actor,
		"moderation-config-update",
	)
	if err != nil || replayed || updated.ConfigVersion != 2 || !updated.Enabled {
		t.Fatalf("update config: result=%+v replayed=%v err=%v", updated, replayed, err)
	}
	encoded, _ := json.Marshal(updated)
	if strings.Contains(string(encoded), credentialRef) ||
		strings.Contains(string(encoded), "credentialRef") {
		t.Fatalf("config response leaked credential: %s", encoded)
	}
	reloaded, err := service.GetConfig(ctx, actor)
	if err != nil || reloaded.ConfigVersion != updated.ConfigVersion || !reloaded.Enabled {
		t.Fatalf("config did not apply immediately: result=%+v err=%v", reloaded, err)
	}
	var configCount int64
	if err := db.Model(&orderfoodModel.ModerationConfig{}).Count(&configCount).Error; err != nil {
		t.Fatal(err)
	}
	if configCount != 1 {
		t.Fatalf("moderation config count = %d, want 1", configCount)
	}

	expectedVersion := updated.ConfigVersion
	connection, _, err := service.TestConnection(
		ctx,
		orderfoodRequest.ModerationConfigTestInput{ExpectedVersion: &expectedVersion},
		actor,
		"moderation-connection-test",
	)
	if err != nil || !connection.Success || connection.ConfigVersion != updated.ConfigVersion {
		t.Fatalf("test current connection: result=%+v err=%v", connection, err)
	}
	imageResult, _, err := service.TestImage(
		ctx,
		orderfoodRequest.ModerationConfigTestInput{ExpectedVersion: &expectedVersion},
		ModerationProviderImage{
			Bytes:       moderationTinyPNG,
			ContentType: "image/png",
		},
		actor,
		"moderation-image-test",
	)
	if err != nil || imageResult.MappedStatus != string(orderfoodModel.ModerationStatusPassed) ||
		imageResult.ConfigVersion != updated.ConfigVersion {
		t.Fatalf("test current image config: result=%+v err=%v", imageResult, err)
	}
	if _, exists := imageResult.ProviderResponseSummary["AccessKeySecret"]; exists {
		t.Fatal("provider response summary leaked credential-like field")
	}
}

// TestImageModerationGatePersistsOnlyExplicitPass 验证用户图片只有明确审核通过时才写入业务数据。
func TestImageModerationGatePersistsOnlyExplicitPass(t *testing.T) {
	testCases := []struct {
		name          string                   // 用例名称
		result        ModerationProviderResult // 供应商返回
		wantAllowed   bool                     // 是否允许写入
		wantErrorCode appErrors.ErrorType      // 预期错误码
	}{
		{
			name: "explicit pass",
			result: ModerationProviderResult{
				Status:      orderfoodModel.ModerationStatusPassed,
				Category:    ModerationCategoryPassed,
				RiskLabels:  []string{},
				SafeMessage: "图片审核通过",
			},
			wantAllowed: true,
		},
		{
			name: "content risk",
			result: ModerationProviderResult{
				Status:      orderfoodModel.ModerationStatusRejected,
				Category:    ModerationCategoryContentRisk,
				RiskLabels:  []string{"block"},
				SafeMessage: "图片内容审核未通过",
			},
			wantErrorCode: appErrors.FrontImageRejected,
		},
		{
			name: "provider timeout",
			result: moderationFailedProviderResult(
				ModerationCategoryTimeout,
				"图片审核服务调用超时",
				1000,
			),
			wantErrorCode: appErrors.FrontInvalidImage,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			db := openModerationTestDB(t)
			now := time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC)
			provider := &moderationFakeProvider{imageResult: testCase.result}
			service := newModerationTestService(db, provider, now)
			credentialRef := "env://ORDERFOOD_MODERATION_TEST_CREDENTIAL"
			if _, _, err := service.UpdateConfig(
				context.Background(),
				orderfoodRequest.ModerationConfigUpdateInput{
					Enabled:         true,
					Region:          "cn-shanghai",
					Endpoint:        "https://green-cip.cn-shanghai.aliyuncs.com",
					ServiceCode:     "baselineCheck",
					TimeoutMS:       1000,
					RetryCount:      0,
					RetryBackoffMS:  0,
					CredentialRef:   &credentialRef,
					Reason:          "配置图片审核测试",
					ExpectedVersion: 1,
				},
				moderationTestActor(),
				"gate-config-"+strings.ReplaceAll(testCase.name, " ", "-"),
			); err != nil {
				t.Fatalf("configure moderation: %v", err)
			}

			writerCalls := 0
			persisted, decision, err := service.AcceptUserImage(
				context.Background(),
				ModerationGateInput{
					RequestID:     "req-gate",
					StagingFileID: "staging-1",
					Scene:         orderfoodModel.ModerationSceneDishCover,
					ContentType:   "image/png",
					Bytes:         moderationTinyPNG,
				},
				func(tx *gorm.DB, permit PassedImagePermit) (interface{}, error) {
					writerCalls++
					row := moderationAcceptedBusinessRow{
						StagingFileID:      "staging-1",
						ModerationRecordID: permit.ModerationRecordID,
					}
					if err := tx.Create(&row).Error; err != nil {
						return nil, err
					}
					return row.ID, nil
				},
			)
			if decision.Allowed != testCase.wantAllowed {
				t.Fatalf("unexpected decision: %+v", decision)
			}
			if testCase.wantAllowed {
				if err != nil || persisted == nil || writerCalls != 1 {
					t.Fatalf("expected persistence: result=%v calls=%d err=%v", persisted, writerCalls, err)
				}
			} else {
				if appErrors.Code(err) != int(testCase.wantErrorCode) ||
					persisted != nil || writerCalls != 0 {
					t.Fatalf("fail-closed violation: result=%v calls=%d err=%v", persisted, writerCalls, err)
				}
			}
		})
	}
}

// TestGeneratedCoverBypassesUploadModeration 验证 AI 生成封面不调用图片审核供应商并直接标记为无需审核。
func TestGeneratedCoverBypassesUploadModeration(t *testing.T) {
	db := openModerationTestDB(t)
	now := time.Date(2026, 7, 26, 8, 30, 0, 0, time.UTC)
	provider := &moderationFakeProvider{}
	service := newModerationTestService(db, provider, now)
	writerCalls := 0

	persisted, decision, err := service.AcceptUserImage(
		context.Background(),
		ModerationGateInput{
			RequestID:     "request-generated-cover",
			StagingFileID: "generated-cover-1",
			Scene:         orderfoodModel.ModerationSceneGeneratedCover,
			ContentType:   "image/png",
			Bytes:         moderationTinyPNG,
		},
		func(tx *gorm.DB, permit PassedImagePermit) (interface{}, error) {
			writerCalls++
			row := moderationAcceptedBusinessRow{
				StagingFileID:      "generated-cover-1",
				ModerationRecordID: permit.ModerationRecordID,
			}
			if err := tx.Create(&row).Error; err != nil {
				return nil, err
			}
			return row.ID, nil
		},
	)
	if err != nil || persisted == nil || writerCalls != 1 {
		t.Fatalf("generated cover persistence: result=%v calls=%d err=%v", persisted, writerCalls, err)
	}
	if provider.imageCalls != 0 || !decision.Allowed ||
		decision.Status != string(orderfoodModel.ModerationStatusNotRequired) {
		t.Fatalf("generated cover decision=%+v providerCalls=%d", decision, provider.imageCalls)
	}
}

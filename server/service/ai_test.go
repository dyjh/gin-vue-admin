package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TestValidateProviderAPIKey 验证供应商 API Key 必填、长度受限且不能包含空白字符。
func TestValidateProviderAPIKey(t *testing.T) {
	if err := validateProviderAPIKey("sk-test_123-ABC.xyz"); err != nil {
		t.Fatalf("validate API key: %v", err)
	}
	for _, apiKey := range []string{
		"",
		"   ",
		"sk-test key",
		strings.Repeat("a", 501),
	} {
		if err := validateProviderAPIKey(apiKey); err == nil {
			t.Fatalf("expected API key to fail validation: %q", apiKey)
		}
	}
}

// TestHTTPAIConnectorCredentialUsesStoredAPIKey 验证连接器直接读取数据库模型中的 API Key。
func TestHTTPAIConnectorCredentialUsesStoredAPIKey(t *testing.T) {
	connector := NewHTTPAIConnector(nil)
	value, err := connector.credential(
		context.Background(),
		orderfoodModel.AIProvider{APIKey: " test-api-key "},
	)
	if err != nil {
		t.Fatalf("read stored AI API key: %v", err)
	}
	if value != "test-api-key" {
		t.Fatalf("unexpected AI API key: %q", value)
	}
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := connector.credential(
		cancelledContext,
		orderfoodModel.AIProvider{APIKey: "test-api-key"},
	); err == nil {
		t.Fatal("expected cancelled context to fail")
	}
}

// TestHTTPAIConnectorListModels 验证实时模型列表会鉴权、过滤、去重并排序。
func TestHTTPAIConnectorListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/models" || request.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(writer, "invalid request", http.StatusUnauthorized)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"id": "z-model"},
				{"id": "a-model"},
				{"id": "z-model"},
				{"id": "bad model"},
				{"id": ""},
				{"id": strings.Repeat("m", 121)},
			},
		})
	}))
	defer server.Close()
	connector := NewHTTPAIConnector(server.Client())
	modelKeys, err := connector.ListModels(
		context.Background(),
		orderfoodModel.AIProvider{
			BaseURL: server.URL,
			APIKey:  "test-api-key",
		},
		time.Second,
	)
	if err != nil {
		t.Fatalf("list provider models: %v", err)
	}
	if len(modelKeys) != 2 || modelKeys[0] != "a-model" || modelKeys[1] != "z-model" {
		t.Fatalf("unexpected normalized model keys: %#v", modelKeys)
	}
}

// fakeAIConnector 为配置测试提供稳定的供应商连接和提示词执行结果。
type fakeAIConnector struct{}

// ListModels 返回稳定的供应商模型选项。
func (fakeAIConnector) ListModels(
	context.Context,
	orderfoodModel.AIProvider,
	time.Duration,
) ([]string, error) {
	return []string{"test-model"}, nil
}

// TestConnection 返回成功的供应商连接测试结果。
func (fakeAIConnector) TestConnection(
	context.Context,
	orderfoodModel.AIProvider,
	time.Duration,
) (AIConnectionResult, error) {
	return AIConnectionResult{
		Success:     true,
		Category:    "ok",
		DurationMS:  12,
		SafeMessage: "连接成功",
	}, nil
}

// TestPrompt 返回成功的提示词测试结果。
func (fakeAIConnector) TestPrompt(
	context.Context,
	orderfoodModel.AIProvider,
	orderfoodModel.AIModel,
	AIPromptRunRequest,
	time.Duration,
) (AIPromptRunResult, error) {
	inputTokens := 20
	outputTokens := 8
	return AIPromptRunResult{
		Success:      true,
		DurationMS:   25,
		InputTokens:  &inputTokens,
		OutputTokens: &outputTokens,
		SafeMessage:  "测试成功",
	}, nil
}

// failingAIMutationAuditWriter 模拟审计写入失败。
type failingAIMutationAuditWriter struct{}

// WriteMutationAudit 固定返回内部错误，用于验证事务回滚。
func (failingAIMutationAuditWriter) WriteMutationAudit(
	context.Context,
	*gorm.DB,
	*orderfoodModel.AdminAuditLog,
) error {
	return appErrors.AdminInternal.New("forced audit failure")
}

// openAITestDB 创建 AI 服务测试所需的独立 MySQL 数据库。
func openAITestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	models := append(
		AIModelsForMigration(),
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAccessAudit{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.MiniAppUser{},
		&system.SysBaseMenuBtn{},
		&system.SysAuthorityBtn{},
	)
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate AI models: %v", err)
	}
	seedTestPermissionTemplates(t, db)
	return db
}

// aiTestActor 返回具备全部 AI 配置权限的测试管理员。
func aiTestActor() orderfoodRequest.AIAdminActor {
	nickname := "AI 管理员"
	return orderfoodRequest.AIAdminActor{
		AdministratorID:  7,
		AuthorityID:      orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:         "ai-admin",
		Nickname:         &nickname,
		RequestID:        "req-ai-test",
		SourceIPMasked:   "127.0.0.0",
		UserAgentSummary: "go-test",
	}
}

// boolPointer 返回布尔值指针。
func boolPointer(value bool) *bool {
	return &value
}

// mustJSON 将测试值编码为JSON，编码失败时立即中止当前测试进程。
func mustJSON(value interface{}) datatypes.JSON {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(encoded)
}

// seedAllAICapabilitiesReady 将六项能力初始化为可用状态。
func seedAllAICapabilitiesReady(t *testing.T, db *gorm.DB, now time.Time) {
	t.Helper()
	provider := orderfoodModel.AIProvider{
		ID: "platform-ready-provider", Name: "平台开关测试供应商",
		Type: orderfoodModel.AIProviderOpenAI, BaseURL: "https://example.test/v1",
		APIKey: "test-api-key", TimeoutMS: 30000,
		Enabled: true, Version: 1, UpdatedByID: 7,
		UpdatedByUsername: "ai-admin", CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&provider).Error; err != nil {
		t.Fatalf("seed ready provider: %v", err)
	}
	model := orderfoodModel.AIModel{
		ID: "platform-ready-model", ProviderID: provider.ID,
		Name: "平台开关测试模型", ModelKey: "ready-model",
		CapabilitiesJSON: datatypes.JSON(`["text","vision","image_generation"]`),
		Enabled:          true, Version: 1, UpdatedByID: 7,
		UpdatedByUsername: "ai-admin", CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&model).Error; err != nil {
		t.Fatalf("seed ready model: %v", err)
	}
	for _, item := range defaultCapabilities() {
		config := orderfoodModel.AICapabilityConfig{
			CapabilityCode: item.Code, Version: 1,
			PrimaryModelID: model.ID, TimeoutMS: 30000,
			PointCost:                   item.SortOrder,
			FreeQuotaPerDay:             item.SortOrder + 10,
			PromptMode:                  orderfoodModel.AIPromptPreset,
			PromptPresetVersion:         1,
			SystemPrompt:                item.SystemPrompt,
			UserPromptTemplate:          item.UserPromptTemplate,
			PromptAllowedVariablesJSON:  mustJSON(item.AllowedVariables),
			PromptRequiredVariablesJSON: mustJSON(item.RequiredVariables),
			PromptOutputSchemaVersion:   item.OutputSchemaVersion,
			PromptHash:                  strings.Repeat("f", 64),
			AppliedByID:                 7,
			AppliedByUsername:           "ai-admin",
			AppliedAt:                   now,
			Reason:                      "平台整体开关测试",
		}
		if err := db.Create(&config).Error; err != nil {
			t.Fatalf("seed ready capability %s: %v", item.Code, err)
		}
	}
}

// TestAIPlatformPolicySaveAppliesImmediatelyAndEmergencyRemainsIndependent 验证平台策略直接生效且紧急开关独立。
func TestAIPlatformPolicySaveAppliesImmediatelyAndEmergencyRemainsIndependent(t *testing.T) {
	db := openAITestDB(t)
	service := NewAIService(
		db,
		NewPermissionService(db),
		NewIdempotencyService(db),
		fakeAIConnector{},
	)
	now := time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC)
	service.Now = func() time.Time { return now }
	service.Idempotency.Now = service.Now
	ctx := context.Background()
	if err := service.EnsureDefaults(ctx); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	seedAllAICapabilitiesReady(t, db, now)
	users := []orderfoodModel.MiniAppUser{
		{ID: "normal-a", OpenIDHash: strings.Repeat("a", 64), OpenIDEncrypted: "enc-a", Nickname: "A", Status: orderfoodModel.UserStatusNormal, Version: 1, RegisteredAt: now},
		{ID: "normal-b", OpenIDHash: strings.Repeat("b", 64), OpenIDEncrypted: "enc-b", Nickname: "B", Status: orderfoodModel.UserStatusNormal, Version: 1, RegisteredAt: now},
		{ID: "blocked", OpenIDHash: strings.Repeat("c", 64), OpenIDEncrypted: "enc-c", Nickname: "C", Status: orderfoodModel.UserStatusDisabled, Version: 1, RegisteredAt: now},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	workspace, err := service.PlatformPolicyWorkspace(ctx)
	if err != nil {
		t.Fatalf("get platform workspace: %v", err)
	}
	if workspace.Config.PlatformDefaultEnabled {
		t.Fatal("platform default must start disabled")
	}
	labels := make([]orderfoodRequest.ClientFeatureLabelInput, 0, len(workspace.Config.FeatureLabels))
	for _, label := range workspace.Config.FeatureLabels {
		labels = append(labels, orderfoodRequest.ClientFeatureLabelInput{
			Code:        label.Code,
			Title:       label.Title,
			ActionLabel: label.ActionLabel,
			Description: label.Description,
			SortOrder:   label.SortOrder,
		})
	}
	if workspace.Config.FeatureLabels[0].CostHint == nil ||
		*workspace.Config.FeatureLabels[0].CostHint !=
			"菜品文本解析：每次消耗 1 积分；菜谱长截图解析：每次消耗 2 积分" {
		t.Fatalf("platform billing hints were not derived from capability configs: %+v", workspace.Config.FeatureLabels[0])
	}
	updated, replayed, err := service.UpdatePlatformPolicy(
		ctx,
		aiTestActor(),
		"platform-policy-update",
		orderfoodRequest.PlatformPolicyUpdateInput{
			PlatformDefaultEnabled: true,
			FeatureLabels:          labels,
			Reason:                 "开放平台全部 AI 功能",
			ExpectedVersion:        workspace.Config.PolicyVersion,
		},
	)
	if err != nil || replayed {
		t.Fatalf("update platform policy: result=%+v replayed=%v err=%v", updated, replayed, err)
	}
	if !updated.PlatformDefaultEnabled || updated.PolicyVersion != workspace.Config.PolicyVersion+1 {
		t.Fatalf("platform policy did not apply immediately: %+v", updated)
	}
	var persisted orderfoodModel.PlatformCapabilityPolicy
	if err := db.First(&persisted, "singleton_key = ?", platformPolicySingletonKey).Error; err != nil {
		t.Fatalf("load persisted platform policy: %v", err)
	}
	if strings.Contains(string(persisted.FeatureLabelsJSON), "costHint") ||
		strings.Contains(string(persisted.FeatureLabelsJSON), "freeQuotaHint") {
		t.Fatalf("platform policy duplicated live billing hints: %s", persisted.FeatureLabelsJSON)
	}
	replayedConfig, replayed, err := service.UpdatePlatformPolicy(
		ctx,
		aiTestActor(),
		"platform-policy-update",
		orderfoodRequest.PlatformPolicyUpdateInput{
			PlatformDefaultEnabled: true,
			FeatureLabels:          labels,
			Reason:                 "开放平台全部 AI 功能",
			ExpectedVersion:        workspace.Config.PolicyVersion,
		},
	)
	if err != nil || !replayed || replayedConfig.PolicyVersion != updated.PolicyVersion {
		t.Fatalf("replay platform update: result=%+v replayed=%v err=%v", replayedConfig, replayed, err)
	}

	emergency, _, err := service.SetPlatformEmergencyStatus(
		ctx,
		aiTestActor(),
		"platform-emergency",
		orderfoodRequest.PlatformPolicyEmergencyInput{
			EmergencyDisabled: boolPointer(true),
			Reason:            "供应商异常立即止损",
			ExpectedVersion:   updated.PolicyVersion,
		},
	)
	if err != nil || !emergency.EmergencyDisabled {
		t.Fatalf("set emergency status: result=%+v err=%v", emergency, err)
	}
	snapshot, err := service.Snapshot(ctx)
	if err != nil || !snapshot.EmergencyDisabled || !snapshot.PlatformDefaultEnabled {
		t.Fatalf("unexpected runtime snapshot: %+v err=%v", snapshot, err)
	}
	workspace, err = service.PlatformPolicyWorkspace(ctx)
	if err != nil {
		t.Fatalf("reload platform workspace: %v", err)
	}
	if workspace.UserCounts.EffectiveEnabledUserCount != 0 {
		t.Fatalf("effective users under emergency = %d, want 0", workspace.UserCounts.EffectiveEnabledUserCount)
	}
	var policyCount int64
	if err := db.Model(&orderfoodModel.PlatformCapabilityPolicy{}).
		Count(&policyCount).Error; err != nil {
		t.Fatal(err)
	}
	if policyCount != 1 {
		t.Fatalf("platform policy count = %d, want 1", policyCount)
	}
}

// TestAIPlatformPolicyRejectsUnreadyEnableDuringEmergency 验证紧急停用期间也不能绕过六项能力就绪校验预先打开总开关。
func TestAIPlatformPolicyRejectsUnreadyEnableDuringEmergency(t *testing.T) {
	db := openAITestDB(t)
	service := NewAIService(
		db,
		NewPermissionService(db),
		NewIdempotencyService(db),
		fakeAIConnector{},
	)
	now := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	service.Now = func() time.Time { return now }
	service.Idempotency.Now = service.Now
	ctx := context.Background()
	if err := service.EnsureDefaults(ctx); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	initial, err := service.PlatformPolicyWorkspace(ctx)
	if err != nil {
		t.Fatalf("load initial platform policy: %v", err)
	}
	emergency, _, err := service.SetPlatformEmergencyStatus(
		ctx,
		aiTestActor(),
		"platform-unready-emergency",
		orderfoodRequest.PlatformPolicyEmergencyInput{
			EmergencyDisabled: boolPointer(true),
			Reason:            "供应商故障需要紧急停用",
			ExpectedVersion:   initial.Config.PolicyVersion,
		},
	)
	if err != nil {
		t.Fatalf("enable emergency status: %v", err)
	}
	labels := make([]orderfoodRequest.ClientFeatureLabelInput, 0, len(emergency.FeatureLabels))
	for _, label := range emergency.FeatureLabels {
		labels = append(labels, orderfoodRequest.ClientFeatureLabelInput{
			Code:        label.Code,
			Title:       label.Title,
			ActionLabel: label.ActionLabel,
			Description: label.Description,
			SortOrder:   label.SortOrder,
		})
	}
	_, _, err = service.UpdatePlatformPolicy(
		ctx,
		aiTestActor(),
		"platform-unready-enable",
		orderfoodRequest.PlatformPolicyUpdateInput{
			PlatformDefaultEnabled: true,
			FeatureLabels:          labels,
			Reason:                 "预先开启平台能力",
			ExpectedVersion:        emergency.PolicyVersion,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInvalidConfig {
		t.Fatalf("unready enable error = %v, want admin invalid config", err)
	}
	current, err := service.PlatformPolicyWorkspace(ctx)
	if err != nil {
		t.Fatalf("reload platform policy: %v", err)
	}
	if current.Config.PlatformDefaultEnabled || !current.Config.EmergencyDisabled {
		t.Fatalf("invalid enable changed current policy: %+v", current.Config)
	}
}

// TestAIProviderModelCapabilityAndPromptSaveApplyImmediately 验证能力与提示词直接保存后立即成为当前配置。
func TestAIProviderModelCapabilityAndPromptSaveApplyImmediately(t *testing.T) {
	db := openAITestDB(t)
	service := NewAIService(
		db,
		NewPermissionService(db),
		NewIdempotencyService(db),
		fakeAIConnector{},
	)
	ctx := context.Background()
	if err := service.EnsureDefaults(ctx); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	provider, _, err := service.CreateAIProvider(
		ctx,
		aiTestActor(),
		"provider-create",
		orderfoodRequest.AIProviderCreateInput{
			Name:      "DeepSeek",
			Type:      orderfoodModel.AIProviderDeepSeek,
			BaseURL:   "https://api.deepseek.example/v1",
			APIKey:    "sk-test-deepseek",
			TimeoutMS: 30000,
		},
	)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	encodedProvider, _ := json.Marshal(provider)
	if strings.Contains(string(encodedProvider), "sk-test-deepseek") ||
		strings.Contains(string(encodedProvider), "apiKey") {
		t.Fatalf("provider response leaked credential: %s", encodedProvider)
	}
	_, _, err = service.UpdateAIProviderStatus(
		ctx,
		aiTestActor(),
		provider.ID,
		"provider-enable-without-test",
		orderfoodRequest.AIStatusUpdateInput{
			Enabled:         boolPointer(true),
			Reason:          "尚未执行当前配置连接测试",
			ExpectedVersion: provider.Version,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInvalidConfig {
		t.Fatalf("enable untested provider error = %v", err)
	}
	connectionTest, _, err := service.TestAIProviderConnection(
		ctx,
		aiTestActor(),
		provider.ID,
		"provider-connection-test",
		orderfoodRequest.AIProviderConnectionTestInput{
			ExpectedVersion: provider.Version,
		},
	)
	if err != nil || !connectionTest.Success {
		t.Fatalf("test provider connection: result=%+v err=%v", connectionTest, err)
	}
	provider, _, err = service.UpdateAIProviderStatus(
		ctx,
		aiTestActor(),
		provider.ID,
		"provider-enable",
		orderfoodRequest.AIStatusUpdateInput{
			Enabled:         boolPointer(true),
			Reason:          "连接配置验证完成",
			ExpectedVersion: provider.Version,
		},
	)
	if err != nil {
		t.Fatalf("enable provider: %v", err)
	}
	model, _, err := service.CreateAIModel(
		ctx,
		aiTestActor(),
		"model-create",
		orderfoodRequest.AIModelCreateInput{
			ProviderID:   provider.ID,
			Name:         "DeepSeek Chat",
			ModelKey:     "deepseek-chat",
			Capabilities: []orderfoodModel.AIModelCapability{orderfoodModel.AIModelText},
		},
	)
	if err != nil {
		t.Fatalf("create model: %v", err)
	}
	model, _, err = service.UpdateAIModelStatus(
		ctx,
		aiTestActor(),
		model.ID,
		"model-enable",
		orderfoodRequest.AIStatusUpdateInput{
			Enabled:         boolPointer(true),
			Reason:          "模型配置验证完成",
			ExpectedVersion: model.Version,
		},
	)
	if err != nil {
		t.Fatalf("enable model: %v", err)
	}

	config, _, err := service.UpdateAICapability(
		ctx,
		aiTestActor(),
		orderfoodModel.AICapabilityMealSuggest,
		"capability-update",
		orderfoodRequest.AICapabilityUpdateInput{
			PrimaryModelID:    model.ID,
			PointCost:         1,
			DailyLimitPerUser: 10,
			TimeoutMS:         30000,
			FreeQuotaPerDay:   1,
			Reason:            "配置推荐主模型",
			ExpectedVersion:   0,
		},
	)
	if err != nil || config.Version != 1 {
		t.Fatalf("update capability: result=%+v err=%v", config, err)
	}
	if len(config.FixedValidationRules) < 2 {
		t.Fatalf("fixed validation rules were not returned: %+v", config)
	}
	providerDetail, err := service.AIProviderDetail(ctx, provider.ID)
	if err != nil {
		t.Fatalf("load provider references: %v", err)
	}
	if len(providerDetail.ReferencedModels) != 1 ||
		providerDetail.ReferencedModels[0].ID != model.ID ||
		len(providerDetail.ReferencedCapabilities) != 1 ||
		providerDetail.ReferencedCapabilities[0].CapabilityCode !=
			orderfoodModel.AICapabilityMealSuggest {
		t.Fatalf("provider reference details mismatch: %+v", providerDetail)
	}
	modelDetail, err := service.AIModelDetail(ctx, model.ID)
	if err != nil {
		t.Fatalf("load model references: %v", err)
	}
	if len(modelDetail.ReferencedCapabilities) != 1 ||
		modelDetail.ReferencedCapabilities[0].CapabilityCode !=
			orderfoodModel.AICapabilityMealSuggest ||
		modelDetail.UsageCount != 0 ||
		modelDetail.ReferenceCount != 1 {
		t.Fatalf("model reference details mismatch: %+v", modelDetail)
	}

	systemPrompt := "只允许从服务端给出的候选菜中选择，不能新增菜品。"
	userTemplate := "候选：{{candidate_dishes}}\n限制：{{constraints}}\n人数：{{servings}}\n数量：{{target_count}}\n索引：{{metadata}}"
	_, _, err = service.UpdateAICapability(
		ctx,
		aiTestActor(),
		orderfoodModel.AICapabilityMealSuggest,
		"capability-invalid-bounds",
		orderfoodRequest.AICapabilityUpdateInput{
			PrimaryModelID:    model.ID,
			PointCost:         -1,
			DailyLimitPerUser: 10,
			TimeoutMS:         30000,
			FreeQuotaPerDay:   1,
			Reason:            "验证服务层参数边界",
			ExpectedVersion:   config.Version,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid capability bounds error = %v, want admin bad request", err)
	}
	if err := db.Model(&orderfoodModel.AIModel{}).
		Where("id = ?", model.ID).
		Update("context_length", 1).Error; err != nil {
		t.Fatalf("set model context limit: %v", err)
	}
	validation, _, err := service.ValidateAICapabilityPrompt(
		ctx,
		aiTestActor(),
		orderfoodModel.AICapabilityMealSuggest,
		"prompt-context-validation",
		orderfoodRequest.AIPromptValidationInput{
			Prompt: orderfoodRequest.AIPromptConfigInput{
				Mode:               orderfoodModel.AIPromptCustom,
				SystemPrompt:       &systemPrompt,
				UserPromptTemplate: &userTemplate,
			},
		},
	)
	if err != nil || validation.Valid ||
		!strings.Contains(strings.Join(validation.Errors, "；"), "上下文上限") {
		t.Fatalf("prompt context validation = %+v, err=%v", validation, err)
	}
	if err := db.Model(&orderfoodModel.AIModel{}).
		Where("id = ?", model.ID).
		Update("context_length", nil).Error; err != nil {
		t.Fatalf("clear model context limit: %v", err)
	}
	prompt, _, err := service.UpdateAIPrompt(
		ctx,
		aiTestActor(),
		orderfoodModel.AICapabilityMealSuggest,
		"prompt-update",
		orderfoodRequest.AIPromptUpdateInput{
			Prompt: orderfoodRequest.AIPromptConfigInput{
				Mode:               orderfoodModel.AIPromptCustom,
				SystemPrompt:       &systemPrompt,
				UserPromptTemplate: &userTemplate,
			},
			Reason:          "优化推荐提示词",
			ExpectedVersion: config.Version,
		},
	)
	if err != nil || prompt.Mode != "custom" || prompt.Version != 2 {
		t.Fatalf("update prompt: result=%+v err=%v", prompt, err)
	}
	promptWorkspace, err := service.AIPromptWorkspace(
		ctx,
		aiTestActor(),
		orderfoodModel.AICapabilityMealSuggest,
	)
	if err != nil || promptWorkspace.Current == nil ||
		promptWorkspace.Current.SystemPrompt != systemPrompt {
		t.Fatalf("prompt was not echoed as current config: %+v err=%v", promptWorkspace, err)
	}
	if promptWorkspace.FixedOutputSchema["type"] != "object" ||
		promptWorkspace.FixedOutputSchema["properties"] == nil {
		t.Fatalf("fixed output schema was not returned: %+v", promptWorkspace.FixedOutputSchema)
	}
	var capabilityConfigCount int64
	if err := db.Model(&orderfoodModel.AICapabilityConfig{}).
		Where("capability_code = ?", orderfoodModel.AICapabilityMealSuggest).
		Count(&capabilityConfigCount).Error; err != nil {
		t.Fatal(err)
	}
	if capabilityConfigCount != 1 {
		t.Fatalf("AI capability config count = %d, want 1", capabilityConfigCount)
	}
	var promptDefaultCount int64
	if err := db.Model(&orderfoodModel.AIPromptDefaultConfig{}).
		Count(&promptDefaultCount).Error; err != nil {
		t.Fatal(err)
	}
	if promptDefaultCount != int64(len(defaultCapabilities())) {
		t.Fatalf(
			"AI prompt default count = %d, want %d",
			promptDefaultCount,
			len(defaultCapabilities()),
		)
	}
	testResult, _, err := service.TestAICapabilityPrompt(
		ctx,
		aiTestActor(),
		orderfoodModel.AICapabilityMealSuggest,
		"prompt-test",
		orderfoodRequest.AIPromptTestInput{
			Prompt: orderfoodRequest.AIPromptConfigInput{
				Mode:               orderfoodModel.AIPromptCustom,
				SystemPrompt:       &systemPrompt,
				UserPromptTemplate: &userTemplate,
			},
			Variables: map[string]string{
				"candidate_dishes": "番茄炒蛋、青椒肉丝",
				"constraints":      "不吃辣",
				"servings":         "2",
				"target_count":     "2",
				"metadata":         `{"categories":["家常菜"],"tags":[],"units":[]}`,
			},
		},
	)
	if err != nil || !testResult.Success {
		t.Fatalf("test submitted prompt: result=%+v err=%v", testResult, err)
	}
}

// TestAIProviderServiceRejectsDirectInvalidInputs 验证服务层会拒绝绕过HTTP绑定的非法供应商配置、排序和连接测试参数。
func TestAIProviderServiceRejectsDirectInvalidInputs(t *testing.T) {
	service := &AIService{}
	_, _, err := service.CreateAIProvider(
		context.Background(),
		aiTestActor(),
		"invalid-provider-create",
		orderfoodRequest.AIProviderCreateInput{
			Name:      "",
			Type:      orderfoodModel.AIProviderDeepSeek,
			BaseURL:   "https://api.deepseek.com?api_key=should-not-be-here",
			APIKey:    "sk-test-deepseek",
			TimeoutMS: 999,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid direct provider create error = %v", err)
	}
	_, err = service.ListAIProviders(
		context.Background(),
		orderfoodRequest.AIProviderListQuery{SortOrder: "desc; DROP TABLE of_ai_providers"},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe direct provider sort error = %v", err)
	}
	tooLongTimeout := 30001
	_, _, err = service.TestAIProviderConnection(
		context.Background(),
		aiTestActor(),
		"provider-id",
		"invalid-provider-test",
		orderfoodRequest.AIProviderConnectionTestInput{
			TimeoutMS: &tooLongTimeout, ExpectedVersion: 1,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid direct provider connection test error = %v", err)
	}
}

// TestAIModelServiceRejectsDirectInvalidInputs 验证服务层会拒绝绕过HTTP绑定的非法模型配置、筛选和状态参数。
func TestAIModelServiceRejectsDirectInvalidInputs(t *testing.T) {
	service := &AIService{}
	_, _, err := service.CreateAIModel(
		context.Background(),
		aiTestActor(),
		"invalid-model-create",
		orderfoodRequest.AIModelCreateInput{
			ProviderID:   "provider-id",
			Name:         "",
			ModelKey:     strings.Repeat("m", 121),
			Capabilities: []orderfoodModel.AIModelCapability{orderfoodModel.AIModelText},
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid direct model create error = %v", err)
	}
	_, err = service.ListAIModels(
		context.Background(),
		orderfoodRequest.AIModelListQuery{SortOrder: "desc; DROP TABLE of_ai_models"},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe direct model sort error = %v", err)
	}
	_, _, err = service.UpdateAIModelStatus(
		context.Background(),
		aiTestActor(),
		"",
		"invalid-model-status",
		orderfoodRequest.AIStatusUpdateInput{
			Enabled: boolPointer(true),
			Reason:  "非法模型状态参数",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid direct model status error = %v", err)
	}
}

// TestAIMutationAuditFailureRollsBackMutationAndIdempotency 验证审计失败会回滚配置变更和幂等记录。
func TestAIMutationAuditFailureRollsBackMutationAndIdempotency(t *testing.T) {
	db := openAITestDB(t)
	service := NewAIService(
		db,
		NewPermissionService(db),
		NewIdempotencyService(db),
		fakeAIConnector{},
		WithAIMutationAuditWriter(failingAIMutationAuditWriter{}),
	)
	_, _, err := service.CreateAIProvider(
		context.Background(),
		aiTestActor(),
		"rollback-provider-key",
		orderfoodRequest.AIProviderCreateInput{
			Name:      "千问",
			Type:      orderfoodModel.AIProviderBailian,
			BaseURL:   "https://dashscope.example/v1",
			APIKey:    "sk-test-dashscope",
			TimeoutMS: 30000,
		},
	)
	if err == nil || appErrors.GetType(err) != appErrors.AdminInternal {
		t.Fatalf("audit failure error = %v", err)
	}
	var providerCount int64
	if countErr := db.Model(&orderfoodModel.AIProvider{}).Count(&providerCount).Error; countErr != nil {
		t.Fatalf("count providers: %v", countErr)
	}
	var idempotencyCount int64
	if countErr := db.Model(&orderfoodModel.AdminIdempotencyRecord{}).
		Count(&idempotencyCount).Error; countErr != nil {
		t.Fatalf("count idempotency records: %v", countErr)
	}
	if providerCount != 0 || idempotencyCount != 0 {
		t.Fatalf(
			"audit failure did not rollback: providers=%d idempotency=%d",
			providerCount,
			idempotencyCount,
		)
	}
}

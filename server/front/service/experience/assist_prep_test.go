package experience

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// prepTestDB 创建备菜顺序能力测试所需的独立 MySQL 数据库。
func prepTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&mealModel.FrontMeal{},
		&mealModel.MealParticipant{},
		&mealModel.FrontMealFinalDish{},
		&aiModel.FrontFeatureUsage{},
		&engagementModel.FrontPointEntry{},
		&mealModel.FrontPrepPlan{},
		&engagementModel.UserNotification{},
		&aiModel.PlatformCapabilityPolicy{},
		&engagementModel.SubscribeMessageTemplate{},
		&aiModel.AICapabilityDefinition{},
		&aiModel.AICapabilityConfig{},
		&aiModel.AIModel{},
		&aiModel.AIProvider{},
	); err != nil {
		t.Fatalf("migrate prep test database: %v", err)
	}
	return db
}

func seedPrepRuntime(t *testing.T, db *gorm.DB, endpoint string) {
	t.Helper()
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	labels, _ := json.Marshal([]map[string]interface{}{{
		"code": "prep_sequence", "title": "饭局备菜顺序", "actionLabel": "生成顺序",
		"description": "根据已确认菜品整理备菜顺序", "sortOrder": 1,
	}})
	stringList := datatypes.JSON([]byte(`["meal_snapshot","dish_steps","servings"]`))
	rows := []interface{}{
		&aiModel.PlatformCapabilityPolicy{
			SingletonKey: "platform", Version: 1,
			PlatformDefaultEnabled: true, FeatureLabelsJSON: datatypes.JSON(labels),
			AppliedByUsername: "test", AppliedAt: now, Reason: "test",
		},
		&aiModel.AIProvider{
			ID: "provider-1", Name: "test", Type: aiModel.AIProviderOpenAI,
			BaseURL: endpoint, APIKey: "secret",
			TimeoutMS: 1000, Enabled: true, UpdatedByUsername: "test",
			CreatedAt: now, UpdatedAt: now,
		},
		&aiModel.AIModel{
			ID: "model-1", ProviderID: "provider-1", Name: "test", ModelKey: "test-model",
			CapabilitiesJSON: datatypes.JSON([]byte(`["text"]`)), Enabled: true,
			UpdatedByUsername: "test", CreatedAt: now, UpdatedAt: now,
		},
		&aiModel.AICapabilityDefinition{
			Code: aiModel.AICapabilityPrepSequence, Name: "饭局备菜顺序",
			ClientFeatureCode: "prep_sequence", RequiredModelCapability: aiModel.AIModelText,
			SortOrder: 1, CreatedAt: now, UpdatedAt: now,
		},
		&aiModel.AICapabilityConfig{
			CapabilityCode: aiModel.AICapabilityPrepSequence,
			Version:        1, PrimaryModelID: "model-1", PointCost: 7,
			TimeoutMS:  1000,
			PromptMode: aiModel.AIPromptPreset, PromptPresetVersion: 1,
			SystemPrompt:               "按输入步骤生成顺序",
			UserPromptTemplate:         "{{meal_snapshot}}\n{{dish_steps}}\n{{servings}}",
			PromptAllowedVariablesJSON: stringList, PromptRequiredVariablesJSON: stringList,
			PromptOutputSchemaVersion: "v1", PromptHash: "test",
			AppliedByUsername: "test", AppliedAt: now, Reason: "test",
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed prep runtime: %v", err)
		}
	}
}

func seedConfirmedPrepMeal(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	steps, _ := json.Marshal([]mealDishStepSnapshot{
		{ID: "step-1", Description: "土豆切丝", SortOrder: 1},
		{ID: "step-2", Description: "下锅翻炒", SortOrder: 2},
	})
	rows := []interface{}{
		&userModel.MiniAppUser{
			ID: "user-1", OpenIDHash: "hash", OpenIDEncrypted: "encrypted",
			Nickname: "测试用户", Points: 100, Status: userModel.UserStatusNormal,
			RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
		},
		&mealModel.FrontMeal{
			ID: "meal-1", CreatorID: "user-1", Name: "晚饭", Code: "ABC123",
			Status: mealModel.MealConfirmed, DeadlineAt: now.Add(time.Hour),
			ConfirmedAt: &now, CreatedAt: now, UpdatedAt: now,
		},
		&mealModel.FrontMealFinalDish{
			ID: "final-1", MealID: "meal-1", CandidateID: "candidate-1",
			DishID: 1, Name: "炒土豆丝", FinalServings: 2,
			Ingredients: datatypes.JSON([]byte(`[]`)), Steps: datatypes.JSON(steps), CreatedAt: now,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed confirmed prep meal: %v", err)
		}
	}
}

func newPrepAssistService(t *testing.T, db *gorm.DB, client *http.Client) AssistService {
	t.Helper()
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	runtime := &RuntimeService{DB: db}
	engagement := &EngagementService{
		DB: db, Runtime: runtime, Now: func() time.Time { return now },
	}
	previousMealService := ServiceGroupApp.MealService
	ServiceGroupApp.MealService = MealService{DB: db, Now: func() time.Time { return now }}
	t.Cleanup(func() { ServiceGroupApp.MealService = previousMealService })
	return AssistService{
		DB: db, Engagement: engagement, Client: client, Now: func() time.Time { return now },
	}
}

type prepRoundTripFunc func(*http.Request) (*http.Response, error)

func (function prepRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func prepModelClient(t *testing.T, content string, calls *atomic.Int64) *http.Client {
	t.Helper()
	return &http.Client{Transport: prepRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		var body strings.Builder
		if err := json.NewEncoder(&body).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{
				"message": map[string]string{"content": content},
			}},
		}); err != nil {
			t.Errorf("encode model response: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body.String())),
			Request:    request,
		}, nil
	})}
}

// TestGeneratePrepPlanUsesCurrentModelAndPersistsValidatedSnapshotSteps 验证备菜计划使用当前模型并保存校验后的快照步骤。
func TestGeneratePrepPlanUsesCurrentModelAndPersistsValidatedSnapshotSteps(t *testing.T) {
	var calls atomic.Int64
	client := prepModelClient(
		t,
		`{"estimatedMinutes":25,"steps":[`+
			`{"sourceStepId":"step-2","parallelSourceStepIds":["step-1"]},`+
			`{"sourceStepId":"step-1","parallelSourceStepIds":[]}]}`,
		&calls,
	)

	db := prepTestDB(t)
	seedPrepRuntime(t, db, "https://provider.test")
	seedConfirmedPrepMeal(t, db)
	service := newPrepAssistService(t, db, client)

	result, err := service.GeneratePrepPlan(context.Background(), "user-1", frontRequest.PrepPlanInput{
		MealID: "meal-1",
	})
	if err != nil {
		t.Fatalf("generate prep plan: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("model calls = %d, want 1", calls.Load())
	}
	if result.EstimatedMinutes != 25 || len(result.Steps) != 2 ||
		result.Steps[0].ID != "step-2" || result.Steps[0].Instruction != "下锅翻炒" {
		t.Fatalf("unexpected validated preparation plan: %+v", result)
	}
	if len(result.Steps[0].ParallelActions) != 1 ||
		result.Steps[0].ParallelActions[0] != "炒土豆丝：土豆切丝" {
		t.Fatalf("parallel action was not derived from the snapshot: %+v", result.Steps[0])
	}
	var user userModel.MiniAppUser
	var usage aiModel.FrontFeatureUsage
	var planCount int64
	_ = db.First(&user, "id = ?", "user-1").Error
	_ = db.First(&usage, "id = ?", result.Usage.UsageID).Error
	_ = db.Model(&mealModel.FrontPrepPlan{}).Count(&planCount).Error
	if user.Points != 93 || usage.ExecutionStatus != "succeeded" ||
		usage.BillingStatus != "charged" || planCount != 1 {
		t.Fatalf("unexpected persisted success state: user=%+v usage=%+v plans=%d", user, usage, planCount)
	}
}

func TestGeneratePrepPlanRejectsUnknownModelStepAndRefundsPoints(t *testing.T) {
	var calls atomic.Int64
	client := prepModelClient(
		t,
		`{"estimatedMinutes":25,"steps":[`+
			`{"sourceStepId":"unknown","parallelSourceStepIds":[]},`+
			`{"sourceStepId":"step-1","parallelSourceStepIds":[]}]}`,
		&calls,
	)

	db := prepTestDB(t)
	seedPrepRuntime(t, db, "https://provider.test")
	seedConfirmedPrepMeal(t, db)
	service := newPrepAssistService(t, db, client)

	_, err := service.GeneratePrepPlan(context.Background(), "user-1", frontRequest.PrepPlanInput{
		MealID: "meal-1",
	})
	if appErrors.GetType(err) != appErrors.FrontResultInvalid {
		t.Fatalf("error type = %d, want %d: %v", appErrors.GetType(err), appErrors.FrontResultInvalid, err)
	}
	if calls.Load() != 1 {
		t.Fatalf("model calls = %d, want 1", calls.Load())
	}
	var user userModel.MiniAppUser
	var usage aiModel.FrontFeatureUsage
	var refund engagementModel.FrontPointEntry
	var planCount int64
	_ = db.First(&user, "id = ?", "user-1").Error
	_ = db.Order("created_at desc").First(&usage).Error
	_ = db.First(&refund, "type = ?", engagementModel.PointRefund).Error
	_ = db.Model(&mealModel.FrontPrepPlan{}).Count(&planCount).Error
	if user.Points != 100 || usage.ExecutionStatus != "failed" ||
		usage.BillingStatus != "refunded" || refund.Amount != 7 ||
		refund.RelatedObjectType == nil || *refund.RelatedObjectType != "feature_usage" ||
		refund.RelatedObjectID == nil || *refund.RelatedObjectID != usage.ID ||
		refund.RelatedEntryID == nil || *refund.RelatedEntryID == "" || planCount != 0 {
		t.Fatalf(
			"unexpected refund state: user=%+v usage=%+v refund=%+v plans=%d",
			user, usage, refund, planCount,
		)
	}
}

func TestValidatePrepModelOutputRejectsExcludedOrder(t *testing.T) {
	sources := []prepSourceStep{
		{ID: "step-1", DishName: "菜一", Instruction: "步骤一"},
		{ID: "step-2", DishName: "菜二", Instruction: "步骤二"},
	}
	_, err := validatePrepModelOutput(
		prepModelOutput{
			EstimatedMinutes: 20,
			Steps: []prepModelStep{
				{SourceStepID: "step-1"},
				{SourceStepID: "step-2"},
			},
		},
		sources,
		[]prepExcludedPlan{{PlanID: "old-plan", StepIDs: []string{"step-1", "step-2"}}},
	)
	if appErrors.GetType(err) != appErrors.FrontResultInvalid {
		t.Fatalf("error type = %d, want result invalid: %v", appErrors.GetType(err), err)
	}
}

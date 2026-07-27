package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type suggestionRoundTripFunc func(*http.Request) (*http.Response, error)

func (function suggestionRoundTripFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return function(request)
}

func suggestionProviderClient(
	calls *atomic.Int32,
	content func(call int32) string,
) *http.Client {
	return &http.Client{
		Transport: suggestionRoundTripFunc(func(request *http.Request) (*http.Response, error) {
			call := calls.Add(1)
			payload, _ := json.Marshal(map[string]interface{}{
				"choices": []map[string]interface{}{{
					"message": map[string]string{"content": content(call)},
				}},
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(payload)),
				Request:    request,
			}, nil
		}),
	}
}

// suggestionGenerationTestDB 创建推荐菜生成能力测试所需的独立 MySQL 数据库。
func suggestionGenerationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.UserPreferenceProfile{},
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.ContentTag{},
		&orderfoodModel.ContentUnit{},
		&orderfoodModel.UserDish{},
		&orderfoodModel.PlatformRecommendation{},
		&orderfoodModel.StandardIngredient{},
		&orderfoodModel.StandardDishIndex{},
		&orderfoodModel.StandardDishIngredient{},
		&orderfoodModel.SuggestionValidationPolicy{},
		&orderfoodModel.PlatformCapabilityPolicy{},
		&orderfoodModel.SubscribeMessageTemplate{},
		&orderfoodModel.FrontFeatureUsage{},
		&orderfoodModel.FrontPointEntry{},
		&orderfoodModel.FrontMealSuggestion{},
		&orderfoodModel.FrontMealSuggestionDish{},
		&orderfoodModel.UserNotification{},
		&orderfoodModel.AICapabilityDefinition{},
		&orderfoodModel.AICapabilityConfig{},
		&orderfoodModel.AIModel{},
		&orderfoodModel.AIProvider{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedSuggestionGenerationRuntime(
	t *testing.T,
	db *gorm.DB,
	endpoint string,
) {
	t.Helper()
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	category := orderfoodModel.ContentCategory{
		PublicID: "category-main", Name: "主菜", Enabled: true,
	}
	tag := orderfoodModel.ContentTag{
		PublicID: "tag-quick", Name: "快手", Enabled: true,
	}
	unit := orderfoodModel.ContentUnit{
		PublicID: "unit-piece", Name: "个", Enabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&unit).Error; err != nil {
		t.Fatal(err)
	}
	tomato := orderfoodModel.StandardIngredient{
		ID: "ingredient-tomato", Name: "番茄",
		AliasesJSON: datatypes.JSON(`["西红柿"]`), Enabled: true,
		Version: 1, UpdatedByUsername: "test", LastChangeReason: "test",
		CreatedAt: now, UpdatedAt: now,
	}
	egg := orderfoodModel.StandardIngredient{
		ID: "ingredient-egg", Name: "鸡蛋",
		AliasesJSON: datatypes.JSON(`[]`), Enabled: true,
		Version: 1, UpdatedByUsername: "test", LastChangeReason: "test",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&tomato).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&egg).Error; err != nil {
		t.Fatal(err)
	}
	dish := orderfoodModel.StandardDishIndex{
		ID: "standard-dish-tomato-egg", Name: "番茄炒蛋",
		AliasesJSON: datatypes.JSON(`["西红柿炒鸡蛋"]`),
		Cuisine:     "家常菜", CategoryID: category.ID,
		SourceName: "test", Enabled: true, Version: 1,
		UpdatedByUsername: "test", LastChangeReason: "test",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatal(err)
	}
	relations := []orderfoodModel.StandardDishIngredient{
		{
			DishID: dish.ID, IngredientID: tomato.ID,
			Required: true, SortOrder: 1, CreatedAt: now,
		},
		{
			DishID: dish.ID, IngredientID: egg.ID,
			Required: true, SortOrder: 2, CreatedAt: now,
		},
	}
	if err := db.Create(&relations).Error; err != nil {
		t.Fatal(err)
	}
	rows := []interface{}{
		&orderfoodModel.PlatformCapabilityPolicy{
			SingletonKey: "platform", Version: 1, PlatformDefaultEnabled: true,
			FeatureLabelsJSON: datatypes.JSON(`[{"code":"meal_suggest","title":"不知道吃什么","actionLabel":"帮我选菜","description":"生成菜品建议","sortOrder":1}]`),
			AppliedByUsername: "test", AppliedAt: now, Reason: "test",
		},
		&orderfoodModel.SuggestionValidationPolicy{
			SingletonKey: "platform", Version: 1, CatalogValidationEnabled: true,
			AppliedByUsername: "test", AppliedAt: now, Reason: "test",
		},
		&orderfoodModel.AIProvider{
			ID: "suggestion-provider", Name: "test",
			Type: orderfoodModel.AIProviderOpenAI, BaseURL: endpoint,
			APIKey: "test-key",
			TimeoutMS:     1000, Enabled: true, Version: 1,
			UpdatedByUsername: "test", CreatedAt: now, UpdatedAt: now,
		},
		&orderfoodModel.AIModel{
			ID: "suggestion-model", ProviderID: "suggestion-provider",
			Name: "test", ModelKey: "test-model",
			CapabilitiesJSON: datatypes.JSON(`["text"]`),
			Enabled:          true, Version: 1, UpdatedByUsername: "test",
			CreatedAt: now, UpdatedAt: now,
		},
		&orderfoodModel.AICapabilityDefinition{
			Code: orderfoodModel.AICapabilityMealSuggest, Name: "不知道吃什么",
			ClientFeatureCode:       "meal_suggest",
			RequiredModelCapability: orderfoodModel.AIModelText,
			SortOrder:               1, CreatedAt: now, UpdatedAt: now,
		},
		&orderfoodModel.AICapabilityConfig{
			CapabilityCode: orderfoodModel.AICapabilityMealSuggest,
			Version:        1, PrimaryModelID: "suggestion-model", PointCost: 7,
			TimeoutMS:  1000,
			PromptMode: orderfoodModel.AIPromptPreset, PromptPresetVersion: 2,
			SystemPrompt: "生成菜品", UserPromptTemplate: "{{metadata}}\n{{target_count}}",
			PromptAllowedVariablesJSON:  datatypes.JSON(`["metadata","target_count"]`),
			PromptRequiredVariablesJSON: datatypes.JSON(`["metadata","target_count"]`),
			PromptOutputSchemaVersion:   "meal_suggestion_v2", PromptHash: "test",
			AppliedByUsername: "test", AppliedAt: now, Reason: "test",
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func TestSuggestionCatalogValidationRetriesOnceThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	validOutput := map[string]interface{}{
		"reason": "一人份快手家常菜",
		"dishes": []map[string]interface{}{{
			"name": "西红柿炒鸡蛋", "cuisine": "家常菜",
			"category": "主菜", "tags": []string{"快手"},
			"serving": 1, "description": "酸甜下饭，适合一人份。",
			"ingredients": []map[string]interface{}{
				{"name": "西红柿", "amount": "1", "unit": "个", "note": nil},
				{"name": "鸡蛋", "amount": "2", "unit": "个", "note": nil},
			},
			"steps": []map[string]interface{}{
				{"text": "西红柿切块，鸡蛋打散。"},
				{"text": "炒熟鸡蛋后加入西红柿翻炒。"},
			},
		}},
	}
	client := suggestionProviderClient(&calls, func(call int32) string {
		content := `{"reason":"invalid","dishes":[]}`
		if call == 2 {
			encoded, _ := json.Marshal(validOutput)
			content = string(encoded)
		}
		return content
	})
	db := suggestionGenerationTestDB(t)
	seedSuggestionGenerationRuntime(t, db, "http://suggestion.test")
	service := &AssistService{DB: db, Client: client}

	dishes, _, err := service.generateSuggestionDishes(
		context.Background(),
		frontRequest.MealSuggestionInput{People: 1, Tags: []string{"快手"}},
		1,
		map[string]struct{}{},
		nil,
		true,
	)
	if err != nil {
		t.Fatalf("generate suggestion with retry: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("provider calls = %d, want 2", calls.Load())
	}
	if len(dishes) != 1 ||
		dishes[0].Dish.Name != "番茄炒蛋" ||
		dishes[0].Dish.Source != "generated" {
		t.Fatalf("unexpected generated dish: %+v", dishes)
	}
}

func TestSuggestionWithoutCatalogValidationDoesNotRetry(t *testing.T) {
	var calls atomic.Int32
	client := suggestionProviderClient(&calls, func(int32) string {
		return `{"reason":"invalid","dishes":[]}`
	})
	db := suggestionGenerationTestDB(t)
	seedSuggestionGenerationRuntime(t, db, "http://suggestion.test")
	service := &AssistService{DB: db, Client: client}

	_, _, err := service.generateSuggestionDishes(
		context.Background(),
		frontRequest.MealSuggestionInput{People: 1},
		1,
		map[string]struct{}{},
		nil,
		false,
	)
	if err == nil {
		t.Fatal("invalid result unexpectedly succeeded")
	}
	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
}

// TestSuggestionWithoutCatalogValidationAcceptsNonIndexedDish 校验关闭索引时只跳过目录匹配且仍保留基础结构校验。
func TestSuggestionWithoutCatalogValidationAcceptsNonIndexedDish(t *testing.T) {
	var calls atomic.Int32
	output := map[string]interface{}{
		"reason": "一人份快手家常菜",
		"dishes": []map[string]interface{}{{
			"name": "青椒炒肉", "cuisine": "家常菜",
			"category": "主菜", "tags": []string{"快手"},
			"serving": 1, "description": "咸香下饭，适合一人份。",
			"ingredients": []map[string]interface{}{
				{"name": "青椒", "amount": "1", "unit": "个", "note": nil},
				{"name": "猪肉", "amount": "100", "unit": "个", "note": nil},
			},
			"steps": []map[string]interface{}{
				{"text": "青椒切块，猪肉切片。"},
				{"text": "先炒猪肉，再加入青椒翻炒。"},
			},
		}},
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	client := suggestionProviderClient(&calls, func(int32) string {
		return string(encoded)
	})
	db := suggestionGenerationTestDB(t)
	seedSuggestionGenerationRuntime(t, db, "http://suggestion.test")
	service := &AssistService{DB: db, Client: client}

	dishes, _, err := service.generateSuggestionDishes(
		context.Background(),
		frontRequest.MealSuggestionInput{People: 1, Tags: []string{"快手"}},
		1,
		map[string]struct{}{},
		nil,
		false,
	)
	if err != nil {
		t.Fatalf("generate non-indexed suggestion with catalog disabled: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
	if len(dishes) != 1 || dishes[0].Dish.Name != "青椒炒肉" {
		t.Fatalf("unexpected non-indexed generated dish: %+v", dishes)
	}
}

func TestSuggestionCatalogValidationRefundsAfterTwoFailedAttempts(t *testing.T) {
	var calls atomic.Int32
	client := suggestionProviderClient(&calls, func(int32) string {
		return `{"reason":"invalid","dishes":[]}`
	})
	db := suggestionGenerationTestDB(t)
	seedSuggestionGenerationRuntime(t, db, "http://suggestion.test")
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	if err := db.Create(&orderfoodModel.MiniAppUser{
		ID: "suggestion-user", OpenIDHash: "suggestion-user-hash",
		OpenIDEncrypted: "suggestion-user-encrypted", Nickname: "测试用户",
		Points: 100, CheckinDayCount: 7, Status: orderfoodModel.UserStatusNormal,
		Version: 1, RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed suggestion user: %v", err)
	}
	engagement := &EngagementService{
		DB: db, Runtime: &RuntimeService{DB: db}, Now: func() time.Time { return now },
	}
	service := &AssistService{
		DB: db, Engagement: engagement, Client: client, Now: func() time.Time { return now },
	}

	_, err := service.CreateSuggestion(
		context.Background(),
		"suggestion-user",
		frontRequest.MealSuggestionInput{People: 1},
	)
	if err == nil {
		t.Fatal("invalid generated results unexpectedly succeeded")
	}
	if calls.Load() != 2 {
		t.Fatalf("provider calls = %d, want 2", calls.Load())
	}
	var user orderfoodModel.MiniAppUser
	var usage orderfoodModel.FrontFeatureUsage
	var refund orderfoodModel.FrontPointEntry
	var suggestionCount int64
	_ = db.First(&user, "id = ?", "suggestion-user").Error
	_ = db.Order("created_at desc").First(&usage).Error
	_ = db.First(&refund, "type = ?", orderfoodModel.PointRefund).Error
	_ = db.Model(&orderfoodModel.FrontMealSuggestion{}).Count(&suggestionCount).Error
	if user.Points != 100 || usage.ExecutionStatus != "failed" ||
		usage.BillingStatus != "refunded" || refund.Amount != 7 ||
		refund.RelatedObjectType == nil || *refund.RelatedObjectType != "feature_usage" ||
		refund.RelatedObjectID == nil || *refund.RelatedObjectID != usage.ID ||
		refund.RelatedEntryID == nil || *refund.RelatedEntryID == "" ||
		suggestionCount != 0 {
		t.Fatalf(
			"unexpected suggestion refund state: user=%+v usage=%+v refund=%+v suggestions=%d",
			user,
			usage,
			refund,
			suggestionCount,
		)
	}
}

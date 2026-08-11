package experience

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
)

// TestRuntimeServiceUsesOnlyCurrentPlatformSwitch 验证运行态只读取当前平台总开关。
func TestRuntimeServiceUsesOnlyCurrentPlatformSwitch(t *testing.T) {
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&aiModel.PlatformCapabilityPolicy{},
		&aiModel.AICapabilityDefinition{},
		&aiModel.AICapabilityConfig{},
		&engagementModel.SubscribeMessageTemplate{},
	); err != nil {
		t.Fatal(err)
	}
	staleCostHint := "旧积分提示"
	staleFreeQuotaHint := "旧免费提示"
	labels, _ := json.Marshal([]map[string]interface{}{{
		"code": "meal_suggest", "title": "不知道吃什么", "actionLabel": "帮我选菜",
		"description": "按真实菜品提供建议", "costHint": staleCostHint,
		"freeQuotaHint": staleFreeQuotaHint, "sortOrder": 1,
	}})
	now := time.Now().UTC()
	disabled := aiModel.PlatformCapabilityPolicy{
		SingletonKey: "platform", Version: 1,
		PlatformDefaultEnabled: false, EmergencyDisabled: false,
		FeatureLabelsJSON: datatypes.JSON(labels), AppliedByUsername: "test",
		AppliedAt: now, Reason: "initial",
	}
	if err := db.Create(&disabled).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&aiModel.AICapabilityDefinition{
		Code:                    aiModel.AICapabilityMealSuggest,
		Name:                    "不知道吃什么",
		ClientFeatureCode:       "meal_suggest",
		RequiredModelCapability: aiModel.AIModelText,
		SortOrder:               1,
		CreatedAt:               now,
		UpdatedAt:               now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&aiModel.AICapabilityConfig{
		CapabilityCode:              aiModel.AICapabilityMealSuggest,
		Version:                     1,
		PrimaryModelID:              "model-1",
		PointCost:                   2,
		FreeQuotaPerDay:             3,
		PromptMode:                  aiModel.AIPromptPreset,
		PromptAllowedVariablesJSON:  datatypes.JSON(`[]`),
		PromptRequiredVariablesJSON: datatypes.JSON(`[]`),
		AppliedByUsername:           "test",
		AppliedAt:                   now,
		Reason:                      "test",
	}).Error; err != nil {
		t.Fatal(err)
	}
	service := RuntimeService{DB: db}
	config, err := service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if config.EnhancedFeaturesEnabled || config.PointsEnabled || len(config.Features) != 0 {
		t.Fatalf("disabled platform leaked features: %+v", config)
	}
	enabled := disabled
	enabled.Version = 2
	enabled.PlatformDefaultEnabled = true
	enabled.Reason = "enable all"
	if err := db.Save(&enabled).Error; err != nil {
		t.Fatal(err)
	}
	config, err = service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !config.EnhancedFeaturesEnabled || !config.PointsEnabled || len(config.Features) != 1 {
		t.Fatalf("enabled platform did not expose all configured features: %+v", config)
	}
	if config.Features[0].CostHint == nil || *config.Features[0].CostHint != "每次消耗 2 积分" ||
		config.Features[0].FreeQuotaHint == nil ||
		*config.Features[0].FreeQuotaHint != "每天免费 3 次" {
		t.Fatalf("live capability billing hints were not returned: %+v", config.Features[0])
	}
}

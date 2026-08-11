package model_test

import (
	"reflect"
	"testing"

	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
)

// TestDirectConfigurationModelsUseCurrentTables 验证配置模型只映射当前记录表。
func TestDirectConfigurationModelsUseCurrentTables(t *testing.T) {
	tableNames := map[string]string{
		"point rule":        (engagementModel.PointRuleConfig{}).TableName(),
		"suggestion policy": (dishModel.SuggestionValidationPolicy{}).TableName(),
		"moderation":        (contentModel.ModerationConfig{}).TableName(),
		"platform policy":   (aiModel.PlatformCapabilityPolicy{}).TableName(),
		"AI capability":     (aiModel.AICapabilityConfig{}).TableName(),
		"prompt default":    (aiModel.AIPromptDefaultConfig{}).TableName(),
		"wechat":            (userModel.WeChatConfig{}).TableName(),
	}
	expected := map[string]string{
		"point rule":        "of_point_rule",
		"suggestion policy": "of_suggest_policy",
		"moderation":        "of_mod_config",
		"platform policy":   "of_ai_policy",
		"AI capability":     "of_ai_cap_config",
		"prompt default":    "of_ai_prompt_defaults",
		"wechat":            "of_wx_config",
	}
	for name, tableName := range tableNames {
		if tableName != expected[name] {
			t.Fatalf("%s table = %q, want %q", name, tableName, expected[name])
		}
	}
}

// TestAICapabilityConfigHasNoFakeRuntimeControls 验证能力配置没有备用模型、单项开关和不生效的结果策略。
func TestAICapabilityConfigHasNoFakeRuntimeControls(t *testing.T) {
	configType := reflect.TypeOf(aiModel.AICapabilityConfig{})
	for _, fieldName := range []string{
		"FallbackModelID",
		"Enabled",
		"RetryCount",
		"ResultValidationPolicyJSON",
	} {
		if _, exists := configType.FieldByName(fieldName); exists {
			t.Fatalf("AI capability config must not contain %s", fieldName)
		}
	}
	definitionType := reflect.TypeOf(aiModel.AICapabilityDefinition{})
	if _, exists := definitionType.FieldByName("CurrentVersion"); exists {
		t.Fatal("AI capability definition must not retain a historical version pointer")
	}
}

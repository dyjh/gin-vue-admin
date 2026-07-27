package model

import (
	"reflect"
	"testing"
)

// TestDirectConfigurationModelsUseCurrentTables 验证配置模型只映射当前记录表。
func TestDirectConfigurationModelsUseCurrentTables(t *testing.T) {
	tableNames := map[string]string{
		"point rule":        (PointRuleConfig{}).TableName(),
		"suggestion policy": (SuggestionValidationPolicy{}).TableName(),
		"moderation":        (ModerationConfig{}).TableName(),
		"platform policy":   (PlatformCapabilityPolicy{}).TableName(),
		"AI capability":     (AICapabilityConfig{}).TableName(),
		"prompt default":    (AIPromptDefaultConfig{}).TableName(),
		"wechat":            (WeChatConfig{}).TableName(),
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
	configType := reflect.TypeOf(AICapabilityConfig{})
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
	definitionType := reflect.TypeOf(AICapabilityDefinition{})
	if _, exists := definitionType.FieldByName("CurrentVersion"); exists {
		t.Fatal("AI capability definition must not retain a historical version pointer")
	}
}

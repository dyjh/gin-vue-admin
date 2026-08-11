package ai

import (
	"gorm.io/datatypes"
	"time"
)

// AIPromptMode 表示能力使用平台默认或自定义提示词。
type AIPromptMode string

const (
	AIPromptPreset AIPromptMode = "preset"
	AIPromptCustom AIPromptMode = "custom"
)

// AICapabilityConfig 表示当前生效的AI能力配置。
type AICapabilityConfig struct {
	CapabilityCode              string         `json:"capabilityCode" gorm:"column:capability_code;type:varchar(64);primaryKey;not null;comment:AI能力编码;"`                 // AI能力编码
	Version                     int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                        // 数据版本
	PrimaryModelID              string         `json:"primaryModelId" gorm:"column:primary_model_id;type:varchar(64);not null;index;comment:主模型ID;"`                      // 主模型ID
	AuxiliaryModelID            *string        `json:"auxiliaryModelId" gorm:"column:auxiliary_model_id;type:varchar(64);default:null;index;comment:辅助模型ID;"`             // 辅助模型ID
	PointCost                   int            `json:"pointCost" gorm:"column:point_cost;type:int;not null;comment:积分消耗;"`                                                // 积分消耗
	DailyLimitPerUser           int            `json:"dailyLimitPerUser" gorm:"column:daily_limit_per_user;type:int;not null;comment:单用户每日调用上限;"`                         // 单用户每日调用上限
	TimeoutMS                   int            `json:"timeoutMs" gorm:"column:timeout_ms;type:int;not null;comment:超时时间（毫秒）;"`                                            // 超时时间（毫秒）
	FreeQuotaPerDay             int            `json:"freeQuotaPerDay" gorm:"column:free_quota_per_day;type:int;not null;comment:每日免费额度;"`                                // 每日免费额度
	PromptMode                  AIPromptMode   `json:"promptMode" gorm:"column:prompt_mode;type:varchar(16);not null;comment:提示词模式;"`                                     // 提示词模式
	PromptPresetVersion         int64          `json:"promptPresetVersion" gorm:"column:prompt_preset_version;type:bigint;not null;comment:默认提示词版本;"`                     // 默认提示词版本
	SystemPrompt                string         `json:"-" gorm:"column:system_prompt;type:text;not null;comment:系统提示词;"`                                                   // 系统提示词
	UserPromptTemplate          string         `json:"-" gorm:"column:user_prompt_template;type:text;not null;comment:用户提示词模板;"`                                          // 用户提示词模板
	PromptAllowedVariablesJSON  datatypes.JSON `json:"-" gorm:"column:prompt_allowed_variables_json;type:json;not null;comment:提示词允许变量集合JSON;"`                           // 提示词允许变量集合JSON
	PromptRequiredVariablesJSON datatypes.JSON `json:"-" gorm:"column:prompt_required_variables_json;type:json;not null;comment:提示词必填变量集合JSON;"`                          // 提示词必填变量集合JSON
	PromptOutputSchemaVersion   string         `json:"promptOutputSchemaVersion" gorm:"column:prompt_output_schema_version;type:varchar(40);not null;comment:提示词输出结构版本;"` // 提示词输出结构版本
	PromptHash                  string         `json:"promptHash" gorm:"column:prompt_hash;type:varchar(64);not null;comment:提示词摘要;"`                                     // 提示词摘要
	AppliedByID                 uint           `json:"-" gorm:"column:applied_by_id;type:bigint unsigned;not null;comment:应用管理员ID;"`                                      // 应用管理员ID
	AppliedByUsername           string         `json:"-" gorm:"column:applied_by_username;type:varchar(80);not null;comment:应用管理员用户名;"`                                   // 应用管理员用户名
	AppliedByNickname           *string        `json:"-" gorm:"column:applied_by_nickname;type:varchar(80);default:null;comment:应用管理员昵称;"`                                // 应用管理员昵称
	AppliedAt                   time.Time      `json:"appliedAt" gorm:"column:applied_at;type:datetime;not null;index;comment:应用时间;"`                                     // 应用时间
	Reason                      string         `json:"reason" gorm:"column:reason;type:varchar(200);not null;comment:原因;"`                                                // 原因
}

// TableName 指定AICapabilityConfig对应的数据表名。
func (AICapabilityConfig) TableName() string {
	return "of_ai_cap_config"
}

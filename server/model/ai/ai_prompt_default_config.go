package ai

import (
	"gorm.io/datatypes"
	"time"
)

// AIPromptDefaultConfig 表示每项能力当前的平台默认提示词。
type AIPromptDefaultConfig struct {
	CapabilityCode        string         `json:"capabilityCode" gorm:"column:capability_code;type:varchar(64);primaryKey;not null;comment:AI能力编码;"` // AI能力编码
	Version               int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                        // 数据版本
	SystemPrompt          string         `json:"-" gorm:"column:system_prompt;type:text;not null;comment:系统提示词;"`                                   // 系统提示词
	UserPromptTemplate    string         `json:"-" gorm:"column:user_prompt_template;type:text;not null;comment:用户提示词模板;"`                          // 用户提示词模板
	AllowedVariablesJSON  datatypes.JSON `json:"-" gorm:"column:allowed_variables_json;type:json;not null;comment:允许变量集合JSON;"`                     // 允许变量集合JSON
	RequiredVariablesJSON datatypes.JSON `json:"-" gorm:"column:required_variables_json;type:json;not null;comment:必填变量集合JSON;"`                    // 必填变量集合JSON
	OutputSchemaVersion   string         `json:"outputSchemaVersion" gorm:"column:output_schema_version;type:varchar(40);not null;comment:输出结构版本;"` // 输出结构版本
	ContentHash           string         `json:"contentHash" gorm:"column:content_hash;type:varchar(64);not null;comment:内容摘要;"`                    // 内容摘要
	UpdatedAt             time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                           // 更新时间
}

// TableName 指定AIPromptDefaultConfig对应的数据表名。
func (AIPromptDefaultConfig) TableName() string {
	return "of_ai_prompt_defaults"
}

package ai

import (
	"time"
)

// AIPromptTestRecord 表示AI提示词测试记录。
type AIPromptTestRecord struct {
	ID                    string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                            // 主键ID
	CapabilityCode        string    `json:"capabilityCode" gorm:"column:capability_code;type:varchar(64);not null;index:idx_of_prompt_test_config,priority:1;comment:AI能力编码;"` // AI能力编码
	ConfigVersion         int64     `json:"configVersion" gorm:"column:config_version;type:bigint;not null;index:idx_of_prompt_test_config,priority:2;comment:配置版本;"`          // 配置版本
	ModelID               string    `json:"modelId" gorm:"column:model_id;type:varchar(64);not null;index;comment:模型ID;"`                                                      // 模型ID
	PromptHash            string    `json:"promptHash" gorm:"column:prompt_hash;type:varchar(64);not null;comment:提示词摘要;"`                                                     // 提示词摘要
	AdministratorID       uint      `json:"-" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`                                               // 管理员ID
	AdministratorUsername string    `json:"-" gorm:"column:administrator_username;type:varchar(80);not null;comment:管理员用户名;"`                                                  // 管理员用户名
	RequestID             string    `json:"requestId" gorm:"column:request_id;type:varchar(80);not null;index;comment:请求ID;"`                                                  // 请求ID
	Success               bool      `json:"success" gorm:"column:success;type:tinyint(1) unsigned;not null;index;comment:是否成功;"`                                               // 是否成功
	DurationMS            int       `json:"durationMs" gorm:"column:duration_ms;type:int;not null;comment:耗时（毫秒）;"`                                                            // 耗时（毫秒）
	InputTokens           *int      `json:"inputTokens" gorm:"column:input_tokens;type:int;default:null;comment:输入Token数;"`                                                    // 输入Token数
	OutputTokens          *int      `json:"outputTokens" gorm:"column:output_tokens;type:int;default:null;comment:输出Token数;"`                                                  // 输出Token数
	EstimatedCostCNY      *string   `json:"estimatedCostCny" gorm:"column:estimated_cost_cny;type:varchar(40);default:null;comment:预估成本（元）;"`                                  // 预估成本（元）
	SafeMessage           string    `json:"safeMessage" gorm:"column:safe_message;type:varchar(240);not null;comment:可安全展示的提示;"`                                               // 可安全展示的提示
	CreatedAt             time.Time `json:"testedAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                                      // 创建时间
}

// TableName 指定AIPromptTestRecord对应的数据表名。
func (AIPromptTestRecord) TableName() string {
	return "of_ai_prompt_tests"
}

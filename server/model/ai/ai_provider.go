package ai

import (
	"time"
)

// AIProviderType 表示平台支持的AI供应商类型。
type AIProviderType string

const (
	AIProviderBailian  AIProviderType = "bailian"
	AIProviderDeepSeek AIProviderType = "deepseek"
	AIProviderOpenAI   AIProviderType = "openai"
)

// AIProvider stores the provider API key in the database. It is deliberately
// from JSON and no response DTO has a field capable of returning it.
type AIProvider struct {
	ID                  string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                      // 主键ID
	Name                string         `json:"name" gorm:"column:name;type:varchar(60);not null;comment:名称;"`                               // 名称
	Type                AIProviderType `json:"type" gorm:"column:type;type:varchar(24);not null;index;comment:类型;"`                         // 类型
	BaseURL             string         `json:"baseUrl" gorm:"column:base_url;type:varchar(300);not null;comment:供应商基础地址;"`                  // 供应商基础地址
	APIKey              string         `json:"-" gorm:"column:api_key;type:text;not null;comment:API密钥（明文）;"`                               // API密钥（明文）
	TimeoutMS           int            `json:"timeoutMs" gorm:"column:timeout_ms;type:int;not null;comment:超时时间（毫秒）;"`                      // 超时时间（毫秒）
	RetryCount          int            `json:"retryCount" gorm:"column:retry_count;type:int;not null;comment:重试次数;"`                        // 重试次数
	Enabled             bool           `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;index;comment:是否启用;"`         // 是否启用
	Remark              *string        `json:"remark" gorm:"column:remark;type:varchar(300);default:null;comment:备注;"`                      // 备注
	Version             int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                  // 数据版本
	LastTestSuccess     *bool          `json:"-" gorm:"column:last_test_success;type:tinyint(1) unsigned;default:null;comment:最近连接测试是否成功;"` // 最近连接测试是否成功
	LastTestCategory    *string        `json:"-" gorm:"column:last_test_category;type:varchar(40);default:null;comment:最近连接测试分类;"`          // 最近连接测试分类
	LastTestDurationMS  *int           `json:"-" gorm:"column:last_test_duration_ms;type:int;default:null;comment:最近连接测试耗时（毫秒）;"`           // 最近连接测试耗时（毫秒）
	LastTestSafeMessage *string        `json:"-" gorm:"column:last_test_safe_message;type:varchar(240);default:null;comment:最近连接测试安全提示;"`   // 最近连接测试安全提示
	LastTestedAt        *time.Time     `json:"-" gorm:"column:last_tested_at;type:datetime;default:null;comment:最近连接测试时间;"`                 // 最近连接测试时间
	UpdatedByID         uint           `json:"-" gorm:"column:updated_by_id;type:bigint unsigned;not null;comment:更新管理员ID;"`                // 更新管理员ID
	UpdatedByUsername   string         `json:"-" gorm:"column:updated_by_username;type:varchar(80);not null;comment:更新管理员用户名;"`             // 更新管理员用户名
	UpdatedByNickname   *string        `json:"-" gorm:"column:updated_by_nickname;type:varchar(80);default:null;comment:更新管理员昵称;"`          // 更新管理员昵称
	LastChangeReason    *string        `json:"-" gorm:"column:last_change_reason;type:varchar(200);default:null;comment:最近变更原因;"`           // 最近变更原因
	CreatedAt           time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                     // 创建时间
	UpdatedAt           time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`               // 更新时间
}

// TableName 指定AIProvider对应的数据表名。
func (AIProvider) TableName() string {
	return "of_ai_providers"
}

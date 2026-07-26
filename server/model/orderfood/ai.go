package orderfood

import (
	"time"

	"gorm.io/datatypes"
)

// AIProviderType 表示平台支持的AI供应商类型。
type AIProviderType string

const (
	AIProviderBailian  AIProviderType = "bailian"
	AIProviderDeepSeek AIProviderType = "deepseek"
	AIProviderOpenAI   AIProviderType = "openai"
)

// AIModelCapability 表示模型支持的输入输出模态。
type AIModelCapability string

const (
	AIModelText            AIModelCapability = "text"
	AIModelVision          AIModelCapability = "vision"
	AIModelImageGeneration AIModelCapability = "image_generation"
)

// AIPromptMode 表示能力使用平台默认或自定义提示词。
type AIPromptMode string

const (
	AIPromptPreset AIPromptMode = "preset"
	AIPromptCustom AIPromptMode = "custom"
)

const (
	AICapabilityDishTextExtract     = "dish_text_extract"
	AICapabilityRecipeImageExtract  = "recipe_image_extract"
	AICapabilityDishCoverCreate     = "dish_cover_create"
	AICapabilityCheckinImageAnalyze = "checkin_image_analyze"
	AICapabilityMealSuggest         = "meal_suggest"
	AICapabilityPrepSequence        = "prep_sequence"
)

// PlatformCapabilityPolicy 表示当前生效的平台AI整体策略。
type PlatformCapabilityPolicy struct {
	SingletonKey           string         `json:"-" gorm:"column:singleton_key;type:varchar(32);primaryKey;not null;comment:单例记录键;"`                                // 单例记录键
	Version                int64          `json:"policyVersion" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                 // 数据版本
	PlatformDefaultEnabled bool           `json:"platformDefaultEnabled" gorm:"column:platform_default_enabled;type:tinyint(1) unsigned;not null;comment:平台AI总开关;"` // 平台AI总开关
	EmergencyDisabled      bool           `json:"emergencyDisabled" gorm:"column:emergency_disabled;type:tinyint(1) unsigned;not null;comment:是否紧急停用;"`             // 是否紧急停用
	FeatureLabelsJSON      datatypes.JSON `json:"-" gorm:"column:feature_labels_json;type:json;not null;comment:功能文案JSON;"`                                         // 功能文案JSON
	AppliedByID            uint           `json:"-" gorm:"column:applied_by_id;type:bigint unsigned;not null;comment:应用管理员ID;"`                                     // 应用管理员ID
	AppliedByUsername      string         `json:"-" gorm:"column:applied_by_username;type:varchar(80);not null;comment:应用管理员用户名;"`                                  // 应用管理员用户名
	AppliedByNickname      *string        `json:"-" gorm:"column:applied_by_nickname;type:varchar(80);default:null;comment:应用管理员昵称;"`                               // 应用管理员昵称
	AppliedAt              time.Time      `json:"appliedAt" gorm:"column:applied_at;type:datetime;not null;index;comment:应用时间;"`                                    // 应用时间
	Reason                 string         `json:"reason" gorm:"column:reason;type:varchar(200);not null;comment:原因;"`                                               // 原因
}

// TableName 指定PlatformCapabilityPolicy对应的数据表名。
func (PlatformCapabilityPolicy) TableName() string {
	return "of_ai_policy"
}

// AIProvider stores only a credential reference. It is deliberately excluded
// from JSON and no response DTO has a field capable of returning it.
type AIProvider struct {
	ID                  string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                      // 主键ID
	Name                string         `json:"name" gorm:"column:name;type:varchar(60);not null;comment:名称;"`                               // 名称
	Type                AIProviderType `json:"type" gorm:"column:type;type:varchar(24);not null;index;comment:类型;"`                         // 类型
	BaseURL             string         `json:"baseUrl" gorm:"column:base_url;type:varchar(300);not null;comment:供应商基础地址;"`                  // 供应商基础地址
	CredentialRef       string         `json:"-" gorm:"column:credential_ref;type:text;not null;comment:密钥引用;"`                             // 密钥引用
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

// AIModel 表示AI模型配置。
type AIModel struct {
	ID                             string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                                  // 主键ID
	ProviderID                     string         `json:"providerId" gorm:"column:provider_id;type:varchar(64);not null;uniqueIndex:uk_of_ai_model_key,priority:1;index;comment:供应商ID;"`           // 供应商ID
	Name                           string         `json:"name" gorm:"column:name;type:varchar(60);not null;comment:名称;"`                                                                           // 名称
	ModelKey                       string         `json:"modelKey" gorm:"column:model_key;type:varchar(120);not null;uniqueIndex:uk_of_ai_model_key,priority:2;comment:供应商模型标识;"`                  // 供应商模型标识
	CapabilitiesJSON               datatypes.JSON `json:"-" gorm:"column:capabilities_json;type:json;not null;comment:模型能力集合JSON;"`                                                                // 模型能力集合JSON
	ContextLength                  *int           `json:"contextLength" gorm:"column:context_length;type:int;default:null;comment:上下文长度;"`                                                         // 上下文长度
	InputPricePerMillionTokensCNY  *string        `json:"inputPricePerMillionTokensCny" gorm:"column:input_price_per_million_tokens_cny;type:varchar(40);default:null;comment:每百万输入Token价格（元）;"`   // 每百万输入Token价格（元）
	OutputPricePerMillionTokensCNY *string        `json:"outputPricePerMillionTokensCny" gorm:"column:output_price_per_million_tokens_cny;type:varchar(40);default:null;comment:每百万输出Token价格（元）;"` // 每百万输出Token价格（元）
	ImagePricePerUnitCNY           *string        `json:"imagePricePerUnitCny" gorm:"column:image_price_per_unit_cny;type:varchar(40);default:null;comment:单张图片价格（元）;"`                            // 单张图片价格（元）
	Enabled                        bool           `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;index;comment:是否启用;"`                                                     // 是否启用
	Remark                         *string        `json:"remark" gorm:"column:remark;type:varchar(300);default:null;comment:备注;"`                                                                  // 备注
	Version                        int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                              // 数据版本
	UpdatedByID                    uint           `json:"-" gorm:"column:updated_by_id;type:bigint unsigned;not null;comment:更新管理员ID;"`                                                            // 更新管理员ID
	UpdatedByUsername              string         `json:"-" gorm:"column:updated_by_username;type:varchar(80);not null;comment:更新管理员用户名;"`                                                         // 更新管理员用户名
	UpdatedByNickname              *string        `json:"-" gorm:"column:updated_by_nickname;type:varchar(80);default:null;comment:更新管理员昵称;"`                                                      // 更新管理员昵称
	LastChangeReason               *string        `json:"-" gorm:"column:last_change_reason;type:varchar(200);default:null;comment:最近变更原因;"`                                                       // 最近变更原因
	CreatedAt                      time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                 // 创建时间
	UpdatedAt                      time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                                           // 更新时间
}

// TableName 指定AIModel对应的数据表名。
func (AIModel) TableName() string {
	return "of_ai_models"
}

// AICapabilityDefinition contains stable product metadata. Client-facing
// labels remain in the platform policy and are not duplicated here.
type AICapabilityDefinition struct {
	Code                    string            `json:"code" gorm:"column:code;type:varchar(64);primaryKey;not null;comment:AI能力编码;"`                              // AI能力编码
	Name                    string            `json:"name" gorm:"column:name;type:varchar(80);not null;comment:名称;"`                                             // 名称
	ClientFeatureCode       string            `json:"clientFeatureCode" gorm:"column:client_feature_code;type:varchar(40);not null;index;comment:小程序功能编码;"`      // 小程序功能编码
	RequiredModelCapability AIModelCapability `json:"requiredModelCapability" gorm:"column:required_model_capability;type:varchar(32);not null;comment:所需模型能力;"` // 所需模型能力
	SortOrder               int               `json:"sortOrder" gorm:"column:sort_order;type:int;not null;index;comment:排序值;"`                                   // 排序值
	CreatedAt               time.Time         `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                           // 创建时间
	UpdatedAt               time.Time         `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                   // 更新时间
}

// TableName 指定AICapabilityDefinition对应的数据表名。
func (AICapabilityDefinition) TableName() string {
	return "of_ai_capabilities"
}

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

// AICapabilityConfig 表示当前生效的AI能力配置。
type AICapabilityConfig struct {
	CapabilityCode              string         `json:"capabilityCode" gorm:"column:capability_code;type:varchar(64);primaryKey;not null;comment:AI能力编码;"`                 // AI能力编码
	Version                     int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                        // 数据版本
	PrimaryModelID              string         `json:"primaryModelId" gorm:"column:primary_model_id;type:varchar(64);not null;index;comment:主模型ID;"`                      // 主模型ID
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

// AIConfigurationChange 表示AI配置变更记录。
type AIConfigurationChange struct {
	ID                    string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                      // 主键ID
	ResourceType          string    `json:"resourceType" gorm:"column:resource_type;type:varchar(40);not null;index:idx_of_ai_change_resource,priority:1;comment:资源类型;"` // 资源类型
	ResourceID            string    `json:"resourceId" gorm:"column:resource_id;type:varchar(80);not null;index:idx_of_ai_change_resource,priority:2;comment:资源ID;"`     // 资源ID
	Action                string    `json:"action" gorm:"column:action;type:varchar(40);not null;comment:操作类型;"`                                                         // 操作类型
	ResourceVersion       int64     `json:"resourceVersion" gorm:"column:resource_version;type:bigint;not null;comment:资源版本;"`                                           // 资源版本
	AdministratorID       uint      `json:"-" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`                                         // 管理员ID
	AdministratorUsername string    `json:"-" gorm:"column:administrator_username;type:varchar(80);not null;comment:管理员用户名;"`                                            // 管理员用户名
	Reason                *string   `json:"reason" gorm:"column:reason;type:varchar(200);default:null;comment:原因;"`                                                      // 原因
	CreatedAt             time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                               // 创建时间
}

// TableName 指定AIConfigurationChange对应的数据表名。
func (AIConfigurationChange) TableName() string {
	return "of_ai_changes"
}

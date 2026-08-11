package request

import aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"

// AIAdminActor 表示 AI 管理操作的管理员上下文。
type AIAdminActor struct {
	AdministratorID  uint    // 管理员ID
	AuthorityID      uint    // 角色ID
	Username         string  // 用户名
	Nickname         *string // 昵称
	RequestID        string  // 请求ID
	SourceIPMasked   string  // 脱敏后的来源IP
	UserAgentSummary string  // 用户代理摘要
}

// AIEmptyQuery 表示AI空查询条件。
type AIEmptyQuery struct{}

// ClientFeatureLabelInput 表示客户端功能显示名称输入参数。
type ClientFeatureLabelInput struct {
	Code        string `json:"code" binding:"required,oneof=dish_extract cover_create meal_suggest prep_sequence"` // 编码
	Title       string `json:"title" binding:"required,max=40"`                                                    // 标题
	ActionLabel string `json:"actionLabel" binding:"required,max=20"`                                              // 操作按钮文案
	Description string `json:"description" binding:"required,max=160"`                                             // 说明
	SortOrder   int    `json:"sortOrder" binding:"required,min=1,max=4"`                                           // 排序值
}

// PlatformPolicyUpdateInput 表示平台策略直接保存参数。
type PlatformPolicyUpdateInput struct {
	PlatformDefaultEnabled bool                      `json:"platformDefaultEnabled"`                                   // 平台增强能力总开关
	FeatureLabels          []ClientFeatureLabelInput `json:"featureLabels" binding:"required,len=4,dive"`              // 客户端增强功能动态文案
	Reason                 string                    `json:"reason" binding:"required,min=2,max=200" checksql:"false"` // 修改原因
	ExpectedVersion        int64                     `json:"expectedVersion" binding:"required,min=1"`                 // 当前配置版本
}

// PlatformPolicyEmergencyInput 表示平台策略紧急输入参数。
type PlatformPolicyEmergencyInput struct {
	EmergencyDisabled *bool  `json:"emergencyDisabled" binding:"required"`     // 紧急停用
	Reason            string `json:"reason" binding:"required,min=4,max=200"`  // 原因
	ExpectedVersion   int64  `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// AIProviderPath 表示AI供应商路径参数。
type AIProviderPath struct {
	ProviderID string `uri:"providerId" json:"providerId" binding:"required,max=64"` // 供应商ID
}

// AIProviderListQuery 表示AI供应商列表查询条件。
type AIProviderListQuery struct {
	Page      int                    `form:"page" json:"page" binding:"omitempty,min=1"`                              // 页码
	PageSize  int                    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`            // 每页数量
	Keyword   string                 `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                       // 关键词
	Type      aiModel.AIProviderType `form:"type" json:"type" binding:"omitempty,oneof=bailian deepseek openai"`      // 类型
	Enabled   *bool                  `form:"enabled" json:"enabled" binding:"omitempty"`                              // 是否启用
	SortBy    string                 `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt updatedAt name"` // 排序字段
	SortOrder string                 `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`           // 排序方向
}

// ApplyDefaults 补齐分页、筛选和排序默认值。
func (query *AIProviderListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "updatedAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// AIProviderCreateInput 表示AI供应商创建输入参数。
type AIProviderCreateInput struct {
	Name      string                 `json:"name" binding:"required,max=60"`                          // 名称
	Type      aiModel.AIProviderType `json:"type" binding:"required,oneof=bailian deepseek openai"`   // 类型
	BaseURL   string                 `json:"baseUrl" binding:"required,url,max=300" checksql:"false"` // 基础地址
	APIKey    string                 `json:"apiKey" binding:"required,max=500" checksql:"false"`      // 供应商 API Key（数据库明文保存，响应不回显）
	TimeoutMS int                    `json:"timeoutMs" binding:"required,min=1000,max=120000"`        // 超时时间（毫秒）
	Remark    *string                `json:"remark" binding:"omitempty,max=300"`                      // 备注
}

// AIProviderUpdateInput 表示AI供应商更新输入参数。
type AIProviderUpdateInput struct {
	Name            string  `json:"name" binding:"required,max=60"`                          // 名称
	BaseURL         string  `json:"baseUrl" binding:"required,url,max=300" checksql:"false"` // 基础地址
	APIKey          *string `json:"apiKey" binding:"omitempty,max=500" checksql:"false"`     // 新 API Key；未传表示保留当前值
	TimeoutMS       int     `json:"timeoutMs" binding:"required,min=1000,max=120000"`        // 超时时间（毫秒）
	Remark          *string `json:"remark" binding:"omitempty,max=300"`                      // 备注
	ExpectedVersion int64   `json:"expectedVersion" binding:"required,min=1"`                // 预期版本
}

// AIStatusUpdateInput 表示AI状态更新输入参数。
type AIStatusUpdateInput struct {
	Enabled         *bool  `json:"enabled" binding:"required"`               // 是否启用
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// AIDeleteInput 表示AI删除输入参数。
type AIDeleteInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// AIProviderConnectionTestInput 表示AI供应商连接测试输入参数。
type AIProviderConnectionTestInput struct {
	TimeoutMS       *int  `json:"timeoutMs" binding:"omitempty,min=1000,max=30000"` // 超时时间（毫秒）
	ExpectedVersion int64 `json:"expectedVersion" binding:"required,min=1"`         // 预期版本
}

// AIModelPath 表示AI模型路径参数。
type AIModelPath struct {
	ModelID string `uri:"modelId" json:"modelId" binding:"required,max=64"` // 模型ID
}

// AIModelListQuery 表示AI模型列表查询条件。
type AIModelListQuery struct {
	Page       int                       `form:"page" json:"page" binding:"omitempty,min=1"`                                          // 页码
	PageSize   int                       `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                        // 每页数量
	Keyword    string                    `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                                   // 关键词
	ProviderID string                    `form:"providerId" json:"providerId" binding:"omitempty,max=64"`                             // 供应商ID
	Capability aiModel.AIModelCapability `form:"capability" json:"capability" binding:"omitempty,oneof=text vision image_generation"` // 能力
	Enabled    *bool                     `form:"enabled" json:"enabled" binding:"omitempty"`                                          // 是否启用
	SortBy     string                    `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt updatedAt name"`             // 排序字段
	SortOrder  string                    `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                       // 排序方向
}

// ApplyDefaults 补齐分页、筛选和排序默认值。
func (query *AIModelListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "updatedAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// AIModelCreateInput 表示AI模型创建输入参数。
type AIModelCreateInput struct {
	ProviderID                     string                      `json:"providerId" binding:"required,max=64"`                                                 // 供应商ID
	Name                           string                      `json:"name" binding:"required,max=60"`                                                       // 名称
	ModelKey                       string                      `json:"modelKey" binding:"required,max=120" checksql:"false"`                                 // 模型键
	Capabilities                   []aiModel.AIModelCapability `json:"capabilities" binding:"required,min=1,unique,dive,oneof=text vision image_generation"` // 能力列表
	ContextLength                  *int                        `json:"contextLength" binding:"omitempty,min=1"`                                              // 上下文长度
	InputPricePerMillionTokensCNY  *string                     `json:"inputPricePerMillionTokensCny" binding:"omitempty,max=40"`                             // 每百万输入 Token 成本（人民币）
	OutputPricePerMillionTokensCNY *string                     `json:"outputPricePerMillionTokensCny" binding:"omitempty,max=40"`                            // 每百万输出 Token 成本（人民币）
	ImagePricePerUnitCNY           *string                     `json:"imagePricePerUnitCny" binding:"omitempty,max=40"`                                      // 单张图片调用成本（人民币）
	Remark                         *string                     `json:"remark" binding:"omitempty,max=300"`                                                   // 备注
}

// AIModelUpdateInput 表示AI模型更新输入参数。
type AIModelUpdateInput struct {
	ProviderID                     string                      `json:"providerId" binding:"required,max=64"`                                                 // 供应商ID
	Name                           string                      `json:"name" binding:"required,max=60"`                                                       // 名称
	ModelKey                       string                      `json:"modelKey" binding:"required,max=120" checksql:"false"`                                 // 模型键
	Capabilities                   []aiModel.AIModelCapability `json:"capabilities" binding:"required,min=1,unique,dive,oneof=text vision image_generation"` // 能力列表
	ContextLength                  *int                        `json:"contextLength" binding:"omitempty,min=1"`                                              // 上下文长度
	InputPricePerMillionTokensCNY  *string                     `json:"inputPricePerMillionTokensCny" binding:"omitempty,max=40"`                             // 每百万输入 Token 成本（人民币）
	OutputPricePerMillionTokensCNY *string                     `json:"outputPricePerMillionTokensCny" binding:"omitempty,max=40"`                            // 每百万输出 Token 成本（人民币）
	ImagePricePerUnitCNY           *string                     `json:"imagePricePerUnitCny" binding:"omitempty,max=40"`                                      // 单张图片调用成本（人民币）
	Remark                         *string                     `json:"remark" binding:"omitempty,max=300"`                                                   // 备注
	ExpectedVersion                int64                       `json:"expectedVersion" binding:"required,min=1"`                                             // 预期版本
}

// AICapabilityPath 表示AI能力路径参数。
type AICapabilityPath struct {
	CapabilityCode string `uri:"capabilityCode" json:"capabilityCode" binding:"required,oneof=dish_text_extract recipe_image_extract dish_cover_create checkin_image_analyze preference_profile_summarize meal_suggest prep_sequence"` // 能力编码
}

// AICapabilityListQuery 表示AI能力列表查询条件。
type AICapabilityListQuery struct {
	ClientFeatureCode string `form:"clientFeatureCode" json:"clientFeatureCode" binding:"omitempty,oneof=dish_extract cover_create meal_suggest prep_sequence checkin_image_analyze"` // 客户端功能编码
	SortBy            string `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=sortOrder code"`                                                                                   // 排序字段
	SortOrder         string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                                                                   // 排序方向
}

// ApplyDefaults 补齐分页、筛选和排序默认值。
func (query *AICapabilityListQuery) ApplyDefaults() {
	if query.SortBy == "" {
		query.SortBy = "sortOrder"
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
}

// AIPromptConfigInput 表示AI提示词配置输入参数。
type AIPromptConfigInput struct {
	Mode               aiModel.AIPromptMode `json:"mode" binding:"required,oneof=default custom"`                      // 模式，default使用当前平台默认提示词，custom使用自定义提示词
	SystemPrompt       *string              `json:"systemPrompt" binding:"omitempty,max=12000" checksql:"false"`       // 系统提示词
	UserPromptTemplate *string              `json:"userPromptTemplate" binding:"omitempty,max=12000" checksql:"false"` // 用户提示词模板
}

// AIPromptUpdateInput 表示独立保存并立即应用提示词的输入参数。
type AIPromptUpdateInput struct {
	Prompt          AIPromptConfigInput `json:"prompt" binding:"required"`                                // 提示词配置
	Reason          string              `json:"reason" binding:"required,min=2,max=200" checksql:"false"` // 修改原因
	ExpectedVersion int64               `json:"expectedVersion" binding:"required,min=1"`                 // 当前能力配置版本
}

// AICapabilityUpdateInput 表示AI能力配置直接保存参数。
type AICapabilityUpdateInput struct {
	PrimaryModelID    string  `json:"primaryModelId" binding:"required,max=64"`                 // 主模型ID
	AuxiliaryModelID  *string `json:"auxiliaryModelId" binding:"omitempty,max=64"`              // 辅助模型ID，仅双模型能力必填
	PointCost         int     `json:"pointCost" binding:"min=0,max=100000"`                     // 积分成本
	DailyLimitPerUser int     `json:"dailyLimitPerUser" binding:"min=0,max=100000"`             // 每日限额每用户
	TimeoutMS         int     `json:"timeoutMs" binding:"required,min=1000,max=120000"`         // 超时时间（毫秒）
	FreeQuotaPerDay   int     `json:"freeQuotaPerDay" binding:"min=0,max=100000"`               // 免费额度每天
	Reason            string  `json:"reason" binding:"required,min=2,max=200" checksql:"false"` // 修改原因
	ExpectedVersion   int64   `json:"expectedVersion" binding:"min=0"`                          // 当前能力配置版本，尚未配置时为0
}

// AIPromptValidationInput 表示AI提示词校验输入参数。
type AIPromptValidationInput struct {
	Prompt AIPromptConfigInput `json:"prompt" binding:"required"` // 提示词
}

// AIPromptRenderPreviewInput 表示AI提示词变量替换预览输入参数。
type AIPromptRenderPreviewInput struct {
	Prompt    AIPromptConfigInput `json:"prompt" binding:"required"`                     // 待预览的提示词
	Variables map[string]string   `json:"variables" binding:"required" checksql:"false"` // 变量
}

// AIPromptTestInput 表示AI提示词测试输入参数。
type AIPromptTestInput struct {
	Prompt    AIPromptConfigInput `json:"prompt" binding:"required"`                     // 待测试的提示词配置
	Variables map[string]string   `json:"variables" binding:"required" checksql:"false"` // 变量
}

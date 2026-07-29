package response

import "time"

// ClientFeatureLabel 表示客户端功能显示名称响应数据。
type ClientFeatureLabel struct {
	Code          string  `json:"code"`          // 编码
	Title         string  `json:"title"`         // 标题
	ActionLabel   string  `json:"actionLabel"`   // 操作按钮文案
	Description   string  `json:"description"`   // 说明
	CostHint      *string `json:"costHint"`      // 根据当前能力配置实时计算的成本提示
	FreeQuotaHint *string `json:"freeQuotaHint"` // 根据当前能力配置实时计算的免费额度提示
	SortOrder     int     `json:"sortOrder"`     // 排序值
}

// AICapabilityReadiness 表示AI能力就绪状态响应数据。
type AICapabilityReadiness struct {
	TotalCount          int      `json:"totalCount"`          // 总数量
	ReadyCount          int      `json:"readyCount"`          // 已就绪能力数量
	AllReady            bool     `json:"allReady"`            // 全部能力是否就绪
	UnreadyCapabilities []string `json:"unreadyCapabilities"` // 未就绪能力列表
}

// PlatformCapabilitySnapshot 表示平台能力快照响应数据。
type PlatformCapabilitySnapshot struct {
	PlatformDefaultEnabled        bool  `json:"platformDefaultEnabled"`        // 平台增强能力总开关
	EmergencyDisabled             bool  `json:"emergencyDisabled"`             // 紧急停用
	NormalUserCount               int64 `json:"normalUserCount"`               // 正常用户数量
	EffectiveEnabledUserCount     int64 `json:"effectiveEnabledUserCount"`     // 能力实际启用用户数
	IndividuallyDisabledUserCount int64 `json:"individuallyDisabledUserCount"` // 被单独关闭能力的用户数
	PolicyVersion                 int64 `json:"policyVersion"`                 // 策略版本
}

// PlatformPolicyConfig 表示当前生效的平台策略。
type PlatformPolicyConfig struct {
	PlatformDefaultEnabled bool                 `json:"platformDefaultEnabled"` // 平台增强能力总开关
	EmergencyDisabled      bool                 `json:"emergencyDisabled"`      // 紧急停用
	FeatureLabels          []ClientFeatureLabel `json:"featureLabels"`          // 客户端增强功能动态文案
	PolicyVersion          int64                `json:"policyVersion"`          // 策略版本
	UpdatedBy              AdministratorSummary `json:"updatedBy"`              // 更新人
	UpdatedAt              time.Time            `json:"updatedAt"`              // 更新时间
	Reason                 string               `json:"reason"`                 // 原因
}

// PlatformPolicyWorkspace 表示平台策略配置响应数据。
type PlatformPolicyWorkspace struct {
	Config     PlatformPolicyConfig       `json:"config"`     // 当前生效配置
	UserCounts PlatformCapabilitySnapshot `json:"userCounts"` // 用户数量统计
	Readiness  AICapabilityReadiness      `json:"readiness"`  // 就绪状态
}

// ProviderConnectionTestResult 表示供应商连接测试结果响应数据。
type ProviderConnectionTestResult struct {
	Success     bool      `json:"success"`     // 是否成功
	Category    string    `json:"category"`    // 分类
	DurationMS  int       `json:"durationMs"`  // 耗时（毫秒）
	TestedAt    time.Time `json:"testedAt"`    // 测试时间
	SafeMessage string    `json:"safeMessage"` // 安全消息
}

// AIProviderModelOption 表示供应商实时返回的可选模型。
type AIProviderModelOption struct {
	Name       string `json:"name"`       // 显示名称
	ModelKey   string `json:"modelKey"`   // 供应商模型标识
	Configured bool   `json:"configured"` // 是否已在当前供应商下配置
}

// AIModelProviderOption 表示模型表单可选择的供应商基础信息。
type AIModelProviderOption struct {
	ID      string `json:"id"`      // 供应商ID
	Name    string `json:"name"`    // 供应商名称
	Type    string `json:"type"`    // 供应商类型
	Enabled bool   `json:"enabled"` // 是否启用
}

// AIProviderSummary 表示AI供应商摘要响应数据。
type AIProviderSummary struct {
	ID                     string                        `json:"id"`                     // ID
	Name                   string                        `json:"name"`                   // 名称
	Type                   string                        `json:"type"`                   // 类型
	BaseURL                string                        `json:"baseUrl"`                // 基础地址
	CredentialConfigured   bool                          `json:"credentialConfigured"`   // 供应商凭证是否已配置
	CredentialMaskedLabel  *string                       `json:"credentialMaskedLabel"`  // 供应商凭证脱敏标识
	TimeoutMS              int                           `json:"timeoutMs"`              // 超时时间（毫秒）
	Enabled                bool                          `json:"enabled"`                // 是否启用
	ModelCount             int64                         `json:"modelCount"`             // 模型数量
	EnabledModelCount      int64                         `json:"enabledModelCount"`      // 已启用模型数量
	EnabledCapabilityCount int64                         `json:"enabledCapabilityCount"` // 当前启用能力引用数量
	UsageCount             int64                         `json:"usageCount"`             // 历史调用引用数量
	LastConnectionTest     *ProviderConnectionTestResult `json:"lastConnectionTest"`     // 最近连接测试
	Version                int64                         `json:"version"`                // 版本
	UpdatedAt              time.Time                     `json:"updatedAt"`              // 更新时间
}

// AIProviderModelReference 表示供应商详情中的模型引用。
type AIProviderModelReference struct {
	ID       string `json:"id"`       // 模型ID
	Name     string `json:"name"`     // 模型名称
	ModelKey string `json:"modelKey"` // 供应商模型标识
	Enabled  bool   `json:"enabled"`  // 是否启用
}

// AIProviderCapabilityReference 表示供应商详情中的能力引用。
type AIProviderCapabilityReference struct {
	CapabilityCode string `json:"capabilityCode"` // 能力编码
	CapabilityName string `json:"capabilityName"` // 能力名称
	ModelID        string `json:"modelId"`        // 主模型ID
	ModelName      string `json:"modelName"`      // 主模型名称
}

// AIProviderDetail 表示AI供应商详情响应数据。
type AIProviderDetail struct {
	AIProviderSummary                                      // AI 供应商摘要
	ReferencedModels       []AIProviderModelReference      `json:"referencedModels"`       // 引用模型
	ReferencedCapabilities []AIProviderCapabilityReference `json:"referencedCapabilities"` // 引用能力
	Remark                 *string                         `json:"remark"`                 // 备注
	CreatedAt              time.Time                       `json:"createdAt"`              // 创建时间
}

// AIModelSummary 表示AI模型摘要响应数据。
type AIModelSummary struct {
	ID                             string            `json:"id"`                             // ID
	Provider                       AIProviderSummary `json:"provider"`                       // 供应商
	Name                           string            `json:"name"`                           // 名称
	ModelKey                       string            `json:"modelKey"`                       // 模型键
	Capabilities                   []string          `json:"capabilities"`                   // 能力列表
	ContextLength                  *int              `json:"contextLength"`                  // 上下文长度
	InputPricePerMillionTokensCNY  *string           `json:"inputPricePerMillionTokensCny"`  // 每百万输入 Token 成本（人民币）
	OutputPricePerMillionTokensCNY *string           `json:"outputPricePerMillionTokensCny"` // 每百万输出 Token 成本（人民币）
	ImagePricePerUnitCNY           *string           `json:"imagePricePerUnitCny"`           // 单张图片调用成本（人民币）
	Enabled                        bool              `json:"enabled"`                        // 是否启用
	EnabledCapabilityCount         int64             `json:"enabledCapabilityCount"`         // 当前启用能力引用数量
	UsageCount                     int64             `json:"usageCount"`                     // 历史调用引用数量
	ReferenceCount                 int64             `json:"referenceCount"`                 // 能力配置和调用记录引用总数
	Version                        int64             `json:"version"`                        // 版本
	UpdatedAt                      time.Time         `json:"updatedAt"`                      // 更新时间
}

// AIModelCapabilityReference 表示模型详情中的能力引用。
type AIModelCapabilityReference struct {
	CapabilityCode string `json:"capabilityCode"` // 能力编码
	CapabilityName string `json:"capabilityName"` // 能力名称
}

// AIModelDetail 表示AI模型详情响应数据。
type AIModelDetail struct {
	AIModelSummary                                      // AI 模型摘要
	ReferencedCapabilities []AIModelCapabilityReference `json:"referencedCapabilities"` // 引用该模型的能力
	Currency               string                       `json:"currency"`               // 币种
	Remark                 *string                      `json:"remark"`                 // 备注
	CreatedAt              time.Time                    `json:"createdAt"`              // 创建时间
}

// AIPromptMetadata is safe for every capability reader. Prompt bodies are
// exposed only through pointer fields after the prompt-read permission and
// synchronous access audit have both succeeded.
type AIPromptMetadata struct {
	Mode                  string   `json:"mode"`                         // 模式
	PresetVersion         int64    `json:"presetVersion"`                // 预设版本
	LatestPresetVersion   int64    `json:"latestPresetVersion"`          // 最新平台预设提示词版本
	PresetUpdateAvailable bool     `json:"presetUpdateAvailable"`        // 是否存在新版平台预设
	ContentHash           string   `json:"contentHash"`                  // 内容摘要
	AllowedVariables      []string `json:"allowedVariables"`             // 允许使用的提示词变量
	RequiredVariables     []string `json:"requiredVariables"`            // 必填变量列表
	OutputSchemaVersion   string   `json:"outputSchemaVersion"`          // 输出结构版本
	BodyReadable          bool     `json:"bodyReadable"`                 // 供应商响应正文是否可读取
	SystemPrompt          *string  `json:"systemPrompt,omitempty"`       // 系统提示词
	UserPromptTemplate    *string  `json:"userPromptTemplate,omitempty"` // 用户提示词模板
}

// AICapabilityConfig 表示AI能力配置响应数据。
type AICapabilityConfig struct {
	CapabilityCode       string               `json:"capabilityCode"`       // 能力编码
	PrimaryModelID       string               `json:"primaryModelId"`       // 主模型ID
	AuxiliaryModelID     *string              `json:"auxiliaryModelId"`     // 辅助模型ID
	PointCost            int                  `json:"pointCost"`            // 积分成本
	DailyLimitPerUser    int                  `json:"dailyLimitPerUser"`    // 每日限额每用户
	TimeoutMS            int                  `json:"timeoutMs"`            // 超时时间（毫秒）
	FreeQuotaPerDay      int                  `json:"freeQuotaPerDay"`      // 免费额度每天
	FixedValidationRules []string             `json:"fixedValidationRules"` // 服务端固定且不可编辑的结果校验说明
	Prompt               AIPromptMetadata     `json:"-"`                    // 内部提示词元数据，外部通过独立提示词接口读取
	PromptMode           string               `json:"promptMode"`           // 提示词模式
	PromptHash           string               `json:"promptHash"`           // 提示词摘要
	Version              int64                `json:"version"`              // 当前配置版本
	UpdatedBy            AdministratorSummary `json:"updatedBy"`            // 更新人
	UpdatedAt            time.Time            `json:"updatedAt"`            // 更新时间
	Reason               string               `json:"reason"`               // 修改原因
}

// AICapabilitySummary 表示AI能力摘要响应数据。
type AICapabilitySummary struct {
	Code                             string    `json:"code"`                             // 编码
	Name                             string    `json:"name"`                             // 名称
	ClientFeatureCode                string    `json:"clientFeatureCode"`                // 客户端功能编码
	Ready                            bool      `json:"ready"`                            // 是否就绪
	RequiredModelCapability          string    `json:"requiredModelCapability"`          // 必需模型能力
	RequiredAuxiliaryModelCapability *string   `json:"requiredAuxiliaryModelCapability"` // 辅助模型必需能力
	PrimaryModelName                 string    `json:"primaryModelName"`                 // 主模型名称
	AuxiliaryModelName               string    `json:"auxiliaryModelName"`               // 辅助模型名称
	PointCost                        int       `json:"pointCost"`                        // 积分成本
	DailyLimitPerUser                int       `json:"dailyLimitPerUser"`                // 每日限额每用户
	FreeQuotaPerDay                  int       `json:"freeQuotaPerDay"`                  // 免费额度每天
	Version                          int64     `json:"version"`                          // 当前配置版本
	PromptMode                       string    `json:"promptMode"`                       // 提示词模式
	PromptPresetVersion              int64     `json:"-"`                                // 内部平台预设提示词版本
	PromptHash                       string    `json:"promptHash"`                       // 提示词摘要
	SortOrder                        int       `json:"sortOrder"`                        // 排序值
	UpdatedAt                        time.Time `json:"updatedAt"`                        // 更新时间
}

// AICapabilityListResult 表示 AI 能力列表响应数据。
type AICapabilityListResult struct {
	List []AICapabilitySummary `json:"list"` // AI 能力列表
}

// AICapabilityWorkspace 表示AI能力当前配置响应数据。
type AICapabilityWorkspace struct {
	Summary AICapabilitySummary `json:"summary"` // 摘要
	Config  *AICapabilityConfig `json:"config"`  // 当前配置，尚未配置时为空
}

// AIPromptPreset 表示AI提示词预设响应数据。
type AIPromptPreset struct {
	CapabilityCode      string    `json:"capabilityCode"`      // 能力编码
	Version             int64     `json:"version"`             // 版本
	ContentHash         string    `json:"contentHash"`         // 内容摘要
	AllowedVariables    []string  `json:"allowedVariables"`    // 允许使用的提示词变量
	RequiredVariables   []string  `json:"requiredVariables"`   // 必填变量列表
	OutputSchemaVersion string    `json:"outputSchemaVersion"` // 输出结构版本
	SystemPrompt        string    `json:"systemPrompt"`        // 系统提示词
	UserPromptTemplate  string    `json:"userPromptTemplate"`  // 用户提示词模板
	CreatedAt           time.Time `json:"createdAt"`           // 创建时间
}

// AIPromptVariable 表示AI提示词可用变量响应数据。
type AIPromptVariable struct {
	Name        string `json:"name"`        // 变量名
	Label       string `json:"label"`       // 展示名称
	Description string `json:"description"` // 变量说明
	Required    bool   `json:"required"`    // 是否必填
	Example     string `json:"example"`     // 示例值
}

// AIPromptWorkspace 表示AI提示词工作区响应数据。
type AIPromptWorkspace struct {
	CapabilityCode            string                 `json:"capabilityCode"`            // 能力编码
	Current                   *AIPromptConfig        `json:"current"`                   // 当前提示词配置，能力未配置时为空
	DefaultSystemPrompt       string                 `json:"defaultSystemPrompt"`       // 平台默认系统提示词
	DefaultUserPromptTemplate string                 `json:"defaultUserPromptTemplate"` // 平台默认用户提示词模板
	AllowedVariables          []AIPromptVariable     `json:"allowedVariables"`          // 允许使用的变量
	FixedOutputSchema         map[string]interface{} `json:"fixedOutputSchema"`         // 服务端固定输出结构
	EffectivePromptHash       string                 `json:"effectivePromptHash"`       // 当前有效提示词摘要
	LastSuccessfulTest        *AIPromptTestResult    `json:"lastSuccessfulTest"`        // 最近成功测试
	AccessAuditID             string                 `json:"-"`                         // 内部敏感访问审计记录ID
}

// AIPromptValidationResult 表示AI提示词校验结果响应数据。
type AIPromptValidationResult struct {
	Valid                    bool     `json:"valid"`                    // 是否有效
	ContentHash              string   `json:"-"`                        // 内部内容摘要
	AllowedVariables         []string `json:"-"`                        // 内部允许使用的提示词变量
	RequiredVariables        []string `json:"-"`                        // 内部必填变量列表
	ReferencedVariables      []string `json:"-"`                        // 内部被引用变量
	UnknownVariables         []string `json:"unknownVariables"`         // 未在白名单中的提示词变量
	MissingRequiredVariables []string `json:"missingRequiredVariables"` // 缺失的必填提示词变量
	Errors                   []string `json:"errors"`                   // 校验错误列表
	PromptHash               *string  `json:"promptHash"`               // 校验通过后的提示词摘要
	Warnings                 []string `json:"-"`                        // 内部兼容的警告列表
}

// AIPromptRenderPreviewResult 表示AI提示词变量替换预览响应数据。
type AIPromptRenderPreviewResult struct {
	CapabilityCode       string `json:"-"`            // 内部能力编码
	ConfigVersion        int64  `json:"-"`            // 内部配置版本
	PromptHash           string `json:"promptHash"`   // 提示词摘要
	RenderedSystemPrompt string `json:"systemPrompt"` // 渲染后的系统提示词
	RenderedUserPrompt   string `json:"userPrompt"`   // 渲染后的用户提示词
	AccessAuditID        string `json:"-"`            // 内部敏感访问审计记录ID
}

// AIPromptTestResult 表示AI提示词测试结果响应数据。
type AIPromptTestResult struct {
	ID               string               `json:"-"`                // 内部测试记录ID
	CapabilityCode   string               `json:"capabilityCode"`   // 能力编码
	ConfigVersion    int64                `json:"configVersion"`    // 当前配置版本
	ModelID          string               `json:"-"`                // 内部模型ID
	PromptHash       string               `json:"promptHash"`       // 提示词摘要
	ProviderName     string               `json:"providerName"`     // 供应商名称
	ModelName        string               `json:"modelName"`        // 模型名称
	RequestID        string               `json:"requestId"`        // 请求ID
	Success          bool                 `json:"success"`          // 是否成功
	DurationMS       int                  `json:"durationMs"`       // 耗时（毫秒）
	InputTokens      *int                 `json:"-"`                // 内部输入Token数
	OutputTokens     *int                 `json:"-"`                // 内部输出Token数
	EstimatedCostCNY *string              `json:"estimatedCostCny"` // 预计成本（人民币）
	SafeMessage      string               `json:"-"`                // 内部安全消息
	SafeSummary      string               `json:"safeSummary"`      // 可安全展示的测试摘要
	TestedBy         AdministratorSummary `json:"testedBy"`         // 测试管理员
	TestedAt         time.Time            `json:"testedAt"`         // 测试时间
}

// DeletedResult 表示删除结果响应数据。
type DeletedResult struct {
	Deleted bool `json:"deleted"` // 是否已删除
}

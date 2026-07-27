package response

import "time"

// ModerationCredentialStatus 表示图片审核凭证状态响应数据。
type ModerationCredentialStatus struct {
	Configured bool       `json:"configured"` // 是否已配置
	UpdatedAt  *time.Time `json:"updatedAt"`  // 更新时间
}

// ModerationSafetyPolicy 表示图片审核安全策略响应数据。
type ModerationSafetyPolicy struct {
	MandatoryScenes         []string `json:"mandatoryScenes"`         // 强制审核的图片使用场景
	ExcludedScenes          []string `json:"excludedScenes"`          // 排除场景列表
	ExecutionMode           string   `json:"executionMode"`           // 执行模式
	FailurePolicy           string   `json:"failurePolicy"`           // 失败策略
	DisabledBehavior        string   `json:"disabledBehavior"`        // 能力关闭时的客户端行为
	UserUploadSuccessStatus string   `json:"userUploadSuccessStatus"` // 用户上传图片审核通过状态
	ProviderPassedMapping   string   `json:"providerPassedMapping"`   // 供应商审核通过状态映射
	ProviderRiskMapping     string   `json:"providerRiskMapping"`     // 供应商风险等级映射
	ProviderFailureMapping  string   `json:"providerFailureMapping"`  // 供应商审核失败状态映射
}

// ModerationConfigValues 表示图片审核当前配置项响应数据。
type ModerationConfigValues struct {
	Provider        string                     `json:"provider"`        // 供应商
	Enabled         bool                       `json:"enabled"`         // 是否启用
	Region          string                     `json:"region"`          // 地域
	Endpoint        string                     `json:"endpoint"`        // 接口地址
	ServiceCode     string                     `json:"serviceCode"`     // 服务编码
	TimeoutMS       int                        `json:"timeoutMs"`       // 超时时间（毫秒）
	RetryCount      int                        `json:"retryCount"`      // 重试数量
	RetryBackoffMS  int                        `json:"retryBackoffMs"`  // 重试退避时间（毫秒）
	WorstCaseWaitMS int                        `json:"worstCaseWaitMs"` // 最坏情况下等待时间（毫秒）
	Credential      ModerationCredentialStatus `json:"credential"`      // 本次提交的供应商凭证
	SafetyPolicy    ModerationSafetyPolicy     `json:"safetyPolicy"`    // 安全策略
}

// ModerationConnectionTestResult 表示图片审核连接测试结果响应数据。
type ModerationConnectionTestResult struct {
	TestID        string               `json:"testId"`        // 测试ID
	ConfigVersion int64                `json:"configVersion"` // 当前配置版本
	ConfigHash    string               `json:"configHash"`    // 配置摘要
	Success       bool                 `json:"success"`       // 是否成功
	Category      string               `json:"category"`      // 分类
	RequestID     string               `json:"requestId"`     // 请求ID
	DurationMS    int                  `json:"durationMs"`    // 耗时（毫秒）
	TestedBy      AdministratorSummary `json:"testedBy"`      // 测试人
	TestedAt      time.Time            `json:"testedAt"`      // 测试时间
	ValidUntil    *time.Time           `json:"validUntil"`    // 有效期
	SafeMessage   string               `json:"safeMessage"`   // 安全消息
}

// ModerationConfigHealth 表示图片审核配置健康状态响应数据。
type ModerationConfigHealth struct {
	Status                    string                          `json:"status"`                    // 状态
	EffectiveConfigVersion    *int64                          `json:"effectiveConfigVersion"`    // 生效配置版本
	LastConnectionTest        *ModerationConnectionTestResult `json:"lastConnectionTest"`        // 最近连接测试
	LastModerationSucceededAt *time.Time                      `json:"lastModerationSucceededAt"` // 最近图片审核成功时间
	LastFailureAt             *time.Time                      `json:"lastFailureAt"`             // 最近失败时间
	LastFailureSafeSummary    *string                         `json:"lastFailureSafeSummary"`    // 最近失败的安全摘要
	ConsecutiveFailureCount   int                             `json:"consecutiveFailureCount"`   // 连续失败数量
	UpdatedAt                 time.Time                       `json:"updatedAt"`                 // 更新时间
}

// ModerationConfig 表示当前生效的图片审核配置。
type ModerationConfig struct {
	ModerationConfigValues                                 // 图片审核当前配置项
	ConfigVersion          int64                           `json:"configVersion"`      // 配置版本
	ConfigHash             string                          `json:"configHash"`         // 配置摘要
	Health                 ModerationConfigHealth          `json:"health"`             // 健康状态
	LastConnectionTest     *ModerationConnectionTestResult `json:"lastConnectionTest"` // 最近连接测试
	UpdatedBy              AdministratorSummary            `json:"updatedBy"`          // 更新人
	UpdatedAt              time.Time                       `json:"updatedAt"`          // 更新时间
	Reason                 string                          `json:"reason"`             // 修改原因
}

// ModerationImageTestResult 表示图片审核图片测试结果响应数据。
type ModerationImageTestResult struct {
	TestID                  string                 `json:"testId"`                  // 测试ID
	ConfigVersion           int64                  `json:"configVersion"`           // 当前配置版本
	MappedStatus            string                 `json:"mappedStatus"`            // 映射后状态
	RiskLabels              []string               `json:"riskLabels"`              // 供应商返回的风险标签
	RiskLevel               *string                `json:"riskLevel"`               // 风险等级
	Category                string                 `json:"category"`                // 分类
	ProviderRequestID       *string                `json:"providerRequestId"`       // 供应商请求ID
	RequestID               string                 `json:"requestId"`               // 请求ID
	DurationMS              int                    `json:"durationMs"`              // 耗时（毫秒）
	SafeMessage             string                 `json:"safeMessage"`             // 安全消息
	ProviderResponseSummary map[string]interface{} `json:"providerResponseSummary"` // 供应商响应脱敏摘要
	TemporaryFileCleaned    bool                   `json:"temporaryFileCleaned"`    // 临时文件是否已清理
	TestedAt                time.Time              `json:"testedAt"`                // 测试时间
}

// ImageModerationDecision 表示图片图片审核审核判定响应数据。
type ImageModerationDecision struct {
	RecordID   string   `json:"recordId"`   // 记录ID
	Status     string   `json:"status"`     // 状态
	Allowed    bool     `json:"allowed"`    // 是否允许
	RiskLabels []string `json:"riskLabels"` // 供应商返回的风险标签
	RiskLevel  *string  `json:"riskLevel"`  // 风险等级
}

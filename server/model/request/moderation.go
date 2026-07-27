package request

// ModerationEmptyQuery 表示图片审核空查询条件。
type ModerationEmptyQuery struct{}

// ModerationConfigUpdateInput 表示图片审核配置直接保存参数。
type ModerationConfigUpdateInput struct {
	Enabled         bool    `json:"enabled"`                                                    // 是否启用
	Region          string  `json:"region" binding:"required,max=64"`                           // 地域
	Endpoint        string  `json:"endpoint" binding:"required,url,max=300" checksql:"false"`   // 接口地址
	ServiceCode     string  `json:"serviceCode" binding:"required,max=64" checksql:"false"`     // 服务编码
	TimeoutMS       int     `json:"timeoutMs" binding:"required,min=1000,max=10000"`            // 超时时间（毫秒）
	RetryCount      int     `json:"retryCount" binding:"min=0,max=3"`                           // 重试数量
	RetryBackoffMS  int     `json:"retryBackoffMs" binding:"min=0,max=2000"`                    // 重试退避时间（毫秒）
	CredentialRef   *string `json:"credentialRef" binding:"omitempty,max=300" checksql:"false"` // 新供应商凭证环境变量引用，格式为 env://变量名
	Reason          string  `json:"reason" binding:"required,min=2,max=200" checksql:"false"`   // 修改原因
	ExpectedVersion int64   `json:"expectedVersion" binding:"required,min=1"`                   // 当前配置版本
}

// ModerationConfigTestInput 表示图片审核配置测试输入参数。
type ModerationConfigTestInput struct {
	ExpectedVersion *int64 `json:"expectedVersion" form:"expectedVersion" binding:"omitempty,min=1"` // 可选的当前配置版本
}

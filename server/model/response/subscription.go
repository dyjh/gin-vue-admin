package response

import "time"

// NotificationAdminSummary 表示管理端站内通知摘要。
type NotificationAdminSummary struct {
	ID         string        `json:"id"`         // 通知ID
	User       UserReference `json:"user"`       // 接收用户
	Type       string        `json:"type"`       // 通知类型
	Title      string        `json:"title"`      // 通知标题
	Read       bool          `json:"read"`       // 用户是否已读
	TargetType *string       `json:"targetType"` // 目标类型
	TargetID   *string       `json:"targetId"`   // 目标ID
	CreatedAt  time.Time     `json:"createdAt"`  // 创建时间
	ReadAt     *time.Time    `json:"readAt"`     // 用户阅读时间
}

// NotificationAdminDetail 表示管理端站内通知详情。
type NotificationAdminDetail struct {
	NotificationAdminSummary         // 站内通知摘要
	Content                  string  `json:"content"`           // 通知正文
	GenerationSource         string  `json:"generationSource"`  // 通知生成来源
	SubscribeRequired        bool    `json:"subscribeRequired"` // 是否要求订阅消息授权
	SubscribeLogID           *string `json:"subscribeLogId"`    // 对应订阅消息发送记录ID
}

// SubscribeSceneFieldDefinition 表示固定场景可提供的一个业务字段。
type SubscribeSceneFieldDefinition struct {
	Key         string `json:"key"`         // 业务字段编码
	Label       string `json:"label"`       // 管理端显示名称
	Description string `json:"description"` // 字段含义
}

// SubscribeSceneSummary 表示固定订阅消息场景及其当前绑定摘要。
type SubscribeSceneSummary struct {
	Scene                  string                          `json:"scene"`                  // 固定业务场景编码
	Name                   string                          `json:"name"`                   // 场景名称
	Purpose                string                          `json:"purpose"`                // 场景用途
	TriggerDescription     string                          `json:"triggerDescription"`     // 固定触发条件
	RecipientDescription   string                          `json:"recipientDescription"`   // 固定接收对象规则
	RequiredFields         []SubscribeSceneFieldDefinition `json:"requiredFields"`         // 固定业务字段定义
	Configured             bool                            `json:"configured"`             // 是否已绑定微信模板
	TemplateID             *string                         `json:"templateId"`             // 当前模板绑定内部ID
	WechatTemplateIDMasked *string                         `json:"wechatTemplateIdMasked"` // 脱敏后的微信模板ID
	FieldMappingSummary    []string                        `json:"fieldMappingSummary"`    // 字段映射摘要
	Enabled                bool                            `json:"enabled"`                // 是否启用
	LogCount               int64                           `json:"logCount"`               // 发送记录总数
	SendCount              int64                           `json:"sendCount"`              // 发送次数
	FailureCount           int64                           `json:"failureCount"`           // 失败次数
	Version                int                             `json:"version"`                // 当前绑定版本；未配置时为0
	UpdatedAt              *time.Time                      `json:"updatedAt"`              // 最近配置时间
}

// SubscribeSceneDetail 表示固定订阅消息场景及其完整当前绑定。
type SubscribeSceneDetail struct {
	SubscribeSceneSummary                   // 固定场景摘要
	WechatTemplateID      string            `json:"wechatTemplateId"` // 完整微信模板ID；未配置时为空
	FieldMappings         map[string]string `json:"fieldMappings"`    // 固定业务字段到微信字段的映射
}

// SubscribeTemplateSummary 表示订阅消息模板摘要。
type SubscribeTemplateSummary struct {
	ID                     string    `json:"id"`                     // 模板ID
	Name                   string    `json:"name"`                   // 模板名称
	WechatTemplateIDMasked string    `json:"wechatTemplateIdMasked"` // 脱敏后的微信模板ID
	Scene                  string    `json:"scene"`                  // 订阅场景
	Purpose                string    `json:"purpose"`                // 模板用途
	FieldMappingSummary    []string  `json:"fieldMappingSummary"`    // 字段映射摘要
	Enabled                bool      `json:"enabled"`                // 是否启用
	LogCount               int64     `json:"logCount"`               // 发送记录总数
	SendCount              int64     `json:"sendCount"`              // 发送次数
	FailureCount           int64     `json:"failureCount"`           // 失败次数
	Version                int       `json:"version"`                // 数据版本
	UpdatedAt              time.Time `json:"updatedAt"`              // 更新时间
}

// SubscribeTemplateDetail 表示订阅消息模板详情。
type SubscribeTemplateDetail struct {
	SubscribeTemplateSummary                   // 订阅消息模板摘要
	WechatTemplateID         string            `json:"wechatTemplateId"` // 微信模板ID
	FieldMappings            map[string]string `json:"fieldMappings"`    // 业务字段到微信模板字段的映射
	CreatedAt                time.Time         `json:"createdAt"`        // 创建时间
}

// SubscribeLogSummary 表示订阅消息发送记录摘要。
type SubscribeLogSummary struct {
	ID              string        `json:"id"`              // 发送记录ID
	User            UserReference `json:"user"`            // 接收用户
	TemplateID      string        `json:"templateId"`      // 模板ID
	TemplateName    string        `json:"templateName"`    // 模板名称
	Scene           string        `json:"scene"`           // 订阅场景
	Status          string        `json:"status"`          // 发送状态
	RetryCount      int           `json:"retryCount"`      // 重试次数
	WechatErrorCode *string       `json:"wechatErrorCode"` // 微信错误码
	RequestID       string        `json:"requestId"`       // 请求ID
	CreatedAt       time.Time     `json:"createdAt"`       // 创建时间
	SentAt          *time.Time    `json:"sentAt"`          // 发送成功时间
}

// SubscribeAttempt 表示订阅消息的一次发送尝试。
type SubscribeAttempt struct {
	Attempt         int       `json:"attempt"`         // 尝试序号
	Status          string    `json:"status"`          // 尝试结果
	WechatErrorCode *string   `json:"wechatErrorCode"` // 微信错误码
	ErrorSummary    *string   `json:"errorSummary"`    // 安全失败说明
	StartedAt       time.Time `json:"startedAt"`       // 开始时间
	FinishedAt      time.Time `json:"finishedAt"`      // 结束时间
}

// SubscribeLogDetail 表示订阅消息发送记录详情。
type SubscribeLogDetail struct {
	SubscribeLogSummary                           // 订阅消息发送记录摘要
	AuthorizationChecked   bool                   `json:"authorizationChecked"`   // 是否已校验授权
	AuthorizationAvailable bool                   `json:"authorizationAvailable"` // 授权是否可用
	RelatedMealID          *string                `json:"relatedMealId"`          // 关联饭局ID
	TargetPage             *string                `json:"targetPage"`             // 小程序跳转页面
	SafePayloadSummary     map[string]interface{} `json:"safePayloadSummary"`     // 脱敏消息摘要
	ErrorSummary           *string                `json:"errorSummary"`           // 安全失败说明
	Attempts               []SubscribeAttempt     `json:"attempts"`               // 各次发送尝试
}

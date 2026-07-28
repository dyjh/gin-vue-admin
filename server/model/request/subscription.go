package request

import "time"

// NotificationAdminListQuery 表示管理端站内通知列表查询条件。
type NotificationAdminListQuery struct {
	Page        int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                       // 页码
	PageSize    int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                     // 每页数量
	UserID      string     `form:"userId" json:"userId" binding:"omitempty,max=64"`                                                  // 用户ID
	Type        string     `form:"type" json:"type" binding:"omitempty,oneof=governance discoverability points feature_refund meal"` // 通知类型
	Read        *bool      `form:"read" json:"read"`                                                                                 // 是否已读
	TargetType  string     `form:"targetType" json:"targetType" binding:"omitempty,max=40"`                                          // 目标类型
	TargetID    string     `form:"targetId" json:"targetId" binding:"omitempty,max=64"`                                              // 目标ID
	CreatedFrom *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                           // 创建起始时间
	CreatedTo   *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                               // 创建结束时间
	SortBy      string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt readAt"`                                  // 排序字段
	SortOrder   string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                    // 排序方向
}

// ApplyDefaults 补齐站内通知列表的分页和排序默认值。
func (query *NotificationAdminListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "createdAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// NotificationAdminPath 表示站内通知ID路径参数。
type NotificationAdminPath struct {
	NotificationID string `uri:"notificationId" json:"notificationId" binding:"required,max=64"` // 站内通知ID
}

// SubscribeTemplateListQuery 表示订阅消息模板列表查询条件。
type SubscribeTemplateListQuery struct {
	Page      int    `form:"page" json:"page" binding:"omitempty,min=1"`                               // 页码
	PageSize  int    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`             // 每页数量
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=60" checksql:"false"`       // 名称或用途关键词
	Scene     string `form:"scene" json:"scene" binding:"omitempty,oneof=meal_status"`                 // 订阅场景
	Enabled   *bool  `form:"enabled" json:"enabled"`                                                   // 是否启用
	SortBy    string `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=updatedAt createdAt scene"` // 排序字段
	SortOrder string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`            // 排序方向
}

// ApplyDefaults 补齐订阅消息模板列表的分页和排序默认值。
func (query *SubscribeTemplateListQuery) ApplyDefaults() {
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

// SubscribeTemplatePath 表示订阅消息模板ID路径参数。
type SubscribeTemplatePath struct {
	TemplateID string `uri:"templateId" json:"templateId" binding:"required,max=64"` // 订阅消息模板ID
}

// SubscribeTemplateCreateInput 表示订阅消息模板创建参数。
type SubscribeTemplateCreateInput struct {
	Name             string            `json:"name" binding:"required,max=60" checksql:"false"`              // 模板名称
	WechatTemplateID string            `json:"wechatTemplateId" binding:"required,max=100" checksql:"false"` // 微信模板ID
	Scene            string            `json:"scene" binding:"required,oneof=meal_status"`                   // 订阅场景
	Purpose          string            `json:"purpose" binding:"required,max=160" checksql:"false"`          // 模板用途
	FieldMappings    map[string]string `json:"fieldMappings" binding:"required,min=1"`                       // 业务字段到微信模板字段的映射
}

// SubscribeTemplateUpdateInput 表示订阅消息模板更新参数。
type SubscribeTemplateUpdateInput struct {
	SubscribeTemplateCreateInput     // 模板基础配置
	ExpectedVersion              int `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// SubscribeTemplateStatusInput 表示订阅消息模板启停参数。
type SubscribeTemplateStatusInput struct {
	Enabled         *bool  `json:"enabled" binding:"required"`               // 是否启用
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 启停原因
	ExpectedVersion int    `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// SubscribeTemplateDeleteInput 表示订阅消息模板删除参数。
type SubscribeTemplateDeleteInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 删除原因
	ExpectedVersion int    `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// SubscribeScenePath 表示固定订阅消息场景路径参数。
type SubscribeScenePath struct {
	Scene string `uri:"scene" json:"scene" binding:"required,oneof=meal_status"` // 固定业务场景
}

// SubscribeSceneConfigureInput 表示固定场景的微信模板绑定参数。
type SubscribeSceneConfigureInput struct {
	WechatTemplateID string            `json:"wechatTemplateId" binding:"required,max=100" checksql:"false"` // 微信模板ID
	FieldMappings    map[string]string `json:"fieldMappings" binding:"required,len=3"`                       // 固定业务字段到微信字段的映射
	ExpectedVersion  *int              `json:"expectedVersion" binding:"omitempty,min=1"`                    // 已配置场景的预期版本
}

// SubscribeSceneStatusInput 表示固定订阅场景启停参数。
type SubscribeSceneStatusInput struct {
	Enabled         *bool  `json:"enabled" binding:"required"`               // 是否启用
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 启停原因
	ExpectedVersion int    `json:"expectedVersion" binding:"required,min=1"` // 预期版本
}

// SubscribeLogListQuery 表示订阅消息发送记录列表查询条件。
type SubscribeLogListQuery struct {
	Page        int        `form:"page" json:"page" binding:"omitempty,min=1"`                                 // 页码
	PageSize    int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`               // 每页数量
	UserID      string     `form:"userId" json:"userId" binding:"omitempty,max=64"`                            // 用户ID
	TemplateID  string     `form:"templateId" json:"templateId" binding:"omitempty,max=64"`                    // 订阅消息模板ID
	Scene       string     `form:"scene" json:"scene" binding:"omitempty,oneof=meal_status"`                   // 订阅场景
	Status      string     `form:"status" json:"status" binding:"omitempty,oneof=pending sending sent failed"` // 发送状态
	RequestID   string     `form:"requestId" json:"requestId" binding:"omitempty,max=128" checksql:"false"`    // 请求ID
	CreatedFrom *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`     // 创建起始时间
	CreatedTo   *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`         // 创建结束时间
	SortBy      string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt sentAt retryCount"` // 排序字段
	SortOrder   string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`              // 排序方向
}

// ApplyDefaults 补齐订阅消息发送记录列表的分页和排序默认值。
func (query *SubscribeLogListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "createdAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// SubscribeLogPath 表示订阅消息发送记录ID路径参数。
type SubscribeLogPath struct {
	LogID string `uri:"logId" json:"logId" binding:"required,max=64"` // 订阅消息发送记录ID
}

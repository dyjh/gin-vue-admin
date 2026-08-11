package response

import (
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	"time"
)

// AuditLog 表示审计日志响应数据。
type AuditLog struct {
	ID             string                              `json:"id"`             // ID
	Administrator  commonResponse.AdministratorSummary `json:"administrator"`  // 管理员
	Action         string                              `json:"action"`         // 操作
	TargetType     string                              `json:"targetType"`     // 目标类型
	TargetID       string                              `json:"targetId"`       // 目标ID
	TargetLabel    *string                             `json:"targetLabel"`    // 目标显示名称
	Reason         *string                             `json:"reason"`         // 原因
	RequestID      string                              `json:"requestId"`      // 请求ID
	IdempotencyKey *string                             `json:"idempotencyKey"` // 幂等键
	CreatedAt      time.Time                           `json:"createdAt"`      // 创建时间
}

// AuditLogDetail 表示审计日志详情响应数据。
type AuditLogDetail struct {
	AuditLog                     // 操作日志摘要
	BeforeSummary    interface{} `json:"beforeSummary"`    // 变更前摘要
	AfterSummary     interface{} `json:"afterSummary"`     // 变更后摘要
	SourceIPMasked   *string     `json:"sourceIpMasked"`   // 脱敏后的来源IP
	UserAgentSummary *string     `json:"userAgentSummary"` // 用户代理摘要
}

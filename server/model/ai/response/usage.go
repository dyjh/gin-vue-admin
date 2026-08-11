package response

import (
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	"time"
)

// AIUsageSummary 表示管理端AI调用记录摘要。
type AIUsageSummary struct {
	ID                      string                       `json:"id"`                      // 调用记录ID
	User                    commonResponse.UserReference `json:"user"`                    // 调用用户
	CapabilityCode          string                       `json:"capabilityCode"`          // 实际AI能力编码
	ProviderID              string                       `json:"providerId"`              // 实际供应商ID
	ProviderName            string                       `json:"providerName"`            // 实际供应商名称
	ModelID                 string                       `json:"modelId"`                 // 实际模型ID
	ModelName               string                       `json:"modelName"`               // 实际模型名称
	RequestID               string                       `json:"requestId"`               // 请求ID
	IdempotencyKey          string                       `json:"idempotencyKey"`          // 请求幂等键
	ExecutionStatus         string                       `json:"executionStatus"`         // 执行状态
	BillingStatus           string                       `json:"billingStatus"`           // 计费状态
	PromptMode              string                       `json:"promptMode"`              // 提示词模式
	PromptHash              string                       `json:"promptHash"`              // 提示词内容摘要
	DurationMS              *int                         `json:"durationMs"`              // 调用耗时毫秒数
	InputTokens             *int                         `json:"inputTokens"`             // 输入Token数
	OutputTokens            *int                         `json:"outputTokens"`            // 输出Token数
	ImageCount              int                          `json:"imageCount"`              // 图片数
	PointCost               int                          `json:"pointCost"`               // 积分消耗
	EstimatedCostCNY        string                       `json:"estimatedCostCny"`        // 估算成本人民币
	Currency                string                       `json:"currency"`                // 成本币种
	SensitiveContentCleared bool                         `json:"sensitiveContentCleared"` // 敏感内容是否已清除
	CreatedAt               time.Time                    `json:"createdAt"`               // 创建时间
	FinishedAt              *time.Time                   `json:"finishedAt"`              // 调用结束时间
}

// AIUsageTimelineEvent 表示AI调用执行、计费和敏感内容清除时间线事件。
type AIUsageTimelineEvent struct {
	Type       string    `json:"type"`       // 事件类型
	Label      string    `json:"label"`      // 事件说明
	OccurredAt time.Time `json:"occurredAt"` // 发生时间
}

// AIUsageDetail 表示管理端AI调用详情。
type AIUsageDetail struct {
	AIUsageSummary                                           // AI调用记录摘要
	FailureCategory     *string                              `json:"failureCategory"`         // 失败分类
	FailureSummary      *string                              `json:"failureSummary"`          // 安全失败摘要
	ChargedPointEntryID *string                              `json:"chargedPointEntryId"`     // 扣积分流水ID
	RefundPointEntryID  *string                              `json:"refundPointEntryId"`      // 退款流水ID
	Timeline            []AIUsageTimelineEvent               `json:"timeline"`                // 调用与计费时间线
	OriginalInput       map[string]interface{}               `json:"originalInput,omitempty"` // 原始输入，仅按敏感读取权限返回
	ImageMetadata       []map[string]interface{}             `json:"imageMetadata,omitempty"` // 图片元信息，仅按敏感读取权限返回
	ModelOutput         interface{}                          `json:"modelOutput,omitempty"`   // 模型输出，仅按敏感读取权限返回
	ClearedAt           *time.Time                           `json:"clearedAt"`               // 敏感内容清除时间
	ClearedBy           *commonResponse.AdministratorSummary `json:"clearedBy"`               // 清除管理员
	Version             int64                                `json:"version"`                 // 数据版本
}

// AIUsageSensitiveDeleteResult 表示清除AI调用敏感内容的结果。
type AIUsageSensitiveDeleteResult struct {
	Cleared             bool      `json:"cleared"`             // 是否已清除
	ClearedAt           time.Time `json:"clearedAt"`           // 清除时间
	RetainedAuditFields []string  `json:"retainedAuditFields"` // 保留的审计字段
}

// AIPromptConfig 表示当前生效的AI提示词配置。
type AIPromptConfig struct {
	CapabilityCode     string                              `json:"capabilityCode"`     // AI能力编码
	Mode               string                              `json:"mode"`               // 提示词模式
	SystemPrompt       string                              `json:"systemPrompt"`       // 系统提示词
	UserPromptTemplate string                              `json:"userPromptTemplate"` // 用户提示词模板
	PromptHash         string                              `json:"promptHash"`         // 提示词内容摘要
	Version            int64                               `json:"version"`            // 当前能力配置版本
	UpdatedBy          commonResponse.AdministratorSummary `json:"updatedBy"`          // 更新管理员
	UpdatedAt          time.Time                           `json:"updatedAt"`          // 更新时间
	Reason             string                              `json:"reason"`             // 修改原因
}

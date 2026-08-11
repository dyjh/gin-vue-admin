package request

import (
	"time"
)

// AIUsageListQuery 表示管理端AI调用记录分页查询条件。
type AIUsageListQuery struct {
	Page            int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                               // 页码
	PageSize        int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                             // 每页数量
	UserID          string     `form:"userId" json:"userId" binding:"omitempty,max=64"`                                                          // 用户ID
	CapabilityCode  string     `form:"capabilityCode" json:"capabilityCode" binding:"omitempty,max=64"`                                          // AI能力编码
	ProviderID      string     `form:"providerId" json:"providerId" binding:"omitempty,max=64"`                                                  // 供应商ID
	ModelID         string     `form:"modelId" json:"modelId" binding:"omitempty,max=64"`                                                        // 模型ID
	RequestID       string     `form:"requestId" json:"requestId" binding:"omitempty,max=128"`                                                   // 请求ID
	IdempotencyKey  string     `form:"idempotencyKey" json:"idempotencyKey" binding:"omitempty,max=128"`                                         // 请求幂等键
	ExecutionStatus string     `form:"executionStatus" json:"executionStatus" binding:"omitempty,oneof=pending processing succeeded failed"`     // 执行状态
	BillingStatus   string     `form:"billingStatus" json:"billingStatus" binding:"omitempty,oneof=not_charged charged refund_pending refunded"` // 计费状态
	CreatedFrom     *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                   // 创建起始时间
	CreatedTo       *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                                       // 创建结束时间
	SortBy          string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt durationMs estimatedCostCny"`                     // 排序字段
	SortOrder       string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                            // 排序方向
}

// ApplyDefaults 补齐AI调用记录分页和排序默认值。
func (query *AIUsageListQuery) ApplyDefaults() {
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

// AIUsageIDPath 表示AI调用记录ID路径参数。
type AIUsageIDPath struct {
	UsageID string `uri:"usageId" json:"usageId" binding:"required,max=64"` // AI调用记录ID
}

// AIUsageDetailQuery 表示AI调用详情查询条件。
type AIUsageDetailQuery struct {
	IncludeSensitiveContent bool `form:"includeSensitiveContent" json:"includeSensitiveContent"` // 是否读取敏感输入输出
}

// AIUsageSensitiveDeleteInput 表示清除AI调用敏感内容的输入参数。
type AIUsageSensitiveDeleteInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 清除原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"` // 预期数据版本
}

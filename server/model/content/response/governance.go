package response

import (
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	"time"
)

// GovernanceRecordSummary 表示违规处理记录摘要。
type GovernanceRecordSummary struct {
	ID            string                              `json:"id"`            // 记录ID
	TargetType    string                              `json:"targetType"`    // 目标类型
	TargetID      string                              `json:"targetId"`      // 目标ID
	TargetLabel   string                              `json:"targetLabel"`   // 目标名称
	Actions       []string                            `json:"actions"`       // 处理动作列表
	ViolationType string                              `json:"violationType"` // 违规类型
	Severity      string                              `json:"severity"`      // 严重程度
	Reason        string                              `json:"reason"`        // 处理原因
	AffectedCount int                                 `json:"affectedCount"` // 已受影响数量
	Administrator commonResponse.AdministratorSummary `json:"administrator"` // 执行管理员
	JobStatus     *string                             `json:"jobStatus"`     // 异步任务状态
	JobProgress   *GovernanceJobProgress              `json:"jobProgress"`   // 异步任务进度
	CreatedAt     time.Time                           `json:"createdAt"`     // 创建时间
}

// GovernanceJobProgress 表示违规处理异步任务的数量进度。
type GovernanceJobProgress struct {
	TotalCount     int `json:"totalCount"`     // 目标总数
	SucceededCount int `json:"succeededCount"` // 成功数量
	FailedCount    int `json:"failedCount"`    // 失败数量
	PendingCount   int `json:"pendingCount"`   // 待处理数量
}

// GovernanceJobItem 表示违规处理异步任务中的单个处理项。
type GovernanceJobItem struct {
	ID               string     `json:"id"`               // 任务项ID
	DishID           string     `json:"dishId"`           // 菜品公开ID
	UserID           string     `json:"userId"`           // 所属用户ID
	Status           string     `json:"status"`           // 处理状态
	AttemptCount     int        `json:"attemptCount"`     // 执行次数
	LastErrorSummary *string    `json:"lastErrorSummary"` // 最近错误摘要
	NotificationID   *string    `json:"notificationId"`   // 站内通知ID
	ProcessedAt      *time.Time `json:"processedAt"`      // 最近处理时间
	UpdatedAt        time.Time  `json:"updatedAt"`        // 更新时间
}

// GovernanceJob 表示违规处理异步任务进度。
type GovernanceJob struct {
	ID               string              `json:"id"`               // 任务ID
	RecordID         string              `json:"recordId"`         // 违规处理记录ID
	Status           string              `json:"status"`           // 任务状态
	TotalCount       int                 `json:"totalCount"`       // 目标总数
	SucceededCount   int                 `json:"succeededCount"`   // 成功数量
	FailedCount      int                 `json:"failedCount"`      // 失败数量
	PendingCount     int                 `json:"pendingCount"`     // 待处理数量
	LastErrorSummary *string             `json:"lastErrorSummary"` // 最近错误摘要
	Version          int64               `json:"version"`          // 数据版本
	StartedAt        *time.Time          `json:"startedAt"`        // 开始时间
	FinishedAt       *time.Time          `json:"finishedAt"`       // 完成时间
	UpdatedAt        time.Time           `json:"updatedAt"`        // 更新时间
	Items            []GovernanceJobItem `json:"items"`            // 任务项明细
}

// GovernanceRecordDetail 表示违规处理记录详情。
type GovernanceRecordDetail struct {
	GovernanceRecordSummary                         // 违规处理记录摘要
	ImpactSnapshot          GovernanceImpactPreview `json:"impactSnapshot"`  // 影响范围快照
	BeforeSummary           interface{}             `json:"beforeSummary"`   // 处理前摘要
	AfterSummary            interface{}             `json:"afterSummary"`    // 处理后摘要
	Job                     *GovernanceJob          `json:"job"`             // 异步任务
	NotificationIDs         []string                `json:"notificationIds"` // 站内通知ID列表
	RequestID               string                  `json:"requestId"`       // 请求ID
	IdempotencyKey          string                  `json:"idempotencyKey"`  // 幂等键
	TargetExists            bool                    `json:"targetExists"`    // 当前目标是否仍存在
	TargetVersion           int                     `json:"targetVersion"`   // 当前目标版本
}

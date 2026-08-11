package content

import (
	"time"
)

const (
	GovernanceJobPending            = "pending"
	GovernanceJobProcessing         = "processing"
	GovernanceJobPartiallySucceeded = "partially_succeeded"
	GovernanceJobSucceeded          = "succeeded"
	GovernanceJobFailed             = "failed"
)

// GovernanceJob 表示违规复制链异步处理任务。
type GovernanceJob struct {
	ID               string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:任务ID;"`                           // 任务ID
	RecordID         string     `json:"recordId" gorm:"column:record_id;type:varchar(64);not null;uniqueIndex;comment:违规处理记录ID;"`         // 违规处理记录ID
	Status           string     `json:"status" gorm:"column:status;type:varchar(32);not null;index;comment:任务状态;"`                        // 任务状态
	TotalCount       int        `json:"totalCount" gorm:"column:total_count;type:int;not null;default:0;comment:目标总数;"`                   // 目标总数
	SucceededCount   int        `json:"succeededCount" gorm:"column:succeeded_count;type:int;not null;default:0;comment:成功数量;"`           // 成功数量
	FailedCount      int        `json:"failedCount" gorm:"column:failed_count;type:int;not null;default:0;comment:失败数量;"`                 // 失败数量
	PendingCount     int        `json:"pendingCount" gorm:"column:pending_count;type:int;not null;default:0;comment:待处理数量;"`              // 待处理数量
	LastErrorSummary *string    `json:"lastErrorSummary" gorm:"column:last_error_summary;type:varchar(240);default:null;comment:最近错误摘要;"` // 最近错误摘要
	Version          int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                       // 数据版本
	StartedAt        *time.Time `json:"startedAt" gorm:"column:started_at;type:datetime;default:null;comment:开始时间;"`                      // 开始时间
	FinishedAt       *time.Time `json:"finishedAt" gorm:"column:finished_at;type:datetime;default:null;comment:完成时间;"`                    // 完成时间
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                    // 创建时间
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                    // 更新时间
}

// TableName 指定GovernanceJob对应的数据表名。
func (GovernanceJob) TableName() string { return "of_gov_jobs" }

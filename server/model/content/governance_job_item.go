package content

import (
	"time"
)

const (
	GovernanceJobItemPending   = "pending"
	GovernanceJobItemSucceeded = "succeeded"
	GovernanceJobItemFailed    = "failed"
)

// GovernanceJobItem 表示违规复制链任务中的单个菜品处理项。
type GovernanceJobItem struct {
	ID               string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:任务项ID;"`                                                // 任务项ID
	JobID            string     `json:"jobId" gorm:"column:job_id;type:varchar(64);not null;uniqueIndex:uk_of_gov_item,priority:1;index;comment:任务ID;"`         // 任务ID
	DishID           uint       `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;uniqueIndex:uk_of_gov_item,priority:2;index;comment:菜品内部ID;"` // 菜品内部ID
	DishPublicID     string     `json:"dishPublicId" gorm:"column:dish_public_id;type:varchar(64);not null;comment:菜品公开ID;"`                                    // 菜品公开ID
	UserID           string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:所属用户ID;"`                                           // 所属用户ID
	Status           string     `json:"status" gorm:"column:status;type:varchar(24);not null;index;comment:任务项状态;"`                                             // 任务项状态
	AttemptCount     int        `json:"attemptCount" gorm:"column:attempt_count;type:int;not null;default:0;comment:执行次数;"`                                     // 执行次数
	LastErrorSummary *string    `json:"lastErrorSummary" gorm:"column:last_error_summary;type:varchar(240);default:null;comment:最近错误摘要;"`                       // 最近错误摘要
	NotificationID   *string    `json:"notificationId" gorm:"column:notification_id;type:varchar(64);default:null;comment:已生成站内通知ID;"`                          // 已生成站内通知ID
	ProcessedAt      *time.Time `json:"processedAt" gorm:"column:processed_at;type:datetime;default:null;comment:最近处理时间;"`                                      // 最近处理时间
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                // 创建时间
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                          // 更新时间
}

// TableName 指定GovernanceJobItem对应的数据表名。
func (GovernanceJobItem) TableName() string { return "of_gov_items" }

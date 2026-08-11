package user

import (
	"time"
)

// PreferenceEvidenceAggregate 表示偏好画像证据聚合。
type PreferenceEvidenceAggregate struct {
	ID               uint       `json:"id" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                                      // 主键ID
	UserID           string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_preference_evidence,priority:1;comment:用户ID;"`         // 用户ID
	SourceType       string     `json:"sourceType" gorm:"column:source_type;type:varchar(40);not null;uniqueIndex:uk_of_preference_evidence,priority:2;comment:来源类型;"` // 来源类型
	TotalCount       int        `json:"totalCount" gorm:"column:total_count;type:int;not null;default:0;comment:证据总数;"`                                                // 证据总数
	AggregatedCount  int        `json:"aggregatedCount" gorm:"column:aggregated_count;type:int;not null;default:0;comment:已聚合证据数;"`                                    // 已聚合证据数
	PendingCount     int        `json:"pendingCount" gorm:"column:pending_count;type:int;not null;default:0;comment:待聚合证据数;"`                                          // 待聚合证据数
	LatestOccurredAt *time.Time `json:"latestOccurredAt" gorm:"column:latest_occurred_at;type:datetime;default:null;comment:最近证据发生时间;"`                                // 最近证据发生时间
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                       // 创建时间
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                       // 更新时间
}

// TableName 指定PreferenceEvidenceAggregate对应的数据表名。
func (PreferenceEvidenceAggregate) TableName() string {
	return "of_preference_evidence"
}

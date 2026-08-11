package user

import (
	"gorm.io/datatypes"
	"time"
)

// PreferenceUpdateState 表示用户偏好画像的聚合更新状态。
type PreferenceUpdateState string

const (
	// PreferenceUpdateActive 表示偏好画像已经完成最近一次聚合。
	PreferenceUpdateActive PreferenceUpdateState = "active"
	// PreferenceUpdatePending 表示存在待处理的偏好证据。
	PreferenceUpdatePending PreferenceUpdateState = "pending"
	// PreferenceUpdatePaused 表示平台整体增强能力关闭后暂停更新。
	PreferenceUpdatePaused PreferenceUpdateState = "paused"
	// PreferenceUpdateFailed 表示最近一次偏好证据处理失败。
	PreferenceUpdateFailed PreferenceUpdateState = "failed"
)

// UserPreferenceProfile 表示用户偏好画像。
type UserPreferenceProfile struct {
	ID                  string                `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                      // 主键ID
	UserID              string                `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex;comment:用户ID;"`                            // 用户ID
	HasProfile          bool                  `json:"hasProfile" gorm:"column:has_profile;type:tinyint(1) unsigned;not null;default:false;comment:是否已生成画像;"`       // 是否已生成画像
	UpdateState         PreferenceUpdateState `json:"updateState" gorm:"column:update_state;type:varchar(16);not null;default:pending;index;comment:画像更新状态;"`      // 画像更新状态
	UpdateEnabled       bool                  `json:"updateEnabled" gorm:"column:update_enabled;type:tinyint(1) unsigned;not null;default:true;comment:是否允许更新画像;"` // 是否允许更新画像
	PausedReasonCode    *string               `json:"pausedReasonCode" gorm:"column:paused_reason_code;type:varchar(80);default:null;comment:暂停原因码;"`              // 暂停原因码
	PausedReasonSummary *string               `json:"pausedReasonSummary" gorm:"column:paused_reason_summary;type:varchar(240);default:null;comment:暂停原因说明;"`      // 暂停原因说明
	LastEvidenceAt      *time.Time            `json:"lastEvidenceAt" gorm:"column:last_evidence_at;type:datetime;default:null;comment:最近证据时间;"`                    // 最近证据时间
	LastAggregatedAt    *time.Time            `json:"lastAggregatedAt" gorm:"column:last_aggregated_at;type:datetime;default:null;comment:最近聚合时间;"`                // 最近聚合时间
	ProfileUpdatedAt    *time.Time            `json:"profileUpdatedAt" gorm:"column:profile_updated_at;type:datetime;default:null;comment:画像更新时间;"`                // 画像更新时间
	ProfileJSON         datatypes.JSON        `json:"-" gorm:"column:profile_json;type:json;default:null;comment:偏好画像JSON;"`                                       // 偏好画像JSON
	LatestFailureJSON   datatypes.JSON        `json:"-" gorm:"column:latest_failure_json;type:json;default:null;comment:最近画像失败信息JSON;"`                            // 最近画像失败信息JSON
	Version             int64                 `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                  // 数据版本
	CreatedAt           time.Time             `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                     // 创建时间
	UpdatedAt           time.Time             `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                     // 更新时间
}

// TableName 指定UserPreferenceProfile对应的数据表名。
func (UserPreferenceProfile) TableName() string {
	return "of_user_preferences"
}

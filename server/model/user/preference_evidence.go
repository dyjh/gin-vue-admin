package user

import (
	"gorm.io/datatypes"
	"time"
)

const (
	// PreferenceSourceCheckinImage 表示打卡图片证据。
	PreferenceSourceCheckinImage = "checkin_image"
	// PreferenceSourceRecommendationAdopted 表示采用推荐证据。
	PreferenceSourceRecommendationAdopted = "recommendation_adopted"
	// PreferenceSourceDishSaved 表示保存可用菜品证据。
	PreferenceSourceDishSaved = "dish_saved"
	// PreferenceSourceDishOrdered 表示饭局点菜证据。
	PreferenceSourceDishOrdered = "dish_ordered"
	// PreferenceSourceReshuffle 表示换一道的弱负向证据。
	PreferenceSourceReshuffle = "reshuffle"
	// PreferenceSourceSkipForNow 表示先不考虑的中等负向证据。
	PreferenceSourceSkipForNow = "skip_for_now"
	// PreferenceSourceExplicitSetting 表示用户明确设置的硬条件或偏好。
	PreferenceSourceExplicitSetting = "explicit_setting"

	// PreferenceEvidencePending 表示证据等待分析或聚合。
	PreferenceEvidencePending = "pending"
	// PreferenceEvidenceAggregated 表示证据已经进入画像。
	PreferenceEvidenceAggregated = "aggregated"
	// PreferenceEvidenceFailed 表示证据达到重试上限后处理失败。
	PreferenceEvidenceFailed = "failed"
)

// PreferenceEvidence 表示一条可追溯但不向管理端暴露明细的偏好画像证据。
type PreferenceEvidence struct {
	ID             string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                                 // 主键ID
	UserID         string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;uniqueIndex:uk_of_pref_evidence_source,priority:1;comment:用户ID;"`           // 用户ID
	SourceType     string         `json:"sourceType" gorm:"column:source_type;type:varchar(40);not null;index;uniqueIndex:uk_of_pref_evidence_source,priority:2;comment:证据来源类型;"` // 证据来源类型
	SourceID       string         `json:"sourceId" gorm:"column:source_id;type:varchar(96);not null;uniqueIndex:uk_of_pref_evidence_source,priority:3;comment:来源业务ID;"`           // 来源业务ID
	Direction      string         `json:"direction" gorm:"column:direction;type:varchar(12);not null;comment:正向或负向;"`                                                             // 正向或负向
	Confidence     float64        `json:"confidence" gorm:"column:confidence;type:decimal(5,4);not null;default:1;comment:证据置信度;"`                                                // 证据置信度
	Weight         float64        `json:"weight" gorm:"column:weight;type:decimal(6,3);not null;comment:聚合权重;"`                                                                   // 聚合权重
	FactsJSON      datatypes.JSON `json:"-" gorm:"column:facts_json;type:json;not null;comment:结构化事实JSON;"`                                                                       // 结构化事实JSON
	Status         string         `json:"status" gorm:"column:status;type:varchar(16);not null;default:pending;index;comment:处理状态;"`                                              // 处理状态
	RetryCount     int            `json:"retryCount" gorm:"column:retry_count;type:int;not null;default:0;comment:处理重试次数;"`                                                       // 处理重试次数
	FailureSummary *string        `json:"failureSummary" gorm:"column:failure_summary;type:varchar(240);default:null;comment:安全失败摘要;"`                                            // 安全失败摘要
	RequestID      string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;default:'';index;comment:来源请求ID;"`                                         // 来源请求ID
	OccurredAt     time.Time      `json:"occurredAt" gorm:"column:occurred_at;type:datetime;not null;index;comment:证据发生时间;"`                                                      // 证据发生时间
	AggregatedAt   *time.Time     `json:"aggregatedAt" gorm:"column:aggregated_at;type:datetime;default:null;comment:证据聚合时间;"`                                                    // 证据聚合时间
	CreatedAt      time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                // 创建时间
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                                          // 更新时间
}

// TableName 指定PreferenceEvidence对应的数据表名。
func (PreferenceEvidence) TableName() string {
	return "of_pref_evidence"
}

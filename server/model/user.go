package model

import (
	"time"

	"gorm.io/datatypes"
)

// UserStatus 表示小程序用户账号状态。
type UserStatus string

const (
	UserStatusNormal   UserStatus = "normal"
	UserStatusDisabled UserStatus = "disabled"
)

// MiniAppUser keeps WeChat credentials as storage-only fields. API response
// types live in model/orderfood/response and never include OpenID or
// SessionKeyEncrypted.
type MiniAppUser struct {
	ID                  string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                      // 主键ID
	OpenIDHash          string     `json:"-" gorm:"column:open_id_hash;type:varchar(64);not null;uniqueIndex;comment:微信OpenID摘要;"`      // 微信OpenID摘要
	OpenIDEncrypted     string     `json:"-" gorm:"column:open_id_encrypted;type:text;not null;comment:加密后的微信OpenID;"`                  // 加密后的微信OpenID
	UnionIDHash         *string    `json:"-" gorm:"column:union_id_hash;type:varchar(64);index;default:null;comment:微信UnionID摘要;"`      // 微信UnionID摘要
	UnionIDEncrypted    *string    `json:"-" gorm:"column:union_id_encrypted;type:text;default:null;comment:加密后的微信UnionID;"`            // 加密后的微信UnionID
	SessionKeyEncrypted *string    `json:"-" gorm:"column:session_key_encrypted;type:text;default:null;comment:加密后的微信会话密钥;"`            // 加密后的微信会话密钥
	AvatarURL           *string    `json:"avatarUrl" gorm:"column:avatar_url;type:varchar(500);default:null;comment:头像地址;"`             // 头像地址
	Nickname            string     `json:"nickname" gorm:"column:nickname;type:varchar(30);not null;comment:用户昵称;"`                     // 用户昵称
	Points              int64      `json:"points" gorm:"column:points;type:bigint;not null;default:0;index;comment:当前积分;"`              // 当前积分
	CheckinDayCount     int        `json:"checkinDayCount" gorm:"column:checkin_day_count;type:int;not null;default:0;comment:累计打卡天数;"` // 累计打卡天数
	DishCount           int        `json:"dishCount" gorm:"column:dish_count;type:int;not null;default:0;comment:菜品数量;"`                // 菜品数量
	RecipeCount         int        `json:"recipeCount" gorm:"column:recipe_count;type:int;not null;default:0;comment:菜谱数量;"`            // 菜谱数量
	MealCount           int        `json:"mealCount" gorm:"column:meal_count;type:int;not null;default:0;comment:饭局数量;"`                // 饭局数量
	CheckinCount        int        `json:"checkinCount" gorm:"column:checkin_count;type:int;not null;default:0;comment:打卡次数;"`          // 打卡次数
	Status              UserStatus `json:"status" gorm:"column:status;type:varchar(16);not null;default:normal;index;comment:状态;"`      // 状态
	DisabledReason      *string    `json:"disabledReason" gorm:"column:disabled_reason;type:varchar(200);default:null;comment:禁用原因;"`   // 禁用原因
	DisabledAt          *time.Time `json:"disabledAt" gorm:"column:disabled_at;type:datetime;default:null;comment:禁用时间;"`               // 禁用时间
	Version             int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                  // 数据版本
	RegisteredAt        time.Time  `json:"registeredAt" gorm:"column:registered_at;type:datetime;not null;index;comment:注册时间;"`         // 注册时间
	LastLoginAt         *time.Time `json:"lastLoginAt" gorm:"column:last_login_at;type:datetime;index;default:null;comment:最近登录时间;"`    // 最近登录时间
	CreatedAt           time.Time  `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                             // 创建时间
	UpdatedAt           time.Time  `json:"-" gorm:"column:updated_at;type:datetime;index;not null;comment:更新时间;"`                       // 更新时间
}

// TableName 指定MiniAppUser对应的数据表名。
func (MiniAppUser) TableName() string {
	return "of_users"
}

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

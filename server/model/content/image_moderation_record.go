package content

import (
	"gorm.io/datatypes"
	"time"
)

// ModerationScene 表示图片进入平台时的业务使用场景。
type ModerationScene string

const (
	ModerationSceneProfileAvatar  ModerationScene = "profile_avatar"
	ModerationSceneDishCover      ModerationScene = "dish_cover"
	ModerationSceneDishStep       ModerationScene = "dish_step"
	ModerationSceneDishExtract    ModerationScene = "dish_extract"
	ModerationSceneCheckin        ModerationScene = "checkin"
	ModerationSceneGeneratedCover ModerationScene = "generated_cover"
	ModerationSceneOfficialCover  ModerationScene = "official_dish_cover"
)

// ModerationStatus 表示平台归一化后的图片审核状态。
type ModerationStatus string

const (
	ModerationStatusPending     ModerationStatus = "pending"
	ModerationStatusPassed      ModerationStatus = "passed"
	ModerationStatusRejected    ModerationStatus = "rejected"
	ModerationStatusFailed      ModerationStatus = "failed"
	ModerationStatusNotRequired ModerationStatus = "not_required"
)

// ImageModerationRecord is written for every gate decision. FileID is a
// staging identifier until a passed decision allows the business resource to
// be committed.
type ImageModerationRecord struct {
	ID                  string           `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                              // 主键ID
	RequestID           string           `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                   // 请求ID
	FileID              string           `json:"fileId" gorm:"column:file_id;type:varchar(128);not null;index;comment:文件ID;"`                         // 文件ID
	Scene               ModerationScene  `json:"scene" gorm:"column:scene;type:varchar(32);not null;index;comment:审核场景;"`                             // 审核场景
	Status              ModerationStatus `json:"status" gorm:"column:status;type:varchar(20);not null;index;comment:状态;"`                             // 状态
	RiskLabelsJSON      datatypes.JSON   `json:"-" gorm:"column:risk_labels_json;type:json;not null;comment:风险标签JSON;"`                               // 风险标签JSON
	RiskLevel           *string          `json:"riskLevel" gorm:"column:risk_level;type:varchar(40);default:null;comment:风险等级;"`                      // 风险等级
	DurationMS          *int             `json:"durationMs" gorm:"column:duration_ms;type:int;default:null;comment:耗时（毫秒）;"`                          // 耗时（毫秒）
	ErrorCode           *string          `json:"errorCode" gorm:"column:error_code;type:varchar(80);default:null;comment:错误码;"`                       // 错误码
	ErrorSummary        *string          `json:"errorSummary" gorm:"column:error_summary;type:varchar(240);default:null;comment:错误摘要;"`               // 错误摘要
	ProviderRequestID   *string          `json:"providerRequestId" gorm:"column:provider_request_id;type:varchar(128);default:null;comment:供应商请求ID;"` // 供应商请求ID
	ProviderSummaryJSON datatypes.JSON   `json:"-" gorm:"column:provider_summary_json;type:json;default:null;comment:供应商响应摘要JSON;"`                   // 供应商响应摘要JSON
	ConfigVersion       *int64           `json:"configVersion" gorm:"column:config_version;type:bigint;default:null;comment:配置版本;"`                   // 配置版本
	ObjectType          *string          `json:"objectType" gorm:"column:object_type;type:varchar(60);default:null;comment:对象类型;"`                    // 对象类型
	ObjectID            *string          `json:"objectId" gorm:"column:object_id;type:varchar(80);default:null;comment:对象ID;"`                        // 对象ID
	CreatedAt           time.Time        `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                       // 创建时间
}

// TableName 指定ImageModerationRecord对应的数据表名。
func (ImageModerationRecord) TableName() string {
	return "of_mod_records"
}

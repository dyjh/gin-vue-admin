package orderfood

import (
	"time"

	"gorm.io/datatypes"
)

// ModerationProviderType 表示图片审核供应商类型。
type ModerationProviderType string

const (
	ModerationProviderAliyun ModerationProviderType = "aliyun"
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

// ModerationHealthStatus 表示当前图片审核配置的运行健康状态。
type ModerationHealthStatus string

const (
	ModerationHealthUnconfigured ModerationHealthStatus = "unconfigured"
	ModerationHealthDisabled     ModerationHealthStatus = "disabled"
	ModerationHealthHealthy      ModerationHealthStatus = "healthy"
	ModerationHealthDegraded     ModerationHealthStatus = "degraded"
	ModerationHealthUnavailable  ModerationHealthStatus = "unavailable"
)

// ModerationConfig 保存当前生效的图片审核配置。
// The credential reference is retained only so the runtime can resolve it.
type ModerationConfig struct {
	SingletonKey        string                 `json:"-" gorm:"column:singleton_key;type:varchar(32);primaryKey;not null;comment:单例记录键;"`       // 单例记录键
	ConfigVersion       int64                  `json:"configVersion" gorm:"column:config_version;type:bigint;not null;default:1;comment:配置版本;"` // 配置版本
	Provider            ModerationProviderType `json:"provider" gorm:"column:provider;type:varchar(20);not null;comment:审核供应商;"`                // 审核供应商
	Enabled             bool                   `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;comment:是否启用;"`           // 是否启用
	Region              string                 `json:"region" gorm:"column:region;type:varchar(64);not null;comment:服务区域;"`                     // 服务区域
	Endpoint            string                 `json:"endpoint" gorm:"column:endpoint;type:varchar(300);not null;comment:服务端点;"`                // 服务端点
	ServiceCode         string                 `json:"serviceCode" gorm:"column:service_code;type:varchar(64);not null;comment:服务编码;"`          // 服务编码
	TimeoutMS           int                    `json:"timeoutMs" gorm:"column:timeout_ms;type:int;not null;comment:超时时间（毫秒）;"`                  // 超时时间（毫秒）
	RetryCount          int                    `json:"retryCount" gorm:"column:retry_count;type:int;not null;comment:重试次数;"`                    // 重试次数
	RetryBackoffMS      int                    `json:"retryBackoffMs" gorm:"column:retry_backoff_ms;type:int;not null;comment:重试间隔（毫秒）;"`       // 重试间隔（毫秒）
	CredentialRef       string                 `json:"-" gorm:"column:credential_ref;type:text;not null;comment:密钥引用;"`                         // 密钥引用
	CredentialUpdatedAt *time.Time             `json:"-" gorm:"column:credential_updated_at;type:datetime;default:null;comment:凭证更新时间;"`        // 凭证更新时间
	ConfigHash          string                 `json:"configHash" gorm:"column:config_hash;type:varchar(64);not null;comment:配置摘要;"`            // 配置摘要
	AppliedByID         uint                   `json:"-" gorm:"column:applied_by_id;type:bigint unsigned;not null;comment:应用管理员ID;"`            // 应用管理员ID
	AppliedByUsername   string                 `json:"-" gorm:"column:applied_by_username;type:varchar(80);not null;comment:应用管理员用户名;"`         // 应用管理员用户名
	AppliedByNickname   *string                `json:"-" gorm:"column:applied_by_nickname;type:varchar(80);default:null;comment:应用管理员昵称;"`      // 应用管理员昵称
	AppliedAt           time.Time              `json:"appliedAt" gorm:"column:applied_at;type:datetime;not null;index;comment:应用时间;"`           // 应用时间
	Reason              string                 `json:"reason" gorm:"column:reason;type:varchar(200);not null;comment:原因;"`                      // 原因
}

// TableName 指定ModerationConfig对应的数据表名。
func (ModerationConfig) TableName() string {
	return "of_mod_config"
}

// ModerationConnectionTest 表示图片审核连接测试。
type ModerationConnectionTest struct {
	ID               string     `json:"testId" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`          // 主键ID
	ConfigVersion    int64      `json:"configVersion" gorm:"column:config_version;type:bigint;not null;index;comment:配置版本;"` // 配置版本
	ConfigHash       string     `json:"configHash" gorm:"column:config_hash;type:varchar(64);not null;comment:配置摘要;"`        // 配置摘要
	Success          bool       `json:"success" gorm:"column:success;type:tinyint(1) unsigned;not null;comment:是否成功;"`       // 是否成功
	Category         string     `json:"category" gorm:"column:category;type:varchar(40);not null;comment:分类关联数据;"`           // 分类关联数据
	RequestID        string     `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`   // 请求ID
	DurationMS       int        `json:"durationMs" gorm:"column:duration_ms;type:int;not null;comment:耗时（毫秒）;"`              // 耗时（毫秒）
	TestedByID       uint       `json:"-" gorm:"column:tested_by_id;type:bigint unsigned;not null;comment:测试管理员ID;"`         // 测试管理员ID
	TestedByUsername string     `json:"-" gorm:"column:tested_by_username;type:varchar(80);not null;comment:测试管理员用户名;"`      // 测试管理员用户名
	TestedByNickname *string    `json:"-" gorm:"column:tested_by_nickname;type:varchar(80);default:null;comment:测试管理员昵称;"`   // 测试管理员昵称
	TestedAt         time.Time  `json:"testedAt" gorm:"column:tested_at;type:datetime;not null;index;comment:测试时间;"`         // 测试时间
	ValidUntil       *time.Time `json:"validUntil" gorm:"column:valid_until;type:datetime;default:null;comment:有效截止时间;"`     // 有效截止时间
	SafeMessage      string     `json:"safeMessage" gorm:"column:safe_message;type:varchar(240);not null;comment:可安全展示的提示;"` // 可安全展示的提示
}

// TableName 指定ModerationConnectionTest对应的数据表名。
func (ModerationConnectionTest) TableName() string {
	return "of_mod_conn_tests"
}

// ModerationImageTest 表示图片审核样例测试。
type ModerationImageTest struct {
	ID                   string           `json:"testId" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                     // 主键ID
	ConfigVersion        int64            `json:"configVersion" gorm:"column:config_version;type:bigint;not null;index;comment:配置版本;"`                            // 配置版本
	MappedStatus         ModerationStatus `json:"mappedStatus" gorm:"column:mapped_status;type:varchar(20);not null;comment:映射后的审核状态;"`                           // 映射后的审核状态
	RiskLabelsJSON       datatypes.JSON   `json:"-" gorm:"column:risk_labels_json;type:json;not null;comment:风险标签JSON;"`                                          // 风险标签JSON
	RiskLevel            *string          `json:"riskLevel" gorm:"column:risk_level;type:varchar(40);default:null;comment:风险等级;"`                                 // 风险等级
	Category             string           `json:"category" gorm:"column:category;type:varchar(40);not null;comment:分类关联数据;"`                                      // 分类关联数据
	ProviderRequestID    *string          `json:"providerRequestId" gorm:"column:provider_request_id;type:varchar(128);default:null;comment:供应商请求ID;"`            // 供应商请求ID
	RequestID            string           `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                              // 请求ID
	DurationMS           int              `json:"durationMs" gorm:"column:duration_ms;type:int;not null;comment:耗时（毫秒）;"`                                         // 耗时（毫秒）
	SafeMessage          string           `json:"safeMessage" gorm:"column:safe_message;type:varchar(240);not null;comment:可安全展示的提示;"`                            // 可安全展示的提示
	ProviderSummaryJSON  datatypes.JSON   `json:"-" gorm:"column:provider_summary_json;type:json;default:null;comment:供应商响应摘要JSON;"`                              // 供应商响应摘要JSON
	TemporaryFileCleaned bool             `json:"temporaryFileCleaned" gorm:"column:temporary_file_cleaned;type:tinyint(1) unsigned;not null;comment:临时文件是否已清理;"` // 临时文件是否已清理
	TestedByID           uint             `json:"-" gorm:"column:tested_by_id;type:bigint unsigned;not null;comment:测试管理员ID;"`                                    // 测试管理员ID
	TestedAt             time.Time        `json:"testedAt" gorm:"column:tested_at;type:datetime;not null;index;comment:测试时间;"`                                    // 测试时间
}

// TableName 指定ModerationImageTest对应的数据表名。
func (ModerationImageTest) TableName() string {
	return "of_mod_image_tests"
}

// ModerationConfigHealth 表示图片审核运行状态。
type ModerationConfigHealth struct {
	ID                        uint                   `json:"-" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                           // 主键ID
	SingletonKey              string                 `json:"-" gorm:"column:singleton_key;type:varchar(32);not null;uniqueIndex;comment:单例记录键;"`                                // 单例记录键
	Status                    ModerationHealthStatus `json:"status" gorm:"column:status;type:varchar(20);not null;comment:状态;"`                                                 // 状态
	EffectiveConfigVersion    *int64                 `json:"effectiveConfigVersion" gorm:"column:effective_config_version;type:bigint;default:null;comment:当前生效配置版本;"`          // 当前生效配置版本
	LastConnectionTestID      *string                `json:"-" gorm:"column:last_connection_test_id;type:varchar(64);default:null;comment:最近连接测试ID;"`                           // 最近连接测试ID
	LastModerationSucceededAt *time.Time             `json:"lastModerationSucceededAt" gorm:"column:last_moderation_succeeded_at;type:datetime;default:null;comment:最近审核成功时间;"` // 最近审核成功时间
	LastFailureAt             *time.Time             `json:"lastFailureAt" gorm:"column:last_failure_at;type:datetime;default:null;comment:最近失败时间;"`                            // 最近失败时间
	LastFailureSafeSummary    *string                `json:"lastFailureSafeSummary" gorm:"column:last_failure_safe_summary;type:varchar(240);default:null;comment:最近失败安全摘要;"`   // 最近失败安全摘要
	ConsecutiveFailureCount   int                    `json:"consecutiveFailureCount" gorm:"column:consecutive_failure_count;type:int;not null;comment:连续失败次数;"`                 // 连续失败次数
	CreatedAt                 time.Time              `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                   // 创建时间
	UpdatedAt                 time.Time              `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                           // 更新时间
}

// TableName 指定ModerationConfigHealth对应的数据表名。
func (ModerationConfigHealth) TableName() string {
	return "of_mod_health"
}

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

// ModerationPersistenceModels is the shared AutoMigrate integration point.
func ModerationPersistenceModels() []interface{} {
	return []interface{}{
		&ModerationConfig{},
		&ModerationConnectionTest{},
		&ModerationImageTest{},
		&ModerationConfigHealth{},
		&ImageModerationRecord{},
	}
}

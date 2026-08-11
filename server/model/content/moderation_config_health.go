package content

import (
	"time"
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

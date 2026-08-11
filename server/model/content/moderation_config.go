package content

import (
	"time"
)

// ModerationProviderType 表示图片审核供应商类型。
type ModerationProviderType string

const (
	ModerationProviderAliyun ModerationProviderType = "aliyun"
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

package content

import (
	"time"
)

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

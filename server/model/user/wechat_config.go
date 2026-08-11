package user

import (
	"time"
)

// WeChatConfig 保存当前生效的微信小程序服务端配置。
type WeChatConfig struct {
	SingletonKey       string     `json:"-" gorm:"column:singleton_key;type:varchar(32);primaryKey;not null;comment:单例记录键;"`   // 单例记录键
	Version            int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:配置版本;"`          // 配置版本
	AppID              string     `json:"appId" gorm:"column:app_id;type:varchar(64);not null;comment:微信小程序AppID;"`            // 微信小程序AppID
	AppSecretEncrypted string     `json:"-" gorm:"column:app_secret_encrypted;type:text;not null;comment:加密后的微信小程序AppSecret;"` // 加密后的微信小程序AppSecret
	SecretUpdatedAt    *time.Time `json:"-" gorm:"column:secret_updated_at;type:datetime;default:null;comment:AppSecret更新时间;"` // AppSecret更新时间
	AppliedByID        uint       `json:"-" gorm:"column:applied_by_id;type:bigint unsigned;not null;comment:应用管理员ID;"`        // 应用管理员ID
	AppliedByUsername  string     `json:"-" gorm:"column:applied_by_username;type:varchar(80);not null;comment:应用管理员用户名;"`     // 应用管理员用户名
	AppliedByNickname  *string    `json:"-" gorm:"column:applied_by_nickname;type:varchar(80);default:null;comment:应用管理员昵称;"`  // 应用管理员昵称
	AppliedAt          time.Time  `json:"appliedAt" gorm:"column:applied_at;type:datetime;not null;index;comment:应用时间;"`       // 应用时间
	Reason             string     `json:"reason" gorm:"column:reason;type:varchar(200);not null;comment:修改原因;"`                // 修改原因
}

// TableName 指定WeChatConfig对应的数据表名。
func (WeChatConfig) TableName() string {
	return "of_wx_config"
}

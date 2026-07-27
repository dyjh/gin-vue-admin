package response

import "time"

// WeChatConfig 表示管理端可安全展示的微信小程序当前配置。
type WeChatConfig struct {
	Configured          bool                  `json:"configured"`          // 是否已完成配置
	AppID               string                `json:"appId"`               // 微信小程序AppID
	AppSecretConfigured bool                  `json:"appSecretConfigured"` // AppSecret是否已配置
	Version             int64                 `json:"version"`             // 配置版本，未配置时为0
	UpdatedBy           *AdministratorSummary `json:"updatedBy"`           // 最近更新管理员
	UpdatedAt           *time.Time            `json:"updatedAt"`           // 最近更新时间
	SecretUpdatedAt     *time.Time            `json:"secretUpdatedAt"`     // AppSecret最近更新时间
	Reason              string                `json:"reason"`              // 最近修改原因
}

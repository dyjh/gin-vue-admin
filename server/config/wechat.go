package config

type Wechat struct {
	Test      bool   `mapstructure:"test" json:"test" yaml:"test"`                   // 是否测试
	AppId     string `mapstructure:"app-id" json:"app-id" yaml:"app-id"`             // 小程序app_id
	AppSecret string `mapstructure:"app-secret" json:"app-secret" yaml:"app-secret"` // 小程序app_secret
}

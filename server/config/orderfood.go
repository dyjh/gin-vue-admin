package config

// OrderFood 表示来干饭业务模块的服务端安全配置。
type OrderFood struct {
	IdentityKey string `mapstructure:"identity-key" json:"-" yaml:"identity-key"` // 微信敏感字段加密密钥，必须是32字节密钥的标准Base64编码
}

package user

// ServiceGroup aggregates user-domain services.
type ServiceGroup struct {
	User         *UserService         // 小程序用户服务
	WeChatConfig *WeChatConfigService // 微信配置服务
}

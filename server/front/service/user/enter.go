package user

// ServiceGroup 聚合小程序用户相关服务。
type ServiceGroup struct {
	AuthService                 AuthService                 // 微信登录认证服务
	ProfileService              ProfileService              // 个人资料服务
	SubscriptionDeliveryService SubscriptionDeliveryService // 微信订阅消息投递服务
}

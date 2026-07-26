package service

var ServiceGroupApp = new(ServiceGroup)

// ServiceGroup 聚合当前模块的业务服务。
type ServiceGroup struct {
	AuthService                 // 微信登录服务
	RuntimeService              // 小程序运行时配置服务
	ProfileService              // 个人资料服务
	ContentService              // 菜品与菜谱服务
	UploadService               // 图片上传服务
	EngagementService           // 打卡、积分与通知服务
	MealService                 // 饭局与采购服务
	AssistService               // AI增强能力服务
	SubscriptionDeliveryService // 微信订阅消息投递服务
	PreferenceService           // 用户偏好聚合服务
}

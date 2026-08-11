package experience

// ServiceGroup 聚合小程序核心体验服务。
type ServiceGroup struct {
	RuntimeService    RuntimeService    // 运行时配置服务
	ContentService    ContentService    // 菜品与菜谱服务
	UploadService     UploadService     // 图片上传服务
	EngagementService EngagementService // 打卡、积分与通知服务
	MealService       MealService       // 饭局与采购服务
	AssistService     AssistService     // AI 增强能力服务
	PreferenceService PreferenceService // 用户偏好服务
}

var ServiceGroupApp = new(ServiceGroup)

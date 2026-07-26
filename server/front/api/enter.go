package api

import "github.com/dyjh/order-food-mini-app/server/front/service"

var ApiGroupApp = new(ApiGroup)

// ApiGroup 聚合当前模块的 API 处理器。
type ApiGroup struct {
	AuthApi       // 微信登录接口
	SystemApi     // 系统元数据接口
	ProfileApi    // 个人资料接口
	ContentApi    // 菜品与菜谱接口
	EngagementApi // 打卡、积分与通知接口
	MealApi       // 饭局与采购接口
	AssistApi     // AI增强能力接口
}

var (
	authService       = service.ServiceGroupApp.AuthService
	runtimeService    = service.ServiceGroupApp.RuntimeService
	profileService    = service.ServiceGroupApp.ProfileService
	contentService    = service.ServiceGroupApp.ContentService
	uploadService     = service.ServiceGroupApp.UploadService
	engagementService = service.ServiceGroupApp.EngagementService
	mealService       = service.ServiceGroupApp.MealService
	assistService     = service.ServiceGroupApp.AssistService
)

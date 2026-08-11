package system

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序系统与测试接口。
type ApiGroup struct {
	SystemApi // 小程序系统元数据接口
	TestApi   // 前台连通性测试接口
}

var (
	runtimeService    = service.ServiceGroupApp.ExperienceServiceGroup.RuntimeService
	contentService    = service.ServiceGroupApp.ExperienceServiceGroup.ContentService
	mealService       = service.ServiceGroupApp.ExperienceServiceGroup.MealService
	engagementService = service.ServiceGroupApp.ExperienceServiceGroup.EngagementService
)

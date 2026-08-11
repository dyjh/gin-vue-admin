package api

import (
	assistApi "github.com/dyjh/order-food-mini-app/server/front/api/assist"
	authApi "github.com/dyjh/order-food-mini-app/server/front/api/auth"
	contentApi "github.com/dyjh/order-food-mini-app/server/front/api/content"
	engagementApi "github.com/dyjh/order-food-mini-app/server/front/api/engagement"
	mealApi "github.com/dyjh/order-food-mini-app/server/front/api/meal"
	profileApi "github.com/dyjh/order-food-mini-app/server/front/api/profile"
	systemApi "github.com/dyjh/order-food-mini-app/server/front/api/system"
)

// ApiGroup 聚合小程序各功能域接口组。
type ApiGroup struct {
	AuthApiGroup       authApi.ApiGroup       // 认证接口组
	SystemApiGroup     systemApi.ApiGroup     // 系统接口组
	ProfileApiGroup    profileApi.ApiGroup    // 个人资料接口组
	ContentApiGroup    contentApi.ApiGroup    // 内容接口组
	EngagementApiGroup engagementApi.ApiGroup // 互动接口组
	MealApiGroup       mealApi.ApiGroup       // 饭局接口组
	AssistApiGroup     assistApi.ApiGroup     // 增强能力接口组
}

var ApiGroupApp = new(ApiGroup)

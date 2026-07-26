package router

import "github.com/dyjh/order-food-mini-app/server/front/api"

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	AuthRouter
	SystemRouter
	ProfileRouter
	ContentRouter
	EngagementRouter
	MealRouter
	AssistRouter
}

var (
	authApi       = api.ApiGroupApp.AuthApi
	systemApi     = api.ApiGroupApp.SystemApi
	profileApi    = api.ApiGroupApp.ProfileApi
	contentApi    = api.ApiGroupApp.ContentApi
	engagementApi = api.ApiGroupApp.EngagementApi
	mealApi       = api.ApiGroupApp.MealApi
	assistApi     = api.ApiGroupApp.AssistApi
)

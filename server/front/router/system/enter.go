package system

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	SystemRouter
	TestRouter
}

var (
	systemApi  = api.ApiGroupApp.SystemApiGroup.SystemApi
	testApi    = api.ApiGroupApp.SystemApiGroup.TestApi
	contentApi = api.ApiGroupApp.ContentApiGroup.ContentApi
)

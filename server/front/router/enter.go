package router

import "github.com/dyjh/order-food-mini-app/server/front/api"

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	TestRouter
}

var (
	testApi = api.ApiGroupApp.TestApi
)

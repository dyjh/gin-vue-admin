package api

import "github.com/dyjh/order-food-mini-app/server/front/service"

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	TestApi
}

var (
	testService = service.ServiceGroupApp.TestService
)

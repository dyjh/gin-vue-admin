package api

import "github.com/flipped-aurora/gin-vue-admin/server/front/service"

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	TestApi
}

var (
	testService = service.ServiceGroupApp.TestService
)

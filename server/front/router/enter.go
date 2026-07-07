package router

import "github.com/flipped-aurora/gin-vue-admin/server/front/api"

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	TestRouter
}

var (
	testApi = api.ApiGroupApp.TestApi
)

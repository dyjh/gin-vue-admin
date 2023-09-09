package router

import "github.com/flipped-aurora/gin-vue-admin/server/business/router/demo"

type RouterGroup struct {
	Demo demo.RouterGroup
}

var BusinessRouterGroupApp = new(RouterGroup)

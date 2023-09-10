package router

import "github.com/flipped-aurora/gin-vue-admin/server/business/router/demo"
import "github.com/flipped-aurora/gin-vue-admin/server/business/router/user"

type RouterGroup struct {
	Demo demo.RouterGroup
	User user.RouterGroup
}

var BusinessRouterGroupApp = new(RouterGroup)

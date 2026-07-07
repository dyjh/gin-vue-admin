package router

import (
	"github.com/dyjh/order-food-mini-app/server/router/example"
	"github.com/dyjh/order-food-mini-app/server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System  system.RouterGroup
	Example example.RouterGroup
}

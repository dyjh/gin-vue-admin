package initialize

import (
	_ "github.com/dyjh/order-food-mini-app/server/source"
	_ "github.com/dyjh/order-food-mini-app/server/source/example"
	_ "github.com/dyjh/order-food-mini-app/server/source/system"
)

func init() {
	// do nothing,only import source package so that inits can be registered
}

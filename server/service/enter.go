package service

import (
	"github.com/dyjh/order-food-mini-app/server/service/example"
	"github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/service/system"
)

var ServiceGroupApp = &ServiceGroup{
	OrderFoodServiceGroup: orderfood.ServiceGroupApp,
}

type ServiceGroup struct {
	SystemServiceGroup    system.ServiceGroup
	ExampleServiceGroup   example.ServiceGroup
	OrderFoodServiceGroup *orderfood.ServiceGroup
}

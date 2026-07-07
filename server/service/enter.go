package service

import (
	"github.com/dyjh/order-food-mini-app/server/service/example"
	"github.com/dyjh/order-food-mini-app/server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup
}

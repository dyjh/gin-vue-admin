package v1

import (
	"github.com/dyjh/order-food-mini-app/server/api/v1/example"
	"github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/dyjh/order-food-mini-app/server/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup    system.ApiGroup
	ExampleApiGroup   example.ApiGroup
	OrderFoodApiGroup orderfood.ApiGroup
}

package assist

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	AssistRouter
}

var assistApi = api.ApiGroupApp.AssistApiGroup.AssistApi

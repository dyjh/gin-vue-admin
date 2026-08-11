package ai

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	AIRouter
	AIUsageRouter
}

var (
	aiApi      = api.ApiGroupApp.AIApiGroup.AIApi
	aiUsageApi = api.ApiGroupApp.AIApiGroup.AIUsageApi
)

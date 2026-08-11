package engagement

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	PointsRouter
	SubscriptionRouter
}

var (
	pointsApi       = api.ApiGroupApp.EngagementApiGroup.PointsApi
	subscriptionApi = api.ApiGroupApp.EngagementApiGroup.SubscriptionApi
)

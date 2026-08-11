package engagement

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	EngagementRouter
}

var engagementApi = api.ApiGroupApp.EngagementApiGroup.EngagementApi

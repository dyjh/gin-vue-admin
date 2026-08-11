package engagement

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合积分与订阅消息接口。
type ApiGroup struct {
	PointsApi       // 积分管理接口
	SubscriptionApi // 订阅消息接口
}

var (
	pointsService       = service.ServiceGroupApp.EngagementServiceGroup.Points
	subscriptionService = service.ServiceGroupApp.EngagementServiceGroup.Subscription
)

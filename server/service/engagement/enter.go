package engagement

// ServiceGroup aggregates engagement-domain services.
type ServiceGroup struct {
	Points       *PointsService       // 积分管理服务
	Subscription *SubscriptionService // 订阅消息服务
}

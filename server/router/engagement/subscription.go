package engagement

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// SubscriptionRouter 提供管理端消息中心路由注册能力。
type SubscriptionRouter struct{}

// InitSubscriptionRouter 注册站内通知、固定订阅场景和发送记录路由。
func (router *SubscriptionRouter) InitSubscriptionRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/notifications", subscriptionApi.ListNotifications)
		group.GET("/notifications/:notificationId", subscriptionApi.GetNotification)
		group.GET("/subscribe-scenes", subscriptionApi.ListSubscribeScenes)
		group.GET("/subscribe-scenes/:scene", subscriptionApi.GetSubscribeScene)
		group.GET("/subscribe-logs", subscriptionApi.ListSubscribeLogs)
		group.GET("/subscribe-logs/:logId", subscriptionApi.GetSubscribeLog)
		mutationGroup.PUT("/subscribe-scenes/:scene", subscriptionApi.ConfigureSubscribeScene)
		mutationGroup.PUT("/subscribe-scenes/:scene/status", subscriptionApi.UpdateSubscribeSceneStatus)
	}
}

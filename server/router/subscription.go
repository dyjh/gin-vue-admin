package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// SubscriptionRouter 提供管理端消息中心路由注册能力。
type SubscriptionRouter struct{}

// InitSubscriptionRouter 注册站内通知、固定订阅场景和发送记录路由。
func (router *SubscriptionRouter) InitSubscriptionRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.SubscriptionApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/notifications", api.ListNotifications)
		group.GET("/notifications/:notificationId", api.GetNotification)
		group.GET("/subscribe-scenes", api.ListSubscribeScenes)
		group.GET("/subscribe-scenes/:scene", api.GetSubscribeScene)
		group.GET("/subscribe-logs", api.ListSubscribeLogs)
		group.GET("/subscribe-logs/:logId", api.GetSubscribeLog)
		mutationGroup.PUT("/subscribe-scenes/:scene", api.ConfigureSubscribeScene)
		mutationGroup.PUT("/subscribe-scenes/:scene/status", api.UpdateSubscribeSceneStatus)
	}
}

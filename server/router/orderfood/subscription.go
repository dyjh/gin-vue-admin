package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// SubscriptionRouter 提供管理端消息中心路由注册能力。
type SubscriptionRouter struct{}

// InitSubscriptionRouter 注册站内通知、订阅消息模板和发送记录路由。
func (router *SubscriptionRouter) InitSubscriptionRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.SubscriptionApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/notifications", api.ListNotifications)
		group.GET("/notifications/:notificationId", api.GetNotification)
		group.GET("/subscribe-templates", api.ListSubscribeTemplates)
		group.GET("/subscribe-templates/:templateId", api.GetSubscribeTemplate)
		group.GET("/subscribe-logs", api.ListSubscribeLogs)
		group.GET("/subscribe-logs/:logId", api.GetSubscribeLog)
		mutationGroup.POST("/subscribe-templates", api.CreateSubscribeTemplate)
		mutationGroup.PUT("/subscribe-templates/:templateId", api.UpdateSubscribeTemplate)
		mutationGroup.PUT("/subscribe-templates/:templateId/status", api.UpdateSubscribeTemplateStatus)
		mutationGroup.DELETE("/subscribe-templates/:templateId", api.DeleteSubscribeTemplate)
	}
}

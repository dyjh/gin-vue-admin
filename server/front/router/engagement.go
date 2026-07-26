package router

import "github.com/gin-gonic/gin"

type EngagementRouter struct{}

func (*EngagementRouter) InitEngagementRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/checkins/calendar", engagementApi.CheckinCalendar)
	privateGroup.GET("/checkins", engagementApi.Checkins)
	privateGroup.POST("/checkins", engagementApi.CreateCheckin)

	privateGroup.GET("/points/summary", engagementApi.PointsSummary)
	privateGroup.GET("/points/entries", engagementApi.PointEntries)
	privateGroup.GET("/feature-usages", engagementApi.FeatureUsages)

	privateGroup.GET("/notifications/summary", engagementApi.NotificationSummary)
	privateGroup.GET("/notifications", engagementApi.Notifications)
	privateGroup.POST("/notifications/:notificationId/read", engagementApi.ReadNotification)
	privateGroup.POST("/notifications/read-all", engagementApi.ReadAllNotifications)
}

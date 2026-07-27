package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

type AuditRouter struct{}

func (router *AuditRouter) InitAuditRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("/orderfood")
	api := orderfoodApi.ApiGroupApp.AuditApi
	{
		group.GET("/audit-logs", api.List)
		group.GET("/audit-logs/:auditLogId", api.Detail)
	}
}

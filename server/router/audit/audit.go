package audit

import "github.com/gin-gonic/gin"

type AuditRouter struct{}

func (router *AuditRouter) InitAuditRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/audit-logs", auditApi.List)
		group.GET("/audit-logs/:auditLogId", auditApi.Detail)
	}
}

package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// GovernanceRouter 提供违规处理记录路由注册能力。
type GovernanceRouter struct{}

// InitGovernanceRouter 注册违规处理记录和异步任务路由。
func (router *GovernanceRouter) InitGovernanceRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/governance-records", orderfoodApi.ApiGroupApp.GovernanceApi.ListGovernanceRecords)
		group.GET("/governance-records/:recordId", orderfoodApi.ApiGroupApp.GovernanceApi.GetGovernanceRecord)
		group.GET("/governance-jobs/:jobId", orderfoodApi.ApiGroupApp.GovernanceApi.GetGovernanceJob)
		mutationGroup.POST("/governance-jobs/:jobId/retry", orderfoodApi.ApiGroupApp.GovernanceApi.RetryGovernanceJob)
	}
}

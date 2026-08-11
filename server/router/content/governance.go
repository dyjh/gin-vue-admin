package content

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// GovernanceRouter 提供违规处理记录路由注册能力。
type GovernanceRouter struct{}

// InitGovernanceRouter 注册违规处理记录和异步任务路由。
func (router *GovernanceRouter) InitGovernanceRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/governance-records", governanceApi.ListGovernanceRecords)
		group.GET("/governance-records/:recordId", governanceApi.GetGovernanceRecord)
		group.GET("/governance-jobs/:jobId", governanceApi.GetGovernanceJob)
		mutationGroup.POST("/governance-jobs/:jobId/retry", governanceApi.RetryGovernanceJob)
	}
}

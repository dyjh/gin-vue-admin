package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/gin-gonic/gin"
)

// OperationsRouter 注册运营概览和AI调用记录路由。
type OperationsRouter struct{}

// InitOperationsRouter 初始化运营概览和AI调用记录路由。
func (router *OperationsRouter) InitOperationsRouter(
	privateGroup *gin.RouterGroup,
	dashboardApi *orderfoodApi.DashboardApi,
	aiUsageApi *orderfoodApi.AIUsageApi,
) {
	group := privateGroup.Group("/orderfood")
	{
		group.GET("/dashboard", dashboardApi.GetDashboard)
		group.GET("/ai-usages", aiUsageApi.ListAIUsages)
		group.GET("/ai-usages/:usageId", aiUsageApi.GetAIUsage)
		group.DELETE(
			"/ai-usages/:usageId/sensitive-content",
			aiUsageApi.DeleteAIUsageSensitiveContent,
		)
	}
}

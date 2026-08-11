package ai

import "github.com/gin-gonic/gin"

// AIUsageRouter 注册 AI 调用记录路由。
type AIUsageRouter struct{}

// InitAIUsageRouter 初始化 AI 调用记录路由。
func (*AIUsageRouter) InitAIUsageRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/ai-usages", aiUsageApi.ListAIUsages)
		group.GET("/ai-usages/:usageId", aiUsageApi.GetAIUsage)
		group.DELETE(
			"/ai-usages/:usageId/sensitive-content",
			aiUsageApi.DeleteAIUsageSensitiveContent,
		)
	}
}

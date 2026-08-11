package ai

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合 AI 平台与调用记录接口。
type ApiGroup struct {
	AIApi      // AI 平台管理接口
	AIUsageApi // AI 调用记录接口
}

var (
	aiService      = service.ServiceGroupApp.AIServiceGroup.AI
	aiUsageService = service.ServiceGroupApp.AIServiceGroup.AIUsage
)

package ai

// ServiceGroup aggregates AI configuration and usage services.
type ServiceGroup struct {
	AI      *AIService      // AI 平台配置服务
	AIUsage *AIUsageService // AI 调用记录服务
}

package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

type AIRouter struct{}

// InitAIRouter registers internal Gin paths under /orderfood. The web client
// sees the frozen external /api/orderfood prefix through the existing proxy.
func (router *AIRouter) InitAIRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.AIApi,
) {
	group := privateGroup.Group("/orderfood")
	{
		group.GET("/platform-capability-policy", api.GetPlatformPolicy)
		group.PUT("/platform-capability-policy", api.UpdatePlatformPolicy)
		group.PUT("/platform-capability-policy/emergency-status", api.SetPlatformEmergencyStatus)

		group.GET("/ai-providers", api.ListAIProviders)
		group.GET("/ai-providers/:providerId", api.GetAIProvider)
		group.POST("/ai-providers", api.CreateAIProvider)
		group.PUT("/ai-providers/:providerId", api.UpdateAIProvider)
		group.PUT("/ai-providers/:providerId/status", api.UpdateAIProviderStatus)
		group.DELETE("/ai-providers/:providerId", api.DeleteAIProvider)
		group.POST("/ai-providers/:providerId/connection-tests", api.TestAIProviderConnection)

		group.GET("/ai-models", api.ListAIModels)
		group.GET("/ai-models/:modelId", api.GetAIModel)
		group.POST("/ai-models", api.CreateAIModel)
		group.PUT("/ai-models/:modelId", api.UpdateAIModel)
		group.PUT("/ai-models/:modelId/status", api.UpdateAIModelStatus)
		group.DELETE("/ai-models/:modelId", api.DeleteAIModel)

		group.GET("/ai-capabilities", api.ListAICapabilities)
		group.GET("/ai-capabilities/:capabilityCode", api.GetAICapability)
		group.PUT("/ai-capabilities/:capabilityCode", api.UpdateAICapability)
		group.GET("/ai-capabilities/:capabilityCode/prompts", api.GetAIPromptWorkspace)
		group.PUT("/ai-capabilities/:capabilityCode/prompts", api.UpdateAIPrompt)
		group.POST("/ai-capabilities/:capabilityCode/prompt-validation", api.ValidateAIPrompt)
		group.POST("/ai-capabilities/:capabilityCode/prompt-render-preview", api.RenderAIPrompt)
		group.POST("/ai-capabilities/:capabilityCode/prompt-tests", api.TestAIPrompt)
	}
}

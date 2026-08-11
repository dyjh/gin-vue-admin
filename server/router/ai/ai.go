package ai

import "github.com/gin-gonic/gin"

type AIRouter struct{}

// InitAIRouter registers internal Gin paths under /orderfood. The web client
// sees the frozen external /api/orderfood prefix through the existing proxy.
func (router *AIRouter) InitAIRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/platform-capability-policy", aiApi.GetPlatformPolicy)
		group.PUT("/platform-capability-policy", aiApi.UpdatePlatformPolicy)
		group.PUT("/platform-capability-policy/emergency-status", aiApi.SetPlatformEmergencyStatus)

		group.GET("/ai-providers", aiApi.ListAIProviders)
		group.GET("/ai-providers/:providerId", aiApi.GetAIProvider)
		group.GET("/ai-providers/:providerId/models", aiApi.ListAIProviderModels)
		group.GET("/ai-model-provider-options", aiApi.ListAIModelProviders)
		group.POST("/ai-providers", aiApi.CreateAIProvider)
		group.PUT("/ai-providers/:providerId", aiApi.UpdateAIProvider)
		group.PUT("/ai-providers/:providerId/status", aiApi.UpdateAIProviderStatus)
		group.DELETE("/ai-providers/:providerId", aiApi.DeleteAIProvider)
		group.POST("/ai-providers/:providerId/connection-tests", aiApi.TestAIProviderConnection)

		group.GET("/ai-models", aiApi.ListAIModels)
		group.GET("/ai-models/:modelId", aiApi.GetAIModel)
		group.POST("/ai-models", aiApi.CreateAIModel)
		group.PUT("/ai-models/:modelId", aiApi.UpdateAIModel)
		group.PUT("/ai-models/:modelId/status", aiApi.UpdateAIModelStatus)
		group.DELETE("/ai-models/:modelId", aiApi.DeleteAIModel)

		group.GET("/ai-capabilities", aiApi.ListAICapabilities)
		group.GET("/ai-capabilities/:capabilityCode", aiApi.GetAICapability)
		group.PUT("/ai-capabilities/:capabilityCode", aiApi.UpdateAICapability)
		group.GET("/ai-capabilities/:capabilityCode/prompts", aiApi.GetAIPromptWorkspace)
		group.PUT("/ai-capabilities/:capabilityCode/prompts", aiApi.UpdateAIPrompt)
		group.POST("/ai-capabilities/:capabilityCode/prompt-validation", aiApi.ValidateAIPrompt)
		group.POST("/ai-capabilities/:capabilityCode/prompt-render-preview", aiApi.RenderAIPrompt)
		group.POST("/ai-capabilities/:capabilityCode/prompt-tests", aiApi.TestAIPrompt)
	}
}

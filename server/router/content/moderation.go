package content

import "github.com/gin-gonic/gin"

type ModerationRouter struct{}

// InitModerationRouter registers the internal Gin paths. The external web
// contract adds the /api proxy prefix, while Gin receives /orderfood.
func (router *ModerationRouter) InitModerationRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/moderation-config", moderationApi.GetConfig)
		group.PUT("/moderation-config", moderationApi.UpdateConfig)
		group.POST("/moderation-config/connection-tests", moderationApi.TestConnection)
		group.POST("/moderation-config/moderation-tests", moderationApi.TestImage)
	}
}

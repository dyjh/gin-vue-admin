package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

type ModerationRouter struct{}

// InitModerationRouter registers the internal Gin paths. The external web
// contract adds the /api proxy prefix, while Gin receives /orderfood.
func (router *ModerationRouter) InitModerationRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.ModerationApi,
) {
	group := privateGroup.Group("/orderfood")
	{
		group.GET("/moderation-config", api.GetConfig)
		group.PUT("/moderation-config", api.UpdateConfig)
		group.POST("/moderation-config/connection-tests", api.TestConnection)
		group.POST("/moderation-config/moderation-tests", api.TestImage)
	}
}

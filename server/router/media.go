package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

// MediaRouter 提供图片资源和图片审核记录路由注册能力。
type MediaRouter struct{}

// InitMediaRouter 注册图片资源和图片审核记录路由。
func (router *MediaRouter) InitMediaRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.MediaApi,
) {
	group := privateGroup.Group("/orderfood")
	{
		group.GET("/media", api.ListMedia)
		group.GET("/media/:fileId", api.GetMedia)
		group.GET("/moderation-records", api.ListModerationRecords)
		group.GET("/moderation-records/:recordId", api.GetModerationRecord)
	}
}

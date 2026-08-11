package content

import "github.com/gin-gonic/gin"

// MediaRouter 提供图片资源和图片审核记录路由注册能力。
type MediaRouter struct{}

// InitMediaRouter 注册图片资源和图片审核记录路由。
func (router *MediaRouter) InitMediaRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/media", mediaApi.ListMedia)
		group.GET("/media/:fileId", mediaApi.GetMedia)
		group.GET("/moderation-records", mediaApi.ListModerationRecords)
		group.GET("/moderation-records/:recordId", mediaApi.GetModerationRecord)
	}
}

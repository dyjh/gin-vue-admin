package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/gin-gonic/gin"
)

// WeChatConfigRouter 注册微信小程序配置路由。
type WeChatConfigRouter struct{}

// InitWeChatConfigRouter 初始化微信小程序配置路由。
func (router *WeChatConfigRouter) InitWeChatConfigRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.WeChatConfigApi,
) {
	group := privateGroup.Group("/orderfood")
	{
		group.GET("/wechat-config", api.GetConfig)
		group.PUT("/wechat-config", api.UpdateConfig)
	}
}

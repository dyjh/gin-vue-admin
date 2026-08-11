package user

import "github.com/gin-gonic/gin"

// WeChatConfigRouter 注册微信小程序配置路由。
type WeChatConfigRouter struct{}

// InitWeChatConfigRouter 初始化微信小程序配置路由。
func (router *WeChatConfigRouter) InitWeChatConfigRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/wechat-config", weChatConfigApi.GetConfig)
		group.PUT("/wechat-config", weChatConfigApi.UpdateConfig)
	}
}

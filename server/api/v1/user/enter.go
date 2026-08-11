package user

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合用户与微信配置接口。
type ApiGroup struct {
	UserApi         // 小程序用户接口
	WeChatConfigApi // 微信配置接口
}

var (
	userService         = service.ServiceGroupApp.UserServiceGroup.User
	weChatConfigService = service.ServiceGroupApp.UserServiceGroup.WeChatConfig
)

package user

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	UserRouter
	WeChatConfigRouter
}

var (
	userApi         = api.ApiGroupApp.UserApiGroup.UserApi
	weChatConfigApi = api.ApiGroupApp.UserApiGroup.WeChatConfigApi
)

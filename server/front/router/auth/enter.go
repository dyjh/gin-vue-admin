package auth

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	AuthRouter
}

var authApi = api.ApiGroupApp.AuthApiGroup.AuthApi

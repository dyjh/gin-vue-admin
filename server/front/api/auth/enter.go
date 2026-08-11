package auth

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序认证接口。
type ApiGroup struct {
	AuthApi // 小程序认证接口
}

var (
	authService    = service.ServiceGroupApp.UserServiceGroup.AuthService
	runtimeService = service.ServiceGroupApp.ExperienceServiceGroup.RuntimeService
)

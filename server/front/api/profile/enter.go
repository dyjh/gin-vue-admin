package profile

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序个人资料接口。
type ApiGroup struct {
	ProfileApi // 小程序个人资料接口
}

var (
	profileService = service.ServiceGroupApp.UserServiceGroup.ProfileService
	runtimeService = service.ServiceGroupApp.ExperienceServiceGroup.RuntimeService
)

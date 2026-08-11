package service

import (
	"github.com/dyjh/order-food-mini-app/server/front/service/experience"
	"github.com/dyjh/order-food-mini-app/server/front/service/system"
	"github.com/dyjh/order-food-mini-app/server/front/service/user"
)

// ServiceGroup 聚合小程序各服务域。
type ServiceGroup struct {
	UserServiceGroup       user.ServiceGroup       // 用户服务组
	ExperienceServiceGroup experience.ServiceGroup // 小程序体验服务组
	SystemServiceGroup     system.ServiceGroup     // 系统服务组
}

var ServiceGroupApp = new(ServiceGroup)

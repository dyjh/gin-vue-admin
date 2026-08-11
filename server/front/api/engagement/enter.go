package engagement

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序互动接口。
type ApiGroup struct {
	EngagementApi // 打卡、积分与通知接口
}

var engagementService = service.ServiceGroupApp.ExperienceServiceGroup.EngagementService

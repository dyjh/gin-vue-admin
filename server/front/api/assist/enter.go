package assist

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序增强能力接口。
type ApiGroup struct {
	AssistApi // 小程序增强能力接口
}

var assistService = service.ServiceGroupApp.ExperienceServiceGroup.AssistService

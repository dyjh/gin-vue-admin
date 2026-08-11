package meal

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序饭局接口。
type ApiGroup struct {
	MealApi // 饭局与采购清单接口
}

var mealService = service.ServiceGroupApp.ExperienceServiceGroup.MealService

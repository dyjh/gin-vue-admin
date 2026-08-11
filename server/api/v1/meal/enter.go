package meal

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合饭局与采购清单管理接口。
type ApiGroup struct {
	MealApi         // 饭局管理接口
	ShoppingListApi // 采购清单管理接口
}

var (
	mealService         = service.ServiceGroupApp.MealServiceGroup.Meal
	shoppingListService = service.ServiceGroupApp.MealServiceGroup.ShoppingList
)

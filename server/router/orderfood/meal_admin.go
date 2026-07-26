package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/gin-gonic/gin"
)

// MealAdminRouter 提供饭局和采购清单管理端路由注册能力。
type MealAdminRouter struct{}

// InitMealAdminRouter 注册饭局和采购清单管理端只读路由。
func (router *MealAdminRouter) InitMealAdminRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.MealAdminApi,
) {
	group := privateGroup.Group("/orderfood")
	{
		group.GET("/meals", api.ListMeals)
		group.GET("/meals/:mealId", api.GetMeal)
		group.GET("/shopping-lists", api.ListShoppingLists)
		group.GET("/shopping-lists/:shoppingListId", api.GetShoppingList)
	}
}

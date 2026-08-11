package meal

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	MealRouter
	ShoppingListRouter
}

var (
	mealApi         = api.ApiGroupApp.MealApiGroup.MealApi
	shoppingListApi = api.ApiGroupApp.MealApiGroup.ShoppingListApi
)

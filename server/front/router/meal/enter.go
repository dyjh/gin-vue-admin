package meal

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	MealRouter
}

var mealApi = api.ApiGroupApp.MealApiGroup.MealApi

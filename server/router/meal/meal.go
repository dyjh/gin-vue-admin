package meal

import "github.com/gin-gonic/gin"

// MealRouter 提供饭局管理端路由注册能力。
type MealRouter struct{}

// InitMealRouter 注册饭局管理端只读路由。
func (router *MealRouter) InitMealRouter(orderFoodGroup *gin.RouterGroup) {
	orderFoodGroup.GET("/meals", mealApi.ListMeals)
	orderFoodGroup.GET("/meals/:mealId", mealApi.GetMeal)
}

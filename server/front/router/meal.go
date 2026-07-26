package router

import "github.com/gin-gonic/gin"

type MealRouter struct{}

func (*MealRouter) InitMealRouter(
	privateGroup *gin.RouterGroup,
	publicGroup *gin.RouterGroup,
) {
	privateGroup.GET("/meals", mealApi.Meals)
	privateGroup.GET("/meals/current", mealApi.CurrentMeal)
	privateGroup.POST("/meals", mealApi.CreateMeal)
	privateGroup.GET("/meals/:mealId", mealApi.Meal)
	privateGroup.POST("/meals/lookup", mealApi.Lookup)
	privateGroup.POST("/meals/join", mealApi.Join)
	privateGroup.POST("/meals/:mealId/final-result-subscriptions", mealApi.RecordFinalResultSubscription)
	privateGroup.POST("/meals/:mealId/close", mealApi.Close)
	privateGroup.POST("/meals/:mealId/cancel", mealApi.Cancel)
	privateGroup.GET("/meals/:mealId/candidates", mealApi.Candidates)
	privateGroup.DELETE("/meals/:mealId/candidates/:candidateId", mealApi.RemoveCandidate)
	privateGroup.GET("/meals/:mealId/votes/me", mealApi.MyVotes)
	privateGroup.PUT("/meals/:mealId/votes/me", mealApi.SaveVotes)
	privateGroup.GET("/meals/:mealId/stats", mealApi.Stats)
	privateGroup.POST("/meals/:mealId/confirm", mealApi.Confirm)
	privateGroup.POST("/meals/:mealId/complete", mealApi.Complete)

	privateGroup.GET("/shopping-lists/current", mealApi.CurrentShopping)
	privateGroup.GET("/shopping-lists/:listId", mealApi.ShoppingByID)
	publicGroup.GET("/public/shopping-lists/:shareToken", mealApi.SharedShopping)
	privateGroup.POST("/shopping-lists/current/items", mealApi.CreateShoppingItem)
	privateGroup.PUT("/shopping-lists/current/items/:itemId", mealApi.UpdateShoppingItem)
	privateGroup.DELETE("/shopping-lists/current/items/:itemId", mealApi.DeleteShoppingItem)
	privateGroup.GET("/shopping-lists/:listId/export-text", mealApi.ExportShopping)
}

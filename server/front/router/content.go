package router

import "github.com/gin-gonic/gin"

type ContentRouter struct{}

func (*ContentRouter) InitContentRouter(privateGroup *gin.RouterGroup) {
	privateGroup.POST("/uploads/images", contentApi.UploadImage)

	privateGroup.GET("/dishes", contentApi.ListDishes)
	privateGroup.GET("/dishes/:dishId", contentApi.Dish)
	privateGroup.POST("/dishes", contentApi.CreateDish)
	privateGroup.PUT("/dishes/:dishId", contentApi.UpdateDish)
	privateGroup.DELETE("/dishes/:dishId", contentApi.DeleteDish)
	privateGroup.PUT("/dishes/:dishId/discoverability", contentApi.Discoverability)

	privateGroup.GET("/recommendations", contentApi.Recommendations)
	privateGroup.GET("/recommendations/:recommendationId", contentApi.Recommendation)
	privateGroup.POST("/recommendations/:recommendationId/copy", contentApi.CopyRecommendation)

	privateGroup.GET("/recipes", contentApi.Recipes)
	privateGroup.POST("/recipes", contentApi.CreateRecipe)
	privateGroup.GET("/recipes/:recipeId", contentApi.Recipe)
	privateGroup.PUT("/recipes/:recipeId", contentApi.UpdateRecipe)
	privateGroup.DELETE("/recipes/:recipeId", contentApi.DeleteRecipe)
	privateGroup.POST("/recipes/:recipeId/dishes", contentApi.AddRecipeDishes)
	privateGroup.DELETE("/recipes/:recipeId/dishes/:dishId", contentApi.RemoveRecipeDish)
}

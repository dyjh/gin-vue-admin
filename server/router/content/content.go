package content

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

type ContentRouter struct{}

func (router *ContentRouter) InitContentRouter(
	orderFoodGroup *gin.RouterGroup,
) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/user-dishes", contentApi.ListUserDishes)
		group.GET("/user-dishes/:dishId", contentApi.GetUserDishDetail)
		group.GET("/user-dishes/:dishId/references", contentApi.ListUserDishReferences)
		group.GET("/user-recipes", contentApi.ListUserRecipes)
		group.GET("/user-recipes/:recipeId", contentApi.GetUserRecipeDetail)
		group.POST("/governance-actions/preview", contentApi.PreviewGovernanceAction)
		mutationGroup.POST("/governance-actions", contentApi.ExecuteGovernanceAction)
	}
}

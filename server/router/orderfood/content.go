package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

type ContentRouter struct{}

func (router *ContentRouter) InitContentRouter(
	privateGroup *gin.RouterGroup,
) {
	RegisterContentRoutes(privateGroup, orderfoodApi.ApiGroupApp.ContentApi)
}

// RegisterContentRoutes keeps content-domain wiring independent from the
// aggregate router while P25/P26 are being integrated.
func RegisterContentRoutes(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.ContentApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/user-dishes", api.ListUserDishes)
		group.GET("/user-dishes/:dishId", api.GetUserDishDetail)
		group.GET("/user-dishes/:dishId/references", api.ListUserDishReferences)
		group.GET("/user-recipes", api.ListUserRecipes)
		group.GET("/user-recipes/:recipeId", api.GetUserRecipeDetail)
		group.POST("/governance-actions/preview", api.PreviewGovernanceAction)
		mutationGroup.POST("/governance-actions", api.ExecuteGovernanceAction)
	}
}

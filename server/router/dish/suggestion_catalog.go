package dish

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

type SuggestionCatalogRouter struct{}

func (router *SuggestionCatalogRouter) InitSuggestionCatalogRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup.Group("/suggestion-catalog")
	mutationGroup := group.Group("").
		Use(middleware.OperationRecord())
	{
		group.GET("", suggestionCatalogApi.GetWorkspace)
		group.GET("/ingredients", suggestionCatalogApi.ListIngredients)
		group.GET("/dishes", suggestionCatalogApi.ListDishes)
		mutationGroup.PUT("/policy", suggestionCatalogApi.UpdatePolicy)
		mutationGroup.POST("/ingredients", suggestionCatalogApi.CreateIngredient)
		mutationGroup.PUT("/ingredients/:id", suggestionCatalogApi.UpdateIngredient)
		mutationGroup.POST("/dishes", suggestionCatalogApi.CreateDish)
		mutationGroup.PUT("/dishes/:id", suggestionCatalogApi.UpdateDish)
	}
}

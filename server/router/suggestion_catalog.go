package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

type SuggestionCatalogRouter struct{}

func (router *SuggestionCatalogRouter) InitSuggestionCatalogRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.SuggestionCatalogApi,
) {
	group := privateGroup.Group("/orderfood/suggestion-catalog")
	mutationGroup := privateGroup.Group("/orderfood/suggestion-catalog").
		Use(middleware.OperationRecord())
	{
		group.GET("", api.GetWorkspace)
		group.GET("/ingredients", api.ListIngredients)
		group.GET("/dishes", api.ListDishes)
		mutationGroup.PUT("/policy", api.UpdatePolicy)
		mutationGroup.POST("/ingredients", api.CreateIngredient)
		mutationGroup.PUT("/ingredients/:id", api.UpdateIngredient)
		mutationGroup.POST("/dishes", api.CreateDish)
		mutationGroup.PUT("/dishes/:id", api.UpdateDish)
	}
}

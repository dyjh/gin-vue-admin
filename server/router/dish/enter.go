package dish

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	CatalogRouter
	RecommendationRouter
	OfficialDishRouter
	SuggestionCatalogRouter
}

var (
	catalogApi           = api.ApiGroupApp.DishApiGroup.CatalogApi
	recommendationApi    = api.ApiGroupApp.DishApiGroup.RecommendationApi
	officialDishApi      = api.ApiGroupApp.DishApiGroup.OfficialDishApi
	suggestionCatalogApi = api.ApiGroupApp.DishApiGroup.SuggestionCatalogApi
)

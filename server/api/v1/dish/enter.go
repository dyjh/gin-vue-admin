package dish

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合菜品目录、推荐与官方菜品接口。
type ApiGroup struct {
	CatalogApi           // 分类标签单位接口
	RecommendationApi    // 推荐管理接口
	OfficialDishApi      // 官方菜品接口
	SuggestionCatalogApi // 建议目录接口
}

var (
	catalogService           = service.ServiceGroupApp.DishServiceGroup.Catalog
	recommendationService    = service.ServiceGroupApp.DishServiceGroup.Recommendation
	officialDishService      = service.ServiceGroupApp.DishServiceGroup.OfficialDish
	suggestionCatalogService = service.ServiceGroupApp.DishServiceGroup.SuggestionCatalog
)

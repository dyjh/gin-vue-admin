package dish

// ServiceGroup aggregates dish-catalog and recommendation services.
type ServiceGroup struct {
	Catalog           *CatalogService           // 分类标签单位服务
	Recommendation    *RecommendationService    // 推荐管理服务
	OfficialDish      *OfficialDishService      // 官方菜品服务
	SuggestionCatalog *SuggestionCatalogService // 建议目录服务
}

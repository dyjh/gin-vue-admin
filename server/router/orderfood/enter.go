package orderfood

type RouterGroup struct {
	UserRouter
	AuditRouter
	ContentRouter
	PointsRouter
	CatalogRouter
	RecommendationRouter
	OfficialDishRouter
	MediaRouter
	GovernanceRouter
	MealAdminRouter
	OperationsRouter
	ModerationRouter
	AIRouter
	SuggestionCatalogRouter
	SubscriptionRouter
	WeChatConfigRouter
}

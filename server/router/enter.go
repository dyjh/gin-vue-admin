package router

import (
	"github.com/dyjh/order-food-mini-app/server/router/example"
	"github.com/dyjh/order-food-mini-app/server/router/system"
)

// RouterGroup aggregates system, example, and business routers.
type RouterGroup struct {
	System  system.RouterGroup
	Example example.RouterGroup

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

var RouterGroupApp = new(RouterGroup)

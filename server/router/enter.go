package router

import (
	aiRouter "github.com/dyjh/order-food-mini-app/server/router/ai"
	auditRouter "github.com/dyjh/order-food-mini-app/server/router/audit"
	contentRouter "github.com/dyjh/order-food-mini-app/server/router/content"
	dashboardRouter "github.com/dyjh/order-food-mini-app/server/router/dashboard"
	dishRouter "github.com/dyjh/order-food-mini-app/server/router/dish"
	engagementRouter "github.com/dyjh/order-food-mini-app/server/router/engagement"
	"github.com/dyjh/order-food-mini-app/server/router/example"
	mealRouter "github.com/dyjh/order-food-mini-app/server/router/meal"
	"github.com/dyjh/order-food-mini-app/server/router/system"
	userRouter "github.com/dyjh/order-food-mini-app/server/router/user"
)

// RouterGroup aggregates system, example, and business routers.
type RouterGroup struct {
	System     system.RouterGroup
	Example    example.RouterGroup
	User       userRouter.RouterGroup
	Engagement engagementRouter.RouterGroup
	Content    contentRouter.RouterGroup
	Dish       dishRouter.RouterGroup
	Meal       mealRouter.RouterGroup
	AI         aiRouter.RouterGroup
	Dashboard  dashboardRouter.RouterGroup
	Audit      auditRouter.RouterGroup
}

var RouterGroupApp = new(RouterGroup)

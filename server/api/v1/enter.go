package v1

import (
	aiApi "github.com/dyjh/order-food-mini-app/server/api/v1/ai"
	auditApi "github.com/dyjh/order-food-mini-app/server/api/v1/audit"
	contentApi "github.com/dyjh/order-food-mini-app/server/api/v1/content"
	dashboardApi "github.com/dyjh/order-food-mini-app/server/api/v1/dashboard"
	dishApi "github.com/dyjh/order-food-mini-app/server/api/v1/dish"
	engagementApi "github.com/dyjh/order-food-mini-app/server/api/v1/engagement"
	"github.com/dyjh/order-food-mini-app/server/api/v1/example"
	mealApi "github.com/dyjh/order-food-mini-app/server/api/v1/meal"
	"github.com/dyjh/order-food-mini-app/server/api/v1/system"
	userApi "github.com/dyjh/order-food-mini-app/server/api/v1/user"
)

// ApiGroup aggregates system, example, and business APIs.
type ApiGroup struct {
	SystemApiGroup     system.ApiGroup
	ExampleApiGroup    example.ApiGroup
	UserApiGroup       userApi.ApiGroup
	EngagementApiGroup engagementApi.ApiGroup
	ContentApiGroup    contentApi.ApiGroup
	DishApiGroup       dishApi.ApiGroup
	MealApiGroup       mealApi.ApiGroup
	AIApiGroup         aiApi.ApiGroup
	DashboardApiGroup  dashboardApi.ApiGroup
	AuditApiGroup      auditApi.ApiGroup
}

var ApiGroupApp = new(ApiGroup)

package service

import (
	"github.com/dyjh/order-food-mini-app/server/service/ai"
	"github.com/dyjh/order-food-mini-app/server/service/audit"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"github.com/dyjh/order-food-mini-app/server/service/content"
	"github.com/dyjh/order-food-mini-app/server/service/dashboard"
	"github.com/dyjh/order-food-mini-app/server/service/dish"
	"github.com/dyjh/order-food-mini-app/server/service/engagement"
	"github.com/dyjh/order-food-mini-app/server/service/example"
	"github.com/dyjh/order-food-mini-app/server/service/meal"
	"github.com/dyjh/order-food-mini-app/server/service/system"
	userService "github.com/dyjh/order-food-mini-app/server/service/user"
	"gorm.io/gorm"
)

// ServiceGroup aggregates system, example, and business services.
type ServiceGroup struct {
	SystemServiceGroup     system.ServiceGroup
	ExampleServiceGroup    example.ServiceGroup
	CommonServiceGroup     serviceCommon.ServiceGroup
	UserServiceGroup       userService.ServiceGroup
	EngagementServiceGroup engagement.ServiceGroup
	ContentServiceGroup    content.ServiceGroup
	DishServiceGroup       dish.ServiceGroup
	MealServiceGroup       meal.ServiceGroup
	AIServiceGroup         ai.ServiceGroup
	DashboardServiceGroup  dashboard.ServiceGroup
	AuditServiceGroup      audit.ServiceGroup
}

// NewServiceGroup constructs the business service aggregate.
func NewServiceGroup(db *gorm.DB) *ServiceGroup {
	accessAudit := serviceCommon.NewAccessAuditService(db)
	idempotency := serviceCommon.NewIdempotencyService(db)
	permission := serviceCommon.NewPermissionService(db)
	group := &ServiceGroup{
		CommonServiceGroup: serviceCommon.ServiceGroup{
			Permission: permission, AccessAudit: accessAudit, Idempotency: idempotency,
		},
		UserServiceGroup: userService.ServiceGroup{
			User: userService.NewUserService(db, accessAudit, idempotency, nil),
			WeChatConfig: userService.NewWeChatConfigService(
				db, permission, idempotency,
			),
		},
		EngagementServiceGroup: engagement.ServiceGroup{
			Points:       engagement.NewPointsService(db, permission, idempotency),
			Subscription: engagement.NewSubscriptionService(db, idempotency),
		},
		DishServiceGroup: dish.ServiceGroup{
			Catalog:           dish.NewCatalogService(db, idempotency),
			Recommendation:    dish.NewRecommendationService(db, idempotency),
			OfficialDish:      dish.NewOfficialDishService(db, idempotency),
			SuggestionCatalog: dish.NewSuggestionCatalogService(db, permission, idempotency),
		},
		MealServiceGroup: meal.ServiceGroup{
			Meal:         meal.NewMealService(db, permission),
			ShoppingList: meal.NewShoppingListService(db, permission),
		},
		AIServiceGroup: ai.ServiceGroup{
			AI:      ai.NewAIService(db, permission, idempotency, nil),
			AIUsage: ai.NewAIUsageService(db, permission, idempotency),
		},
		DashboardServiceGroup: dashboard.ServiceGroup{
			Dashboard: dashboard.NewDashboardService(db, permission),
		},
		AuditServiceGroup: audit.ServiceGroup{
			Audit: audit.NewAuditService(db),
		},
	}
	group.ContentServiceGroup = content.ServiceGroup{
		Content:    content.NewContentService(db, accessAudit, idempotency, permission, nil),
		Media:      content.NewMediaService(db),
		Governance: content.NewGovernanceService(db, permission, idempotency),
		Moderation: content.NewModerationService(
			db, nil, permission, accessAudit, idempotency,
		),
	}
	group.UserServiceGroup.User.Policy = group.AIServiceGroup.AI
	return group
}

var ServiceGroupApp = NewServiceGroup(nil)

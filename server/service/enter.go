package service

import (
	"github.com/dyjh/order-food-mini-app/server/service/example"
	"github.com/dyjh/order-food-mini-app/server/service/system"
	"gorm.io/gorm"
)

// ServiceGroup aggregates system, example, and business services.
type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup

	Permission        *PermissionService
	AccessAudit       *AccessAuditService
	Idempotency       *IdempotencyService
	User              *UserService
	Points            *PointsService
	Catalog           *CatalogService
	Recommendation    *RecommendationService
	OfficialDish      *OfficialDishService
	Media             *MediaService
	Governance        *GovernanceService
	MealAdmin         *MealAdminService
	Dashboard         *DashboardService
	AIUsage           *AIUsageService
	Audit             *AuditService
	Content           *ContentService
	AI                *AIService
	Moderation        *ModerationService
	SuggestionCatalog *SuggestionCatalogService
	Subscription      *SubscriptionService
	WeChatConfig      *WeChatConfigService
}

// NewServiceGroup constructs the business service aggregate.
func NewServiceGroup(db *gorm.DB) *ServiceGroup {
	accessAudit := NewAccessAuditService(db)
	idempotency := NewIdempotencyService(db)
	group := &ServiceGroup{
		Permission:     NewPermissionService(db),
		AccessAudit:    accessAudit,
		Idempotency:    idempotency,
		User:           NewUserService(db, accessAudit, idempotency, nil),
		Points:         NewPointsService(db, NewPermissionService(db), idempotency),
		Catalog:        NewCatalogService(db, idempotency),
		Recommendation: NewRecommendationService(db, idempotency),
		OfficialDish:   NewOfficialDishService(db, idempotency),
		Media:          NewMediaService(db),
		Governance:     NewGovernanceService(db, NewPermissionService(db), idempotency),
		MealAdmin:      NewMealAdminService(db, NewPermissionService(db)),
		Dashboard:      NewDashboardService(db, NewPermissionService(db)),
		AIUsage:        NewAIUsageService(db, NewPermissionService(db), idempotency),
		Audit:          NewAuditService(db),
	}
	group.Content = NewContentService(db, accessAudit, idempotency, group.Permission, nil)
	group.AI = NewAIService(db, group.Permission, idempotency, nil)
	group.Moderation = NewModerationService(
		db,
		nil,
		group.Permission,
		accessAudit,
		idempotency,
	)
	group.SuggestionCatalog = NewSuggestionCatalogService(
		db,
		group.Permission,
		idempotency,
	)
	group.Subscription = NewSubscriptionService(db, idempotency)
	group.WeChatConfig = NewWeChatConfigService(db, group.Permission, idempotency)
	group.User.Policy = group.AI
	return group
}

var ServiceGroupApp = NewServiceGroup(nil)

package orderfood

import "gorm.io/gorm"

// ServiceGroup 聚合当前模块的业务服务。
type ServiceGroup struct {
	Permission        *PermissionService        // 权限服务
	AccessAudit       *AccessAuditService       // 访问审计服务
	Idempotency       *IdempotencyService       // 幂等服务
	User              *UserService              // 用户服务
	Points            *PointsService            // 积分服务
	Catalog           *CatalogService           // 基础数据服务
	Recommendation    *RecommendationService    // 推荐精选服务
	OfficialDish      *OfficialDishService      // 官方菜品服务
	Media             *MediaService             // 图片资源服务
	Governance        *GovernanceService        // 违规处理记录服务
	MealAdmin         *MealAdminService         // 饭局和采购清单管理服务
	Dashboard         *DashboardService         // 运营概览服务
	AIUsage           *AIUsageService           // AI调用记录服务
	Audit             *AuditService             // 审计日志服务
	Content           *ContentService           // 用户内容服务
	AI                *AIService                // AI配置服务
	Moderation        *ModerationService        // 图片审核服务
	SuggestionCatalog *SuggestionCatalogService // 标准菜品索引服务
	Subscription      *SubscriptionService      // 订阅消息服务
	WeChatConfig      *WeChatConfigService      // 微信小程序配置服务
}

// NewServiceGroup 创建服务组合实例。
func NewServiceGroup(db *gorm.DB) *ServiceGroup {
	accessAudit := NewAccessAuditService(db)
	idempotency := NewIdempotencyService(db)
	group := &ServiceGroup{
		Permission:  NewPermissionService(db),
		AccessAudit: accessAudit,
		Idempotency: idempotency,
		User: NewUserService(
			db,
			accessAudit,
			idempotency,
			nil,
		),
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
	group.Content = NewContentService(
		db,
		accessAudit,
		idempotency,
		group.Permission,
		nil,
	)
	group.AI = NewAIService(
		db,
		group.Permission,
		idempotency,
		nil,
	)
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

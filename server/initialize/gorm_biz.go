package initialize

import (
	"context"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
)

// bizModel 创建来干饭业务表并注入各领域服务的默认配置。
func bizModel() error {
	db := global.GVA_DB

	// 基础领域模型在此集中注册；AI 与图片审核模型由各自领域暴露迁移清单。
	models := []interface{}{
		&orderfoodModel.AdminAccessAudit{},
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.MiniAppSession{},
		&orderfoodModel.MiniAppLoginCode{},
		&orderfoodModel.FrontIdempotencyRecord{},
		&orderfoodModel.FrontMediaAsset{},
		&orderfoodModel.FrontUserActivityDay{},
		&orderfoodModel.UserPreferenceProfile{},
		&orderfoodModel.PreferenceEvidenceAggregate{},
		&orderfoodModel.PreferenceEvidence{},
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.ContentTag{},
		&orderfoodModel.ContentUnit{},
		&orderfoodModel.UserDish{},
		&orderfoodModel.DishIngredient{},
		&orderfoodModel.DishStep{},
		&orderfoodModel.UserDishTag{},
		&orderfoodModel.UserRecipe{},
		&orderfoodModel.RecipeDish{},
		&orderfoodModel.DishReference{},
		&orderfoodModel.GovernanceRecord{},
		&orderfoodModel.GovernanceJob{},
		&orderfoodModel.GovernanceJobItem{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.UserNotification{},
		&orderfoodModel.PlatformRecommendation{},
		&orderfoodModel.RecommendationCopy{},
		&orderfoodModel.OfficialDish{},
		&orderfoodModel.OfficialDishIngredient{},
		&orderfoodModel.OfficialDishStep{},
		&orderfoodModel.OfficialDishTag{},
		&orderfoodModel.FrontCheckin{},
		&orderfoodModel.FrontMeal{},
		&orderfoodModel.MealParticipant{},
		&orderfoodModel.FrontMealCandidate{},
		&orderfoodModel.FrontMealVote{},
		&orderfoodModel.FrontMealFinalDish{},
		&orderfoodModel.SubscribeMessageTemplate{},
		&orderfoodModel.MealFinalResultSubscription{},
		&orderfoodModel.SubscribeMessageLog{},
		&orderfoodModel.SubscribeMessageAttempt{},
		&orderfoodModel.FrontShoppingList{},
		&orderfoodModel.FrontShoppingItem{},
		&orderfoodModel.FrontPointEntry{},
		&orderfoodModel.FrontFeatureUsage{},
		&orderfoodModel.FrontMealSuggestion{},
		&orderfoodModel.FrontMealSuggestionDish{},
		&orderfoodModel.FrontPrepPlan{},
		&orderfoodModel.PointRuleConfig{},
		&orderfoodModel.StandardIngredient{},
		&orderfoodModel.StandardDishIndex{},
		&orderfoodModel.StandardDishIngredient{},
		&orderfoodModel.SuggestionValidationPolicy{},
		&orderfoodModel.WeChatConfig{},
	}
	models = append(models, orderfoodService.AIModelsForMigration()...)
	models = append(models, orderfoodModel.ModerationPersistenceModels()...)

	// 必须先完成表结构创建，后续默认配置初始化才可以安全读写。
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}
	if orderfoodService.ServiceGroupApp == nil ||
		orderfoodService.ServiceGroupApp.Content == nil ||
		orderfoodService.ServiceGroupApp.AI == nil ||
		orderfoodService.ServiceGroupApp.OfficialDish == nil ||
		orderfoodService.ServiceGroupApp.MealAdmin == nil ||
		orderfoodService.ServiceGroupApp.Dashboard == nil ||
		orderfoodService.ServiceGroupApp.AIUsage == nil ||
		orderfoodService.ServiceGroupApp.SuggestionCatalog == nil ||
		orderfoodService.ServiceGroupApp.WeChatConfig == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	if err := orderfoodService.ServiceGroupApp.Content.BindDatabase(db); err != nil {
		return err
	}
	orderfoodService.ServiceGroupApp.AI.DB = db
	orderfoodService.ServiceGroupApp.Points.DB = db
	orderfoodService.ServiceGroupApp.Catalog.DB = db
	orderfoodService.ServiceGroupApp.Recommendation.DB = db
	orderfoodService.ServiceGroupApp.OfficialDish.DB = db
	orderfoodService.ServiceGroupApp.Media.DB = db
	orderfoodService.ServiceGroupApp.Governance.DB = db
	orderfoodService.ServiceGroupApp.MealAdmin.DB = db
	orderfoodService.ServiceGroupApp.Dashboard.DB = db
	orderfoodService.ServiceGroupApp.AIUsage.DB = db
	orderfoodService.ServiceGroupApp.AIUsage.Audit.DB = db
	orderfoodService.ServiceGroupApp.Moderation.DB = db
	orderfoodService.ServiceGroupApp.SuggestionCatalog.DB = db
	orderfoodService.ServiceGroupApp.WeChatConfig.DB = db

	// 默认配置按依赖顺序初始化，任一失败都阻止服务以不完整配置启动。
	if err := orderfoodService.ServiceGroupApp.Points.EnsureDefaults(context.Background()); err != nil {
		return err
	}
	if err := orderfoodService.ServiceGroupApp.AI.EnsureDefaults(context.Background()); err != nil {
		return err
	}
	if err := orderfoodService.ServiceGroupApp.SuggestionCatalog.EnsureDefaults(context.Background()); err != nil {
		return err
	}
	return ensureOrderFoodAdminSeed(db)
}

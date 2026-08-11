package initialize

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/ai"
)

// bizModel 仅注册并迁移来干饭业务模型。
func bizModel() error {
	db := global.GVA_DB

	// 基础领域模型在此集中注册；AI 与图片审核模型由各自领域暴露迁移清单。
	models := []interface{}{
		&commonModel.AdminAccessAudit{},
		&commonModel.AdminIdempotencyRecord{},
		&userModel.MiniAppUser{},
		&userModel.MiniAppSession{},
		&userModel.MiniAppLoginCode{},
		&commonModel.FrontIdempotencyRecord{},
		&contentModel.FrontMediaAsset{},
		&engagementModel.FrontUserActivityDay{},
		&userModel.UserPreferenceProfile{},
		&userModel.PreferenceEvidenceAggregate{},
		&userModel.PreferenceEvidence{},
		&dishModel.ContentCategory{},
		&dishModel.ContentTag{},
		&dishModel.ContentUnit{},
		&contentModel.UserDish{},
		&contentModel.DishIngredient{},
		&contentModel.DishStep{},
		&contentModel.UserDishTag{},
		&contentModel.UserRecipe{},
		&contentModel.RecipeDish{},
		&contentModel.DishReference{},
		&contentModel.GovernanceRecord{},
		&contentModel.GovernanceJob{},
		&contentModel.GovernanceJobItem{},
		&auditModel.AdminAuditLog{},
		&engagementModel.UserNotification{},
		&dishModel.PlatformRecommendation{},
		&dishModel.RecommendationCopy{},
		&dishModel.OfficialDish{},
		&dishModel.OfficialDishIngredient{},
		&dishModel.OfficialDishStep{},
		&dishModel.OfficialDishTag{},
		&engagementModel.FrontCheckin{},
		&mealModel.FrontMeal{},
		&mealModel.MealParticipant{},
		&mealModel.FrontMealCandidate{},
		&mealModel.FrontMealVote{},
		&mealModel.FrontMealFinalDish{},
		&engagementModel.SubscribeMessageTemplate{},
		&engagementModel.MealFinalResultSubscription{},
		&engagementModel.SubscribeMessageLog{},
		&engagementModel.SubscribeMessageAttempt{},
		&mealModel.FrontShoppingList{},
		&mealModel.FrontShoppingItem{},
		&engagementModel.FrontPointEntry{},
		&aiModel.FrontFeatureUsage{},
		&mealModel.FrontMealSuggestion{},
		&mealModel.FrontMealSuggestionDish{},
		&mealModel.FrontPrepPlan{},
		&engagementModel.PointRuleConfig{},
		&dishModel.StandardIngredient{},
		&dishModel.StandardDishIndex{},
		&dishModel.StandardDishIngredient{},
		&dishModel.SuggestionValidationPolicy{},
		&userModel.WeChatConfig{},
	}
	models = append(models, orderfoodService.AIModelsForMigration()...)
	models = append(models, contentModel.ModerationPersistenceModels()...)

	return db.AutoMigrate(models...)
}

package initialize

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
)

// bizModel 仅注册并迁移来干饭业务模型。
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

	return db.AutoMigrate(models...)
}

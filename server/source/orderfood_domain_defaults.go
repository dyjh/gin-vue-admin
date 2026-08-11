package source

import (
	"context"

	aiModel "github.com/dyjh/order-food-mini-app/server/model/ai"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	aiService "github.com/dyjh/order-food-mini-app/server/service/ai"
	initSystem "github.com/dyjh/order-food-mini-app/server/service/system"
	"gorm.io/gorm"
)

const initOrderFoodDomainDefaults = initSystem.InitOrderInternal + 15

type initOrderFoodDomainDefaultsData struct{}

func init() {
	initSystem.RegisterInit(initOrderFoodDomainDefaults, &initOrderFoodDomainDefaultsData{})
}

func (i *initOrderFoodDomainDefaultsData) InitializerName() string {
	return "orderfood_domain_defaults"
}

func (i *initOrderFoodDomainDefaultsData) MigrateTable(
	ctx context.Context,
) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, initSystem.ErrMissingDBContext
	}
	models := []interface{}{
		&engagementModel.PointRuleConfig{},
		&dishModel.StandardIngredient{},
		&dishModel.StandardDishIndex{},
		&dishModel.StandardDishIngredient{},
		&dishModel.SuggestionValidationPolicy{},
	}
	models = append(models, aiService.AIModelsForMigration()...)
	return ctx, db.AutoMigrate(models...)
}

func (i *initOrderFoodDomainDefaultsData) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	migrator := db.Migrator()
	return migrator.HasTable(&engagementModel.PointRuleConfig{}) &&
		migrator.HasTable(&aiModel.PlatformCapabilityPolicy{}) &&
		migrator.HasTable(&aiModel.AICapabilityDefinition{}) &&
		migrator.HasTable(&aiModel.AIPromptDefaultConfig{}) &&
		migrator.HasTable(&dishModel.SuggestionValidationPolicy{})
}

func (i *initOrderFoodDomainDefaultsData) InitializeData(
	ctx context.Context,
) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, initSystem.ErrMissingDBContext
	}
	group := orderfoodService.NewServiceGroup(db)
	if err := group.EngagementServiceGroup.Points.EnsureDefaults(ctx); err != nil {
		return ctx, err
	}
	if err := group.AIServiceGroup.AI.EnsureDefaults(ctx); err != nil {
		return ctx, err
	}
	if err := group.DishServiceGroup.SuggestionCatalog.EnsureDefaults(ctx); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (i *initOrderFoodDomainDefaultsData) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return tableHasRows(db, &engagementModel.PointRuleConfig{}) &&
		tableHasRows(db, &aiModel.PlatformCapabilityPolicy{}) &&
		tableHasRows(db, &aiModel.AICapabilityDefinition{}) &&
		tableHasRows(db, &aiModel.AIPromptDefaultConfig{}) &&
		tableHasRows(db, &dishModel.SuggestionValidationPolicy{})
}

func tableHasRows(db *gorm.DB, model interface{}) bool {
	var count int64
	return db.Model(model).Limit(1).Count(&count).Error == nil && count > 0
}

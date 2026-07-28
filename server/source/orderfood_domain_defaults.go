package source

import (
	"context"

	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
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
		&orderfoodModel.PointRuleConfig{},
		&orderfoodModel.StandardIngredient{},
		&orderfoodModel.StandardDishIndex{},
		&orderfoodModel.StandardDishIngredient{},
		&orderfoodModel.SuggestionValidationPolicy{},
	}
	models = append(models, orderfoodService.AIModelsForMigration()...)
	return ctx, db.AutoMigrate(models...)
}

func (i *initOrderFoodDomainDefaultsData) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	migrator := db.Migrator()
	return migrator.HasTable(&orderfoodModel.PointRuleConfig{}) &&
		migrator.HasTable(&orderfoodModel.PlatformCapabilityPolicy{}) &&
		migrator.HasTable(&orderfoodModel.AICapabilityDefinition{}) &&
		migrator.HasTable(&orderfoodModel.AIPromptDefaultConfig{}) &&
		migrator.HasTable(&orderfoodModel.SuggestionValidationPolicy{})
}

func (i *initOrderFoodDomainDefaultsData) InitializeData(
	ctx context.Context,
) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, initSystem.ErrMissingDBContext
	}
	group := orderfoodService.NewServiceGroup(db)
	if err := group.Points.EnsureDefaults(ctx); err != nil {
		return ctx, err
	}
	if err := group.AI.EnsureDefaults(ctx); err != nil {
		return ctx, err
	}
	if err := group.SuggestionCatalog.EnsureDefaults(ctx); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (i *initOrderFoodDomainDefaultsData) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return tableHasRows(db, &orderfoodModel.PointRuleConfig{}) &&
		tableHasRows(db, &orderfoodModel.PlatformCapabilityPolicy{}) &&
		tableHasRows(db, &orderfoodModel.AICapabilityDefinition{}) &&
		tableHasRows(db, &orderfoodModel.AIPromptDefaultConfig{}) &&
		tableHasRows(db, &orderfoodModel.SuggestionValidationPolicy{})
}

func tableHasRows(db *gorm.DB, model interface{}) bool {
	var count int64
	return db.Model(model).Limit(1).Count(&count).Error == nil && count > 0
}

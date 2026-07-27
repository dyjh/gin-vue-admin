package source

import (
	"context"

	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"github.com/dyjh/order-food-mini-app/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const initOrderFoodCatalog = system.InitOrderInternal + 10

type initOrderFoodCatalogData struct{}

func init() {
	system.RegisterInit(initOrderFoodCatalog, &initOrderFoodCatalogData{})
}

type orderFoodCatalogSeed struct {
	PublicID  string
	Name      string
	SortOrder int
}

var orderFoodCategorySeeds = []orderFoodCatalogSeed{
	{PublicID: "category-stir-fry", Name: "炒菜", SortOrder: 10},
	{PublicID: "category-steamed", Name: "蒸菜", SortOrder: 20},
	{PublicID: "category-cold-dish", Name: "凉菜", SortOrder: 30},
	{PublicID: "category-stewed", Name: "炖菜", SortOrder: 40},
	{PublicID: "category-braised", Name: "烧菜", SortOrder: 50},
	{PublicID: "category-fried", Name: "煎炸", SortOrder: 60},
	{PublicID: "category-roasted", Name: "烤菜", SortOrder: 70},
	{PublicID: "category-soup", Name: "汤羹", SortOrder: 80},
	{PublicID: "category-staple", Name: "主食", SortOrder: 90},
	{PublicID: "category-dessert", Name: "甜品", SortOrder: 100},
}

var orderFoodTagSeeds = []orderFoodCatalogSeed{
	{PublicID: "tag-cuisine-sichuan", Name: "川菜", SortOrder: 10},
	{PublicID: "tag-cuisine-hunan", Name: "湘菜", SortOrder: 20},
	{PublicID: "tag-cuisine-cantonese", Name: "粤菜", SortOrder: 30},
	{PublicID: "tag-cuisine-shandong", Name: "鲁菜", SortOrder: 40},
	{PublicID: "tag-cuisine-jiangsu", Name: "苏菜", SortOrder: 50},
	{PublicID: "tag-cuisine-zhejiang", Name: "浙菜", SortOrder: 60},
	{PublicID: "tag-cuisine-fujian", Name: "闽菜", SortOrder: 70},
	{PublicID: "tag-cuisine-anhui", Name: "徽菜", SortOrder: 80},
	{PublicID: "tag-cuisine-northeast", Name: "东北菜", SortOrder: 90},
	{PublicID: "tag-cuisine-northwest", Name: "西北菜", SortOrder: 100},
	{PublicID: "tag-cuisine-yunnan-guizhou", Name: "云贵菜", SortOrder: 110},
	{PublicID: "tag-cuisine-hakka", Name: "客家菜", SortOrder: 120},
	{PublicID: "tag-cuisine-chaoshan", Name: "潮汕菜", SortOrder: 130},
	{PublicID: "tag-cuisine-home-style", Name: "家常菜", SortOrder: 140},
	{PublicID: "tag-flavor-light", Name: "清淡", SortOrder: 150},
	{PublicID: "tag-flavor-savory", Name: "咸鲜", SortOrder: 160},
	{PublicID: "tag-flavor-sweet-sour", Name: "酸甜", SortOrder: 170},
	{PublicID: "tag-flavor-numbing-spicy", Name: "麻辣", SortOrder: 180},
	{PublicID: "tag-spice-none", Name: "不辣", SortOrder: 190},
	{PublicID: "tag-spice-mild", Name: "微辣", SortOrder: 200},
	{PublicID: "tag-spice-medium", Name: "中辣", SortOrder: 210},
	{PublicID: "tag-feature-quick", Name: "快手菜", SortOrder: 220},
	{PublicID: "tag-feature-rice-friendly", Name: "下饭菜", SortOrder: 230},
	{PublicID: "tag-feature-vegetarian", Name: "素菜", SortOrder: 240},
}

var orderFoodUnitSeeds = []orderFoodCatalogSeed{
	{PublicID: "unit-gram", Name: "克", SortOrder: 10},
	{PublicID: "unit-kilogram", Name: "千克", SortOrder: 20},
	{PublicID: "unit-milliliter", Name: "毫升", SortOrder: 30},
	{PublicID: "unit-liter", Name: "升", SortOrder: 40},
	{PublicID: "unit-piece", Name: "个", SortOrder: 50},
	{PublicID: "unit-animal", Name: "只", SortOrder: 60},
	{PublicID: "unit-egg", Name: "枚", SortOrder: 70},
	{PublicID: "unit-slice", Name: "片", SortOrder: 80},
	{PublicID: "unit-chunk", Name: "块", SortOrder: 90},
	{PublicID: "unit-strip", Name: "条", SortOrder: 100},
	{PublicID: "unit-root", Name: "根", SortOrder: 110},
	{PublicID: "unit-grain", Name: "颗", SortOrder: 120},
	{PublicID: "unit-clove", Name: "瓣", SortOrder: 130},
	{PublicID: "unit-plant", Name: "棵", SortOrder: 140},
	{PublicID: "unit-handful", Name: "把", SortOrder: 150},
	{PublicID: "unit-serving", Name: "份", SortOrder: 160},
	{PublicID: "unit-bowl", Name: "碗", SortOrder: 170},
	{PublicID: "unit-cup", Name: "杯", SortOrder: 180},
	{PublicID: "unit-tablespoon", Name: "汤匙", SortOrder: 190},
	{PublicID: "unit-teaspoon", Name: "茶匙", SortOrder: 200},
	{PublicID: "unit-pack", Name: "包", SortOrder: 210},
	{PublicID: "unit-box", Name: "盒", SortOrder: 220},
	{PublicID: "unit-pinch", Name: "少许", SortOrder: 230},
	{PublicID: "unit-as-needed", Name: "适量", SortOrder: 240},
}

func (i *initOrderFoodCatalogData) InitializerName() string {
	return "orderfood_catalog_defaults"
}

func (i *initOrderFoodCatalogData) MigrateTable(
	ctx context.Context,
) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.ContentTag{},
		&orderfoodModel.ContentUnit{},
	)
}

func (i *initOrderFoodCatalogData) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	migrator := db.Migrator()
	return migrator.HasTable(&orderfoodModel.ContentCategory{}) &&
		migrator.HasTable(&orderfoodModel.ContentTag{}) &&
		migrator.HasTable(&orderfoodModel.ContentUnit{})
}

func (i *initOrderFoodCatalogData) InitializeData(
	ctx context.Context,
) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	if err := initializeOrderFoodCatalogData(db); err != nil {
		return ctx, errors.Wrap(err, i.InitializerName()+" data initialization failed")
	}
	return ctx, nil
}

func (i *initOrderFoodCatalogData) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return orderFoodCatalogDataExists(db)
}

func initializeOrderFoodCatalogData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, seed := range orderFoodCategorySeeds {
			item := &orderfoodModel.ContentCategory{
				PublicID:  seed.PublicID,
				Name:      seed.Name,
				SortOrder: seed.SortOrder,
				Enabled:   true,
				Version:   1,
			}
			if err := insertOrderFoodCatalogSeedIfMissing(tx, seed, item); err != nil {
				return err
			}
		}
		for _, seed := range orderFoodTagSeeds {
			item := &orderfoodModel.ContentTag{
				PublicID:  seed.PublicID,
				Name:      seed.Name,
				SortOrder: seed.SortOrder,
				Enabled:   true,
				Version:   1,
			}
			if err := insertOrderFoodCatalogSeedIfMissing(tx, seed, item); err != nil {
				return err
			}
		}
		for _, seed := range orderFoodUnitSeeds {
			item := &orderfoodModel.ContentUnit{
				PublicID:  seed.PublicID,
				Name:      seed.Name,
				SortOrder: seed.SortOrder,
				Enabled:   true,
				Version:   1,
			}
			if err := insertOrderFoodCatalogSeedIfMissing(tx, seed, item); err != nil {
				return err
			}
		}
		return nil
	})
}

func insertOrderFoodCatalogSeedIfMissing(
	db *gorm.DB,
	seed orderFoodCatalogSeed,
	item interface{},
) error {
	var count int64
	if err := db.Unscoped().
		Model(item).
		Where("public_id = ? OR name = ?", seed.PublicID, seed.Name).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(item).Error
}

func orderFoodCatalogDataExists(db *gorm.DB) bool {
	seedSets := []struct {
		table string
		seeds []orderFoodCatalogSeed
	}{
		{table: (orderfoodModel.ContentCategory{}).TableName(), seeds: orderFoodCategorySeeds},
		{table: (orderfoodModel.ContentTag{}).TableName(), seeds: orderFoodTagSeeds},
		{table: (orderfoodModel.ContentUnit{}).TableName(), seeds: orderFoodUnitSeeds},
	}
	for _, seedSet := range seedSets {
		for _, seed := range seedSet.seeds {
			var count int64
			if err := db.Unscoped().
				Table(seedSet.table).
				Where("public_id = ? OR name = ?", seed.PublicID, seed.Name).
				Count(&count).Error; err != nil || count == 0 {
				return false
			}
		}
	}
	return true
}

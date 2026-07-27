package source

import (
	"context"
	"fmt"
	"testing"

	"github.com/dyjh/order-food-mini-app/server/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestOrderFoodCatalogSeedDefinitionsAreStable(t *testing.T) {
	testCases := []struct {
		name  string
		seeds []orderFoodCatalogSeed
	}{
		{name: "categories", seeds: orderFoodCategorySeeds},
		{name: "tags", seeds: orderFoodTagSeeds},
		{name: "units", seeds: orderFoodUnitSeeds},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			publicIDs := make(map[string]struct{}, len(testCase.seeds))
			names := make(map[string]struct{}, len(testCase.seeds))
			lastSortOrder := 0
			for _, seed := range testCase.seeds {
				if seed.PublicID == "" || seed.Name == "" {
					t.Fatalf("seed must have public ID and name: %+v", seed)
				}
				if _, exists := publicIDs[seed.PublicID]; exists {
					t.Fatalf("duplicate public ID %q", seed.PublicID)
				}
				if _, exists := names[seed.Name]; exists {
					t.Fatalf("duplicate name %q", seed.Name)
				}
				if seed.SortOrder <= lastSortOrder {
					t.Fatalf("sort order must be strictly increasing: %+v", seed)
				}
				publicIDs[seed.PublicID] = struct{}{}
				names[seed.Name] = struct{}{}
				lastSortOrder = seed.SortOrder
			}
		})
	}
}

func TestEnsureOrderFoodCatalogSeedIsIdempotentAndPreservesChanges(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get SQLite connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	for _, tableName := range []string{
		(model.ContentCategory{}).TableName(),
		(model.ContentTag{}).TableName(),
		(model.ContentUnit{}).TableName(),
	} {
		statement := fmt.Sprintf(`CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			public_id TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL UNIQUE,
			sort_order INTEGER NOT NULL DEFAULT 1,
			enabled INTEGER NOT NULL DEFAULT 1,
			version INTEGER NOT NULL DEFAULT 1
		)`, tableName)
		if err = db.Exec(statement).Error; err != nil {
			t.Fatalf("create catalog table %s: %v", tableName, err)
		}
	}
	initializer := &initOrderFoodCatalogData{}
	ctx := context.WithValue(context.Background(), "db", db)
	if initializer.DataInserted(ctx) {
		t.Fatal("catalog defaults unexpectedly exist before initialization")
	}
	if _, err = initializer.InitializeData(ctx); err != nil {
		t.Fatalf("seed catalog defaults: %v", err)
	}
	if !initializer.DataInserted(ctx) {
		t.Fatal("catalog initializer did not report complete data")
	}
	assertCatalogSeedCount(t, db, &model.ContentCategory{}, len(orderFoodCategorySeeds))
	assertCatalogSeedCount(t, db, &model.ContentTag{}, len(orderFoodTagSeeds))
	assertCatalogSeedCount(t, db, &model.ContentUnit{}, len(orderFoodUnitSeeds))

	if err = db.Model(&model.ContentCategory{}).
		Where("public_id = ?", orderFoodCategorySeeds[0].PublicID).
		Updates(map[string]interface{}{
			"name":       "自定义分类",
			"sort_order": 999,
			"enabled":    false,
		}).Error; err != nil {
		t.Fatalf("customize seeded category: %v", err)
	}
	if err = db.Where("public_id = ?", orderFoodTagSeeds[0].PublicID).
		Delete(&model.ContentTag{}).Error; err != nil {
		t.Fatalf("soft-delete seeded tag: %v", err)
	}
	if _, err = initializer.InitializeData(ctx); err != nil {
		t.Fatalf("reseed catalog defaults: %v", err)
	}

	var category model.ContentCategory
	if err = db.Where("public_id = ?", orderFoodCategorySeeds[0].PublicID).
		First(&category).Error; err != nil {
		t.Fatalf("load customized category: %v", err)
	}
	if category.Name != "自定义分类" || category.SortOrder != 999 || category.Enabled {
		t.Fatalf("catalog seed overwrote administrator changes: %+v", category)
	}
	var deletedTag model.ContentTag
	if err = db.Unscoped().
		Where("public_id = ?", orderFoodTagSeeds[0].PublicID).
		First(&deletedTag).Error; err != nil {
		t.Fatalf("load soft-deleted tag: %v", err)
	}
	if !deletedTag.DeletedAt.Valid {
		t.Fatal("catalog seed restored a soft-deleted tag")
	}
	if !initializer.DataInserted(ctx) {
		t.Fatal("soft-deleted catalog data should remain recognized as initialized")
	}
	assertCatalogSeedCount(t, db.Unscoped(), &model.ContentCategory{}, len(orderFoodCategorySeeds))
	assertCatalogSeedCount(t, db.Unscoped(), &model.ContentTag{}, len(orderFoodTagSeeds))
	assertCatalogSeedCount(t, db.Unscoped(), &model.ContentUnit{}, len(orderFoodUnitSeeds))
}

func assertCatalogSeedCount(
	t *testing.T,
	db *gorm.DB,
	value interface{},
	expected int,
) {
	t.Helper()
	var count int64
	if err := db.Model(value).Count(&count).Error; err != nil {
		t.Fatalf("count catalog seeds: %v", err)
	}
	if count != int64(expected) {
		t.Fatalf("catalog seed count = %d, want %d", count, expected)
	}
}

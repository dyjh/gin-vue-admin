package service

import (
	"context"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// catalogUserDishReference 表示分类引用计数测试使用的用户菜品记录。
type catalogUserDishReference struct {
	ID         uint `gorm:"column:id;primaryKey;autoIncrement"` // 内部主键
	CategoryID uint `gorm:"column:category_id;index;not null"`  // 分类主键
}

// TableName 指定分类引用测试使用的用户菜品表。
func (catalogUserDishReference) TableName() string {
	return (orderfoodModel.UserDish{}).TableName()
}

// catalogStandardDishReference 表示分类引用计数测试使用的标准菜品记录。
type catalogStandardDishReference struct {
	ID         string `gorm:"column:id;type:varchar(64);primaryKey"` // 标准菜品ID
	CategoryID uint   `gorm:"column:category_id;index;not null"`     // 分类主键
}

// TableName 指定分类引用测试使用的标准菜品表。
func (catalogStandardDishReference) TableName() string {
	return (orderfoodModel.StandardDishIndex{}).TableName()
}

// catalogOfficialDishReference 表示分类引用计数测试使用的官方菜品记录。
type catalogOfficialDishReference struct {
	ID         uint `gorm:"column:id;primaryKey;autoIncrement"` // 内部主键
	CategoryID uint `gorm:"column:category_id;index;not null"`  // 分类主键
}

// TableName 指定分类引用测试使用的官方菜品表。
func (catalogOfficialDishReference) TableName() string {
	return (orderfoodModel.OfficialDish{}).TableName()
}

// catalogDishIngredientReference 表示单位引用计数测试使用的用户菜品配料记录。
type catalogDishIngredientReference struct {
	ID     uint `gorm:"column:id;primaryKey;autoIncrement"` // 内部主键
	UnitID uint `gorm:"column:unit_id;index;not null"`      // 单位主键
}

// TableName 指定单位引用测试使用的用户菜品配料表。
func (catalogDishIngredientReference) TableName() string {
	return (orderfoodModel.DishIngredient{}).TableName()
}

// catalogOfficialIngredientReference 表示单位引用计数测试使用的官方菜品配料记录。
type catalogOfficialIngredientReference struct {
	ID     uint `gorm:"column:id;primaryKey;autoIncrement"` // 内部主键
	UnitID uint `gorm:"column:unit_id;index;not null"`      // 单位主键
}

// TableName 指定单位引用测试使用的官方菜品配料表。
func (catalogOfficialIngredientReference) TableName() string {
	return (orderfoodModel.OfficialDishIngredient{}).TableName()
}

// newCatalogTestService 创建基础数据服务及其独立MySQL测试数据库。
func newCatalogTestService(t *testing.T) (*CatalogService, *gorm.DB) {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.ContentTag{},
		&orderfoodModel.ContentUnit{},
		&orderfoodModel.UserDishTag{},
		&orderfoodModel.OfficialDishTag{},
		&catalogUserDishReference{},
		&catalogStandardDishReference{},
		&catalogOfficialDishReference{},
		&catalogDishIngredientReference{},
		&catalogOfficialIngredientReference{},
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAuditLog{},
	); err != nil {
		t.Fatal(err)
	}
	service := NewCatalogService(db, NewIdempotencyService(db))
	service.Now = func() time.Time {
		return time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC)
	}
	return service, db
}

// TestCatalogListsAllResourceReferencesAndValidatesQuery 校验三类基础数据的引用汇总与服务层查询边界。
func TestCatalogListsAllResourceReferencesAndValidatesQuery(t *testing.T) {
	service, db := newCatalogTestService(t)
	ctx := context.Background()
	enabled := true
	actor := orderfoodRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-catalog-list",
	}
	category, _, err := service.CreateCategory(
		ctx, actor, "catalog-category-list-create",
		orderfoodRequest.CatalogItemCreateInput{Name: "热菜", SortOrder: 1, Enabled: &enabled},
	)
	if err != nil {
		t.Fatal(err)
	}
	tag, _, err := service.CreateTag(
		ctx, actor, "catalog-tag-list-create",
		orderfoodRequest.CatalogItemCreateInput{Name: "下饭", SortOrder: 1, Enabled: &enabled},
	)
	if err != nil {
		t.Fatal(err)
	}
	unit, _, err := service.CreateUnit(
		ctx, actor, "catalog-unit-list-create",
		orderfoodRequest.CatalogItemCreateInput{Name: "克", SortOrder: 1, Enabled: &enabled},
	)
	if err != nil {
		t.Fatal(err)
	}

	var categoryRow orderfoodModel.ContentCategory
	var tagRow orderfoodModel.ContentTag
	var unitRow orderfoodModel.ContentUnit
	if err := db.First(&categoryRow, "public_id = ?", category.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&tagRow, "public_id = ?", tag.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&unitRow, "public_id = ?", unit.ID).Error; err != nil {
		t.Fatal(err)
	}
	references := []interface{}{
		&catalogUserDishReference{CategoryID: categoryRow.ID},
		&catalogStandardDishReference{ID: "standard-dish-1", CategoryID: categoryRow.ID},
		&catalogOfficialDishReference{CategoryID: categoryRow.ID},
		&orderfoodModel.UserDishTag{DishID: 1, TagID: tagRow.ID, CreatedAt: service.now()},
		&orderfoodModel.OfficialDishTag{DishID: 1, TagID: tagRow.ID, CreatedAt: service.now()},
		&catalogDishIngredientReference{UnitID: unitRow.ID},
		&catalogOfficialIngredientReference{UnitID: unitRow.ID},
	}
	for _, reference := range references {
		if err := db.Create(reference).Error; err != nil {
			t.Fatal(err)
		}
	}

	categoryPage, err := service.ListCategories(ctx, orderfoodRequest.CatalogItemListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	tagPage, err := service.ListTags(ctx, orderfoodRequest.CatalogItemListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	unitPage, err := service.ListUnits(ctx, orderfoodRequest.CatalogItemListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(categoryPage.List) != 1 || categoryPage.List[0].ReferenceCount != 3 {
		t.Fatalf("category page = %+v", categoryPage)
	}
	if len(tagPage.List) != 1 || tagPage.List[0].ReferenceCount != 2 {
		t.Fatalf("tag page = %+v", tagPage)
	}
	if len(unitPage.List) != 1 || unitPage.List[0].ReferenceCount != 2 {
		t.Fatalf("unit page = %+v", unitPage)
	}

	_, err = service.ListCategories(ctx, orderfoodRequest.CatalogItemListQuery{
		SortOrder: "asc; drop table of_categories",
	})
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe sort order error = %v", err)
	}
	_, err = service.ListCategories(ctx, orderfoodRequest.CatalogItemListQuery{PageSize: 21})
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("invalid page size error = %v", err)
	}
}

// TestCatalogTagLifecycleAndReferenceProtection 校验标签的幂等变更、版本排序和引用删除保护。
func TestCatalogTagLifecycleAndReferenceProtection(t *testing.T) {
	service, db := newCatalogTestService(t)
	ctx := context.Background()
	enabled := true
	actor := orderfoodRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-catalog-tag",
	}
	created, replayed, err := service.CreateTag(
		ctx,
		actor,
		"catalog-tag-create",
		orderfoodRequest.CatalogItemCreateInput{
			Name: "  家常菜  ", SortOrder: 10, Enabled: &enabled,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || created.Name != "家常菜" || created.Version != 1 {
		t.Fatalf("created tag = %+v, replayed = %v", created, replayed)
	}
	replayedTag, replayed, err := service.CreateTag(
		ctx,
		actor,
		"catalog-tag-create",
		orderfoodRequest.CatalogItemCreateInput{
			Name: "家常菜", SortOrder: 10, Enabled: &enabled,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !replayed || replayedTag.ID != created.ID {
		t.Fatalf("replayed tag = %+v, replayed = %v", replayedTag, replayed)
	}
	_, _, err = service.CreateTag(
		ctx,
		actor,
		"catalog-tag-create-duplicate",
		orderfoodRequest.CatalogItemCreateInput{
			Name: "家常菜", SortOrder: 11, Enabled: &enabled,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminAlreadyExists {
		t.Fatalf("duplicate tag error = %v", err)
	}

	disabled := false
	updated, replayed, err := service.UpdateTag(
		ctx,
		created.ID,
		actor,
		"catalog-tag-update",
		orderfoodRequest.CatalogItemUpdateInput{
			Name: "家常", SortOrder: 20, Enabled: &disabled, ExpectedVersion: 1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || updated.Enabled || updated.Version != 2 || updated.Name != "家常" {
		t.Fatalf("updated tag = %+v, replayed = %v", updated, replayed)
	}

	sorted, replayed, err := service.UpdateTagSortOrder(
		ctx,
		actor,
		"catalog-tag-sort",
		orderfoodRequest.CatalogSortOrderInput{
			Items: []orderfoodRequest.CatalogSortOrderItem{
				{ID: created.ID, SortOrder: 1, ExpectedVersion: 2},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || sorted.Updated != 1 {
		t.Fatalf("sorted result = %+v, replayed = %v", sorted, replayed)
	}

	var stored orderfoodModel.ContentTag
	if err := db.First(&stored, "public_id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&orderfoodModel.UserDishTag{
		DishID: 999, TagID: stored.ID, CreatedAt: service.now(),
	}).Error; err != nil {
		t.Fatal(err)
	}
	_, _, err = service.DeleteTag(
		ctx,
		created.ID,
		actor,
		"catalog-tag-delete",
		orderfoodRequest.CatalogDeleteInput{
			Reason: "清理测试标签", ExpectedVersion: 3,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminResourceInUse {
		t.Fatalf("delete referenced tag error = %v", err)
	}
	if err := db.Where("tag_id = ?", stored.ID).
		Delete(&orderfoodModel.UserDishTag{}).Error; err != nil {
		t.Fatal(err)
	}
	deleted, replayed, err := service.DeleteTag(
		ctx,
		created.ID,
		actor,
		"catalog-tag-delete",
		orderfoodRequest.CatalogDeleteInput{
			Reason: "清理测试标签", ExpectedVersion: 3,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || !deleted.Deleted {
		t.Fatalf("deleted result = %+v, replayed = %v", deleted, replayed)
	}
}

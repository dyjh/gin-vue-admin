package orderfood

import (
	"context"
	"testing"
	"time"

	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// newSuggestionCatalogTestService 创建推荐菜索引服务及其独立 MySQL 测试数据库。
func newSuggestionCatalogTestService(
	t *testing.T,
) (*SuggestionCatalogService, *gorm.DB) {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.StandardIngredient{},
		&orderfoodModel.StandardDishIndex{},
		&orderfoodModel.StandardDishIngredient{},
		&orderfoodModel.SuggestionValidationPolicy{},
	); err != nil {
		t.Fatal(err)
	}
	service := NewSuggestionCatalogService(
		db,
		NewPermissionService(db),
		NewIdempotencyService(db),
	)
	service.Now = func() time.Time {
		return time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)
	}
	return service, db
}

func suggestionCatalogTestActor() orderfoodRequest.AdminActor {
	return orderfoodRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:        "catalog-admin",
		RequestID:       "request-suggestion-catalog",
	}
}

func suggestionCatalogBool(value bool) *bool {
	return &value
}

// TestSuggestionCatalogPolicyAndIndexClosure 验证校验策略直接生效且索引维护闭环。
func TestSuggestionCatalogPolicyAndIndexClosure(t *testing.T) {
	service, db := newSuggestionCatalogTestService(t)
	ctx := context.Background()
	category := orderfoodModel.ContentCategory{
		PublicID: "category-main", Name: "主菜", Enabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}

	workspace, err := service.Workspace(ctx)
	if err != nil {
		t.Fatalf("workspace: %v", err)
	}
	if workspace.Policy.CatalogValidationEnabled ||
		workspace.Policy.RetryCount != SuggestionValidationRetryCount ||
		len(workspace.Categories) != 1 ||
		workspace.Readiness.Ready {
		t.Fatalf("unexpected default workspace: %+v", workspace)
	}

	updated, replayed, err := service.UpdatePolicy(
		ctx,
		suggestionCatalogTestActor(),
		"policy-update",
		orderfoodRequest.SuggestionValidationPolicyUpdateInput{
			CatalogValidationEnabled: false,
			Reason:                   "暂时关闭索引校验",
			ExpectedVersion:          workspace.Policy.Version,
		},
	)
	if err != nil || replayed || updated.CatalogValidationEnabled ||
		updated.Version != workspace.Policy.Version+1 {
		t.Fatalf("update policy: updated=%+v replayed=%v err=%v", updated, replayed, err)
	}
	runtimeEnabled, err := CurrentSuggestionCatalogValidation(ctx, db)
	if err != nil || runtimeEnabled {
		t.Fatalf("runtime policy = %v, err=%v", runtimeEnabled, err)
	}
	var policyCount int64
	if err := db.Model(&orderfoodModel.SuggestionValidationPolicy{}).
		Count(&policyCount).Error; err != nil {
		t.Fatal(err)
	}
	if policyCount != 1 {
		t.Fatalf("suggestion policy count = %d, want 1", policyCount)
	}

	ingredient, _, err := service.CreateIngredient(
		ctx,
		suggestionCatalogTestActor(),
		"ingredient-create",
		orderfoodRequest.StandardIngredientInput{
			Name:    "番茄",
			Aliases: []string{"西红柿"},
			Enabled: suggestionCatalogBool(true),
			Reason:  "补充标准食材",
		},
	)
	if err != nil {
		t.Fatalf("create ingredient: %v", err)
	}
	dish, _, err := service.CreateDish(
		ctx,
		suggestionCatalogTestActor(),
		"dish-create",
		orderfoodRequest.StandardDishInput{
			Name:       "番茄炒蛋",
			Aliases:    []string{"西红柿炒鸡蛋"},
			Cuisine:    "家常菜",
			CategoryID: category.PublicID,
			SourceName: "平台整理",
			Enabled:    suggestionCatalogBool(true),
			Ingredients: []orderfoodRequest.StandardDishIngredientInput{
				{
					IngredientID: ingredient.ID,
					Required:     suggestionCatalogBool(true),
					SortOrder:    1,
				},
			},
			Reason: "补充标准菜品索引",
		},
	)
	if err != nil {
		t.Fatalf("create dish: %v", err)
	}
	if dish.CategoryID != category.PublicID ||
		len(dish.Ingredients) != 1 ||
		!dish.Ingredients[0].Required {
		t.Fatalf("unexpected dish response: %+v", dish)
	}
	workspace, err = service.Workspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.DishCount != 1 || workspace.IngredientCount != 1 {
		t.Fatalf("unexpected catalog counts: %+v", workspace)
	}
}

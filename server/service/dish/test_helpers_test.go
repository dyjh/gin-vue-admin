package dish

import (
	"testing"
	"time"

	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

func newOrderFoodTestDB(t *testing.T) *gorm.DB {
	return testutil.NewOrderFoodServiceDB(t, serviceCommon.DefaultRolePermissionMatrix())
}

func stringPointer(value string) *string {
	return &value
}

func migrateContentTestTables(t *testing.T, db *gorm.DB, includeAccessAudit bool) {
	t.Helper()
	models := []interface{}{
		&userModel.MiniAppUser{},
		&commonModel.AdminIdempotencyRecord{},
		&dishModel.ContentCategory{},
		&dishModel.ContentTag{},
		&dishModel.ContentUnit{},
		&dishModel.OfficialDish{},
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
		&engagementModel.FrontCheckin{},
		&engagementModel.FrontPointEntry{},
		&mealModel.FrontMeal{},
		&mealModel.FrontMealCandidate{},
		&mealModel.FrontMealVote{},
		&contentModel.FrontMediaAsset{},
		&contentModel.ImageModerationRecord{},
	}
	if includeAccessAudit {
		models = append(models, &commonModel.AdminAccessAudit{})
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate content tables: %v", err)
	}
}

func seedContentOwnerAndCategory(
	t *testing.T,
	db *gorm.DB,
) (userModel.MiniAppUser, dishModel.ContentCategory) {
	t.Helper()
	now := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	user := userModel.MiniAppUser{
		ID: "content-user-001", OpenIDHash: "content-user-openid-hash",
		OpenIDEncrypted: "content-user-openid-encrypted", Nickname: "内容用户",
		Status: userModel.UserStatusNormal, Version: 1,
		RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create content owner: %v", err)
	}
	category := dishModel.ContentCategory{
		PublicID: "category-home", Name: "家常菜", Enabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create content category: %v", err)
	}
	return user, category
}

func createContentDish(
	t *testing.T,
	db *gorm.DB,
	owner userModel.MiniAppUser,
	category dishModel.ContentCategory,
	publicID string,
	status string,
	discoverable bool,
) contentModel.UserDish {
	t.Helper()
	coverURL := "https://cdn.example.test/" + publicID + ".jpg"
	dish := contentModel.UserDish{
		PublicID: publicID, OwnerID: owner.ID, CoverFileID: "file-" + publicID,
		CoverURL: &coverURL, Name: "菜品-" + publicID, CategoryID: category.ID,
		Status: status, Discoverable: discoverable,
		SourceType:        contentModel.SourceTypeManual,
		MediaReviewStatus: contentModel.MediaReviewPassed,
		Serving:           2, Version: 1,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatalf("create dish %s: %v", publicID, err)
	}
	return dish
}

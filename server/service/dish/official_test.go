package dish

import (
	"bytes"
	"context"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// newOfficialDishTestService 创建官方菜品服务及其独立MySQL测试数据库。
func newOfficialDishTestService(t *testing.T) (*OfficialDishService, *gorm.DB) {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&dishModel.ContentCategory{},
		&dishModel.ContentTag{},
		&dishModel.ContentUnit{},
		&contentModel.FrontMediaAsset{},
		&dishModel.OfficialDish{},
		&dishModel.OfficialDishIngredient{},
		&dishModel.OfficialDishStep{},
		&dishModel.OfficialDishTag{},
		&dishModel.PlatformRecommendation{},
		&dishModel.RecommendationCopy{},
		&commonModel.AdminIdempotencyRecord{},
		&auditModel.AdminAuditLog{},
	); err != nil {
		t.Fatal(err)
	}
	service := NewOfficialDishService(db, serviceCommon.NewIdempotencyService(db))
	service.Now = func() time.Time {
		return time.Date(2026, time.July, 25, 10, 0, 0, 0, time.UTC)
	}
	return service, db
}

// officialDishTestActor 返回官方菜品测试管理员上下文。
func officialDishTestActor() commonRequest.AdminActor {
	return commonRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     commonModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-official-dish",
	}
}

// officialDishTestPNG 生成可用于上传校验的PNG图片。
func officialDishTestPNG(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	picture.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&buffer, picture); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// TestOfficialDishLifecycleAndAdminCoverBypass 校验管理端封面免审核、幂等创建和推荐联动下线。
func TestOfficialDishLifecycleAndAdminCoverBypass(t *testing.T) {
	service, db := newOfficialDishTestService(t)
	ctx := context.Background()
	actor := officialDishTestActor()

	originalLocal := global.GVA_CONFIG.Local
	global.GVA_CONFIG.Local.StorePath = t.TempDir()
	global.GVA_CONFIG.Local.Path = "/uploads"
	t.Cleanup(func() {
		global.GVA_CONFIG.Local = originalLocal
	})

	cover, replayed, err := service.UploadOfficialDishCover(
		ctx,
		actor,
		"official-cover-upload",
		"红烧肉.png",
		"image/png",
		officialDishTestPNG(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || cover.UploadSource != officialDishUploadSource ||
		cover.SourceScene != officialDishCoverScene ||
		cover.ReviewStatus != contentModel.MediaReviewNotRequired {
		t.Fatalf("cover = %+v, replayed = %v", cover, replayed)
	}
	var asset contentModel.FrontMediaAsset
	if err := db.First(&asset, "id = ?", cover.FileID).Error; err != nil {
		t.Fatal(err)
	}
	if asset.AdministratorID == nil || *asset.AdministratorID != actor.AdministratorID ||
		asset.ReviewStatus != contentModel.MediaReviewNotRequired {
		t.Fatalf("asset = %+v", asset)
	}

	category := dishModel.ContentCategory{
		PublicID: "category-home", Name: "家常菜", SortOrder: 1, Enabled: true, Version: 1,
	}
	tag := dishModel.ContentTag{
		PublicID: "tag-pork", Name: "猪肉", SortOrder: 1, Enabled: true, Version: 1,
	}
	unit := dishModel.ContentUnit{
		PublicID: "unit-gram", Name: "克", SortOrder: 1, Enabled: true, Version: 1,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&unit).Error; err != nil {
		t.Fatal(err)
	}
	unitID := unit.PublicID
	input := dishRequest.OfficialDishCreateInput{
		Name: "红烧肉", CategoryID: category.PublicID, TagIDs: []string{tag.PublicID},
		Serving: 4, CoverFileID: cover.FileID, Status: contentModel.DishStatusUsable,
		Ingredients: []dishRequest.OfficialDishIngredientInput{
			{Name: "五花肉", Quantity: stringPointer("500"), UnitID: &unitID, SortOrder: 1},
		},
		Steps: []dishRequest.OfficialDishStepInput{
			{Description: "焯水后炖煮", SortOrder: 1},
		},
	}
	created, replayed, err := service.CreateOfficialDish(
		ctx, actor, "official-dish-create", input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || created.Version != 1 || len(created.Ingredients) != 1 ||
		len(created.Steps) != 1 || len(created.Tags) != 1 {
		t.Fatalf("created = %+v, replayed = %v", created, replayed)
	}
	replayedDish, replayed, err := service.CreateOfficialDish(
		ctx, actor, "official-dish-create", input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !replayed || replayedDish.ID != created.ID {
		t.Fatalf("replayed dish = %+v, replayed = %v", replayedDish, replayed)
	}

	var stored dishModel.OfficialDish
	if err := db.First(&stored, "public_id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}
	now := service.now()
	recommendation := dishModel.PlatformRecommendation{
		ID: "recommendation-official", OfficialDishID: &stored.ID,
		SourceType: recommendationSourceOfficial, SourceDishID: stored.PublicID,
		Position: serviceCommon.RecommendationPositionHome, Status: serviceCommon.RecommendationStatusPublished,
		Selected: true, SortOrder: 1, Version: 1,
		CreatedByID: actor.AdministratorID, CreatedByUsername: actor.Username,
		UpdatedByID: actor.AdministratorID, UpdatedByUsername: actor.Username,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&recommendation).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&stored).UpdateColumn("recommendation_count", 1).Error; err != nil {
		t.Fatal(err)
	}

	updated, replayed, err := service.UpdateOfficialDish(
		ctx,
		created.ID,
		actor,
		"official-dish-update",
		dishRequest.OfficialDishUpdateInput{
			OfficialDishCreateInput: input,
			ExpectedVersion:         1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || updated.Dish.Version != 2 || updated.OfflineRecommendationCount != 1 {
		t.Fatalf("updated = %+v, replayed = %v", updated, replayed)
	}
	if err := db.First(&recommendation, "id = ?", recommendation.ID).Error; err != nil {
		t.Fatal(err)
	}
	if recommendation.Status != serviceCommon.RecommendationStatusOffline || recommendation.Selected ||
		recommendation.OfflineReason == nil {
		t.Fatalf("recommendation after update = %+v", recommendation)
	}

	deleted, replayed, err := service.DeleteOfficialDish(
		ctx,
		created.ID,
		actor,
		"official-dish-delete",
		dishRequest.OfficialDishDeleteInput{
			Reason: "测试软删除菜品", ExpectedVersion: 2,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || !deleted.Deleted || deleted.OfflineRecommendationCount != 0 {
		t.Fatalf("deleted = %+v, replayed = %v", deleted, replayed)
	}
	if err := db.First(&stored, "public_id = ?", created.ID).Error; err != gorm.ErrRecordNotFound {
		t.Fatalf("soft deleted query error = %v", err)
	}
	if err := db.Unscoped().First(&stored, "public_id = ?", created.ID).Error; err != nil ||
		!stored.DeletedAt.Valid {
		t.Fatalf("unscoped official dish = %+v, error = %v", stored, err)
	}
}

// TestOfficialDishDraftAllowsIncompleteContentAndReportsOnlineRecommendations 验证草稿完整度和在线推荐计数。
func TestOfficialDishDraftAllowsIncompleteContentAndReportsOnlineRecommendations(t *testing.T) {
	service, db := newOfficialDishTestService(t)
	ctx := context.Background()
	actor := officialDishTestActor()
	now := service.now()
	category := dishModel.ContentCategory{
		PublicID:  "category-draft",
		Name:      "草稿分类",
		SortOrder: 1,
		Enabled:   true,
		Version:   1,
	}
	tag := dishModel.ContentTag{
		PublicID:  "tag-draft",
		Name:      "草稿标签",
		SortOrder: 1,
		Enabled:   true,
		Version:   1,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatal(err)
	}
	cover := contentModel.FrontMediaAsset{
		ID:                    "official-draft-cover",
		AdministratorID:       &actor.AdministratorID,
		AdministratorUsername: &actor.Username,
		FileName:              "official-draft-cover.png",
		UploadSource:          officialDishUploadSource,
		Scene:                 officialDishCoverScene,
		ResourceStatus:        serviceCommon.MediaResourceActive,
		URL:                   "/uploads/official-draft-cover.png",
		StoragePath:           "official/official-draft-cover.png",
		ContentType:           "image/png",
		Width:                 800,
		Height:                600,
		SizeBytes:             128,
		Checksum:              "official-draft-cover-checksum",
		ReviewStatus:          contentModel.MediaReviewNotRequired,
		CreatedAt:             now,
	}
	if err := db.Create(&cover).Error; err != nil {
		t.Fatal(err)
	}
	input := dishRequest.OfficialDishCreateInput{
		Name:        "未完成官方菜品",
		CategoryID:  category.PublicID,
		TagIDs:      []string{tag.PublicID},
		Serving:     2,
		CoverFileID: cover.ID,
		Status:      contentModel.DishStatusDraft,
		Ingredients: []dishRequest.OfficialDishIngredientInput{},
		Steps:       []dishRequest.OfficialDishStepInput{},
	}
	created, replayed, err := service.CreateOfficialDish(
		ctx,
		actor,
		"official-incomplete-draft-create",
		input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || created.Status != contentModel.DishStatusDraft ||
		len(created.Ingredients) != 0 || len(created.Steps) != 0 {
		t.Fatalf("incomplete draft result=%+v replayed=%v", created, replayed)
	}
	usableInput := input
	usableInput.Status = contentModel.DishStatusUsable
	_, _, err = service.UpdateOfficialDish(
		ctx,
		created.ID,
		actor,
		"official-incomplete-usable-update",
		dishRequest.OfficialDishUpdateInput{
			OfficialDishCreateInput: usableInput,
			ExpectedVersion:         created.Version,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("incomplete usable error=%v type=%d", err, appErrors.GetType(err))
	}
	var stored dishModel.OfficialDish
	if err := db.First(&stored, "public_id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}
	recommendation := dishModel.PlatformRecommendation{
		ID:                "recommendation-official-draft",
		OfficialDishID:    &stored.ID,
		SourceType:        recommendationSourceOfficial,
		SourceDishID:      stored.PublicID,
		Position:          serviceCommon.RecommendationPositionHome,
		Status:            serviceCommon.RecommendationStatusPublished,
		Selected:          true,
		SortOrder:         1,
		Version:           1,
		CreatedByID:       actor.AdministratorID,
		CreatedByUsername: actor.Username,
		UpdatedByID:       actor.AdministratorID,
		UpdatedByUsername: actor.Username,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&recommendation).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&stored).UpdateColumn("recommendation_count", 1).Error; err != nil {
		t.Fatal(err)
	}
	from := now.Add(-time.Minute)
	to := now.Add(time.Minute)
	page, err := service.ListOfficialDishes(
		ctx,
		dishRequest.OfficialDishListQuery{
			TagIDs:      []string{tag.PublicID},
			CreatedFrom: &from,
			CreatedTo:   &to,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.List) != 1 ||
		page.List[0].OnlineRecommendationCount != 1 ||
		page.List[0].UpdatedBy.ID == "" {
		t.Fatalf("official dish list result=%+v", page)
	}
	updated, replayed, err := service.UpdateOfficialDish(
		ctx,
		created.ID,
		actor,
		"official-incomplete-draft-update",
		dishRequest.OfficialDishUpdateInput{
			OfficialDishCreateInput: input,
			ExpectedVersion:         created.Version,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed || updated.OfflineRecommendationCount != 1 ||
		updated.Dish.OnlineRecommendationCount != 0 {
		t.Fatalf("draft update result=%+v replayed=%v", updated, replayed)
	}
}

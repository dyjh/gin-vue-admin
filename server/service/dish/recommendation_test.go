package dish

import (
	"context"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"testing"
	"time"

	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	"gorm.io/datatypes"
)

// TestDiscoverableDishFiltersLockedSourceAndDetailSummary 验证候选池筛选、来源锁和详情审核摘要。
func TestDiscoverableDishFiltersLockedSourceAndDetailSummary(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	discoverableAt := time.Date(2026, time.July, 24, 15, 0, 0, 0, time.UTC)
	dish := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-discoverable-detail",
		contentModel.DishStatusUsable,
		true,
	)
	lockedDish := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-discoverable-locked",
		contentModel.DishStatusUsable,
		true,
	)
	if err := db.Model(&contentModel.UserDish{}).
		Where("id IN ?", []uint{dish.ID, lockedDish.ID}).
		Update("discoverable_at", discoverableAt).Error; err != nil {
		t.Fatalf("set discoverable time: %v", err)
	}
	if err := db.Model(&lockedDish).Update("source_locked", true).Error; err != nil {
		t.Fatalf("lock copied source: %v", err)
	}
	tag := dishModel.ContentTag{
		PublicID: "tag-discoverable",
		Name:     "候选标签",
		Enabled:  true,
		Version:  1,
	}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("create discoverable tag: %v", err)
	}
	relations := []contentModel.UserDishTag{
		{DishID: dish.ID, TagID: tag.ID, CreatedAt: discoverableAt},
		{DishID: lockedDish.ID, TagID: tag.ID, CreatedAt: discoverableAt},
	}
	if err := db.Create(&relations).Error; err != nil {
		t.Fatalf("attach discoverable tags: %v", err)
	}
	asset := contentModel.FrontMediaAsset{
		ID:             dish.CoverFileID,
		UserID:         owner.ID,
		FileName:       "dish-discoverable-detail.jpg",
		UploadSource:   "user_upload",
		Scene:          string(contentModel.ModerationSceneDishCover),
		ResourceStatus: "active",
		URL:            *dish.CoverURL,
		StoragePath:    "test/dish-discoverable-detail.jpg",
		ContentType:    "image/jpeg",
		Width:          800,
		Height:         600,
		SizeBytes:      128,
		Checksum:       "dish-discoverable-detail-checksum",
		ReviewStatus:   string(contentModel.ModerationStatusPassed),
		CreatedAt:      discoverableAt,
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("create discoverable media: %v", err)
	}
	objectType := "user_dish"
	objectID := dish.PublicID
	record := contentModel.ImageModerationRecord{
		ID:             "moderation-discoverable-detail",
		RequestID:      "request-moderation-discoverable",
		FileID:         dish.CoverFileID,
		Scene:          contentModel.ModerationSceneDishCover,
		Status:         contentModel.ModerationStatusPassed,
		RiskLabelsJSON: datatypes.JSON([]byte(`["normal"]`)),
		ObjectType:     &objectType,
		ObjectID:       &objectID,
		CreatedAt:      discoverableAt,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create discoverable moderation record: %v", err)
	}
	recommendation := dishModel.PlatformRecommendation{
		ID:                "recommendation-discoverable-detail",
		DishID:            &dish.ID,
		SourceType:        recommendationSourceCreator,
		SourceDishID:      dish.PublicID,
		Position:          serviceCommon.RecommendationPositionHome,
		Status:            recommendationStatusDraft,
		Selected:          false,
		SortOrder:         3,
		Version:           1,
		CreatedByID:       41,
		CreatedByUsername: "recommendation-admin",
		UpdatedByID:       41,
		UpdatedByUsername: "recommendation-admin",
		CreatedAt:         discoverableAt,
		UpdatedAt:         discoverableAt,
	}
	if err := db.Create(&recommendation).Error; err != nil {
		t.Fatalf("create discoverable recommendation: %v", err)
	}

	service := NewRecommendationService(db, serviceCommon.NewIdempotencyService(db))
	from := discoverableAt.Add(-time.Minute)
	to := discoverableAt.Add(time.Minute)
	page, err := service.ListDiscoverableDishes(
		context.Background(),
		dishRequest.DiscoverableDishListQuery{
			AuthorKeyword:    "内容用",
			TagIDs:           []string{tag.PublicID},
			DiscoverableFrom: &from,
			DiscoverableTo:   &to,
		},
	)
	if err != nil {
		t.Fatalf("list discoverable dishes with filters: %v", err)
	}
	if page.Total != 2 || len(page.List) != 2 {
		t.Fatalf("discoverable filters result: %+v", page)
	}
	foundLocked := false
	for _, item := range page.List {
		if item.ID == lockedDish.PublicID {
			foundLocked = item.SourceLocked
		}
	}
	if !foundLocked {
		t.Fatalf("locked source flag missing from list: %+v", page.List)
	}
	unselected := false
	unselectedPage, err := service.ListDiscoverableDishes(
		context.Background(),
		dishRequest.DiscoverableDishListQuery{Selected: &unselected},
	)
	if err != nil {
		t.Fatalf("list unselected discoverable dishes: %v", err)
	}
	if unselectedPage.Total != 1 ||
		len(unselectedPage.List) != 1 ||
		unselectedPage.List[0].ID != lockedDish.PublicID {
		t.Fatalf("unselected filter result: %+v", unselectedPage)
	}
	outsideFrom := discoverableAt.Add(time.Hour)
	outside, err := service.ListDiscoverableDishes(
		context.Background(),
		dishRequest.DiscoverableDishListQuery{DiscoverableFrom: &outsideFrom},
	)
	if err != nil {
		t.Fatalf("list discoverable dishes outside range: %v", err)
	}
	if outside.Total != 0 || len(outside.List) != 0 {
		t.Fatalf("discoverable time filter leaked rows: %+v", outside)
	}

	detail, err := service.GetDiscoverableDish(context.Background(), dish.PublicID, true)
	if err != nil {
		t.Fatalf("get discoverable detail: %v", err)
	}
	if detail.ModerationSummary == nil || detail.ModerationSummary.ID != record.ID {
		t.Fatalf("discoverable moderation summary: %+v", detail.ModerationSummary)
	}
	if detail.Recommendation == nil || detail.Recommendation.ID != recommendation.ID {
		t.Fatalf("discoverable recommendation summary: %+v", detail.Recommendation)
	}
	withoutModeration, err := service.GetDiscoverableDish(
		context.Background(),
		dish.PublicID,
		false,
	)
	if err != nil {
		t.Fatalf("get discoverable detail without moderation: %v", err)
	}
	if withoutModeration.ModerationSummary != nil {
		t.Fatalf("moderation summary leaked without conditional permission: %+v", withoutModeration)
	}
}

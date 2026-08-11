package content

import (
	"context"
	"encoding/json"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const contentTestPreviewSecret = "content-test-preview-secret"

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

// TestRecommendationEntryGovernanceProcessesCreatorSource 验证推荐入口会处理用户源菜品及全部相关推荐。
func TestRecommendationEntryGovernanceProcessesCreatorSource(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	source := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-governance-entry",
		contentModel.DishStatusUsable,
		true,
	)
	recipe := createContentRecipe(t, db, owner, "recipe-governance-entry", source)
	meal := mealModel.FrontMeal{
		ID:         "meal-governance-entry",
		CreatorID:  owner.ID,
		Name:       "待确认饭局",
		Code:       "GOV001",
		Status:     mealModel.MealCollecting,
		DeadlineAt: time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC),
		CreatedAt:  time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&meal).Error; err != nil {
		t.Fatalf("create governed meal: %v", err)
	}
	candidate := mealModel.FrontMealCandidate{
		ID: "candidate-governance-entry", MealID: meal.ID, DishID: source.ID,
		Available: true, SortOrder: 1, CreatedAt: meal.CreatedAt,
	}
	if err := db.Create(&candidate).Error; err != nil {
		t.Fatalf("create governed meal candidate: %v", err)
	}
	recommendations := []dishModel.PlatformRecommendation{
		{
			ID: "recommendation-governance-entry", DishID: &source.ID,
			SourceType: "creator", SourceDishID: source.PublicID, Position: serviceCommon.RecommendationPositionHome,
			Status: serviceCommon.RecommendationStatusPublished, Selected: true, SortOrder: 1, Version: 1,
			CreatedByID: 71, CreatedByUsername: "platform-admin",
			UpdatedByID: 71, UpdatedByUsername: "platform-admin",
			CreatedAt: meal.CreatedAt, UpdatedAt: meal.UpdatedAt,
		},
		{
			ID:         "recommendation-governance-linked",
			SourceType: "creator", SourceDishID: source.PublicID, Position: serviceCommon.RecommendationPositionHome,
			Status: serviceCommon.RecommendationStatusPublished, Selected: true, SortOrder: 2, Version: 1,
			CreatedByID: 71, CreatedByUsername: "platform-admin",
			UpdatedByID: 71, UpdatedByUsername: "platform-admin",
			CreatedAt: meal.CreatedAt, UpdatedAt: meal.UpdatedAt,
		},
	}
	if err := db.Create(&recommendations).Error; err != nil {
		t.Fatalf("create governed recommendations: %v", err)
	}
	input := contentRequest.GovernanceActionPreviewInput{
		TargetType:            contentModel.GovernanceTargetDish,
		TargetID:              source.PublicID,
		EntryRecommendationID: recommendations[0].ID,
		Actions:               []string{contentModel.GovernanceActionSoftDeleteDish},
		ViolationType:         "content_violation",
		Severity:              contentModel.GovernanceSeverityNormal,
		Reason:                "来源菜品存在违规内容",
		ExpectedVersion:       int(source.Version),
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 71, AuthorityID: commonModel.AuthorityPlatformSuperAdmin,
		Username: "platform-admin", RequestID: "request-recommendation-entry-governance",
	}
	service := contentServiceForTest(db)
	preview, err := service.PreviewGovernance(context.Background(), input, actor)
	if err != nil {
		t.Fatalf("preview recommendation entry governance: %v", err)
	}
	if preview.EntryRecommendationID == nil ||
		*preview.EntryRecommendationID != recommendations[0].ID ||
		preview.RecommendationCount != 2 {
		t.Fatalf("unexpected recommendation entry preview: %+v", preview)
	}
	result, err := service.ExecuteGovernance(
		context.Background(),
		contentRequest.GovernanceActionExecuteInput{
			GovernanceActionPreviewInput: input,
			PreviewToken:                 preview.PreviewToken,
		},
		actor,
		"recommendation-entry-governance",
	)
	if err != nil || result.Status != "succeeded" {
		t.Fatalf("execute recommendation entry governance: result=%+v err=%v", result, err)
	}
	var persistedSource contentModel.UserDish
	if err := db.Unscoped().First(&persistedSource, source.ID).Error; err != nil {
		t.Fatalf("load governed source: %v", err)
	}
	if !persistedSource.DeletedAt.Valid || persistedSource.Discoverable {
		t.Fatalf("creator source was not fully governed: %+v", persistedSource)
	}
	var onlineRecommendationCount int64
	if err := db.Model(&dishModel.PlatformRecommendation{}).
		Where("id IN ? AND status = ?", []string{recommendations[0].ID, recommendations[1].ID}, serviceCommon.RecommendationStatusPublished).
		Count(&onlineRecommendationCount).Error; err != nil || onlineRecommendationCount != 0 {
		t.Fatalf("linked recommendations remain online: count=%d err=%v", onlineRecommendationCount, err)
	}
	var recipeRelationCount int64
	if err := db.Model(&contentModel.RecipeDish{}).
		Where("recipe_id = ? AND dish_id = ?", recipe.ID, source.ID).
		Count(&recipeRelationCount).Error; err != nil || recipeRelationCount != 0 {
		t.Fatalf("governed source remains in recipe: count=%d err=%v", recipeRelationCount, err)
	}
	var persistedCandidate mealModel.FrontMealCandidate
	if err := db.First(&persistedCandidate, "id = ?", candidate.ID).Error; err != nil {
		t.Fatalf("load governed meal candidate: %v", err)
	}
	if persistedCandidate.Available ||
		persistedCandidate.UnavailableReason == nil ||
		*persistedCandidate.UnavailableReason != "source_deleted" {
		t.Fatalf("meal candidate was not invalidated: %+v", persistedCandidate)
	}
	var notificationCount int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("user_id = ? AND type = ?", owner.ID, "governance").
		Count(&notificationCount).Error; err != nil || notificationCount != 1 {
		t.Fatalf("creator governance notification count=%d err=%v", notificationCount, err)
	}
	var record contentModel.GovernanceRecord
	if err := db.Where("public_id = ?", result.RecordID).First(&record).Error; err != nil {
		t.Fatalf("load recommendation entry governance record: %v", err)
	}
	var storedImpact struct {
		EntryRecommendationID *string `json:"entryRecommendationId"`
	}
	if err := json.Unmarshal(record.ImpactSnapshot, &storedImpact); err != nil ||
		storedImpact.EntryRecommendationID == nil ||
		*storedImpact.EntryRecommendationID != recommendations[0].ID {
		t.Fatalf("entry recommendation trace was not persisted: impact=%s err=%v", record.ImpactSnapshot, err)
	}
}

// TestRecommendationEntryGovernanceRejectsMismatchedSource 验证推荐入口不能串联处理其他源菜品。
func TestRecommendationEntryGovernanceRejectsMismatchedSource(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	first := createContentDish(t, db, owner, category, "dish-entry-first", contentModel.DishStatusUsable, true)
	second := createContentDish(t, db, owner, category, "dish-entry-second", contentModel.DishStatusUsable, true)
	recommendation := dishModel.PlatformRecommendation{
		ID: "recommendation-entry-mismatch", DishID: &first.ID,
		SourceType: "creator", SourceDishID: first.PublicID, Position: serviceCommon.RecommendationPositionHome,
		Status: serviceCommon.RecommendationStatusPublished, Selected: true, SortOrder: 1, Version: 1,
		CreatedByID: 72, CreatedByUsername: "platform-admin",
		UpdatedByID: 72, UpdatedByUsername: "platform-admin",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := db.Create(&recommendation).Error; err != nil {
		t.Fatalf("create mismatched recommendation: %v", err)
	}
	_, err := contentServiceForTest(db).PreviewGovernance(
		context.Background(),
		contentRequest.GovernanceActionPreviewInput{
			TargetType:            contentModel.GovernanceTargetDish,
			TargetID:              second.PublicID,
			EntryRecommendationID: recommendation.ID,
			Actions:               []string{contentModel.GovernanceActionDisableDiscoverability},
			ViolationType:         "content_violation",
			Severity:              contentModel.GovernanceSeverityNormal,
			Reason:                "尝试处理不关联的源菜品",
			ExpectedVersion:       int(second.Version),
		},
		commonRequest.AdminActor{
			AdministratorID: 72, AuthorityID: commonModel.AuthorityPlatformSuperAdmin,
			Username: "platform-admin", RequestID: "request-entry-mismatch",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInvalidGovernance {
		t.Fatalf("mismatched recommendation source error=%v type=%v", err, appErrors.GetType(err))
	}
}

// TestRecommendationEntryGovernanceProcessesOfficialCopyChain 验证官方源及其用户复制链可由推荐入口一次处理。
func TestRecommendationEntryGovernanceProcessesOfficialCopyChain(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	now := time.Date(2026, time.July, 24, 14, 0, 0, 0, time.UTC)
	official := dishModel.OfficialDish{
		PublicID: "official-governance-entry", Name: "官方违规菜品",
		CoverFileID: "file-official-governance-entry",
		CoverURL:    "https://cdn.example.test/official-governance-entry.jpg",
		CategoryID:  category.ID, Serving: 2, Status: contentModel.DishStatusUsable,
		Version: 1, CreatedByID: 73, CreatedByUsername: "platform-admin",
		UpdatedByID: 73, UpdatedByUsername: "platform-admin",
	}
	if err := db.Create(&official).Error; err != nil {
		t.Fatalf("create governed official source: %v", err)
	}
	copyDish := createContentDish(
		t, db, owner, category, "official-governance-copy",
		contentModel.DishStatusUsable, false,
	)
	if err := db.Model(&copyDish).Updates(map[string]interface{}{
		"source_type":                    contentModel.SourceTypeOfficialCopy,
		"direct_source_official_dish_id": official.ID,
		"root_source_official_dish_id":   official.ID,
		"source_locked":                  true,
		"chain_depth":                    1,
	}).Error; err != nil {
		t.Fatalf("link official copy chain: %v", err)
	}
	recommendations := []dishModel.PlatformRecommendation{
		{
			ID: "recommendation-official-entry", OfficialDishID: &official.ID,
			SourceType: "official", SourceDishID: official.PublicID, Position: serviceCommon.RecommendationPositionHome,
			Status: serviceCommon.RecommendationStatusPublished, Selected: true, SortOrder: 1, Version: 1,
			CreatedByID: 73, CreatedByUsername: "platform-admin",
			UpdatedByID: 73, UpdatedByUsername: "platform-admin",
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:         "recommendation-official-copy",
			SourceType: "creator", SourceDishID: copyDish.PublicID, Position: serviceCommon.RecommendationPositionHome,
			Status: serviceCommon.RecommendationStatusPublished, Selected: true, SortOrder: 2, Version: 1,
			CreatedByID: 73, CreatedByUsername: "platform-admin",
			UpdatedByID: 73, UpdatedByUsername: "platform-admin",
			CreatedAt: now, UpdatedAt: now,
		},
	}
	if err := db.Create(&recommendations).Error; err != nil {
		t.Fatalf("create official governance recommendations: %v", err)
	}
	input := contentRequest.GovernanceActionPreviewInput{
		TargetType:            contentModel.GovernanceTargetOfficialDish,
		TargetID:              official.PublicID,
		EntryRecommendationID: recommendations[0].ID,
		Actions:               []string{contentModel.GovernanceActionDeleteCopyChain},
		ViolationType:         "serious_content_violation",
		Severity:              contentModel.GovernanceSeveritySerious,
		Reason:                "官方源菜品存在严重违规内容",
		ExpectedVersion:       int(official.Version),
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 73, AuthorityID: commonModel.AuthorityPlatformSuperAdmin,
		Username: "platform-admin", RequestID: "request-official-entry-governance",
	}
	contentService := contentServiceForTest(db)
	preview, err := contentService.PreviewGovernance(context.Background(), input, actor)
	if err != nil {
		t.Fatalf("preview official recommendation governance: %v", err)
	}
	if !preview.Asynchronous || preview.CopiedDishCount != 1 ||
		preview.RecommendationCount != 2 || preview.AffectedUserCount != 1 ||
		preview.NotificationCount != 1 {
		t.Fatalf("unexpected official source governance preview: %+v", preview)
	}
	execution, err := contentService.ExecuteGovernance(
		context.Background(),
		contentRequest.GovernanceActionExecuteInput{
			GovernanceActionPreviewInput: input,
			PreviewToken:                 preview.PreviewToken,
			ConfirmText:                  governanceCopyChainConfirmText,
		},
		actor,
		"official-recommendation-entry-governance",
	)
	if err != nil || execution.Status != contentModel.GovernanceJobPending ||
		execution.JobID == nil || execution.AffectedCount != 2 {
		t.Fatalf("execute official recommendation governance: result=%+v err=%v", execution, err)
	}
	var persistedOfficial dishModel.OfficialDish
	if err := db.Unscoped().First(&persistedOfficial, official.ID).Error; err != nil {
		t.Fatalf("load governed official source: %v", err)
	}
	if !persistedOfficial.DeletedAt.Valid {
		t.Fatalf("official source was not deleted before copy-chain processing: %+v", persistedOfficial)
	}
	var entryRecommendation dishModel.PlatformRecommendation
	if err := db.First(&entryRecommendation, "id = ?", recommendations[0].ID).Error; err != nil {
		t.Fatalf("load official entry recommendation: %v", err)
	}
	if entryRecommendation.Status != serviceCommon.RecommendationStatusOffline || entryRecommendation.Selected {
		t.Fatalf("official entry recommendation remains online: %+v", entryRecommendation)
	}
	var notificationCountBefore int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("type = ?", "governance").
		Count(&notificationCountBefore).Error; err != nil || notificationCountBefore != 0 {
		t.Fatalf("official source generated an unexpected notification: count=%d err=%v", notificationCountBefore, err)
	}
	governanceService := NewGovernanceService(
		db,
		serviceCommon.NewPermissionService(db),
		serviceCommon.NewIdempotencyService(db),
	)
	if err := governanceService.ProcessPendingJobs(context.Background(), 100); err != nil {
		t.Fatalf("process official copy chain governance job: %v", err)
	}
	var persistedCopy contentModel.UserDish
	if err := db.Unscoped().First(&persistedCopy, copyDish.ID).Error; err != nil {
		t.Fatalf("load governed official copy: %v", err)
	}
	if !persistedCopy.DeletedAt.Valid {
		t.Fatalf("official copy remains active: %+v", persistedCopy)
	}
	var copyRecommendation dishModel.PlatformRecommendation
	if err := db.First(&copyRecommendation, "id = ?", recommendations[1].ID).Error; err != nil {
		t.Fatalf("load official copy recommendation: %v", err)
	}
	if copyRecommendation.Status != serviceCommon.RecommendationStatusOffline || copyRecommendation.Selected {
		t.Fatalf("official copy recommendation remains online: %+v", copyRecommendation)
	}
	var notificationCount int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("user_id = ? AND type = ?", owner.ID, "governance").
		Count(&notificationCount).Error; err != nil || notificationCount != 1 {
		t.Fatalf("official copy notification count=%d err=%v", notificationCount, err)
	}
	var record contentModel.GovernanceRecord
	if err := db.Where("public_id = ?", execution.RecordID).First(&record).Error; err != nil {
		t.Fatalf("load official governance record: %v", err)
	}
	if record.TargetType != contentModel.GovernanceTargetOfficialDish ||
		record.AffectedCount != 2 {
		t.Fatalf("official governance record progress is incorrect: %+v", record)
	}
}

// TestRecommendationEntryGovernanceProcessesOfficialSource 验证一般违规会处理官方源并下线相关推荐且不通知用户。
func TestRecommendationEntryGovernanceProcessesOfficialSource(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	_, category := seedContentOwnerAndCategory(t, db)
	now := time.Date(2026, time.July, 24, 15, 0, 0, 0, time.UTC)
	official := dishModel.OfficialDish{
		PublicID: "official-governance-normal", Name: "一般违规官方菜品",
		CoverFileID: "file-official-governance-normal",
		CoverURL:    "https://cdn.example.test/official-governance-normal.jpg",
		CategoryID:  category.ID, Serving: 2, Status: contentModel.DishStatusUsable,
		Version: 1, CreatedByID: 74, CreatedByUsername: "platform-admin",
		UpdatedByID: 74, UpdatedByUsername: "platform-admin",
	}
	if err := db.Create(&official).Error; err != nil {
		t.Fatalf("create normal governed official source: %v", err)
	}
	recommendation := dishModel.PlatformRecommendation{
		ID: "recommendation-official-normal", OfficialDishID: &official.ID,
		SourceType: "official", SourceDishID: official.PublicID, Position: serviceCommon.RecommendationPositionHome,
		Status: serviceCommon.RecommendationStatusPublished, Selected: true, SortOrder: 1, Version: 1,
		CreatedByID: 74, CreatedByUsername: "platform-admin",
		UpdatedByID: 74, UpdatedByUsername: "platform-admin",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&recommendation).Error; err != nil {
		t.Fatalf("create normal official recommendation: %v", err)
	}
	input := contentRequest.GovernanceActionPreviewInput{
		TargetType:            contentModel.GovernanceTargetOfficialDish,
		TargetID:              official.PublicID,
		EntryRecommendationID: recommendation.ID,
		Actions:               []string{contentModel.GovernanceActionSoftDeleteOfficialDish},
		ViolationType:         "content_violation",
		Severity:              contentModel.GovernanceSeverityNormal,
		Reason:                "官方源菜品存在违规内容",
		ExpectedVersion:       int(official.Version),
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 74, AuthorityID: commonModel.AuthorityPlatformSuperAdmin,
		Username: "platform-admin", RequestID: "request-official-normal-governance",
	}
	service := contentServiceForTest(db)
	preview, err := service.PreviewGovernance(context.Background(), input, actor)
	if err != nil {
		t.Fatalf("preview normal official governance: %v", err)
	}
	if preview.Asynchronous || preview.RecommendationCount != 1 ||
		preview.AffectedUserCount != 0 || preview.NotificationCount != 0 {
		t.Fatalf("unexpected normal official preview: %+v", preview)
	}
	result, err := service.ExecuteGovernance(
		context.Background(),
		contentRequest.GovernanceActionExecuteInput{
			GovernanceActionPreviewInput: input,
			PreviewToken:                 preview.PreviewToken,
		},
		actor,
		"official-normal-governance",
	)
	if err != nil || result.Status != "succeeded" || result.JobID != nil {
		t.Fatalf("execute normal official governance: result=%+v err=%v", result, err)
	}
	var persistedOfficial dishModel.OfficialDish
	if err := db.Unscoped().First(&persistedOfficial, official.ID).Error; err != nil {
		t.Fatalf("load normal governed official source: %v", err)
	}
	if !persistedOfficial.DeletedAt.Valid || persistedOfficial.Version != 2 {
		t.Fatalf("normal official source was not deleted: %+v", persistedOfficial)
	}
	var persistedRecommendation dishModel.PlatformRecommendation
	if err := db.First(&persistedRecommendation, "id = ?", recommendation.ID).Error; err != nil {
		t.Fatalf("load normal official recommendation: %v", err)
	}
	if persistedRecommendation.Status != serviceCommon.RecommendationStatusOffline ||
		persistedRecommendation.Selected {
		t.Fatalf("normal official recommendation remains online: %+v", persistedRecommendation)
	}
	var notificationCount int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("type = ?", "governance").
		Count(&notificationCount).Error; err != nil || notificationCount != 0 {
		t.Fatalf("normal official governance notification count=%d err=%v", notificationCount, err)
	}
}

func seedContentOwnerAndCategory(
	t *testing.T,
	db *gorm.DB,
) (userModel.MiniAppUser, dishModel.ContentCategory) {
	t.Helper()
	now := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	user := userModel.MiniAppUser{
		ID:              "content-user-001",
		OpenIDHash:      "content-user-openid-hash",
		OpenIDEncrypted: "content-user-openid-encrypted",
		Nickname:        "内容用户",
		Status:          userModel.UserStatusNormal,
		Version:         1,
		RegisteredAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create content owner: %v", err)
	}
	category := dishModel.ContentCategory{
		PublicID: "category-home",
		Name:     "家常菜",
		Enabled:  true,
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
		PublicID:          publicID,
		OwnerID:           owner.ID,
		CoverFileID:       "file-" + publicID,
		CoverURL:          &coverURL,
		Name:              "菜品-" + publicID,
		CategoryID:        category.ID,
		Status:            status,
		Discoverable:      discoverable,
		SourceType:        contentModel.SourceTypeManual,
		MediaReviewStatus: contentModel.MediaReviewPassed,
		Serving:           2,
		Version:           1,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatalf("create dish %s: %v", publicID, err)
	}
	return dish
}

func createContentRecipe(
	t *testing.T,
	db *gorm.DB,
	owner userModel.MiniAppUser,
	publicID string,
	dishes ...contentModel.UserDish,
) contentModel.UserRecipe {
	t.Helper()
	note := "家庭菜谱备注"
	recipe := contentModel.UserRecipe{
		PublicID: publicID,
		OwnerID:  owner.ID,
		Name:     "菜谱-" + publicID,
		Note:     &note,
		Version:  1,
	}
	if err := db.Create(&recipe).Error; err != nil {
		t.Fatalf("create recipe %s: %v", publicID, err)
	}
	for index, dish := range dishes {
		relation := contentModel.RecipeDish{
			PublicID:  "relation-" + publicID + "-" + dish.PublicID,
			RecipeID:  recipe.ID,
			DishID:    dish.ID,
			SortOrder: index + 1,
			AddedAt:   time.Date(2026, time.July, 24, 10, index, 0, 0, time.UTC),
		}
		if err := db.Create(&relation).Error; err != nil {
			t.Fatalf("create recipe relation: %v", err)
		}
	}
	return recipe
}

func contentServiceForTest(
	db *gorm.DB,
	options ...ContentServiceOption,
) *ContentService {
	return NewContentService(
		db,
		serviceCommon.NewAccessAuditService(db),
		serviceCommon.NewIdempotencyService(db),
		serviceCommon.NewPermissionService(db),
		[]byte(contentTestPreviewSecret),
		options...,
	)
}

func grantContentReadPermissions(t *testing.T, db *gorm.DB, authorityID uint) {
	t.Helper()
	grantTestButtonPermission(t, db, authorityID, PermissionUserDishRead)
	grantTestButtonPermission(t, db, authorityID, PermissionUserRecipeRead)
}

func TestContentServiceUsesGlobalDatabaseAfterConstruction(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	previousDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previousDB
	})

	heldByAPI := NewContentService(
		nil,
		serviceCommon.NewAccessAuditService(nil),
		serviceCommon.NewIdempotencyService(nil),
		serviceCommon.NewPermissionService(nil),
		nil,
	)
	if heldByAPI == nil {
		t.Fatal("content singleton was not constructed")
	}
	page, err := heldByAPI.ListUserDishes(
		context.Background(),
		contentRequest.UserDishSearch{},
		commonRequest.AdminActor{
			AdministratorID: 51,
			AuthorityID:     commonModel.AuthorityPlatformSuperAdmin,
			Username:        "platform-admin",
			RequestID:       "request-bind-content-db",
		},
	)
	if err != nil {
		t.Fatalf("content service did not use the runtime database: %v", err)
	}
	if page.Total != 0 || len(page.List) != 0 {
		t.Fatalf("unexpected page after bind: %+v", page)
	}
}

func TestContentListsMaskRestrictedCoversWithoutPrivateRead(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	publicDish := createContentDish(t, db, owner, category, "dish-public", contentModel.DishStatusUsable, true)
	privateDish := createContentDish(t, db, owner, category, "dish-private", contentModel.DishStatusDraft, false)
	deletedDish := createContentDish(t, db, owner, category, "dish-deleted", contentModel.DishStatusUsable, true)
	if err := db.Delete(&deletedDish).Error; err != nil {
		t.Fatalf("soft delete dish: %v", err)
	}

	publicRecipe := createContentRecipe(t, db, owner, "recipe-public", publicDish)
	privateRecipe := createContentRecipe(t, db, owner, "recipe-private", privateDish)
	deletedRecipe := createContentRecipe(t, db, owner, "recipe-deleted", publicDish)
	if err := db.Delete(&deletedRecipe).Error; err != nil {
		t.Fatalf("soft delete recipe: %v", err)
	}

	const authorityID uint = 7001
	grantContentReadPermissions(t, db, authorityID)
	actor := commonRequest.AdminActor{AdministratorID: 31, AuthorityID: authorityID, RequestID: "request-list-mask"}
	service := contentServiceForTest(db)

	dishPage, err := service.ListUserDishes(context.Background(), contentRequest.UserDishSearch{}, actor)
	if err != nil {
		t.Fatalf("list active dishes: %v", err)
	}
	dishCovers := make(map[string]*string, len(dishPage.List))
	for _, item := range dishPage.List {
		dishCovers[item.ID] = item.CoverURL
	}
	if dishCovers[publicDish.PublicID] == nil {
		t.Fatalf("public usable dish cover was masked: %+v", dishPage.List)
	}
	if dishCovers[privateDish.PublicID] != nil {
		t.Fatal("private draft dish cover was returned without private-read")
	}
	deletedDishPage, err := service.ListUserDishes(
		context.Background(),
		contentRequest.UserDishSearch{DeletionStatus: "deleted"},
		actor,
	)
	if err != nil {
		t.Fatalf("list deleted dishes: %v", err)
	}
	if len(deletedDishPage.List) != 1 || deletedDishPage.List[0].CoverURL != nil {
		t.Fatalf("deleted dish cover was not masked: %+v", deletedDishPage.List)
	}

	recipePage, err := service.ListUserRecipes(context.Background(), contentRequest.UserRecipeSearch{}, actor)
	if err != nil {
		t.Fatalf("list active recipes: %v", err)
	}
	recipeCovers := make(map[string]*string, len(recipePage.List))
	for _, item := range recipePage.List {
		recipeCovers[item.ID] = item.CoverURL
	}
	if recipeCovers[publicRecipe.PublicID] == nil {
		t.Fatal("recipe derived from public dish lost its cover")
	}
	if recipeCovers[privateRecipe.PublicID] != nil {
		t.Fatal("recipe derived from private dish returned its cover without private-read")
	}
	deletedRecipePage, err := service.ListUserRecipes(
		context.Background(),
		contentRequest.UserRecipeSearch{DeletionStatus: "deleted"},
		actor,
	)
	if err != nil {
		t.Fatalf("list deleted recipes: %v", err)
	}
	if len(deletedRecipePage.List) != 1 || deletedRecipePage.List[0].CoverURL != nil {
		t.Fatalf("deleted recipe cover was not masked: %+v", deletedRecipePage.List)
	}
}

// TestUserDishListFiltersOwnerKeywordAndDetailReturnsModerationSummary 验证用户关键词筛选及菜品最新审核摘要。
func TestUserDishListFiltersOwnerKeywordAndDetailReturnsModerationSummary(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	dish := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-owner-moderation",
		contentModel.DishStatusUsable,
		true,
	)
	now := time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC)
	asset := contentModel.FrontMediaAsset{
		ID:             dish.CoverFileID,
		UserID:         owner.ID,
		FileName:       "dish-owner-moderation.jpg",
		UploadSource:   "user_upload",
		Scene:          string(contentModel.ModerationSceneDishCover),
		ResourceStatus: "active",
		URL:            *dish.CoverURL,
		StoragePath:    "test/dish-owner-moderation.jpg",
		ContentType:    "image/jpeg",
		Width:          800,
		Height:         600,
		SizeBytes:      128,
		Checksum:       "dish-owner-moderation-checksum",
		ReviewStatus:   string(contentModel.ModerationStatusPassed),
		CreatedAt:      now,
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("create dish cover media: %v", err)
	}
	objectType := "user_dish"
	objectID := dish.PublicID
	providerRequestID := "aliyun-dish-owner-moderation"
	record := contentModel.ImageModerationRecord{
		ID:                  "moderation-dish-owner",
		RequestID:           "request-moderation-dish-owner",
		FileID:              dish.CoverFileID,
		Scene:               contentModel.ModerationSceneDishCover,
		Status:              contentModel.ModerationStatusPassed,
		RiskLabelsJSON:      datatypes.JSON([]byte(`["normal"]`)),
		ProviderRequestID:   &providerRequestID,
		ProviderSummaryJSON: datatypes.JSON([]byte(`{"source":"aliyun"}`)),
		ObjectType:          &objectType,
		ObjectID:            &objectID,
		CreatedAt:           now,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create dish moderation record: %v", err)
	}

	const authorityID uint = 7004
	grantContentReadPermissions(t, db, authorityID)
	grantTestButtonPermission(t, db, authorityID, PermissionUserDishPrivateRead)
	actor := commonRequest.AdminActor{
		AdministratorID: 34,
		AuthorityID:     authorityID,
		Username:        "content-admin",
		RequestID:       "request-owner-moderation",
	}
	service := contentServiceForTest(db)
	page, err := service.ListUserDishes(
		context.Background(),
		contentRequest.UserDishSearch{UserKeyword: "内容用"},
		actor,
	)
	if err != nil {
		t.Fatalf("list dishes by owner nickname: %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 || page.List[0].ID != dish.PublicID {
		t.Fatalf("owner nickname filter result: %+v", page)
	}
	unmatched, err := service.ListUserDishes(
		context.Background(),
		contentRequest.UserDishSearch{UserKeyword: "不存在的用户"},
		actor,
	)
	if err != nil {
		t.Fatalf("list dishes by unmatched owner: %v", err)
	}
	if unmatched.Total != 0 || len(unmatched.List) != 0 {
		t.Fatalf("unmatched owner filter leaked dishes: %+v", unmatched)
	}

	detail, err := service.GetUserDishDetail(context.Background(), dish.PublicID, actor)
	if err != nil {
		t.Fatalf("get dish detail with moderation summary: %v", err)
	}
	if detail.ModerationSummary == nil ||
		detail.ModerationSummary.ID != record.ID ||
		detail.ModerationSummary.ProviderRequestID == nil ||
		*detail.ModerationSummary.ProviderRequestID != providerRequestID ||
		len(detail.ModerationSummary.RiskLabels) != 1 ||
		detail.ModerationSummary.RiskLabels[0] != "normal" {
		t.Fatalf("unexpected moderation summary: %+v", detail.ModerationSummary)
	}
	if detail.ModerationSummary.User == nil || detail.ModerationSummary.User.ID != owner.ID {
		t.Fatalf("moderation uploader was not resolved: %+v", detail.ModerationSummary)
	}
	if detail.AccessAuditID == "" {
		t.Fatal("private dish detail did not return access audit ID")
	}
}

func TestContentPrivateDetailRejectsMissingPrivatePermission(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	dish := createContentDish(t, db, owner, category, "dish-private-denied", contentModel.DishStatusDraft, false)
	const authorityID uint = 7002
	grantContentReadPermissions(t, db, authorityID)

	detail, err := contentServiceForTest(db).GetUserDishDetail(
		context.Background(),
		dish.PublicID,
		commonRequest.AdminActor{
			AdministratorID: 32,
			AuthorityID:     authorityID,
			RequestID:       "request-private-denied",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminNoPermission {
		t.Fatalf("error code = %d, want %d", appErrors.GetType(err), appErrors.AdminNoPermission)
	}
	if detail.ID != "" || detail.AccessAuditID != "" {
		t.Fatalf("private detail leaked on permission denial: %+v", detail)
	}
	var auditCount int64
	if err := db.Model(&commonModel.AdminAccessAudit{}).Count(&auditCount).Error; err != nil {
		t.Fatalf("count access audits: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("permission denial unexpectedly wrote %d access audits", auditCount)
	}
}

func TestContentPrivateDetailFailsClosedWhenAccessAuditWriteFails(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, false)
	owner, category := seedContentOwnerAndCategory(t, db)
	dish := createContentDish(t, db, owner, category, "dish-audit-failure", contentModel.DishStatusDraft, false)

	detail, err := contentServiceForTest(db).GetUserDishDetail(
		context.Background(),
		dish.PublicID,
		commonRequest.AdminActor{
			AdministratorID: 33,
			AuthorityID:     commonModel.AuthorityPlatformSuperAdmin,
			RequestID:       "request-audit-failure",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInternal {
		t.Fatalf("error code = %d, want %d", appErrors.GetType(err), appErrors.AdminInternal)
	}
	if detail.ID != "" || detail.AccessAuditID != "" || detail.CoverFileID != "" {
		t.Fatalf("private data returned after audit failure: %+v", detail)
	}
}

func TestUserRecipeDetailKeepsDishOrderAndReturnsAccessAuditID(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	firstDish := createContentDish(t, db, owner, category, "dish-first", contentModel.DishStatusUsable, true)
	secondDish := createContentDish(t, db, owner, category, "dish-second", contentModel.DishStatusUsable, true)
	recipe := createContentRecipe(t, db, owner, "recipe-ordered", firstDish, secondDish)
	if err := db.Model(&contentModel.RecipeDish{}).
		Where("recipe_id = ? AND dish_id = ?", recipe.ID, firstDish.ID).
		Update("sort_order", 20).Error; err != nil {
		t.Fatalf("move first relation: %v", err)
	}
	if err := db.Model(&contentModel.RecipeDish{}).
		Where("recipe_id = ? AND dish_id = ?", recipe.ID, secondDish.ID).
		Update("sort_order", 10).Error; err != nil {
		t.Fatalf("move second relation: %v", err)
	}

	detail, err := contentServiceForTest(db).GetUserRecipeDetail(
		context.Background(),
		recipe.PublicID,
		commonRequest.AdminActor{
			AdministratorID: 34,
			AuthorityID:     commonModel.AuthorityPlatformSuperAdmin,
			RequestID:       "request-recipe-order",
		},
	)
	if err != nil {
		t.Fatalf("get ordered recipe detail: %v", err)
	}
	if len(detail.Dishes) != 2 ||
		detail.Dishes[0].DishID != secondDish.PublicID ||
		detail.Dishes[1].DishID != firstDish.PublicID {
		t.Fatalf("recipe dishes out of order: %+v", detail.Dishes)
	}
	if detail.AccessAuditID == "" {
		t.Fatal("missing accessAuditId")
	}
	var audit commonModel.AdminAccessAudit
	if err := db.First(&audit, "id = ?", detail.AccessAuditID).Error; err != nil {
		t.Fatalf("load recipe access audit: %v", err)
	}
	if audit.Permission != PermissionUserRecipePrivateRead ||
		audit.TargetID != recipe.PublicID ||
		audit.RequestID != "request-recipe-order" {
		t.Fatalf("unexpected access audit: %+v", audit)
	}
	var operationAudit auditModel.AdminAuditLog
	if err := db.First(&operationAudit, "public_id = ?", detail.AccessAuditID).Error; err != nil {
		t.Fatalf("load recipe operation audit: %v", err)
	}
	if operationAudit.Action != "read_user_recipe_private_detail" ||
		operationAudit.TargetID != recipe.PublicID ||
		operationAudit.RequestID != "request-recipe-order" {
		t.Fatalf("unexpected recipe operation audit: %+v", operationAudit)
	}
}

// TestUserRecipeListFiltersOwnerKeywordAndUsesFirstUsableDishCover 验证菜谱用户筛选和首个可用菜品封面规则。
func TestUserRecipeListFiltersOwnerKeywordAndUsesFirstUsableDishCover(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	draftDish := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-recipe-draft-cover",
		contentModel.DishStatusDraft,
		false,
	)
	usableDish := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-recipe-usable-cover",
		contentModel.DishStatusUsable,
		true,
	)
	recipe := createContentRecipe(
		t,
		db,
		owner,
		"recipe-first-usable-cover",
		draftDish,
		usableDish,
	)
	const authorityID uint = 7005
	grantContentReadPermissions(t, db, authorityID)
	actor := commonRequest.AdminActor{
		AdministratorID: 36,
		AuthorityID:     authorityID,
		Username:        "recipe-admin",
		RequestID:       "request-recipe-owner-filter",
	}
	service := contentServiceForTest(db)
	page, err := service.ListUserRecipes(
		context.Background(),
		contentRequest.UserRecipeSearch{UserKeyword: "内容用"},
		actor,
	)
	if err != nil {
		t.Fatalf("list recipes by owner nickname: %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 || page.List[0].ID != recipe.PublicID {
		t.Fatalf("owner nickname filter result: %+v", page)
	}
	if page.List[0].CoverURL == nil || *page.List[0].CoverURL != *usableDish.CoverURL {
		t.Fatalf("recipe did not use first usable dish cover: %+v", page.List[0])
	}
	if page.List[0].UnavailableDishCount != 1 ||
		page.List[0].ContentState != "contains_unavailable" {
		t.Fatalf("recipe unavailable state mismatch: %+v", page.List[0])
	}
	unmatched, err := service.ListUserRecipes(
		context.Background(),
		contentRequest.UserRecipeSearch{UserKeyword: "不存在的用户"},
		actor,
	)
	if err != nil {
		t.Fatalf("list recipes by unmatched owner: %v", err)
	}
	if unmatched.Total != 0 || len(unmatched.List) != 0 {
		t.Fatalf("unmatched owner filter leaked recipes: %+v", unmatched)
	}
}

func setupRecipeGovernanceTest(
	t *testing.T,
	options ...ContentServiceOption,
) (
	*gorm.DB,
	*ContentService,
	contentModel.UserRecipe,
	[]contentModel.UserDish,
	contentRequest.GovernanceActionPreviewInput,
	commonRequest.AdminActor,
) {
	t.Helper()
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	firstDish := createContentDish(t, db, owner, category, "dish-governance-1", contentModel.DishStatusUsable, true)
	secondDish := createContentDish(t, db, owner, category, "dish-governance-2", contentModel.DishStatusUsable, true)
	recipe := createContentRecipe(t, db, owner, "recipe-governance", firstDish, secondDish)
	reference := contentModel.DishReference{
		PublicID:           "reference-history",
		DishID:             firstDish.ID,
		ReferenceType:      contentModel.ReferenceTypeConfirmedMealSnapshot,
		ObjectID:           "meal-history-001",
		ObjectLabel:        "历史饭局",
		ObjectStatus:       "completed",
		HistoricalSnapshot: true,
		OccurredAt:         time.Date(2026, time.July, 24, 11, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&reference).Error; err != nil {
		t.Fatalf("create historical reference: %v", err)
	}
	input := contentRequest.GovernanceActionPreviewInput{
		TargetType:      contentModel.GovernanceTargetRecipe,
		TargetID:        recipe.PublicID,
		Actions:         []string{contentModel.GovernanceActionSoftDeleteRecipe},
		ViolationType:   "content_violation",
		Severity:        contentModel.GovernanceSeverityNormal,
		Reason:          "内容不符合平台规范",
		ExpectedVersion: 1,
	}
	actor := commonRequest.AdminActor{
		AdministratorID:  35,
		AuthorityID:      commonModel.AuthorityPlatformSuperAdmin,
		Username:         "platform-admin",
		RequestID:        "request-governance-recipe",
		SourceIPMasked:   "192.168.0.0",
		UserAgentSummary: "content-governance-test",
	}
	return db, contentServiceForTest(db, options...), recipe, []contentModel.UserDish{firstDish, secondDish}, input, actor
}

func TestSoftDeleteRecipeDoesNotCascadeAndWritesNotificationAndAudit(t *testing.T) {
	db, service, recipe, dishes, input, actor := setupRecipeGovernanceTest(t)
	preview, err := service.PreviewGovernance(context.Background(), input, actor)
	if err != nil {
		t.Fatalf("preview recipe governance: %v", err)
	}
	executeInput := contentRequest.GovernanceActionExecuteInput{
		GovernanceActionPreviewInput: input,
		PreviewToken:                 preview.PreviewToken,
	}
	first, err := service.ExecuteGovernance(
		context.Background(),
		executeInput,
		actor,
		"governance-recipe-001",
	)
	if err != nil {
		t.Fatalf("execute recipe governance: %v", err)
	}
	if first.RecordID == "" || first.Status != "succeeded" || first.AffectedCount != 1 {
		t.Fatalf("unexpected governance result: %+v", first)
	}

	var storedRecipe contentModel.UserRecipe
	if err := db.Unscoped().First(&storedRecipe, recipe.ID).Error; err != nil {
		t.Fatalf("reload deleted recipe: %v", err)
	}
	if !storedRecipe.DeletedAt.Valid || storedRecipe.Version != 2 {
		t.Fatalf("recipe was not soft deleted with version increment: %+v", storedRecipe)
	}
	var dishCount int64
	if err := db.Model(&contentModel.UserDish{}).Where("id IN ?", []uint{dishes[0].ID, dishes[1].ID}).Count(&dishCount).Error; err != nil {
		t.Fatalf("count active dishes: %v", err)
	}
	if dishCount != 2 {
		t.Fatalf("soft deleting recipe affected %d dishes", 2-dishCount)
	}
	var relationCount int64
	if err := db.Model(&contentModel.RecipeDish{}).Where("recipe_id = ?", recipe.ID).Count(&relationCount).Error; err != nil {
		t.Fatalf("count recipe relations: %v", err)
	}
	if relationCount != 2 {
		t.Fatalf("recipe relations were cascaded: count=%d", relationCount)
	}
	var historyCount int64
	if err := db.Model(&contentModel.DishReference{}).
		Where("reference_type = ?", contentModel.ReferenceTypeConfirmedMealSnapshot).
		Count(&historyCount).Error; err != nil {
		t.Fatalf("count history snapshots: %v", err)
	}
	if historyCount != 1 {
		t.Fatalf("historical snapshot references changed: count=%d", historyCount)
	}

	var governanceRecord contentModel.GovernanceRecord
	if err := db.First(&governanceRecord, "public_id = ?", first.RecordID).Error; err != nil {
		t.Fatalf("load governance record: %v", err)
	}
	var notificationIDs []string
	if err := json.Unmarshal(governanceRecord.NotificationIDs, &notificationIDs); err != nil {
		t.Fatalf("decode notification ids: %v", err)
	}
	if len(notificationIDs) != 1 || notificationIDs[0] == "" {
		t.Fatalf("governance record lacks notification id: %s", governanceRecord.NotificationIDs)
	}
	var notification engagementModel.UserNotification
	if err := db.First(&notification, "id = ?", notificationIDs[0]).Error; err != nil {
		t.Fatalf("load governance notification: %v", err)
	}
	if notification.UserID != storedRecipe.OwnerID ||
		notification.Type != "governance" ||
		notification.TargetID == nil ||
		*notification.TargetID != recipe.PublicID {
		t.Fatalf("unexpected governance notification: %+v", notification)
	}
	var audit auditModel.AdminAuditLog
	if err := db.First(&audit, "target_type = ? AND target_id = ?", input.TargetType, input.TargetID).Error; err != nil {
		t.Fatalf("load governance audit: %v", err)
	}
	if audit.RequestID != actor.RequestID ||
		audit.IdempotencyKey == nil ||
		*audit.IdempotencyKey != "governance-recipe-001" ||
		len(audit.BeforeSummary) == 0 ||
		len(audit.AfterSummary) == 0 {
		t.Fatalf("unexpected mutation audit: %+v", audit)
	}

	replayed, err := service.ExecuteGovernance(
		context.Background(),
		executeInput,
		actor,
		"governance-recipe-001",
	)
	if err != nil {
		t.Fatalf("replay governance execution: %v", err)
	}
	if replayed.RecordID != first.RecordID {
		t.Fatalf("replayed record id = %s, want %s", replayed.RecordID, first.RecordID)
	}
	for name, model := range map[string]interface{}{
		"governance records": &contentModel.GovernanceRecord{},
		"mutation audits":    &auditModel.AdminAuditLog{},
		"notifications":      &engagementModel.UserNotification{},
	} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", name, err)
		}
		if count != 1 {
			t.Fatalf("%s count after replay = %d, want 1", name, count)
		}
	}
}

// TestSoftDeleteCheckinPreservesPointsAndWritesNotificationAndAudit 验证违规打卡软删除不会回退积分或删除流水。
func TestSoftDeleteCheckinPreservesPointsAndWritesNotificationAndAudit(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, false)
	now := time.Date(2026, time.July, 26, 11, 0, 0, 0, time.UTC)
	user := userModel.MiniAppUser{
		ID: "checkin-governance-user", OpenIDHash: "checkin-governance-openid-hash",
		OpenIDEncrypted: "checkin-governance-openid-encrypted", Nickname: "打卡治理用户",
		Status: userModel.UserStatusNormal, Points: 8, CheckinCount: 1,
		CheckinDayCount: 1, Version: 1, RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create checkin governance user: %v", err)
	}
	checkin := engagementModel.FrontCheckin{
		ID: "checkin-governance-target", UserID: user.ID, DishName: "违规图片打卡",
		ImageFileID: "checkin-governance-file", ImageURL: "https://example.invalid/checkin.png",
		CheckedDate: "2026-07-26", Rewarded: true, CheckedAt: now, Version: 1, CreatedAt: now,
	}
	if err := db.Create(&checkin).Error; err != nil {
		t.Fatalf("create governed checkin: %v", err)
	}
	objectType := "checkin"
	entry := engagementModel.FrontPointEntry{
		ID: "checkin-governance-points", UserID: user.ID, Type: engagementModel.PointEarned,
		Scene: "daily_checkin", Title: "做菜打卡奖励", Description: "每日首次做菜打卡",
		Amount: 1, BalanceAfter: user.Points, RelatedObjectType: &objectType,
		RelatedObjectID: &checkin.ID, CreatedAt: now,
	}
	if err := db.Create(&entry).Error; err != nil {
		t.Fatalf("create checkin point entry: %v", err)
	}
	input := contentRequest.GovernanceActionPreviewInput{
		TargetType: contentModel.GovernanceTargetCheckin, TargetID: checkin.ID,
		Actions:       []string{contentModel.GovernanceActionSoftDeleteCheckin},
		ViolationType: "image_violation", Severity: contentModel.GovernanceSeverityNormal,
		Reason: "打卡图片存在违规内容", ExpectedVersion: checkin.Version,
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 51, AuthorityID: commonModel.AuthorityPlatformSuperAdmin,
		Username: "platform-admin", RequestID: "request-checkin-governance",
	}
	service := contentServiceForTest(db)
	preview, err := service.PreviewGovernance(context.Background(), input, actor)
	if err != nil || preview.AffectedUserCount != 1 || preview.NotificationCount != 1 {
		t.Fatalf("preview checkin governance: result=%+v err=%v", preview, err)
	}
	execution, err := service.ExecuteGovernance(
		context.Background(),
		contentRequest.GovernanceActionExecuteInput{
			GovernanceActionPreviewInput: input,
			PreviewToken:                 preview.PreviewToken,
		},
		actor,
		"checkin-governance-key",
	)
	if err != nil || execution.Status != "succeeded" || execution.RecordID == "" {
		t.Fatalf("execute checkin governance: result=%+v err=%v", execution, err)
	}
	var storedCheckin engagementModel.FrontCheckin
	if err := db.Unscoped().First(&storedCheckin, "id = ?", checkin.ID).Error; err != nil {
		t.Fatalf("reload governed checkin: %v", err)
	}
	if !storedCheckin.DeletedAt.Valid || storedCheckin.Version != 2 {
		t.Fatalf("checkin was not soft deleted with version increment: %+v", storedCheckin)
	}
	var storedUser userModel.MiniAppUser
	if err := db.First(&storedUser, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("reload checkin user: %v", err)
	}
	if storedUser.Points != user.Points || storedUser.CheckinCount != 0 || storedUser.CheckinDayCount != 0 {
		t.Fatalf("unexpected governed checkin user counters: %+v", storedUser)
	}
	var pointCount int64
	if err := db.Model(&engagementModel.FrontPointEntry{}).
		Where("id = ?", entry.ID).Count(&pointCount).Error; err != nil || pointCount != 1 {
		t.Fatalf("checkin point entry was changed: count=%d err=%v", pointCount, err)
	}
	var notificationCount int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("target_type = ? AND target_id = ?", input.TargetType, input.TargetID).
		Count(&notificationCount).Error; err != nil || notificationCount != 1 {
		t.Fatalf("unexpected checkin governance notifications: count=%d err=%v", notificationCount, err)
	}
	var auditCount int64
	if err := db.Model(&auditModel.AdminAuditLog{}).
		Where("target_type = ? AND target_id = ?", input.TargetType, input.TargetID).
		Count(&auditCount).Error; err != nil || auditCount != 1 {
		t.Fatalf("unexpected checkin governance audits: count=%d err=%v", auditCount, err)
	}
}

type failingMutationAuditWriter struct{}

func (failingMutationAuditWriter) WriteMutationAudit(
	context.Context,
	*gorm.DB,
	*auditModel.AdminAuditLog,
) error {
	return appErrors.AdminInternal.DefaultMsg()
}

type failingGovernanceNotificationWriter struct{}

func (failingGovernanceNotificationWriter) WriteGovernanceNotification(
	context.Context,
	*gorm.DB,
	*engagementModel.UserNotification,
) (string, error) {
	return "", appErrors.AdminInternal.DefaultMsg()
}

func TestGovernanceWriterFailureRollsBackRecipeMutation(t *testing.T) {
	tests := []struct {
		name   string
		option ContentServiceOption
	}{
		{"notification failure", WithGovernanceNotificationWriter(failingGovernanceNotificationWriter{})},
		{"mutation audit failure", WithMutationAuditWriter(failingMutationAuditWriter{})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, service, recipe, dishes, input, actor := setupRecipeGovernanceTest(t, test.option)
			preview, err := service.PreviewGovernance(context.Background(), input, actor)
			if err != nil {
				t.Fatalf("preview governance: %v", err)
			}
			_, err = service.ExecuteGovernance(
				context.Background(),
				contentRequest.GovernanceActionExecuteInput{
					GovernanceActionPreviewInput: input,
					PreviewToken:                 preview.PreviewToken,
				},
				actor,
				"writer-failure-"+test.name,
			)
			if appErrors.GetType(err) != appErrors.AdminInternal {
				t.Fatalf("error code = %d, want %d", appErrors.GetType(err), appErrors.AdminInternal)
			}

			var storedRecipe contentModel.UserRecipe
			if err := db.Unscoped().First(&storedRecipe, recipe.ID).Error; err != nil {
				t.Fatalf("reload recipe after rollback: %v", err)
			}
			if storedRecipe.DeletedAt.Valid || storedRecipe.Version != 1 {
				t.Fatalf("recipe mutation was not rolled back: %+v", storedRecipe)
			}
			var activeDishCount int64
			if err := db.Model(&contentModel.UserDish{}).
				Where("id IN ?", []uint{dishes[0].ID, dishes[1].ID}).
				Count(&activeDishCount).Error; err != nil {
				t.Fatalf("count dishes after rollback: %v", err)
			}
			if activeDishCount != 2 {
				t.Fatalf("dish state changed after rollback: active=%d", activeDishCount)
			}
			for name, model := range map[string]interface{}{
				"governance records":  &contentModel.GovernanceRecord{},
				"mutation audits":     &auditModel.AdminAuditLog{},
				"notifications":       &engagementModel.UserNotification{},
				"idempotency records": &commonModel.AdminIdempotencyRecord{},
			} {
				var count int64
				if err := db.Model(model).Count(&count).Error; err != nil {
					t.Fatalf("count %s: %v", name, err)
				}
				if count != 0 {
					t.Fatalf("%s count after rollback = %d, want 0", name, count)
				}
			}
		})
	}
}

// TestCopyChainGovernanceRequiresConfirmationAndProcessesJobMySQL 验证复制链确认词和后台任务闭环。
func TestCopyChainGovernanceRequiresConfirmationAndProcessesJobMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	migrateContentTestTables(t, db, true)
	owner, category := seedContentOwnerAndCategory(t, db)
	now := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	copyOwner := userModel.MiniAppUser{
		ID:              "content-copy-user-001",
		OpenIDHash:      "content-copy-openid-hash",
		OpenIDEncrypted: "content-copy-openid-encrypted",
		Nickname:        "复制用户",
		Status:          userModel.UserStatusNormal,
		Version:         1,
		RegisteredAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&copyOwner).Error; err != nil {
		t.Fatalf("create copy owner: %v", err)
	}
	source := createContentDish(
		t,
		db,
		owner,
		category,
		"dish-copy-chain-source",
		contentModel.DishStatusUsable,
		true,
	)
	copyDish := createContentDish(
		t,
		db,
		copyOwner,
		category,
		"dish-copy-chain-child",
		contentModel.DishStatusUsable,
		false,
	)
	if err := db.Model(&copyDish).Updates(map[string]interface{}{
		"source_type":           contentModel.SourceTypeCreatorCopy,
		"direct_source_dish_id": source.ID,
		"root_source_dish_id":   source.ID,
		"source_locked":         true,
		"chain_depth":           1,
	}).Error; err != nil {
		t.Fatalf("link copy chain: %v", err)
	}
	recipe := createContentRecipe(t, db, copyOwner, "recipe-copy-chain", copyDish)
	input := contentRequest.GovernanceActionPreviewInput{
		TargetType:      contentModel.GovernanceTargetDish,
		TargetID:        source.PublicID,
		Actions:         []string{contentModel.GovernanceActionDeleteCopyChain},
		ViolationType:   "serious_content_violation",
		Severity:        contentModel.GovernanceSeveritySerious,
		Reason:          "源菜品存在严重违规内容",
		ExpectedVersion: source.Version,
	}
	actor := commonRequest.AdminActor{
		AdministratorID: 41,
		AuthorityID:     commonModel.AuthorityPlatformSuperAdmin,
		Username:        "platform-admin",
		RequestID:       "request-copy-chain-governance",
	}
	contentService := contentServiceForTest(db)
	preview, err := contentService.PreviewGovernance(context.Background(), input, actor)
	if err != nil {
		t.Fatalf("preview copy chain governance: %v", err)
	}
	if !preview.Asynchronous || preview.CopiedDishCount != 1 ||
		preview.AffectedUserCount != 2 || preview.NotificationCount != 2 {
		t.Fatalf("unexpected copy chain preview: %+v", preview)
	}
	if _, err := contentService.ExecuteGovernance(
		context.Background(),
		contentRequest.GovernanceActionExecuteInput{
			GovernanceActionPreviewInput: input,
			PreviewToken:                 preview.PreviewToken,
		},
		actor,
		"copy-chain-without-confirmation",
	); err == nil {
		t.Fatal("copy chain governance without confirmation unexpectedly succeeded")
	}
	execution, err := contentService.ExecuteGovernance(
		context.Background(),
		contentRequest.GovernanceActionExecuteInput{
			GovernanceActionPreviewInput: input,
			PreviewToken:                 preview.PreviewToken,
			ConfirmText:                  governanceCopyChainConfirmText,
		},
		actor,
		"copy-chain-with-confirmation",
	)
	if err != nil || execution.Status != contentModel.GovernanceJobPending ||
		execution.JobID == nil || execution.AffectedCount != 2 {
		t.Fatalf("execute copy chain governance: result=%+v err=%v", execution, err)
	}
	governanceService := NewGovernanceService(
		db,
		serviceCommon.NewPermissionService(db),
		serviceCommon.NewIdempotencyService(db),
	)
	if err := governanceService.ProcessPendingJobs(context.Background(), 100); err != nil {
		t.Fatalf("process copy chain governance job: %v", err)
	}
	for _, dish := range []contentModel.UserDish{source, copyDish} {
		var persisted contentModel.UserDish
		if err := db.Unscoped().First(&persisted, dish.ID).Error; err != nil {
			t.Fatalf("load governed dish %s: %v", dish.PublicID, err)
		}
		if !persisted.DeletedAt.Valid || persisted.Discoverable {
			t.Fatalf("governed dish remains active: %+v", persisted)
		}
	}
	var recipeDishCount int64
	if err := db.Model(&contentModel.RecipeDish{}).
		Where("recipe_id = ?", recipe.ID).
		Count(&recipeDishCount).Error; err != nil || recipeDishCount != 0 {
		t.Fatalf("copy dish recipe relation remains: count=%d err=%v", recipeDishCount, err)
	}
	var notificationCount int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("type = ?", "governance").
		Count(&notificationCount).Error; err != nil || notificationCount != 2 {
		t.Fatalf("unexpected copy chain notifications: count=%d err=%v", notificationCount, err)
	}
	job, err := governanceService.GetJob(context.Background(), *execution.JobID, actor)
	if err != nil || job.Status != contentModel.GovernanceJobSucceeded ||
		job.SucceededCount != 2 || job.FailedCount != 0 {
		t.Fatalf("unexpected copy chain job: job=%+v err=%v", job, err)
	}
	var record contentModel.GovernanceRecord
	if err := db.First(&record, "public_id = ?", execution.RecordID).Error; err != nil {
		t.Fatalf("load copy chain governance record: %v", err)
	}
	if record.JobStatus == nil || *record.JobStatus != contentModel.GovernanceJobSucceeded ||
		record.AffectedCount != 2 {
		t.Fatalf("copy chain governance record not synchronized: %+v", record)
	}
}

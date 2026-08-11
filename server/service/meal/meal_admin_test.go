package meal

import (
	"context"
	"encoding/json"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	mealRequest "github.com/dyjh/order-food-mini-app/server/model/meal/request"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/datatypes"
)

// TestMealAdminReadModelsMySQL 验证饭局和采购清单管理视图只依赖MySQL语义。
func TestMealAdminReadModelsMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&dishModel.ContentCategory{},
		&contentModel.UserDish{},
		&mealModel.FrontMeal{},
		&mealModel.MealParticipant{},
		&mealModel.FrontMealCandidate{},
		&mealModel.FrontMealVote{},
		&mealModel.FrontMealFinalDish{},
		&mealModel.FrontShoppingList{},
		&mealModel.FrontShoppingItem{},
	); err != nil {
		t.Fatalf("migrate meal admin tables: %v", err)
	}

	now := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	creator := userModel.MiniAppUser{
		ID: "meal-creator", OpenIDHash: "meal-creator-hash",
		OpenIDEncrypted: "meal-creator-encrypted", Nickname: "饭局发起人",
		Status: userModel.UserStatusNormal, Version: 1,
		RegisteredAt: now.Add(-24 * time.Hour), CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}
	member := userModel.MiniAppUser{
		ID: "meal-member", OpenIDHash: "meal-member-hash",
		OpenIDEncrypted: "meal-member-encrypted", Nickname: "饭局成员",
		Status: userModel.UserStatusNormal, Version: 1,
		RegisteredAt: now.Add(-12 * time.Hour), CreatedAt: now.Add(-12 * time.Hour),
		UpdatedAt: now,
	}
	if err := db.Create(&[]userModel.MiniAppUser{creator, member}).Error; err != nil {
		t.Fatalf("create meal users: %v", err)
	}
	category := dishModel.ContentCategory{
		PublicID: "meal-category", Name: "饭局测试分类", SortOrder: 1,
		Enabled: true, Version: 1,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create meal category: %v", err)
	}
	dish := contentModel.UserDish{
		PublicID: "meal-dish", OwnerID: creator.ID, CoverFileID: "meal-cover",
		Name: "番茄炒蛋", CategoryID: category.ID, Status: contentModel.DishStatusUsable,
		SourceType: contentModel.SourceTypeManual, Serving: 2,
		MediaReviewStatus: contentModel.MediaReviewPassed, Version: 1,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatalf("create meal dish: %v", err)
	}

	expiredMeal := mealModel.FrontMeal{
		ID: "expired-meal", CreatorID: creator.ID, Name: "超时饭局", Code: "EX1234",
		Status: mealModel.MealCollecting, DeadlineAt: now.Add(-time.Hour),
		CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour),
	}
	confirmedAt := now.Add(-30 * time.Minute)
	closedAt := now.Add(-45 * time.Minute)
	closeSource := mealModel.MealCloseSourceCreatorAction
	shoppingListID := "shopping-list-1"
	confirmedMeal := mealModel.FrontMeal{
		ID: "confirmed-meal", CreatorID: creator.ID, Name: "周末聚餐", Code: "CF1234",
		Status: mealModel.MealConfirmed, CloseReason: stringAddress("manual"),
		CloseSource: &closeSource, ClosedAt: &closedAt,
		DeadlineAt: now.Add(time.Hour), ShoppingListID: &shoppingListID,
		ConfirmedAt: &confirmedAt, CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: confirmedAt,
	}
	if err := db.Create(&[]mealModel.FrontMeal{expiredMeal, confirmedMeal}).Error; err != nil {
		t.Fatalf("create meals: %v", err)
	}
	members := []mealModel.MealParticipant{
		{ID: "meal-member-1", MealID: confirmedMeal.ID, UserID: creator.ID, JoinedAt: confirmedMeal.CreatedAt},
		{ID: "meal-member-2", MealID: confirmedMeal.ID, UserID: member.ID, JoinedAt: confirmedMeal.CreatedAt.Add(time.Minute)},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("create meal participants: %v", err)
	}
	candidate := mealModel.FrontMealCandidate{
		ID: "candidate-1", MealID: confirmedMeal.ID, DishID: dish.ID,
		Available: true, SortOrder: 1, CreatedAt: confirmedMeal.CreatedAt.Add(2 * time.Minute),
	}
	if err := db.Create(&candidate).Error; err != nil {
		t.Fatalf("create meal candidate: %v", err)
	}
	votes := []mealModel.FrontMealVote{
		{
			ID: "vote-1", MealID: confirmedMeal.ID, UserID: creator.ID,
			CandidateID: candidate.ID, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "vote-2", MealID: confirmedMeal.ID, UserID: member.ID,
			CandidateID: candidate.ID, CreatedAt: now, UpdatedAt: now,
		},
	}
	if err := db.Create(&votes).Error; err != nil {
		t.Fatalf("create meal votes: %v", err)
	}
	snapshot := mealModel.FrontMealFinalDish{
		ID: "snapshot-1", MealID: confirmedMeal.ID, CandidateID: candidate.ID,
		DishID: dish.ID, Name: dish.Name, FinalServings: 3,
		Ingredients: datatypes.JSON([]byte(`[{"name":"番茄","amount":"2个"}]`)),
		Steps:       datatypes.JSON([]byte(`[{"description":"翻炒"}]`)),
		CreatedAt:   confirmedAt,
	}
	if err := db.Create(&snapshot).Error; err != nil {
		t.Fatalf("create final dish snapshot: %v", err)
	}
	shareExpiresAt := now.Add(24 * time.Hour)
	shoppingList := mealModel.FrontShoppingList{
		ID: shoppingListID, MealID: confirmedMeal.ID, OwnerID: creator.ID,
		ShareTokenHash: "shopping-token-hash", ShareToken: "shopping-secret-token-123456",
		ShareExpiresAt: &shareExpiresAt, CreatedAt: confirmedAt, UpdatedAt: confirmedAt,
	}
	if err := db.Create(&shoppingList).Error; err != nil {
		t.Fatalf("create shopping list: %v", err)
	}
	sourceNames, err := json.Marshal([]string{dish.Name})
	if err != nil {
		t.Fatalf("encode shopping source names: %v", err)
	}
	shoppingItems := []mealModel.FrontShoppingItem{
		{
			ID: "shopping-item-1", ListID: shoppingList.ID, Name: "番茄", Amount: "2个",
			Completed: false, SourceDishNames: datatypes.JSON(sourceNames),
			SortOrder: 1, CreatedAt: confirmedAt, UpdatedAt: confirmedAt,
		},
		{
			ID: "shopping-item-2", ListID: shoppingList.ID, Name: "鸡蛋", Amount: "3个",
			Completed: true, SortOrder: 2, CreatedAt: confirmedAt, UpdatedAt: confirmedAt,
		},
	}
	if err := db.Create(&shoppingItems).Error; err != nil {
		t.Fatalf("create shopping items: %v", err)
	}

	permission := serviceCommon.NewPermissionService(db)
	mealService := NewMealService(db, permission)
	mealService.Now = func() time.Time { return now }
	shoppingListService := NewShoppingListService(db, permission)
	shoppingListService.Now = func() time.Time { return now }
	actor := commonRequest.AdminActor{
		AdministratorID: 1, AuthorityID: commonModel.AuthorityOrderFoodSuperAdmin,
		Username: "admin", RequestID: "meal-admin-request",
	}
	list, err := mealService.ListMeals(
		context.Background(),
		mealRequest.MealAdminListQuery{},
		actor,
	)
	if err != nil || list.Total != 2 || len(list.List) != 2 {
		t.Fatalf("list admin meals: result=%+v err=%v", list, err)
	}
	confirmedFinalDishCount := -1
	for index := range list.List {
		if list.List[index].ID == confirmedMeal.ID {
			confirmedFinalDishCount = list.List[index].FinalDishCount
			break
		}
	}
	if confirmedFinalDishCount != 1 {
		t.Fatalf("confirmed meal final dish count = %d, want 1", confirmedFinalDishCount)
	}
	joinedFrom := confirmedMeal.CreatedAt.Add(30 * time.Second)
	joinedTo := confirmedMeal.CreatedAt.Add(2 * time.Minute)
	joinedList, err := mealService.ListMeals(
		context.Background(),
		mealRequest.MealAdminListQuery{
			JoinedFrom: &joinedFrom,
			JoinedTo:   &joinedTo,
		},
		actor,
	)
	if err != nil || joinedList.Total != 1 || joinedList.List[0].ID != confirmedMeal.ID {
		t.Fatalf("joined time filter mismatch: result=%+v err=%v", joinedList, err)
	}
	confirmedFrom := confirmedAt.Add(-time.Minute)
	confirmedTo := confirmedAt.Add(time.Minute)
	confirmedList, err := mealService.ListMeals(
		context.Background(),
		mealRequest.MealAdminListQuery{
			ConfirmedFrom: &confirmedFrom,
			ConfirmedTo:   &confirmedTo,
		},
		actor,
	)
	if err != nil || confirmedList.Total != 1 || confirmedList.List[0].ID != confirmedMeal.ID {
		t.Fatalf("confirmed time filter mismatch: result=%+v err=%v", confirmedList, err)
	}
	var expiredAfterGuard mealModel.FrontMeal
	if err := db.First(&expiredAfterGuard, "id = ?", expiredMeal.ID).Error; err != nil ||
		expiredAfterGuard.Status != mealModel.MealCollecting ||
		expiredAfterGuard.ClosedAt != nil {
		t.Fatalf("admin read unexpectedly mutated an expired meal: meal=%+v err=%v", expiredAfterGuard, err)
	}
	detail, err := mealService.GetMeal(context.Background(), confirmedMeal.ID, actor)
	if err != nil || len(detail.Participants) != 2 || len(detail.Candidates) != 1 ||
		detail.Candidates[0].VoteCount != 2 || !detail.Candidates[0].Selected ||
		len(detail.FinalDishSnapshots) != 1 || detail.ShoppingListSummary == nil ||
		detail.CloseSource == nil ||
		*detail.CloseSource != string(mealModel.MealCloseSourceCreatorAction) {
		t.Fatalf("get admin meal detail: result=%+v err=%v", detail, err)
	}
	shoppingPage, err := shoppingListService.ListShoppingLists(
		context.Background(),
		mealRequest.ShoppingListAdminListQuery{ShareStatus: "active"},
		actor,
	)
	if err != nil || shoppingPage.Total != 1 || len(shoppingPage.List) != 1 ||
		shoppingPage.List[0].ID != shoppingList.ID ||
		shoppingPage.List[0].TotalCount != 2 {
		t.Fatalf("list active shopping lists: result=%+v err=%v", shoppingPage, err)
	}
	if _, err := mealService.ListMeals(
		context.Background(),
		mealRequest.MealAdminListQuery{SortOrder: "desc; DROP TABLE of_meals"},
		actor,
	); appErrors.Code(err) != int(appErrors.AdminBadRequest) {
		t.Fatalf("unsafe meal sort order error=%v", err)
	}
	shoppingDetail, err := shoppingListService.GetShoppingList(
		context.Background(),
		shoppingList.ID,
		actor,
	)
	if err != nil || shoppingDetail.TotalCount != 2 ||
		shoppingDetail.PendingCount != 1 ||
		shoppingDetail.CompletedCount != 1 || len(shoppingDetail.Items) != 2 ||
		len(shoppingDetail.GeneratedFromSnapshotIDs) != 1 ||
		len(shoppingDetail.Items[0].SourceDishNames) != 1 {
		t.Fatalf("get admin shopping detail: result=%+v err=%v", shoppingDetail, err)
	}
	if shoppingDetail.ShareTokenMasked == nil ||
		strings.Contains(*shoppingDetail.ShareTokenMasked, shoppingList.ShareToken) {
		t.Fatalf("shopping share token was not masked: %+v", shoppingDetail.ShareTokenMasked)
	}
	if _, err := shoppingListService.ListShoppingLists(
		context.Background(),
		mealRequest.ShoppingListAdminListQuery{
			SortOrder: "desc; DROP TABLE of_shopping_lists",
		},
		actor,
	); appErrors.Code(err) != int(appErrors.AdminBadRequest) {
		t.Fatalf("unsafe shopping sort order error=%v", err)
	}

	cancelledAt := now.Add(time.Minute)
	cancelReason := "creator_disabled"
	cancelledFrom := mealModel.MealConfirmed
	if err := db.Model(&mealModel.FrontMeal{}).
		Where("id = ?", confirmedMeal.ID).
		Updates(map[string]interface{}{
			"status": mealModel.MealCancelled, "cancelled_reason": cancelReason,
			"cancelled_from_status": cancelledFrom, "cancelled_at": cancelledAt,
		}).Error; err != nil {
		t.Fatalf("cancel meal for disabled creator: %v", err)
	}
	if err := db.Model(&mealModel.FrontShoppingList{}).
		Where("id = ?", shoppingList.ID).
		Updates(map[string]interface{}{
			"share_revoked_at": cancelledAt, "share_expires_at": cancelledAt,
		}).Error; err != nil {
		t.Fatalf("revoke disabled creator share: %v", err)
	}
	disabledDetail, err := mealService.GetMeal(context.Background(), confirmedMeal.ID, actor)
	if err != nil ||
		disabledDetail.CancelReason == nil ||
		*disabledDetail.CancelReason != "creator_disabled" ||
		disabledDetail.CancelledFromStatus == nil ||
		*disabledDetail.CancelledFromStatus != string(mealModel.MealConfirmed) ||
		disabledDetail.ShareRevocationStatus == nil ||
		*disabledDetail.ShareRevocationStatus != "revoked" ||
		len(disabledDetail.Participants) != 2 ||
		len(disabledDetail.FinalDishSnapshots) != 1 {
		t.Fatalf("disabled creator retention detail=%+v err=%v", disabledDetail, err)
	}
	revokedShoppingDetail, err := shoppingListService.GetShoppingList(
		context.Background(),
		shoppingList.ID,
		actor,
	)
	if err != nil ||
		revokedShoppingDetail.ShareStatus != "revoked" ||
		revokedShoppingDetail.ShareRevokedAt == nil ||
		!revokedShoppingDetail.ShareRevokedAt.Equal(cancelledAt) {
		t.Fatalf("revoked shopping detail=%+v err=%v", revokedShoppingDetail, err)
	}
	revokedPage, err := shoppingListService.ListShoppingLists(
		context.Background(),
		mealRequest.ShoppingListAdminListQuery{ShareStatus: "revoked"},
		actor,
	)
	if err != nil || revokedPage.Total != 1 || len(revokedPage.List) != 1 ||
		revokedPage.List[0].ID != shoppingList.ID {
		t.Fatalf("list revoked shopping lists: result=%+v err=%v", revokedPage, err)
	}
}

package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	"gorm.io/datatypes"
)

// TestMealAdminReadModelsMySQL 验证饭局和采购清单管理视图只依赖MySQL语义。
func TestMealAdminReadModelsMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.UserDish{},
		&orderfoodModel.FrontMeal{},
		&orderfoodModel.MealParticipant{},
		&orderfoodModel.FrontMealCandidate{},
		&orderfoodModel.FrontMealVote{},
		&orderfoodModel.FrontMealFinalDish{},
		&orderfoodModel.FrontShoppingList{},
		&orderfoodModel.FrontShoppingItem{},
	); err != nil {
		t.Fatalf("migrate meal admin tables: %v", err)
	}

	now := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	creator := orderfoodModel.MiniAppUser{
		ID: "meal-creator", OpenIDHash: "meal-creator-hash",
		OpenIDEncrypted: "meal-creator-encrypted", Nickname: "饭局发起人",
		Status: orderfoodModel.UserStatusNormal, Version: 1,
		RegisteredAt: now.Add(-24 * time.Hour), CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}
	member := orderfoodModel.MiniAppUser{
		ID: "meal-member", OpenIDHash: "meal-member-hash",
		OpenIDEncrypted: "meal-member-encrypted", Nickname: "饭局成员",
		Status: orderfoodModel.UserStatusNormal, Version: 1,
		RegisteredAt: now.Add(-12 * time.Hour), CreatedAt: now.Add(-12 * time.Hour),
		UpdatedAt: now,
	}
	if err := db.Create(&[]orderfoodModel.MiniAppUser{creator, member}).Error; err != nil {
		t.Fatalf("create meal users: %v", err)
	}
	category := orderfoodModel.ContentCategory{
		PublicID: "meal-category", Name: "饭局测试分类", SortOrder: 1,
		Enabled: true, Version: 1,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create meal category: %v", err)
	}
	dish := orderfoodModel.UserDish{
		PublicID: "meal-dish", OwnerID: creator.ID, CoverFileID: "meal-cover",
		Name: "番茄炒蛋", CategoryID: category.ID, Status: orderfoodModel.DishStatusUsable,
		SourceType: orderfoodModel.SourceTypeManual, Serving: 2,
		MediaReviewStatus: orderfoodModel.MediaReviewPassed, Version: 1,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatalf("create meal dish: %v", err)
	}

	expiredMeal := orderfoodModel.FrontMeal{
		ID: "expired-meal", CreatorID: creator.ID, Name: "超时饭局", Code: "EX1234",
		Status: orderfoodModel.MealCollecting, DeadlineAt: now.Add(-time.Hour),
		CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour),
	}
	confirmedAt := now.Add(-30 * time.Minute)
	closedAt := now.Add(-45 * time.Minute)
	closeSource := orderfoodModel.MealCloseSourceCreatorAction
	shoppingListID := "shopping-list-1"
	confirmedMeal := orderfoodModel.FrontMeal{
		ID: "confirmed-meal", CreatorID: creator.ID, Name: "周末聚餐", Code: "CF1234",
		Status: orderfoodModel.MealConfirmed, CloseReason: stringAddress("manual"),
		CloseSource: &closeSource, ClosedAt: &closedAt,
		DeadlineAt: now.Add(time.Hour), ShoppingListID: &shoppingListID,
		ConfirmedAt: &confirmedAt, CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: confirmedAt,
	}
	if err := db.Create(&[]orderfoodModel.FrontMeal{expiredMeal, confirmedMeal}).Error; err != nil {
		t.Fatalf("create meals: %v", err)
	}
	members := []orderfoodModel.MealParticipant{
		{ID: "meal-member-1", MealID: confirmedMeal.ID, UserID: creator.ID, JoinedAt: confirmedMeal.CreatedAt},
		{ID: "meal-member-2", MealID: confirmedMeal.ID, UserID: member.ID, JoinedAt: confirmedMeal.CreatedAt.Add(time.Minute)},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("create meal participants: %v", err)
	}
	candidate := orderfoodModel.FrontMealCandidate{
		ID: "candidate-1", MealID: confirmedMeal.ID, DishID: dish.ID,
		Available: true, SortOrder: 1, CreatedAt: confirmedMeal.CreatedAt.Add(2 * time.Minute),
	}
	if err := db.Create(&candidate).Error; err != nil {
		t.Fatalf("create meal candidate: %v", err)
	}
	votes := []orderfoodModel.FrontMealVote{
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
	snapshot := orderfoodModel.FrontMealFinalDish{
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
	shoppingList := orderfoodModel.FrontShoppingList{
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
	shoppingItems := []orderfoodModel.FrontShoppingItem{
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

	service := NewMealAdminService(db, NewPermissionService(db))
	service.Now = func() time.Time { return now }
	actor := orderfoodRequest.AdminActor{
		AdministratorID: 1, AuthorityID: orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username: "admin", RequestID: "meal-admin-request",
	}
	list, err := service.ListMeals(
		context.Background(),
		orderfoodRequest.MealAdminListQuery{},
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
	joinedList, err := service.ListMeals(
		context.Background(),
		orderfoodRequest.MealAdminListQuery{
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
	confirmedList, err := service.ListMeals(
		context.Background(),
		orderfoodRequest.MealAdminListQuery{
			ConfirmedFrom: &confirmedFrom,
			ConfirmedTo:   &confirmedTo,
		},
		actor,
	)
	if err != nil || confirmedList.Total != 1 || confirmedList.List[0].ID != confirmedMeal.ID {
		t.Fatalf("confirmed time filter mismatch: result=%+v err=%v", confirmedList, err)
	}
	var expiredAfterGuard orderfoodModel.FrontMeal
	if err := db.First(&expiredAfterGuard, "id = ?", expiredMeal.ID).Error; err != nil ||
		expiredAfterGuard.Status != orderfoodModel.MealCollecting ||
		expiredAfterGuard.ClosedAt != nil {
		t.Fatalf("admin read unexpectedly mutated an expired meal: meal=%+v err=%v", expiredAfterGuard, err)
	}
	detail, err := service.GetMeal(context.Background(), confirmedMeal.ID, actor)
	if err != nil || len(detail.Participants) != 2 || len(detail.Candidates) != 1 ||
		detail.Candidates[0].VoteCount != 2 || !detail.Candidates[0].Selected ||
		len(detail.FinalDishSnapshots) != 1 || detail.ShoppingListSummary == nil ||
		detail.CloseSource == nil ||
		*detail.CloseSource != string(orderfoodModel.MealCloseSourceCreatorAction) {
		t.Fatalf("get admin meal detail: result=%+v err=%v", detail, err)
	}
	shoppingPage, err := service.ListShoppingLists(
		context.Background(),
		orderfoodRequest.ShoppingListAdminListQuery{ShareStatus: "active"},
		actor,
	)
	if err != nil || shoppingPage.Total != 1 || len(shoppingPage.List) != 1 ||
		shoppingPage.List[0].ID != shoppingList.ID ||
		shoppingPage.List[0].TotalCount != 2 {
		t.Fatalf("list active shopping lists: result=%+v err=%v", shoppingPage, err)
	}
	if _, err := service.ListMeals(
		context.Background(),
		orderfoodRequest.MealAdminListQuery{SortOrder: "desc; DROP TABLE of_meals"},
		actor,
	); appErrors.Code(err) != int(appErrors.AdminBadRequest) {
		t.Fatalf("unsafe meal sort order error=%v", err)
	}
	shoppingDetail, err := service.GetShoppingList(
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
	if _, err := service.ListShoppingLists(
		context.Background(),
		orderfoodRequest.ShoppingListAdminListQuery{
			SortOrder: "desc; DROP TABLE of_shopping_lists",
		},
		actor,
	); appErrors.Code(err) != int(appErrors.AdminBadRequest) {
		t.Fatalf("unsafe shopping sort order error=%v", err)
	}

	cancelledAt := now.Add(time.Minute)
	cancelReason := "creator_disabled"
	cancelledFrom := orderfoodModel.MealConfirmed
	if err := db.Model(&orderfoodModel.FrontMeal{}).
		Where("id = ?", confirmedMeal.ID).
		Updates(map[string]interface{}{
			"status": orderfoodModel.MealCancelled, "cancelled_reason": cancelReason,
			"cancelled_from_status": cancelledFrom, "cancelled_at": cancelledAt,
		}).Error; err != nil {
		t.Fatalf("cancel meal for disabled creator: %v", err)
	}
	if err := db.Model(&orderfoodModel.FrontShoppingList{}).
		Where("id = ?", shoppingList.ID).
		Updates(map[string]interface{}{
			"share_revoked_at": cancelledAt, "share_expires_at": cancelledAt,
		}).Error; err != nil {
		t.Fatalf("revoke disabled creator share: %v", err)
	}
	disabledDetail, err := service.GetMeal(context.Background(), confirmedMeal.ID, actor)
	if err != nil ||
		disabledDetail.CancelReason == nil ||
		*disabledDetail.CancelReason != "creator_disabled" ||
		disabledDetail.CancelledFromStatus == nil ||
		*disabledDetail.CancelledFromStatus != string(orderfoodModel.MealConfirmed) ||
		disabledDetail.ShareRevocationStatus == nil ||
		*disabledDetail.ShareRevocationStatus != "revoked" ||
		len(disabledDetail.Participants) != 2 ||
		len(disabledDetail.FinalDishSnapshots) != 1 {
		t.Fatalf("disabled creator retention detail=%+v err=%v", disabledDetail, err)
	}
	revokedShoppingDetail, err := service.GetShoppingList(
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
	revokedPage, err := service.ListShoppingLists(
		context.Background(),
		orderfoodRequest.ShoppingListAdminListQuery{ShareStatus: "revoked"},
		actor,
	)
	if err != nil || revokedPage.Total != 1 || len(revokedPage.List) != 1 ||
		revokedPage.List[0].ID != shoppingList.ID {
		t.Fatalf("list revoked shopping lists: result=%+v err=%v", revokedPage, err)
	}
}

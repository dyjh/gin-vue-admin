package service

import (
	"context"
	"sync"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// mealTestDB 创建饭局链路测试所需的独立 MySQL 数据库。
func mealTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.ContentCategory{},
		&orderfoodModel.ContentTag{},
		&orderfoodModel.ContentUnit{},
		&orderfoodModel.UserDish{},
		&orderfoodModel.DishIngredient{},
		&orderfoodModel.DishStep{},
		&orderfoodModel.UserDishTag{},
		&orderfoodModel.UserRecipe{},
		&orderfoodModel.RecipeDish{},
		&orderfoodModel.FrontMeal{},
		&orderfoodModel.MealParticipant{},
		&orderfoodModel.FrontMealCandidate{},
		&orderfoodModel.FrontMealVote{},
		&orderfoodModel.FrontMealFinalDish{},
		&orderfoodModel.FrontShoppingList{},
		&orderfoodModel.FrontShoppingItem{},
		&orderfoodModel.UserNotification{},
		&orderfoodModel.SubscribeMessageTemplate{},
		&orderfoodModel.MealFinalResultSubscription{},
		&orderfoodModel.SubscribeMessageLog{},
		&orderfoodModel.SubscribeMessageAttempt{},
	); err != nil {
		t.Fatalf("migrate meal test database: %v", err)
	}
	return db
}

// seedMealUser 写入饭局测试用户。
func seedMealUser(t *testing.T, db *gorm.DB, id string, now time.Time) {
	t.Helper()
	if err := db.Create(&orderfoodModel.MiniAppUser{
		ID: id, OpenIDHash: "hash-" + id, OpenIDEncrypted: "encrypted-" + id,
		Nickname: id, Status: orderfoodModel.UserStatusNormal,
		RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed meal user %s: %v", id, err)
	}
}

// seedMealDish 写入可供饭局选择的用户菜品。
func seedMealDish(
	t *testing.T,
	db *gorm.DB,
	ownerID string,
	publicID string,
	now time.Time,
) orderfoodModel.UserDish {
	t.Helper()
	category := orderfoodModel.ContentCategory{
		PublicID: "category-" + publicID,
		Name:     "家常菜-" + publicID,
		Enabled:  true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("seed meal category: %v", err)
	}
	dish := orderfoodModel.UserDish{
		PublicID: publicID, OwnerID: ownerID, CoverFileID: "cover-" + publicID,
		Name: "测试菜品", CategoryID: category.ID, Status: orderfoodModel.DishStatusUsable,
		SourceType: orderfoodModel.SourceTypeManual, Serving: 2,
		MediaReviewStatus: orderfoodModel.MediaReviewPassed, Version: 1,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatalf("seed meal dish: %v", err)
	}
	return dish
}

func TestConcurrentMealCreationKeepsSingleActiveCreatorMeal(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "creator", now)
	seedMealDish(t, db, "creator", "dish-1", now)
	service := MealService{DB: db, Now: func() time.Time { return now }}
	input := frontRequest.MealCreateInput{
		Name: "并发饭局", DeadlineAt: now.Add(time.Hour).Format(time.RFC3339),
		CandidateDishIDs: []string{"dish-1"},
	}

	var wait sync.WaitGroup
	errorsByRequest := make([]error, 2)
	wait.Add(len(errorsByRequest))
	for index := range errorsByRequest {
		go func(current int) {
			defer wait.Done()
			_, errorsByRequest[current] = service.Create(context.Background(), "creator", input)
		}(index)
	}
	wait.Wait()

	successCount := 0
	conflictCount := 0
	for _, err := range errorsByRequest {
		if err == nil {
			successCount++
		} else if appErrors.GetType(err) == appErrors.FrontStateConflict {
			conflictCount++
		}
	}
	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("concurrent create results: success=%d conflict=%d errors=%v", successCount, conflictCount, errorsByRequest)
	}
	var activeCount int64
	if err := db.Model(&orderfoodModel.FrontMeal{}).
		Where("creator_id = ? AND status IN ?", "creator", []orderfoodModel.MealStatus{
			orderfoodModel.MealCollecting,
			orderfoodModel.MealClosed,
			orderfoodModel.MealConfirmed,
		}).
		Count(&activeCount).Error; err != nil {
		t.Fatalf("count active meals: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("active meals = %d, want 1", activeCount)
	}
}

// TestMealRemainsActiveThroughPurchaseUntilCreatorCompletes 验证饭局在关闭点餐和采购阶段继续占用唯一进行中名额。
func TestMealRemainsActiveThroughPurchaseUntilCreatorCompletes(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "creator", now)
	seedMealUser(t, db, "other-user", now)
	seedMealDish(t, db, "creator", "dish-1", now)
	service := MealService{DB: db, Now: func() time.Time { return now }}
	input := frontRequest.MealCreateInput{
		Name:             "周末饭局",
		DeadlineAt:       now.Add(time.Hour).Format(time.RFC3339),
		CandidateDishIDs: []string{"dish-1"},
	}

	created, err := service.Create(context.Background(), "creator", input)
	if err != nil {
		t.Fatalf("create meal: %v", err)
	}
	if _, err = service.Create(context.Background(), "creator", input); appErrors.GetType(err) != appErrors.FrontStateConflict {
		t.Fatalf("collecting meal did not block another meal: %v", err)
	}

	closed, err := service.Close(context.Background(), "creator", created.ID)
	if err != nil || closed.Status != string(orderfoodModel.MealClosed) {
		t.Fatalf("close meal: result=%+v err=%v", closed, err)
	}
	var closeNotificationCount int64
	if err := db.Model(&orderfoodModel.UserNotification{}).
		Where("target_id = ? AND title = ?", created.ID, "饭局已关闭点单").
		Count(&closeNotificationCount).Error; err != nil {
		t.Fatalf("count manual close notifications: %v", err)
	}
	if closeNotificationCount != 1 {
		t.Fatalf("manual close notification count = %d, want 1", closeNotificationCount)
	}
	var manuallyClosed orderfoodModel.FrontMeal
	if err := db.First(&manuallyClosed, "id = ?", created.ID).Error; err != nil ||
		manuallyClosed.CloseSource == nil ||
		*manuallyClosed.CloseSource != orderfoodModel.MealCloseSourceCreatorAction {
		t.Fatalf("manual close source: meal=%+v err=%v", manuallyClosed, err)
	}
	if _, err = service.Create(context.Background(), "creator", input); appErrors.GetType(err) != appErrors.FrontStateConflict {
		t.Fatalf("closed meal did not block another meal: %v", err)
	}

	var candidate orderfoodModel.FrontMealCandidate
	if err = db.First(&candidate, "meal_id = ?", created.ID).Error; err != nil {
		t.Fatalf("load meal candidate: %v", err)
	}
	selected := true
	confirmed, shoppingListID, err := service.Confirm(
		context.Background(),
		"creator",
		created.ID,
		frontRequest.MealConfirmInput{Dishes: []frontRequest.FinalMenuDishInput{{
			CandidateID:  candidate.ID,
			Selected:     &selected,
			FinalServing: 2,
		}}},
	)
	if err != nil || confirmed.Status != string(orderfoodModel.MealConfirmed) || shoppingListID == "" {
		t.Fatalf("confirm meal: result=%+v shoppingListID=%q err=%v", confirmed, shoppingListID, err)
	}
	if _, err = service.Create(context.Background(), "creator", input); appErrors.GetType(err) != appErrors.FrontStateConflict {
		t.Fatalf("confirmed meal did not block another meal: %v", err)
	}
	item, err := service.CreateShoppingItem(
		context.Background(),
		"creator",
		frontRequest.ShoppingItemCreateInput{Name: "青菜", Amount: "1 份"},
	)
	if err != nil {
		t.Fatalf("create shopping item before completion: %v", err)
	}
	if _, err = service.Cancel(context.Background(), "creator", created.ID, nil); appErrors.GetType(err) != appErrors.FrontStateConflict {
		t.Fatalf("confirmed meal was unexpectedly cancellable: %v", err)
	}
	if _, err = service.Complete(context.Background(), "other-user", created.ID); appErrors.GetType(err) != appErrors.FrontNoPermission {
		t.Fatalf("non-creator unexpectedly completed meal: %v", err)
	}

	completed, err := service.Complete(context.Background(), "creator", created.ID)
	if err != nil || completed.Status != string(orderfoodModel.MealCompleted) || completed.CompletedAt == nil {
		t.Fatalf("complete meal: result=%+v err=%v", completed, err)
	}
	renamed := "迟到的修改"
	if _, err = service.UpdateShoppingItem(
		context.Background(),
		"creator",
		item.ID,
		frontRequest.ShoppingItemUpdateInput{Name: &renamed},
	); appErrors.GetType(err) != appErrors.FrontNotFound {
		t.Fatalf("completed shopping item was unexpectedly editable: %v", err)
	}
	if err = service.DeleteShoppingItem(
		context.Background(),
		"creator",
		item.ID,
	); appErrors.GetType(err) != appErrors.FrontNotFound {
		t.Fatalf("completed shopping item was unexpectedly deletable: %v", err)
	}
	if _, err = service.CreateShoppingItem(
		context.Background(),
		"creator",
		frontRequest.ShoppingItemCreateInput{Name: "迟到的新增", Amount: "1 份"},
	); appErrors.GetType(err) != appErrors.FrontNotFound {
		t.Fatalf("completed shopping list unexpectedly accepted a new item: %v", err)
	}
	var storedItem orderfoodModel.FrontShoppingItem
	if err = db.First(&storedItem, "id = ?", item.ID).Error; err != nil {
		t.Fatalf("reload completed shopping item: %v", err)
	}
	if storedItem.Name != item.Name {
		t.Fatalf("completed shopping item changed to %q, want %q", storedItem.Name, item.Name)
	}
	nextMeal, err := service.Create(context.Background(), "creator", input)
	if err != nil || nextMeal.ID == created.ID {
		t.Fatalf("completed meal did not release creator slot: result=%+v err=%v", nextMeal, err)
	}
}

// TestDeadlineGuardCommitsClosureBeforeRejectingLateVote 验证截止后的点菜写入被拒绝且关闭状态和通知不会随错误回滚。
func TestDeadlineGuardCommitsClosureBeforeRejectingLateVote(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "deadline-creator", now)
	seedMealUser(t, db, "deadline-participant", now)
	dish := seedMealDish(t, db, "deadline-creator", "deadline-dish", now)
	meal := orderfoodModel.FrontMeal{
		ID: "deadline-meal", CreatorID: "deadline-creator", Name: "到期饭局",
		Code: "EXPIRE", Status: orderfoodModel.MealCollecting,
		DeadlineAt: now.Add(-time.Minute), CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	candidate := orderfoodModel.FrontMealCandidate{
		ID: "deadline-candidate", MealID: meal.ID, DishID: dish.ID,
		Available: true, SortOrder: 1, CreatedAt: now,
	}
	rows := []interface{}{
		&meal,
		&orderfoodModel.MealParticipant{
			ID: "deadline-member-creator", MealID: meal.ID,
			UserID: meal.CreatorID, JoinedAt: now,
		},
		&orderfoodModel.MealParticipant{
			ID: "deadline-member-participant", MealID: meal.ID,
			UserID: "deadline-participant", JoinedAt: now,
		},
		&candidate,
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed deadline guard case: %v", err)
		}
	}
	service := MealService{DB: db, Now: func() time.Time { return now }}

	if _, _, err := service.SaveVotes(
		context.Background(),
		"deadline-participant",
		meal.ID,
		[]string{candidate.ID},
	); appErrors.GetType(err) != appErrors.FrontStateConflict {
		t.Fatalf("late vote error = %v", err)
	}
	var stored orderfoodModel.FrontMeal
	if err := db.First(&stored, "id = ?", meal.ID).Error; err != nil {
		t.Fatalf("load deadline-closed meal: %v", err)
	}
	if stored.Status != orderfoodModel.MealClosed ||
		stored.CloseReason == nil || *stored.CloseReason != "deadline" ||
		stored.CloseSource == nil ||
		*stored.CloseSource != orderfoodModel.MealCloseSourceServiceGuard {
		t.Fatalf("deadline meal = %+v", stored)
	}
	var voteCount int64
	if err := db.Model(&orderfoodModel.FrontMealVote{}).
		Where("meal_id = ?", meal.ID).Count(&voteCount).Error; err != nil {
		t.Fatalf("count late votes: %v", err)
	}
	if voteCount != 0 {
		t.Fatalf("late vote count = %d, want 0", voteCount)
	}
	var notificationCount int64
	if err := db.Model(&orderfoodModel.UserNotification{}).
		Where("target_id = ? AND title = ?", meal.ID, "饭局已关闭点单").
		Count(&notificationCount).Error; err != nil {
		t.Fatalf("count deadline close notifications: %v", err)
	}
	if notificationCount != 2 {
		t.Fatalf("deadline close notification count = %d, want 2", notificationCount)
	}

	scannedMeal := orderfoodModel.FrontMeal{
		ID: "deadline-scan-meal", CreatorID: "deadline-creator", Name: "扫描到期饭局",
		Code: "SCAN01", Status: orderfoodModel.MealCollecting,
		DeadlineAt: now.Add(-30 * time.Second), CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	if err := db.Create(&scannedMeal).Error; err != nil {
		t.Fatalf("seed scanned meal: %v", err)
	}
	if err := db.Create(&orderfoodModel.MealParticipant{
		ID: "deadline-scan-member", MealID: scannedMeal.ID,
		UserID: scannedMeal.CreatorID, JoinedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed scanned meal participant: %v", err)
	}
	if err := service.ProcessExpiredMeals(context.Background(), 100); err != nil {
		t.Fatalf("process expired meals: %v", err)
	}
	var scannedStored orderfoodModel.FrontMeal
	if err := db.First(&scannedStored, "id = ?", scannedMeal.ID).Error; err != nil ||
		scannedStored.Status != orderfoodModel.MealClosed ||
		scannedStored.CloseSource == nil ||
		*scannedStored.CloseSource != orderfoodModel.MealCloseSourceMinuteScan {
		t.Fatalf("minute scan close source: meal=%+v err=%v", scannedStored, err)
	}
}

// TestCreatorRemovesCandidateByCandidateID 验证候选移除保留记录、隐藏点选并阻止移除最后一道候选菜。
func TestCreatorRemovesCandidateByCandidateID(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "remove-creator", now)
	seedMealUser(t, db, "remove-participant", now)
	firstDish := seedMealDish(t, db, "remove-creator", "remove-dish-1", now)
	secondDish := seedMealDish(t, db, "remove-creator", "remove-dish-2", now)
	meal := orderfoodModel.FrontMeal{
		ID: "remove-meal", CreatorID: "remove-creator", Name: "候选移除饭局",
		Code: "REMOVE", Status: orderfoodModel.MealCollecting,
		DeadlineAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
	}
	firstCandidate := orderfoodModel.FrontMealCandidate{
		ID: "remove-candidate-1", MealID: meal.ID, DishID: firstDish.ID,
		Available: true, SortOrder: 1, CreatedAt: now,
	}
	secondCandidate := orderfoodModel.FrontMealCandidate{
		ID: "remove-candidate-2", MealID: meal.ID, DishID: secondDish.ID,
		Available: true, SortOrder: 2, CreatedAt: now,
	}
	rows := []interface{}{
		&meal,
		&orderfoodModel.MealParticipant{
			ID: "remove-member-creator", MealID: meal.ID,
			UserID: meal.CreatorID, JoinedAt: now,
		},
		&orderfoodModel.MealParticipant{
			ID: "remove-member-participant", MealID: meal.ID,
			UserID: "remove-participant", JoinedAt: now,
		},
		&firstCandidate,
		&secondCandidate,
		&orderfoodModel.FrontMealVote{
			ID: "remove-vote", MealID: meal.ID, UserID: "remove-participant",
			CandidateID: firstCandidate.ID, CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed candidate removal case: %v", err)
		}
	}
	service := MealService{DB: db, Now: func() time.Time { return now }}

	if _, err := service.RemoveCandidate(
		context.Background(),
		"remove-participant",
		meal.ID,
		firstCandidate.ID,
	); appErrors.GetType(err) != appErrors.FrontNoPermission {
		t.Fatalf("participant candidate removal error=%v", err)
	}
	removedAt, err := service.RemoveCandidate(
		context.Background(),
		meal.CreatorID,
		meal.ID,
		firstCandidate.ID,
	)
	if err != nil || !removedAt.Equal(now) {
		t.Fatalf("remove candidate at=%v err=%v", removedAt, err)
	}
	var stored orderfoodModel.FrontMealCandidate
	if err := db.First(&stored, "id = ?", firstCandidate.ID).Error; err != nil {
		t.Fatalf("load removed candidate: %v", err)
	}
	if stored.Available || stored.UnavailableReason == nil ||
		*stored.UnavailableReason != "removed" {
		t.Fatalf("removed candidate state=%+v", stored)
	}
	summary, candidates, err := service.CandidateList(
		context.Background(),
		"remove-participant",
		meal.ID,
		"",
	)
	if err != nil || len(summary) != 1 || len(candidates) != 1 ||
		candidates[0].ID != secondCandidate.ID {
		t.Fatalf("visible candidates=%+v categories=%+v err=%v", candidates, summary, err)
	}
	detail, err := service.Meal(context.Background(), meal.CreatorID, meal.ID)
	if err != nil || detail.CandidateCount != 1 {
		t.Fatalf("meal candidate count=%d err=%v", detail.CandidateCount, err)
	}
	if _, err := service.RemoveCandidate(
		context.Background(),
		meal.CreatorID,
		meal.ID,
		secondCandidate.ID,
	); appErrors.GetType(err) != appErrors.FrontStateConflict {
		t.Fatalf("last candidate removal error=%v", err)
	}
	if _, _, err := service.SaveVotes(
		context.Background(),
		"remove-participant",
		meal.ID,
		[]string{secondCandidate.ID},
	); err != nil {
		t.Fatalf("save remaining candidate vote: %v", err)
	}
	var voteCount int64
	if err := db.Model(&orderfoodModel.FrontMealVote{}).
		Where("meal_id = ? AND user_id = ?", meal.ID, "remove-participant").
		Count(&voteCount).Error; err != nil {
		t.Fatalf("count retained removed candidate votes: %v", err)
	}
	if voteCount != 2 {
		t.Fatalf("retained and current candidate vote count=%d, want 2", voteCount)
	}
}

func TestDeleteDishRetainsInvalidCandidateAndExcludesVotes(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "creator", now)
	dish := seedMealDish(t, db, "creator", "dish-1", now)
	meal := orderfoodModel.FrontMeal{
		ID: "meal-1", CreatorID: "creator", Name: "晚饭", Code: "ABC123",
		Status: orderfoodModel.MealCollecting, DeadlineAt: now.Add(time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}
	candidate := orderfoodModel.FrontMealCandidate{
		ID: "candidate-1", MealID: meal.ID, DishID: dish.ID,
		Available: true, SortOrder: 1, CreatedAt: now,
	}
	rows := []interface{}{
		&meal,
		&orderfoodModel.MealParticipant{
			ID: "member-1", MealID: meal.ID, UserID: "creator", JoinedAt: now,
		},
		&candidate,
		&orderfoodModel.FrontMealVote{
			ID: "vote-1", MealID: meal.ID, UserID: "creator",
			CandidateID: candidate.ID, CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed invalid candidate case: %v", err)
		}
	}

	contentService := ContentService{DB: db, Now: func() time.Time { return now }}
	if _, _, err := contentService.DeleteDish(context.Background(), "creator", dish.PublicID); err != nil {
		t.Fatalf("delete dish: %v", err)
	}
	var persisted orderfoodModel.FrontMealCandidate
	if err := db.First(&persisted, "id = ?", candidate.ID).Error; err != nil {
		t.Fatalf("candidate was deleted instead of retained: %v", err)
	}
	if persisted.Available || persisted.UnavailableReason == nil || *persisted.UnavailableReason != "source_deleted" {
		t.Fatalf("unexpected retained candidate state: %+v", persisted)
	}

	mealService := MealService{DB: db, Now: func() time.Time { return now }}
	_, candidates, err := mealService.CandidateList(context.Background(), "creator", meal.ID, "")
	if err != nil {
		t.Fatalf("list retained candidates: %v", err)
	}
	if len(candidates) != 1 || candidates[0].Available ||
		candidates[0].VoteCount != 0 || candidates[0].SelectedByMe {
		t.Fatalf("invalid candidate still contributes votes: %+v", candidates)
	}
}

func TestFinalResultSubscriptionIsConsumedByFirstConfirmedResult(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "creator", now)
	seedMealUser(t, db, "participant", now)
	dish := seedMealDish(t, db, "creator", "dish-1", now)
	meal := orderfoodModel.FrontMeal{
		ID: "meal-1", CreatorID: "creator", Name: "晚饭", Code: "ABC123",
		Status: orderfoodModel.MealClosed, DeadlineAt: now.Add(time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}
	rows := []interface{}{
		&meal,
		&orderfoodModel.MealParticipant{
			ID: "member-creator", MealID: meal.ID, UserID: "creator", JoinedAt: now,
		},
		&orderfoodModel.MealParticipant{
			ID: "member-participant", MealID: meal.ID, UserID: "participant", JoinedAt: now,
		},
		&orderfoodModel.FrontMealCandidate{
			ID: "candidate-1", MealID: meal.ID, DishID: dish.ID,
			Available: true, SortOrder: 1, CreatedAt: now,
		},
		&orderfoodModel.SubscribeMessageTemplate{
			ID: "template-1", Name: "饭局结果", WechatTemplateID: "wechat-template-1",
			Scene: orderfoodModel.SubscribeSceneMealStatus, Purpose: "饭局最终结果",
			FieldMappings: datatypes.JSON([]byte(`{}`)), Enabled: true, Version: 1,
			CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed subscription case: %v", err)
		}
	}

	service := MealService{DB: db, Now: func() time.Time { return now }}
	recorded, err := service.RecordFinalResultSubscription(
		context.Background(),
		"participant",
		meal.ID,
		"wechat-template-1",
		orderfoodModel.MealAuthorizationAccept,
	)
	if err != nil || !recorded.Accepted {
		t.Fatalf("record final result subscription: result=%+v err=%v", recorded, err)
	}
	replayed, err := service.RecordFinalResultSubscription(
		context.Background(),
		"participant",
		meal.ID,
		"wechat-template-1",
		orderfoodModel.MealAuthorizationAccept,
	)
	if err != nil || !replayed.Accepted || !replayed.RecordedAt.Equal(recorded.RecordedAt) {
		t.Fatalf("reuse active final result subscription: result=%+v err=%v", replayed, err)
	}
	var activeSubscriptionCount int64
	if err := db.Model(&orderfoodModel.MealFinalResultSubscription{}).
		Where("meal_id = ? AND user_id = ? AND accepted = ? AND consumed_at IS NULL",
			meal.ID, "participant", true).
		Count(&activeSubscriptionCount).Error; err != nil {
		t.Fatalf("count active final result subscriptions: %v", err)
	}
	if activeSubscriptionCount != 1 {
		t.Fatalf("active final result subscriptions = %d, want 1", activeSubscriptionCount)
	}
	participantMeal, err := service.Meal(context.Background(), "participant", meal.ID)
	if err != nil || !participantMeal.FinalResultSubscriptionAccepted {
		t.Fatalf("meal subscription state: meal=%+v err=%v", participantMeal, err)
	}
	selected := true
	if _, _, err := service.Confirm(
		context.Background(),
		"creator",
		meal.ID,
		frontRequest.MealConfirmInput{Dishes: []frontRequest.FinalMenuDishInput{{
			CandidateID: "candidate-1", Selected: &selected, FinalServing: 2,
		}}},
	); err != nil {
		t.Fatalf("confirm subscribed meal: %v", err)
	}

	var subscription orderfoodModel.MealFinalResultSubscription
	var log orderfoodModel.SubscribeMessageLog
	var notification orderfoodModel.UserNotification
	if err := db.First(&subscription, "meal_id = ? AND user_id = ?", meal.ID, "participant").Error; err != nil {
		t.Fatalf("load consumed subscription: %v", err)
	}
	if err := db.First(&log, "subscription_id = ?", subscription.ID).Error; err != nil {
		t.Fatalf("load subscription send log: %v", err)
	}
	if err := db.First(&notification, "user_id = ? AND target_id = ?", "participant", meal.ID).Error; err != nil {
		t.Fatalf("load participant notification: %v", err)
	}
	if subscription.ConsumedAt == nil || subscription.ConsumedResult == nil ||
		*subscription.ConsumedResult != string(orderfoodModel.MealConfirmed) ||
		log.Status != orderfoodModel.SubscribeLogPending ||
		notification.SubscribeLogID == nil || *notification.SubscribeLogID != log.ID {
		t.Fatalf("subscription closure mismatch: subscription=%+v log=%+v notification=%+v", subscription, log, notification)
	}
}

// TestRetainedShoppingListIsParticipantReadOnly 验证发起人被禁用后只有已加入参与者可查看保留采购清单。
func TestRetainedShoppingListIsParticipantReadOnly(t *testing.T) {
	db := mealTestDB(t)
	now := time.Date(2026, 7, 26, 11, 0, 0, 0, time.UTC)
	seedMealUser(t, db, "disabled-creator", now)
	seedMealUser(t, db, "joined-participant", now)
	seedMealUser(t, db, "unrelated-user", now)
	cancelReason := "creator_disabled"
	cancelledFrom := orderfoodModel.MealConfirmed
	listID := "retained-shopping-list"
	meal := orderfoodModel.FrontMeal{
		ID: "retained-meal", CreatorID: "disabled-creator", Name: "保留饭局",
		Code: "RETAIN", Status: orderfoodModel.MealCancelled,
		CancelledReason: &cancelReason, CancelledFrom: &cancelledFrom,
		ShoppingListID: &listID, DeadlineAt: now.Add(-time.Hour),
		CancelledAt: &now, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now,
	}
	list := orderfoodModel.FrontShoppingList{
		ID: listID, MealID: meal.ID, OwnerID: meal.CreatorID,
		ShareToken: "revoked-share", ShareTokenHash: shareDigest("revoked-share"),
		ShareRevokedAt: &now, CreatedAt: now, UpdatedAt: now,
	}
	rows := []interface{}{
		&meal,
		&orderfoodModel.MealParticipant{
			ID: "retained-creator-member", MealID: meal.ID,
			UserID: meal.CreatorID, JoinedAt: now,
		},
		&orderfoodModel.MealParticipant{
			ID: "retained-joined-member", MealID: meal.ID,
			UserID: "joined-participant", JoinedAt: now,
		},
		&list,
		&orderfoodModel.FrontShoppingItem{
			ID: "retained-item", ListID: list.ID, Name: "番茄",
			Amount: "2个", SortOrder: 1, CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed retained shopping case: %v", err)
		}
	}
	service := MealService{DB: db, Now: func() time.Time { return now }}

	retained, err := service.ShoppingByID(
		context.Background(),
		"joined-participant",
		list.ID,
	)
	if err != nil || !retained.ReadOnly || len(retained.Items) != 1 {
		t.Fatalf("retained shopping result=%+v err=%v", retained, err)
	}
	detail, err := service.Meal(context.Background(), "joined-participant", meal.ID)
	if err != nil || !detail.ReadOnly ||
		detail.ReadOnlyReason == nil || *detail.ReadOnlyReason != cancelReason ||
		detail.CancelledFromStatus == nil ||
		*detail.CancelledFromStatus != string(orderfoodModel.MealConfirmed) {
		t.Fatalf("retained meal detail=%+v err=%v", detail, err)
	}
	if _, err := service.ShoppingByID(
		context.Background(),
		"unrelated-user",
		list.ID,
	); appErrors.GetType(err) != appErrors.FrontNoPermission {
		t.Fatalf("unrelated retained shopping error=%v", err)
	}
	if _, err := service.CurrentShopping(
		context.Background(),
		"disabled-creator",
	); appErrors.GetType(err) != appErrors.FrontNotFound {
		t.Fatalf("cancelled shopping unexpectedly remained current: %v", err)
	}

	completedMealID := "completed-meal"
	completedListID := "completed-shopping-list"
	completedAt := now.Add(-30 * time.Minute)
	completedRows := []interface{}{
		&orderfoodModel.FrontMeal{
			ID: completedMealID, CreatorID: "disabled-creator", Name: "已完成饭局",
			Code: "DONE01", Status: orderfoodModel.MealCompleted,
			ShoppingListID: &completedListID, DeadlineAt: now.Add(-2 * time.Hour),
			CompletedAt: &completedAt, CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: completedAt,
		},
		&orderfoodModel.MealParticipant{
			ID: "completed-creator-member", MealID: completedMealID,
			UserID: "disabled-creator", JoinedAt: now.Add(-3 * time.Hour),
		},
		&orderfoodModel.MealParticipant{
			ID: "completed-joined-member", MealID: completedMealID,
			UserID: "joined-participant", JoinedAt: now.Add(-2 * time.Hour),
		},
		&orderfoodModel.FrontShoppingList{
			ID: completedListID, MealID: completedMealID, OwnerID: "disabled-creator",
			ShareToken: "completed-share", ShareTokenHash: shareDigest("completed-share"),
			CreatedAt: now.Add(-time.Hour), UpdatedAt: completedAt,
		},
	}
	for _, row := range completedRows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed completed shopping case: %v", err)
		}
	}
	completedList, err := service.ShoppingByID(
		context.Background(),
		"joined-participant",
		completedListID,
	)
	if err != nil || !completedList.ReadOnly {
		t.Fatalf("completed participant shopping result=%+v err=%v", completedList, err)
	}
}

package user

import (
	"context"
	"encoding/json"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"reflect"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	userRequest "github.com/dyjh/order-food-mini-app/server/model/user/request"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func testMiniAppUser() userModel.MiniAppUser {
	now := time.Date(2026, time.July, 24, 8, 30, 0, 0, time.UTC)
	sessionKey := "encrypted-session-key"
	return userModel.MiniAppUser{
		ID:                  "user-001",
		OpenIDHash:          strings.Repeat("a", 64),
		OpenIDEncrypted:     "encrypted-openid",
		SessionKeyEncrypted: &sessionKey,
		Nickname:            "测试用户",
		Status:              userModel.UserStatusNormal,
		Version:             1,
		RegisteredAt:        now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func testAdminActor() commonRequest.AdminActor {
	return commonRequest.AdminActor{
		AdministratorID:  8,
		AuthorityID:      commonModel.AuthorityOrderFoodSuperAdmin,
		Username:         "orderfood-admin",
		RequestID:        "request-user-mutation",
		SourceIPMasked:   "192.168.0.0",
		UserAgentSummary: "user-service-test",
	}
}

func TestPreferenceProfileFailsClosedWhenAuditWriteFails(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&userModel.UserPreferenceProfile{},
		&userModel.PreferenceEvidenceAggregate{},
		&commonModel.AdminAccessAudit{},
	); err != nil {
		t.Fatalf("migrate profile tables: %v", err)
	}
	user := testMiniAppUser()
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	profile := userModel.UserPreferenceProfile{
		ID:            "profile-001",
		UserID:        user.ID,
		HasProfile:    true,
		UpdateState:   userModel.PreferenceUpdateActive,
		UpdateEnabled: true,
		ProfileJSON: datatypes.JSON(
			`{"commonDishTags":[],"frequentlyOrderedDishTags":[],"commonIngredients":[],"userSettings":[]}`,
		),
		Version:   1,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	service := NewUserService(db, serviceCommon.NewAccessAuditService(db), nil, nil)
	result, err := service.PreferenceProfile(
		context.Background(),
		user.ID,
		serviceCommon.SensitiveAccess{
			AdministratorID: 1,
			AuthorityID:     commonModel.AuthorityOrderFoodSuperAdmin,
			Permission:      PermissionUserPreferenceRead,
			RequestID:       "request-audit-failure",
		},
	)
	if err == nil {
		t.Fatal("expected audit write failure")
	}
	if appErrors.GetType(err) != appErrors.AdminInternal {
		t.Fatalf("error code = %d, want %d", appErrors.GetType(err), appErrors.AdminInternal)
	}
	encoded, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		t.Fatalf("marshal zero response: %v", marshalErr)
	}
	if strings.Contains(string(encoded), "测试用户") || result.AccessAuditID != "" {
		t.Fatalf("sensitive response was returned after audit failure: %s", encoded)
	}
	var accessAuditCount int64
	if err := db.Model(&commonModel.AdminAccessAudit{}).Count(&accessAuditCount).Error; err != nil {
		t.Fatalf("count rolled-back access audits: %v", err)
	}
	if accessAuditCount != 0 {
		t.Fatalf("access audit half-write count = %d, want 0", accessAuditCount)
	}
}

func TestUserDetailAndPreferenceNeverExposeWechatCredentials(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&userModel.UserPreferenceProfile{},
		&userModel.PreferenceEvidenceAggregate{},
		&commonModel.AdminAccessAudit{},
		&auditModel.AdminAuditLog{},
		&mealModel.FrontMeal{},
		&mealModel.MealParticipant{},
		&mealModel.FrontMealFinalDish{},
		&mealModel.FrontShoppingList{},
	); err != nil {
		t.Fatalf("migrate user tables: %v", err)
	}
	user := testMiniAppUser()
	user.CheckinDayCount = 9
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	service := NewUserService(db, serviceCommon.NewAccessAuditService(db), nil, nil)

	detail, err := service.Detail(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detail.WechatIdentityMasked == nil ||
		*detail.WechatIdentityMasked != "sha256:"+strings.Repeat("a", 12) {
		t.Fatalf("unexpected masked identity: %v", detail.WechatIdentityMasked)
	}
	profile, err := service.PreferenceProfile(
		context.Background(),
		user.ID,
		serviceCommon.SensitiveAccess{
			AdministratorID:       7,
			AdministratorUsername: "profile-admin",
			AuthorityID:           commonModel.AuthorityOrderFoodSuperAdmin,
			Permission:            PermissionUserPreferenceRead,
			RequestID:             "request-profile-success",
			SourceIPMasked:        "192.168.0.0",
		},
	)
	if err != nil {
		t.Fatalf("preference profile: %v", err)
	}
	if profile.AccessAuditID == "" {
		t.Fatal("missing accessAuditId")
	}

	encoded, err := json.Marshal(struct {
		Detail  interface{} `json:"detail"`
		Profile interface{} `json:"profile"`
	}{detail, profile})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	for _, forbidden := range []string{
		user.OpenIDHash,
		user.OpenIDEncrypted,
		*user.SessionKeyEncrypted,
		"openId",
		"sessionKey",
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("response contains forbidden credential %q: %s", forbidden, encoded)
		}
	}

	var audit commonModel.AdminAccessAudit
	if err := db.First(&audit, "id = ?", profile.AccessAuditID).Error; err != nil {
		t.Fatalf("get access audit: %v", err)
	}
	if audit.RequestID != "request-profile-success" ||
		audit.TargetID != user.ID ||
		audit.Permission != PermissionUserPreferenceRead {
		t.Fatalf("unexpected audit row: %+v", audit)
	}
	var operationAudit auditModel.AdminAuditLog
	if err := db.First(&operationAudit, "public_id = ?", profile.AccessAuditID).Error; err != nil {
		t.Fatalf("get sensitive operation audit: %v", err)
	}
	if operationAudit.Action != "read_preference_profile" ||
		operationAudit.TargetID != user.ID ||
		operationAudit.RequestID != "request-profile-success" ||
		operationAudit.AdministratorUsername != "profile-admin" {
		t.Fatalf("unexpected sensitive operation audit: %+v", operationAudit)
	}
}

func TestUserStatusOptimisticLockAndIdempotency(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&commonModel.AdminIdempotencyRecord{},
		&auditModel.AdminAuditLog{},
		&mealModel.FrontMeal{},
		&mealModel.MealParticipant{},
		&mealModel.FrontMealFinalDish{},
		&mealModel.FrontShoppingList{},
		&engagementModel.MealFinalResultSubscription{},
		&engagementModel.SubscribeMessageLog{},
		&engagementModel.UserNotification{},
	); err != nil {
		t.Fatalf("migrate mutation tables: %v", err)
	}
	user := testMiniAppUser()
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	participant := testMiniAppUser()
	participant.ID = "user-002"
	participant.OpenIDHash = strings.Repeat("b", 64)
	participant.OpenIDEncrypted = "encrypted-openid-002"
	participant.Nickname = "饭局参与者"
	if err := db.Create(&participant).Error; err != nil {
		t.Fatalf("create participant: %v", err)
	}
	now := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	listID := "shopping-disable-001"
	meal := mealModel.FrontMeal{
		ID: "meal-disable-001", CreatorID: user.ID, Name: "待收口饭局",
		Code: "654321", Status: mealModel.MealConfirmed,
		DeadlineAt: now.Add(time.Hour), ShoppingListID: &listID,
		ConfirmedAt: &now, CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	if err := db.Create(&meal).Error; err != nil {
		t.Fatalf("create active meal: %v", err)
	}
	if err := db.Create([]mealModel.MealParticipant{
		{ID: "member-creator", MealID: meal.ID, UserID: user.ID, JoinedAt: now.Add(-time.Hour)},
		{ID: "member-guest", MealID: meal.ID, UserID: participant.ID, JoinedAt: now},
	}).Error; err != nil {
		t.Fatalf("create meal participants: %v", err)
	}
	if err := db.Create(&mealModel.FrontMealFinalDish{
		ID: "final-disable-001", MealID: meal.ID, CandidateID: "candidate-disable-001",
		DishID: 101, Name: "保留菜品", FinalServings: 2,
		Ingredients: datatypes.JSON(`[]`), CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create final snapshot: %v", err)
	}
	shareExpiry := now.Add(24 * time.Hour)
	if err := db.Create(&mealModel.FrontShoppingList{
		ID: listID, MealID: meal.ID, OwnerID: user.ID,
		ShareTokenHash: strings.Repeat("c", 64), ShareToken: "share-disable-001",
		ShareExpiresAt: &shareExpiry, CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create shopping list: %v", err)
	}
	if err := db.Create(&engagementModel.MealFinalResultSubscription{
		ID: "meal-sub-disable-001", MealID: meal.ID, UserID: participant.ID,
		TemplateID: "template-meal-status", AuthorizationResult: "accept",
		Accepted: true, RecordedAt: now,
	}).Error; err != nil {
		t.Fatalf("create meal subscription: %v", err)
	}
	service := NewUserService(db, nil, serviceCommon.NewIdempotencyService(db), nil)
	service.Now = func() time.Time { return now }
	detail, err := service.Detail(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("load user disable impact: %v", err)
	}
	if detail.ActiveCreatedMealImpact == nil ||
		detail.ActiveCreatedMealImpact.MealID != meal.ID ||
		detail.ActiveCreatedMealImpact.ParticipantCount != 2 ||
		detail.ActiveCreatedMealImpact.FinalSnapshotCount != 1 ||
		detail.ActiveCreatedMealImpact.ShoppingListCount != 1 {
		t.Fatalf("unexpected user disable impact: %+v", detail.ActiveCreatedMealImpact)
	}
	body := userRequest.UpdateUserStatusBody{
		Status:          userModel.UserStatusDisabled,
		Reason:          "违反平台规则",
		ExpectedVersion: 1,
	}
	first, replayed, err := service.UpdateStatus(
		context.Background(),
		testAdminActor(),
		user.ID,
		"disable-user-001",
		body,
	)
	if err != nil {
		t.Fatalf("first status update: %v", err)
	}
	if replayed || first.Version != 2 || first.Status != string(userModel.UserStatusDisabled) {
		t.Fatalf("unexpected first result: %+v replayed=%v", first, replayed)
	}
	if first.CancelledMealCount != 1 ||
		first.NotifiedParticipantCount != 1 ||
		first.PreservedSnapshotCount != 1 ||
		first.PreservedShoppingListCount != 1 {
		t.Fatalf("unexpected disabled meal effects: %+v", first)
	}

	second, replayed, err := service.UpdateStatus(
		context.Background(),
		testAdminActor(),
		user.ID,
		"disable-user-001",
		body,
	)
	if err != nil {
		t.Fatalf("idempotent replay: %v", err)
	}
	if !replayed || !reflect.DeepEqual(second, first) {
		t.Fatalf("replay mismatch: first=%+v second=%+v replayed=%v", first, second, replayed)
	}

	conflictingBody := userRequest.UpdateUserStatusBody{
		Status:          userModel.UserStatusNormal,
		Reason:          "人工审核恢复",
		ExpectedVersion: 2,
	}
	_, _, err = service.UpdateStatus(
		context.Background(),
		testAdminActor(),
		user.ID,
		"disable-user-001",
		conflictingBody,
	)
	if appErrors.GetType(err) != appErrors.AdminIdempotencyConflict {
		t.Fatalf("idempotency conflict code = %d, want %d", appErrors.GetType(err), appErrors.AdminIdempotencyConflict)
	}

	staleBody := conflictingBody
	staleBody.ExpectedVersion = 1
	_, _, err = service.UpdateStatus(
		context.Background(),
		testAdminActor(),
		user.ID,
		"restore-user-stale",
		staleBody,
	)
	if appErrors.GetType(err) != appErrors.AdminStateConflict {
		t.Fatalf("stale version code = %d, want %d", appErrors.GetType(err), appErrors.AdminStateConflict)
	}

	var stored userModel.MiniAppUser
	if err := db.First(&stored, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if stored.Version != 2 || stored.Status != userModel.UserStatusDisabled {
		t.Fatalf("unexpected stored user: version=%d status=%s", stored.Version, stored.Status)
	}
	var cancelledMeal mealModel.FrontMeal
	if err := db.First(&cancelledMeal, "id = ?", meal.ID).Error; err != nil {
		t.Fatalf("reload cancelled meal: %v", err)
	}
	if cancelledMeal.Status != mealModel.MealCancelled ||
		cancelledMeal.CancelledReason == nil ||
		*cancelledMeal.CancelledReason != "creator_disabled" ||
		cancelledMeal.CancelledFrom == nil ||
		*cancelledMeal.CancelledFrom != mealModel.MealConfirmed {
		t.Fatalf("meal not closed for disabled creator: %+v", cancelledMeal)
	}
	var shoppingList mealModel.FrontShoppingList
	if err := db.First(&shoppingList, "id = ?", listID).Error; err != nil {
		t.Fatalf("reload shopping list: %v", err)
	}
	if shoppingList.ShareRevokedAt == nil {
		t.Fatal("shopping share was not revoked")
	}
	var notificationCount int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("user_id = ? AND target_id = ?", participant.ID, meal.ID).
		Count(&notificationCount).Error; err != nil {
		t.Fatalf("count participant notifications: %v", err)
	}
	if notificationCount != 1 {
		t.Fatalf("participant notification count = %d, want 1", notificationCount)
	}
	var subscription engagementModel.MealFinalResultSubscription
	if err := db.First(&subscription, "id = ?", "meal-sub-disable-001").Error; err != nil {
		t.Fatalf("reload meal subscription: %v", err)
	}
	if subscription.ConsumedResult == nil ||
		*subscription.ConsumedResult != string(mealModel.MealCancelled) ||
		subscription.ConsumedAt == nil {
		t.Fatalf("meal subscription not consumed: %+v", subscription)
	}
	var subscribeLogCount int64
	if err := db.Model(&engagementModel.SubscribeMessageLog{}).
		Count(&subscribeLogCount).Error; err != nil {
		t.Fatalf("count subscription logs: %v", err)
	}
	if subscribeLogCount != 0 {
		t.Fatalf("user disable created %d subscription logs", subscribeLogCount)
	}
	var audits []auditModel.AdminAuditLog
	if err := db.Find(&audits).Error; err != nil {
		t.Fatalf("list user mutation audits: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("audit count after replay/conflicts = %d, want 1", len(audits))
	}
	audit := audits[0]
	if audit.Action != "update_user_status" ||
		audit.TargetID != user.ID ||
		audit.Reason == nil ||
		*audit.Reason != body.Reason ||
		audit.IdempotencyKey == nil ||
		*audit.IdempotencyKey != "disable-user-001" ||
		len(audit.BeforeSummary) == 0 ||
		len(audit.AfterSummary) == 0 {
		t.Fatalf("unexpected user mutation audit: %+v", audit)
	}
}

func TestUserStatusAuditFailureRollsBackMutationAndIdempotency(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&commonModel.AdminIdempotencyRecord{},
		&mealModel.FrontMeal{},
		&mealModel.MealParticipant{},
		&mealModel.FrontMealFinalDish{},
		&mealModel.FrontShoppingList{},
		&engagementModel.MealFinalResultSubscription{},
		&engagementModel.UserNotification{},
	); err != nil {
		t.Fatalf("migrate rollback tables: %v", err)
	}
	user := testMiniAppUser()
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	_, _, err := NewUserService(db, nil, serviceCommon.NewIdempotencyService(db), nil).UpdateStatus(
		context.Background(),
		testAdminActor(),
		user.ID,
		"status-audit-failure",
		userRequest.UpdateUserStatusBody{
			Status:          userModel.UserStatusDisabled,
			Reason:          "审计写入失败",
			ExpectedVersion: 1,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInternal {
		t.Fatalf("audit failure code = %d, want %d", appErrors.GetType(err), appErrors.AdminInternal)
	}
	var stored userModel.MiniAppUser
	if err := db.First(&stored, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if stored.Status != userModel.UserStatusNormal || stored.Version != 1 {
		t.Fatalf("user mutation was not rolled back: status=%s version=%d", stored.Status, stored.Version)
	}
	var idempotencyCount int64
	if err := db.Model(&commonModel.AdminIdempotencyRecord{}).Count(&idempotencyCount).Error; err != nil {
		t.Fatalf("count idempotency records: %v", err)
	}
	if idempotencyCount != 0 {
		t.Fatalf("idempotency record count after rollback = %d, want 0", idempotencyCount)
	}
}

func TestPlatformCapabilityStateIsGlobal(t *testing.T) {
	user := testMiniAppUser()
	summary := toUserSummary(user, CapabilityPolicySnapshot{
		PlatformDefaultEnabled: true,
	})
	if summary.CapabilityEffective != "enabled" ||
		summary.CapabilitySource != "platform_default" {
		t.Fatalf("global enabled state mismatch: %+v", summary)
	}
	summary = toUserSummary(user, CapabilityPolicySnapshot{
		PlatformDefaultEnabled: false,
	})
	if summary.CapabilityEffective != "disabled" ||
		summary.CapabilitySource != "platform_default" {
		t.Fatalf("global disabled state mismatch: %+v", summary)
	}
}

// TestUserListFiltersShanghaiActivityDaysMySQL 验证活跃用户下钻按上海自然日筛选去重用户。
func TestUserListFiltersShanghaiActivityDaysMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&engagementModel.FrontUserActivityDay{},
	); err != nil {
		t.Fatalf("migrate user activity tables: %v", err)
	}
	now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)
	activeUser := testMiniAppUser()
	activeUser.ID = "active-user"
	activeUser.OpenIDHash = "active-user-hash"
	activeUser.OpenIDEncrypted = "active-user-encrypted"
	inactiveUser := testMiniAppUser()
	inactiveUser.ID = "inactive-user"
	inactiveUser.OpenIDHash = "inactive-user-hash"
	inactiveUser.OpenIDEncrypted = "inactive-user-encrypted"
	if err := db.Create(&[]userModel.MiniAppUser{
		activeUser,
		inactiveUser,
	}).Error; err != nil {
		t.Fatalf("create activity users: %v", err)
	}
	activity := engagementModel.FrontUserActivityDay{
		UserID: activeUser.ID, ActiveDate: "2026-07-26",
		FirstAt: now, CreatedAt: now,
	}
	if err := db.Create(&activity).Error; err != nil {
		t.Fatalf("create user activity day: %v", err)
	}
	list, err := NewUserService(db, nil, nil, StaticCapabilityPolicyProvider{}).List(
		context.Background(),
		userRequest.UserListQuery{
			ActiveFrom: "2026-07-26T00:00:00+08:00",
			ActiveTo:   "2026-07-26T23:59:59+08:00",
		},
	)
	if err != nil || list.Total != 1 || len(list.List) != 1 ||
		list.List[0].ID != activeUser.ID {
		t.Fatalf("user activity filter mismatch: result=%+v err=%v", list, err)
	}
}

func TestIdempotencyProcessingAndPayloadConflict(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(&commonModel.AdminIdempotencyRecord{}); err != nil {
		t.Fatalf("migrate idempotency table: %v", err)
	}
	payload := map[string]string{"userId": "user-001", "status": "disabled"}
	requestHash, err := serviceCommon.HashIdempotencyPayload(payload)
	if err != nil {
		t.Fatalf("hash payload: %v", err)
	}
	now := time.Now().UTC()
	record := commonModel.AdminIdempotencyRecord{
		AdministratorID: 8,
		EndpointID:      "user_status_update",
		IdempotencyKey:  "processing-user-001",
		RequestHash:     requestHash,
		State:           commonModel.AdminIdempotencyProcessing,
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       now.Add(time.Hour),
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create processing record: %v", err)
	}
	actionCalled := false
	_, _, err = serviceCommon.NewIdempotencyService(db).Execute(
		context.Background(),
		8,
		"user_status_update",
		"processing-user-001",
		payload,
		func(_ *gorm.DB) (interface{}, error) {
			actionCalled = true
			return nil, nil
		},
	)
	if appErrors.GetType(err) != appErrors.AdminRequestProcessing {
		t.Fatalf("processing code = %d, want %d", appErrors.GetType(err), appErrors.AdminRequestProcessing)
	}
	if actionCalled {
		t.Fatal("action ran while an identical request was processing")
	}

	differentPayload := map[string]string{"userId": "user-001", "status": "normal"}
	_, _, err = serviceCommon.NewIdempotencyService(db).Execute(
		context.Background(),
		8,
		"user_status_update",
		"processing-user-001",
		differentPayload,
		func(_ *gorm.DB) (interface{}, error) {
			return nil, nil
		},
	)
	if appErrors.GetType(err) != appErrors.AdminIdempotencyConflict {
		t.Fatalf("payload conflict code = %d, want %d", appErrors.GetType(err), appErrors.AdminIdempotencyConflict)
	}
}

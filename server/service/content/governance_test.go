package content

import (
	"context"
	"encoding/json"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
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
)

// TestGovernanceRecordsAndFailedItemRetryMySQL 验证违规记录查询和失败项重试只依赖MySQL。
func TestGovernanceRecordsAndFailedItemRetryMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&commonModel.AdminIdempotencyRecord{},
		&auditModel.AdminAuditLog{},
		&userModel.MiniAppUser{},
		&dishModel.ContentCategory{},
		&contentModel.UserDish{},
		&contentModel.RecipeDish{},
		&mealModel.FrontMeal{},
		&mealModel.FrontMealCandidate{},
		&dishModel.PlatformRecommendation{},
		&engagementModel.UserNotification{},
		&contentModel.GovernanceRecord{},
		&contentModel.GovernanceJob{},
		&contentModel.GovernanceJobItem{},
	); err != nil {
		t.Fatalf("migrate governance tables: %v", err)
	}
	now := time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC)
	user := userModel.MiniAppUser{
		ID: "gov-user", OpenIDHash: "gov-openid-hash", OpenIDEncrypted: "gov-openid-encrypted",
		Nickname: "治理测试用户", Status: userModel.UserStatusNormal,
		Version: 1, RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create governance user: %v", err)
	}
	category := dishModel.ContentCategory{
		PublicID: "gov-category", Name: "治理分类", SortOrder: 1, Enabled: true, Version: 1,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create governance category: %v", err)
	}
	dish := contentModel.UserDish{
		PublicID: "gov-copy-dish", OwnerID: user.ID, CoverFileID: "gov-cover",
		Name: "违规副本", CategoryID: category.ID, Status: contentModel.DishStatusUsable,
		Discoverable: true, SourceType: contentModel.SourceTypeCreatorCopy,
		SourceLocked: true, Serving: 1, MediaReviewStatus: contentModel.MediaReviewPassed,
		Version: 1,
	}
	if err := db.Create(&dish).Error; err != nil {
		t.Fatalf("create governed copy dish: %v", err)
	}
	actions, _ := json.Marshal([]string{contentModel.GovernanceActionDeleteCopyChain})
	impact, _ := json.Marshal(map[string]interface{}{
		"targetType": contentModel.GovernanceTargetDish,
		"targetId":   dish.PublicID, "copiedDishCount": 1,
	})
	emptyList, _ := json.Marshal([]string{})
	record := contentModel.GovernanceRecord{
		PublicID: "gov-record", TargetType: contentModel.GovernanceTargetDish,
		TargetID: dish.PublicID, TargetLabel: dish.Name, Actions: datatypes.JSON(actions),
		ViolationType: "serious_violation", Severity: contentModel.GovernanceSeveritySerious,
		Reason: "严重违规复制链处理", AdministratorID: 1,
		AdministratorUsername: "admin", JobStatus: stringPointer(contentModel.GovernanceJobFailed),
		ImpactSnapshot: datatypes.JSON(impact), BeforeSummary: datatypes.JSON([]byte(`{"deleted":false}`)),
		AfterSummary:    datatypes.JSON([]byte(`{"jobCreated":true}`)),
		NotificationIDs: datatypes.JSON(emptyList), RequestID: "gov-request",
		IdempotencyKey: "gov-execute-key", AffectedCount: 1,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create governance record: %v", err)
	}
	failedSummary := "首次处理失败"
	job := contentModel.GovernanceJob{
		ID: "gov-job", RecordID: record.PublicID, Status: contentModel.GovernanceJobFailed,
		TotalCount: 2, SucceededCount: 1, FailedCount: 1, LastErrorSummary: &failedSummary,
		Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("create governance job: %v", err)
	}
	item := contentModel.GovernanceJobItem{
		ID: "gov-item", JobID: job.ID, DishID: dish.ID, DishPublicID: dish.PublicID,
		UserID: user.ID, Status: contentModel.GovernanceJobItemFailed,
		AttemptCount: 1, LastErrorSummary: &failedSummary, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("create governance job item: %v", err)
	}
	existingNotificationID := "existing-notification"
	succeededItem := contentModel.GovernanceJobItem{
		ID: "gov-item-succeeded", JobID: job.ID, DishID: dish.ID + 1,
		DishPublicID: "already-processed-dish", UserID: user.ID,
		Status: contentModel.GovernanceJobItemSucceeded, AttemptCount: 1,
		NotificationID: &existingNotificationID, ProcessedAt: &now,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&succeededItem).Error; err != nil {
		t.Fatalf("create succeeded governance job item: %v", err)
	}

	service := NewGovernanceService(db, serviceCommon.NewPermissionService(db), serviceCommon.NewIdempotencyService(db))
	service.Now = func() time.Time { return now.Add(time.Hour) }
	actor := commonRequest.AdminActor{
		AdministratorID: 1, AuthorityID: commonModel.AuthorityOrderFoodSuperAdmin,
		Username: "admin", RequestID: "gov-retry-request",
	}
	list, err := service.ListRecords(context.Background(), contentRequest.GovernanceRecordListQuery{
		Action: contentModel.GovernanceActionDeleteCopyChain,
	}, actor)
	if err != nil || list.Total != 1 || len(list.List) != 1 {
		t.Fatalf("list governance records: result=%+v err=%v", list, err)
	}
	if list.List[0].AffectedCount != 1 || list.List[0].JobProgress == nil ||
		list.List[0].JobProgress.SucceededCount != 1 ||
		list.List[0].JobProgress.FailedCount != 1 {
		t.Fatalf("governance list progress = %+v", list.List[0])
	}
	abnormalList, err := service.ListRecords(
		context.Background(),
		contentRequest.GovernanceRecordListQuery{JobStatus: "abnormal"},
		actor,
	)
	if err != nil || abnormalList.Total != 1 || abnormalList.List[0].ID != record.PublicID {
		t.Fatalf("list abnormal governance records: result=%+v err=%v", abnormalList, err)
	}
	_, err = service.ListRecords(
		context.Background(),
		contentRequest.GovernanceRecordListQuery{SortOrder: "desc; drop table of_gov_jobs"},
		actor,
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe governance sort error = %v", err)
	}
	detail, err := service.GetRecord(context.Background(), record.PublicID, actor)
	if err != nil || detail.Job == nil || detail.Job.ID != job.ID ||
		len(detail.Job.Items) != 2 || !detail.TargetExists || detail.TargetVersion != 1 {
		t.Fatalf("get governance record detail: result=%+v err=%v", detail, err)
	}
	jobDetail, err := service.GetJob(context.Background(), job.ID, actor)
	if err != nil || len(jobDetail.Items) != 2 ||
		jobDetail.Items[0].ID != item.ID || jobDetail.Items[1].ID != succeededItem.ID {
		t.Fatalf("get governance job detail: result=%+v err=%v", jobDetail, err)
	}
	result, replayed, err := service.RetryJob(
		context.Background(), job.ID,
		contentRequest.GovernanceJobRetryInput{
			Reason: "修复任务依赖后重试", ExpectedVersion: 1,
		},
		actor, "gov-retry-key",
	)
	if err != nil || replayed || result.Status != contentModel.GovernanceJobSucceeded ||
		result.SucceededCount != 2 || result.FailedCount != 0 || result.Version != 2 ||
		len(result.Items) != 2 {
		t.Fatalf("retry governance job: result=%+v replayed=%v err=%v", result, replayed, err)
	}
	var unchangedSucceededItem contentModel.GovernanceJobItem
	if err := db.First(&unchangedSucceededItem, "id = ?", succeededItem.ID).Error; err != nil ||
		unchangedSucceededItem.AttemptCount != 1 ||
		unchangedSucceededItem.NotificationID == nil ||
		*unchangedSucceededItem.NotificationID != existingNotificationID {
		t.Fatalf("succeeded item was unexpectedly retried: item=%+v err=%v", unchangedSucceededItem, err)
	}
	var deleted contentModel.UserDish
	if err := db.Unscoped().First(&deleted, dish.ID).Error; err != nil || !deleted.DeletedAt.Valid {
		t.Fatalf("governed copy was not soft deleted: dish=%+v err=%v", deleted, err)
	}
	detail, err = service.GetRecord(context.Background(), record.PublicID, actor)
	if err != nil || detail.TargetExists {
		t.Fatalf("deleted governance target should not allow append processing: result=%+v err=%v", detail, err)
	}
	var notifications int64
	if err := db.Model(&engagementModel.UserNotification{}).
		Where("target_id = ?", dish.PublicID).Count(&notifications).Error; err != nil || notifications != 1 {
		t.Fatalf("unexpected governance notifications: count=%d err=%v", notifications, err)
	}
	replayedResult, replayed, err := service.RetryJob(
		context.Background(), job.ID,
		contentRequest.GovernanceJobRetryInput{
			Reason: "修复任务依赖后重试", ExpectedVersion: 1,
		},
		actor, "gov-retry-key",
	)
	if err != nil || !replayed || replayedResult.Version != 2 {
		t.Fatalf("replay governance retry: result=%+v replayed=%v err=%v", replayedResult, replayed, err)
	}
}

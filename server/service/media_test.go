package service

import (
	"context"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	"gorm.io/datatypes"
)

// TestModerationRecordSummaryIncludesUserObjectAndProviderRequestMySQL 验证图片审核列表和详情返回排障所需的关联信息。
func TestModerationRecordSummaryIncludesUserObjectAndProviderRequestMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.FrontMediaAsset{},
		&orderfoodModel.ImageModerationRecord{},
	); err != nil {
		t.Fatalf("migrate moderation record tables: %v", err)
	}
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	user := orderfoodModel.MiniAppUser{
		ID: "moderation-user", OpenIDHash: "moderation-openid-hash",
		OpenIDEncrypted: "moderation-openid-encrypted", Nickname: "审核用户",
		Status: orderfoodModel.UserStatusNormal, Version: 1,
		RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create moderation user: %v", err)
	}
	asset := orderfoodModel.FrontMediaAsset{
		ID: "moderation-file", UserID: user.ID, FileName: "cover.png",
		UploadSource: "user_upload", Scene: string(orderfoodModel.ModerationSceneDishCover),
		ResourceStatus: "active", URL: "https://example.invalid/cover.png",
		StoragePath: "/safe/moderation-file.png", ContentType: "image/png",
		Width: 100, Height: 100, SizeBytes: 128, Checksum: "checksum",
		ReviewStatus: string(orderfoodModel.ModerationStatusPassed), CreatedAt: now,
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("create moderation media: %v", err)
	}
	objectType := "dish"
	objectID := "dish-public-id"
	providerRequestID := "aliyun-request-id"
	record := orderfoodModel.ImageModerationRecord{
		ID: "moderation-record", RequestID: "request-id", FileID: asset.ID,
		Scene:             orderfoodModel.ModerationSceneDishCover,
		Status:            orderfoodModel.ModerationStatusPassed,
		RiskLabelsJSON:    datatypes.JSON([]byte(`[]`)),
		ProviderRequestID: &providerRequestID, ObjectType: &objectType,
		ObjectID: &objectID, ProviderSummaryJSON: datatypes.JSON([]byte(`{"risk":"none"}`)),
		CreatedAt: now,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create moderation record: %v", err)
	}
	service := NewMediaService(db)
	list, err := service.ListModerationRecords(
		context.Background(),
		orderfoodRequest.ModerationRecordListQuery{UserID: user.ID, ObjectType: objectType},
	)
	if err != nil || list.Total != 1 || len(list.List) != 1 {
		t.Fatalf("list moderation records: result=%+v err=%v", list, err)
	}
	summary := list.List[0]
	if summary.User == nil || summary.User.ID != user.ID ||
		summary.ObjectType == nil || *summary.ObjectType != objectType ||
		summary.ObjectID == nil || *summary.ObjectID != objectID ||
		summary.ProviderRequestID == nil || *summary.ProviderRequestID != providerRequestID {
		t.Fatalf("moderation summary association mismatch: %+v", summary)
	}
	detail, err := service.GetModerationRecord(context.Background(), record.ID, false)
	if err != nil || detail.User == nil || detail.User.ID != user.ID ||
		detail.ProviderRequestID == nil || *detail.ProviderRequestID != providerRequestID ||
		detail.ProviderResponseSummary != nil {
		t.Fatalf("moderation detail association mismatch: result=%+v err=%v", detail, err)
	}
	sensitiveDetail, err := service.GetModerationRecord(context.Background(), record.ID, true)
	if err != nil || sensitiveDetail.ProviderResponseSummary["risk"] != "none" {
		t.Fatalf("moderation sensitive summary mismatch: result=%+v err=%v", sensitiveDetail, err)
	}
	rejectedRecord := record
	rejectedRecord.ID = "moderation-record-rejected"
	rejectedRecord.RequestID = "request-id-rejected"
	rejectedRecord.Status = orderfoodModel.ModerationStatusRejected
	failedRecord := record
	failedRecord.ID = "moderation-record-failed"
	failedRecord.RequestID = "request-id-failed"
	failedRecord.Status = orderfoodModel.ModerationStatusFailed
	if err := db.Create(&[]orderfoodModel.ImageModerationRecord{
		rejectedRecord,
		failedRecord,
	}).Error; err != nil {
		t.Fatalf("create rejected and failed moderation records: %v", err)
	}
	abnormalList, err := service.ListModerationRecords(
		context.Background(),
		orderfoodRequest.ModerationRecordListQuery{Status: "rejected_or_failed"},
	)
	if err != nil || abnormalList.Total != 2 || len(abnormalList.List) != 2 {
		t.Fatalf("moderation rejected-or-failed filter mismatch: result=%+v err=%v", abnormalList, err)
	}
	_, err = service.ListModerationRecords(
		context.Background(),
		orderfoodRequest.ModerationRecordListQuery{
			SortOrder: "desc; drop table of_mod_records",
		},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe moderation sort error = %v", err)
	}
}

// TestMediaSummaryPreservesDeletedCheckinBindingMySQL 验证打卡软删除后图片仍保留治理目标关联。
func TestMediaSummaryPreservesDeletedCheckinBindingMySQL(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(
		&orderfoodModel.MiniAppUser{},
		&orderfoodModel.FrontMediaAsset{},
		&orderfoodModel.OfficialDish{},
		&orderfoodModel.UserDish{},
		&orderfoodModel.FrontCheckin{},
		&orderfoodModel.ImageModerationRecord{},
	); err != nil {
		t.Fatalf("migrate checkin media tables: %v", err)
	}
	now := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	user := orderfoodModel.MiniAppUser{
		ID: "checkin-media-user", OpenIDHash: "checkin-media-openid-hash",
		OpenIDEncrypted: "checkin-media-openid-encrypted", Nickname: "打卡图片用户",
		Status: orderfoodModel.UserStatusNormal, Version: 1,
		RegisteredAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create checkin media user: %v", err)
	}
	asset := orderfoodModel.FrontMediaAsset{
		ID: "checkin-media-file", UserID: user.ID, FileName: "checkin.png",
		UploadSource: "user_upload", Scene: string(orderfoodModel.ModerationSceneCheckin),
		ResourceStatus: "active", URL: "https://example.invalid/checkin.png",
		StoragePath: "/safe/checkin-media-file.png", ContentType: "image/png",
		Width: 100, Height: 100, SizeBytes: 128, Checksum: "checkin-checksum",
		ReviewStatus: string(orderfoodModel.ModerationStatusPassed), CreatedAt: now,
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("create checkin media: %v", err)
	}
	checkin := orderfoodModel.FrontCheckin{
		ID: "checkin-media-target", UserID: user.ID, DishName: "测试打卡",
		ImageFileID: asset.ID, ImageURL: asset.URL, CheckedDate: "2026-07-26",
		CheckedAt: now, Version: 1, CreatedAt: now,
	}
	if err := db.Create(&checkin).Error; err != nil {
		t.Fatalf("create checkin media binding: %v", err)
	}
	service := NewMediaService(db)
	detail, err := service.GetMedia(context.Background(), asset.ID)
	if err != nil || detail.BoundObjectType == nil || *detail.BoundObjectType != "checkin" ||
		detail.BoundObjectID == nil || *detail.BoundObjectID != checkin.ID ||
		detail.BoundObjectVersion == nil || *detail.BoundObjectVersion != 1 ||
		detail.BoundObjectDeleted {
		t.Fatalf("active checkin media binding mismatch: result=%+v err=%v", detail, err)
	}
	list, err := service.ListMedia(
		context.Background(),
		orderfoodRequest.MediaListQuery{
			BoundObjectType: "checkin", BoundObjectID: checkin.ID,
		},
	)
	if err != nil || list.Total != 1 || len(list.List) != 1 ||
		list.List[0].Width == nil || *list.List[0].Width != asset.Width ||
		list.List[0].Height == nil || *list.List[0].Height != asset.Height {
		t.Fatalf("checkin media list dimensions mismatch: result=%+v err=%v", list, err)
	}
	_, err = service.ListMedia(
		context.Background(),
		orderfoodRequest.MediaListQuery{SortOrder: "desc; drop table of_media"},
	)
	if appErrors.GetType(err) != appErrors.AdminBadRequest {
		t.Fatalf("unsafe media sort error = %v", err)
	}
	if err := db.Delete(&checkin).Error; err != nil {
		t.Fatalf("soft delete checkin media binding: %v", err)
	}
	detail, err = service.GetMedia(context.Background(), asset.ID)
	if err != nil || detail.BoundObjectType == nil || *detail.BoundObjectType != "checkin" ||
		detail.BoundObjectID == nil || *detail.BoundObjectID != checkin.ID ||
		!detail.BoundObjectDeleted {
		t.Fatalf("deleted checkin media binding mismatch: result=%+v err=%v", detail, err)
	}

	category := orderfoodModel.ContentCategory{
		PublicID: "official-media-category", Name: "官方图片分类",
		SortOrder: 1, Enabled: true, Version: 1,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create official media category: %v", err)
	}
	officialAsset := orderfoodModel.FrontMediaAsset{
		ID: "official-media-file", FileName: "official.png",
		UploadSource:   "admin_upload",
		Scene:          string(orderfoodModel.ModerationSceneOfficialCover),
		ResourceStatus: "active", URL: "https://example.invalid/official.png",
		StoragePath: "/safe/official-media-file.png", ContentType: "image/png",
		Width: 300, Height: 200, SizeBytes: 256, Checksum: "official-checksum",
		ReviewStatus: string(orderfoodModel.ModerationStatusNotRequired), CreatedAt: now,
	}
	if err := db.Create(&officialAsset).Error; err != nil {
		t.Fatalf("create official media asset: %v", err)
	}
	officialDish := orderfoodModel.OfficialDish{
		PublicID: "official-media-dish", Name: "官方图片菜品",
		CoverFileID: officialAsset.ID, CoverURL: officialAsset.URL,
		CategoryID: category.ID, Serving: 1, Status: "usable",
		Version: 3, CreatedByID: 1, CreatedByUsername: "admin",
		UpdatedByID: 1, UpdatedByUsername: "admin",
	}
	if err := db.Create(&officialDish).Error; err != nil {
		t.Fatalf("create official media dish: %v", err)
	}
	officialDetail, err := service.GetMedia(context.Background(), officialAsset.ID)
	if err != nil || officialDetail.BoundObjectType == nil ||
		*officialDetail.BoundObjectType != "official_dish" ||
		officialDetail.BoundObjectVersion == nil ||
		*officialDetail.BoundObjectVersion != int(officialDish.Version) ||
		officialDetail.BoundObjectDeleted {
		t.Fatalf("official media governance binding mismatch: result=%+v err=%v", officialDetail, err)
	}
}

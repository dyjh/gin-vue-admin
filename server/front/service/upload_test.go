package service

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

var uploadTestTinyPNG = func() []byte {
	decoded, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZlV8AAAAASUVORK5CYII=",
	)
	if err != nil {
		panic(err)
	}
	return decoded
}()

// openUploadTestDB 创建图片上传测试所需的独立 MySQL 数据库。
func openUploadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.FrontMediaAsset{},
		&orderfoodModel.ImageModerationRecord{},
	); err != nil {
		t.Fatalf("migrate upload schema: %v", err)
	}
	return db
}

// TestGeneratedImageUploadUsesAIGeneratedSource 验证 AI 生成图片免上传审核且记录正确来源。
func TestGeneratedImageUploadUsesAIGeneratedSource(t *testing.T) {
	db := openUploadTestDB(t)
	now := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)
	moderation := orderfoodService.NewModerationService(db, nil, nil, nil, nil)
	moderation.Now = func() time.Time { return now }
	service := &UploadService{
		DB:         db,
		Moderation: moderation,
		Now:        func() time.Time { return now },
	}

	originalLocal := global.GVA_CONFIG.Local
	global.GVA_CONFIG.Local.StorePath = t.TempDir()
	global.GVA_CONFIG.Local.Path = "/uploads"
	t.Cleanup(func() {
		global.GVA_CONFIG.Local = originalLocal
	})

	result, err := service.UploadGeneratedImage(
		context.Background(),
		"user-generated-cover",
		"image/png",
		uploadTestTinyPNG,
		"request-generated-cover",
	)
	if err != nil {
		t.Fatalf("upload generated image: %v", err)
	}
	if result.ReviewStatus != orderfoodModel.MediaReviewNotRequired {
		t.Fatalf("review status = %q, want %q", result.ReviewStatus, orderfoodModel.MediaReviewNotRequired)
	}

	var asset orderfoodModel.FrontMediaAsset
	if err := db.First(&asset, "id = ?", result.FileID).Error; err != nil {
		t.Fatalf("load generated asset: %v", err)
	}
	if asset.UploadSource != frontUploadSourceAI ||
		asset.ResourceStatus != frontMediaResourceActive ||
		asset.Scene != string(orderfoodModel.ModerationSceneGeneratedCover) {
		t.Fatalf("generated asset = %+v", asset)
	}

	var record orderfoodModel.ImageModerationRecord
	if err := db.First(&record, "file_id = ?", result.FileID).Error; err != nil {
		t.Fatalf("load generated moderation record: %v", err)
	}
	if record.Status != orderfoodModel.ModerationStatusNotRequired {
		t.Fatalf("moderation status = %q, want %q", record.Status, orderfoodModel.ModerationStatusNotRequired)
	}
}

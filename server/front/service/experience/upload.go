package experience

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	moderationService "github.com/dyjh/order-food-mini-app/server/service/content"
	"gorm.io/gorm"
)

const (
	maxFrontImageBytes       = 10 << 20
	frontUploadSourceUser    = "user_upload"
	frontUploadSourceAI      = "ai_generated"
	frontMediaResourceActive = "active"
)

var allowedUploadScenes = map[string]contentModel.ModerationScene{
	"profile_avatar": contentModel.ModerationSceneProfileAvatar,
	"dish_cover":     contentModel.ModerationSceneDishCover,
	"dish_step":      contentModel.ModerationSceneDishStep,
	"dish_extract":   contentModel.ModerationSceneDishExtract,
	"checkin":        contentModel.ModerationSceneCheckin,
}

// UploadService 提供小程序图片上传业务能力。
type UploadService struct {
	DB         *gorm.DB                              // 业务数据库
	Moderation moderationService.ImageModerationGate // 用户图片强制审核门禁
	Now        func() time.Time                      // 可注入的当前时间函数
}

func (service *UploadService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *UploadService) moderation() moderationService.ImageModerationGate {
	if service != nil && service.Moderation != nil {
		return service.Moderation
	}
	return orderfoodService.ServiceGroupApp.ContentServiceGroup.Moderation
}

func (service *UploadService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// UploadImage 上传图片。
func (service *UploadService) UploadImage(
	ctx context.Context,
	userID string,
	scene string,
	contentType string,
	payload []byte,
	requestID string,
) (frontResponse.ImageUploadResult, error) {
	moderationScene, ok := allowedUploadScenes[scene]
	if !ok || len(payload) == 0 || len(payload) > maxFrontImageBytes {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInvalidImage.DefaultMsg()
	}
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInvalidImage.DefaultMsg()
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(payload))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 10000 || config.Height > 10000 {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInvalidImage.DefaultMsg()
	}
	return service.persistImage(ctx, userID, scene, moderationScene, contentType, payload, requestID)
}

// UploadGeneratedImage 上传生成图片。
func (service *UploadService) UploadGeneratedImage(
	ctx context.Context,
	userID string,
	contentType string,
	payload []byte,
	requestID string,
) (frontResponse.ImageUploadResult, error) {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType != "image/jpeg" && contentType != "image/png" {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInvalidImage.DefaultMsg()
	}
	return service.persistImage(
		ctx,
		userID,
		"generated_cover",
		contentModel.ModerationSceneGeneratedCover,
		contentType,
		payload,
		requestID,
	)
}

func (service *UploadService) persistImage(
	ctx context.Context,
	userID string,
	scene string,
	moderationScene contentModel.ModerationScene,
	contentType string,
	payload []byte,
	requestID string,
) (frontResponse.ImageUploadResult, error) {
	if len(payload) == 0 || len(payload) > maxFrontImageBytes {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInvalidImage.DefaultMsg()
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(payload))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 10000 || config.Height > 10000 {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInvalidImage.DefaultMsg()
	}
	fileID := commonModel.NewID()
	uploadSource := frontUploadSourceUser
	if moderationScene == contentModel.ModerationSceneGeneratedCover {
		uploadSource = frontUploadSourceAI
	}
	persisted, decision, err := service.moderation().AcceptUserImage(
		ctx,
		moderationService.ModerationGateInput{
			RequestID: requestID, StagingFileID: fileID, Scene: moderationScene,
			ContentType: contentType, Bytes: payload,
		},
		func(tx *gorm.DB, permit moderationService.PassedImagePermit) (interface{}, error) {
			extension := ".jpg"
			if contentType == "image/png" {
				extension = ".png"
			} else if contentType == "image/gif" {
				extension = ".gif"
			}
			filename := fileID + extension
			storePath := global.GVA_CONFIG.Local.StorePath
			if strings.TrimSpace(storePath) == "" {
				return nil, appErrors.FrontInternal.New("local image store path is not configured")
			}
			if err := os.MkdirAll(storePath, 0o750); err != nil {
				return nil, appErrors.FrontInternal.Wrap(err, "create image storage directory")
			}
			storagePath := filepath.Join(storePath, filename)
			if err := os.WriteFile(storagePath, payload, 0o640); err != nil {
				return nil, appErrors.FrontInternal.Wrap(err, "store reviewed image")
			}
			publicURL := strings.TrimRight(global.GVA_CONFIG.Local.Path, "/") + "/" + filename
			checksum := sha256.Sum256(payload)
			asset := contentModel.FrontMediaAsset{
				ID: fileID, UserID: userID, Scene: scene, URL: publicURL,
				StoragePath: storagePath, ContentType: contentType, Width: config.Width,
				Height: config.Height, SizeBytes: int64(len(payload)),
				Checksum:       hex.EncodeToString(checksum[:]),
				UploadSource:   uploadSource,
				ResourceStatus: frontMediaResourceActive,
				ReviewStatus:   string(permit.ReviewStatus),
				CreatedAt:      service.now(),
			}
			if err := tx.Create(&asset).Error; err != nil {
				_ = os.Remove(storagePath)
				return nil, appErrors.FrontInternal.Wrap(err, "persist reviewed image")
			}
			return asset, nil
		},
	)
	if err != nil {
		if decision.Status == string(contentModel.ModerationStatusRejected) {
			return frontResponse.ImageUploadResult{}, appErrors.FrontImageRejected.DefaultMsg()
		}
		return frontResponse.ImageUploadResult{}, err
	}
	asset, ok := persisted.(contentModel.FrontMediaAsset)
	if !ok {
		return frontResponse.ImageUploadResult{}, appErrors.FrontInternal.DefaultMsg()
	}
	return frontResponse.ImageUploadResult{
		FileID: asset.ID, URL: asset.URL, Width: asset.Width, Height: asset.Height,
		ReviewStatus: asset.ReviewStatus, RejectReason: asset.RejectReason,
	}, nil
}

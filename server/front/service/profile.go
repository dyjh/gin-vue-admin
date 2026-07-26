package service

import (
	"context"
	"errors"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	"gorm.io/gorm"
)

// ProfileService 提供小程序个人资料业务能力。
type ProfileService struct {
	DB *gorm.DB // 业务数据库
}

func (service *ProfileService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// Get 获取小程序个人资料。
func (service *ProfileService) Get(
	ctx context.Context,
	userID string,
) (orderfoodModel.MiniAppUser, error) {
	db := service.database()
	if db == nil {
		return orderfoodModel.MiniAppUser{}, appErrors.FrontInternal.DefaultMsg()
	}
	var user orderfoodModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodModel.MiniAppUser{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return orderfoodModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "load miniapp profile")
	}
	if user.Status == orderfoodModel.UserStatusDisabled {
		return orderfoodModel.MiniAppUser{}, appErrors.FrontUserDisabled.DefaultMsg()
	}
	return user, nil
}

// Update 更新小程序个人资料。
func (service *ProfileService) Update(
	ctx context.Context,
	userID string,
	input frontRequest.ProfileUpdateInput,
) (orderfoodModel.MiniAppUser, error) {
	db := service.database()
	if db == nil {
		return orderfoodModel.MiniAppUser{}, appErrors.FrontInternal.DefaultMsg()
	}
	user, err := service.Get(ctx, userID)
	if err != nil {
		return orderfoodModel.MiniAppUser{}, err
	}
	updates := map[string]interface{}{
		"nickname": strings.TrimSpace(input.Nickname),
		"version":  gorm.Expr("version + 1"),
	}
	if input.AvatarFileID != nil {
		var asset orderfoodModel.FrontMediaAsset
		if err := db.WithContext(ctx).First(
			&asset,
			"id = ? AND user_id = ? AND scene = ? AND review_status = ?",
			*input.AvatarFileID,
			userID,
			"profile_avatar",
			orderfoodModel.MediaReviewPassed,
		).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return orderfoodModel.MiniAppUser{}, appErrors.FrontInvalidImage.DefaultMsg()
			}
			return orderfoodModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "resolve profile avatar")
		}
		updates["avatar_url"] = asset.URL
		user.AvatarURL = &asset.URL
	}
	if err := db.WithContext(ctx).Model(&orderfoodModel.MiniAppUser{}).
		Where("id = ? AND status = ?", userID, orderfoodModel.UserStatusNormal).
		Updates(updates).Error; err != nil {
		return orderfoodModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "update miniapp profile")
	}
	user.Nickname = strings.TrimSpace(input.Nickname)
	user.Version++
	return user, nil
}

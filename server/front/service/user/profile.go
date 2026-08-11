package user

import (
	"context"
	"errors"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	"github.com/dyjh/order-food-mini-app/server/global"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
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
) (userModel.MiniAppUser, error) {
	db := service.database()
	if db == nil {
		return userModel.MiniAppUser{}, appErrors.FrontInternal.DefaultMsg()
	}
	var user userModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return userModel.MiniAppUser{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return userModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "load miniapp profile")
	}
	if user.Status == userModel.UserStatusDisabled {
		return userModel.MiniAppUser{}, appErrors.FrontUserDisabled.DefaultMsg()
	}
	return user, nil
}

// Update 更新小程序个人资料。
func (service *ProfileService) Update(
	ctx context.Context,
	userID string,
	input frontRequest.ProfileUpdateInput,
) (userModel.MiniAppUser, error) {
	db := service.database()
	if db == nil {
		return userModel.MiniAppUser{}, appErrors.FrontInternal.DefaultMsg()
	}
	user, err := service.Get(ctx, userID)
	if err != nil {
		return userModel.MiniAppUser{}, err
	}
	updates := map[string]interface{}{
		"nickname": strings.TrimSpace(input.Nickname),
		"version":  gorm.Expr("version + 1"),
	}
	if input.AvatarFileID != nil {
		var asset contentModel.FrontMediaAsset
		if err := db.WithContext(ctx).First(
			&asset,
			"id = ? AND user_id = ? AND scene = ? AND review_status IN ?",
			*input.AvatarFileID,
			userID,
			"profile_avatar",
			contentModel.UsableMediaReviewStatuses(),
		).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return userModel.MiniAppUser{}, appErrors.FrontInvalidImage.DefaultMsg()
			}
			return userModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "resolve profile avatar")
		}
		updates["avatar_url"] = asset.URL
		user.AvatarURL = &asset.URL
	}
	if err := db.WithContext(ctx).Model(&userModel.MiniAppUser{}).
		Where("id = ? AND status = ?", userID, userModel.UserStatusNormal).
		Updates(updates).Error; err != nil {
		return userModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "update miniapp profile")
	}
	user.Nickname = strings.TrimSpace(input.Nickname)
	user.Version++
	return user, nil
}

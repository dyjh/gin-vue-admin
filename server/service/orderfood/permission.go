package orderfood

import (
	"context"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	"gorm.io/gorm"
)

// PermissionService 提供后台权限业务能力。
type PermissionService struct {
	DB *gorm.DB // GVA权限数据库
}

// NewPermissionService 创建权限服务实例。
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{DB: db}
}

func (service *PermissionService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// HasPermission 判断是否具有权限。
func (service *PermissionService) HasPermission(
	ctx context.Context,
	authorityID uint,
	permission string,
) (bool, error) {
	if permission == "" {
		return false, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return false, appErrors.AdminInternal.DefaultMsg()
	}

	var count int64
	err := db.WithContext(ctx).
		Model(&system.SysAuthorityBtn{}).
		Joins(
			"JOIN sys_base_menu_btns ON "+
				"sys_base_menu_btns.id = sys_authority_btns.sys_base_menu_btn_id",
		).
		Where(
			"sys_authority_btns.authority_id = ? AND sys_base_menu_btns.name = ?",
			authorityID,
			permission,
		).
		Count(&count).Error
	if err != nil {
		return false, appErrors.AdminInternal.Wrap(err, "query GVA button permission")
	}
	return count > 0, nil
}

// Require 校验角色是否拥有指定权限，缺失时返回统一无权错误。
func (service *PermissionService) Require(
	ctx context.Context,
	authorityID uint,
	permission string,
) error {
	allowed, err := service.HasPermission(ctx, authorityID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return appErrors.AdminNoPermission.DefaultMsg()
	}
	return nil
}

// DefaultRolePermissionMatrix 返回后台默认角色权限模板。
func DefaultRolePermissionMatrix() map[uint][]string {
	matrix := map[uint][]string{
		orderfoodModel.AuthorityOrderFoodSuperAdmin: {
			"orderfood:ai-usage:read",
			"orderfood:ai-usage:sensitive-delete",
			"orderfood:ai-usage:sensitive-read",
			"orderfood:audit:read",
			"orderfood:prompt:read",
			"orderfood:prompt:test",
			"orderfood:prompt:update",
			"orderfood:capability:read",
			"orderfood:capability:update",
			"orderfood:category:create",
			"orderfood:category:delete",
			"orderfood:category:read",
			"orderfood:category:sort",
			"orderfood:category:update",
			"orderfood:dashboard:read",
			"orderfood:discoverable-dish:read",
			"orderfood:governance:cascade",
			"orderfood:governance:execute",
			"orderfood:governance:read",
			"orderfood:governance:retry",
			"orderfood:meal:read",
			"orderfood:media:read",
			"orderfood:model:create",
			"orderfood:model:delete",
			"orderfood:model:read",
			"orderfood:model:status",
			"orderfood:model:update",
			"orderfood:moderation-config:credential-write",
			"orderfood:moderation-config:read",
			"orderfood:moderation-config:test",
			"orderfood:moderation-config:update",
			"orderfood:moderation:read",
			"orderfood:moderation:sensitive-read",
			"orderfood:notification:read",
			"orderfood:official-dish:create",
			"orderfood:official-dish:cover-upload",
			"orderfood:official-dish:delete",
			"orderfood:official-dish:read",
			"orderfood:official-dish:update",
			"orderfood:platform-policy:read",
			"orderfood:platform-policy:update",
			"orderfood:points:adjust",
			"orderfood:points:read",
			"orderfood:point-rule:read",
			"orderfood:point-rule:update",
			"orderfood:provider:create",
			"orderfood:provider:credential-write",
			"orderfood:provider:delete",
			"orderfood:provider:read",
			"orderfood:provider:status",
			"orderfood:provider:test",
			"orderfood:provider:update",
			"orderfood:recommendation:create",
			"orderfood:recommendation:delete",
			"orderfood:recommendation:offline",
			"orderfood:recommendation:publish",
			"orderfood:recommendation:read",
			"orderfood:recommendation:sort",
			"orderfood:recommendation:update",
			"orderfood:shopping:read",
			"orderfood:suggestion-catalog:read",
			"orderfood:suggestion-catalog:update",
			"orderfood:subscribe-log:read",
			"orderfood:subscribe-template:create",
			"orderfood:subscribe-template:delete",
			"orderfood:subscribe-template:read",
			"orderfood:subscribe-template:status",
			"orderfood:subscribe-template:update",
			"orderfood:tag:create",
			"orderfood:tag:delete",
			"orderfood:tag:read",
			"orderfood:tag:sort",
			"orderfood:tag:update",
			"orderfood:unit:create",
			"orderfood:unit:delete",
			"orderfood:unit:read",
			"orderfood:unit:sort",
			"orderfood:unit:update",
			"orderfood:user-dish:private-read",
			"orderfood:user-dish:read",
			"orderfood:user-recipe:private-read",
			"orderfood:user-recipe:read",
			"orderfood:user:disable",
			"orderfood:user:preference:read",
			"orderfood:user:read",
			"orderfood:wechat-config:read",
			"orderfood:wechat-config:update",
		},
		orderfoodModel.AuthorityOrderFoodOperator: {
			"orderfood:ai-usage:read",
			"orderfood:capability:read",
			"orderfood:category:create",
			"orderfood:category:delete",
			"orderfood:category:read",
			"orderfood:category:sort",
			"orderfood:category:update",
			"orderfood:dashboard:read",
			"orderfood:discoverable-dish:read",
			"orderfood:meal:read",
			"orderfood:media:read",
			"orderfood:model:read",
			"orderfood:notification:read",
			"orderfood:official-dish:create",
			"orderfood:official-dish:cover-upload",
			"orderfood:official-dish:delete",
			"orderfood:official-dish:read",
			"orderfood:official-dish:update",
			"orderfood:points:read",
			"orderfood:provider:read",
			"orderfood:recommendation:create",
			"orderfood:recommendation:delete",
			"orderfood:recommendation:offline",
			"orderfood:recommendation:publish",
			"orderfood:recommendation:read",
			"orderfood:recommendation:sort",
			"orderfood:recommendation:update",
			"orderfood:shopping:read",
			"orderfood:subscribe-log:read",
			"orderfood:tag:create",
			"orderfood:tag:delete",
			"orderfood:tag:read",
			"orderfood:tag:sort",
			"orderfood:tag:update",
			"orderfood:unit:create",
			"orderfood:unit:delete",
			"orderfood:unit:read",
			"orderfood:unit:sort",
			"orderfood:unit:update",
			"orderfood:user:read",
		},
		orderfoodModel.AuthorityOrderFoodGovernance: {
			"orderfood:audit:read",
			"orderfood:discoverable-dish:read",
			"orderfood:governance:execute",
			"orderfood:governance:read",
			"orderfood:governance:retry",
			"orderfood:media:read",
			"orderfood:moderation-config:read",
			"orderfood:moderation:read",
			"orderfood:notification:read",
			"orderfood:recommendation:offline",
			"orderfood:recommendation:read",
			"orderfood:user-dish:private-read",
			"orderfood:user-dish:read",
			"orderfood:user-recipe:private-read",
			"orderfood:user-recipe:read",
			"orderfood:user:disable",
			"orderfood:user:read",
		},
		orderfoodModel.AuthorityOrderFoodSupport: {
			"orderfood:ai-usage:read",
			"orderfood:meal:read",
			"orderfood:notification:read",
			"orderfood:points:read",
			"orderfood:shopping:read",
			"orderfood:subscribe-log:read",
			"orderfood:user:read",
		},
	}
	// GVA自动注入的首个管理员角色只在初始化阶段获得全量权限；
	// 运行时仍以数据库权限表为准，后台撤销后立即生效。
	matrix[orderfoodModel.AuthorityPlatformSuperAdmin] = append(
		[]string(nil),
		matrix[orderfoodModel.AuthorityOrderFoodSuperAdmin]...,
	)
	return matrix
}

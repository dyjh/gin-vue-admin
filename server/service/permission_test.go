package service

import (
	"context"
	"testing"

	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// newOrderFoodTestDB 创建订单服务测试共享的独立 MySQL 数据库。
func newOrderFoodTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(&system.SysBaseMenuBtn{}, &system.SysAuthorityBtn{}); err != nil {
		t.Fatalf("migrate shared permissions: %v", err)
	}
	seedTestPermissionTemplates(t, db)
	return db
}

// grantTestButtonPermission 为测试角色写入一项 GVA 按钮权限。
func grantTestButtonPermission(
	t *testing.T,
	db *gorm.DB,
	authorityID uint,
	permission string,
) {
	t.Helper()
	var button system.SysBaseMenuBtn
	if err := db.Where("name = ?", permission).FirstOrCreate(
		&button,
		system.SysBaseMenuBtn{
			Name:          permission,
			Desc:          permission,
			SysBaseMenuID: 1,
		},
	).Error; err != nil {
		t.Fatalf("create test button permission %q: %v", permission, err)
	}
	authorization := system.SysAuthorityBtn{
		AuthorityId:      authorityID,
		SysMenuID:        button.SysBaseMenuID,
		SysBaseMenuBtnID: button.ID,
	}
	if err := db.Where(
		"authority_id = ? AND sys_menu_id = ? AND sys_base_menu_btn_id = ?",
		authorization.AuthorityId,
		authorization.SysMenuID,
		authorization.SysBaseMenuBtnID,
	).FirstOrCreate(&authorization).Error; err != nil {
		t.Fatalf("grant test button permission %q: %v", permission, err)
	}
}

// revokeTestButtonPermission 从测试角色移除一项 GVA 按钮权限。
func revokeTestButtonPermission(
	t *testing.T,
	db *gorm.DB,
	authorityID uint,
	permission string,
) {
	t.Helper()
	if err := db.Where(
		"authority_id = ? AND sys_base_menu_btn_id IN (?)",
		authorityID,
		db.Model(&system.SysBaseMenuBtn{}).Select("id").Where("name = ?", permission),
	).Delete(&system.SysAuthorityBtn{}).Error; err != nil {
		t.Fatalf("revoke test button permission %q: %v", permission, err)
	}
}

// seedTestPermissionTemplates 注入测试使用的默认角色按钮权限。
func seedTestPermissionTemplates(t *testing.T, db *gorm.DB) {
	t.Helper()
	for authorityID, permissions := range DefaultRolePermissionMatrix() {
		for _, permission := range permissions {
			grantTestButtonPermission(t, db, authorityID, permission)
		}
	}
}

func TestPermissionServiceRoleBoundariesAndDatabaseTruth(t *testing.T) {
	db := newOrderFoodTestDB(t)
	service := NewPermissionService(db)

	tests := []struct {
		name        string
		authorityID uint
		permission  string
		want        bool
	}{
		{"platform initial admin seeded permission", 888, PermissionUserRead, true},
		{"platform initial admin no hardcoded fallback", 888, "orderfood:anything", false},
		{"orderfood super seeded permission", 9901, PermissionUserRead, true},
		{"orderfood super no hardcoded fallback", 9901, "orderfood:anything", false},
		{"operator user read", 9902, PermissionUserRead, true},
		{"operator cannot disable user", 9902, PermissionUserDisable, false},
		{"governance user read", 9903, PermissionUserRead, true},
		{"governance can disable user", 9903, PermissionUserDisable, true},
		{"governance cannot read preference", 9903, PermissionUserPreferenceRead, false},
		{"support user read", 9904, PermissionUserRead, true},
		{"support cannot read audit", 9904, PermissionAuditRead, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := service.HasPermission(
				context.Background(),
				test.authorityID,
				test.permission,
			)
			if err != nil {
				t.Fatalf("HasPermission: %v", err)
			}
			if got != test.want {
				t.Fatalf("allowed = %v, want %v", got, test.want)
			}
		})
	}

	revokeTestButtonPermission(
		t,
		db,
		orderfoodModel.AuthorityOrderFoodOperator,
		PermissionUserRead,
	)
	allowed, err := service.HasPermission(
		context.Background(),
		orderfoodModel.AuthorityOrderFoodOperator,
		PermissionUserRead,
	)
	if err != nil {
		t.Fatalf("HasPermission after disable: %v", err)
	}
	if allowed {
		t.Fatal("disabled database permission still allowed")
	}

	grantTestButtonPermission(
		t,
		db,
		orderfoodModel.AuthorityOrderFoodSupport,
		PermissionAuditRead,
	)
	allowed, err = service.HasPermission(
		context.Background(),
		orderfoodModel.AuthorityOrderFoodSupport,
		PermissionAuditRead,
	)
	if err != nil {
		t.Fatalf("HasPermission after custom grant: %v", err)
	}
	if !allowed {
		t.Fatal("database custom permission was not allowed")
	}
}

func TestOrderFoodSuperAdminPermissionTemplateIsComplete(t *testing.T) {
	permissions := DefaultRolePermissionMatrix()[orderfoodModel.AuthorityOrderFoodSuperAdmin]
	if len(permissions) != 87 {
		t.Fatalf("super admin permission count = %d, want 87", len(permissions))
	}
	required := map[string]bool{
		PermissionUserRead:                 false,
		PermissionUserDisable:              false,
		PermissionUserPreferenceRead:       false,
		PermissionAuditRead:                false,
		PermissionAICapabilityPromptRead:   false,
		PermissionAICapabilityPromptUpdate: false,
		PermissionAICapabilityPromptTest:   false,
		PermissionPointRuleRead:            false,
		PermissionPointRuleUpdate:          false,
		PermissionSuggestionCatalogRead:    false,
		PermissionSuggestionCatalogUpdate:  false,
	}
	for _, permission := range permissions {
		if _, ok := required[permission]; ok {
			required[permission] = true
		}
	}
	for permission, found := range required {
		if !found {
			t.Fatalf("super admin permission template missing %q", permission)
		}
	}
}

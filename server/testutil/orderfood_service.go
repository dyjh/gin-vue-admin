package testutil

import (
	"testing"

	"github.com/dyjh/order-food-mini-app/server/model/system"
	"gorm.io/gorm"
)

// NewOrderFoodServiceDB creates the shared isolated database used by OrderFood service tests.
func NewOrderFoodServiceDB(t testing.TB, permissionMatrix map[uint][]string) *gorm.DB {
	t.Helper()
	db := OpenMySQL(t)
	if err := db.AutoMigrate(&system.SysBaseMenuBtn{}, &system.SysAuthorityBtn{}); err != nil {
		t.Fatalf("migrate shared permissions: %v", err)
	}
	SeedOrderFoodPermissions(t, db, permissionMatrix)
	return db
}

// GrantOrderFoodPermission grants one GVA button permission to a test authority.
func GrantOrderFoodPermission(
	t testing.TB,
	db *gorm.DB,
	authorityID uint,
	permission string,
) {
	t.Helper()
	var button system.SysBaseMenuBtn
	if err := db.Where("name = ?", permission).FirstOrCreate(
		&button,
		system.SysBaseMenuBtn{Name: permission, Desc: permission, SysBaseMenuID: 1},
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

// RevokeOrderFoodPermission removes one test permission from an authority.
func RevokeOrderFoodPermission(
	t testing.TB,
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

// SeedOrderFoodPermissions writes the default role templates used by service tests.
func SeedOrderFoodPermissions(t testing.TB, db *gorm.DB, permissionMatrix map[uint][]string) {
	t.Helper()
	for authorityID, permissions := range permissionMatrix {
		for _, permission := range permissions {
			GrantOrderFoodPermission(t, db, authorityID, permission)
		}
	}
}

package source

import (
	"testing"

	adapter "github.com/casbin/gorm-adapter/v3"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	"github.com/dyjh/order-food-mini-app/server/testutil"
)

// TestEnsureOrderFoodAdminSeedIsIdempotentAndKeepsRoleBoundaries 验证管理端种子数据幂等且不突破角色边界。
func TestEnsureOrderFoodAdminSeedIsIdempotentAndKeepsRoleBoundaries(t *testing.T) {
	db := testutil.OpenMySQL(t)
	err := db.AutoMigrate(
		&system.SysAuthority{},
		&system.SysBaseMenu{},
		&system.SysBaseMenuBtn{},
		&system.SysAuthorityBtn{},
		&system.SysAuthorityMenu{},
		&system.SysApi{},
		&adapter.CasbinRule{},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err = ensureOrderFoodAdminSeed(db); err != nil {
		t.Fatal(err)
	}

	// 模拟管理员在 GVA 中撤销运营角色的供应商只读权限。
	var providerReadButton system.SysBaseMenuBtn
	if err = db.Where("name = ?", "orderfood:provider:read").
		First(&providerReadButton).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Where(
		"authority_id = ? AND sys_base_menu_btn_id = ?",
		commonModel.AuthorityOrderFoodOperator,
		providerReadButton.ID,
	).Delete(&system.SysAuthorityBtn{}).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Where(
		"v0 = ? AND v1 = ? AND v2 = ?",
		"9902",
		"/orderfood/ai-providers",
		"GET",
	).Delete(&adapter.CasbinRule{}).Error; err != nil {
		t.Fatal(err)
	}
	// 平台首个管理员的权限即使缺失，也应在重复初始化时自动补齐。
	if err = db.Where(
		"v0 = ? AND v1 = ? AND v2 = ?",
		"888",
		"/orderfood/ai-providers",
		"GET",
	).Delete(&adapter.CasbinRule{}).Error; err != nil {
		t.Fatal(err)
	}

	if err = ensureOrderFoodAdminSeed(db); err != nil {
		t.Fatal(err)
	}

	var menuCount int64
	db.Model(&system.SysBaseMenu{}).
		Where("name IN ?", []string{
			"OrderFoodDashboard",
			"OrderFoodUsersAndPoints",
			"OrderFoodDishOperations",
			"OrderFoodUsers",
			"OrderFoodPointRules",
			"OrderFoodUserDishes",
			"OrderFoodUserRecipes",
			"OrderFoodAiCapabilities",
			"OrderFoodPlatformPolicy",
			"OrderFoodAiCapabilityConfig",
			"OrderFoodAiProviders",
			"OrderFoodAiModels",
			"OrderFoodWeChatConfig",
		}).
		Count(&menuCount)
	if menuCount != 13 {
		t.Fatalf("implemented menu count = %d, want 13", menuCount)
	}

	var topLevelMenus []system.SysBaseMenu
	if err = db.Where("name IN ?", []string{
		"OrderFoodDashboard",
		"OrderFoodUsersAndPoints",
		"OrderFoodDishOperations",
		"OrderFoodContentSafety",
		"OrderFoodMealManagement",
		"OrderFoodAiCapabilities",
		"OrderFoodMessageCenter",
		"OrderFoodWeChatConfig",
	}).Find(&topLevelMenus).Error; err != nil {
		t.Fatal(err)
	}
	if len(topLevelMenus) != 8 {
		t.Fatalf("top-level orderfood menu count = %d, want 8", len(topLevelMenus))
	}
	for _, menu := range topLevelMenus {
		if menu.ParentId != 0 || menu.MenuLevel != 0 {
			t.Fatalf(
				"menu %s has parent=%d level=%d, want top-level",
				menu.Name,
				menu.ParentId,
				menu.MenuLevel,
			)
		}
	}

	var operatorDishRoutes int64
	db.Model(&adapter.CasbinRule{}).
		Where(
			"v0 = ? AND v1 = ? AND v2 = ?",
			"9902",
			"/orderfood/user-dishes",
			"GET",
		).
		Count(&operatorDishRoutes)
	if operatorDishRoutes != 0 {
		t.Fatalf("operator unexpectedly received user dish route")
	}

	var governanceDishRoutes int64
	db.Model(&adapter.CasbinRule{}).
		Where(
			"v0 = ? AND v1 = ? AND v2 = ?",
			"9903",
			"/orderfood/user-dishes",
			"GET",
		).
		Count(&governanceDishRoutes)
	if governanceDishRoutes != 1 {
		t.Fatalf("governance dish route count = %d, want 1", governanceDishRoutes)
	}

	var supportPreferenceButtons int64
	db.Model(&system.SysAuthorityBtn{}).
		Joins("JOIN sys_base_menu_btns ON sys_base_menu_btns.id = sys_authority_btns.sys_base_menu_btn_id").
		Where(
			"sys_authority_btns.authority_id = ? AND sys_base_menu_btns.name = ?",
			commonModel.AuthorityOrderFoodSupport,
			"orderfood:user:preference:read",
		).
		Count(&supportPreferenceButtons)
	if supportPreferenceButtons != 0 {
		t.Fatalf("support unexpectedly received preference button")
	}

	for _, authorityID := range []uint{
		commonModel.AuthorityPlatformSuperAdmin,
		commonModel.AuthorityOrderFoodSupport,
	} {
		var pointEntryUserReadButtons int64
		db.Model(&system.SysAuthorityBtn{}).
			Joins("JOIN sys_base_menu_btns ON sys_base_menu_btns.id = sys_authority_btns.sys_base_menu_btn_id").
			Joins("JOIN sys_base_menus ON sys_base_menus.id = sys_base_menu_btns.sys_base_menu_id").
			Where(
				"sys_authority_btns.authority_id = ? AND sys_base_menus.name = ? AND sys_base_menu_btns.name = ?",
				authorityID,
				"OrderFoodPointEntries",
				"orderfood:user:read",
			).
			Count(&pointEntryUserReadButtons)
		if pointEntryUserReadButtons != 1 {
			t.Fatalf(
				"authority %d point-entry user-read button count = %d, want 1",
				authorityID,
				pointEntryUserReadButtons,
			)
		}
	}

	for _, authorityID := range []uint{
		commonModel.AuthorityPlatformSuperAdmin,
		commonModel.AuthorityOrderFoodSuperAdmin,
		commonModel.AuthorityOrderFoodOperator,
	} {
		var officialDishRecommendationButtons int64
		db.Model(&system.SysAuthorityBtn{}).
			Joins("JOIN sys_base_menu_btns ON sys_base_menu_btns.id = sys_authority_btns.sys_base_menu_btn_id").
			Joins("JOIN sys_base_menus ON sys_base_menus.id = sys_base_menu_btns.sys_base_menu_id").
			Where(
				"sys_authority_btns.authority_id = ? AND sys_base_menus.name = ? AND sys_base_menu_btns.name = ?",
				authorityID,
				"OrderFoodOfficialDishes",
				"orderfood:recommendation:create",
			).
			Count(&officialDishRecommendationButtons)
		if officialDishRecommendationButtons != 1 {
			t.Fatalf(
				"authority %d official-dish recommendation-create button count = %d, want 1",
				authorityID,
				officialDishRecommendationButtons,
			)
		}
	}
	var operatorProviderReadRoutes int64
	db.Model(&adapter.CasbinRule{}).
		Where(
			"v0 = ? AND v1 = ? AND v2 = ?",
			"9902",
			"/orderfood/ai-providers",
			"GET",
		).
		Count(&operatorProviderReadRoutes)
	if operatorProviderReadRoutes != 0 {
		t.Fatalf("revoked operator provider route count = %d, want 0", operatorProviderReadRoutes)
	}

	var operatorProviderReadButtons int64
	db.Model(&system.SysAuthorityBtn{}).
		Where(
			"authority_id = ? AND sys_base_menu_btn_id = ?",
			commonModel.AuthorityOrderFoodOperator,
			providerReadButton.ID,
		).
		Count(&operatorProviderReadButtons)
	if operatorProviderReadButtons != 0 {
		t.Fatalf("revoked operator provider button count = %d, want 0", operatorProviderReadButtons)
	}

	var platformProviderReadRoutes int64
	db.Model(&adapter.CasbinRule{}).
		Where(
			"v0 = ? AND v1 = ? AND v2 = ?",
			"888",
			"/orderfood/ai-providers",
			"GET",
		).
		Count(&platformProviderReadRoutes)
	if platformProviderReadRoutes != 1 {
		t.Fatalf("platform provider read route count = %d, want 1", platformProviderReadRoutes)
	}

	var operatorProviderCreateRoutes int64
	db.Model(&adapter.CasbinRule{}).
		Where(
			"v0 = ? AND v1 = ? AND v2 = ?",
			"9902",
			"/orderfood/ai-providers",
			"POST",
		).
		Count(&operatorProviderCreateRoutes)
	if operatorProviderCreateRoutes != 0 {
		t.Fatalf("operator unexpectedly received provider create route")
	}

	var operatorPointRuleRoutes int64
	db.Model(&adapter.CasbinRule{}).
		Where(
			"v0 = ? AND v1 = ? AND v2 = ?",
			"9902",
			"/orderfood/point-rules",
			"GET",
		).
		Count(&operatorPointRuleRoutes)
	if operatorPointRuleRoutes != 0 {
		t.Fatalf("operator unexpectedly received point rule route")
	}

	var operatorPromptButtons int64
	db.Model(&system.SysAuthorityBtn{}).
		Joins("JOIN sys_base_menu_btns ON sys_base_menu_btns.id = sys_authority_btns.sys_base_menu_btn_id").
		Where(
			"sys_authority_btns.authority_id = ? AND sys_base_menu_btns.name LIKE ?",
			commonModel.AuthorityOrderFoodOperator,
			"orderfood:prompt:%",
		).
		Count(&operatorPromptButtons)
	if operatorPromptButtons != 0 {
		t.Fatalf("operator unexpectedly received prompt button permissions")
	}

	var permissionCount int64
	db.Model(&system.SysBaseMenuBtn{}).
		Distinct("name").
		Where("name LIKE ?", "orderfood:%").
		Count(&permissionCount)
	if permissionCount != 85 {
		t.Fatalf("configurable orderfood permission count = %d, want 85", permissionCount)
	}
}

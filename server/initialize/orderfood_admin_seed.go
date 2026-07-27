package initialize

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	adapter "github.com/casbin/gorm-adapter/v3"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"gorm.io/gorm"
)

type orderFoodAdminRouteSeed struct {
	Method      string
	Path        string
	Description string
	Roles       []uint
}

type orderFoodMenuSeed struct {
	Path      string
	Name      string
	Component string
	Title     string
	Icon      string
	Sort      int
}

var (
	orderFoodSuperRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
	}
	orderFoodAllRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
		orderfoodModel.AuthorityOrderFoodGovernance,
		orderfoodModel.AuthorityOrderFoodSupport,
	}
	orderFoodGovernanceRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodGovernance,
	}
	orderFoodDishRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
		orderfoodModel.AuthorityOrderFoodGovernance,
	}
	orderFoodCatalogRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
	}
	orderFoodAIRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
	}
	orderFoodDashboardRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
	}
	orderFoodAIUsageRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
		orderfoodModel.AuthorityOrderFoodSupport,
	}
	orderFoodMessageLogRoles = []uint{
		orderfoodModel.AuthorityPlatformSuperAdmin,
		orderfoodModel.AuthorityOrderFoodSuperAdmin,
		orderfoodModel.AuthorityOrderFoodOperator,
		orderfoodModel.AuthorityOrderFoodSupport,
	}
)

// ensureOrderFoodAdminSeed 在当前业务表创建后幂等注入管理端角色、菜单、按钮和路由。
func ensureOrderFoodAdminSeed(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	newAuthorities, err := ensureOrderFoodAuthorities(db)
	if err != nil {
		return err
	}

	menuIDs, err := ensureImplementedOrderFoodMenus(db)
	if err != nil {
		return err
	}
	if err = ensureOrderFoodMenuAccess(db, menuIDs, newAuthorities); err != nil {
		return err
	}
	if err = ensureOrderFoodButtons(db, menuIDs, newAuthorities); err != nil {
		return err
	}
	return ensureImplementedOrderFoodRoutes(db, newAuthorities)
}

// ensureOrderFoodAuthorities 确保预设角色存在，并返回本次新建的角色集合。
func ensureOrderFoodAuthorities(db *gorm.DB) (map[uint]bool, error) {
	zero := uint(0)
	created := make(map[uint]bool)
	authorities := []system.SysAuthority{
		{
			AuthorityId:   orderfoodModel.AuthorityOrderFoodSuperAdmin,
			AuthorityName: "来干饭超级管理员",
			ParentId:      &zero,
			DefaultRouter: "OrderFoodDashboard",
		},
		{
			AuthorityId:   orderfoodModel.AuthorityOrderFoodOperator,
			AuthorityName: "来干饭运营管理员",
			ParentId:      &zero,
			DefaultRouter: "OrderFoodDashboard",
		},
		{
			AuthorityId:   orderfoodModel.AuthorityOrderFoodGovernance,
			AuthorityName: "来干饭内容安全管理员",
			ParentId:      &zero,
			DefaultRouter: "OrderFoodUserDishes",
		},
		{
			AuthorityId:   orderfoodModel.AuthorityOrderFoodSupport,
			AuthorityName: "来干饭客服只读",
			ParentId:      &zero,
			DefaultRouter: "OrderFoodUsers",
		},
	}
	for index := range authorities {
		authority := authorities[index]
		var existing system.SysAuthority
		err := db.Where("authority_id = ?", authority.AuthorityId).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err = db.Create(&authority).Error; err != nil {
				return nil, err
			}
			created[authority.AuthorityId] = true
			continue
		}
		if err != nil {
			return nil, err
		}
		if err = db.Model(&existing).Updates(map[string]interface{}{
			"authority_name": authority.AuthorityName,
			"default_router": authority.DefaultRouter,
		}).Error; err != nil {
			return nil, err
		}
	}
	return created, nil
}

func ensureImplementedOrderFoodMenus(db *gorm.DB) (map[string]uint, error) {
	dashboard, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "dashboard",
		Name:      "OrderFoodDashboard",
		Component: "view/orderFood/dashboard/index.vue",
		Title:     "数据概览",
		Icon:      "data-analysis",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	usersGroup, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "users-and-points",
		Name:      "OrderFoodUsersAndPoints",
		Component: "view/routerHolder.vue",
		Title:     "用户与积分",
		Icon:      "user",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	dishesGroup, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "dish-operations",
		Name:      "OrderFoodDishOperations",
		Component: "view/routerHolder.vue",
		Title:     "菜品运营",
		Icon:      "dish",
		Sort:      3,
	})
	if err != nil {
		return nil, err
	}
	contentSafetyGroup, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "content-safety",
		Name:      "OrderFoodContentSafety",
		Component: "view/routerHolder.vue",
		Title:     "内容安全",
		Icon:      "lock",
		Sort:      4,
	})
	if err != nil {
		return nil, err
	}
	mealGroup, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "meal-management",
		Name:      "OrderFoodMealManagement",
		Component: "view/routerHolder.vue",
		Title:     "饭局管理",
		Icon:      "calendar",
		Sort:      5,
	})
	if err != nil {
		return nil, err
	}
	aiGroup, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "ai-capabilities",
		Name:      "OrderFoodAiCapabilities",
		Component: "view/routerHolder.vue",
		Title:     "AI 能力",
		Icon:      "cpu",
		Sort:      6,
	})
	if err != nil {
		return nil, err
	}
	messageGroup, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "message-center",
		Name:      "OrderFoodMessageCenter",
		Component: "view/routerHolder.vue",
		Title:     "消息中心",
		Icon:      "message",
		Sort:      7,
	})
	if err != nil {
		return nil, err
	}
	wechatConfig, err := ensureOrderFoodMenu(db, 0, 0, orderFoodMenuSeed{
		Path:      "wechat-config",
		Name:      "OrderFoodWeChatConfig",
		Component: "view/orderFood/wechatConfig/index.vue",
		Title:     "微信配置",
		Icon:      "setting",
		Sort:      8,
	})
	if err != nil {
		return nil, err
	}
	users, err := ensureOrderFoodMenu(db, usersGroup.ID, 1, orderFoodMenuSeed{
		Path:      "users",
		Name:      "OrderFoodUsers",
		Component: "view/orderFood/user/index.vue",
		Title:     "小程序用户",
		Icon:      "user",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	pointEntries, err := ensureOrderFoodMenu(db, usersGroup.ID, 1, orderFoodMenuSeed{
		Path:      "point-entries",
		Name:      "OrderFoodPointEntries",
		Component: "view/orderFood/points/index.vue",
		Title:     "积分流水",
		Icon:      "tickets",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	pointRules, err := ensureOrderFoodMenu(db, usersGroup.ID, 1, orderFoodMenuSeed{
		Path:      "point-rules",
		Name:      "OrderFoodPointRules",
		Component: "view/orderFood/pointRule/index.vue",
		Title:     "积分规则",
		Icon:      "coin",
		Sort:      3,
	})
	if err != nil {
		return nil, err
	}
	userDishes, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "user-dishes",
		Name:      "OrderFoodUserDishes",
		Component: "view/orderFood/userDish/index.vue",
		Title:     "用户菜品",
		Icon:      "dish",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	userRecipes, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "user-recipes",
		Name:      "OrderFoodUserRecipes",
		Component: "view/orderFood/userRecipe/index.vue",
		Title:     "用户菜谱",
		Icon:      "notebook",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	suggestionCatalog, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "suggestion-catalog",
		Name:      "OrderFoodSuggestionCatalog",
		Component: "view/orderFood/suggestionCatalog/index.vue",
		Title:     "标准菜品索引",
		Icon:      "collection",
		Sort:      3,
	})
	if err != nil {
		return nil, err
	}
	discoverableDishes, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "discoverable-dishes",
		Name:      "OrderFoodDiscoverableDishes",
		Component: "view/orderFood/discoverableDish/index.vue",
		Title:     "可发现菜品",
		Icon:      "view",
		Sort:      4,
	})
	if err != nil {
		return nil, err
	}
	recommendations, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "recommendations",
		Name:      "OrderFoodRecommendations",
		Component: "view/orderFood/recommendation/index.vue",
		Title:     "推荐精选",
		Icon:      "star",
		Sort:      5,
	})
	if err != nil {
		return nil, err
	}
	officialDishes, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "official-dishes",
		Name:      "OrderFoodOfficialDishes",
		Component: "view/orderFood/officialDish/index.vue",
		Title:     "官方菜品",
		Icon:      "food",
		Sort:      6,
	})
	if err != nil {
		return nil, err
	}
	categories, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "categories",
		Name:      "OrderFoodCategories",
		Component: "view/orderFood/category/index.vue",
		Title:     "分类管理",
		Icon:      "collection-tag",
		Sort:      7,
	})
	if err != nil {
		return nil, err
	}
	tags, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "tags",
		Name:      "OrderFoodTags",
		Component: "view/orderFood/tag/index.vue",
		Title:     "标签管理",
		Icon:      "price-tag",
		Sort:      8,
	})
	if err != nil {
		return nil, err
	}
	units, err := ensureOrderFoodMenu(db, dishesGroup.ID, 1, orderFoodMenuSeed{
		Path:      "units",
		Name:      "OrderFoodUnits",
		Component: "view/orderFood/unit/index.vue",
		Title:     "单位管理",
		Icon:      "scale-to-original",
		Sort:      9,
	})
	if err != nil {
		return nil, err
	}
	governanceRecords, err := ensureOrderFoodMenu(db, contentSafetyGroup.ID, 1, orderFoodMenuSeed{
		Path:      "governance-records",
		Name:      "OrderFoodGovernanceRecords",
		Component: "view/orderFood/governance/index.vue",
		Title:     "违规处理记录",
		Icon:      "warning",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	media, err := ensureOrderFoodMenu(db, contentSafetyGroup.ID, 1, orderFoodMenuSeed{
		Path:      "media",
		Name:      "OrderFoodMedia",
		Component: "view/orderFood/media/index.vue",
		Title:     "图片资源",
		Icon:      "picture",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	moderationRecords, err := ensureOrderFoodMenu(db, contentSafetyGroup.ID, 1, orderFoodMenuSeed{
		Path:      "moderation-records",
		Name:      "OrderFoodModerationRecords",
		Component: "view/orderFood/moderation/index.vue",
		Title:     "图片审核记录",
		Icon:      "document-checked",
		Sort:      3,
	})
	if err != nil {
		return nil, err
	}
	moderationConfig, err := ensureOrderFoodMenu(db, contentSafetyGroup.ID, 1, orderFoodMenuSeed{
		Path:      "moderation-config",
		Name:      "OrderFoodModerationConfig",
		Component: "view/orderFood/moderationConfig/index.vue",
		Title:     "图片审核配置",
		Icon:      "setting",
		Sort:      4,
	})
	if err != nil {
		return nil, err
	}
	meals, err := ensureOrderFoodMenu(db, mealGroup.ID, 1, orderFoodMenuSeed{
		Path:      "meals",
		Name:      "OrderFoodMeals",
		Component: "view/orderFood/meal/index.vue",
		Title:     "饭局列表",
		Icon:      "calendar",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	shoppingLists, err := ensureOrderFoodMenu(db, mealGroup.ID, 1, orderFoodMenuSeed{
		Path:      "shopping-lists",
		Name:      "OrderFoodShoppingLists",
		Component: "view/orderFood/shopping/index.vue",
		Title:     "采购清单",
		Icon:      "shopping-cart",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	platformPolicy, err := ensureOrderFoodMenu(db, aiGroup.ID, 1, orderFoodMenuSeed{
		Path:      "platform-policy",
		Name:      "OrderFoodPlatformPolicy",
		Component: "view/orderFood/platformPolicy/index.vue",
		Title:     "平台策略",
		Icon:      "switch",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	aiCapabilities, err := ensureOrderFoodMenu(db, aiGroup.ID, 1, orderFoodMenuSeed{
		Path:      "capabilities",
		Name:      "OrderFoodAiCapabilityConfig",
		Component: "view/orderFood/aiCapability/index.vue",
		Title:     "能力配置",
		Icon:      "cpu",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	aiUsages, err := ensureOrderFoodMenu(db, aiGroup.ID, 1, orderFoodMenuSeed{
		Path:      "ai-usages",
		Name:      "OrderFoodAIUsages",
		Component: "view/orderFood/aiUsage/index.vue",
		Title:     "AI 调用记录",
		Icon:      "data-line",
		Sort:      3,
	})
	if err != nil {
		return nil, err
	}
	aiProviders, err := ensureOrderFoodMenu(db, aiGroup.ID, 1, orderFoodMenuSeed{
		Path:      "providers",
		Name:      "OrderFoodAiProviders",
		Component: "view/orderFood/aiProvider/index.vue",
		Title:     "供应商管理",
		Icon:      "connection",
		Sort:      4,
	})
	if err != nil {
		return nil, err
	}
	aiModels, err := ensureOrderFoodMenu(db, aiGroup.ID, 1, orderFoodMenuSeed{
		Path:      "models",
		Name:      "OrderFoodAiModels",
		Component: "view/orderFood/aiModel/index.vue",
		Title:     "模型管理",
		Icon:      "set-up",
		Sort:      5,
	})
	if err != nil {
		return nil, err
	}
	notifications, err := ensureOrderFoodMenu(db, messageGroup.ID, 1, orderFoodMenuSeed{
		Path:      "notifications",
		Name:      "OrderFoodNotifications",
		Component: "view/orderFood/notification/index.vue",
		Title:     "站内通知",
		Icon:      "bell",
		Sort:      1,
	})
	if err != nil {
		return nil, err
	}
	subscribeLogs, err := ensureOrderFoodMenu(db, messageGroup.ID, 1, orderFoodMenuSeed{
		Path:      "subscribe-logs",
		Name:      "OrderFoodSubscribeLogs",
		Component: "view/orderFood/subscribeLog/index.vue",
		Title:     "订阅消息记录",
		Icon:      "tickets",
		Sort:      2,
	})
	if err != nil {
		return nil, err
	}
	subscribeTemplates, err := ensureOrderFoodMenu(db, messageGroup.ID, 1, orderFoodMenuSeed{
		Path:      "subscribe-templates",
		Name:      "OrderFoodSubscribeTemplates",
		Component: "view/orderFood/subscribeTemplate/index.vue",
		Title:     "订阅消息模板",
		Icon:      "document",
		Sort:      3,
	})
	if err != nil {
		return nil, err
	}
	return map[string]uint{
		"dashboard":          dashboard.ID,
		"usersGroup":         usersGroup.ID,
		"dishesGroup":        dishesGroup.ID,
		"contentSafetyGroup": contentSafetyGroup.ID,
		"mealGroup":          mealGroup.ID,
		"aiGroup":            aiGroup.ID,
		"messageGroup":       messageGroup.ID,
		"users":              users.ID,
		"pointEntries":       pointEntries.ID,
		"pointRules":         pointRules.ID,
		"userDishes":         userDishes.ID,
		"userRecipes":        userRecipes.ID,
		"suggestionCatalog":  suggestionCatalog.ID,
		"discoverableDishes": discoverableDishes.ID,
		"recommendations":    recommendations.ID,
		"officialDishes":     officialDishes.ID,
		"categories":         categories.ID,
		"tags":               tags.ID,
		"units":              units.ID,
		"governanceRecords":  governanceRecords.ID,
		"media":              media.ID,
		"moderationRecords":  moderationRecords.ID,
		"moderationConfig":   moderationConfig.ID,
		"meals":              meals.ID,
		"shoppingLists":      shoppingLists.ID,
		"platformPolicy":     platformPolicy.ID,
		"aiCapabilities":     aiCapabilities.ID,
		"aiUsages":           aiUsages.ID,
		"aiProviders":        aiProviders.ID,
		"aiModels":           aiModels.ID,
		"notifications":      notifications.ID,
		"subscribeLogs":      subscribeLogs.ID,
		"subscribeTemplates": subscribeTemplates.ID,
		"wechatConfig":       wechatConfig.ID,
	}, nil
}

func ensureOrderFoodMenu(
	db *gorm.DB,
	parentID uint,
	menuLevel uint,
	seed orderFoodMenuSeed,
) (system.SysBaseMenu, error) {
	var menu system.SysBaseMenu
	err := db.Unscoped().Where("name = ?", seed.Name).First(&menu).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return menu, err
	}
	values := map[string]interface{}{
		"deleted_at": nil,
		"menu_level": menuLevel,
		"parent_id":  parentID,
		"path":       seed.Path,
		"name":       seed.Name,
		"hidden":     false,
		"component":  seed.Component,
		"sort":       seed.Sort,
		"title":      seed.Title,
		"icon":       seed.Icon,
		"keep_alive": true,
	}
	if err == gorm.ErrRecordNotFound {
		menu = system.SysBaseMenu{
			MenuLevel: menuLevel,
			ParentId:  parentID,
			Path:      seed.Path,
			Name:      seed.Name,
			Hidden:    false,
			Component: seed.Component,
			Sort:      seed.Sort,
			Meta: system.Meta{
				Title:     seed.Title,
				Icon:      seed.Icon,
				KeepAlive: true,
			},
		}
		return menu, db.Create(&menu).Error
	}
	return menu, db.Unscoped().Model(&menu).Updates(values).Error
}

// ensureOrderFoodMenuAccess 注入新角色的默认菜单，并始终补齐首个管理员菜单。
func ensureOrderFoodMenuAccess(
	db *gorm.DB,
	menuIDs map[string]uint,
	newAuthorities map[uint]bool,
) error {
	assignments := map[string][]uint{
		"dashboard":          orderFoodDashboardRoles,
		"usersGroup":         orderFoodAllRoles,
		"users":              orderFoodAllRoles,
		"pointEntries":       orderFoodAllRoles,
		"pointRules":         orderFoodSuperRoles,
		"dishesGroup":        orderFoodDishRoles,
		"contentSafetyGroup": orderFoodDishRoles,
		"userDishes":         orderFoodGovernanceRoles,
		"userRecipes":        orderFoodGovernanceRoles,
		"suggestionCatalog":  orderFoodSuperRoles,
		"discoverableDishes": orderFoodDishRoles,
		"recommendations":    orderFoodDishRoles,
		"officialDishes":     orderFoodCatalogRoles,
		"categories":         orderFoodCatalogRoles,
		"tags":               orderFoodCatalogRoles,
		"units":              orderFoodCatalogRoles,
		"governanceRecords":  orderFoodGovernanceRoles,
		"media":              orderFoodDishRoles,
		"moderationRecords":  orderFoodGovernanceRoles,
		"moderationConfig":   orderFoodGovernanceRoles,
		"mealGroup":          orderFoodAllRoles,
		"meals":              orderFoodAllRoles,
		"shoppingLists":      orderFoodAllRoles,
		"aiGroup":            orderFoodAIRoles,
		"platformPolicy":     orderFoodSuperRoles,
		"aiCapabilities":     orderFoodAIRoles,
		"aiUsages":           orderFoodAIUsageRoles,
		"aiProviders":        orderFoodAIRoles,
		"aiModels":           orderFoodAIRoles,
		"messageGroup":       orderFoodAllRoles,
		"notifications":      orderFoodAllRoles,
		"subscribeLogs":      orderFoodMessageLogRoles,
		"subscribeTemplates": orderFoodSuperRoles,
		"wechatConfig":       orderFoodSuperRoles,
	}
	for key, roles := range assignments {
		for _, authorityID := range roles {
			if !shouldSeedOrderFoodAuthority(authorityID, newAuthorities) {
				continue
			}
			row := system.SysAuthorityMenu{
				MenuId:      strconv.FormatUint(uint64(menuIDs[key]), 10),
				AuthorityId: strconv.FormatUint(uint64(authorityID), 10),
			}
			if err := db.
				Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", row.MenuId, row.AuthorityId).
				FirstOrCreate(&row).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureOrderFoodButtons 创建按钮定义，并注入新角色及首个管理员的默认按钮权限。
func ensureOrderFoodButtons(
	db *gorm.DB,
	menuIDs map[string]uint,
	newAuthorities map[uint]bool,
) error {
	type buttonSeed struct {
		MenuKey     string
		Permission  string
		Description string
		Roles       []uint
	}
	buttons := []buttonSeed{
		{"pointEntries", "orderfood:user:read", "查看用户", orderFoodAllRoles},
		{"users", "orderfood:user:preference:read", "查看偏好画像", orderFoodSuperRoles},
		{"users", "orderfood:user:disable", "禁用或恢复用户", orderFoodGovernanceRoles},
		{"users", "orderfood:points:adjust", "调整用户积分", orderFoodSuperRoles},
		{"pointEntries", "orderfood:points:adjust", "调整用户积分", orderFoodSuperRoles},
		{"pointRules", "orderfood:point-rule:update", "保存积分规则并立即生效", orderFoodSuperRoles},
		{"userDishes", "orderfood:user-dish:private-read", "查看私有菜品详情", orderFoodGovernanceRoles},
		{"userDishes", "orderfood:governance:execute", "处理违规菜品", orderFoodGovernanceRoles},
		{"userDishes", "orderfood:governance:cascade", "处理严重违规复制链", orderFoodSuperRoles},
		{"userRecipes", "orderfood:user-recipe:private-read", "查看私有菜谱详情", orderFoodGovernanceRoles},
		{"userRecipes", "orderfood:governance:execute", "处理违规菜谱", orderFoodGovernanceRoles},
		{"suggestionCatalog", "orderfood:suggestion-catalog:update", "编辑标准菜品索引、食材词库或立即生效的校验策略", orderFoodSuperRoles},
		{"discoverableDishes", "orderfood:recommendation:create", "将可发现菜品加入推荐草稿", orderFoodCatalogRoles},
		{"recommendations", "orderfood:recommendation:create", "创建推荐草稿", orderFoodCatalogRoles},
		{"recommendations", "orderfood:recommendation:update", "编辑推荐排序和展示说明", orderFoodCatalogRoles},
		{"recommendations", "orderfood:recommendation:delete", "删除推荐草稿", orderFoodCatalogRoles},
		{"recommendations", "orderfood:recommendation:publish", "发布或重新发布推荐", orderFoodCatalogRoles},
		{"recommendations", "orderfood:recommendation:offline", "下线推荐", orderFoodDishRoles},
		{"recommendations", "orderfood:recommendation:sort", "批量调整推荐排序", orderFoodCatalogRoles},
		{"officialDishes", "orderfood:official-dish:create", "新增官方菜品", orderFoodCatalogRoles},
		{"officialDishes", "orderfood:official-dish:update", "编辑官方菜品", orderFoodCatalogRoles},
		{"officialDishes", "orderfood:official-dish:delete", "软删除官方菜品", orderFoodCatalogRoles},
		{"officialDishes", "orderfood:official-dish:cover-upload", "上传官方菜品封面", orderFoodCatalogRoles},
		{"categories", "orderfood:category:create", "新增分类", orderFoodCatalogRoles},
		{"categories", "orderfood:category:update", "编辑或停用分类", orderFoodCatalogRoles},
		{"categories", "orderfood:category:delete", "删除未被引用的分类", orderFoodCatalogRoles},
		{"categories", "orderfood:category:sort", "批量调整分类排序", orderFoodCatalogRoles},
		{"tags", "orderfood:tag:create", "新增标签", orderFoodCatalogRoles},
		{"tags", "orderfood:tag:update", "编辑或停用标签", orderFoodCatalogRoles},
		{"tags", "orderfood:tag:delete", "删除未被引用的标签", orderFoodCatalogRoles},
		{"tags", "orderfood:tag:sort", "批量调整标签排序", orderFoodCatalogRoles},
		{"units", "orderfood:unit:create", "新增单位", orderFoodCatalogRoles},
		{"units", "orderfood:unit:update", "编辑或停用单位", orderFoodCatalogRoles},
		{"units", "orderfood:unit:delete", "删除未被引用的单位", orderFoodCatalogRoles},
		{"units", "orderfood:unit:sort", "批量调整单位排序", orderFoodCatalogRoles},
		{"governanceRecords", "orderfood:governance:retry", "重试违规处理任务失败项", orderFoodGovernanceRoles},
		{"governanceRecords", "orderfood:governance:cascade", "处理严重违规复制链", orderFoodSuperRoles},
		{"moderationRecords", "orderfood:moderation:sensitive-read", "查看供应商响应脱敏摘要", orderFoodGovernanceRoles},
		{"moderationConfig", "orderfood:moderation-config:update", "保存图片审核配置并立即生效", orderFoodSuperRoles},
		{"moderationConfig", "orderfood:moderation-config:credential-write", "更新图片审核凭据引用", orderFoodSuperRoles},
		{"moderationConfig", "orderfood:moderation-config:test", "测试图片审核连接和样例图片", orderFoodSuperRoles},
		{"platformPolicy", "orderfood:platform-policy:update", "保存平台策略或切换紧急状态并立即生效", orderFoodSuperRoles},
		{"aiCapabilities", "orderfood:capability:update", "保存能力配置并立即生效", orderFoodSuperRoles},
		{"aiCapabilities", "orderfood:prompt:read", "查看提示词正文", orderFoodSuperRoles},
		{"aiCapabilities", "orderfood:prompt:update", "编辑提示词", orderFoodSuperRoles},
		{"aiCapabilities", "orderfood:prompt:test", "校验和测试提示词", orderFoodSuperRoles},
		{"aiUsages", "orderfood:ai-usage:sensitive-read", "查看AI调用原始输入和模型输出", orderFoodSuperRoles},
		{"aiUsages", "orderfood:ai-usage:sensitive-delete", "永久清除AI调用敏感内容", orderFoodSuperRoles},
		{"aiProviders", "orderfood:provider:create", "新增供应商", orderFoodSuperRoles},
		{"aiProviders", "orderfood:provider:update", "编辑供应商", orderFoodSuperRoles},
		{"aiProviders", "orderfood:provider:status", "启用或停用供应商", orderFoodSuperRoles},
		{"aiProviders", "orderfood:provider:delete", "删除供应商", orderFoodSuperRoles},
		{"aiProviders", "orderfood:provider:test", "测试供应商连接", orderFoodSuperRoles},
		{"aiProviders", "orderfood:provider:credential-write", "更新供应商凭据引用", orderFoodSuperRoles},
		{"aiModels", "orderfood:model:create", "新增模型", orderFoodSuperRoles},
		{"aiModels", "orderfood:model:update", "编辑模型", orderFoodSuperRoles},
		{"aiModels", "orderfood:model:status", "启用或停用模型", orderFoodSuperRoles},
		{"aiModels", "orderfood:model:delete", "删除模型", orderFoodSuperRoles},
		{"subscribeTemplates", "orderfood:subscribe-template:create", "新增订阅消息模板", orderFoodSuperRoles},
		{"subscribeTemplates", "orderfood:subscribe-template:update", "编辑订阅消息模板", orderFoodSuperRoles},
		{"subscribeTemplates", "orderfood:subscribe-template:status", "启用或停用订阅消息模板", orderFoodSuperRoles},
		{"subscribeTemplates", "orderfood:subscribe-template:delete", "删除未使用的订阅消息模板", orderFoodSuperRoles},
		{"wechatConfig", "orderfood:wechat-config:update", "保存微信小程序配置并立即生效", orderFoodSuperRoles},
	}
	for _, seed := range buttons {
		menuID := menuIDs[seed.MenuKey]
		var button system.SysBaseMenuBtn
		if err := db.
			Where("sys_base_menu_id = ? AND name = ?", menuID, seed.Permission).
			Assign(system.SysBaseMenuBtn{Desc: seed.Description}).
			FirstOrCreate(&button, system.SysBaseMenuBtn{
				Name:          seed.Permission,
				Desc:          seed.Description,
				SysBaseMenuID: menuID,
			}).Error; err != nil {
			return err
		}
		for _, authorityID := range seed.Roles {
			if !shouldSeedOrderFoodAuthority(authorityID, newAuthorities) {
				continue
			}
			authorization := system.SysAuthorityBtn{
				AuthorityId:      authorityID,
				SysMenuID:        menuID,
				SysBaseMenuBtnID: button.ID,
			}
			if err := db.
				Where(
					"authority_id = ? AND sys_menu_id = ? AND sys_base_menu_btn_id = ?",
					authorityID,
					menuID,
					button.ID,
				).
				FirstOrCreate(&authorization).Error; err != nil {
				return err
			}
		}
	}
	return ensureAllOrderFoodPermissionButtons(db, menuIDs, newAuthorities)
}

// ensureAllOrderFoodPermissionButtons 为全部业务权限创建可在 GVA 中配置的按钮定义。
func ensureAllOrderFoodPermissionButtons(
	db *gorm.DB,
	menuIDs map[string]uint,
	newAuthorities map[uint]bool,
) error {
	matrix := orderfoodService.DefaultRolePermissionMatrix()
	permissions := make(map[string]struct{})
	for _, rolePermissions := range matrix {
		for _, permission := range rolePermissions {
			permissions[permission] = struct{}{}
		}
	}
	orderedPermissions := make([]string, 0, len(permissions))
	for permission := range permissions {
		orderedPermissions = append(orderedPermissions, permission)
	}
	sort.Strings(orderedPermissions)
	for _, permission := range orderedPermissions {
		menuKey := orderFoodPermissionMenuKey(permission)
		menuID := menuIDs[menuKey]
		if menuID == 0 {
			return errors.New("orderfood permission has no menu mapping: " + permission)
		}
		var button system.SysBaseMenuBtn
		if err := db.
			Where("sys_base_menu_id = ? AND name = ?", menuID, permission).
			Assign(system.SysBaseMenuBtn{Desc: orderFoodPermissionDescription(permission)}).
			FirstOrCreate(&button, system.SysBaseMenuBtn{
				Name:          permission,
				Desc:          orderFoodPermissionDescription(permission),
				SysBaseMenuID: menuID,
			}).Error; err != nil {
			return err
		}
		for authorityID, rolePermissions := range matrix {
			if !shouldSeedOrderFoodAuthority(authorityID, newAuthorities) ||
				!containsOrderFoodPermission(rolePermissions, permission) {
				continue
			}
			authorization := system.SysAuthorityBtn{
				AuthorityId:      authorityID,
				SysMenuID:        menuID,
				SysBaseMenuBtnID: button.ID,
			}
			if err := db.Where(
				"authority_id = ? AND sys_menu_id = ? AND sys_base_menu_btn_id = ?",
				authorityID,
				menuID,
				button.ID,
			).FirstOrCreate(&authorization).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// containsOrderFoodPermission 判断默认角色模板是否包含指定权限。
func containsOrderFoodPermission(permissions []string, target string) bool {
	for _, permission := range permissions {
		if permission == target {
			return true
		}
	}
	return false
}

// orderFoodPermissionMenuKey 返回权限在 GVA 中归属的叶子菜单。
func orderFoodPermissionMenuKey(permission string) string {
	// 偏好画像是用户详情内的独立敏感权限，沿用四段权限码并归属用户菜单。
	if permission == "orderfood:user:preference:read" {
		return "users"
	}
	parts := strings.Split(permission, ":")
	if len(parts) != 3 || parts[0] != "orderfood" {
		return ""
	}
	return map[string]string{
		"ai-usage":           "aiUsages",
		"audit":              "governanceRecords",
		"prompt":             "aiCapabilities",
		"capability":         "aiCapabilities",
		"category":           "categories",
		"dashboard":          "dashboard",
		"discoverable-dish":  "discoverableDishes",
		"governance":         "governanceRecords",
		"meal":               "meals",
		"media":              "media",
		"model":              "aiModels",
		"moderation-config":  "moderationConfig",
		"moderation":         "moderationRecords",
		"notification":       "notifications",
		"official-dish":      "officialDishes",
		"platform-policy":    "platformPolicy",
		"points":             "pointEntries",
		"point-rule":         "pointRules",
		"provider":           "aiProviders",
		"recommendation":     "recommendations",
		"shopping":           "shoppingLists",
		"suggestion-catalog": "suggestionCatalog",
		"subscribe-log":      "subscribeLogs",
		"subscribe-template": "subscribeTemplates",
		"tag":                "tags",
		"unit":               "units",
		"user-dish":          "userDishes",
		"user-recipe":        "userRecipes",
		"user":               "users",
		"wechat-config":      "wechatConfig",
	}[parts[1]]
}

// orderFoodPermissionDescription 生成权限按钮在 GVA 中展示的简短说明。
func orderFoodPermissionDescription(permission string) string {
	if permission == "orderfood:user:preference:read" {
		return "查看偏好画像"
	}
	parts := strings.Split(permission, ":")
	if len(parts) != 3 {
		return permission
	}
	action := map[string]string{
		"read":             "查看",
		"create":           "新增",
		"update":           "编辑",
		"delete":           "删除",
		"sort":             "调整排序",
		"status":           "启停",
		"test":             "测试",
		"execute":          "执行",
		"retry":            "重试",
		"publish":          "发布",
		"offline":          "下线",
		"disable":          "禁用或恢复",
		"adjust":           "调整",
		"cascade":          "级联处理",
		"private-read":     "查看私有详情",
		"sensitive-read":   "查看敏感详情",
		"sensitive-delete": "清除敏感内容",
		"credential-write": "更新凭据",
		"cover-upload":     "上传封面",
	}[parts[2]]
	if action == "" {
		action = parts[2]
	}
	return action + "权限"
}

// ensureImplementedOrderFoodRoutes 创建 API 定义，并注入新角色及首个管理员的默认路由权限。
func ensureImplementedOrderFoodRoutes(
	db *gorm.DB,
	newAuthorities map[uint]bool,
) error {
	const base = "/orderfood"
	routes := []orderFoodAdminRouteSeed{
		{"GET", base + "/dashboard", "获取来干饭运营概览", orderFoodDashboardRoles},
		{"GET", base + "/users", "分页查询小程序用户", orderFoodAllRoles},
		{"GET", base + "/users/:userId", "获取小程序用户详情", orderFoodAllRoles},
		{"PUT", base + "/users/:userId/status", "禁用或恢复小程序用户", orderFoodGovernanceRoles},
		{"GET", base + "/users/:userId/preference-profile", "获取用户偏好画像", orderFoodSuperRoles},
		{"GET", base + "/point-entries", "分页查询积分流水", orderFoodAllRoles},
		{"POST", base + "/point-adjustments/preview", "预览人工积分调整", orderFoodSuperRoles},
		{"POST", base + "/point-adjustments", "提交人工积分调整", orderFoodSuperRoles},
		{"GET", base + "/point-rules", "获取当前积分规则", orderFoodSuperRoles},
		{"PUT", base + "/point-rules", "保存积分规则并立即生效", orderFoodSuperRoles},
		{"GET", base + "/user-dishes", "分页查询全部用户菜品", orderFoodGovernanceRoles},
		{"GET", base + "/user-dishes/:dishId", "获取用户菜品私有详情", orderFoodGovernanceRoles},
		{"GET", base + "/user-dishes/:dishId/references", "分页查询用户菜品引用位置", orderFoodGovernanceRoles},
		{"GET", base + "/user-recipes", "分页查询全部用户菜谱", orderFoodGovernanceRoles},
		{"GET", base + "/user-recipes/:recipeId", "获取用户菜谱私有详情", orderFoodGovernanceRoles},
		{"GET", base + "/suggestion-catalog", "获取标准菜品索引工作区", orderFoodSuperRoles},
		{"PUT", base + "/suggestion-catalog/policy", "保存生成结果索引校验策略并立即生效", orderFoodSuperRoles},
		{"GET", base + "/suggestion-catalog/ingredients", "分页查询标准食材词库", orderFoodSuperRoles},
		{"POST", base + "/suggestion-catalog/ingredients", "新增标准食材", orderFoodSuperRoles},
		{"PUT", base + "/suggestion-catalog/ingredients/:id", "编辑标准食材", orderFoodSuperRoles},
		{"GET", base + "/suggestion-catalog/dishes", "分页查询标准菜品索引", orderFoodSuperRoles},
		{"POST", base + "/suggestion-catalog/dishes", "新增标准菜品索引", orderFoodSuperRoles},
		{"PUT", base + "/suggestion-catalog/dishes/:id", "编辑标准菜品索引", orderFoodSuperRoles},
		{"GET", base + "/discoverable-dishes", "分页查询可发现候选菜品", orderFoodDishRoles},
		{"GET", base + "/discoverable-dishes/:dishId", "获取可发现候选菜品详情", orderFoodDishRoles},
		{"GET", base + "/recommendations", "分页查询推荐精选", orderFoodDishRoles},
		{"GET", base + "/recommendations/:recommendationId", "获取推荐精选详情", orderFoodDishRoles},
		{"POST", base + "/recommendations", "创建推荐草稿", orderFoodCatalogRoles},
		{"PUT", base + "/recommendations/sort-order", "批量调整推荐排序", orderFoodCatalogRoles},
		{"PUT", base + "/recommendations/:recommendationId", "编辑推荐信息", orderFoodCatalogRoles},
		{"DELETE", base + "/recommendations/:recommendationId", "删除推荐草稿", orderFoodCatalogRoles},
		{"POST", base + "/recommendations/:recommendationId/publish", "发布或重新发布推荐", orderFoodCatalogRoles},
		{"POST", base + "/recommendations/:recommendationId/offline", "下线推荐", orderFoodDishRoles},
		{"GET", base + "/official-dishes", "分页查询官方菜品", orderFoodCatalogRoles},
		{"GET", base + "/official-dishes/:dishId", "获取官方菜品详情", orderFoodCatalogRoles},
		{"POST", base + "/official-dishes", "创建官方菜品", orderFoodCatalogRoles},
		{"PUT", base + "/official-dishes/:dishId", "编辑官方菜品", orderFoodCatalogRoles},
		{"DELETE", base + "/official-dishes/:dishId", "软删除官方菜品", orderFoodCatalogRoles},
		{"GET", base + "/official-dish-covers", "分页查询官方菜品封面", orderFoodCatalogRoles},
		{"POST", base + "/official-dish-covers", "上传官方菜品封面", orderFoodCatalogRoles},
		{"GET", base + "/categories", "分页查询分类", orderFoodCatalogRoles},
		{"POST", base + "/categories", "新增分类", orderFoodCatalogRoles},
		{"PUT", base + "/categories/sort-order", "批量调整分类排序", orderFoodCatalogRoles},
		{"PUT", base + "/categories/:categoryId", "编辑分类", orderFoodCatalogRoles},
		{"DELETE", base + "/categories/:categoryId", "删除分类", orderFoodCatalogRoles},
		{"GET", base + "/tags", "分页查询标签", orderFoodCatalogRoles},
		{"POST", base + "/tags", "新增标签", orderFoodCatalogRoles},
		{"PUT", base + "/tags/sort-order", "批量调整标签排序", orderFoodCatalogRoles},
		{"PUT", base + "/tags/:tagId", "编辑标签", orderFoodCatalogRoles},
		{"DELETE", base + "/tags/:tagId", "删除标签", orderFoodCatalogRoles},
		{"GET", base + "/units", "分页查询单位", orderFoodCatalogRoles},
		{"POST", base + "/units", "新增单位", orderFoodCatalogRoles},
		{"PUT", base + "/units/sort-order", "批量调整单位排序", orderFoodCatalogRoles},
		{"PUT", base + "/units/:unitId", "编辑单位", orderFoodCatalogRoles},
		{"DELETE", base + "/units/:unitId", "删除单位", orderFoodCatalogRoles},
		{"GET", base + "/media", "分页查询图片资源", orderFoodDishRoles},
		{"GET", base + "/media/:fileId", "获取图片资源详情", orderFoodDishRoles},
		{"GET", base + "/moderation-records", "分页查询图片审核记录", orderFoodGovernanceRoles},
		{"GET", base + "/moderation-records/:recordId", "获取图片审核记录详情", orderFoodGovernanceRoles},
		{"GET", base + "/moderation-config", "获取当前图片审核配置", orderFoodGovernanceRoles},
		{"PUT", base + "/moderation-config", "保存图片审核配置并立即生效", orderFoodSuperRoles},
		{"POST", base + "/moderation-config/connection-tests", "测试图片审核连接", orderFoodSuperRoles},
		{"POST", base + "/moderation-config/moderation-tests", "测试图片审核样例", orderFoodSuperRoles},
		{"GET", base + "/governance-records", "分页查询违规处理记录", orderFoodGovernanceRoles},
		{"GET", base + "/governance-records/:recordId", "获取违规处理记录详情", orderFoodGovernanceRoles},
		{"GET", base + "/governance-jobs/:jobId", "获取违规处理任务进度", orderFoodGovernanceRoles},
		{"POST", base + "/governance-jobs/:jobId/retry", "重试违规处理任务失败项", orderFoodSuperRoles},
		{"POST", base + "/governance-actions/preview", "预览违规处理动作影响", orderFoodGovernanceRoles},
		{"POST", base + "/governance-actions", "执行违规处理动作", orderFoodGovernanceRoles},
		{"GET", base + "/meals", "分页查询饭局", orderFoodAllRoles},
		{"GET", base + "/meals/:mealId", "获取饭局详情", orderFoodAllRoles},
		{"GET", base + "/shopping-lists", "分页查询采购清单", orderFoodAllRoles},
		{"GET", base + "/shopping-lists/:shoppingListId", "获取采购清单详情", orderFoodAllRoles},
		{"GET", base + "/audit-logs", "分页查询来干饭管理操作日志", orderFoodGovernanceRoles},
		{"GET", base + "/audit-logs/:auditLogId", "获取来干饭管理操作日志详情", orderFoodGovernanceRoles},
		{"GET", base + "/platform-capability-policy", "获取平台整体能力策略", orderFoodSuperRoles},
		{"PUT", base + "/platform-capability-policy", "保存平台整体能力策略并立即生效", orderFoodSuperRoles},
		{"PUT", base + "/platform-capability-policy/emergency-status", "设置平台能力紧急关闭状态", orderFoodSuperRoles},
		{"GET", base + "/ai-providers", "分页查询 AI 供应商", orderFoodAIRoles},
		{"GET", base + "/ai-providers/:providerId", "获取 AI 供应商详情", orderFoodAIRoles},
		{"POST", base + "/ai-providers", "新增 AI 供应商", orderFoodSuperRoles},
		{"PUT", base + "/ai-providers/:providerId", "编辑 AI 供应商", orderFoodSuperRoles},
		{"PUT", base + "/ai-providers/:providerId/status", "启用或停用 AI 供应商", orderFoodSuperRoles},
		{"DELETE", base + "/ai-providers/:providerId", "删除 AI 供应商", orderFoodSuperRoles},
		{"POST", base + "/ai-providers/:providerId/connection-tests", "测试 AI 供应商连接", orderFoodSuperRoles},
		{"GET", base + "/ai-models", "分页查询 AI 模型", orderFoodAIRoles},
		{"GET", base + "/ai-models/:modelId", "获取 AI 模型详情", orderFoodAIRoles},
		{"POST", base + "/ai-models", "新增 AI 模型", orderFoodSuperRoles},
		{"PUT", base + "/ai-models/:modelId", "编辑 AI 模型", orderFoodSuperRoles},
		{"PUT", base + "/ai-models/:modelId/status", "启用或停用 AI 模型", orderFoodSuperRoles},
		{"DELETE", base + "/ai-models/:modelId", "删除 AI 模型", orderFoodSuperRoles},
		{"GET", base + "/ai-capabilities", "查询 AI 能力配置", orderFoodAIRoles},
		{"GET", base + "/ai-capabilities/:capabilityCode", "获取 AI 能力配置详情", orderFoodAIRoles},
		{"PUT", base + "/ai-capabilities/:capabilityCode", "保存 AI 能力配置并立即生效", orderFoodSuperRoles},
		{"GET", base + "/ai-capabilities/:capabilityCode/prompts", "获取 AI 提示词工作区", orderFoodSuperRoles},
		{"PUT", base + "/ai-capabilities/:capabilityCode/prompts", "保存 AI 提示词并立即生效", orderFoodSuperRoles},
		{"POST", base + "/ai-capabilities/:capabilityCode/prompt-validation", "校验 AI 提示词", orderFoodSuperRoles},
		{"POST", base + "/ai-capabilities/:capabilityCode/prompt-render-preview", "预览 AI 提示词渲染结果", orderFoodSuperRoles},
		{"POST", base + "/ai-capabilities/:capabilityCode/prompt-tests", "测试 AI 提示词", orderFoodSuperRoles},
		{"GET", base + "/ai-usages", "分页查询 AI 调用记录", orderFoodAIUsageRoles},
		{"GET", base + "/ai-usages/:usageId", "获取 AI 调用记录详情", orderFoodAIUsageRoles},
		{"DELETE", base + "/ai-usages/:usageId/sensitive-content", "永久清除 AI 调用敏感内容", orderFoodSuperRoles},
		{"GET", base + "/notifications", "分页查询站内通知投递", orderFoodAllRoles},
		{"GET", base + "/notifications/:notificationId", "获取站内通知投递详情", orderFoodAllRoles},
		{"GET", base + "/subscribe-templates", "分页查询订阅消息模板", orderFoodSuperRoles},
		{"GET", base + "/subscribe-templates/:templateId", "获取订阅消息模板详情", orderFoodSuperRoles},
		{"POST", base + "/subscribe-templates", "创建订阅消息模板", orderFoodSuperRoles},
		{"PUT", base + "/subscribe-templates/:templateId", "编辑订阅消息模板", orderFoodSuperRoles},
		{"PUT", base + "/subscribe-templates/:templateId/status", "启用或停用订阅消息模板", orderFoodSuperRoles},
		{"DELETE", base + "/subscribe-templates/:templateId", "删除未使用的订阅消息模板", orderFoodSuperRoles},
		{"GET", base + "/subscribe-logs", "分页查询订阅消息发送记录", orderFoodMessageLogRoles},
		{"GET", base + "/subscribe-logs/:logId", "获取订阅消息发送详情", orderFoodMessageLogRoles},
		{"GET", base + "/wechat-config", "获取微信小程序配置", orderFoodSuperRoles},
		{"PUT", base + "/wechat-config", "保存微信小程序配置并立即生效", orderFoodSuperRoles},
	}
	for _, seed := range routes {
		api := system.SysApi{
			Path:        seed.Path,
			Description: seed.Description,
			ApiGroup:    "来干饭管理端",
			Method:      seed.Method,
		}
		if err := db.
			Where("path = ? AND method = ?", seed.Path, seed.Method).
			Assign(map[string]interface{}{
				"description": seed.Description,
				"api_group":   "来干饭管理端",
			}).
			FirstOrCreate(&api).Error; err != nil {
			return err
		}
		for _, authorityID := range seed.Roles {
			if !shouldSeedOrderFoodAuthority(authorityID, newAuthorities) {
				continue
			}
			rule := adapter.CasbinRule{
				Ptype: "p",
				V0:    strconv.FormatUint(uint64(authorityID), 10),
				V1:    seed.Path,
				V2:    seed.Method,
			}
			if err := db.
				Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).
				FirstOrCreate(&rule).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// shouldSeedOrderFoodAuthority 判断是否应注入默认授权。
// 平台首个管理员始终补齐全部权限，其他角色只在首次创建时注入默认模板。
func shouldSeedOrderFoodAuthority(authorityID uint, newAuthorities map[uint]bool) bool {
	return authorityID == orderfoodModel.AuthorityPlatformSuperAdmin ||
		newAuthorities[authorityID]
}

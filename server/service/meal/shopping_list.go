package meal

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	mealRequest "github.com/dyjh/order-food-mini-app/server/model/meal/request"
	mealResponse "github.com/dyjh/order-food-mini-app/server/model/meal/response"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"gorm.io/gorm"
)

// PermissionShoppingRead 表示读取采购清单管理数据所需的权限。
const PermissionShoppingRead = "orderfood:shopping:read"

// ShoppingListService 提供管理端采购清单只读业务能力。
type ShoppingListService struct {
	DB         *gorm.DB                         // 数据库连接
	Permission *serviceCommon.PermissionService // 管理端权限服务
	Now        func() time.Time                 // 当前时间函数
}

// NewShoppingListService 创建管理端采购清单查询服务实例。
func NewShoppingListService(
	db *gorm.DB,
	permission *serviceCommon.PermissionService,
) *ShoppingListService {
	if permission == nil {
		permission = serviceCommon.NewPermissionService(db)
	}
	return &ShoppingListService{DB: db, Permission: permission, Now: time.Now}
}

// database 返回服务使用的数据库连接。
func (service *ShoppingListService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回服务使用的UTC时间。
func (service *ShoppingListService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// ListShoppingLists 分页查询采购清单。
func (service *ShoppingListService) ListShoppingLists(
	ctx context.Context,
	query mealRequest.ShoppingListAdminListQuery,
	actor commonRequest.AdminActor,
) (commonResponse.Page[mealResponse.ShoppingListAdminSummary], error) {
	query.ApplyDefaults()
	if service == nil || service.Permission == nil {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionShoppingRead); err != nil {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, err
	}
	if err := validateShoppingListAdminListQuery(query); err != nil {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, err
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&mealModel.FrontShoppingList{})
	for _, filter := range []struct {
		column string // 数据库字段
		value  string // 查询值
	}{
		{"id", query.ShoppingListID},
		{"meal_id", query.MealID},
		{"owner_id", query.CreatorID},
	} {
		if value := strings.TrimSpace(filter.value); value != "" {
			statement = statement.Where(filter.column+" = ?", value)
		}
	}
	now := service.now()
	switch query.ShareStatus {
	case "":
	case "active":
		statement = statement.Where(
			"share_revoked_at IS NULL AND (share_expires_at IS NULL OR share_expires_at > ?)", now,
		)
	case "expired":
		statement = statement.Where("share_revoked_at IS NULL AND share_expires_at <= ?", now)
	case "revoked":
		statement = statement.Where("share_revoked_at IS NOT NULL")
	default:
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", query.CreatedFrom.UTC())
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", query.CreatedTo.UTC())
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{},
			appErrors.AdminInternal.Wrap(err, "count admin shopping lists")
	}
	orderExpression := "created_at " + query.SortOrder
	if query.SortBy == "pendingCount" {
		orderExpression = "(SELECT COUNT(*) FROM of_shopping_items AS item WHERE item.list_id = of_shopping_lists.id AND item.completed = 0) " + query.SortOrder
	} else if query.SortBy != "createdAt" {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var lists []mealModel.FrontShoppingList
	if err := statement.
		Order(orderExpression).
		Order("id " + query.SortOrder).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&lists).Error; err != nil {
		return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{},
			appErrors.AdminInternal.Wrap(err, "list admin shopping lists")
	}
	list := make([]mealResponse.ShoppingListAdminSummary, 0, len(lists))
	for index := range lists {
		summary, err := buildShoppingListSummary(ctx, db, lists[index], now)
		if err != nil {
			return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{}, err
		}
		list = append(list, summary)
	}
	return commonResponse.Page[mealResponse.ShoppingListAdminSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetShoppingList 获取采购清单只读详情。
func (service *ShoppingListService) GetShoppingList(
	ctx context.Context,
	shoppingListID string,
	actor commonRequest.AdminActor,
) (mealResponse.ShoppingListAdminDetail, error) {
	if service == nil || service.Permission == nil {
		return mealResponse.ShoppingListAdminDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionShoppingRead); err != nil {
		return mealResponse.ShoppingListAdminDetail{}, err
	}
	db := service.database()
	if db == nil {
		return mealResponse.ShoppingListAdminDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var list mealModel.FrontShoppingList
	if err := db.WithContext(ctx).First(&list, "id = ?", strings.TrimSpace(shoppingListID)).Error; err != nil {
		return mealResponse.ShoppingListAdminDetail{}, mealAdminLookupError(err)
	}
	summary, err := buildShoppingListSummary(ctx, db, list, service.now())
	if err != nil {
		return mealResponse.ShoppingListAdminDetail{}, err
	}
	var rows []mealModel.FrontShoppingItem
	if err := db.WithContext(ctx).Where("list_id = ?", list.ID).
		Order("sort_order ASC, created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return mealResponse.ShoppingListAdminDetail{},
			appErrors.AdminInternal.Wrap(err, "list admin shopping items")
	}
	items := make([]mealResponse.ShoppingItemAdmin, 0, len(rows))
	for index := range rows {
		sourceDishNames := []string{}
		if len(rows[index].SourceDishNames) > 0 && string(rows[index].SourceDishNames) != "null" {
			if err := json.Unmarshal(rows[index].SourceDishNames, &sourceDishNames); err != nil {
				return mealResponse.ShoppingListAdminDetail{},
					appErrors.AdminInternal.Wrap(err, "decode shopping item source dishes")
			}
		}
		items = append(items, mealResponse.ShoppingItemAdmin{
			ID: rows[index].ID, Name: rows[index].Name, Amount: rows[index].Amount,
			Note: rows[index].Note, Completed: rows[index].Completed,
			SourceDishNames: sourceDishNames, SortOrder: rows[index].SortOrder,
		})
	}
	var snapshots []mealModel.FrontMealFinalDish
	if err := db.WithContext(ctx).Where("meal_id = ?", list.MealID).
		Order("created_at ASC, id ASC").Find(&snapshots).Error; err != nil {
		return mealResponse.ShoppingListAdminDetail{},
			appErrors.AdminInternal.Wrap(err, "list shopping source snapshots")
	}
	snapshotIDs := make([]string, 0, len(snapshots))
	for index := range snapshots {
		snapshotIDs = append(snapshotIDs, snapshots[index].ID)
	}
	var tokenMasked *string
	if value := maskShareToken(list.ShareToken); value != "" {
		tokenMasked = &value
	}
	return mealResponse.ShoppingListAdminDetail{
		ShoppingListAdminSummary: summary, ShareTokenMasked: tokenMasked,
		Items: items, GeneratedFromSnapshotIDs: snapshotIDs,
	}, nil
}

// buildShoppingListSummary 将采购清单转换为管理端摘要。
func buildShoppingListSummary(
	ctx context.Context,
	db *gorm.DB,
	list mealModel.FrontShoppingList,
	now time.Time,
) (mealResponse.ShoppingListAdminSummary, error) {
	var meal mealModel.FrontMeal
	if err := db.WithContext(ctx).First(&meal, "id = ?", list.MealID).Error; err != nil {
		return mealResponse.ShoppingListAdminSummary{}, mealAdminLookupError(err)
	}
	var creator userModel.MiniAppUser
	if err := db.WithContext(ctx).First(&creator, "id = ?", list.OwnerID).Error; err != nil {
		return mealResponse.ShoppingListAdminSummary{}, mealAdminLookupError(err)
	}
	var pendingCount int64
	if err := db.WithContext(ctx).Model(&mealModel.FrontShoppingItem{}).
		Where("list_id = ? AND completed = ?", list.ID, false).Count(&pendingCount).Error; err != nil {
		return mealResponse.ShoppingListAdminSummary{},
			appErrors.AdminInternal.Wrap(err, "count pending shopping items")
	}
	var completedCount int64
	if err := db.WithContext(ctx).Model(&mealModel.FrontShoppingItem{}).
		Where("list_id = ? AND completed = ?", list.ID, true).Count(&completedCount).Error; err != nil {
		return mealResponse.ShoppingListAdminSummary{},
			appErrors.AdminInternal.Wrap(err, "count completed shopping items")
	}
	return mealResponse.ShoppingListAdminSummary{
		ID: list.ID, MealID: list.MealID, MealName: meal.Name, Creator: serviceCommon.UserReference(creator),
		TotalCount:   int(pendingCount + completedCount),
		PendingCount: int(pendingCount), CompletedCount: int(completedCount),
		ShareStatus:    shoppingShareStatus(list, now),
		ShareExpiresAt: list.ShareExpiresAt, ShareRevokedAt: list.ShareRevokedAt,
		CreatedAt: list.CreatedAt, UpdatedAt: list.UpdatedAt,
	}, nil
}

// validateShoppingListAdminListQuery 校验采购清单列表参数并阻止不安全排序表达式。
func validateShoppingListAdminListQuery(
	query mealRequest.ShoppingListAdminListQuery,
) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.ShoppingListID))) > 64 ||
		len([]rune(strings.TrimSpace(query.MealID))) > 64 ||
		len([]rune(strings.TrimSpace(query.CreatorID))) > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, ok := map[string]struct{}{
		"": {}, "active": {}, "expired": {}, "revoked": {},
	}[query.ShareStatus]; !ok {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.SortBy != "createdAt" && query.SortBy != "pendingCount" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.SortOrder != "asc" && query.SortOrder != "desc" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// shoppingShareStatus 返回采购清单分享状态。
func shoppingShareStatus(list mealModel.FrontShoppingList, now time.Time) string {
	if list.ShareRevokedAt != nil {
		return "revoked"
	}
	if list.ShareExpiresAt != nil && !list.ShareExpiresAt.After(now) {
		return "expired"
	}
	return "active"
}

// maskShareToken 返回不可直接使用的分享令牌掩码。
func maskShareToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 6 {
		return "******"
	}
	return "******" + token[len(token)-6:]
}

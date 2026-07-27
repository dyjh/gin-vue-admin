package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// PermissionCategoryRead 表示读取分类所需的权限。
	PermissionCategoryRead = "orderfood:category:read"
	// PermissionCategoryCreate 表示新增分类所需的权限。
	PermissionCategoryCreate = "orderfood:category:create"
	// PermissionCategoryUpdate 表示编辑分类所需的权限。
	PermissionCategoryUpdate = "orderfood:category:update"
	// PermissionCategoryDelete 表示删除分类所需的权限。
	PermissionCategoryDelete = "orderfood:category:delete"
	// PermissionCategorySort 表示调整分类排序所需的权限。
	PermissionCategorySort = "orderfood:category:sort"
	// PermissionTagRead 表示读取标签所需的权限。
	PermissionTagRead = "orderfood:tag:read"
	// PermissionTagCreate 表示新增标签所需的权限。
	PermissionTagCreate = "orderfood:tag:create"
	// PermissionTagUpdate 表示编辑标签所需的权限。
	PermissionTagUpdate = "orderfood:tag:update"
	// PermissionTagDelete 表示删除标签所需的权限。
	PermissionTagDelete = "orderfood:tag:delete"
	// PermissionTagSort 表示调整标签排序所需的权限。
	PermissionTagSort = "orderfood:tag:sort"
	// PermissionUnitRead 表示读取单位所需的权限。
	PermissionUnitRead = "orderfood:unit:read"
	// PermissionUnitCreate 表示新增单位所需的权限。
	PermissionUnitCreate = "orderfood:unit:create"
	// PermissionUnitUpdate 表示编辑单位所需的权限。
	PermissionUnitUpdate = "orderfood:unit:update"
	// PermissionUnitDelete 表示删除单位所需的权限。
	PermissionUnitDelete = "orderfood:unit:delete"
	// PermissionUnitSort 表示调整单位排序所需的权限。
	PermissionUnitSort = "orderfood:unit:sort"
)

// catalogReferenceSource 表示一种业务引用计数来源。
type catalogReferenceSource struct {
	Table  string // 关联数据表名
	Column string // 指向基础数据主键的字段名
}

// catalogResource 描述一类基础数据的存储和审计信息。
type catalogResource struct {
	Key              string                   // 资源代码
	Table            string                   // 数据表名
	PublicIDPrefix   string                   // 对外ID前缀
	DisplayName      string                   // 中文名称
	ReferenceSources []catalogReferenceSource // 业务引用来源
}

var (
	categoryCatalogResource = catalogResource{
		Key:            "category",
		Table:          (orderfoodModel.ContentCategory{}).TableName(),
		PublicIDPrefix: "category",
		DisplayName:    "分类",
		ReferenceSources: []catalogReferenceSource{
			{Table: (orderfoodModel.UserDish{}).TableName(), Column: "category_id"},
			{Table: (orderfoodModel.StandardDishIndex{}).TableName(), Column: "category_id"},
			{Table: (orderfoodModel.OfficialDish{}).TableName(), Column: "category_id"},
		},
	}
	tagCatalogResource = catalogResource{
		Key:            "tag",
		Table:          (orderfoodModel.ContentTag{}).TableName(),
		PublicIDPrefix: "tag",
		DisplayName:    "标签",
		ReferenceSources: []catalogReferenceSource{
			{Table: (orderfoodModel.UserDishTag{}).TableName(), Column: "tag_id"},
			{Table: (orderfoodModel.OfficialDishTag{}).TableName(), Column: "tag_id"},
		},
	}
	unitCatalogResource = catalogResource{
		Key:            "unit",
		Table:          (orderfoodModel.ContentUnit{}).TableName(),
		PublicIDPrefix: "unit",
		DisplayName:    "单位",
		ReferenceSources: []catalogReferenceSource{
			{Table: (orderfoodModel.DishIngredient{}).TableName(), Column: "unit_id"},
			{Table: (orderfoodModel.OfficialDishIngredient{}).TableName(), Column: "unit_id"},
		},
	}
)

// catalogRecord 表示基础数据表的通用持久化字段。
type catalogRecord struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement"` // 内部主键
	PublicID  string     `gorm:"column:public_id"`                   // 对外公开ID
	Name      string     `gorm:"column:name"`                        // 名称
	SortOrder int        `gorm:"column:sort_order"`                  // 排序值
	Enabled   bool       `gorm:"column:enabled"`                     // 是否启用
	Version   int64      `gorm:"column:version"`                     // 数据版本
	CreatedAt time.Time  `gorm:"column:created_at"`                  // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"`                  // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"`                  // 删除时间
}

// CatalogService 提供分类、标签和单位的管理业务能力。
type CatalogService struct {
	DB            *gorm.DB            // 数据库连接
	Idempotency   *IdempotencyService // 幂等处理服务
	MutationAudit MutationAuditWriter // 变更审计写入器
	Now           func() time.Time    // 当前时间函数
}

// NewCatalogService 创建基础数据服务实例。
func NewCatalogService(
	db *gorm.DB,
	idempotency *IdempotencyService,
) *CatalogService {
	if idempotency == nil {
		idempotency = NewIdempotencyService(db)
	}
	return &CatalogService{
		DB:            db,
		Idempotency:   idempotency,
		MutationAudit: gormMutationAuditWriter{},
		Now:           time.Now,
	}
}

// database 返回服务使用的数据库连接。
func (service *CatalogService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回统一为UTC的当前时间。
func (service *CatalogService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// normalizeCatalogItemName 校验并规范化基础数据名称。
func normalizeCatalogItemName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 40 {
		return "", appErrors.AdminBadRequest.DefaultMsg()
	}
	return name, nil
}

// validateCatalogActor 校验管理操作上下文。
func validateCatalogActor(actor orderfoodRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// catalogResponse 将通用持久化记录转换为管理端响应。
func catalogResponse(
	row catalogRecord,
	referenceCount int64,
) orderfoodResponse.CatalogItem {
	return orderfoodResponse.CatalogItem{
		ID:             row.PublicID,
		Name:           row.Name,
		SortOrder:      row.SortOrder,
		Enabled:        row.Enabled,
		ReferenceCount: referenceCount,
		Version:        row.Version,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

// referenceCount 统计一条基础数据的全部业务引用。
func (service *CatalogService) referenceCount(
	ctx context.Context,
	db *gorm.DB,
	resource catalogResource,
	internalID uint,
) (int64, error) {
	var total int64
	for _, source := range resource.ReferenceSources {
		var count int64
		if err := db.WithContext(ctx).
			Table(source.Table).
			Where(source.Column+" = ?", internalID).
			Count(&count).Error; err != nil {
			return 0, appErrors.AdminInternal.Wrap(err, "count catalog references")
		}
		total += count
	}
	return total, nil
}

// loadRecord 查询一条未删除的基础数据，可选择加行锁。
func (service *CatalogService) loadRecord(
	ctx context.Context,
	db *gorm.DB,
	resource catalogResource,
	publicID string,
	lock bool,
) (catalogRecord, error) {
	var row catalogRecord
	statement := db.WithContext(ctx).Table(resource.Table)
	if lock {
		statement = statement.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := statement.
		Where("public_id = ? AND deleted_at IS NULL", strings.TrimSpace(publicID)).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return row, appErrors.AdminNotFound.DefaultMsg()
		}
		return row, appErrors.AdminInternal.Wrap(err, "load catalog item")
	}
	return row, nil
}

// ensureUniqueName 确保同类未删除资源中不存在重名记录。
func (service *CatalogService) ensureUniqueName(
	ctx context.Context,
	db *gorm.DB,
	resource catalogResource,
	name string,
	excludeID uint,
) error {
	var count int64
	statement := db.WithContext(ctx).
		Table(resource.Table).
		Where("LOWER(name) = LOWER(?) AND deleted_at IS NULL", name)
	if excludeID != 0 {
		statement = statement.Where("id <> ?", excludeID)
	}
	if err := statement.Count(&count).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "check catalog name")
	}
	if count > 0 {
		return appErrors.AdminAlreadyExists.DefaultMsg()
	}
	return nil
}

// writeMutationAudit 在同一事务内写入基础数据变更审计。
func (service *CatalogService) writeMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor orderfoodRequest.AdminActor,
	resource catalogResource,
	action string,
	targetID string,
	reason string,
	idempotencyKey string,
	before interface{},
	after interface{},
) error {
	if service == nil || service.MutationAudit == nil || tx == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	beforeJSON, err := safeAIAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := safeAIAuditJSON(after)
	if err != nil {
		return err
	}
	key := strings.TrimSpace(idempotencyKey)
	var reasonPointer *string
	if value := truncateRunes(strings.TrimSpace(reason), 200); value != "" {
		reasonPointer = &value
	}
	audit := orderfoodModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            resource.Key,
		TargetID:              targetID,
		Reason:                reasonPointer,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             truncateRunes(strings.TrimSpace(actor.RequestID), 96),
		IdempotencyKey:        &key,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

// listCatalogItems 分页查询指定类型的基础数据。
func (service *CatalogService) listCatalogItems(
	ctx context.Context,
	resource catalogResource,
	query orderfoodRequest.CatalogItemListQuery,
) (orderfoodResponse.Page[orderfoodResponse.CatalogItem], error) {
	query.ApplyDefaults()
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		utf8.RuneCountInString(strings.TrimSpace(query.Keyword)) > 40 ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	sortColumns := map[string]string{
		"sortOrder": "sort_order",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).
		Table(resource.Table).
		Where("deleted_at IS NULL")
	if value := strings.TrimSpace(query.Keyword); value != "" {
		statement = statement.Where("name LIKE ?", "%"+value+"%")
	}
	if query.Enabled != nil {
		statement = statement.Where("enabled = ?", *query.Enabled)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{}, appErrors.AdminInternal.Wrap(err, "count catalog items")
	}
	rows := make([]catalogRecord, 0)
	if err := statement.
		Order(column + " " + query.SortOrder).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{}, appErrors.AdminInternal.Wrap(err, "list catalog items")
	}
	list := make([]orderfoodResponse.CatalogItem, 0, len(rows))
	for _, row := range rows {
		count, err := service.referenceCount(ctx, db, resource, row.ID)
		if err != nil {
			return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{}, err
		}
		list = append(list, catalogResponse(row, count))
	}
	return orderfoodResponse.Page[orderfoodResponse.CatalogItem]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// createCatalogItem 创建一条基础数据并记录审计。
func (service *CatalogService) createCatalogItem(
	ctx context.Context,
	resource catalogResource,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemCreateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	if err := validateCatalogActor(actor); err != nil {
		return orderfoodResponse.CatalogItem{}, false, err
	}
	name, err := normalizeCatalogItemName(input.Name)
	if err != nil || input.Enabled == nil || input.SortOrder < 1 {
		return orderfoodResponse.CatalogItem{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	input.Name = name
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		resource.Key+"_create",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			if err := service.ensureUniqueName(ctx, tx, resource, name, 0); err != nil {
				return nil, err
			}
			now := service.now()
			row := catalogRecord{
				PublicID:  newPublicID(resource.PublicIDPrefix),
				Name:      name,
				SortOrder: input.SortOrder,
				Enabled:   *input.Enabled,
				Version:   1,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.WithContext(ctx).Table(resource.Table).Create(&row).Error; err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) ||
					strings.Contains(strings.ToLower(err.Error()), "duplicate") {
					return nil, appErrors.AdminAlreadyExists.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "create catalog item")
			}
			result := catalogResponse(row, 0)
			if err := service.writeMutationAudit(
				ctx, tx, actor, resource, resource.Key+"_create", row.PublicID,
				"新增"+resource.DisplayName, idempotencyKey, nil, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return orderfoodResponse.CatalogItem{}, false, err
	}
	var result orderfoodResponse.CatalogItem
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode catalog create result")
	}
	return result, replayed, nil
}

// updateCatalogItem 编辑一条基础数据并使用版本号防止覆盖并发修改。
func (service *CatalogService) updateCatalogItem(
	ctx context.Context,
	resource catalogResource,
	publicID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemUpdateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	if err := validateCatalogActor(actor); err != nil {
		return orderfoodResponse.CatalogItem{}, false, err
	}
	name, err := normalizeCatalogItemName(input.Name)
	if err != nil || input.Enabled == nil || input.SortOrder < 1 || input.ExpectedVersion < 1 {
		return orderfoodResponse.CatalogItem{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return orderfoodResponse.CatalogItem{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	input.Name = name
	payload := struct {
		ID    string                                  `json:"id"`    // 资源ID
		Input orderfoodRequest.CatalogItemUpdateInput `json:"input"` // 编辑参数
	}{ID: publicID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		resource.Key+"_update",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			row, err := service.loadRecord(ctx, tx, resource, publicID, true)
			if err != nil {
				return nil, err
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if err := service.ensureUniqueName(ctx, tx, resource, name, row.ID); err != nil {
				return nil, err
			}
			count, err := service.referenceCount(ctx, tx, resource, row.ID)
			if err != nil {
				return nil, err
			}
			before := catalogResponse(row, count)
			now := service.now()
			update := tx.WithContext(ctx).
				Table(resource.Table).
				Where("id = ? AND version = ? AND deleted_at IS NULL", row.ID, input.ExpectedVersion).
				Updates(map[string]interface{}{
					"name":       name,
					"sort_order": input.SortOrder,
					"enabled":    *input.Enabled,
					"version":    gorm.Expr("version + 1"),
					"updated_at": now,
				})
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "update catalog item")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			row.Name = name
			row.SortOrder = input.SortOrder
			row.Enabled = *input.Enabled
			row.Version++
			row.UpdatedAt = now
			if resource.Key == categoryCatalogResource.Key {
				if err := ensureCurrentSuggestionCatalogReady(tx); err != nil {
					return nil, err
				}
			}
			result := catalogResponse(row, count)
			if err := service.writeMutationAudit(
				ctx, tx, actor, resource, resource.Key+"_update", row.PublicID,
				"编辑"+resource.DisplayName, idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return orderfoodResponse.CatalogItem{}, false, err
	}
	var result orderfoodResponse.CatalogItem
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode catalog update result")
	}
	return result, replayed, nil
}

// deleteCatalogItem 删除未被业务数据引用的一条基础数据。
func (service *CatalogService) deleteCatalogItem(
	ctx context.Context,
	resource catalogResource,
	publicID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogDeleteInput,
) (orderfoodResponse.CatalogDeleteResult, bool, error) {
	if err := validateCatalogActor(actor); err != nil {
		return orderfoodResponse.CatalogDeleteResult{}, false, err
	}
	publicID = strings.TrimSpace(publicID)
	input.Reason = strings.TrimSpace(input.Reason)
	if publicID == "" || utf8.RuneCountInString(input.Reason) < 4 ||
		utf8.RuneCountInString(input.Reason) > 200 || input.ExpectedVersion < 1 {
		return orderfoodResponse.CatalogDeleteResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		ID    string                              `json:"id"`    // 资源ID
		Input orderfoodRequest.CatalogDeleteInput `json:"input"` // 删除参数
	}{ID: publicID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		resource.Key+"_delete",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			row, err := service.loadRecord(ctx, tx, resource, publicID, true)
			if err != nil {
				return nil, err
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			count, err := service.referenceCount(ctx, tx, resource, row.ID)
			if err != nil {
				return nil, err
			}
			if count > 0 {
				return nil, appErrors.AdminResourceInUse.DefaultMsg()
			}
			before := catalogResponse(row, count)
			deletion := tx.WithContext(ctx).
				Unscoped().
				Table(resource.Table).
				Where("id = ? AND version = ?", row.ID, input.ExpectedVersion).
				Delete(&catalogRecord{})
			if deletion.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(deletion.Error, "delete catalog item")
			}
			if deletion.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			result := orderfoodResponse.CatalogDeleteResult{Deleted: true}
			if err := service.writeMutationAudit(
				ctx, tx, actor, resource, resource.Key+"_delete", row.PublicID,
				input.Reason, idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return orderfoodResponse.CatalogDeleteResult{}, false, err
	}
	var result orderfoodResponse.CatalogDeleteResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode catalog delete result")
	}
	return result, replayed, nil
}

// updateCatalogSortOrder 在一个事务中批量调整基础数据排序。
func (service *CatalogService) updateCatalogSortOrder(
	ctx context.Context,
	resource catalogResource,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogSortOrderInput,
) (orderfoodResponse.CatalogSortOrderResult, bool, error) {
	if err := validateCatalogActor(actor); err != nil {
		return orderfoodResponse.CatalogSortOrderResult{}, false, err
	}
	if len(input.Items) < 1 || len(input.Items) > 100 {
		return orderfoodResponse.CatalogSortOrderResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	seen := make(map[string]struct{}, len(input.Items))
	for index := range input.Items {
		item := &input.Items[index]
		item.ID = strings.TrimSpace(item.ID)
		if item.ID == "" || item.SortOrder < 1 || item.ExpectedVersion < 1 {
			return orderfoodResponse.CatalogSortOrderResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
		}
		if _, exists := seen[item.ID]; exists {
			return orderfoodResponse.CatalogSortOrderResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
		}
		seen[item.ID] = struct{}{}
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		resource.Key+"_sort_update",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			before := make([]orderfoodResponse.CatalogItem, 0, len(input.Items))
			after := make([]orderfoodResponse.CatalogItem, 0, len(input.Items))
			for _, item := range input.Items {
				row, err := service.loadRecord(ctx, tx, resource, item.ID, true)
				if err != nil {
					return nil, err
				}
				if row.Version != item.ExpectedVersion {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				count, err := service.referenceCount(ctx, tx, resource, row.ID)
				if err != nil {
					return nil, err
				}
				before = append(before, catalogResponse(row, count))
				now := service.now()
				update := tx.WithContext(ctx).
					Table(resource.Table).
					Where("id = ? AND version = ? AND deleted_at IS NULL", row.ID, item.ExpectedVersion).
					Updates(map[string]interface{}{
						"sort_order": item.SortOrder,
						"version":    gorm.Expr("version + 1"),
						"updated_at": now,
					})
				if update.Error != nil {
					return nil, appErrors.AdminInternal.Wrap(update.Error, "update catalog sort order")
				}
				if update.RowsAffected != 1 {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				row.SortOrder = item.SortOrder
				row.Version++
				row.UpdatedAt = now
				after = append(after, catalogResponse(row, count))
			}
			result := orderfoodResponse.CatalogSortOrderResult{Updated: len(input.Items)}
			if err := service.writeMutationAudit(
				ctx, tx, actor, resource, resource.Key+"_sort_update", "batch",
				"批量调整"+resource.DisplayName+"排序", idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return orderfoodResponse.CatalogSortOrderResult{}, false, err
	}
	var result orderfoodResponse.CatalogSortOrderResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode catalog sort result")
	}
	return result, replayed, nil
}

// ListCategories 分页查询分类。
func (service *CatalogService) ListCategories(
	ctx context.Context,
	query orderfoodRequest.CatalogItemListQuery,
) (orderfoodResponse.Page[orderfoodResponse.CatalogItem], error) {
	return service.listCatalogItems(ctx, categoryCatalogResource, query)
}

// CreateCategory 新增分类。
func (service *CatalogService) CreateCategory(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemCreateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	return service.createCatalogItem(ctx, categoryCatalogResource, actor, idempotencyKey, input)
}

// UpdateCategory 编辑分类。
func (service *CatalogService) UpdateCategory(
	ctx context.Context,
	categoryID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemUpdateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	return service.updateCatalogItem(ctx, categoryCatalogResource, categoryID, actor, idempotencyKey, input)
}

// DeleteCategory 删除未被引用的分类。
func (service *CatalogService) DeleteCategory(
	ctx context.Context,
	categoryID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogDeleteInput,
) (orderfoodResponse.CatalogDeleteResult, bool, error) {
	return service.deleteCatalogItem(ctx, categoryCatalogResource, categoryID, actor, idempotencyKey, input)
}

// UpdateCategorySortOrder 批量调整分类排序。
func (service *CatalogService) UpdateCategorySortOrder(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogSortOrderInput,
) (orderfoodResponse.CatalogSortOrderResult, bool, error) {
	return service.updateCatalogSortOrder(ctx, categoryCatalogResource, actor, idempotencyKey, input)
}

// ListTags 分页查询标签。
func (service *CatalogService) ListTags(
	ctx context.Context,
	query orderfoodRequest.CatalogItemListQuery,
) (orderfoodResponse.Page[orderfoodResponse.CatalogItem], error) {
	return service.listCatalogItems(ctx, tagCatalogResource, query)
}

// CreateTag 新增标签。
func (service *CatalogService) CreateTag(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemCreateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	return service.createCatalogItem(ctx, tagCatalogResource, actor, idempotencyKey, input)
}

// UpdateTag 编辑标签。
func (service *CatalogService) UpdateTag(
	ctx context.Context,
	tagID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemUpdateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	return service.updateCatalogItem(ctx, tagCatalogResource, tagID, actor, idempotencyKey, input)
}

// DeleteTag 删除未被引用的标签。
func (service *CatalogService) DeleteTag(
	ctx context.Context,
	tagID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogDeleteInput,
) (orderfoodResponse.CatalogDeleteResult, bool, error) {
	return service.deleteCatalogItem(ctx, tagCatalogResource, tagID, actor, idempotencyKey, input)
}

// UpdateTagSortOrder 批量调整标签排序。
func (service *CatalogService) UpdateTagSortOrder(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogSortOrderInput,
) (orderfoodResponse.CatalogSortOrderResult, bool, error) {
	return service.updateCatalogSortOrder(ctx, tagCatalogResource, actor, idempotencyKey, input)
}

// ListUnits 分页查询单位。
func (service *CatalogService) ListUnits(
	ctx context.Context,
	query orderfoodRequest.CatalogItemListQuery,
) (orderfoodResponse.Page[orderfoodResponse.CatalogItem], error) {
	return service.listCatalogItems(ctx, unitCatalogResource, query)
}

// CreateUnit 新增单位。
func (service *CatalogService) CreateUnit(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemCreateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	return service.createCatalogItem(ctx, unitCatalogResource, actor, idempotencyKey, input)
}

// UpdateUnit 编辑单位。
func (service *CatalogService) UpdateUnit(
	ctx context.Context,
	unitID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogItemUpdateInput,
) (orderfoodResponse.CatalogItem, bool, error) {
	return service.updateCatalogItem(ctx, unitCatalogResource, unitID, actor, idempotencyKey, input)
}

// DeleteUnit 删除未被引用的单位。
func (service *CatalogService) DeleteUnit(
	ctx context.Context,
	unitID string,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogDeleteInput,
) (orderfoodResponse.CatalogDeleteResult, bool, error) {
	return service.deleteCatalogItem(ctx, unitCatalogResource, unitID, actor, idempotencyKey, input)
}

// UpdateUnitSortOrder 批量调整单位排序。
func (service *CatalogService) UpdateUnitSortOrder(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.CatalogSortOrderInput,
) (orderfoodResponse.CatalogSortOrderResult, bool, error) {
	return service.updateCatalogSortOrder(ctx, unitCatalogResource, actor, idempotencyKey, input)
}

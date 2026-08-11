package dish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"sort"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	dishResponse "github.com/dyjh/order-food-mini-app/server/model/dish/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	PermissionSuggestionCatalogRead   = "orderfood:suggestion-catalog:read"
	PermissionSuggestionCatalogUpdate = "orderfood:suggestion-catalog:update"

	suggestionValidationPolicySingletonKey = "platform"
	SuggestionValidationRetryCount         = 1
	SuggestionCatalogMinimumDishCount      = 6
)

// SuggestionCatalogService 提供推荐菜索引业务能力。
type SuggestionCatalogService struct {
	DB            *gorm.DB                          // 业务数据库
	Permission    *serviceCommon.PermissionService  // GVA权限服务
	Idempotency   *serviceCommon.IdempotencyService // 幂等执行服务
	MutationAudit serviceCommon.MutationAuditWriter // 配置变更审计写入器
	Now           func() time.Time                  // 可注入的当前时间函数
}

// NewSuggestionCatalogService 创建推荐菜索引服务实例。
func NewSuggestionCatalogService(
	db *gorm.DB,
	permission *serviceCommon.PermissionService,
	idempotency *serviceCommon.IdempotencyService,
) *SuggestionCatalogService {
	if permission == nil {
		permission = serviceCommon.NewPermissionService(db)
	}
	if idempotency == nil {
		idempotency = serviceCommon.NewIdempotencyService(db)
	}
	return &SuggestionCatalogService{
		DB:            db,
		Permission:    permission,
		Idempotency:   idempotency,
		MutationAudit: serviceCommon.GormMutationAuditWriter{},
		Now:           time.Now,
	}
}

func (service *SuggestionCatalogService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *SuggestionCatalogService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// EnsureDefaults 初始化并补齐默认值。
func (service *SuggestionCatalogService) EnsureDefaults(ctx context.Context) error {
	db := service.database()
	if db == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	row := dishModel.SuggestionValidationPolicy{
		SingletonKey:             suggestionValidationPolicySingletonKey,
		Version:                  1,
		CatalogValidationEnabled: false,
		AppliedByID:              0,
		AppliedByUsername:        "system",
		AppliedAt:                service.now(),
		Reason:                   "系统默认关闭生成结果索引校验",
	}
	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&row).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "seed default suggestion validation policy")
	}
	return nil
}

func latestSuggestionValidationPolicy(
	db *gorm.DB,
) (dishModel.SuggestionValidationPolicy, error) {
	var row dishModel.SuggestionValidationPolicy
	if err := db.Where(
		"singleton_key = ?",
		suggestionValidationPolicySingletonKey,
	).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return row, appErrors.AdminNotFound.DefaultMsg()
		}
		return row, appErrors.AdminInternal.Wrap(err, "load suggestion validation policy")
	}
	return row, nil
}

// CurrentSuggestionCatalogValidation 返回当前生效的索引校验开关。
// The retry count is intentionally fixed in code: when enabled the same paid
// generation may make one extra attempt before points are refunded.
func CurrentSuggestionCatalogValidation(ctx context.Context, db *gorm.DB) (bool, error) {
	if db == nil {
		db = global.GVA_DB
	}
	if db == nil {
		return false, appErrors.FrontInternal.DefaultMsg()
	}
	var row dishModel.SuggestionValidationPolicy
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", suggestionValidationPolicySingletonKey).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, appErrors.FrontInternal.Wrap(err, "load suggestion validation policy")
	}
	return row.CatalogValidationEnabled, nil
}

func suggestionCatalogAdministrator(
	id uint,
	username string,
	nickname *string,
) commonResponse.AdministratorSummary {
	return commonResponse.AdministratorSummary{
		ID:       strconv.FormatUint(uint64(id), 10),
		Username: username,
		Nickname: nickname,
	}
}

func suggestionPolicyConfigResponse(
	row dishModel.SuggestionValidationPolicy,
) dishResponse.SuggestionValidationPolicyConfig {
	return dishResponse.SuggestionValidationPolicyConfig{
		Version:                  row.Version,
		CatalogValidationEnabled: row.CatalogValidationEnabled,
		RetryCount:               SuggestionValidationRetryCount,
		RefundOnFailure:          true,
		UpdatedBy: suggestionCatalogAdministrator(
			row.AppliedByID,
			row.AppliedByUsername,
			row.AppliedByNickname,
		),
		UpdatedAt: row.AppliedAt,
		Reason:    row.Reason,
	}
}

func addCatalogNames(
	owners map[string]string,
	collisions map[string]struct{},
	ownerID string,
	name string,
	aliases []string,
) {
	for _, value := range append([]string{name}, aliases...) {
		normalized := normalizeCatalogName(value)
		if normalized == "" {
			continue
		}
		if existingOwner, exists := owners[normalized]; exists && existingOwner != ownerID {
			collisions[normalized] = struct{}{}
			continue
		}
		owners[normalized] = ownerID
	}
}

// validateCatalogNameUniqueness 确保同类标准目录中的名称和别名不会指向不同记录。
func validateCatalogNameUniqueness(
	db *gorm.DB,
	kind string,
	currentID string,
	name string,
	aliases []string,
) error {
	candidates := make(map[string]struct{}, len(aliases)+1)
	for _, value := range append([]string{name}, aliases...) {
		if normalized := normalizeCatalogName(value); normalized != "" {
			candidates[normalized] = struct{}{}
		}
	}
	checkRow := func(rowID string, rowName string, rowAliases datatypes.JSON) error {
		if rowID == currentID {
			return nil
		}
		for _, value := range append([]string{rowName}, decodeCatalogAliases(rowAliases)...) {
			if _, exists := candidates[normalizeCatalogName(value)]; exists {
				return appErrors.AdminAlreadyExists.DefaultMsg()
			}
		}
		return nil
	}
	switch kind {
	case "ingredient":
		var rows []dishModel.StandardIngredient
		if err := db.Select("id", "name", "aliases_json").Find(&rows).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "load standard ingredients for name uniqueness")
		}
		for _, row := range rows {
			if err := checkRow(row.ID, row.Name, row.AliasesJSON); err != nil {
				return err
			}
		}
	case "dish":
		var rows []dishModel.StandardDishIndex
		if err := db.Select("id", "name", "aliases_json").Find(&rows).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "load standard dishes for name uniqueness")
		}
		for _, row := range rows {
			if err := checkRow(row.ID, row.Name, row.AliasesJSON); err != nil {
				return err
			}
		}
	default:
		return appErrors.AdminInternal.DefaultMsg()
	}
	return nil
}

// suggestionCatalogReadiness 计算开启标准菜品索引校验前必须满足的完整就绪条件。
func suggestionCatalogReadiness(
	db *gorm.DB,
) (dishResponse.SuggestionCatalogReadiness, error) {
	readiness := dishResponse.SuggestionCatalogReadiness{
		RequiredMinimumDishCount: SuggestionCatalogMinimumDishCount,
		Blockers:                 []string{},
	}
	var ingredients []dishModel.StandardIngredient
	if err := db.Where("enabled = ?", true).
		Order("id asc").
		Find(&ingredients).Error; err != nil {
		return readiness, appErrors.AdminInternal.Wrap(err, "load enabled standard ingredients for readiness")
	}
	enabledIngredientIDs := make(map[string]struct{}, len(ingredients))
	ingredientOwners := make(map[string]string)
	ingredientCollisions := make(map[string]struct{})
	for _, ingredient := range ingredients {
		enabledIngredientIDs[ingredient.ID] = struct{}{}
		addCatalogNames(
			ingredientOwners,
			ingredientCollisions,
			ingredient.ID,
			ingredient.Name,
			decodeCatalogAliases(ingredient.AliasesJSON),
		)
	}

	var dishes []dishModel.StandardDishIndex
	if err := db.Where("enabled = ?", true).
		Order("id asc").
		Find(&dishes).Error; err != nil {
		return readiness, appErrors.AdminInternal.Wrap(err, "load enabled standard dishes for readiness")
	}
	readiness.EnabledDishCount = int64(len(dishes))
	enabledCategoryIDs := make(map[uint]struct{})
	var categories []dishModel.ContentCategory
	if err := db.Where("enabled = ?", true).Find(&categories).Error; err != nil {
		return readiness, appErrors.AdminInternal.Wrap(err, "load enabled categories for readiness")
	}
	for _, category := range categories {
		enabledCategoryIDs[category.ID] = struct{}{}
	}

	dishOwners := make(map[string]string)
	dishCollisions := make(map[string]struct{})
	dishIDs := make([]string, 0, len(dishes))
	relationCountByDish := make(map[string]int)
	for _, dish := range dishes {
		dishIDs = append(dishIDs, dish.ID)
		addCatalogNames(
			dishOwners,
			dishCollisions,
			dish.ID,
			dish.Name,
			decodeCatalogAliases(dish.AliasesJSON),
		)
		if _, exists := enabledCategoryIDs[dish.CategoryID]; !exists {
			readiness.InvalidReferenceCount++
		}
	}
	if len(dishIDs) > 0 {
		var relations []dishModel.StandardDishIngredient
		if err := db.Where("dish_id IN ?", dishIDs).
			Order("dish_id asc, sort_order asc").
			Find(&relations).Error; err != nil {
			return readiness, appErrors.AdminInternal.Wrap(err, "load standard dish ingredient references for readiness")
		}
		for _, relation := range relations {
			relationCountByDish[relation.DishID]++
			if _, exists := enabledIngredientIDs[relation.IngredientID]; !exists {
				readiness.InvalidReferenceCount++
			}
		}
		for _, dishID := range dishIDs {
			if relationCountByDish[dishID] == 0 {
				readiness.InvalidReferenceCount++
			}
		}
	}

	readiness.NameCollisionCount = int64(len(ingredientCollisions) + len(dishCollisions))
	if readiness.EnabledDishCount < readiness.RequiredMinimumDishCount {
		readiness.Blockers = append(
			readiness.Blockers,
			fmt.Sprintf(
				"启用菜品至少需要 %d 道，当前 %d 道",
				readiness.RequiredMinimumDishCount,
				readiness.EnabledDishCount,
			),
		)
	}
	if readiness.InvalidReferenceCount > 0 {
		readiness.Blockers = append(
			readiness.Blockers,
			fmt.Sprintf(
				"存在 %d 处未启用或不存在的分类、食材引用",
				readiness.InvalidReferenceCount,
			),
		)
	}
	if readiness.NameCollisionCount > 0 {
		readiness.Blockers = append(
			readiness.Blockers,
			fmt.Sprintf(
				"存在 %d 个标准名称或别名冲突",
				readiness.NameCollisionCount,
			),
		)
	}
	readiness.Ready = len(readiness.Blockers) == 0
	return readiness, nil
}

// ensureCurrentSuggestionCatalogReady 防止已开启的当前校验目录被后续管理操作破坏。
func ensureCurrentSuggestionCatalogReady(db *gorm.DB) error {
	var policy dishModel.SuggestionValidationPolicy
	if err := db.Where(
		"singleton_key = ?",
		suggestionValidationPolicySingletonKey,
	).First(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return appErrors.AdminInternal.Wrap(err, "load current suggestion validation policy")
	}
	if !policy.CatalogValidationEnabled {
		return nil
	}
	readiness, err := suggestionCatalogReadiness(db)
	if err != nil {
		return err
	}
	if !readiness.Ready {
		return appErrors.AdminInvalidConfig.DefaultMsg()
	}
	return nil
}

// Workspace 获取当前推荐菜索引配置。
func (service *SuggestionCatalogService) Workspace(
	ctx context.Context,
) (dishResponse.SuggestionCatalogWorkspace, error) {
	if err := service.EnsureDefaults(ctx); err != nil {
		return dishResponse.SuggestionCatalogWorkspace{}, err
	}
	db := service.database().WithContext(ctx)
	current, err := latestSuggestionValidationPolicy(db)
	if err != nil {
		return dishResponse.SuggestionCatalogWorkspace{}, err
	}
	result := dishResponse.SuggestionCatalogWorkspace{
		Policy:     suggestionPolicyConfigResponse(current),
		Categories: []commonResponse.CatalogReference{},
	}
	var categories []dishModel.ContentCategory
	if err := db.Where("enabled = ?", true).Order("name asc").Find(&categories).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "load standard dish index categories")
	}
	for _, category := range categories {
		result.Categories = append(result.Categories, commonResponse.CatalogReference{
			ID: category.PublicID, Name: category.Name, Enabled: category.Enabled,
		})
	}
	if err := db.Model(&dishModel.StandardDishIndex{}).Count(&result.DishCount).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "count standard dish indexes")
	}
	if err := db.Model(&dishModel.StandardDishIndex{}).
		Where("enabled = ?", true).
		Count(&result.EnabledDishCount).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "count enabled standard dish indexes")
	}
	if err := db.Model(&dishModel.StandardIngredient{}).Count(&result.IngredientCount).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "count standard ingredients")
	}
	if err := db.Model(&dishModel.StandardIngredient{}).
		Where("enabled = ?", true).
		Count(&result.EnabledIngredientCount).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "count enabled standard ingredients")
	}
	result.Readiness, err = suggestionCatalogReadiness(db)
	if err != nil {
		return result, err
	}
	return result, nil
}

func validateSuggestionCatalogActor(actor commonRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

func (service *SuggestionCatalogService) writeMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor commonRequest.AdminActor,
	action string,
	targetType string,
	targetID string,
	reason string,
	idempotencyKey string,
	before interface{},
	after interface{},
) error {
	if service == nil || service.MutationAudit == nil || tx == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	requestID := serviceCommon.TruncateRunes(strings.TrimSpace(actor.RequestID), 96)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if requestID == "" || idempotencyKey == "" || targetType == "" || targetID == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	beforeJSON, err := serviceCommon.SafeAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := serviceCommon.SafeAuditJSON(after)
	if err != nil {
		return err
	}
	var reasonPointer *string
	if value := serviceCommon.TruncateRunes(strings.TrimSpace(reason), 200); value != "" {
		reasonPointer = &value
	}
	audit := auditModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            targetType,
		TargetID:              targetID,
		Reason:                reasonPointer,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             requestID,
		IdempotencyKey:        &idempotencyKey,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := serviceCommon.TruncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

// UpdatePolicy 保存并立即应用生成结果索引校验策略。
func (service *SuggestionCatalogService) UpdatePolicy(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.SuggestionValidationPolicyUpdateInput,
) (dishResponse.SuggestionValidationPolicyConfig, bool, error) {
	if err := validateSuggestionCatalogActor(actor); err != nil {
		return dishResponse.SuggestionValidationPolicyConfig{}, false, err
	}
	reason := strings.TrimSpace(input.Reason)
	if input.ExpectedVersion < 1 || len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return dishResponse.SuggestionValidationPolicyConfig{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := service.EnsureDefaults(ctx); err != nil {
		return dishResponse.SuggestionValidationPolicyConfig{}, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"suggestion_validation_policy_update",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			current, err := latestSuggestionValidationPolicy(
				tx,
			)
			if err != nil {
				return nil, err
			}
			if current.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if input.CatalogValidationEnabled {
				readiness, readinessErr := suggestionCatalogReadiness(tx)
				if readinessErr != nil {
					return nil, readinessErr
				}
				if !readiness.Ready {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
			}
			next := dishModel.SuggestionValidationPolicy{
				SingletonKey:             suggestionValidationPolicySingletonKey,
				Version:                  current.Version + 1,
				CatalogValidationEnabled: input.CatalogValidationEnabled,
				AppliedByID:              actor.AdministratorID,
				AppliedByUsername:        actor.Username,
				AppliedByNickname:        actor.Nickname,
				AppliedAt:                service.now(),
				Reason:                   reason,
			}
			update := tx.Model(&dishModel.SuggestionValidationPolicy{}).
				Where("singleton_key = ? AND version = ?", suggestionValidationPolicySingletonKey, input.ExpectedVersion).
				Select("*").Omit("singleton_key").Updates(&next)
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "save suggestion validation policy")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before := suggestionPolicyConfigResponse(current)
			after := suggestionPolicyConfigResponse(next)
			if err := service.writeMutationAudit(
				ctx, tx, actor,
				"update_suggestion_validation_policy",
				"suggestion_validation_policy",
				suggestionValidationPolicySingletonKey,
				reason, idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[dishResponse.SuggestionValidationPolicyConfig](raw, replayed, err)
}

func normalizeCatalogName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Join(strings.Fields(value), "")
}

func normalizeAliases(name string, aliases []string, maxLength int) ([]string, error) {
	seen := map[string]struct{}{normalizeCatalogName(name): {}}
	result := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		normalized := normalizeCatalogName(alias)
		if normalized == "" || len([]rune(alias)) > maxLength {
			return nil, appErrors.AdminBadRequest.DefaultMsg()
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, alias)
	}
	sort.Slice(result, func(left, right int) bool {
		return normalizeCatalogName(result[left]) < normalizeCatalogName(result[right])
	})
	return result, nil
}

func encodeCatalogAliases(aliases []string) (datatypes.JSON, error) {
	encoded, err := json.Marshal(aliases)
	if err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "encode catalog aliases")
	}
	return datatypes.JSON(encoded), nil
}

func decodeCatalogAliases(raw datatypes.JSON) []string {
	var aliases []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &aliases)
	}
	if aliases == nil {
		return []string{}
	}
	return aliases
}

func standardIngredientResponse(
	row dishModel.StandardIngredient,
) dishResponse.StandardIngredient {
	return dishResponse.StandardIngredient{
		ID:      row.ID,
		Name:    row.Name,
		Aliases: decodeCatalogAliases(row.AliasesJSON),
		Enabled: row.Enabled,
		Version: row.Version,
		UpdatedBy: suggestionCatalogAdministrator(
			row.UpdatedByID,
			row.UpdatedByUsername,
			row.UpdatedByNickname,
		),
		UpdatedAt: row.UpdatedAt,
	}
}

func standardDishResponse(row dishModel.StandardDishIndex) dishResponse.StandardDish {
	ingredients := make([]dishResponse.StandardDishIngredient, 0, len(row.Ingredients))
	sort.Slice(row.Ingredients, func(left, right int) bool {
		return row.Ingredients[left].SortOrder < row.Ingredients[right].SortOrder
	})
	for _, ingredient := range row.Ingredients {
		ingredients = append(ingredients, dishResponse.StandardDishIngredient{
			IngredientID: ingredient.IngredientID,
			Name:         ingredient.Ingredient.Name,
			Required:     ingredient.Required,
			SortOrder:    ingredient.SortOrder,
		})
	}
	return dishResponse.StandardDish{
		ID:          row.ID,
		Name:        row.Name,
		Aliases:     decodeCatalogAliases(row.AliasesJSON),
		Cuisine:     row.Cuisine,
		CategoryID:  row.Category.PublicID,
		Category:    row.Category.Name,
		SourceName:  row.SourceName,
		SourceURL:   row.SourceURL,
		Enabled:     row.Enabled,
		Ingredients: ingredients,
		Version:     row.Version,
		UpdatedBy: suggestionCatalogAdministrator(
			row.UpdatedByID,
			row.UpdatedByUsername,
			row.UpdatedByNickname,
		),
		UpdatedAt: row.UpdatedAt,
	}
}

func applyStandardCatalogFilters(
	db *gorm.DB,
	query dishRequest.StandardCatalogListQuery,
) *gorm.DB {
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR aliases_json LIKE ?", like, like)
	}
	if query.Enabled != nil {
		db = db.Where("enabled = ?", *query.Enabled)
	}
	return db
}

// ListIngredients 分页查询食材列表。
func (service *SuggestionCatalogService) ListIngredients(
	ctx context.Context,
	query dishRequest.StandardCatalogListQuery,
) (commonResponse.Page[dishResponse.StandardIngredient], error) {
	query.ApplyDefaults()
	order := strings.ToLower(query.SortOrder)
	if order != "asc" && order != "desc" {
		return commonResponse.Page[dishResponse.StandardIngredient]{},
			appErrors.AdminBadRequest.DefaultMsg()
	}
	db := applyStandardCatalogFilters(
		service.database().WithContext(ctx).Model(&dishModel.StandardIngredient{}),
		query,
	)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return commonResponse.Page[dishResponse.StandardIngredient]{},
			appErrors.AdminInternal.Wrap(err, "count standard ingredients")
	}
	var rows []dishModel.StandardIngredient
	if err := db.Order("name " + order).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[dishResponse.StandardIngredient]{},
			appErrors.AdminInternal.Wrap(err, "list standard ingredients")
	}
	list := make([]dishResponse.StandardIngredient, 0, len(rows))
	for _, row := range rows {
		list = append(list, standardIngredientResponse(row))
	}
	return commonResponse.Page[dishResponse.StandardIngredient]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

func (service *SuggestionCatalogService) saveIngredient(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	id string,
	input dishRequest.StandardIngredientInput,
) (dishResponse.StandardIngredient, bool, error) {
	if err := validateSuggestionCatalogActor(actor); err != nil {
		return dishResponse.StandardIngredient{}, false, err
	}
	name := strings.TrimSpace(input.Name)
	reason := strings.TrimSpace(input.Reason)
	if name == "" || input.Enabled == nil || len([]rune(reason)) < 4 {
		return dishResponse.StandardIngredient{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	aliases, err := normalizeAliases(name, input.Aliases, 80)
	if err != nil {
		return dishResponse.StandardIngredient{}, false, err
	}
	aliasesJSON, err := encodeCatalogAliases(aliases)
	if err != nil {
		return dishResponse.StandardIngredient{}, false, err
	}
	creating := strings.TrimSpace(id) == ""
	endpoint := "standard_ingredient_update"
	if creating {
		id = serviceCommon.NewPublicID("ingredient")
		endpoint = "standard_ingredient_create"
	}
	payload := struct {
		ID    string
		Input dishRequest.StandardIngredientInput
	}{ID: id, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, endpoint, idempotencyKey, payload,
		func(tx *gorm.DB) (interface{}, error) {
			now := service.now()
			var before interface{}
			var row dishModel.StandardIngredient
			if err := validateCatalogNameUniqueness(
				tx, "ingredient", id, name, aliases,
			); err != nil {
				return nil, err
			}
			if creating {
				if input.ExpectedVersion != nil {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				row = dishModel.StandardIngredient{
					ID: id, Name: name, AliasesJSON: aliasesJSON,
					Enabled: *input.Enabled, Version: 1,
					UpdatedByID: actor.AdministratorID, UpdatedByUsername: actor.Username,
					UpdatedByNickname: actor.Nickname, LastChangeReason: reason,
					CreatedAt: now, UpdatedAt: now,
				}
				if err := tx.Create(&row).Error; err != nil {
					return nil, appErrors.AdminStateConflict.Wrap(err, "create standard ingredient")
				}
			} else {
				if err := tx.
					First(&row, "id = ?", id).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, appErrors.AdminNotFound.DefaultMsg()
					}
					return nil, appErrors.AdminInternal.Wrap(err, "load standard ingredient")
				}
				if input.ExpectedVersion == nil || *input.ExpectedVersion != row.Version {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				before = standardIngredientResponse(row)
				row.Name = name
				row.AliasesJSON = aliasesJSON
				row.Enabled = *input.Enabled
				row.Version++
				row.UpdatedByID = actor.AdministratorID
				row.UpdatedByUsername = actor.Username
				row.UpdatedByNickname = actor.Nickname
				row.LastChangeReason = reason
				row.UpdatedAt = now
				update := tx.Model(&dishModel.StandardIngredient{}).
					Where("id = ? AND version = ?", row.ID, *input.ExpectedVersion).
					Select("*").Omit("id").Updates(&row)
				if update.Error != nil {
					return nil, appErrors.AdminInternal.Wrap(update.Error, "update standard ingredient")
				}
				if update.RowsAffected != 1 {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
			}
			if err := ensureCurrentSuggestionCatalogReady(tx); err != nil {
				return nil, err
			}
			after := standardIngredientResponse(row)
			action := "update_standard_ingredient"
			if creating {
				action = "create_standard_ingredient"
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, action, "standard_ingredient", id,
				reason, idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[dishResponse.StandardIngredient](raw, replayed, err)
}

// CreateIngredient 创建食材。
func (service *SuggestionCatalogService) CreateIngredient(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.StandardIngredientInput,
) (dishResponse.StandardIngredient, bool, error) {
	return service.saveIngredient(ctx, actor, idempotencyKey, "", input)
}

// UpdateIngredient 更新食材。
func (service *SuggestionCatalogService) UpdateIngredient(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	id string,
	input dishRequest.StandardIngredientInput,
) (dishResponse.StandardIngredient, bool, error) {
	if strings.TrimSpace(id) == "" {
		return dishResponse.StandardIngredient{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	return service.saveIngredient(ctx, actor, idempotencyKey, id, input)
}

// ListDishes 分页查询菜品列表。
func (service *SuggestionCatalogService) ListDishes(
	ctx context.Context,
	query dishRequest.StandardCatalogListQuery,
) (commonResponse.Page[dishResponse.StandardDish], error) {
	query.ApplyDefaults()
	order := strings.ToLower(query.SortOrder)
	if order != "asc" && order != "desc" {
		return commonResponse.Page[dishResponse.StandardDish]{},
			appErrors.AdminBadRequest.DefaultMsg()
	}
	db := applyStandardCatalogFilters(
		service.database().WithContext(ctx).Model(&dishModel.StandardDishIndex{}),
		query,
	)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return commonResponse.Page[dishResponse.StandardDish]{},
			appErrors.AdminInternal.Wrap(err, "count standard dish indexes")
	}
	var rows []dishModel.StandardDishIndex
	if err := db.Preload("Category").
		Preload("Ingredients.Ingredient").
		Order("name " + order).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[dishResponse.StandardDish]{},
			appErrors.AdminInternal.Wrap(err, "list standard dish indexes")
	}
	list := make([]dishResponse.StandardDish, 0, len(rows))
	for _, row := range rows {
		list = append(list, standardDishResponse(row))
	}
	return commonResponse.Page[dishResponse.StandardDish]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

func (service *SuggestionCatalogService) saveDish(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	id string,
	input dishRequest.StandardDishInput,
) (dishResponse.StandardDish, bool, error) {
	if err := validateSuggestionCatalogActor(actor); err != nil {
		return dishResponse.StandardDish{}, false, err
	}
	name := strings.TrimSpace(input.Name)
	cuisine := strings.TrimSpace(input.Cuisine)
	sourceName := strings.TrimSpace(input.SourceName)
	reason := strings.TrimSpace(input.Reason)
	if name == "" || cuisine == "" || sourceName == "" || input.Enabled == nil ||
		len([]rune(reason)) < 4 || len(input.Ingredients) == 0 {
		return dishResponse.StandardDish{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	aliases, err := normalizeAliases(name, input.Aliases, 120)
	if err != nil {
		return dishResponse.StandardDish{}, false, err
	}
	aliasesJSON, err := encodeCatalogAliases(aliases)
	if err != nil {
		return dishResponse.StandardDish{}, false, err
	}
	creating := strings.TrimSpace(id) == ""
	endpoint := "standard_dish_index_update"
	if creating {
		id = serviceCommon.NewPublicID("standard_dish")
		endpoint = "standard_dish_index_create"
	}
	payload := struct {
		ID    string
		Input dishRequest.StandardDishInput
	}{ID: id, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, endpoint, idempotencyKey, payload,
		func(tx *gorm.DB) (interface{}, error) {
			var category dishModel.ContentCategory
			if err := tx.First(&category, "public_id = ? AND enabled = ?", input.CategoryID, true).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load standard dish category")
			}
			relations := make([]dishModel.StandardDishIngredient, 0, len(input.Ingredients))
			seen := make(map[string]struct{}, len(input.Ingredients))
			for _, item := range input.Ingredients {
				if item.Required == nil || item.SortOrder < 1 {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				if _, exists := seen[item.IngredientID]; exists {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				seen[item.IngredientID] = struct{}{}
				var count int64
				if err := tx.Model(&dishModel.StandardIngredient{}).
					Where("id = ? AND enabled = ?", item.IngredientID, true).
					Count(&count).Error; err != nil {
					return nil, appErrors.AdminInternal.Wrap(err, "validate standard dish ingredient")
				}
				if count != 1 {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				relations = append(relations, dishModel.StandardDishIngredient{
					DishID: id, IngredientID: item.IngredientID,
					Required: *item.Required, SortOrder: item.SortOrder,
					CreatedAt: service.now(),
				})
			}
			if err := validateCatalogNameUniqueness(
				tx, "dish", id, name, aliases,
			); err != nil {
				return nil, err
			}
			now := service.now()
			var before interface{}
			var row dishModel.StandardDishIndex
			if creating {
				if input.ExpectedVersion != nil {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				row = dishModel.StandardDishIndex{
					ID: id, Name: name, AliasesJSON: aliasesJSON,
					Cuisine: cuisine, CategoryID: category.ID,
					SourceName: sourceName, SourceURL: input.SourceURL,
					Enabled: *input.Enabled, Version: 1,
					UpdatedByID: actor.AdministratorID, UpdatedByUsername: actor.Username,
					UpdatedByNickname: actor.Nickname, LastChangeReason: reason,
					CreatedAt: now, UpdatedAt: now,
				}
				if err := tx.Create(&row).Error; err != nil {
					return nil, appErrors.AdminStateConflict.Wrap(err, "create standard dish index")
				}
			} else {
				if err := tx.Preload("Category").Preload("Ingredients.Ingredient").
					First(&row, "id = ?", id).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, appErrors.AdminNotFound.DefaultMsg()
					}
					return nil, appErrors.AdminInternal.Wrap(err, "load standard dish index")
				}
				if input.ExpectedVersion == nil || *input.ExpectedVersion != row.Version {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				before = standardDishResponse(row)
				row.Name = name
				row.AliasesJSON = aliasesJSON
				row.Cuisine = cuisine
				row.CategoryID = category.ID
				row.SourceName = sourceName
				row.SourceURL = input.SourceURL
				row.Enabled = *input.Enabled
				row.Version++
				row.UpdatedByID = actor.AdministratorID
				row.UpdatedByUsername = actor.Username
				row.UpdatedByNickname = actor.Nickname
				row.LastChangeReason = reason
				row.UpdatedAt = now
				update := tx.Model(&dishModel.StandardDishIndex{}).
					Where("id = ? AND version = ?", row.ID, *input.ExpectedVersion).
					Select("*").Omit("id").Updates(&row)
				if update.Error != nil {
					return nil, appErrors.AdminInternal.Wrap(update.Error, "update standard dish index")
				}
				if update.RowsAffected != 1 {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				if err := tx.Where("dish_id = ?", id).
					Delete(&dishModel.StandardDishIngredient{}).Error; err != nil {
					return nil, appErrors.AdminInternal.Wrap(err, "replace standard dish ingredients")
				}
			}
			if err := tx.Create(&relations).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "save standard dish ingredients")
			}
			row.Category = category
			if err := tx.Preload("Ingredients.Ingredient").First(&row, "id = ?", id).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "reload standard dish index")
			}
			if err := ensureCurrentSuggestionCatalogReady(tx); err != nil {
				return nil, err
			}
			after := standardDishResponse(row)
			action := "update_standard_dish_index"
			if creating {
				action = "create_standard_dish_index"
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, action, "standard_dish_index", id,
				reason, idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[dishResponse.StandardDish](raw, replayed, err)
}

// CreateDish 创建菜品。
func (service *SuggestionCatalogService) CreateDish(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.StandardDishInput,
) (dishResponse.StandardDish, bool, error) {
	return service.saveDish(ctx, actor, idempotencyKey, "", input)
}

// UpdateDish 更新菜品。
func (service *SuggestionCatalogService) UpdateDish(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	id string,
	input dishRequest.StandardDishInput,
) (dishResponse.StandardDish, bool, error) {
	if strings.TrimSpace(id) == "" {
		return dishResponse.StandardDish{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	return service.saveDish(ctx, actor, idempotencyKey, id, input)
}

package orderfood

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	contentReq "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	contentRes "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

const (
	governancePreviewTTL           = 5 * time.Minute
	governanceCopyChainConfirmText = "确认处理复制链"
)

const (
	PermissionUserDishRead          = "orderfood:user-dish:read"
	PermissionUserDishPrivateRead   = "orderfood:user-dish:private-read"
	PermissionUserRecipeRead        = "orderfood:user-recipe:read"
	PermissionUserRecipePrivateRead = "orderfood:user-recipe:private-read"
	PermissionGovernanceExecute     = "orderfood:governance:execute"
)

// MutationAuditWriter 定义变更审计写入器所需的业务能力。
type MutationAuditWriter interface {
	WriteMutationAudit(context.Context, *gorm.DB, *contentModel.AdminAuditLog) error
}

// GovernanceNotificationWriter 定义违规处理通知写入器所需的业务能力。
type GovernanceNotificationWriter interface {
	WriteGovernanceNotification(context.Context, *gorm.DB, *contentModel.UserNotification) (string, error)
}

// ContentServiceOption 定义内容服务配置选项。
type ContentServiceOption func(*ContentService)

// WithMutationAuditWriter 配置内容变更审计写入器。
func WithMutationAuditWriter(writer MutationAuditWriter) ContentServiceOption {
	return func(service *ContentService) {
		service.mutationAudit = writer
	}
}

// WithGovernanceNotificationWriter 配置违规处理通知写入器。
func WithGovernanceNotificationWriter(writer GovernanceNotificationWriter) ContentServiceOption {
	return func(service *ContentService) {
		service.notification = writer
	}
}

type gormMutationAuditWriter struct{}

// WriteMutationAudit 写入变更审计。
func (gormMutationAuditWriter) WriteMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	audit *contentModel.AdminAuditLog,
) error {
	if tx == nil || audit == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	if audit.PublicID == "" {
		audit.PublicID = newPublicID("audit")
	}
	if err := tx.WithContext(ctx).Create(audit).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "create mutation audit log")
	}
	return nil
}

type gormGovernanceNotificationWriter struct{}

// WriteGovernanceNotification 写入违规处理通知。
func (gormGovernanceNotificationWriter) WriteGovernanceNotification(
	ctx context.Context,
	tx *gorm.DB,
	notification *contentModel.UserNotification,
) (string, error) {
	if tx == nil || notification == nil {
		return "", appErrors.AdminInternal.DefaultMsg()
	}
	if notification.ID == "" {
		notification.ID = newPublicID("notice")
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now().UTC()
	}
	if err := tx.WithContext(ctx).Create(notification).Error; err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "create governance notification")
	}
	return notification.ID, nil
}

// ContentService 提供内容管理业务能力。
type ContentService struct {
	db            *gorm.DB
	accessAudit   *AccessAuditService
	idempotency   *IdempotencyService
	permission    *PermissionService
	mutationAudit MutationAuditWriter
	notification  GovernanceNotificationWriter
	previewSecret []byte
	now           func() time.Time
}

// NewContentService 创建内容服务实例。
func NewContentService(
	db *gorm.DB,
	accessAudit *AccessAuditService,
	idempotency *IdempotencyService,
	permission *PermissionService,
	previewSecret []byte,
	options ...ContentServiceOption,
) *ContentService {
	service := &ContentService{
		db:            db,
		accessAudit:   accessAudit,
		idempotency:   idempotency,
		permission:    permission,
		mutationAudit: gormMutationAuditWriter{},
		notification:  gormGovernanceNotificationWriter{},
		previewSecret: append([]byte(nil), previewSecret...),
		now:           time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

// BindDatabase 原地绑定内容服务及其依赖服务的数据库。
// 包级单例早于数据库初始化创建，ContentApi 会一直持有同一个服务指针。
func (s *ContentService) BindDatabase(db *gorm.DB) error {
	if s == nil || db == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	s.db = db
	// 内容读取会继续调用权限、访问审计和幂等服务，必须同步绑定同一数据库。
	if s.accessAudit != nil {
		s.accessAudit.DB = db
	}
	if s.idempotency != nil {
		s.idempotency.DB = db
	}
	if s.permission != nil {
		s.permission.DB = db
	}
	return nil
}

// ListUserDishes 分页查询用户菜品列表。
func (s *ContentService) ListUserDishes(
	ctx context.Context,
	search contentReq.UserDishSearch,
	actor contentReq.AdminActor,
) (contentRes.Page[contentRes.UserDishAdminSummary], error) {
	var result contentRes.Page[contentRes.UserDishAdminSummary]
	if s.db == nil || s.permission == nil {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionUserDishRead); err != nil {
		return result, err
	}
	includePrivateCover, err := s.permission.HasPermission(ctx, actor.AuthorityID, PermissionUserDishPrivateRead)
	if err != nil {
		return result, err
	}
	if !validTimeRange(search.CreatedFrom, search.CreatedTo) ||
		!validTimeRange(search.UpdatedFrom, search.UpdatedTo) ||
		!validTimeRange(search.DeletedFrom, search.DeletedTo) ||
		!validPagination(search.Page, search.PageSize) ||
		utf8.RuneCountInString(search.Keyword) > 60 ||
		utf8.RuneCountInString(search.UserKeyword) > 60 ||
		(search.SourceType != "" && !validSourceType(search.SourceType)) ||
		(search.MediaReviewStatus != "" && !validMediaReviewStatus(search.MediaReviewStatus)) {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	page, pageSize := normalizePage(search.Page, search.PageSize)
	query := s.db.WithContext(ctx).Model(&contentModel.UserDish{})
	if search.DeletionStatus == "deleted" {
		query = query.Unscoped().Where("of_dishes.deleted_at IS NOT NULL")
	} else if search.DeletionStatus != "" && search.DeletionStatus != "active" {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		query = query.Where("(of_dishes.public_id LIKE ? OR of_dishes.name LIKE ?)", keyword, keyword)
	}
	if search.UserID != "" {
		query = query.Where("of_dishes.owner_id = ?", strings.TrimSpace(search.UserID))
	}
	if value := strings.TrimSpace(search.UserKeyword); value != "" {
		keyword := "%" + value + "%"
		ownerIDs := s.db.Model(&contentModel.MiniAppUser{}).
			Select("id").
			Where("id LIKE ? OR nickname LIKE ?", keyword, keyword)
		query = query.Where("of_dishes.owner_id IN (?)", ownerIDs)
	}
	if search.Status != "" {
		if search.Status != contentModel.DishStatusDraft && search.Status != contentModel.DishStatusUsable {
			return result, appErrors.AdminBadRequest.DefaultMsg()
		}
		query = query.Where("of_dishes.status = ?", search.Status)
	}
	if search.Discoverable != nil {
		query = query.Where("of_dishes.discoverable = ?", *search.Discoverable)
	}
	if search.SourceType != "" {
		query = query.Where("of_dishes.source_type = ?", search.SourceType)
	}
	if search.SourceLocked != nil {
		query = query.Where("of_dishes.source_locked = ?", *search.SourceLocked)
	}
	if search.CategoryID != "" {
		query = query.Where(
			"of_dishes.category_id IN (?)",
			s.db.Model(&contentModel.ContentCategory{}).Select("id").Where("public_id = ?", search.CategoryID),
		)
	}
	if len(search.TagIDs) > 0 {
		tagDishIDs := s.db.Table("of_dish_tags").
			Select("DISTINCT of_dish_tags.dish_id").
			Joins("JOIN of_tags ON of_tags.id = of_dish_tags.tag_id").
			Where("of_tags.public_id IN ?", search.TagIDs)
		query = query.Where("of_dishes.id IN (?)", tagDishIDs)
	}
	if search.MediaReviewStatus != "" {
		query = query.Where("of_dishes.media_review_status = ?", search.MediaReviewStatus)
	}
	if search.RecipeID != "" {
		referencedDishIDs := s.db.Table("of_recipe_dishes").
			Select("of_recipe_dishes.dish_id").
			Joins("JOIN of_recipes ON of_recipes.id = of_recipe_dishes.recipe_id").
			Where("of_recipes.public_id = ?", search.RecipeID).
			Where("of_recipes.deleted_at IS NULL")
		query = query.Where("of_dishes.id IN (?)", referencedDishIDs)
	}
	query = applyTimeRange(query, "of_dishes.created_at", search.CreatedFrom, search.CreatedTo)
	query = applyTimeRange(query, "of_dishes.updated_at", search.UpdatedFrom, search.UpdatedTo)
	if search.DeletionStatus == "deleted" {
		query = applyTimeRange(query, "of_dishes.deleted_at", search.DeletedFrom, search.DeletedTo)
	}

	if err := query.Session(&gorm.Session{}).
		Distinct("of_dishes.id").
		Count(&result.Total).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "count user dishes")
	}

	sortColumn, ok := map[string]string{
		"":          "created_at",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		"deletedAt": "deleted_at",
		"name":      "name",
	}[search.SortBy]
	if !ok {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	sortOrder, err := normalizeSortOrder(search.SortOrder)
	if err != nil {
		return result, err
	}

	var dishes []contentModel.UserDish
	err = query.
		Preload("Owner").
		Preload("Category").
		Preload("Tags").
		Order("of_dishes." + sortColumn + " " + sortOrder).
		Order("of_dishes.id " + sortOrder).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&dishes).Error
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "query user dishes")
	}

	result.Page = page
	result.PageSize = pageSize
	result.List = make([]contentRes.UserDishAdminSummary, 0, len(dishes))
	for i := range dishes {
		summary, summaryErr := s.buildDishSummary(ctx, s.db, &dishes[i], includePrivateCover)
		if summaryErr != nil {
			return contentRes.Page[contentRes.UserDishAdminSummary]{}, summaryErr
		}
		result.List = append(result.List, summary)
	}
	return result, nil
}

// GetUserDishDetail 获取用户菜品详情。
func (s *ContentService) GetUserDishDetail(
	ctx context.Context,
	dishID string,
	actor contentReq.AdminActor,
) (contentRes.UserDishAdminDetail, error) {
	var detail contentRes.UserDishAdminDetail
	if s.db == nil || s.permission == nil || s.accessAudit == nil {
		return detail, appErrors.AdminInternal.DefaultMsg()
	}
	if strings.TrimSpace(dishID) == "" {
		return detail, appErrors.AdminBadRequest.DefaultMsg()
	}
	dishID = strings.TrimSpace(dishID)
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionUserDishPrivateRead); err != nil {
		return detail, err
	}
	accessAuditID, err := s.accessAudit.Record(ctx, s.sensitiveAccess(
		actor,
		PermissionUserDishPrivateRead,
		"read_user_dish_private_detail",
		contentModel.GovernanceTargetDish,
		dishID,
	))
	if err != nil {
		return detail, err
	}

	var dish contentModel.UserDish
	err = s.db.WithContext(ctx).Unscoped().
		Preload("Owner").
		Preload("Category").
		Preload("Tags").
		Preload("Ingredients", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("Ingredients.Unit").
		Preload("Steps", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Where("public_id = ?", dishID).
		First(&dish).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return detail, appErrors.AdminNotFound.DefaultMsg()
		}
		return detail, appErrors.AdminInternal.Wrap(err, "query user dish detail")
	}

	summary, err := s.buildDishSummary(ctx, s.db, &dish, true)
	if err != nil {
		return detail, err
	}
	sourceOverview, err := s.buildDishSourceOverview(ctx, s.db, &dish)
	if err != nil {
		return detail, err
	}
	referenceOverview, err := s.buildDishReferenceOverview(ctx, s.db, dish.ID)
	if err != nil {
		return detail, err
	}
	moderationSummary, err := s.latestDishModerationSummary(ctx, &dish)
	if err != nil {
		return detail, err
	}
	detail = contentRes.UserDishAdminDetail{
		UserDishAdminSummary: summary,
		CoverFileID:          dish.CoverFileID,
		Description:          dish.Description,
		Serving:              dish.Serving,
		Ingredients:          make([]contentRes.DishIngredient, 0, len(dish.Ingredients)),
		Steps:                make([]contentRes.DishStep, 0, len(dish.Steps)),
		SourceOverview:       sourceOverview,
		ReferenceOverview:    referenceOverview,
		ModerationSummary:    moderationSummary,
		DeletedReason:        dish.DeletedReason,
		DeletedViolationType: dish.DeletedViolationType,
		DeletedBy:            deletedByAdministrator(dish.DeletedByAdminID, dish.DeletedByUsername, dish.DeletedByNickname),
		AccessAuditID:        accessAuditID,
	}
	for _, ingredient := range dish.Ingredients {
		var unit *contentRes.CatalogReference
		if ingredient.Unit != nil {
			unit = &contentRes.CatalogReference{
				ID: ingredient.Unit.PublicID, Name: ingredient.Unit.Name, Enabled: ingredient.Unit.Enabled,
			}
		}
		detail.Ingredients = append(detail.Ingredients, contentRes.DishIngredient{
			Name: ingredient.Name, Quantity: ingredient.Quantity, Unit: unit, Note: ingredient.Note, SortOrder: ingredient.SortOrder,
		})
	}
	for _, step := range dish.Steps {
		detail.Steps = append(detail.Steps, contentRes.DishStep{
			Description: step.Description, ImageURL: step.ImageURL, SortOrder: step.SortOrder,
		})
	}
	var governanceCount int64
	if err := s.db.WithContext(ctx).Model(&contentModel.GovernanceRecord{}).
		Where("target_type = ? AND target_id = ?", contentModel.GovernanceTargetDish, dish.PublicID).
		Count(&governanceCount).Error; err != nil {
		return contentRes.UserDishAdminDetail{}, appErrors.AdminInternal.Wrap(err, "count dish governance records")
	}
	detail.GovernanceRecordCount = int(governanceCount)
	var auditCount int64
	if err := s.db.WithContext(ctx).Model(&contentModel.AdminAuditLog{}).
		Where("target_type = ? AND target_id = ?", contentModel.GovernanceTargetDish, dish.PublicID).
		Count(&auditCount).Error; err != nil {
		return contentRes.UserDishAdminDetail{}, appErrors.AdminInternal.Wrap(err, "count dish audit logs")
	}
	detail.AuditLogCount = int(auditCount)
	return detail, nil
}

// latestDishModerationSummary 返回用户菜品封面的最新图片审核摘要。
func (s *ContentService) latestDishModerationSummary(
	ctx context.Context,
	dish *contentModel.UserDish,
) (*contentRes.ModerationRecordSummary, error) {
	if s == nil || s.db == nil || dish == nil {
		return nil, appErrors.AdminInternal.DefaultMsg()
	}
	statement := s.db.WithContext(ctx).Model(&contentModel.ImageModerationRecord{})
	if strings.TrimSpace(dish.CoverFileID) != "" {
		statement = statement.Where(
			"file_id = ? OR (object_id = ? AND object_type IN ?)",
			dish.CoverFileID,
			dish.PublicID,
			[]string{"dish", "user_dish"},
		)
	} else {
		statement = statement.Where(
			"object_id = ? AND object_type IN ?",
			dish.PublicID,
			[]string{"dish", "user_dish"},
		)
	}
	var record contentModel.ImageModerationRecord
	if err := statement.Order("created_at DESC, id DESC").First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "load latest user dish moderation record")
	}
	summary, err := moderationRecordSummary(ctx, s.db, record)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ListUserDishReferences 分页查询用户菜品引用列表。
func (s *ContentService) ListUserDishReferences(
	ctx context.Context,
	dishID string,
	search contentReq.UserDishReferenceSearch,
	actor contentReq.AdminActor,
) (contentRes.AuditedPage[contentRes.UserDishReference], error) {
	var result contentRes.AuditedPage[contentRes.UserDishReference]
	if s.db == nil || s.permission == nil || s.accessAudit == nil {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	if strings.TrimSpace(dishID) == "" {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	dishID = strings.TrimSpace(dishID)
	if !validPagination(search.Page, search.PageSize) ||
		(search.ReferenceType != "" && !validReferenceType(search.ReferenceType)) ||
		(search.SortBy != "" && search.SortBy != "occurredAt") {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	sortOrder, err := normalizeSortOrder(search.SortOrder)
	if err != nil {
		return result, err
	}
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionUserDishPrivateRead); err != nil {
		return result, err
	}
	accessAuditID, err := s.accessAudit.Record(ctx, s.sensitiveAccess(
		actor,
		PermissionUserDishPrivateRead,
		"read_user_dish_references",
		contentModel.GovernanceTargetDish,
		dishID,
	))
	if err != nil {
		return result, err
	}
	page, pageSize := normalizePage(search.Page, search.PageSize)
	var dish contentModel.UserDish
	if err := s.db.WithContext(ctx).Unscoped().Select("id").Where("public_id = ?", dishID).First(&dish).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return result, appErrors.AdminNotFound.DefaultMsg()
		}
		return result, appErrors.AdminInternal.Wrap(err, "find referenced dish")
	}
	query := s.db.WithContext(ctx).Model(&contentModel.DishReference{}).Where("dish_id = ?", dish.ID)
	if search.ReferenceType != "" {
		query = query.Where("reference_type = ?", search.ReferenceType)
	}
	if err := query.Session(&gorm.Session{}).Count(&result.Total).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "count dish references")
	}
	var references []contentModel.DishReference
	if err := query.Order("occurred_at " + sortOrder).Order("id " + sortOrder).
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&references).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "query dish references")
	}
	result.Page = page
	result.PageSize = pageSize
	result.AccessAuditID = accessAuditID
	result.List = make([]contentRes.UserDishReference, 0, len(references))
	for _, reference := range references {
		result.List = append(result.List, contentRes.UserDishReference{
			ID: reference.PublicID, ReferenceType: reference.ReferenceType, ObjectID: reference.ObjectID,
			ObjectLabel: reference.ObjectLabel, ObjectStatus: reference.ObjectStatus,
			HistoricalSnapshot: reference.HistoricalSnapshot, OccurredAt: reference.OccurredAt,
		})
	}
	return result, nil
}

// ListUserRecipes 分页查询用户菜谱列表。
func (s *ContentService) ListUserRecipes(
	ctx context.Context,
	search contentReq.UserRecipeSearch,
	actor contentReq.AdminActor,
) (contentRes.Page[contentRes.UserRecipeAdminSummary], error) {
	var result contentRes.Page[contentRes.UserRecipeAdminSummary]
	if s.db == nil || s.permission == nil {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionUserRecipeRead); err != nil {
		return result, err
	}
	includePrivateCover, err := s.permission.HasPermission(ctx, actor.AuthorityID, PermissionUserRecipePrivateRead)
	if err != nil {
		return result, err
	}
	if !validTimeRange(search.CreatedFrom, search.CreatedTo) ||
		!validTimeRange(search.UpdatedFrom, search.UpdatedTo) ||
		!validPagination(search.Page, search.PageSize) ||
		utf8.RuneCountInString(search.Keyword) > 60 ||
		utf8.RuneCountInString(search.UserKeyword) > 60 ||
		(search.ContentState != "" && !validRecipeContentState(search.ContentState)) ||
		(search.MinDishCount != nil && *search.MinDishCount < 0) ||
		(search.MaxDishCount != nil && *search.MaxDishCount < 0) ||
		(search.MinDishCount != nil && search.MaxDishCount != nil && *search.MinDishCount > *search.MaxDishCount) {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	page, pageSize := normalizePage(search.Page, search.PageSize)
	query := s.db.WithContext(ctx).Model(&contentModel.UserRecipe{})
	if search.DeletionStatus == "deleted" {
		query = query.Unscoped().Where("of_recipes.deleted_at IS NOT NULL")
	} else if search.DeletionStatus != "" && search.DeletionStatus != "active" {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		query = query.Where("(of_recipes.public_id LIKE ? OR of_recipes.name LIKE ?)", keyword, keyword)
	}
	if search.UserID != "" {
		query = query.Where("of_recipes.owner_id = ?", strings.TrimSpace(search.UserID))
	}
	if value := strings.TrimSpace(search.UserKeyword); value != "" {
		keyword := "%" + value + "%"
		ownerIDs := s.db.Model(&contentModel.MiniAppUser{}).
			Select("id").
			Where("id LIKE ? OR nickname LIKE ?", keyword, keyword)
		query = query.Where("of_recipes.owner_id IN (?)", ownerIDs)
	}
	if search.DishID != "" {
		recipeIDs := s.db.Table("of_recipe_dishes").
			Select("of_recipe_dishes.recipe_id").
			Joins("JOIN of_dishes ON of_dishes.id = of_recipe_dishes.dish_id").
			Where("of_dishes.public_id = ?", search.DishID)
		query = query.Where("of_recipes.id IN (?)", recipeIDs)
	}
	query = applyTimeRange(query, "of_recipes.created_at", search.CreatedFrom, search.CreatedTo)
	query = applyTimeRange(query, "of_recipes.updated_at", search.UpdatedFrom, search.UpdatedTo)

	var recipes []contentModel.UserRecipe
	err = query.
		Preload("Owner").
		Preload("RecipeDishes", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("RecipeDishes.Dish", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Find(&recipes).Error
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "query user recipes")
	}
	summaries := make([]contentRes.UserRecipeAdminSummary, 0, len(recipes))
	for i := range recipes {
		summary := buildRecipeSummary(&recipes[i], includePrivateCover)
		if search.MinDishCount != nil && summary.DishCount < *search.MinDishCount {
			continue
		}
		if search.MaxDishCount != nil && summary.DishCount > *search.MaxDishCount {
			continue
		}
		if search.ContentState != "" && summary.ContentState != search.ContentState {
			continue
		}
		if search.HasNote != nil && summary.HasNote != *search.HasNote {
			continue
		}
		summaries = append(summaries, summary)
	}
	sortOrder, err := normalizeSortOrder(search.SortOrder)
	if err != nil {
		return result, err
	}
	sortBy := search.SortBy
	if sortBy == "" {
		sortBy = "createdAt"
	}
	if !validRecipeSort(sortBy) {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	sort.SliceStable(summaries, func(i, j int) bool {
		less := recipeSummaryLess(summaries[i], summaries[j], sortBy)
		if sortOrder == "DESC" {
			return !less && !recipeSummaryEqual(summaries[i], summaries[j], sortBy)
		}
		return less
	})
	result.Total = int64(len(summaries))
	start := (page - 1) * pageSize
	if start > len(summaries) {
		start = len(summaries)
	}
	end := start + pageSize
	if end > len(summaries) {
		end = len(summaries)
	}
	result.Page = page
	result.PageSize = pageSize
	result.List = summaries[start:end]
	return result, nil
}

// GetUserRecipeDetail 获取用户菜谱详情。
func (s *ContentService) GetUserRecipeDetail(
	ctx context.Context,
	recipeID string,
	actor contentReq.AdminActor,
) (contentRes.UserRecipeAdminDetail, error) {
	var detail contentRes.UserRecipeAdminDetail
	if s.db == nil || s.permission == nil || s.accessAudit == nil {
		return detail, appErrors.AdminInternal.DefaultMsg()
	}
	if strings.TrimSpace(recipeID) == "" {
		return detail, appErrors.AdminBadRequest.DefaultMsg()
	}
	recipeID = strings.TrimSpace(recipeID)
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionUserRecipePrivateRead); err != nil {
		return detail, err
	}
	accessAuditID, err := s.accessAudit.Record(ctx, s.sensitiveAccess(
		actor,
		PermissionUserRecipePrivateRead,
		"read_user_recipe_private_detail",
		contentModel.GovernanceTargetRecipe,
		recipeID,
	))
	if err != nil {
		return detail, err
	}
	var recipe contentModel.UserRecipe
	err = s.db.WithContext(ctx).Unscoped().
		Preload("Owner").
		Preload("RecipeDishes", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("RecipeDishes.Dish", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Where("public_id = ?", recipeID).
		First(&recipe).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return detail, appErrors.AdminNotFound.DefaultMsg()
		}
		return detail, appErrors.AdminInternal.Wrap(err, "query user recipe detail")
	}
	detail = contentRes.UserRecipeAdminDetail{
		UserRecipeAdminSummary: buildRecipeSummary(&recipe, true),
		Note:                   recipe.Note,
		Dishes:                 buildRecipeDishItems(&recipe),
		DeletedReason:          recipe.DeletedReason,
		DeletedViolationType:   recipe.DeletedViolationType,
		DeletedBy:              deletedByAdministrator(recipe.DeletedByAdminID, recipe.DeletedByUsername, recipe.DeletedByNickname),
		AccessAuditID:          accessAuditID,
	}
	var governanceCount int64
	if err := s.db.WithContext(ctx).Model(&contentModel.GovernanceRecord{}).
		Where("target_type = ? AND target_id = ?", contentModel.GovernanceTargetRecipe, recipe.PublicID).
		Count(&governanceCount).Error; err != nil {
		return contentRes.UserRecipeAdminDetail{}, appErrors.AdminInternal.Wrap(err, "count recipe governance records")
	}
	detail.GovernanceRecordCount = int(governanceCount)
	var auditCount int64
	if err := s.db.WithContext(ctx).Model(&contentModel.AdminAuditLog{}).
		Where("target_type = ? AND target_id = ?", contentModel.GovernanceTargetRecipe, recipe.PublicID).
		Count(&auditCount).Error; err != nil {
		return contentRes.UserRecipeAdminDetail{}, appErrors.AdminInternal.Wrap(err, "count recipe audit logs")
	}
	detail.AuditLogCount = int(auditCount)
	return detail, nil
}

// PreviewGovernance 预览违规处理。
func (s *ContentService) PreviewGovernance(
	ctx context.Context,
	input contentReq.GovernanceActionPreviewInput,
	actor contentReq.AdminActor,
) (contentRes.GovernanceImpactPreview, error) {
	var preview contentRes.GovernanceImpactPreview
	if s.db == nil || s.permission == nil || len(s.effectivePreviewSecret()) == 0 {
		return preview, appErrors.AdminInternal.DefaultMsg()
	}
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionGovernanceExecute); err != nil {
		return preview, err
	}
	if err := validateGovernanceInput(input); err != nil {
		return preview, err
	}
	if containsGovernanceAction(input.Actions, contentModel.GovernanceActionDeleteCopyChain) {
		if err := s.permission.Require(ctx, actor.AuthorityID, PermissionGovernanceCascade); err != nil {
			return preview, err
		}
	}
	impact, err := s.buildGovernanceImpact(ctx, s.db, input)
	if err != nil {
		return preview, err
	}
	expiresAt := s.now().Add(governancePreviewTTL)
	requestHash := governanceRequestHash(input)
	token, err := s.signPreviewToken(requestHash, expiresAt)
	if err != nil {
		return preview, appErrors.AdminInternal.Wrap(err, "sign governance preview")
	}
	impact.PreviewToken = token
	impact.ExpiresAt = expiresAt
	return impact, nil
}

// ExecuteGovernance 执行违规处理。
func (s *ContentService) ExecuteGovernance(
	ctx context.Context,
	input contentReq.GovernanceActionExecuteInput,
	actor contentReq.AdminActor,
	idempotencyKey string,
) (contentRes.GovernanceExecutionResult, error) {
	var result contentRes.GovernanceExecutionResult
	if s.db == nil || s.idempotency == nil || s.permission == nil ||
		s.mutationAudit == nil || s.notification == nil || len(s.effectivePreviewSecret()) == 0 {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	if actor.AdministratorID == 0 || strings.TrimSpace(actor.RequestID) == "" {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := s.permission.Require(ctx, actor.AuthorityID, PermissionGovernanceExecute); err != nil {
		return result, err
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return result, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateGovernanceInput(input.GovernanceActionPreviewInput); err != nil {
		return result, err
	}
	if containsGovernanceAction(input.Actions, contentModel.GovernanceActionDeleteCopyChain) {
		if err := s.permission.Require(ctx, actor.AuthorityID, PermissionGovernanceCascade); err != nil {
			return result, err
		}
		if input.ConfirmText != governanceCopyChainConfirmText {
			return result, appErrors.AdminInvalidGovernance.DefaultMsg()
		}
	}
	requestHash := governanceRequestHash(input.GovernanceActionPreviewInput)
	if err := s.verifyPreviewToken(input.PreviewToken, requestHash); err != nil {
		return result, err
	}
	rawResult, _, err := s.idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"governance_action_execute",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			return s.executeGovernanceOnce(ctx, tx, input.GovernanceActionPreviewInput, actor, idempotencyKey)
		},
	)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(rawResult, &result); err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "decode governance result")
	}
	return result, nil
}

func (s *ContentService) executeGovernanceOnce(
	ctx context.Context,
	tx *gorm.DB,
	input contentReq.GovernanceActionPreviewInput,
	actor contentReq.AdminActor,
	idempotencyKey string,
) (contentRes.GovernanceExecutionResult, error) {
	var result contentRes.GovernanceExecutionResult
	if tx == nil {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	impact, err := s.buildGovernanceImpact(ctx, tx, input)
	if err != nil {
		return result, err
	}
	if containsGovernanceAction(input.Actions, contentModel.GovernanceActionDeleteCopyChain) {
		if input.TargetType != contentModel.GovernanceTargetOfficialDish || impact.CopiedDishCount > 0 {
			return s.createCopyChainGovernanceJob(ctx, tx, input, actor, idempotencyKey, impact)
		}
	}
	beforeSummary := map[string]interface{}{}
	afterSummary := map[string]interface{}{}
	targetLabel := ""
	recipientUserID := ""
	notificationTitle := ""

	switch input.TargetType {
	case contentModel.GovernanceTargetDish:
		var dish contentModel.UserDish
		if err := tx.Unscoped().Where("public_id = ?", input.TargetID).First(&dish).Error; err != nil {
			return result, mapTargetLookupError(err)
		}
		if dish.DeletedAt.Valid || dish.Version != input.ExpectedVersion {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		targetLabel = dish.Name
		recipientUserID = dish.OwnerID
		notificationTitle = "菜品处理通知"
		beforeSummary = map[string]interface{}{
			"status": dish.Status, "discoverable": dish.Discoverable, "deleted": dish.DeletedAt.Valid, "version": dish.Version,
		}
		updates := map[string]interface{}{"version": dish.Version + 1}
		softDelete := false
		discoverableAfter := dish.Discoverable
		for _, action := range input.Actions {
			switch action {
			case contentModel.GovernanceActionDisableDiscoverability:
				updates["discoverable"] = false
				updates["discoverable_at"] = nil
				discoverableAfter = false
			case contentModel.GovernanceActionSoftDeleteDish:
				updates["discoverable"] = false
				updates["discoverable_at"] = nil
				updates["deleted_reason"] = input.Reason
				updates["deleted_violation_type"] = input.ViolationType
				updates["deleted_by_admin_id"] = actor.AdministratorID
				updates["deleted_by_username"] = actor.Username
				updates["deleted_by_nickname"] = actor.Nickname
				softDelete = true
			default:
				return result, appErrors.AdminInvalidGovernance.DefaultMsg()
			}
		}
		update := tx.Model(&dish).
			Where("version = ?", input.ExpectedVersion).
			Updates(updates)
		if update.Error != nil {
			return result, appErrors.AdminInternal.Wrap(update.Error, "update governed dish")
		}
		if update.RowsAffected != 1 {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		offlineCount, err := offlineGovernedSourceRecommendations(
			ctx, tx, "creator", dish.PublicID, &dish.ID, nil, actor, input.Reason, s.now().UTC(),
		)
		if err != nil {
			return result, err
		}
		if softDelete {
			if err := removeGovernedDishLiveReferences(ctx, tx, dish.ID); err != nil {
				return result, err
			}
			if err := tx.Delete(&dish).Error; err != nil {
				return result, appErrors.AdminInternal.Wrap(err, "soft delete governed dish")
			}
		}
		afterSummary = map[string]interface{}{
			"status": dish.Status, "discoverable": discoverableAfter, "deleted": softDelete,
			"version": dish.Version + 1, "offlineRecommendationCount": offlineCount,
		}

	case contentModel.GovernanceTargetOfficialDish:
		var dish contentModel.OfficialDish
		if err := tx.Unscoped().Where("public_id = ?", input.TargetID).First(&dish).Error; err != nil {
			return result, mapTargetLookupError(err)
		}
		if dish.DeletedAt.Valid || dish.Version != int64(input.ExpectedVersion) {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		if len(input.Actions) != 1 ||
			(input.Actions[0] != contentModel.GovernanceActionSoftDeleteOfficialDish &&
				input.Actions[0] != contentModel.GovernanceActionDeleteCopyChain) {
			return result, appErrors.AdminInvalidGovernance.DefaultMsg()
		}
		targetLabel = dish.Name
		beforeSummary = map[string]interface{}{
			"status": dish.Status, "deleted": false, "version": dish.Version,
		}
		update := tx.Model(&dish).
			Where("version = ?", input.ExpectedVersion).
			Updates(map[string]interface{}{
				"deleted_reason":      input.Reason,
				"deleted_by_id":       actor.AdministratorID,
				"updated_by_id":       actor.AdministratorID,
				"updated_by_username": actor.Username,
				"updated_by_nickname": actor.Nickname,
				"version":             dish.Version + 1,
			})
		if update.Error != nil {
			return result, appErrors.AdminInternal.Wrap(update.Error, "update governed official dish")
		}
		if update.RowsAffected != 1 {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		offlineCount, err := offlineGovernedSourceRecommendations(
			ctx, tx, "official", dish.PublicID, nil, &dish.ID, actor, input.Reason, s.now().UTC(),
		)
		if err != nil {
			return result, err
		}
		if err := tx.Delete(&dish).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "soft delete governed official dish")
		}
		afterSummary = map[string]interface{}{
			"status": dish.Status, "deleted": true, "version": dish.Version + 1,
			"offlineRecommendationCount": offlineCount,
		}
	case contentModel.GovernanceTargetRecipe:
		var recipe contentModel.UserRecipe
		if err := tx.Unscoped().Where("public_id = ?", input.TargetID).First(&recipe).Error; err != nil {
			return result, mapTargetLookupError(err)
		}
		if recipe.DeletedAt.Valid || recipe.Version != input.ExpectedVersion {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		if len(input.Actions) != 1 || input.Actions[0] != contentModel.GovernanceActionSoftDeleteRecipe {
			return result, appErrors.AdminInvalidGovernance.DefaultMsg()
		}
		targetLabel = recipe.Name
		recipientUserID = recipe.OwnerID
		notificationTitle = "菜谱处理通知"
		beforeSummary = map[string]interface{}{"deleted": false, "version": recipe.Version}
		updates := map[string]interface{}{
			"version":                recipe.Version + 1,
			"deleted_reason":         input.Reason,
			"deleted_violation_type": input.ViolationType,
			"deleted_by_admin_id":    actor.AdministratorID,
			"deleted_by_username":    actor.Username,
			"deleted_by_nickname":    actor.Nickname,
		}
		update := tx.Model(&recipe).
			Where("version = ?", input.ExpectedVersion).
			Updates(updates)
		if update.Error != nil {
			return result, appErrors.AdminInternal.Wrap(update.Error, "update governed recipe")
		}
		if update.RowsAffected != 1 {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		if err := tx.Delete(&recipe).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "soft delete governed recipe")
		}
		afterSummary = map[string]interface{}{"deleted": true, "version": recipe.Version + 1}
	case contentModel.GovernanceTargetCheckin:
		var checkin contentModel.FrontCheckin
		if err := tx.Unscoped().Where("id = ?", input.TargetID).First(&checkin).Error; err != nil {
			return result, mapTargetLookupError(err)
		}
		if checkin.DeletedAt.Valid || checkin.Version != input.ExpectedVersion {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		if len(input.Actions) != 1 || input.Actions[0] != contentModel.GovernanceActionSoftDeleteCheckin {
			return result, appErrors.AdminInvalidGovernance.DefaultMsg()
		}
		targetLabel = checkin.DishName + "打卡"
		recipientUserID = checkin.UserID
		notificationTitle = "打卡处理通知"
		beforeSummary = map[string]interface{}{
			"deleted": false, "rewarded": checkin.Rewarded, "version": checkin.Version,
		}
		updates := map[string]interface{}{
			"version":                checkin.Version + 1,
			"deleted_reason":         input.Reason,
			"deleted_violation_type": input.ViolationType,
			"deleted_by_admin_id":    actor.AdministratorID,
			"deleted_by_username":    actor.Username,
			"deleted_by_nickname":    actor.Nickname,
		}
		update := tx.Model(&checkin).
			Where("version = ?", input.ExpectedVersion).
			Updates(updates)
		if update.Error != nil {
			return result, appErrors.AdminInternal.Wrap(update.Error, "update governed checkin")
		}
		if update.RowsAffected != 1 {
			return result, appErrors.AdminStateConflict.DefaultMsg()
		}
		if err := tx.Delete(&checkin).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "soft delete governed checkin")
		}
		var remainingDailyCount int64
		if err := tx.Model(&contentModel.FrontCheckin{}).
			Where("user_id = ? AND checked_date = ?", checkin.UserID, checkin.CheckedDate).
			Count(&remainingDailyCount).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "count remaining governed checkins")
		}
		userUpdates := map[string]interface{}{
			"checkin_count": gorm.Expr("GREATEST(checkin_count - 1, 0)"),
		}
		if remainingDailyCount == 0 {
			userUpdates["checkin_day_count"] = gorm.Expr("GREATEST(checkin_day_count - 1, 0)")
		}
		if err := tx.Model(&contentModel.MiniAppUser{}).
			Where("id = ?", checkin.UserID).
			Updates(userUpdates).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "update governed checkin counters")
		}
		afterSummary = map[string]interface{}{
			"deleted": true, "pointsPreserved": true, "version": checkin.Version + 1,
		}
	default:
		return result, appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	if recipientUserID == "" && input.TargetType != contentModel.GovernanceTargetOfficialDish {
		return result, appErrors.AdminInternal.DefaultMsg()
	}

	notificationIDs := make([]string, 0, 1)
	if recipientUserID != "" {
		targetTypeCopy := input.TargetType
		targetIDCopy := input.TargetID
		notificationID, err := s.notification.WriteGovernanceNotification(ctx, tx, &contentModel.UserNotification{
			UserID:            recipientUserID,
			Type:              "governance",
			Title:             notificationTitle,
			Content:           fmt.Sprintf("“%s”已被平台处理。原因：%s", targetLabel, input.Reason),
			TargetType:        &targetTypeCopy,
			TargetID:          &targetIDCopy,
			SubscribeRequired: false,
			CreatedAt:         s.now().UTC(),
		})
		if err != nil {
			return result, err
		}
		notificationIDs = append(notificationIDs, notificationID)
	}

	actionsJSON, err := json.Marshal(input.Actions)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode governance actions")
	}
	impactJSON, err := json.Marshal(impact)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode governance impact")
	}
	beforeJSON, err := json.Marshal(beforeSummary)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode governance before summary")
	}
	afterJSON, err := json.Marshal(afterSummary)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode governance after summary")
	}
	notificationJSON, err := json.Marshal(notificationIDs)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode governance notification ids")
	}
	record := contentModel.GovernanceRecord{
		PublicID:              newPublicID("gov"),
		TargetType:            input.TargetType,
		TargetID:              input.TargetID,
		TargetLabel:           targetLabel,
		Actions:               actionsJSON,
		ViolationType:         input.ViolationType,
		Severity:              input.Severity,
		Reason:                input.Reason,
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		ImpactSnapshot:        impactJSON,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		NotificationIDs:       notificationJSON,
		RequestID:             actor.RequestID,
		IdempotencyKey:        idempotencyKey,
		AffectedCount:         1,
	}
	if err := tx.Create(&record).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create governance record")
	}
	targetLabelCopy := targetLabel
	reasonCopy := input.Reason
	idempotencyKeyCopy := idempotencyKey
	audit := contentModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                "execute_governance",
		TargetType:            input.TargetType,
		TargetID:              input.TargetID,
		TargetLabel:           &targetLabelCopy,
		Reason:                &reasonCopy,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             actor.RequestID,
		IdempotencyKey:        &idempotencyKeyCopy,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	if err := s.mutationAudit.WriteMutationAudit(ctx, tx, &audit); err != nil {
		return result, err
	}
	result = contentRes.GovernanceExecutionResult{
		RecordID: record.PublicID, JobID: nil, Status: "succeeded", AffectedCount: 1,
	}
	return result, nil
}

// offlineGovernedSourceRecommendations 下线与违规源菜品关联的全部在线推荐。
func offlineGovernedSourceRecommendations(
	ctx context.Context,
	tx *gorm.DB,
	sourceType string,
	sourcePublicID string,
	userDishID *uint,
	officialDishID *uint,
	actor contentReq.AdminActor,
	reason string,
	now time.Time,
) (int64, error) {
	if tx == nil {
		return 0, appErrors.AdminInternal.DefaultMsg()
	}
	statement := tx.WithContext(ctx).Model(&contentModel.PlatformRecommendation{}).
		Where("status = ?", recommendationStatusPublished)
	switch {
	case sourceType == "creator" && userDishID != nil:
		statement = statement.Where(
			"(source_type = ? AND source_dish_id = ?) OR dish_id = ?",
			sourceType,
			sourcePublicID,
			*userDishID,
		)
	case sourceType == "official" && officialDishID != nil:
		statement = statement.Where(
			"(source_type = ? AND source_dish_id = ?) OR official_dish_id = ?",
			sourceType,
			sourcePublicID,
			*officialDishID,
		)
	default:
		return 0, appErrors.AdminInternal.DefaultMsg()
	}
	offlineReason := truncateRunes("来源菜品被平台处理："+strings.TrimSpace(reason), 200)
	update := statement.Updates(map[string]interface{}{
		"status":              recommendationStatusOffline,
		"selected":            false,
		"offline_at":          now,
		"offline_reason":      offlineReason,
		"updated_by_id":       actor.AdministratorID,
		"updated_by_username": actor.Username,
		"updated_by_nickname": actor.Nickname,
		"updated_at":          now,
		"version":             gorm.Expr("version + 1"),
	})
	if update.Error != nil {
		return 0, appErrors.AdminInternal.Wrap(update.Error, "offline governed source recommendations")
	}
	return update.RowsAffected, nil
}

// countGovernedSourceRecommendations 统计与源菜品关联且当前在线的推荐数量。
func countGovernedSourceRecommendations(
	ctx context.Context,
	db *gorm.DB,
	sourceType string,
	sourcePublicID string,
	userDishID *uint,
	officialDishID *uint,
) (int64, error) {
	if db == nil {
		return 0, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&contentModel.PlatformRecommendation{}).
		Where("status = ?", recommendationStatusPublished)
	switch {
	case sourceType == "creator" && userDishID != nil:
		statement = statement.Where(
			"(source_type = ? AND source_dish_id = ?) OR dish_id = ?",
			sourceType,
			sourcePublicID,
			*userDishID,
		)
	case sourceType == "official" && officialDishID != nil:
		statement = statement.Where(
			"(source_type = ? AND source_dish_id = ?) OR official_dish_id = ?",
			sourceType,
			sourcePublicID,
			*officialDishID,
		)
	default:
		return 0, appErrors.AdminInternal.DefaultMsg()
	}
	var count int64
	if err := statement.Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count governed source recommendations")
	}
	return count, nil
}

// removeGovernedDishLiveReferences 清理被软删除菜品的实时关系并保留历史快照。
func removeGovernedDishLiveReferences(ctx context.Context, tx *gorm.DB, dishID uint) error {
	if tx == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	if err := tx.WithContext(ctx).Where("dish_id = ?", dishID).
		Delete(&contentModel.RecipeDish{}).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "remove governed dish from recipes")
	}
	unavailableReason := "source_deleted"
	if err := tx.WithContext(ctx).Model(&contentModel.FrontMealCandidate{}).
		Where(
			"dish_id = ? AND meal_id IN (?)",
			dishID,
			tx.Model(&contentModel.FrontMeal{}).Select("id").
				Where("status IN ?", []contentModel.MealStatus{
					contentModel.MealCollecting,
					contentModel.MealClosed,
				}),
		).
		Updates(map[string]interface{}{
			"available": false, "unavailable_reason": unavailableReason,
		}).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "invalidate governed dish meal candidates")
	}
	return nil
}

// createCopyChainGovernanceJob 创建严重违规复制链后台处理任务。
func (s *ContentService) createCopyChainGovernanceJob(
	ctx context.Context,
	tx *gorm.DB,
	input contentReq.GovernanceActionPreviewInput,
	actor contentReq.AdminActor,
	idempotencyKey string,
	impact contentRes.GovernanceImpactPreview,
) (contentRes.GovernanceExecutionResult, error) {
	var result contentRes.GovernanceExecutionResult
	if tx == nil {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	if input.TargetType == contentModel.GovernanceTargetOfficialDish {
		return s.createOfficialCopyChainGovernanceJob(
			ctx, tx, input, actor, idempotencyKey, impact,
		)
	}
	var source contentModel.UserDish
	if err := tx.WithContext(ctx).Unscoped().
		Where("public_id = ?", input.TargetID).
		First(&source).Error; err != nil {
		return result, mapTargetLookupError(err)
	}
	if source.DeletedAt.Valid || source.Version != input.ExpectedVersion {
		return result, appErrors.AdminStateConflict.DefaultMsg()
	}
	var targets []contentModel.UserDish
	if err := tx.WithContext(ctx).
		Where(
			"id = ? OR direct_source_dish_id = ? OR root_source_dish_id = ?",
			source.ID,
			source.ID,
			source.ID,
		).
		Order("id ASC").
		Find(&targets).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "list governed copy chain")
	}
	if len(targets) == 0 {
		return result, appErrors.AdminStateConflict.DefaultMsg()
	}

	now := s.now().UTC()
	recordID := newPublicID("gov")
	jobID := newPublicID("govjob")
	jobStatus := contentModel.GovernanceJobPending
	actionsJSON, err := json.Marshal(input.Actions)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode copy chain governance actions")
	}
	impactJSON, err := json.Marshal(impact)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode copy chain governance impact")
	}
	beforeSummary, err := json.Marshal(map[string]interface{}{
		"sourceDishId":  source.PublicID,
		"sourceVersion": source.Version,
		"targetCount":   len(targets),
	})
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode copy chain before summary")
	}
	afterSummary, err := json.Marshal(map[string]interface{}{
		"jobCreated": true,
		"jobId":      jobID,
		"jobStatus":  jobStatus,
	})
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode copy chain after summary")
	}
	emptyNotifications, err := json.Marshal([]string{})
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode empty governance notifications")
	}
	record := contentModel.GovernanceRecord{
		PublicID:              recordID,
		TargetType:            input.TargetType,
		TargetID:              input.TargetID,
		TargetLabel:           source.Name,
		Actions:               actionsJSON,
		ViolationType:         input.ViolationType,
		Severity:              input.Severity,
		Reason:                input.Reason,
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		JobStatus:             &jobStatus,
		ImpactSnapshot:        impactJSON,
		BeforeSummary:         beforeSummary,
		AfterSummary:          afterSummary,
		NotificationIDs:       emptyNotifications,
		RequestID:             actor.RequestID,
		IdempotencyKey:        idempotencyKey,
		AffectedCount:         0,
	}
	if err := tx.WithContext(ctx).Create(&record).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create copy chain governance record")
	}
	job := contentModel.GovernanceJob{
		ID:           jobID,
		RecordID:     recordID,
		Status:       contentModel.GovernanceJobPending,
		TotalCount:   len(targets),
		PendingCount: len(targets),
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := tx.WithContext(ctx).Create(&job).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create copy chain governance job")
	}
	items := make([]contentModel.GovernanceJobItem, 0, len(targets))
	for index := range targets {
		items = append(items, contentModel.GovernanceJobItem{
			ID:           newPublicID("govitem"),
			JobID:        jobID,
			DishID:       targets[index].ID,
			DishPublicID: targets[index].PublicID,
			UserID:       targets[index].OwnerID,
			Status:       contentModel.GovernanceJobItemPending,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if err := tx.WithContext(ctx).Create(&items).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create copy chain governance job items")
	}
	targetLabel := source.Name
	reason := input.Reason
	idempotencyKeyCopy := idempotencyKey
	audit := contentModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                "create_copy_chain_governance_job",
		TargetType:            input.TargetType,
		TargetID:              input.TargetID,
		TargetLabel:           &targetLabel,
		Reason:                &reason,
		BeforeSummary:         beforeSummary,
		AfterSummary:          afterSummary,
		RequestID:             actor.RequestID,
		IdempotencyKey:        &idempotencyKeyCopy,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	if err := s.mutationAudit.WriteMutationAudit(ctx, tx, &audit); err != nil {
		return result, err
	}
	return contentRes.GovernanceExecutionResult{
		RecordID:      recordID,
		JobID:         &jobID,
		Status:        contentModel.GovernanceJobPending,
		AffectedCount: len(targets),
	}, nil
}

// createOfficialCopyChainGovernanceJob 处理官方源菜品并为其用户复制菜品创建后台任务。
func (s *ContentService) createOfficialCopyChainGovernanceJob(
	ctx context.Context,
	tx *gorm.DB,
	input contentReq.GovernanceActionPreviewInput,
	actor contentReq.AdminActor,
	idempotencyKey string,
	impact contentRes.GovernanceImpactPreview,
) (contentRes.GovernanceExecutionResult, error) {
	var result contentRes.GovernanceExecutionResult
	if tx == nil {
		return result, appErrors.AdminInternal.DefaultMsg()
	}
	var source contentModel.OfficialDish
	if err := tx.WithContext(ctx).Unscoped().
		Where("public_id = ?", input.TargetID).
		First(&source).Error; err != nil {
		return result, mapTargetLookupError(err)
	}
	if source.DeletedAt.Valid || source.Version != int64(input.ExpectedVersion) {
		return result, appErrors.AdminStateConflict.DefaultMsg()
	}
	var targets []contentModel.UserDish
	if err := tx.WithContext(ctx).
		Where(
			"direct_source_official_dish_id = ? OR root_source_official_dish_id = ?",
			source.ID,
			source.ID,
		).
		Order("id ASC").
		Find(&targets).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "list governed official dish copies")
	}
	if len(targets) == 0 {
		return result, appErrors.AdminStateConflict.DefaultMsg()
	}

	now := s.now().UTC()
	update := tx.WithContext(ctx).Model(&source).
		Where("version = ?", input.ExpectedVersion).
		Updates(map[string]interface{}{
			"deleted_reason":      input.Reason,
			"deleted_by_id":       actor.AdministratorID,
			"updated_by_id":       actor.AdministratorID,
			"updated_by_username": actor.Username,
			"updated_by_nickname": actor.Nickname,
			"version":             source.Version + 1,
		})
	if update.Error != nil {
		return result, appErrors.AdminInternal.Wrap(update.Error, "update governed official source")
	}
	if update.RowsAffected != 1 {
		return result, appErrors.AdminStateConflict.DefaultMsg()
	}
	offlineCount, err := offlineGovernedSourceRecommendations(
		ctx, tx, "official", source.PublicID, nil, &source.ID, actor, input.Reason, now,
	)
	if err != nil {
		return result, err
	}
	if err := tx.WithContext(ctx).Delete(&source).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "soft delete governed official source")
	}

	recordID := newPublicID("gov")
	jobID := newPublicID("govjob")
	jobStatus := contentModel.GovernanceJobPending
	actionsJSON, err := json.Marshal(input.Actions)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode official copy chain actions")
	}
	impactJSON, err := json.Marshal(impact)
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode official copy chain impact")
	}
	beforeSummary, err := json.Marshal(map[string]interface{}{
		"sourceDishId":  source.PublicID,
		"sourceVersion": source.Version,
		"targetCount":   len(targets),
	})
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode official copy chain before summary")
	}
	afterSummary, err := json.Marshal(map[string]interface{}{
		"sourceDeleted":              true,
		"offlineRecommendationCount": offlineCount,
		"jobCreated":                 true,
		"jobId":                      jobID,
		"jobStatus":                  jobStatus,
	})
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode official copy chain after summary")
	}
	emptyNotifications, err := json.Marshal([]string{})
	if err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "encode official copy chain notifications")
	}
	record := contentModel.GovernanceRecord{
		PublicID:              recordID,
		TargetType:            input.TargetType,
		TargetID:              input.TargetID,
		TargetLabel:           source.Name,
		Actions:               actionsJSON,
		ViolationType:         input.ViolationType,
		Severity:              input.Severity,
		Reason:                input.Reason,
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		JobStatus:             &jobStatus,
		ImpactSnapshot:        impactJSON,
		BeforeSummary:         beforeSummary,
		AfterSummary:          afterSummary,
		NotificationIDs:       emptyNotifications,
		RequestID:             actor.RequestID,
		IdempotencyKey:        idempotencyKey,
		AffectedCount:         1,
	}
	if err := tx.WithContext(ctx).Create(&record).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create official copy chain record")
	}
	job := contentModel.GovernanceJob{
		ID:           jobID,
		RecordID:     recordID,
		Status:       contentModel.GovernanceJobPending,
		TotalCount:   len(targets),
		PendingCount: len(targets),
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := tx.WithContext(ctx).Create(&job).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create official copy chain job")
	}
	items := make([]contentModel.GovernanceJobItem, 0, len(targets))
	for index := range targets {
		items = append(items, contentModel.GovernanceJobItem{
			ID:           newPublicID("govitem"),
			JobID:        jobID,
			DishID:       targets[index].ID,
			DishPublicID: targets[index].PublicID,
			UserID:       targets[index].OwnerID,
			Status:       contentModel.GovernanceJobItemPending,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if err := tx.WithContext(ctx).Create(&items).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "create official copy chain items")
	}
	targetLabel := source.Name
	reason := input.Reason
	idempotencyKeyCopy := idempotencyKey
	audit := contentModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                "create_official_copy_chain_governance_job",
		TargetType:            input.TargetType,
		TargetID:              input.TargetID,
		TargetLabel:           &targetLabel,
		Reason:                &reason,
		BeforeSummary:         beforeSummary,
		AfterSummary:          afterSummary,
		RequestID:             actor.RequestID,
		IdempotencyKey:        &idempotencyKeyCopy,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	if err := s.mutationAudit.WriteMutationAudit(ctx, tx, &audit); err != nil {
		return result, err
	}
	return contentRes.GovernanceExecutionResult{
		RecordID:      recordID,
		JobID:         &jobID,
		Status:        contentModel.GovernanceJobPending,
		AffectedCount: len(targets) + 1,
	}, nil
}

func (s *ContentService) buildDishSummary(
	ctx context.Context,
	db *gorm.DB,
	dish *contentModel.UserDish,
	includePrivateCover bool,
) (contentRes.UserDishAdminSummary, error) {
	var deletedAt *time.Time
	if dish.DeletedAt.Valid {
		value := dish.DeletedAt.Time
		deletedAt = &value
	}
	coverURL := dish.CoverURL
	if !includePrivateCover && (!dish.Discoverable || dish.Status == contentModel.DishStatusDraft || dish.DeletedAt.Valid) {
		coverURL = nil
	}
	var recipeReferenceCount int64
	err := db.WithContext(ctx).Table("of_recipe_dishes").
		Joins("JOIN of_recipes ON of_recipes.id = of_recipe_dishes.recipe_id").
		Where("of_recipe_dishes.dish_id = ?", dish.ID).
		Where("of_recipes.deleted_at IS NULL").
		Count(&recipeReferenceCount).Error
	if err != nil {
		return contentRes.UserDishAdminSummary{}, appErrors.AdminInternal.Wrap(err, "count recipe references")
	}
	var unconfirmedMealReferenceCount int64
	if err := db.WithContext(ctx).Model(&contentModel.DishReference{}).
		Where("dish_id = ? AND reference_type = ?", dish.ID, contentModel.ReferenceTypeUnconfirmedMeal).
		Count(&unconfirmedMealReferenceCount).Error; err != nil {
		return contentRes.UserDishAdminSummary{}, appErrors.AdminInternal.Wrap(err, "count meal references")
	}
	tags := make([]contentRes.CatalogReference, 0, len(dish.Tags))
	for _, tag := range dish.Tags {
		tags = append(tags, contentRes.CatalogReference{ID: tag.PublicID, Name: tag.Name, Enabled: tag.Enabled})
	}
	deletionStatus := "active"
	if dish.DeletedAt.Valid {
		deletionStatus = "deleted"
	}
	return contentRes.UserDishAdminSummary{
		ID:       dish.PublicID,
		CoverURL: coverURL,
		Name:     dish.Name,
		Owner:    userReference(dish.Owner),
		Category: contentRes.CatalogReference{
			ID: dish.Category.PublicID, Name: dish.Category.Name, Enabled: dish.Category.Enabled,
		},
		Tags:                          tags,
		Status:                        dish.Status,
		DeletionStatus:                deletionStatus,
		Discoverable:                  dish.Discoverable,
		DiscoverableAt:                dish.DiscoverableAt,
		SourceType:                    dish.SourceType,
		SourceLocked:                  dish.SourceLocked,
		RecipeReferenceCount:          int(recipeReferenceCount),
		UnconfirmedMealReferenceCount: int(unconfirmedMealReferenceCount),
		MediaReviewStatus:             dish.MediaReviewStatus,
		RecommendationCount:           dish.RecommendationCount,
		Version:                       dish.Version,
		CreatedAt:                     dish.CreatedAt,
		UpdatedAt:                     dish.UpdatedAt,
		DeletedAt:                     deletedAt,
	}, nil
}

func (s *ContentService) buildDishReferenceOverview(
	ctx context.Context,
	db *gorm.DB,
	dishID uint,
) (contentRes.UserDishReferenceOverview, error) {
	var overview contentRes.UserDishReferenceOverview
	var activeRecipeCount int64
	if err := db.WithContext(ctx).Table("of_recipe_dishes").
		Joins("JOIN of_recipes ON of_recipes.id = of_recipe_dishes.recipe_id").
		Where("of_recipe_dishes.dish_id = ? AND of_recipes.deleted_at IS NULL", dishID).
		Count(&activeRecipeCount).Error; err != nil {
		return overview, appErrors.AdminInternal.Wrap(err, "count active recipe references")
	}
	overview.ActiveRecipeCount = int(activeRecipeCount)
	counts := map[string]*int{
		contentModel.ReferenceTypeUnconfirmedMeal:       &overview.UnconfirmedMealCount,
		contentModel.ReferenceTypeConfirmedMealSnapshot: &overview.ConfirmedMealSnapshotCount,
		contentModel.ReferenceTypeShoppingSnapshot:      &overview.ShoppingSnapshotCount,
	}
	for referenceType, target := range counts {
		var count int64
		if err := db.WithContext(ctx).Model(&contentModel.DishReference{}).
			Where("dish_id = ? AND reference_type = ?", dishID, referenceType).
			Count(&count).Error; err != nil {
			return overview, appErrors.AdminInternal.Wrap(err, "count dish reference overview")
		}
		*target = int(count)
	}
	return overview, nil
}

func (s *ContentService) buildDishSourceOverview(
	ctx context.Context,
	db *gorm.DB,
	dish *contentModel.UserDish,
) (contentRes.UserDishSourceOverview, error) {
	overview := contentRes.UserDishSourceOverview{
		SourceType: dish.SourceType, OperationSourceType: dish.OperationSourceType, OperationSourceID: dish.OperationSourceID,
		ChainDepth: dish.ChainDepth,
	}
	var err error
	if dish.DirectSourceDishID != nil {
		overview.DirectSource, err = loadDishSourceNode(ctx, db, *dish.DirectSourceDishID)
		if err != nil {
			return overview, err
		}
	}
	if dish.RootSourceDishID != nil {
		overview.RootSource, err = loadDishSourceNode(ctx, db, *dish.RootSourceDishID)
		if err != nil {
			return overview, err
		}
	}
	if dish.DirectSourceOfficialDishID != nil {
		overview.DirectSource, err = loadOfficialDishSourceNode(
			ctx, db, *dish.DirectSourceOfficialDishID,
		)
		if err != nil {
			return overview, err
		}
	}
	if dish.RootSourceOfficialDishID != nil {
		overview.RootSource, err = loadOfficialDishSourceNode(
			ctx, db, *dish.RootSourceOfficialDishID,
		)
		if err != nil {
			return overview, err
		}
	}
	if dish.OriginalAuthorID != nil {
		var author contentModel.MiniAppUser
		if err := db.WithContext(ctx).Where("id = ?", *dish.OriginalAuthorID).First(&author).Error; err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				return overview, appErrors.AdminInternal.Wrap(err, "query original dish author")
			}
		} else {
			reference := userReference(author)
			overview.OriginalAuthor = &reference
		}
	}
	var directCopyCount int64
	if err := db.WithContext(ctx).Unscoped().Model(&contentModel.UserDish{}).
		Where("direct_source_dish_id = ?", dish.ID).Count(&directCopyCount).Error; err != nil {
		return overview, appErrors.AdminInternal.Wrap(err, "count direct dish copies")
	}
	overview.DirectCopyCount = int(directCopyCount)
	var descendantCopyCount int64
	if err := db.WithContext(ctx).Unscoped().Model(&contentModel.UserDish{}).
		Where("root_source_dish_id = ? AND id <> ?", dish.ID, dish.ID).Count(&descendantCopyCount).Error; err != nil {
		return overview, appErrors.AdminInternal.Wrap(err, "count descendant dish copies")
	}
	overview.DescendantCopyCount = int(descendantCopyCount)
	return overview, nil
}

func (s *ContentService) buildGovernanceImpact(
	ctx context.Context,
	db *gorm.DB,
	input contentReq.GovernanceActionPreviewInput,
) (contentRes.GovernanceImpactPreview, error) {
	impact := contentRes.GovernanceImpactPreview{
		TargetType: input.TargetType, TargetID: input.TargetID, AffectedUserCount: 1,
		NotificationCount: 1, Asynchronous: false, Warnings: []string{}, ExpectedVersion: input.ExpectedVersion,
	}
	if entryID := strings.TrimSpace(input.EntryRecommendationID); entryID != "" {
		if err := validateGovernanceEntryRecommendation(ctx, db, input, entryID); err != nil {
			return impact, err
		}
		impact.EntryRecommendationID = &entryID
	}
	switch input.TargetType {
	case contentModel.GovernanceTargetDish:
		var dish contentModel.UserDish
		if err := db.WithContext(ctx).Unscoped().Where("public_id = ?", input.TargetID).First(&dish).Error; err != nil {
			return impact, mapTargetLookupError(err)
		}
		if dish.DeletedAt.Valid || dish.Version != input.ExpectedVersion {
			return impact, appErrors.AdminStateConflict.DefaultMsg()
		}
		impact.SourceDishCount = 1
		recommendationCount, err := countGovernedSourceRecommendations(
			ctx, db, "creator", dish.PublicID, &dish.ID, nil,
		)
		if err != nil {
			return impact, err
		}
		impact.RecommendationCount = int(recommendationCount)
		var copiedDishCount int64
		// 每次统计都创建全新的查询，避免 GORM 将某次追加的 id<> 条件带到后续统计。
		copyChainQuery := func() *gorm.DB {
			return db.WithContext(ctx).Model(&contentModel.UserDish{}).
				Where(
					"id = ? OR direct_source_dish_id = ? OR root_source_dish_id = ?",
					dish.ID,
					dish.ID,
					dish.ID,
				)
		}
		if err := copyChainQuery().
			Where("id <> ?", dish.ID).
			Count(&copiedDishCount).Error; err != nil {
			return impact, appErrors.AdminInternal.Wrap(err, "count governed dish copies")
		}
		impact.CopiedDishCount = int(copiedDishCount)
		for _, action := range input.Actions {
			switch action {
			case contentModel.GovernanceActionSoftDeleteDish:
				impact.Warnings = append(
					impact.Warnings,
					"将从现有菜谱和未确认饭局中停用源菜品；已确认饭局和采购历史快照不变。",
				)
			case contentModel.GovernanceActionDisableDiscoverability:
				impact.Warnings = append(
					impact.Warnings,
					"将关闭允许被发现，并同步下线该源菜品关联的全部在线推荐。",
				)
			case contentModel.GovernanceActionDeleteCopyChain:
				impact.Asynchronous = true
				var affectedUserCount int64
				if err := copyChainQuery().
					Distinct("owner_id").
					Count(&affectedUserCount).Error; err != nil {
					return impact, appErrors.AdminInternal.Wrap(err, "count governed copy chain users")
				}
				impact.AffectedUserCount = int(affectedUserCount)
				impact.NotificationCount = int(copiedDishCount) + 1
				var chainRecommendationCount int64
				if err := db.WithContext(ctx).Model(&contentModel.PlatformRecommendation{}).
					Where("status = ?", recommendationStatusPublished).
					Where(
						"dish_id IN (?) OR (source_type = ? AND source_dish_id IN (?))",
						copyChainQuery().Select("id"),
						"creator",
						copyChainQuery().Select("public_id"),
					).
					Count(&chainRecommendationCount).Error; err != nil {
					return impact, appErrors.AdminInternal.Wrap(err, "count governed copy chain recommendations")
				}
				impact.RecommendationCount = int(chainRecommendationCount)
				impact.Warnings = append(
					impact.Warnings,
					"将后台处理源菜品及全部复制菜品；未确认饭局仅保留不可选占位，历史饭局和采购快照不变。",
				)
			}
		}
	case contentModel.GovernanceTargetOfficialDish:
		var dish contentModel.OfficialDish
		if err := db.WithContext(ctx).Unscoped().Where("public_id = ?", input.TargetID).First(&dish).Error; err != nil {
			return impact, mapTargetLookupError(err)
		}
		if dish.DeletedAt.Valid || dish.Version != int64(input.ExpectedVersion) {
			return impact, appErrors.AdminStateConflict.DefaultMsg()
		}
		impact.SourceDishCount = 1
		impact.AffectedUserCount = 0
		impact.NotificationCount = 0
		recommendationCount, err := countGovernedSourceRecommendations(
			ctx, db, "official", dish.PublicID, nil, &dish.ID,
		)
		if err != nil {
			return impact, err
		}
		impact.RecommendationCount = int(recommendationCount)
		copyQuery := func() *gorm.DB {
			return db.WithContext(ctx).Model(&contentModel.UserDish{}).
				Where(
					"direct_source_official_dish_id = ? OR root_source_official_dish_id = ?",
					dish.ID,
					dish.ID,
				)
		}
		var copiedDishCount int64
		if err := copyQuery().Count(&copiedDishCount).Error; err != nil {
			return impact, appErrors.AdminInternal.Wrap(err, "count governed official dish copies")
		}
		impact.CopiedDishCount = int(copiedDishCount)
		if containsGovernanceAction(input.Actions, contentModel.GovernanceActionDeleteCopyChain) {
			impact.Asynchronous = copiedDishCount > 0
			var affectedUserCount int64
			if err := copyQuery().Distinct("owner_id").Count(&affectedUserCount).Error; err != nil {
				return impact, appErrors.AdminInternal.Wrap(err, "count governed official copy users")
			}
			impact.AffectedUserCount = int(affectedUserCount)
			impact.NotificationCount = int(copiedDishCount)
			if copiedDishCount > 0 {
				var copiedRecommendationCount int64
				if err := db.WithContext(ctx).Model(&contentModel.PlatformRecommendation{}).
					Where("status = ?", recommendationStatusPublished).
					Where(
						"dish_id IN (?) OR (source_type = ? AND source_dish_id IN (?))",
						copyQuery().Select("id"),
						"creator",
						copyQuery().Select("public_id"),
					).
					Count(&copiedRecommendationCount).Error; err != nil {
					return impact, appErrors.AdminInternal.Wrap(err, "count governed official copy recommendations")
				}
				impact.RecommendationCount += int(copiedRecommendationCount)
			}
			impact.Warnings = append(
				impact.Warnings,
				"将立即处理官方源菜品并下线相关推荐；用户复制菜品由后台任务继续处理，历史饭局和采购快照不变。",
			)
		} else {
			impact.Warnings = append(
				impact.Warnings,
				"将软删除官方源菜品并同步下线全部在线推荐；用户已经复制的菜品不受影响。",
			)
		}
	case contentModel.GovernanceTargetRecipe:
		var recipe contentModel.UserRecipe
		if err := db.WithContext(ctx).Unscoped().Where("public_id = ?", input.TargetID).First(&recipe).Error; err != nil {
			return impact, mapTargetLookupError(err)
		}
		if recipe.DeletedAt.Valid || recipe.Version != input.ExpectedVersion {
			return impact, appErrors.AdminStateConflict.DefaultMsg()
		}
		var dishCount int64
		if err := db.WithContext(ctx).Model(&contentModel.RecipeDish{}).
			Where("recipe_id = ?", recipe.ID).Count(&dishCount).Error; err != nil {
			return impact, appErrors.AdminInternal.Wrap(err, "count governed recipe dishes")
		}
		impact.SourceDishCount = int(dishCount)
		impact.Warnings = append(impact.Warnings, "软删除菜谱不会删除菜谱内菜品、关系审计或历史快照。")
	case contentModel.GovernanceTargetCheckin:
		var checkin contentModel.FrontCheckin
		if err := db.WithContext(ctx).Unscoped().
			Where("id = ?", input.TargetID).
			First(&checkin).Error; err != nil {
			return impact, mapTargetLookupError(err)
		}
		if checkin.DeletedAt.Valid || checkin.Version != input.ExpectedVersion {
			return impact, appErrors.AdminStateConflict.DefaultMsg()
		}
		impact.Warnings = append(
			impact.Warnings,
			"软删除打卡并从用户打卡记录中隐藏，但不撤销已经发放的积分，也不删除积分流水。",
		)
	default:
		return impact, appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	return impact, nil
}

// validateGovernanceEntryRecommendation 校验推荐入口与实际治理源头保持一致。
func validateGovernanceEntryRecommendation(
	ctx context.Context,
	db *gorm.DB,
	input contentReq.GovernanceActionPreviewInput,
	entryID string,
) error {
	if db == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	var recommendation contentModel.PlatformRecommendation
	if err := db.WithContext(ctx).Where("id = ?", entryID).First(&recommendation).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.AdminInvalidGovernance.DefaultMsg()
		}
		return appErrors.AdminInternal.Wrap(err, "load governance entry recommendation")
	}
	expectedSourceType := ""
	switch input.TargetType {
	case contentModel.GovernanceTargetDish:
		expectedSourceType = "creator"
	case contentModel.GovernanceTargetOfficialDish:
		expectedSourceType = "official"
	default:
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	if recommendation.SourceType != expectedSourceType ||
		recommendation.SourceDishID != input.TargetID {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	return nil
}

func (s *ContentService) sensitiveAccess(
	actor contentReq.AdminActor,
	permission string,
	action string,
	targetType string,
	targetID string,
) SensitiveAccess {
	return SensitiveAccess{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		AuthorityID:           actor.AuthorityID,
		Permission:            permission,
		Action:                action,
		TargetType:            targetType,
		TargetID:              targetID,
		RequestID:             actor.RequestID,
		SourceIPMasked:        actor.SourceIPMasked,
		UserAgentSummary:      actor.UserAgentSummary,
	}
}

type previewTokenClaims struct {
	RequestHash string `json:"requestHash"`
	ExpiresAt   int64  `json:"expiresAt"`
}

func (s *ContentService) signPreviewToken(requestHash string, expiresAt time.Time) (string, error) {
	claimsJSON, err := json.Marshal(previewTokenClaims{RequestHash: requestHash, ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	mac := hmac.New(sha256.New, s.effectivePreviewSecret())
	_, _ = mac.Write([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + signature, nil
}

func (s *ContentService) verifyPreviewToken(token string, expectedHash string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	mac := hmac.New(sha256.New, s.effectivePreviewSecret())
	_, _ = mac.Write([]byte(parts[0]))
	expectedSignature := mac.Sum(nil)
	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(actualSignature, expectedSignature) {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	var claims previewTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	if claims.RequestHash != expectedHash || s.now().Unix() > claims.ExpiresAt {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	return nil
}

func (s *ContentService) effectivePreviewSecret() []byte {
	if s != nil && len(s.previewSecret) > 0 {
		return s.previewSecret
	}
	signingKey := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	if signingKey == "" {
		return nil
	}
	sum := sha256.Sum256([]byte("orderfood-governance-preview:" + signingKey))
	return sum[:]
}

func governanceRequestHash(input contentReq.GovernanceActionPreviewInput) string {
	actions := append([]string(nil), input.Actions...)
	sort.Strings(actions)
	canonical := struct {
		TargetType            string   `json:"targetType"`
		TargetID              string   `json:"targetId"`
		EntryRecommendationID string   `json:"entryRecommendationId"`
		Actions               []string `json:"actions"`
		ViolationType         string   `json:"violationType"`
		Severity              string   `json:"severity"`
		Reason                string   `json:"reason"`
		ExpectedVersion       int      `json:"expectedVersion"`
	}{
		TargetType: input.TargetType, TargetID: input.TargetID, Actions: actions,
		EntryRecommendationID: strings.TrimSpace(input.EntryRecommendationID),
		ViolationType:         input.ViolationType, Severity: input.Severity,
		Reason: input.Reason, ExpectedVersion: input.ExpectedVersion,
	}
	encoded, _ := json.Marshal(canonical)
	sum := sha256.Sum256(encoded)
	return fmt.Sprintf("%x", sum[:])
}

func validateGovernanceInput(input contentReq.GovernanceActionPreviewInput) error {
	if input.TargetID == "" || input.ExpectedVersion < 1 ||
		(input.Severity != contentModel.GovernanceSeverityNormal && input.Severity != contentModel.GovernanceSeveritySerious) ||
		utf8.RuneCountInString(strings.TrimSpace(input.Reason)) < 4 ||
		utf8.RuneCountInString(input.Reason) > 300 ||
		utf8.RuneCountInString(input.ViolationType) == 0 ||
		utf8.RuneCountInString(input.ViolationType) > 40 ||
		utf8.RuneCountInString(strings.TrimSpace(input.EntryRecommendationID)) > 64 ||
		len(input.Actions) == 0 {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	seen := make(map[string]struct{}, len(input.Actions))
	hasCopyChainAction := false
	for _, action := range input.Actions {
		if _, exists := seen[action]; exists {
			return appErrors.AdminInvalidGovernance.DefaultMsg()
		}
		seen[action] = struct{}{}
		switch input.TargetType {
		case contentModel.GovernanceTargetDish:
			if action != contentModel.GovernanceActionDisableDiscoverability &&
				action != contentModel.GovernanceActionSoftDeleteDish &&
				action != contentModel.GovernanceActionDeleteCopyChain {
				return appErrors.AdminInvalidGovernance.DefaultMsg()
			}
			hasCopyChainAction = hasCopyChainAction ||
				action == contentModel.GovernanceActionDeleteCopyChain
		case contentModel.GovernanceTargetOfficialDish:
			if action != contentModel.GovernanceActionSoftDeleteOfficialDish &&
				action != contentModel.GovernanceActionDeleteCopyChain {
				return appErrors.AdminInvalidGovernance.DefaultMsg()
			}
			hasCopyChainAction = hasCopyChainAction ||
				action == contentModel.GovernanceActionDeleteCopyChain
		case contentModel.GovernanceTargetRecipe:
			if action != contentModel.GovernanceActionSoftDeleteRecipe {
				return appErrors.AdminInvalidGovernance.DefaultMsg()
			}
		case contentModel.GovernanceTargetCheckin:
			if action != contentModel.GovernanceActionSoftDeleteCheckin {
				return appErrors.AdminInvalidGovernance.DefaultMsg()
			}
		default:
			return appErrors.AdminInvalidGovernance.DefaultMsg()
		}
	}
	// 复制链处理是独立的严重违规动作，不能与普通治理动作混合提交。
	if hasCopyChainAction &&
		(len(input.Actions) != 1 || input.Severity != contentModel.GovernanceSeveritySerious) {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	if strings.TrimSpace(input.EntryRecommendationID) != "" &&
		input.TargetType != contentModel.GovernanceTargetDish &&
		input.TargetType != contentModel.GovernanceTargetOfficialDish {
		return appErrors.AdminInvalidGovernance.DefaultMsg()
	}
	return nil
}

// containsGovernanceAction 判断治理动作集合是否包含指定动作。
func containsGovernanceAction(actions []string, expected string) bool {
	for _, action := range actions {
		if action == expected {
			return true
		}
	}
	return false
}

func buildRecipeSummary(recipe *contentModel.UserRecipe, includePrivateCover bool) contentRes.UserRecipeAdminSummary {
	dishItems := buildRecipeDishItems(recipe)
	unavailableCount := 0
	for _, item := range dishItems {
		if item.DishStatus != contentModel.DishStatusUsable {
			unavailableCount++
		}
	}
	contentState := "ready"
	if len(dishItems) == 0 {
		contentState = "empty"
	} else if unavailableCount > 0 {
		contentState = "contains_unavailable"
	}
	var coverURL *string
	for _, relation := range recipe.RecipeDishes {
		firstDish := relation.Dish
		if firstDish.DeletedAt.Valid || firstDish.Status != contentModel.DishStatusUsable {
			continue
		}
		restricted := recipe.DeletedAt.Valid ||
			!firstDish.Discoverable
		if includePrivateCover || !restricted {
			coverURL = firstDish.CoverURL
		}
		break
	}
	var deletedAt *time.Time
	deletionStatus := "active"
	if recipe.DeletedAt.Valid {
		value := recipe.DeletedAt.Time
		deletedAt = &value
		deletionStatus = "deleted"
	}
	return contentRes.UserRecipeAdminSummary{
		ID: recipe.PublicID, Name: recipe.Name, CoverURL: coverURL,
		Owner: userReference(recipe.Owner), DishCount: len(dishItems), UnavailableDishCount: unavailableCount,
		HasNote:        recipe.Note != nil && strings.TrimSpace(*recipe.Note) != "",
		DeletionStatus: deletionStatus, ContentState: contentState, Version: recipe.Version,
		CreatedAt: recipe.CreatedAt, UpdatedAt: recipe.UpdatedAt, DeletedAt: deletedAt,
	}
}

func buildRecipeDishItems(recipe *contentModel.UserRecipe) []contentRes.UserRecipeDishItem {
	relations := append([]contentModel.RecipeDish(nil), recipe.RecipeDishes...)
	sort.SliceStable(relations, func(i, j int) bool {
		if relations[i].SortOrder == relations[j].SortOrder {
			return relations[i].ID < relations[j].ID
		}
		return relations[i].SortOrder < relations[j].SortOrder
	})
	items := make([]contentRes.UserRecipeDishItem, 0, len(relations))
	for _, relation := range relations {
		dishStatus := relation.Dish.Status
		var unavailableReason *string
		if relation.Dish.DeletedAt.Valid {
			dishStatus = "deleted"
			reason := "菜品已删除"
			unavailableReason = &reason
		} else if relation.Dish.Status != contentModel.DishStatusUsable {
			reason := "菜品为草稿"
			unavailableReason = &reason
		}
		items = append(items, contentRes.UserRecipeDishItem{
			RelationID: relation.PublicID, DishID: relation.Dish.PublicID, DishName: relation.Dish.Name,
			CoverURL: relation.Dish.CoverURL, DishStatus: dishStatus, SourceType: relation.Dish.SourceType,
			UnavailableReason: unavailableReason, SortOrder: relation.SortOrder, AddedAt: relation.AddedAt,
		})
	}
	return items
}

func loadDishSourceNode(ctx context.Context, db *gorm.DB, dishID uint) (*contentRes.UserDishSourceNode, error) {
	var dish contentModel.UserDish
	err := db.WithContext(ctx).Unscoped().Preload("Owner").First(&dish, dishID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "query dish source node")
	}
	owner := userReference(dish.Owner)
	deletionStatus := "active"
	if dish.DeletedAt.Valid {
		deletionStatus = "deleted"
	}
	return &contentRes.UserDishSourceNode{
		DishID: dish.PublicID, DishType: "user", Name: dish.Name, Owner: &owner,
		DeletionStatus: deletionStatus, Accessible: true,
	}, nil
}

// loadOfficialDishSourceNode 查询官方菜品来源节点。
func loadOfficialDishSourceNode(
	ctx context.Context,
	db *gorm.DB,
	dishID uint,
) (*contentRes.UserDishSourceNode, error) {
	var dish contentModel.OfficialDish
	err := db.WithContext(ctx).Unscoped().First(&dish, dishID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "query official dish source node")
	}
	deletionStatus := "active"
	if dish.DeletedAt.Valid {
		deletionStatus = "deleted"
	}
	return &contentRes.UserDishSourceNode{
		DishID: dish.PublicID, DishType: "official", Name: dish.Name,
		DeletionStatus: deletionStatus, Accessible: !dish.DeletedAt.Valid,
	}, nil
}

func userReference(user contentModel.MiniAppUser) contentRes.UserReference {
	return contentRes.UserReference{
		ID: user.ID, Nickname: user.Nickname, AvatarURL: user.AvatarURL, Status: string(user.Status),
	}
}

func deletedByAdministrator(
	id *uint,
	username *string,
	nickname *string,
) *contentRes.AdministratorSummary {
	if id == nil {
		return nil
	}
	value := ""
	if username != nil {
		value = *username
	}
	return &contentRes.AdministratorSummary{ID: strconv.FormatUint(uint64(*id), 10), Username: value, Nickname: nickname}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	switch pageSize {
	case 20, 50, 100:
	default:
		pageSize = 20
	}
	return page, pageSize
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func normalizeSortOrder(value string) (string, error) {
	switch strings.ToLower(value) {
	case "", "desc":
		return "DESC", nil
	case "asc":
		return "ASC", nil
	default:
		return "", appErrors.AdminBadRequest.DefaultMsg()
	}
}

func applyTimeRange(db *gorm.DB, column string, from, to *time.Time) *gorm.DB {
	if from != nil {
		db = db.Where(column+" >= ?", *from)
	}
	if to != nil {
		db = db.Where(column+" <= ?", *to)
	}
	return db
}

func validTimeRange(from, to *time.Time) bool {
	return from == nil || to == nil || !from.After(*to)
}

func validPagination(page, pageSize int) bool {
	if page < 0 {
		return false
	}
	switch pageSize {
	case 0, 20, 50, 100:
		return true
	default:
		return false
	}
}

func validSourceType(value string) bool {
	switch value {
	case contentModel.SourceTypeManual,
		contentModel.SourceTypeCreatorCopy,
		contentModel.SourceTypeOfficialCopy,
		contentModel.SourceTypeMealSuggestionCopy:
		return true
	default:
		return false
	}
}

func validMediaReviewStatus(value string) bool {
	switch value {
	case contentModel.MediaReviewPending,
		contentModel.MediaReviewPassed,
		contentModel.MediaReviewRejected,
		contentModel.MediaReviewFailed,
		contentModel.MediaReviewNotRequired:
		return true
	default:
		return false
	}
}

func validRecipeContentState(value string) bool {
	switch value {
	case "ready", "empty", "contains_unavailable":
		return true
	default:
		return false
	}
}

func validReferenceType(value string) bool {
	switch value {
	case contentModel.ReferenceTypeRecipe,
		contentModel.ReferenceTypeUnconfirmedMeal,
		contentModel.ReferenceTypeConfirmedMealSnapshot,
		contentModel.ReferenceTypeShoppingSnapshot:
		return true
	default:
		return false
	}
}

func validRecipeSort(value string) bool {
	switch value {
	case "createdAt", "updatedAt", "deletedAt", "name", "dishCount":
		return true
	default:
		return false
	}
}

func recipeSummaryLess(left, right contentRes.UserRecipeAdminSummary, sortBy string) bool {
	switch sortBy {
	case "updatedAt":
		return left.UpdatedAt.Before(right.UpdatedAt)
	case "deletedAt":
		if left.DeletedAt == nil {
			return right.DeletedAt != nil
		}
		if right.DeletedAt == nil {
			return false
		}
		return left.DeletedAt.Before(*right.DeletedAt)
	case "name":
		return left.Name < right.Name
	case "dishCount":
		return left.DishCount < right.DishCount
	default:
		return left.CreatedAt.Before(right.CreatedAt)
	}
}

func recipeSummaryEqual(left, right contentRes.UserRecipeAdminSummary, sortBy string) bool {
	switch sortBy {
	case "updatedAt":
		return left.UpdatedAt.Equal(right.UpdatedAt)
	case "deletedAt":
		if left.DeletedAt == nil || right.DeletedAt == nil {
			return left.DeletedAt == nil && right.DeletedAt == nil
		}
		return left.DeletedAt.Equal(*right.DeletedAt)
	case "name":
		return left.Name == right.Name
	case "dishCount":
		return left.DishCount == right.DishCount
	default:
		return left.CreatedAt.Equal(right.CreatedAt)
	}
}

func mapTargetLookupError(err error) error {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.AdminNotFound.DefaultMsg()
	}
	return appErrors.AdminInternal.Wrap(err, "query governance target")
}

func newPublicID(prefix string) string {
	return prefix + "_" + xid.New().String()
}

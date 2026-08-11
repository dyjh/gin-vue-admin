package dish

import (
	"context"
	"encoding/json"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	dishResponse "github.com/dyjh/order-food-mini-app/server/model/dish/response"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	contentService "github.com/dyjh/order-food-mini-app/server/service/content"
	"gorm.io/gorm"
)

const (
	// PermissionDiscoverableDishRead 表示读取可发现候选菜品所需的权限。
	PermissionDiscoverableDishRead = "orderfood:discoverable-dish:read"
	// PermissionRecommendationRead 表示读取推荐精选所需的权限。
	PermissionRecommendationRead = "orderfood:recommendation:read"
	// PermissionRecommendationCreate 表示创建推荐草稿所需的权限。
	PermissionRecommendationCreate = "orderfood:recommendation:create"
	// PermissionRecommendationUpdate 表示编辑推荐信息所需的权限。
	PermissionRecommendationUpdate = "orderfood:recommendation:update"
	// PermissionRecommendationDelete 表示删除推荐草稿所需的权限。
	PermissionRecommendationDelete = "orderfood:recommendation:delete"
	// PermissionRecommendationPublish 表示发布推荐所需的权限。
	PermissionRecommendationPublish = "orderfood:recommendation:publish"
	// PermissionRecommendationOffline 表示下线推荐所需的权限。
	PermissionRecommendationOffline = "orderfood:recommendation:offline"
	// PermissionRecommendationSort 表示调整推荐排序所需的权限。
	PermissionRecommendationSort = "orderfood:recommendation:sort"

	recommendationSourceCreator  = "creator"
	recommendationSourceOfficial = "official"
	recommendationStatusDraft    = "draft"
	recommendationHomeCapacity   = int64(5)
)

// RecommendationService 提供可发现候选菜品和推荐精选管理能力。
type RecommendationService struct {
	DB            *gorm.DB                          // 数据库连接
	Idempotency   *serviceCommon.IdempotencyService // 幂等处理服务
	MutationAudit serviceCommon.MutationAuditWriter // 变更审计写入器
	Now           func() time.Time                  // 当前时间函数
}

// NewRecommendationService 创建推荐管理服务实例。
func NewRecommendationService(
	db *gorm.DB,
	idempotency *serviceCommon.IdempotencyService,
) *RecommendationService {
	if idempotency == nil {
		idempotency = serviceCommon.NewIdempotencyService(db)
	}
	return &RecommendationService{
		DB:            db,
		Idempotency:   idempotency,
		MutationAudit: serviceCommon.GormMutationAuditWriter{},
		Now:           time.Now,
	}
}

// database 返回推荐服务使用的数据库连接。
func (service *RecommendationService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回统一为UTC的当前时间。
func (service *RecommendationService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// validateRecommendationActor 校验推荐管理操作上下文。
func validateRecommendationActor(actor commonRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// recommendationAdministrator 将持久化的管理员信息转换为响应摘要。
func recommendationAdministrator(
	id uint,
	username string,
	nickname *string,
) commonResponse.AdministratorSummary {
	return commonResponse.AdministratorSummary{
		ID: strconv.FormatUint(uint64(id), 10), Username: username, Nickname: nickname,
	}
}

// writeMutationAudit 在推荐变更事务内写入审计记录。
func (service *RecommendationService) writeMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor commonRequest.AdminActor,
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
	beforeJSON, err := serviceCommon.SafeAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := serviceCommon.SafeAuditJSON(after)
	if err != nil {
		return err
	}
	key := strings.TrimSpace(idempotencyKey)
	var reasonPointer *string
	if value := serviceCommon.TruncateRunes(strings.TrimSpace(reason), 200); value != "" {
		reasonPointer = &value
	}
	audit := auditModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            "recommendation",
		TargetID:              targetID,
		Reason:                reasonPointer,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             serviceCommon.TruncateRunes(strings.TrimSpace(actor.RequestID), 96),
		IdempotencyKey:        &key,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := serviceCommon.TruncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

// discoverableListRow 表示可发现菜品列表查询的聚合字段。
type discoverableListRow struct {
	ID                     uint    `gorm:"column:id"`                      // 菜品内部ID
	VoteCount              int64   `gorm:"column:vote_count"`              // 饭局点选次数
	CopyCount              int64   `gorm:"column:copy_count"`              // 被复制次数
	RecommendationID       *string `gorm:"column:recommendation_id"`       // 推荐ID
	RecommendationPosition *string `gorm:"column:recommendation_position"` // 推荐位置
}

// preloadDiscoverableDish 查询可发现菜品响应所需的全部关联数据。
func preloadDiscoverableDish(db *gorm.DB, dishID uint) (contentModel.UserDish, error) {
	var dish contentModel.UserDish
	err := db.
		Preload("Owner").
		Preload("Category").
		Preload("Tags").
		Preload("Ingredients", func(statement *gorm.DB) *gorm.DB {
			return statement.Order("sort_order ASC, id ASC")
		}).
		Preload("Ingredients.Unit").
		Preload("Steps", func(statement *gorm.DB) *gorm.DB {
			return statement.Order("sort_order ASC, id ASC")
		}).
		First(&dish, dishID).Error
	return dish, err
}

// discoverableDishCounts 统计可发现菜品的点选、复制和推荐关联。
func (service *RecommendationService) discoverableDishCounts(
	ctx context.Context,
	db *gorm.DB,
	dish contentModel.UserDish,
) (discoverableListRow, error) {
	row := discoverableListRow{ID: dish.ID}
	if err := db.WithContext(ctx).
		Table((mealModel.FrontMealVote{}).TableName()+" AS votes").
		Joins(
			"JOIN "+(mealModel.FrontMealCandidate{}).TableName()+
				" AS candidates ON candidates.id = votes.candidate_id",
		).
		Where("candidates.dish_id = ?", dish.ID).
		Count(&row.VoteCount).Error; err != nil {
		return row, appErrors.AdminInternal.Wrap(err, "count discoverable dish votes")
	}
	if err := db.WithContext(ctx).
		Unscoped().
		Model(&contentModel.UserDish{}).
		Where("direct_source_dish_id = ? AND deleted_at IS NULL", dish.ID).
		Count(&row.CopyCount).Error; err != nil {
		return row, appErrors.AdminInternal.Wrap(err, "count discoverable dish copies")
	}
	var recommendation dishModel.PlatformRecommendation
	if err := db.WithContext(ctx).
		Where("source_type = ? AND source_dish_id = ?", recommendationSourceCreator, dish.PublicID).
		Order("created_at desc").
		First(&recommendation).Error; err == nil {
		row.RecommendationID = &recommendation.ID
		row.RecommendationPosition = &recommendation.Position
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return row, appErrors.AdminInternal.Wrap(err, "load discoverable dish recommendation")
	}
	return row, nil
}

// discoverableDishSummary 将菜品及聚合数据转换为候选菜品摘要。
func discoverableDishSummary(
	dish contentModel.UserDish,
	counts discoverableListRow,
) dishResponse.DiscoverableDishSummary {
	tags := make([]commonResponse.CatalogReference, 0, len(dish.Tags))
	for _, tag := range dish.Tags {
		tags = append(tags, commonResponse.CatalogReference{
			ID: tag.PublicID, Name: tag.Name, Enabled: tag.Enabled,
		})
	}
	discoverableAt := dish.CreatedAt
	if dish.DiscoverableAt != nil {
		discoverableAt = *dish.DiscoverableAt
	}
	coverURL := ""
	if dish.CoverURL != nil {
		coverURL = *dish.CoverURL
	}
	return dishResponse.DiscoverableDishSummary{
		ID:       dish.PublicID,
		CoverURL: coverURL,
		Name:     dish.Name,
		Category: commonResponse.CatalogReference{
			ID: dish.Category.PublicID, Name: dish.Category.Name, Enabled: dish.Category.Enabled,
		},
		Tags:                   tags,
		Author:                 serviceCommon.UserReference(dish.Owner),
		DiscoverableAt:         discoverableAt,
		VoteCount:              counts.VoteCount,
		CopyCount:              counts.CopyCount,
		Selected:               counts.RecommendationID != nil,
		RecommendationPosition: counts.RecommendationPosition,
		SourceLocked:           dish.SourceLocked,
		Version:                dish.Version,
	}
}

// discoverableDishDetail 将完整菜品模型转换为候选菜品详情。
func (service *RecommendationService) discoverableDishDetail(
	ctx context.Context,
	db *gorm.DB,
	dish contentModel.UserDish,
	includeModerationSummary bool,
) (dishResponse.DiscoverableDishDetail, error) {
	counts, err := service.discoverableDishCounts(ctx, db, dish)
	if err != nil {
		return dishResponse.DiscoverableDishDetail{}, err
	}
	ingredients := make([]contentResponse.DishIngredient, 0, len(dish.Ingredients))
	for _, ingredient := range dish.Ingredients {
		var unit *commonResponse.CatalogReference
		if ingredient.UnitID != nil {
			unit = &commonResponse.CatalogReference{
				ID: ingredient.Unit.PublicID, Name: ingredient.Unit.Name, Enabled: ingredient.Unit.Enabled,
			}
		}
		ingredients = append(ingredients, contentResponse.DishIngredient{
			Name: ingredient.Name, Quantity: ingredient.Quantity, Unit: unit,
			Note: ingredient.Note, SortOrder: ingredient.SortOrder,
		})
	}
	steps := make([]contentResponse.DishStep, 0, len(dish.Steps))
	for _, step := range dish.Steps {
		steps = append(steps, contentResponse.DishStep{
			Description: step.Description, ImageURL: step.ImageURL, SortOrder: step.SortOrder,
		})
	}
	var recommendation *dishResponse.RecommendationSummary
	if counts.RecommendationID != nil {
		var row dishModel.PlatformRecommendation
		if err := db.WithContext(ctx).First(&row, "id = ?", *counts.RecommendationID).Error; err != nil {
			return dishResponse.DiscoverableDishDetail{}, appErrors.AdminInternal.Wrap(err, "load recommendation summary")
		}
		value, err := service.recommendationSummary(ctx, db, row)
		if err != nil {
			return dishResponse.DiscoverableDishDetail{}, err
		}
		recommendation = &value
	}
	var moderationSummary *contentResponse.ModerationRecordSummary
	if includeModerationSummary && strings.TrimSpace(dish.CoverFileID) != "" {
		var record contentModel.ImageModerationRecord
		if err := db.WithContext(ctx).
			Where("file_id = ?", dish.CoverFileID).
			Order("created_at DESC, id DESC").
			First(&record).Error; err == nil {
			value, summaryErr := contentService.ModerationRecordSummary(ctx, db, record)
			if summaryErr != nil {
				return dishResponse.DiscoverableDishDetail{}, summaryErr
			}
			moderationSummary = &value
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return dishResponse.DiscoverableDishDetail{}, appErrors.AdminInternal.Wrap(
				err, "load discoverable dish moderation summary",
			)
		}
	}
	return dishResponse.DiscoverableDishDetail{
		DiscoverableDishSummary: discoverableDishSummary(dish, counts),
		Description:             dish.Description,
		Serving:                 dish.Serving,
		Ingredients:             ingredients,
		Steps:                   steps,
		MediaReviewStatus:       dish.MediaReviewStatus,
		Recommendation:          recommendation,
		ModerationSummary:       moderationSummary,
		UpdatedAt:               dish.UpdatedAt,
	}, nil
}

// ListDiscoverableDishes 分页查询用户主动允许被发现的候选菜品。
func (service *RecommendationService) ListDiscoverableDishes(
	ctx context.Context,
	query dishRequest.DiscoverableDishListQuery,
) (commonResponse.Page[dishResponse.DiscoverableDishSummary], error) {
	query.ApplyDefaults()
	if !serviceCommon.ValidPagination(query.Page, query.PageSize) ||
		!serviceCommon.ValidTimeRange(query.DiscoverableFrom, query.DiscoverableTo) ||
		utf8.RuneCountInString(query.Keyword) > 60 ||
		utf8.RuneCountInString(query.AuthorKeyword) > 60 ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	base := db.WithContext(ctx).
		Model(&contentModel.UserDish{}).
		Where("discoverable = ? AND status = ?", true, contentModel.DishStatusUsable)
	if value := strings.TrimSpace(query.DishID); value != "" {
		base = base.Where("public_id = ?", value)
	}
	if value := strings.TrimSpace(query.Keyword); value != "" {
		base = base.Where("name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(query.AuthorID); value != "" {
		base = base.Where("owner_id = ?", value)
	}
	if value := strings.TrimSpace(query.AuthorKeyword); value != "" {
		keyword := "%" + value + "%"
		ownerIDs := db.Model(&userModel.MiniAppUser{}).
			Select("id").
			Where("id LIKE ? OR nickname LIKE ?", keyword, keyword)
		base = base.Where("owner_id IN (?)", ownerIDs)
	}
	if value := strings.TrimSpace(query.CategoryID); value != "" {
		base = base.Where("category_id IN (?)",
			db.Model(&dishModel.ContentCategory{}).Select("id").Where("public_id = ?", value),
		)
	}
	if len(query.TagIDs) > 0 {
		base = base.Where("id IN (?)",
			db.Table((contentModel.UserDishTag{}).TableName()+" AS dish_tags").
				Select("dish_tags.dish_id").
				Joins("JOIN "+(dishModel.ContentTag{}).TableName()+" AS tags ON tags.id = dish_tags.tag_id").
				Where("tags.public_id IN ?", query.TagIDs).
				Group("dish_tags.dish_id").
				Having("COUNT(DISTINCT tags.public_id) = ?", len(query.TagIDs)),
		)
	}
	if query.DiscoverableFrom != nil {
		base = base.Where("discoverable_at >= ?", *query.DiscoverableFrom)
	}
	if query.DiscoverableTo != nil {
		base = base.Where("discoverable_at <= ?", *query.DiscoverableTo)
	}
	if query.Selected != nil {
		exists := db.Model(&dishModel.PlatformRecommendation{}).
			Select("1").
			Where(
				"source_type = ? AND source_dish_id = of_dishes.public_id",
				recommendationSourceCreator,
			)
		if *query.Selected {
			base = base.Where("EXISTS (?)", exists)
		} else {
			base = base.Where("NOT EXISTS (?)", exists)
		}
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminInternal.Wrap(err, "count discoverable dishes")
	}
	sortColumns := map[string]string{
		"discoverableAt": "discoverable_at",
		"createdAt":      "created_at",
	}
	ids := make([]uint, 0)
	switch query.SortBy {
	case "voteCount", "copyCount":
		aggregate := base.Select(
			"of_dishes.id, " +
				"(SELECT COUNT(*) FROM of_meal_votes votes JOIN of_meal_candidates candidates ON candidates.id = votes.candidate_id WHERE candidates.dish_id = of_dishes.id) AS vote_count, " +
				"(SELECT COUNT(*) FROM of_dishes copies WHERE copies.direct_source_dish_id = of_dishes.id AND copies.deleted_at IS NULL) AS copy_count",
		)
		rows := make([]discoverableListRow, 0)
		column := "vote_count"
		if query.SortBy == "copyCount" {
			column = "copy_count"
		}
		if err := aggregate.
			Order(column + " " + query.SortOrder).
			Order("of_dishes.id asc").
			Limit(query.PageSize).
			Offset((query.Page - 1) * query.PageSize).
			Scan(&rows).Error; err != nil {
			return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminInternal.Wrap(err, "list discoverable dish ids")
		}
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
	default:
		column, ok := sortColumns[query.SortBy]
		if !ok {
			return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		if err := base.
			Order(column+" "+query.SortOrder).
			Order("id asc").
			Limit(query.PageSize).
			Offset((query.Page-1)*query.PageSize).
			Pluck("id", &ids).Error; err != nil {
			return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminInternal.Wrap(err, "list discoverable dish ids")
		}
	}
	list := make([]dishResponse.DiscoverableDishSummary, 0, len(ids))
	for _, dishID := range ids {
		dish, err := preloadDiscoverableDish(db.WithContext(ctx), dishID)
		if err != nil {
			return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, appErrors.AdminInternal.Wrap(err, "load discoverable dish")
		}
		counts, err := service.discoverableDishCounts(ctx, db, dish)
		if err != nil {
			return commonResponse.Page[dishResponse.DiscoverableDishSummary]{}, err
		}
		list = append(list, discoverableDishSummary(dish, counts))
	}
	return commonResponse.Page[dishResponse.DiscoverableDishSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetDiscoverableDish 获取用户主动允许被发现的候选菜品详情。
func (service *RecommendationService) GetDiscoverableDish(
	ctx context.Context,
	dishID string,
	includeModerationSummary bool,
) (dishResponse.DiscoverableDishDetail, error) {
	db := service.database()
	if db == nil || strings.TrimSpace(dishID) == "" {
		return dishResponse.DiscoverableDishDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var row contentModel.UserDish
	if err := db.WithContext(ctx).
		Where("public_id = ? AND discoverable = ? AND status = ?",
			strings.TrimSpace(dishID), true, contentModel.DishStatusUsable,
		).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dishResponse.DiscoverableDishDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return dishResponse.DiscoverableDishDetail{}, appErrors.AdminInternal.Wrap(err, "load discoverable dish")
	}
	dish, err := preloadDiscoverableDish(db.WithContext(ctx), row.ID)
	if err != nil {
		return dishResponse.DiscoverableDishDetail{}, appErrors.AdminInternal.Wrap(err, "load discoverable dish detail")
	}
	result, err := service.discoverableDishDetail(ctx, db, dish, includeModerationSummary)
	if err != nil {
		return result, err
	}
	// 图片审核摘要的条件权限已由API层控制；媒体记录详情将在图片资源模块统一补齐。
	if !includeModerationSummary {
		result.ModerationSummary = nil
	}
	return result, nil
}

// recommendationSource 描述推荐来源菜品的当前状态。
type recommendationSource struct {
	Type              string                  // 来源类型
	PublicID          string                  // 来源菜品公开ID
	CreatorDish       *contentModel.UserDish  // 用户来源菜品
	OfficialDish      *dishModel.OfficialDish // 官方来源菜品
	Name              string                  // 菜品名称
	CoverURL          string                  // 封面地址
	AuthorLabel       string                  // 来源作者说明
	Available         bool                    // 是否可发布
	UnavailableReason *string                 // 不可用原因
}

// unavailableReason 返回不可用原因指针。
func unavailableReason(value string) *string {
	return &value
}

// loadRecommendationSource 查询推荐来源及其当前可用状态。
func (service *RecommendationService) loadRecommendationSource(
	ctx context.Context,
	db *gorm.DB,
	sourceType string,
	sourceDishID string,
) (recommendationSource, error) {
	sourceDishID = strings.TrimSpace(sourceDishID)
	source := recommendationSource{Type: sourceType, PublicID: sourceDishID}
	statement := db.WithContext(ctx)
	switch sourceType {
	case recommendationSourceCreator:
		var dish contentModel.UserDish
		if err := statement.Unscoped().
			Preload("Owner").
			First(&dish, "public_id = ?", sourceDishID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return source, appErrors.AdminNotFound.DefaultMsg()
			}
			return source, appErrors.AdminInternal.Wrap(err, "load creator recommendation source")
		}
		source.CreatorDish = &dish
		source.Name = dish.Name
		source.AuthorLabel = dish.Owner.Nickname
		if dish.CoverURL != nil {
			source.CoverURL = *dish.CoverURL
		}
		switch {
		case dish.DeletedAt.Valid:
			source.UnavailableReason = unavailableReason("来源菜品已删除")
		case !dish.Discoverable:
			source.UnavailableReason = unavailableReason("用户已关闭允许被发现")
		case dish.SourceLocked:
			source.UnavailableReason = unavailableReason("来源为锁定副本")
		case dish.Status != contentModel.DishStatusUsable:
			source.UnavailableReason = unavailableReason("来源菜品不可用")
		case strings.TrimSpace(source.CoverURL) == "":
			source.UnavailableReason = unavailableReason("来源封面不可用")
		case !contentModel.IsMediaReviewUsable(dish.MediaReviewStatus):
			source.UnavailableReason = unavailableReason("来源封面未通过审核")
		default:
			source.Available = true
		}
	case recommendationSourceOfficial:
		var dish dishModel.OfficialDish
		if err := statement.Unscoped().
			First(&dish, "public_id = ?", sourceDishID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return source, appErrors.AdminNotFound.DefaultMsg()
			}
			return source, appErrors.AdminInternal.Wrap(err, "load official recommendation source")
		}
		source.OfficialDish = &dish
		source.Name = dish.Name
		source.CoverURL = dish.CoverURL
		source.AuthorLabel = "平台官方"
		switch {
		case dish.DeletedAt.Valid:
			source.UnavailableReason = unavailableReason("官方菜品已删除")
		case dish.Status != contentModel.DishStatusUsable:
			source.UnavailableReason = unavailableReason("官方菜品不可用")
		case strings.TrimSpace(dish.CoverURL) == "":
			source.UnavailableReason = unavailableReason("官方菜品封面不可用")
		default:
			source.Available = true
		}
	default:
		return source, appErrors.AdminBadRequest.DefaultMsg()
	}
	return source, nil
}

// recommendationSummary 将推荐模型转换为列表摘要。
func (service *RecommendationService) recommendationSummary(
	ctx context.Context,
	db *gorm.DB,
	row dishModel.PlatformRecommendation,
) (dishResponse.RecommendationSummary, error) {
	source, err := service.loadRecommendationSource(ctx, db, row.SourceType, row.SourceDishID)
	if err != nil {
		if appErrors.GetType(err) != appErrors.AdminNotFound {
			return dishResponse.RecommendationSummary{}, err
		}
		source.Name = "来源已不存在"
		source.AuthorLabel = "—"
	}
	var copyCount int64
	if err := db.WithContext(ctx).
		Model(&dishModel.RecommendationCopy{}).
		Where("recommendation_id = ?", row.ID).
		Count(&copyCount).Error; err != nil {
		return dishResponse.RecommendationSummary{}, appErrors.AdminInternal.Wrap(err, "count recommendation copies")
	}
	return dishResponse.RecommendationSummary{
		ID:                row.ID,
		SourceType:        row.SourceType,
		SourceDishID:      row.SourceDishID,
		DishName:          source.Name,
		CoverURL:          source.CoverURL,
		SourceAuthorLabel: source.AuthorLabel,
		Position:          row.Position,
		SortOrder:         row.SortOrder,
		Status:            row.Status,
		CopyCount:         copyCount,
		Version:           row.Version,
		PublishedAt:       row.PublishedAt,
		OfflineAt:         row.OfflineAt,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}

// recommendationDetail 将推荐模型及来源转换为详情。
func (service *RecommendationService) recommendationDetail(
	ctx context.Context,
	db *gorm.DB,
	row dishModel.PlatformRecommendation,
) (dishResponse.RecommendationDetail, error) {
	summary, err := service.recommendationSummary(ctx, db, row)
	if err != nil {
		return dishResponse.RecommendationDetail{}, err
	}
	source, sourceErr := service.loadRecommendationSource(
		ctx, db, row.SourceType, row.SourceDishID,
	)
	var dish interface{}
	available := false
	var unavailable *string
	if sourceErr != nil {
		if appErrors.GetType(sourceErr) != appErrors.AdminNotFound {
			return dishResponse.RecommendationDetail{}, sourceErr
		}
		unavailable = unavailableReason("来源菜品已不存在")
	} else {
		available = source.Available
		unavailable = source.UnavailableReason
		if source.CreatorDish != nil {
			full, err := preloadDiscoverableDish(db.WithContext(ctx), source.CreatorDish.ID)
			if err != nil {
				return dishResponse.RecommendationDetail{}, appErrors.AdminInternal.Wrap(err, "load creator recommendation dish")
			}
			value, err := service.discoverableDishDetail(ctx, db, full, false)
			if err != nil {
				return dishResponse.RecommendationDetail{}, err
			}
			dish = value
		} else if source.OfficialDish != nil {
			full, err := preloadOfficialDish(db.WithContext(ctx), source.OfficialDish.ID)
			if err != nil {
				return dishResponse.RecommendationDetail{}, appErrors.AdminInternal.Wrap(
					err, "load official recommendation dish",
				)
			}
			copyCount, err := officialDishCopyCount(ctx, db, full.ID)
			if err != nil {
				return dishResponse.RecommendationDetail{}, err
			}
			dish = officialDishDetail(full, copyCount)
		}
	}
	return dishResponse.RecommendationDetail{
		RecommendationSummary:   summary,
		DisplayNote:             row.DisplayNote,
		SourceAvailable:         available,
		SourceUnavailableReason: unavailable,
		Dish:                    dish,
		CreatedBy: recommendationAdministrator(
			row.CreatedByID, row.CreatedByUsername, row.CreatedByNickname,
		),
		UpdatedBy: recommendationAdministrator(
			row.UpdatedByID, row.UpdatedByUsername, row.UpdatedByNickname,
		),
	}, nil
}

// ListRecommendations 分页查询推荐精选记录。
func (service *RecommendationService) ListRecommendations(
	ctx context.Context,
	query dishRequest.RecommendationListQuery,
) (commonResponse.Page[dishResponse.RecommendationSummary], error) {
	query.ApplyDefaults()
	db := service.database()
	if db == nil {
		return commonResponse.Page[dishResponse.RecommendationSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&dishModel.PlatformRecommendation{})
	if query.SourceType != "" {
		statement = statement.Where("source_type = ?", query.SourceType)
	}
	if value := strings.TrimSpace(query.SourceDishID); value != "" {
		statement = statement.Where("source_dish_id = ?", value)
	}
	if query.Status != "" {
		statement = statement.Where("status = ?", query.Status)
	}
	if query.Position != "" {
		statement = statement.Where("position = ?", query.Position)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", *query.CreatedTo)
	}
	if value := strings.TrimSpace(query.Keyword); value != "" {
		pattern := "%" + value + "%"
		statement = statement.Where(
			"(source_type = ? AND dish_id IN (?)) OR (source_type = ? AND official_dish_id IN (?))",
			recommendationSourceCreator,
			db.Unscoped().Model(&contentModel.UserDish{}).Select("id").Where("name LIKE ?", pattern),
			recommendationSourceOfficial,
			db.Unscoped().Model(&dishModel.OfficialDish{}).Select("id").Where("name LIKE ?", pattern),
		)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[dishResponse.RecommendationSummary]{}, appErrors.AdminInternal.Wrap(err, "count recommendations")
	}
	sortColumns := map[string]string{
		"sortOrder":   "sort_order",
		"publishedAt": "published_at",
		"createdAt":   "created_at",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[dishResponse.RecommendationSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	rows := make([]dishModel.PlatformRecommendation, 0)
	if err := statement.
		Order(column + " " + query.SortOrder).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[dishResponse.RecommendationSummary]{}, appErrors.AdminInternal.Wrap(err, "list recommendations")
	}
	list := make([]dishResponse.RecommendationSummary, 0, len(rows))
	for _, row := range rows {
		summary, err := service.recommendationSummary(ctx, db, row)
		if err != nil {
			return commonResponse.Page[dishResponse.RecommendationSummary]{}, err
		}
		list = append(list, summary)
	}
	return commonResponse.Page[dishResponse.RecommendationSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetRecommendation 获取推荐精选详情。
func (service *RecommendationService) GetRecommendation(
	ctx context.Context,
	recommendationID string,
) (dishResponse.RecommendationDetail, error) {
	db := service.database()
	if db == nil || strings.TrimSpace(recommendationID) == "" {
		return dishResponse.RecommendationDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var row dishModel.PlatformRecommendation
	if err := db.WithContext(ctx).
		First(&row, "id = ?", strings.TrimSpace(recommendationID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dishResponse.RecommendationDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return dishResponse.RecommendationDetail{}, appErrors.AdminInternal.Wrap(err, "load recommendation")
	}
	return service.recommendationDetail(ctx, db, row)
}

// updateRecommendationCount 调整来源菜品的推荐记录计数。
func updateRecommendationCount(
	tx *gorm.DB,
	source recommendationSource,
	delta int,
) error {
	var model interface{}
	var id uint
	if source.CreatorDish != nil {
		model = &contentModel.UserDish{}
		id = source.CreatorDish.ID
	} else if source.OfficialDish != nil {
		model = &dishModel.OfficialDish{}
		id = source.OfficialDish.ID
	} else {
		return appErrors.AdminInternal.DefaultMsg()
	}
	expression := "recommendation_count + ?"
	if delta < 0 {
		expression = "GREATEST(recommendation_count + ?, 0)"
	}
	update := tx.Unscoped().Model(model).Where("id = ?", id).
		UpdateColumn("recommendation_count", gorm.Expr(expression, delta))
	if update.Error != nil {
		return appErrors.AdminInternal.Wrap(update.Error, "update recommendation count")
	}
	if update.RowsAffected != 1 {
		return appErrors.AdminStateConflict.DefaultMsg()
	}
	return nil
}

// normalizeDisplayNote 规范化可选展示说明。
func normalizeDisplayNote(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > 160 {
		return nil, appErrors.AdminBadRequest.DefaultMsg()
	}
	return &trimmed, nil
}

// CreateRecommendation 创建推荐草稿。
func (service *RecommendationService) CreateRecommendation(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.RecommendationCreateInput,
) (dishResponse.RecommendationDetail, bool, error) {
	if err := validateRecommendationActor(actor); err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	input.SourceDishID = strings.TrimSpace(input.SourceDishID)
	note, err := normalizeDisplayNote(input.DisplayNote)
	if err != nil || input.SourceDishID == "" || input.SortOrder < 1 ||
		input.Position != serviceCommon.RecommendationPositionHome ||
		(input.SourceType != recommendationSourceCreator && input.SourceType != recommendationSourceOfficial) {
		return dishResponse.RecommendationDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	input.DisplayNote = note
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"recommendation_create",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			var existing int64
			if err := tx.Model(&dishModel.PlatformRecommendation{}).
				Where("source_type = ? AND source_dish_id = ?", input.SourceType, input.SourceDishID).
				Count(&existing).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check recommendation source")
			}
			if existing > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			source, err := service.loadRecommendationSource(
				ctx, tx, input.SourceType, input.SourceDishID,
			)
			if err != nil {
				return nil, err
			}
			if !source.Available {
				return nil, appErrors.AdminInvalidImage.DefaultMsg()
			}
			now := service.now()
			row := dishModel.PlatformRecommendation{
				ID:                serviceCommon.NewPublicID("rec"),
				SourceType:        input.SourceType,
				SourceDishID:      input.SourceDishID,
				Position:          input.Position,
				DisplayNote:       input.DisplayNote,
				Status:            recommendationStatusDraft,
				Selected:          false,
				SortOrder:         input.SortOrder,
				Version:           1,
				CreatedByID:       actor.AdministratorID,
				CreatedByUsername: actor.Username,
				CreatedByNickname: actor.Nickname,
				UpdatedByID:       actor.AdministratorID,
				UpdatedByUsername: actor.Username,
				UpdatedByNickname: actor.Nickname,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			if source.CreatorDish != nil {
				row.DishID = &source.CreatorDish.ID
			} else {
				row.OfficialDishID = &source.OfficialDish.ID
			}
			if err := tx.Create(&row).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create recommendation")
			}
			if err := updateRecommendationCount(tx, source, 1); err != nil {
				return nil, err
			}
			result, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "recommendation_create", row.ID, "",
				idempotencyKey, nil, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	var result dishResponse.RecommendationDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode recommendation create result")
	}
	return result, replayed, nil
}

// loadRecommendation 查询推荐记录。
func loadRecommendation(
	ctx context.Context,
	tx *gorm.DB,
	recommendationID string,
) (dishModel.PlatformRecommendation, error) {
	var row dishModel.PlatformRecommendation
	if err := tx.WithContext(ctx).
		First(&row, "id = ?", strings.TrimSpace(recommendationID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return row, appErrors.AdminNotFound.DefaultMsg()
		}
		return row, appErrors.AdminInternal.Wrap(err, "load recommendation for mutation")
	}
	return row, nil
}

// UpdateRecommendation 编辑推荐排序和展示说明。
func (service *RecommendationService) UpdateRecommendation(
	ctx context.Context,
	recommendationID string,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.RecommendationUpdateInput,
) (dishResponse.RecommendationDetail, bool, error) {
	if err := validateRecommendationActor(actor); err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	recommendationID = strings.TrimSpace(recommendationID)
	note, err := normalizeDisplayNote(input.DisplayNote)
	if err != nil || recommendationID == "" || input.SortOrder < 1 ||
		input.ExpectedVersion < 1 || input.Position != serviceCommon.RecommendationPositionHome {
		return dishResponse.RecommendationDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	input.DisplayNote = note
	payload := struct {
		ID    string                                `json:"id"`    // 推荐ID
		Input dishRequest.RecommendationUpdateInput `json:"input"` // 编辑参数
	}{ID: recommendationID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "recommendation_update",
		strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			row, err := loadRecommendation(ctx, tx, recommendationID)
			if err != nil {
				return nil, err
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			now := service.now()
			update := tx.Model(&row).
				Where("version = ?", input.ExpectedVersion).
				Updates(map[string]interface{}{
					"position":            input.Position,
					"sort_order":          input.SortOrder,
					"display_note":        input.DisplayNote,
					"version":             gorm.Expr("version + 1"),
					"updated_by_id":       actor.AdministratorID,
					"updated_by_username": actor.Username,
					"updated_by_nickname": actor.Nickname,
					"updated_at":          now,
				})
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "update recommendation")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			row.Position = input.Position
			row.SortOrder = input.SortOrder
			row.DisplayNote = input.DisplayNote
			row.Version++
			row.UpdatedByID = actor.AdministratorID
			row.UpdatedByUsername = actor.Username
			row.UpdatedByNickname = actor.Nickname
			row.UpdatedAt = now
			result, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "recommendation_update", row.ID, "",
				idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	var result dishResponse.RecommendationDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode recommendation update result")
	}
	return result, replayed, nil
}

// DeleteRecommendation 删除仍处于草稿状态的推荐。
func (service *RecommendationService) DeleteRecommendation(
	ctx context.Context,
	recommendationID string,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.RecommendationVersionInput,
) (dishResponse.RecommendationDeleteResult, bool, error) {
	if err := validateRecommendationActor(actor); err != nil {
		return dishResponse.RecommendationDeleteResult{}, false, err
	}
	recommendationID = strings.TrimSpace(recommendationID)
	if recommendationID == "" || input.ExpectedVersion < 1 {
		return dishResponse.RecommendationDeleteResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		ID    string                                 `json:"id"`    // 推荐ID
		Input dishRequest.RecommendationVersionInput `json:"input"` // 版本参数
	}{ID: recommendationID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "recommendation_delete",
		strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			row, err := loadRecommendation(ctx, tx, recommendationID)
			if err != nil {
				return nil, err
			}
			if row.Version != input.ExpectedVersion || row.Status != recommendationStatusDraft {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			source, err := service.loadRecommendationSource(
				ctx, tx, row.SourceType, row.SourceDishID,
			)
			if err != nil && appErrors.GetType(err) != appErrors.AdminNotFound {
				return nil, err
			}
			deletion := tx.Where("id = ? AND version = ?", row.ID, input.ExpectedVersion).
				Delete(&dishModel.PlatformRecommendation{})
			if deletion.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(deletion.Error, "delete recommendation")
			}
			if deletion.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if err == nil {
				if err := updateRecommendationCount(tx, source, -1); err != nil {
					return nil, err
				}
			}
			result := dishResponse.RecommendationDeleteResult{Deleted: true}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "recommendation_delete", row.ID, "",
				idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.RecommendationDeleteResult{}, false, err
	}
	var result dishResponse.RecommendationDeleteResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode recommendation delete result")
	}
	return result, replayed, nil
}

// PublishRecommendation 发布或重新发布推荐。
func (service *RecommendationService) PublishRecommendation(
	ctx context.Context,
	recommendationID string,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.RecommendationVersionInput,
) (dishResponse.RecommendationDetail, bool, error) {
	if err := validateRecommendationActor(actor); err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	recommendationID = strings.TrimSpace(recommendationID)
	if recommendationID == "" || input.ExpectedVersion < 1 {
		return dishResponse.RecommendationDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		ID    string                                 `json:"id"`    // 推荐ID
		Input dishRequest.RecommendationVersionInput `json:"input"` // 版本参数
	}{ID: recommendationID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "recommendation_publish",
		strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			row, err := loadRecommendation(ctx, tx, recommendationID)
			if err != nil {
				return nil, err
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if row.Status == serviceCommon.RecommendationStatusPublished {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			source, err := service.loadRecommendationSource(
				ctx, tx, row.SourceType, row.SourceDishID,
			)
			if err != nil {
				return nil, err
			}
			if !source.Available {
				return nil, appErrors.AdminInvalidImage.DefaultMsg()
			}
			var publishedCount int64
			if err := tx.Model(&dishModel.PlatformRecommendation{}).
				Where("position = ? AND status = ? AND id <> ?",
					row.Position, serviceCommon.RecommendationStatusPublished, row.ID,
				).
				Count(&publishedCount).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "count published recommendations")
			}
			if row.Position == serviceCommon.RecommendationPositionHome &&
				publishedCount >= recommendationHomeCapacity {
				return nil, appErrors.AdminAlreadyExists.New("首页精选最多同时发布5条")
			}
			before, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			now := service.now()
			update := tx.Model(&row).
				Where("version = ?", input.ExpectedVersion).
				Updates(map[string]interface{}{
					"status":              serviceCommon.RecommendationStatusPublished,
					"selected":            true,
					"published_at":        now,
					"offline_at":          nil,
					"offline_reason":      nil,
					"version":             gorm.Expr("version + 1"),
					"updated_by_id":       actor.AdministratorID,
					"updated_by_username": actor.Username,
					"updated_by_nickname": actor.Nickname,
					"updated_at":          now,
				})
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "publish recommendation")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			row.Status = serviceCommon.RecommendationStatusPublished
			row.Selected = true
			row.PublishedAt = &now
			row.OfflineAt = nil
			row.OfflineReason = nil
			row.Version++
			row.UpdatedByID = actor.AdministratorID
			row.UpdatedByUsername = actor.Username
			row.UpdatedByNickname = actor.Nickname
			row.UpdatedAt = now
			result, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "recommendation_publish", row.ID, "",
				idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	var result dishResponse.RecommendationDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode recommendation publish result")
	}
	return result, replayed, nil
}

// OfflineRecommendation 下线已发布推荐。
func (service *RecommendationService) OfflineRecommendation(
	ctx context.Context,
	recommendationID string,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.RecommendationOfflineInput,
) (dishResponse.RecommendationDetail, bool, error) {
	if err := validateRecommendationActor(actor); err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	recommendationID = strings.TrimSpace(recommendationID)
	input.Reason = strings.TrimSpace(input.Reason)
	if recommendationID == "" || input.ExpectedVersion < 1 ||
		utf8.RuneCountInString(input.Reason) < 4 ||
		utf8.RuneCountInString(input.Reason) > 200 {
		return dishResponse.RecommendationDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		ID    string                                 `json:"id"`    // 推荐ID
		Input dishRequest.RecommendationOfflineInput `json:"input"` // 下线参数
	}{ID: recommendationID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "recommendation_offline",
		strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			row, err := loadRecommendation(ctx, tx, recommendationID)
			if err != nil {
				return nil, err
			}
			if row.Version != input.ExpectedVersion ||
				row.Status != serviceCommon.RecommendationStatusPublished {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			now := service.now()
			update := tx.Model(&row).
				Where("version = ?", input.ExpectedVersion).
				Updates(map[string]interface{}{
					"status":              serviceCommon.RecommendationStatusOffline,
					"selected":            false,
					"offline_at":          now,
					"offline_reason":      input.Reason,
					"version":             gorm.Expr("version + 1"),
					"updated_by_id":       actor.AdministratorID,
					"updated_by_username": actor.Username,
					"updated_by_nickname": actor.Nickname,
					"updated_at":          now,
				})
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "offline recommendation")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			row.Status = serviceCommon.RecommendationStatusOffline
			row.Selected = false
			row.OfflineAt = &now
			row.OfflineReason = &input.Reason
			row.Version++
			row.UpdatedByID = actor.AdministratorID
			row.UpdatedByUsername = actor.Username
			row.UpdatedByNickname = actor.Nickname
			row.UpdatedAt = now
			result, err := service.recommendationDetail(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "recommendation_offline", row.ID, input.Reason,
				idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.RecommendationDetail{}, false, err
	}
	var result dishResponse.RecommendationDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode recommendation offline result")
	}
	return result, replayed, nil
}

// UpdateRecommendationSortOrder 在一个事务中批量调整同一推荐位的排序。
func (service *RecommendationService) UpdateRecommendationSortOrder(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.RecommendationSortOrderInput,
) (dishResponse.RecommendationSortOrderResult, bool, error) {
	if err := validateRecommendationActor(actor); err != nil {
		return dishResponse.RecommendationSortOrderResult{}, false, err
	}
	if len(input.Items) < 1 || len(input.Items) > 100 {
		return dishResponse.RecommendationSortOrderResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	seen := make(map[string]struct{}, len(input.Items))
	for index := range input.Items {
		input.Items[index].ID = strings.TrimSpace(input.Items[index].ID)
		item := input.Items[index]
		if item.ID == "" || item.SortOrder < 1 || item.ExpectedVersion < 1 {
			return dishResponse.RecommendationSortOrderResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
		}
		if _, exists := seen[item.ID]; exists {
			return dishResponse.RecommendationSortOrderResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
		}
		seen[item.ID] = struct{}{}
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "recommendation_sort_update",
		strings.TrimSpace(idempotencyKey), input,
		func(tx *gorm.DB) (interface{}, error) {
			before := make([]dishResponse.RecommendationSummary, 0, len(input.Items))
			after := make([]dishResponse.RecommendationSummary, 0, len(input.Items))
			position := ""
			for _, item := range input.Items {
				row, err := loadRecommendation(ctx, tx, item.ID)
				if err != nil {
					return nil, err
				}
				if row.Version != item.ExpectedVersion {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				if position == "" {
					position = row.Position
				} else if row.Position != position {
					return nil, appErrors.AdminBadRequest.DefaultMsg()
				}
				beforeItem, err := service.recommendationSummary(ctx, tx, row)
				if err != nil {
					return nil, err
				}
				before = append(before, beforeItem)
				now := service.now()
				update := tx.Model(&row).
					Where("version = ?", item.ExpectedVersion).
					Updates(map[string]interface{}{
						"sort_order":          item.SortOrder,
						"version":             gorm.Expr("version + 1"),
						"updated_by_id":       actor.AdministratorID,
						"updated_by_username": actor.Username,
						"updated_by_nickname": actor.Nickname,
						"updated_at":          now,
					})
				if update.Error != nil {
					return nil, appErrors.AdminInternal.Wrap(update.Error, "update recommendation sort")
				}
				if update.RowsAffected != 1 {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				row.SortOrder = item.SortOrder
				row.Version++
				row.UpdatedByID = actor.AdministratorID
				row.UpdatedByUsername = actor.Username
				row.UpdatedByNickname = actor.Nickname
				row.UpdatedAt = now
				afterItem, err := service.recommendationSummary(ctx, tx, row)
				if err != nil {
					return nil, err
				}
				after = append(after, afterItem)
			}
			result := dishResponse.RecommendationSortOrderResult{Updated: len(input.Items)}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "recommendation_sort_update", "batch", "",
				idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.RecommendationSortOrderResult{}, false, err
	}
	var result dishResponse.RecommendationSortOrderResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode recommendation sort result")
	}
	return result, replayed, nil
}

package dish

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
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
	"gorm.io/gorm"
)

const (
	// PermissionOfficialDishRead 表示读取官方菜品所需的权限。
	PermissionOfficialDishRead = "orderfood:official-dish:read"
	// PermissionOfficialDishCreate 表示创建官方菜品所需的权限。
	PermissionOfficialDishCreate = "orderfood:official-dish:create"
	// PermissionOfficialDishUpdate 表示编辑官方菜品所需的权限。
	PermissionOfficialDishUpdate = "orderfood:official-dish:update"
	// PermissionOfficialDishDelete 表示软删除官方菜品所需的权限。
	PermissionOfficialDishDelete = "orderfood:official-dish:delete"
	// PermissionOfficialDishCoverUpload 表示上传官方菜品封面所需的权限。
	PermissionOfficialDishCoverUpload = "orderfood:official-dish:cover-upload"

	officialDishUploadSource = "admin_upload"
	officialDishCoverScene   = "official_dish_cover"
	maxOfficialCoverBytes    = 10 << 20
)

// MaxOfficialCoverBytes 返回官方菜品封面允许的最大字节数。
func MaxOfficialCoverBytes() int64 {
	return maxOfficialCoverBytes
}

// OfficialDishService 提供官方菜品和官方封面管理能力。
type OfficialDishService struct {
	DB            *gorm.DB                          // 数据库连接
	Idempotency   *serviceCommon.IdempotencyService // 幂等处理服务
	MutationAudit serviceCommon.MutationAuditWriter // 变更审计写入器
	Now           func() time.Time                  // 当前时间函数
}

// NewOfficialDishService 创建官方菜品服务实例。
func NewOfficialDishService(
	db *gorm.DB,
	idempotency *serviceCommon.IdempotencyService,
) *OfficialDishService {
	if idempotency == nil {
		idempotency = serviceCommon.NewIdempotencyService(db)
	}
	return &OfficialDishService{
		DB:            db,
		Idempotency:   idempotency,
		MutationAudit: serviceCommon.GormMutationAuditWriter{},
		Now:           time.Now,
	}
}

// database 返回服务使用的数据库连接。
func (service *OfficialDishService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回统一为UTC的当前时间。
func (service *OfficialDishService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// validateOfficialDishActor 校验官方菜品管理操作上下文。
func validateOfficialDishActor(actor commonRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// writeMutationAudit 在官方菜品变更事务内写入审计记录。
func (service *OfficialDishService) writeMutationAudit(
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
		TargetType:            "official_dish",
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

// normalizeOptionalOfficialText 规范化可选文本并限制字符数。
func normalizeOptionalOfficialText(value *string, maximum int) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maximum {
		return nil, appErrors.AdminBadRequest.DefaultMsg()
	}
	return &trimmed, nil
}

// normalizeOfficialDishInput 规范化并校验官方菜品请求数据。
func normalizeOfficialDishInput(
	input dishRequest.OfficialDishCreateInput,
) (dishRequest.OfficialDishCreateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.CategoryID = strings.TrimSpace(input.CategoryID)
	input.CoverFileID = strings.TrimSpace(input.CoverFileID)
	description, err := normalizeOptionalOfficialText(input.Description, 180)
	if err != nil {
		return input, err
	}
	input.Description = description
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 40 ||
		input.CategoryID == "" || input.CoverFileID == "" ||
		input.Serving < 1 || input.Serving > 20 ||
		(input.Status != contentModel.DishStatusDraft &&
			input.Status != contentModel.DishStatusUsable) ||
		len(input.Ingredients) > 100 ||
		len(input.Steps) > 100 ||
		(input.Status == contentModel.DishStatusUsable &&
			(len(input.Ingredients) == 0 || len(input.Steps) == 0)) ||
		len(input.TagIDs) > 3 {
		return input, appErrors.AdminBadRequest.DefaultMsg()
	}
	seenTags := make(map[string]struct{}, len(input.TagIDs))
	for index, tagID := range input.TagIDs {
		tagID = strings.TrimSpace(tagID)
		if tagID == "" {
			return input, appErrors.AdminBadRequest.DefaultMsg()
		}
		if _, exists := seenTags[tagID]; exists {
			return input, appErrors.AdminBadRequest.DefaultMsg()
		}
		seenTags[tagID] = struct{}{}
		input.TagIDs[index] = tagID
	}
	for index := range input.Ingredients {
		ingredient := &input.Ingredients[index]
		ingredient.Name = strings.TrimSpace(ingredient.Name)
		quantity, quantityErr := normalizeOptionalOfficialText(ingredient.Quantity, 20)
		note, noteErr := normalizeOptionalOfficialText(ingredient.Note, 80)
		if quantityErr != nil || noteErr != nil ||
			ingredient.Name == "" || utf8.RuneCountInString(ingredient.Name) > 30 ||
			ingredient.SortOrder < 1 {
			return input, appErrors.AdminBadRequest.DefaultMsg()
		}
		ingredient.Quantity = quantity
		ingredient.Note = note
		if ingredient.UnitID != nil {
			value := strings.TrimSpace(*ingredient.UnitID)
			if value == "" {
				ingredient.UnitID = nil
			} else {
				ingredient.UnitID = &value
			}
		}
	}
	for index := range input.Steps {
		step := &input.Steps[index]
		step.Description = strings.TrimSpace(step.Description)
		imageURL, imageErr := normalizeOptionalOfficialText(step.ImageURL, 512)
		if imageErr != nil || step.Description == "" ||
			utf8.RuneCountInString(step.Description) > 500 ||
			step.SortOrder < 1 {
			return input, appErrors.AdminBadRequest.DefaultMsg()
		}
		step.ImageURL = imageURL
	}
	return input, nil
}

// officialDishReferences 保存校验后的分类、标签、单位和封面引用。
type officialDishReferences struct {
	Category dishModel.ContentCategory        // 分类
	Tags     []dishModel.ContentTag           // 标签列表
	Units    map[string]dishModel.ContentUnit // 单位映射
	Cover    contentModel.FrontMediaAsset     // 官方封面
}

// loadOfficialDishReferences 校验并查询官方菜品使用的关联数据。
func (service *OfficialDishService) loadOfficialDishReferences(
	ctx context.Context,
	db *gorm.DB,
	input dishRequest.OfficialDishCreateInput,
) (officialDishReferences, error) {
	result := officialDishReferences{
		Tags:  make([]dishModel.ContentTag, 0, len(input.TagIDs)),
		Units: make(map[string]dishModel.ContentUnit),
	}
	if err := db.WithContext(ctx).
		Where("public_id = ? AND enabled = ?", input.CategoryID, true).
		First(&result.Category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, appErrors.AdminNotFound.DefaultMsg()
		}
		return result, appErrors.AdminInternal.Wrap(err, "load official dish category")
	}
	if len(input.TagIDs) > 0 {
		if err := db.WithContext(ctx).
			Where("public_id IN ? AND enabled = ?", input.TagIDs, true).
			Find(&result.Tags).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "load official dish tags")
		}
		if len(result.Tags) != len(input.TagIDs) {
			return result, appErrors.AdminNotFound.DefaultMsg()
		}
	}
	unitIDs := make([]string, 0)
	seenUnits := make(map[string]struct{})
	for _, ingredient := range input.Ingredients {
		if ingredient.UnitID == nil {
			continue
		}
		if _, exists := seenUnits[*ingredient.UnitID]; !exists {
			seenUnits[*ingredient.UnitID] = struct{}{}
			unitIDs = append(unitIDs, *ingredient.UnitID)
		}
	}
	if len(unitIDs) > 0 {
		var units []dishModel.ContentUnit
		if err := db.WithContext(ctx).
			Where("public_id IN ? AND enabled = ?", unitIDs, true).
			Find(&units).Error; err != nil {
			return result, appErrors.AdminInternal.Wrap(err, "load official dish units")
		}
		if len(units) != len(unitIDs) {
			return result, appErrors.AdminNotFound.DefaultMsg()
		}
		for _, unit := range units {
			result.Units[unit.PublicID] = unit
		}
	}
	if err := db.WithContext(ctx).
		First(&result.Cover, "id = ? AND upload_source = ? AND scene = ? AND resource_status = ? AND review_status = ?",
			input.CoverFileID,
			officialDishUploadSource,
			officialDishCoverScene,
			serviceCommon.MediaResourceActive,
			contentModel.MediaReviewNotRequired,
		).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, appErrors.AdminInvalidImage.DefaultMsg()
		}
		return result, appErrors.AdminInternal.Wrap(err, "load official dish cover")
	}
	return result, nil
}

// preloadOfficialDish 查询官方菜品详情所需的全部关联数据。
func preloadOfficialDish(
	db *gorm.DB,
	dishID uint,
) (dishModel.OfficialDish, error) {
	var dish dishModel.OfficialDish
	err := db.
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

// officialDishCopyCount 统计官方菜品通过推荐产生的用户副本数量。
func officialDishCopyCount(
	ctx context.Context,
	db *gorm.DB,
	dishID uint,
) (int64, error) {
	var count int64
	err := db.WithContext(ctx).
		Table((dishModel.RecommendationCopy{}).TableName()+" AS copies").
		Joins("JOIN "+(dishModel.PlatformRecommendation{}).TableName()+
			" AS recommendations ON recommendations.id = copies.recommendation_id").
		Where("recommendations.source_type = ? AND recommendations.official_dish_id = ?",
			recommendationSourceOfficial, dishID).
		Count(&count).Error
	if err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count official dish copies")
	}
	return count, nil
}

// officialDishSummary 将官方菜品模型转换为列表摘要。
func officialDishSummary(
	dish dishModel.OfficialDish,
	copyCount int64,
) dishResponse.OfficialDishSummary {
	tags := make([]commonResponse.CatalogReference, 0, len(dish.Tags))
	for _, tag := range dish.Tags {
		tags = append(tags, commonResponse.CatalogReference{
			ID: tag.PublicID, Name: tag.Name, Enabled: tag.Enabled,
		})
	}
	return dishResponse.OfficialDishSummary{
		ID:       dish.PublicID,
		Name:     dish.Name,
		CoverURL: dish.CoverURL,
		Category: commonResponse.CatalogReference{
			ID: dish.Category.PublicID, Name: dish.Category.Name, Enabled: dish.Category.Enabled,
		},
		Tags:                tags,
		Serving:             dish.Serving,
		Status:              dish.Status,
		RecommendationCount: dish.RecommendationCount,
		CopyCount:           copyCount,
		Version:             dish.Version,
		CreatedAt:           dish.CreatedAt,
		UpdatedAt:           dish.UpdatedAt,
		UpdatedBy: serviceCommon.AdministratorSummary(
			dish.UpdatedByID, dish.UpdatedByUsername, dish.UpdatedByNickname,
		),
	}
}

// officialDishDetail 将官方菜品模型转换为详情。
func officialDishDetail(
	dish dishModel.OfficialDish,
	copyCount int64,
) dishResponse.OfficialDishDetail {
	ingredients := make([]contentResponse.DishIngredient, 0, len(dish.Ingredients))
	for _, ingredient := range dish.Ingredients {
		var unit *commonResponse.CatalogReference
		if ingredient.UnitID != nil {
			unit = &commonResponse.CatalogReference{
				ID:      ingredient.Unit.PublicID,
				Name:    ingredient.Unit.Name,
				Enabled: ingredient.Unit.Enabled,
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
	return dishResponse.OfficialDishDetail{
		OfficialDishSummary: officialDishSummary(dish, copyCount),
		Description:         dish.Description,
		CoverFileID:         dish.CoverFileID,
		Ingredients:         ingredients,
		Steps:               steps,
		CreatedBy: serviceCommon.AdministratorSummary(
			dish.CreatedByID, dish.CreatedByUsername, dish.CreatedByNickname,
		),
	}
}

// officialDishOnlineRecommendationCount 统计官方菜品当前在线推荐数量。
func officialDishOnlineRecommendationCount(
	ctx context.Context,
	db *gorm.DB,
	dishID uint,
) (int64, error) {
	var count int64
	if err := db.WithContext(ctx).
		Model(&dishModel.PlatformRecommendation{}).
		Where(
			"source_type = ? AND official_dish_id = ? AND status = ?",
			recommendationSourceOfficial,
			dishID,
			serviceCommon.RecommendationStatusPublished,
		).
		Count(&count).Error; err != nil {
		return 0, appErrors.AdminInternal.Wrap(err, "count online official dish recommendations")
	}
	return count, nil
}

// loadOfficialDishDetail 查询并转换官方菜品详情。
func (service *OfficialDishService) loadOfficialDishDetail(
	ctx context.Context,
	db *gorm.DB,
	dishID uint,
) (dishResponse.OfficialDishDetail, error) {
	dish, err := preloadOfficialDish(db.WithContext(ctx), dishID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dishResponse.OfficialDishDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return dishResponse.OfficialDishDetail{}, appErrors.AdminInternal.Wrap(err, "load official dish detail")
	}
	copyCount, err := officialDishCopyCount(ctx, db, dish.ID)
	if err != nil {
		return dishResponse.OfficialDishDetail{}, err
	}
	detail := officialDishDetail(dish, copyCount)
	onlineCount, err := officialDishOnlineRecommendationCount(ctx, db, dish.ID)
	if err != nil {
		return dishResponse.OfficialDishDetail{}, err
	}
	detail.OnlineRecommendationCount = onlineCount
	return detail, nil
}

// ListOfficialDishes 分页查询官方菜品。
func (service *OfficialDishService) ListOfficialDishes(
	ctx context.Context,
	query dishRequest.OfficialDishListQuery,
) (commonResponse.Page[dishResponse.OfficialDishSummary], error) {
	query.ApplyDefaults()
	if !serviceCommon.ValidPagination(query.Page, query.PageSize) ||
		!serviceCommon.ValidTimeRange(query.CreatedFrom, query.CreatedTo) ||
		utf8.RuneCountInString(query.Keyword) > 60 ||
		len(query.TagIDs) > 3 ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return commonResponse.Page[dishResponse.OfficialDishSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[dishResponse.OfficialDishSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&dishModel.OfficialDish{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		statement = statement.Where("of_off_dishes.name LIKE ?", "%"+keyword+"%")
	}
	if categoryID := strings.TrimSpace(query.CategoryID); categoryID != "" {
		statement = statement.Where("category_id IN (?)",
			db.Model(&dishModel.ContentCategory{}).
				Select("id").Where("public_id = ?", categoryID))
	}
	if query.Status != "" {
		statement = statement.Where("of_off_dishes.status = ?", query.Status)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("of_off_dishes.created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("of_off_dishes.created_at <= ?", *query.CreatedTo)
	}
	if len(query.TagIDs) > 0 {
		statement = statement.Where(
			"of_off_dishes.id IN (?)",
			db.Table((dishModel.OfficialDishTag{}).TableName()+" AS dish_tags").
				Joins("JOIN "+(dishModel.ContentTag{}).TableName()+
					" AS tags ON tags.id = dish_tags.tag_id").
				Select("dish_tags.dish_id").
				Where("tags.public_id IN ?", query.TagIDs).
				Group("dish_tags.dish_id").
				Having("COUNT(DISTINCT tags.public_id) = ?", len(query.TagIDs)),
		)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[dishResponse.OfficialDishSummary]{}, appErrors.AdminInternal.Wrap(err, "count official dishes")
	}
	sortColumns := map[string]string{
		"updatedAt": "of_off_dishes.updated_at",
		"createdAt": "of_off_dishes.created_at",
		"name":      "of_off_dishes.name",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[dishResponse.OfficialDishSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	rows := make([]dishModel.OfficialDish, 0)
	if err := statement.
		Preload("Category").
		Preload("Tags").
		Order(column + " " + query.SortOrder).
		Order("of_off_dishes.id ASC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[dishResponse.OfficialDishSummary]{}, appErrors.AdminInternal.Wrap(err, "list official dishes")
	}
	list := make([]dishResponse.OfficialDishSummary, 0, len(rows))
	for _, dish := range rows {
		copyCount, err := officialDishCopyCount(ctx, db, dish.ID)
		if err != nil {
			return commonResponse.Page[dishResponse.OfficialDishSummary]{}, err
		}
		summary := officialDishSummary(dish, copyCount)
		onlineCount, err := officialDishOnlineRecommendationCount(ctx, db, dish.ID)
		if err != nil {
			return commonResponse.Page[dishResponse.OfficialDishSummary]{}, err
		}
		summary.OnlineRecommendationCount = onlineCount
		list = append(list, summary)
	}
	return commonResponse.Page[dishResponse.OfficialDishSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetOfficialDish 获取官方菜品详情。
func (service *OfficialDishService) GetOfficialDish(
	ctx context.Context,
	publicID string,
) (dishResponse.OfficialDishDetail, error) {
	db := service.database()
	publicID = strings.TrimSpace(publicID)
	if db == nil || publicID == "" {
		return dishResponse.OfficialDishDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var dish dishModel.OfficialDish
	if err := db.WithContext(ctx).First(&dish, "public_id = ?", publicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dishResponse.OfficialDishDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return dishResponse.OfficialDishDetail{}, appErrors.AdminInternal.Wrap(err, "load official dish")
	}
	return service.loadOfficialDishDetail(ctx, db, dish.ID)
}

// persistOfficialDishParts 替换官方菜品的食材、步骤和标签关联。
func persistOfficialDishParts(
	tx *gorm.DB,
	dishID uint,
	input dishRequest.OfficialDishCreateInput,
	references officialDishReferences,
) error {
	if err := tx.Unscoped().Where("dish_id = ?", dishID).
		Delete(&dishModel.OfficialDishIngredient{}).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "replace official dish ingredients")
	}
	if err := tx.Unscoped().Where("dish_id = ?", dishID).
		Delete(&dishModel.OfficialDishStep{}).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "replace official dish steps")
	}
	if err := tx.Where("dish_id = ?", dishID).
		Delete(&dishModel.OfficialDishTag{}).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "replace official dish tags")
	}
	ingredients := make([]dishModel.OfficialDishIngredient, 0, len(input.Ingredients))
	for _, item := range input.Ingredients {
		var unitID *uint
		if item.UnitID != nil {
			unit := references.Units[*item.UnitID]
			unitID = &unit.ID
		}
		ingredients = append(ingredients, dishModel.OfficialDishIngredient{
			DishID: dishID, Name: item.Name, Quantity: item.Quantity,
			UnitID: unitID, Note: item.Note, SortOrder: item.SortOrder,
		})
	}
	if len(ingredients) > 0 {
		if err := tx.Create(&ingredients).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "create official dish ingredients")
		}
	}
	steps := make([]dishModel.OfficialDishStep, 0, len(input.Steps))
	for _, item := range input.Steps {
		steps = append(steps, dishModel.OfficialDishStep{
			DishID: dishID, Description: item.Description,
			ImageURL: item.ImageURL, SortOrder: item.SortOrder,
		})
	}
	if len(steps) > 0 {
		if err := tx.Create(&steps).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "create official dish steps")
		}
	}
	if len(references.Tags) > 0 {
		joins := make([]dishModel.OfficialDishTag, 0, len(references.Tags))
		now := time.Now().UTC()
		for _, tag := range references.Tags {
			joins = append(joins, dishModel.OfficialDishTag{
				DishID: dishID, TagID: tag.ID, CreatedAt: now,
			})
		}
		if err := tx.Create(&joins).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "create official dish tags")
		}
	}
	return nil
}

// CreateOfficialDish 创建官方菜品。
func (service *OfficialDishService) CreateOfficialDish(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.OfficialDishCreateInput,
) (dishResponse.OfficialDishDetail, bool, error) {
	if err := validateOfficialDishActor(actor); err != nil {
		return dishResponse.OfficialDishDetail{}, false, err
	}
	normalized, err := normalizeOfficialDishInput(input)
	if err != nil {
		return dishResponse.OfficialDishDetail{}, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"official_dish_create",
		strings.TrimSpace(idempotencyKey),
		normalized,
		func(tx *gorm.DB) (interface{}, error) {
			references, err := service.loadOfficialDishReferences(ctx, tx, normalized)
			if err != nil {
				return nil, err
			}
			now := service.now()
			dish := dishModel.OfficialDish{
				PublicID:          serviceCommon.NewPublicID("official"),
				Name:              normalized.Name,
				CoverFileID:       references.Cover.ID,
				CoverURL:          references.Cover.URL,
				CategoryID:        references.Category.ID,
				Serving:           normalized.Serving,
				Description:       normalized.Description,
				Status:            normalized.Status,
				Version:           1,
				CreatedByID:       actor.AdministratorID,
				CreatedByUsername: actor.Username,
				CreatedByNickname: actor.Nickname,
				UpdatedByID:       actor.AdministratorID,
				UpdatedByUsername: actor.Username,
				UpdatedByNickname: actor.Nickname,
			}
			dish.CreatedAt = now
			dish.UpdatedAt = now
			if err := tx.Create(&dish).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create official dish")
			}
			if err := persistOfficialDishParts(tx, dish.ID, normalized, references); err != nil {
				return nil, err
			}
			result, err := service.loadOfficialDishDetail(ctx, tx, dish.ID)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "official_dish_create", dish.PublicID, "",
				idempotencyKey, nil, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.OfficialDishDetail{}, false, err
	}
	var result dishResponse.OfficialDishDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode official dish create result")
	}
	return result, replayed, nil
}

// offlineOfficialDishRecommendations 下线官方菜品当前在线的所有推荐。
func (service *OfficialDishService) offlineOfficialDishRecommendations(
	tx *gorm.DB,
	dishID uint,
	actor commonRequest.AdminActor,
	reason string,
) (int64, error) {
	now := service.now()
	update := tx.Model(&dishModel.PlatformRecommendation{}).
		Where("source_type = ? AND official_dish_id = ? AND status = ?",
			recommendationSourceOfficial, dishID, serviceCommon.RecommendationStatusPublished).
		Updates(map[string]interface{}{
			"status":              serviceCommon.RecommendationStatusOffline,
			"selected":            false,
			"offline_at":          now,
			"offline_reason":      reason,
			"updated_by_id":       actor.AdministratorID,
			"updated_by_username": actor.Username,
			"updated_by_nickname": actor.Nickname,
			"updated_at":          now,
			"version":             gorm.Expr("version + 1"),
		})
	if update.Error != nil {
		return 0, appErrors.AdminInternal.Wrap(update.Error, "offline official dish recommendations")
	}
	return update.RowsAffected, nil
}

// UpdateOfficialDish 编辑官方菜品并自动下线在线推荐。
func (service *OfficialDishService) UpdateOfficialDish(
	ctx context.Context,
	publicID string,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.OfficialDishUpdateInput,
) (dishResponse.OfficialDishMutationResult, bool, error) {
	if err := validateOfficialDishActor(actor); err != nil {
		return dishResponse.OfficialDishMutationResult{}, false, err
	}
	publicID = strings.TrimSpace(publicID)
	normalized, err := normalizeOfficialDishInput(input.OfficialDishCreateInput)
	if err != nil || publicID == "" || input.ExpectedVersion < 1 {
		return dishResponse.OfficialDishMutationResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		ID              string                              `json:"id"`              // 官方菜品ID
		Input           dishRequest.OfficialDishCreateInput `json:"input"`           // 菜品内容
		ExpectedVersion int64                               `json:"expectedVersion"` // 预期数据版本
	}{ID: publicID, Input: normalized, ExpectedVersion: input.ExpectedVersion}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"official_dish_update",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var dish dishModel.OfficialDish
			if err := tx.
				First(&dish, "public_id = ?", publicID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load official dish for mutation")
			}
			if dish.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before, err := service.loadOfficialDishDetail(ctx, tx, dish.ID)
			if err != nil {
				return nil, err
			}
			references, err := service.loadOfficialDishReferences(ctx, tx, normalized)
			if err != nil {
				return nil, err
			}
			now := service.now()
			update := tx.Model(&dishModel.OfficialDish{}).
				Where("id = ? AND version = ?", dish.ID, input.ExpectedVersion).
				Updates(map[string]interface{}{
					"name":                normalized.Name,
					"cover_file_id":       references.Cover.ID,
					"cover_url":           references.Cover.URL,
					"category_id":         references.Category.ID,
					"serving":             normalized.Serving,
					"description":         normalized.Description,
					"status":              normalized.Status,
					"updated_by_id":       actor.AdministratorID,
					"updated_by_username": actor.Username,
					"updated_by_nickname": actor.Nickname,
					"updated_at":          now,
					"version":             gorm.Expr("version + 1"),
				})
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "update official dish")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if err := persistOfficialDishParts(tx, dish.ID, normalized, references); err != nil {
				return nil, err
			}
			offlineCount, err := service.offlineOfficialDishRecommendations(
				tx, dish.ID, actor, "官方菜品内容更新，系统自动下线",
			)
			if err != nil {
				return nil, err
			}
			detail, err := service.loadOfficialDishDetail(ctx, tx, dish.ID)
			if err != nil {
				return nil, err
			}
			result := dishResponse.OfficialDishMutationResult{
				Dish: detail, OfflineRecommendationCount: offlineCount,
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "official_dish_update", dish.PublicID,
				"编辑官方菜品", idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.OfficialDishMutationResult{}, false, err
	}
	var result dishResponse.OfficialDishMutationResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode official dish update result")
	}
	return result, replayed, nil
}

// DeleteOfficialDish 软删除官方菜品并自动下线在线推荐。
func (service *OfficialDishService) DeleteOfficialDish(
	ctx context.Context,
	publicID string,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input dishRequest.OfficialDishDeleteInput,
) (dishResponse.OfficialDishDeleteResult, bool, error) {
	if err := validateOfficialDishActor(actor); err != nil {
		return dishResponse.OfficialDishDeleteResult{}, false, err
	}
	publicID = strings.TrimSpace(publicID)
	input.Reason = strings.TrimSpace(input.Reason)
	if publicID == "" || input.ExpectedVersion < 1 ||
		utf8.RuneCountInString(input.Reason) < 4 ||
		utf8.RuneCountInString(input.Reason) > 200 {
		return dishResponse.OfficialDishDeleteResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		ID    string                              `json:"id"`    // 官方菜品ID
		Input dishRequest.OfficialDishDeleteInput `json:"input"` // 删除参数
	}{ID: publicID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"official_dish_delete",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var dish dishModel.OfficialDish
			if err := tx.
				First(&dish, "public_id = ?", publicID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load official dish for delete")
			}
			if dish.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before, err := service.loadOfficialDishDetail(ctx, tx, dish.ID)
			if err != nil {
				return nil, err
			}
			offlineCount, err := service.offlineOfficialDishRecommendations(
				tx, dish.ID, actor, "官方菜品已删除，系统自动下线",
			)
			if err != nil {
				return nil, err
			}
			now := service.now()
			update := tx.Model(&dishModel.OfficialDish{}).
				Where("id = ? AND version = ?", dish.ID, input.ExpectedVersion).
				Updates(map[string]interface{}{
					"deleted_reason": input.Reason,
					"deleted_by_id":  actor.AdministratorID,
					"updated_at":     now,
					"version":        gorm.Expr("version + 1"),
				})
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "mark official dish deleted")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if err := tx.Delete(&dish).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "soft delete official dish")
			}
			result := dishResponse.OfficialDishDeleteResult{
				Deleted: true, OfflineRecommendationCount: offlineCount,
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "official_dish_delete", dish.PublicID,
				input.Reason, idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		return dishResponse.OfficialDishDeleteResult{}, false, err
	}
	var result dishResponse.OfficialDishDeleteResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode official dish delete result")
	}
	return result, replayed, nil
}

// officialDishCoverSummary 将媒体文件转换为官方封面摘要。
func officialDishCoverSummary(
	asset contentModel.FrontMediaAsset,
) dishResponse.OfficialDishCoverSummary {
	administratorID := uint(0)
	if asset.AdministratorID != nil {
		administratorID = *asset.AdministratorID
	}
	username := ""
	if asset.AdministratorUsername != nil {
		username = *asset.AdministratorUsername
	}
	return dishResponse.OfficialDishCoverSummary{
		FileID:         asset.ID,
		URL:            asset.URL,
		FileName:       asset.FileName,
		FileSize:       asset.SizeBytes,
		MimeType:       asset.ContentType,
		Width:          asset.Width,
		Height:         asset.Height,
		UploadSource:   asset.UploadSource,
		SourceScene:    asset.Scene,
		ReviewStatus:   asset.ReviewStatus,
		ResourceStatus: asset.ResourceStatus,
		UploadedBy: serviceCommon.AdministratorSummary(
			administratorID, username, asset.AdministratorNickname,
		),
		CreatedAt: asset.CreatedAt,
	}
}

// ListOfficialDishCovers 分页查询管理端上传的官方菜品封面。
func (service *OfficialDishService) ListOfficialDishCovers(
	ctx context.Context,
	query dishRequest.OfficialDishCoverListQuery,
) (commonResponse.Page[dishResponse.OfficialDishCoverSummary], error) {
	query.ApplyDefaults()
	db := service.database()
	if db == nil {
		return commonResponse.Page[dishResponse.OfficialDishCoverSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&contentModel.FrontMediaAsset{}).
		Where("upload_source = ? AND scene = ? AND resource_status = ?",
			officialDishUploadSource, officialDishCoverScene, serviceCommon.MediaResourceActive)
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		statement = statement.Where("file_name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[dishResponse.OfficialDishCoverSummary]{}, appErrors.AdminInternal.Wrap(err, "count official dish covers")
	}
	sortColumns := map[string]string{
		"createdAt": "created_at",
		"fileSize":  "size_bytes",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[dishResponse.OfficialDishCoverSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	rows := make([]contentModel.FrontMediaAsset, 0)
	if err := statement.Order(column + " " + query.SortOrder).
		Order("id ASC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[dishResponse.OfficialDishCoverSummary]{}, appErrors.AdminInternal.Wrap(err, "list official dish covers")
	}
	list := make([]dishResponse.OfficialDishCoverSummary, 0, len(rows))
	for _, asset := range rows {
		list = append(list, officialDishCoverSummary(asset))
	}
	return commonResponse.Page[dishResponse.OfficialDishCoverSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// UploadOfficialDishCover 上传无需外部审核的管理端官方菜品封面。
func (service *OfficialDishService) UploadOfficialDishCover(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	fileName string,
	contentType string,
	payload []byte,
) (dishResponse.OfficialDishCoverSummary, bool, error) {
	if err := validateOfficialDishActor(actor); err != nil {
		return dishResponse.OfficialDishCoverSummary{}, false, err
	}
	fileName = filepath.Base(strings.TrimSpace(fileName))
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if fileName == "" || len(payload) == 0 || len(payload) > maxOfficialCoverBytes ||
		(contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif") {
		return dishResponse.OfficialDishCoverSummary{}, false, appErrors.AdminInvalidImage.DefaultMsg()
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(payload))
	if err != nil || config.Width < 1 || config.Height < 1 ||
		config.Width > 10000 || config.Height > 10000 {
		return dishResponse.OfficialDishCoverSummary{}, false, appErrors.AdminInvalidImage.DefaultMsg()
	}
	digest := sha256.Sum256(payload)
	idempotencyPayload := struct {
		FileName    string `json:"fileName"`    // 原始文件名
		ContentType string `json:"contentType"` // MIME类型
		ContentHash string `json:"contentHash"` // 文件内容摘要
	}{
		FileName: fileName, ContentType: contentType,
		ContentHash: hex.EncodeToString(digest[:]),
	}
	var writtenPath string
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"official_dish_cover_upload",
		strings.TrimSpace(idempotencyKey),
		idempotencyPayload,
		func(tx *gorm.DB) (interface{}, error) {
			storePath := strings.TrimSpace(global.GVA_CONFIG.Local.StorePath)
			if storePath == "" {
				return nil, appErrors.AdminInternal.New("local image store path is not configured")
			}
			if err := os.MkdirAll(storePath, 0o750); err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create official image directory")
			}
			extension := ".jpg"
			if contentType == "image/png" {
				extension = ".png"
			} else if contentType == "image/gif" {
				extension = ".gif"
			}
			fileID := serviceCommon.NewPublicID("cover")
			storagePath := filepath.Join(storePath, fileID+extension)
			if err := os.WriteFile(storagePath, payload, 0o640); err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "store official image")
			}
			writtenPath = storagePath
			username := actor.Username
			publicURL := strings.TrimRight(global.GVA_CONFIG.Local.Path, "/") + "/" + fileID + extension
			asset := contentModel.FrontMediaAsset{
				ID:                    fileID,
				UserID:                "",
				AdministratorID:       &actor.AdministratorID,
				AdministratorUsername: &username,
				AdministratorNickname: actor.Nickname,
				FileName:              fileName,
				UploadSource:          officialDishUploadSource,
				Scene:                 officialDishCoverScene,
				ResourceStatus:        serviceCommon.MediaResourceActive,
				URL:                   publicURL,
				StoragePath:           storagePath,
				ContentType:           contentType,
				Width:                 config.Width,
				Height:                config.Height,
				SizeBytes:             int64(len(payload)),
				Checksum:              hex.EncodeToString(digest[:]),
				ReviewStatus:          contentModel.MediaReviewNotRequired,
				CreatedAt:             service.now(),
			}
			if err := tx.Create(&asset).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "persist official image")
			}
			result := officialDishCoverSummary(asset)
			if err := service.writeMutationAudit(
				ctx, tx, actor, "official_dish_cover_upload", asset.ID,
				"管理端上传官方菜品封面", idempotencyKey, nil, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	if err != nil {
		if writtenPath != "" {
			_ = os.Remove(writtenPath)
		}
		return dishResponse.OfficialDishCoverSummary{}, false, err
	}
	var result dishResponse.OfficialDishCoverSummary
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode official image upload result")
	}
	return result, replayed, nil
}

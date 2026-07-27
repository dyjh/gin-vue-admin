package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"gorm.io/gorm"
)

// ContentService 提供内容管理业务能力。
type ContentService struct {
	DB  *gorm.DB         // 业务数据库
	Now func() time.Time // 可注入的当前时间函数
}

func (service *ContentService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *ContentService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// Metadata 获取菜品分类、标签、单位和运行时配置。
func (service *ContentService) Metadata(
	ctx context.Context,
	runtime frontResponse.RuntimeConfig,
) (frontResponse.Metadata, error) {
	db := service.database()
	if db == nil {
		return frontResponse.Metadata{}, appErrors.FrontInternal.DefaultMsg()
	}
	var categories []orderfoodModel.ContentCategory
	var tags []orderfoodModel.ContentTag
	var units []orderfoodModel.ContentUnit
	if err := db.WithContext(ctx).Order("id asc").Find(&categories).Error; err != nil {
		return frontResponse.Metadata{}, appErrors.FrontInternal.Wrap(err, "load dish categories")
	}
	if err := db.WithContext(ctx).Order("id asc").Find(&tags).Error; err != nil {
		return frontResponse.Metadata{}, appErrors.FrontInternal.Wrap(err, "load dish tags")
	}
	if err := db.WithContext(ctx).Where("enabled = ?", true).Order("id asc").Find(&units).Error; err != nil {
		return frontResponse.Metadata{}, appErrors.FrontInternal.Wrap(err, "load ingredient units")
	}
	result := frontResponse.Metadata{
		DishCategories:  make([]frontResponse.DishCategory, 0, len(categories)),
		DishTags:        make([]frontResponse.DishTag, 0, len(tags)),
		IngredientUnits: make([]string, 0, len(units)),
		RuntimeConfig:   runtime,
	}
	for _, category := range categories {
		result.DishCategories = append(result.DishCategories, frontResponse.DishCategory{
			ID: category.PublicID, Name: category.Name, SortOrder: int(category.ID), Enabled: category.Enabled,
		})
	}
	for _, tag := range tags {
		result.DishTags = append(result.DishTags, frontResponse.DishTag{
			ID: tag.PublicID, Name: tag.Name, SortOrder: int(tag.ID), Enabled: tag.Enabled,
		})
	}
	for _, unit := range units {
		result.IngredientUnits = append(result.IngredientUnits, unit.Name)
	}
	return result, nil
}

func preloadDish(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Category").
		Preload("Tags").
		Preload("Ingredients", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order asc, id asc") }).
		Preload("Ingredients.Unit").
		Preload("Steps", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order asc, id asc") }).
		Preload("Owner")
}

// preloadOfficialDish 查询官方菜品及其详情关联。
func preloadOfficialDish(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Category").
		Preload("Tags").
		Preload("Ingredients", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order asc, id asc") }).
		Preload("Ingredients.Unit").
		Preload("Steps", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order asc, id asc") })
}

func dishSummary(dish orderfoodModel.UserDish) frontResponse.DishSummary {
	tags := make([]string, 0, len(dish.Tags))
	for _, tag := range dish.Tags {
		tags = append(tags, tag.Name)
	}
	sort.Strings(tags)
	return frontResponse.DishSummary{
		ID:           dish.PublicID,
		Name:         dish.Name,
		Category:     dish.Category.Name,
		Tags:         tags,
		CoverURL:     stringValue(dish.CoverURL),
		Serving:      dish.Serving,
		Status:       dish.Status,
		Discoverable: dish.Discoverable,
		SourceLocked: dish.SourceLocked,
	}
}

func dishResponse(dish orderfoodModel.UserDish) frontResponse.Dish {
	result := frontResponse.Dish{
		DishSummary: dishSummary(dish),
		Description: dish.Description,
		Ingredients: make([]frontResponse.Ingredient, 0, len(dish.Ingredients)),
		Steps:       make([]frontResponse.DishStep, 0, len(dish.Steps)),
		CreatedAt:   dish.CreatedAt,
		UpdatedAt:   dish.UpdatedAt,
	}
	for _, ingredient := range dish.Ingredients {
		unit := ""
		if ingredient.Unit != nil {
			unit = ingredient.Unit.Name
		}
		result.Ingredients = append(result.Ingredients, frontResponse.Ingredient{
			ID: orderfoodModel.NewID(), Name: ingredient.Name, Amount: stringValue(ingredient.Quantity),
			Unit: unit, Note: ingredient.Note, SortOrder: ingredient.SortOrder,
		})
	}
	for _, step := range dish.Steps {
		result.Steps = append(result.Steps, frontResponse.DishStep{
			ID: orderfoodModel.NewID(), Text: step.Description, ImageURL: step.ImageURL, SortOrder: step.SortOrder,
		})
	}
	return result
}

// officialDishSummary 将官方菜品转换为小程序菜品摘要。
func officialDishSummary(dish orderfoodModel.OfficialDish) frontResponse.DishSummary {
	tags := make([]string, 0, len(dish.Tags))
	for _, tag := range dish.Tags {
		tags = append(tags, tag.Name)
	}
	sort.Strings(tags)
	return frontResponse.DishSummary{
		ID:           dish.PublicID,
		Name:         dish.Name,
		Category:     dish.Category.Name,
		Tags:         tags,
		CoverURL:     dish.CoverURL,
		Serving:      dish.Serving,
		Status:       dish.Status,
		Discoverable: false,
		SourceLocked: true,
	}
}

// officialDishResponse 将官方菜品转换为小程序菜品详情。
func officialDishResponse(dish orderfoodModel.OfficialDish) frontResponse.Dish {
	result := frontResponse.Dish{
		DishSummary: officialDishSummary(dish),
		Description: dish.Description,
		Ingredients: make([]frontResponse.Ingredient, 0, len(dish.Ingredients)),
		Steps:       make([]frontResponse.DishStep, 0, len(dish.Steps)),
		CreatedAt:   dish.CreatedAt,
		UpdatedAt:   dish.UpdatedAt,
	}
	for _, ingredient := range dish.Ingredients {
		unit := ""
		if ingredient.Unit != nil {
			unit = ingredient.Unit.Name
		}
		result.Ingredients = append(result.Ingredients, frontResponse.Ingredient{
			ID: orderfoodModel.NewID(), Name: ingredient.Name,
			Amount: stringValue(ingredient.Quantity), Unit: unit,
			Note: ingredient.Note, SortOrder: ingredient.SortOrder,
		})
	}
	for _, step := range dish.Steps {
		result.Steps = append(result.Steps, frontResponse.DishStep{
			ID: orderfoodModel.NewID(), Text: step.Description,
			ImageURL: step.ImageURL, SortOrder: step.SortOrder,
		})
	}
	return result
}

// ListDishes 分页查询菜品列表。
func (service *ContentService) ListDishes(
	ctx context.Context,
	userID string,
	query frontRequest.DishListQuery,
) (frontResponse.Page[frontResponse.DishSummary], error) {
	query.PageQuery.Defaults()
	db := service.database()
	statement := db.WithContext(ctx).Model(&orderfoodModel.UserDish{}).
		Where("owner_id = ?", userID)
	if keyword := strings.TrimSpace(query.Q); keyword != "" {
		statement = statement.Where("name LIKE ?", "%"+keyword+"%")
	}
	if query.Status != "" {
		statement = statement.Where("status = ?", query.Status)
	}
	if query.Category != "" {
		statement = statement.Where(
			"category_id IN (?)",
			db.Model(&orderfoodModel.ContentCategory{}).
				Select("id").
				Where("public_id = ? OR name = ?", query.Category, query.Category),
		)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.DishSummary]{}, appErrors.FrontInternal.Wrap(err, "count dishes")
	}
	var rows []orderfoodModel.UserDish
	if err := preloadDish(statement).
		Order("updated_at desc, id desc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return frontResponse.Page[frontResponse.DishSummary]{}, appErrors.FrontInternal.Wrap(err, "list dishes")
	}
	list := make([]frontResponse.DishSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, dishSummary(row))
	}
	return frontResponse.Page[frontResponse.DishSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// Dish 获取当前用户可访问的个人菜品详情。
func (service *ContentService) Dish(
	ctx context.Context,
	userID string,
	dishID string,
) (frontResponse.Dish, error) {
	var row orderfoodModel.UserDish
	err := preloadDish(service.database().WithContext(ctx)).
		First(&row, "public_id = ? AND owner_id = ?", dishID, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.Dish{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.Dish{}, appErrors.FrontInternal.Wrap(err, "load dish")
	}
	return dishResponse(row), nil
}

func (service *ContentService) resolveCategory(
	tx *gorm.DB,
	value string,
) (orderfoodModel.ContentCategory, error) {
	var category orderfoodModel.ContentCategory
	if err := tx.First(&category, "(public_id = ? OR name = ?) AND enabled = ?", value, value, true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return category, appErrors.FrontBadRequest.DefaultMsg()
		}
		return category, appErrors.FrontInternal.Wrap(err, "resolve dish category")
	}
	return category, nil
}

func (service *ContentService) resolveAsset(
	tx *gorm.DB,
	userID string,
	fileID string,
	scenes ...string,
) (orderfoodModel.FrontMediaAsset, error) {
	var asset orderfoodModel.FrontMediaAsset
	statement := tx.First(
		&asset,
		"id = ? AND user_id = ? AND review_status IN ?",
		fileID,
		userID,
		orderfoodModel.UsableMediaReviewStatuses(),
	)
	if statement.Error != nil {
		if errors.Is(statement.Error, gorm.ErrRecordNotFound) {
			return asset, appErrors.FrontInvalidImage.DefaultMsg()
		}
		return asset, appErrors.FrontInternal.Wrap(statement.Error, "resolve front image")
	}
	if len(scenes) > 0 {
		valid := false
		for _, scene := range scenes {
			if asset.Scene == scene {
				valid = true
				break
			}
		}
		if !valid {
			return asset, appErrors.FrontInvalidImage.DefaultMsg()
		}
	}
	return asset, nil
}

func (service *ContentService) persistDishChildren(
	tx *gorm.DB,
	userID string,
	dish *orderfoodModel.UserDish,
	input frontRequest.DishUpsertInput,
) error {
	if err := tx.Where("dish_id = ?", dish.ID).Delete(&orderfoodModel.DishIngredient{}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "replace dish ingredients")
	}
	if err := tx.Where("dish_id = ?", dish.ID).Delete(&orderfoodModel.DishStep{}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "replace dish steps")
	}
	if err := tx.Where("dish_id = ?", dish.ID).Delete(&orderfoodModel.UserDishTag{}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "replace dish tags")
	}
	for _, item := range input.Ingredients {
		var unit orderfoodModel.ContentUnit
		if err := tx.First(&unit, "(public_id = ? OR name = ?) AND enabled = ?", item.Unit, item.Unit, true).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontBadRequest.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "resolve ingredient unit")
		}
		quantity := item.Amount
		row := orderfoodModel.DishIngredient{
			DishID: dish.ID, Name: strings.TrimSpace(item.Name), Quantity: &quantity,
			UnitID: &unit.ID, Note: item.Note, SortOrder: item.SortOrder,
		}
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create dish ingredient")
		}
	}
	for _, item := range input.Steps {
		var imageURL *string
		if item.ImageFileID != nil {
			asset, err := service.resolveAsset(tx, userID, *item.ImageFileID, "dish_step")
			if err != nil {
				return err
			}
			imageURL = &asset.URL
		}
		row := orderfoodModel.DishStep{
			DishID: dish.ID, Description: strings.TrimSpace(item.Text),
			ImageURL: imageURL, SortOrder: item.SortOrder,
		}
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create dish step")
		}
	}
	for _, value := range input.Tags {
		var tag orderfoodModel.ContentTag
		if err := tx.First(&tag, "(public_id = ? OR name = ?) AND enabled = ?", value, value, true).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontBadRequest.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "resolve dish tag")
		}
		if err := tx.Create(&orderfoodModel.UserDishTag{
			DishID: dish.ID, TagID: tag.ID, CreatedAt: service.now(),
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create dish tag")
		}
	}
	return nil
}

// UpsertDish 创建或更新个人菜品及其食材、步骤关联数据。
func (service *ContentService) UpsertDish(
	ctx context.Context,
	userID string,
	dishID *string,
	input frontRequest.DishUpsertInput,
) (frontResponse.Dish, error) {
	db := service.database()
	var publicID string
	preferenceEnabled := preferenceUpdatesEnabled(ctx)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		category, err := service.resolveCategory(tx, input.Category)
		if err != nil {
			return err
		}
		cover, err := service.resolveAsset(tx, userID, input.CoverFileID, "dish_cover", "generated_cover")
		if err != nil {
			return err
		}
		now := service.now()
		if dishID == nil {
			row := orderfoodModel.UserDish{
				PublicID: orderfoodModel.NewID(), OwnerID: userID, CoverFileID: cover.ID,
				CoverURL: &cover.URL, Name: strings.TrimSpace(input.Name), CategoryID: category.ID,
				Status: input.Status, Discoverable: false, SourceType: orderfoodModel.SourceTypeManual,
				SourceLocked: false, Description: input.Description, Serving: input.Serving,
				MediaReviewStatus: cover.ReviewStatus, Version: 1,
			}
			if err := tx.Create(&row).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create dish")
			}
			if err := service.persistDishChildren(tx, userID, &row, input); err != nil {
				return err
			}
			if preferenceEnabled && input.Status == orderfoodModel.DishStatusUsable {
				if err := ServiceGroupApp.PreferenceService.queueDishEvidenceTx(
					ctx, tx, userID,
					orderfoodModel.PreferenceSourceDishSaved, row.PublicID,
					"positive", 2.5, row.ID, now,
				); err != nil {
					return err
				}
			}
			publicID = row.PublicID
			return nil
		}
		var row orderfoodModel.UserDish
		if err := tx.First(&row, "public_id = ? AND owner_id = ?", *dishID, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load dish for update")
		}
		updates := map[string]interface{}{
			"cover_file_id": cover.ID, "cover_url": cover.URL, "name": strings.TrimSpace(input.Name),
			"category_id": category.ID, "status": input.Status, "description": input.Description,
			"serving": input.Serving, "media_review_status": cover.ReviewStatus,
			"version": gorm.Expr("version + 1"), "updated_at": now,
		}
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "update dish")
		}
		if input.Status != orderfoodModel.DishStatusUsable {
			// 未确认饭局继续保留候选记录，只将源菜品失效状态固化为不可点选。
			reason := "source_deleted"
			if err := tx.Model(&orderfoodModel.FrontMealCandidate{}).
				Where("dish_id = ? AND available = ? AND meal_id IN (?)", row.ID, true,
					tx.Model(&orderfoodModel.FrontMeal{}).Select("id").
						Where("status IN ?", []orderfoodModel.MealStatus{
							orderfoodModel.MealCollecting,
							orderfoodModel.MealClosed,
						}),
				).
				Updates(map[string]interface{}{
					"available": false, "unavailable_reason": reason,
				}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "invalidate draft dish meal candidates")
			}
		}
		if err := service.persistDishChildren(tx, userID, &row, input); err != nil {
			return err
		}
		if preferenceEnabled && input.Status == orderfoodModel.DishStatusUsable {
			if err := ServiceGroupApp.PreferenceService.queueDishEvidenceTx(
				ctx, tx, userID,
				orderfoodModel.PreferenceSourceDishSaved, row.PublicID,
				"positive", 2.5, row.ID, now,
			); err != nil {
				return err
			}
		}
		publicID = row.PublicID
		return nil
	})
	if err != nil {
		return frontResponse.Dish{}, err
	}
	return service.Dish(ctx, userID, publicID)
}

// DeleteDish 删除菜品。
func (service *ContentService) DeleteDish(
	ctx context.Context,
	userID string,
	dishID string,
) (int64, int64, error) {
	db := service.database()
	var dish orderfoodModel.UserDish
	if err := db.WithContext(ctx).First(&dish, "public_id = ? AND owner_id = ?", dishID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, appErrors.FrontNotFound.DefaultMsg()
		}
		return 0, 0, appErrors.FrontInternal.Wrap(err, "load dish for delete")
	}
	var recipeCount int64
	var mealCount int64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&orderfoodModel.RecipeDish{}).Where("dish_id = ?", dish.ID).Count(&recipeCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "count dish recipe references")
		}
		if err := tx.Model(&orderfoodModel.FrontMealCandidate{}).
			Where("dish_id = ? AND meal_id IN (?)", dish.ID,
				tx.Model(&orderfoodModel.FrontMeal{}).Select("id").
					Where("status IN ?", []orderfoodModel.MealStatus{orderfoodModel.MealCollecting, orderfoodModel.MealClosed}),
			).Count(&mealCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "count dish meal references")
		}
		if err := tx.Where("dish_id = ?", dish.ID).Delete(&orderfoodModel.RecipeDish{}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "remove dish from recipes")
		}
		// 删除源菜品时保留未确认饭局的候选记录，供历史和客户端展示失效原因。
		reason := "source_deleted"
		if err := tx.Model(&orderfoodModel.FrontMealCandidate{}).
			Where("dish_id = ? AND meal_id IN (?)", dish.ID,
				tx.Model(&orderfoodModel.FrontMeal{}).Select("id").
					Where("status IN ?", []orderfoodModel.MealStatus{
						orderfoodModel.MealCollecting,
						orderfoodModel.MealClosed,
					}),
			).
			Updates(map[string]interface{}{
				"available": false, "unavailable_reason": reason,
			}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "invalidate deleted dish meal candidates")
		}
		if err := tx.Delete(&dish).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "delete dish")
		}
		return nil
	})
	return recipeCount, mealCount, err
}

// SetDiscoverable 设置是否允许被发现。
func (service *ContentService) SetDiscoverable(
	ctx context.Context,
	userID string,
	dishID string,
	discoverable bool,
) (frontResponse.Dish, error) {
	db := service.database()
	var dish orderfoodModel.UserDish
	if err := db.WithContext(ctx).First(&dish, "public_id = ? AND owner_id = ?", dishID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.Dish{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.Dish{}, appErrors.FrontInternal.Wrap(err, "load dish discoverability")
	}
	if discoverable && dish.SourceLocked {
		return frontResponse.Dish{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	var at *time.Time
	if discoverable {
		now := service.now()
		at = &now
	}
	if err := db.WithContext(ctx).Model(&dish).Updates(map[string]interface{}{
		"discoverable": discoverable, "discoverable_at": at, "version": gorm.Expr("version + 1"),
	}).Error; err != nil {
		return frontResponse.Dish{}, appErrors.FrontInternal.Wrap(err, "update dish discoverability")
	}
	return service.Dish(ctx, userID, dishID)
}

func recipeResponse(recipe orderfoodModel.UserRecipe) frontResponse.Recipe {
	result := frontResponse.Recipe{
		RecipeSummary: frontResponse.RecipeSummary{
			ID: recipe.PublicID, Name: recipe.Name, Note: recipe.Note, DishCount: len(recipe.RecipeDishes),
		},
		DishIDs: []string{}, Dishes: []frontResponse.DishSummary{}, UpdatedAt: recipe.UpdatedAt,
	}
	sort.SliceStable(recipe.RecipeDishes, func(i, j int) bool {
		return recipe.RecipeDishes[i].SortOrder < recipe.RecipeDishes[j].SortOrder
	})
	for _, relation := range recipe.RecipeDishes {
		result.DishIDs = append(result.DishIDs, relation.Dish.PublicID)
		result.Dishes = append(result.Dishes, dishSummary(relation.Dish))
		if result.CoverURL == nil && relation.Dish.CoverURL != nil {
			result.CoverURL = relation.Dish.CoverURL
		}
	}
	return result
}

func preloadRecipe(db *gorm.DB) *gorm.DB {
	return db.
		Preload("RecipeDishes", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order asc, id asc") }).
		Preload("RecipeDishes.Dish.Category").
		Preload("RecipeDishes.Dish.Tags")
}

// Recipes 查询当前用户的个人菜谱列表。
func (service *ContentService) Recipes(
	ctx context.Context,
	userID string,
) ([]frontResponse.RecipeSummary, error) {
	var rows []orderfoodModel.UserRecipe
	if err := preloadRecipe(service.database().WithContext(ctx)).
		Where("owner_id = ?", userID).
		Order("updated_at desc, id desc").
		Find(&rows).Error; err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "list recipes")
	}
	list := make([]frontResponse.RecipeSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, recipeResponse(row).RecipeSummary)
	}
	return list, nil
}

// Recipe 获取当前用户的个人菜谱及其菜品明细。
func (service *ContentService) Recipe(
	ctx context.Context,
	userID string,
	recipeID string,
) (frontResponse.Recipe, error) {
	var row orderfoodModel.UserRecipe
	err := preloadRecipe(service.database().WithContext(ctx)).
		First(&row, "public_id = ? AND owner_id = ?", recipeID, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.Recipe{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.Recipe{}, appErrors.FrontInternal.Wrap(err, "load recipe")
	}
	return recipeResponse(row), nil
}

func (service *ContentService) replaceRecipeDishes(
	tx *gorm.DB,
	userID string,
	recipe *orderfoodModel.UserRecipe,
	dishIDs []string,
) error {
	if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&orderfoodModel.RecipeDish{}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "replace recipe dishes")
	}
	for index, publicID := range dishIDs {
		var dish orderfoodModel.UserDish
		if err := tx.First(&dish, "public_id = ? AND owner_id = ? AND status = ?", publicID, userID, orderfoodModel.DishStatusUsable).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontBadRequest.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "resolve recipe dish")
		}
		relation := orderfoodModel.RecipeDish{
			PublicID: orderfoodModel.NewID(), RecipeID: recipe.ID, DishID: dish.ID,
			SortOrder: index + 1, AddedAt: service.now(),
		}
		if err := tx.Create(&relation).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create recipe dish")
		}
	}
	return nil
}

// CreateRecipe 创建菜谱。
func (service *ContentService) CreateRecipe(
	ctx context.Context,
	userID string,
	input frontRequest.RecipeCreateInput,
) (frontResponse.Recipe, error) {
	row := orderfoodModel.UserRecipe{
		PublicID: orderfoodModel.NewID(), OwnerID: userID, Name: strings.TrimSpace(input.Name),
		Note: input.Note, Version: 1,
	}
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create recipe")
		}
		return service.replaceRecipeDishes(tx, userID, &row, input.DishIDs)
	})
	if err != nil {
		return frontResponse.Recipe{}, err
	}
	return service.Recipe(ctx, userID, row.PublicID)
}

// UpdateRecipe 更新菜谱。
func (service *ContentService) UpdateRecipe(
	ctx context.Context,
	userID string,
	recipeID string,
	input frontRequest.RecipeUpdateInput,
) (frontResponse.Recipe, error) {
	db := service.database()
	var row orderfoodModel.UserRecipe
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&row, "public_id = ? AND owner_id = ?", recipeID, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load recipe for update")
		}
		updates := map[string]interface{}{"version": gorm.Expr("version + 1")}
		if input.Name != nil {
			updates["name"] = strings.TrimSpace(*input.Name)
		}
		if input.Note != nil {
			updates["note"] = input.Note
		}
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "update recipe")
		}
		if input.DishIDs != nil {
			return service.replaceRecipeDishes(tx, userID, &row, *input.DishIDs)
		}
		return nil
	})
	if err != nil {
		return frontResponse.Recipe{}, err
	}
	return service.Recipe(ctx, userID, recipeID)
}

// DeleteRecipe 删除菜谱。
func (service *ContentService) DeleteRecipe(
	ctx context.Context,
	userID string,
	recipeID string,
) error {
	db := service.database()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row orderfoodModel.UserRecipe
		if err := tx.First(&row, "public_id = ? AND owner_id = ?", recipeID, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load recipe for delete")
		}
		if err := tx.Where("recipe_id = ?", row.ID).Delete(&orderfoodModel.RecipeDish{}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "delete recipe dishes")
		}
		if err := tx.Delete(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "delete recipe")
		}
		return nil
	})
}

// AddRecipeDishes 添加菜谱菜品列表。
func (service *ContentService) AddRecipeDishes(
	ctx context.Context,
	userID string,
	recipeID string,
	dishIDs []string,
) (frontResponse.Recipe, error) {
	current, err := service.Recipe(ctx, userID, recipeID)
	if err != nil {
		return frontResponse.Recipe{}, err
	}
	seen := make(map[string]bool, len(current.DishIDs)+len(dishIDs))
	merged := make([]string, 0, len(current.DishIDs)+len(dishIDs))
	for _, id := range append(current.DishIDs, dishIDs...) {
		if !seen[id] {
			seen[id] = true
			merged = append(merged, id)
		}
	}
	return service.UpdateRecipe(ctx, userID, recipeID, frontRequest.RecipeUpdateInput{DishIDs: &merged})
}

// RemoveRecipeDish 移除菜谱菜品。
func (service *ContentService) RemoveRecipeDish(
	ctx context.Context,
	userID string,
	recipeID string,
	dishID string,
) (frontResponse.Recipe, error) {
	current, err := service.Recipe(ctx, userID, recipeID)
	if err != nil {
		return frontResponse.Recipe{}, err
	}
	filtered := make([]string, 0, len(current.DishIDs))
	found := false
	for _, id := range current.DishIDs {
		if id == dishID {
			found = true
			continue
		}
		filtered = append(filtered, id)
	}
	if !found {
		return frontResponse.Recipe{}, appErrors.FrontNotFound.DefaultMsg()
	}
	return service.UpdateRecipe(ctx, userID, recipeID, frontRequest.RecipeUpdateInput{DishIDs: &filtered})
}

// Recommendations 分页查询平台当前可用的推荐菜。
func (service *ContentService) Recommendations(
	ctx context.Context,
	userID string,
	query frontRequest.RecommendationListQuery,
) (frontResponse.Page[frontResponse.RecommendationSummary], error) {
	query.PageQuery.Defaults()
	db := service.database()
	statement := db.WithContext(ctx).Model(&orderfoodModel.PlatformRecommendation{}).
		Where("selected = ?", true)
	if query.Category != "" {
		categoryIDs := db.Model(&orderfoodModel.ContentCategory{}).Select("id").
			Where("public_id = ? OR name = ?", query.Category, query.Category)
		statement = statement.Where(
			"(source_type = ? AND dish_id IN (?)) OR (source_type = ? AND official_dish_id IN (?))",
			"creator",
			db.Model(&orderfoodModel.UserDish{}).Select("id").Where("category_id IN (?)", categoryIDs),
			"official",
			db.Model(&orderfoodModel.OfficialDish{}).Select("id").Where("category_id IN (?)", categoryIDs),
		)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.RecommendationSummary]{}, appErrors.FrontInternal.Wrap(err, "count recommendations")
	}
	var recommendations []orderfoodModel.PlatformRecommendation
	if err := statement.Order("sort_order asc, updated_at desc").
		Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).
		Find(&recommendations).Error; err != nil {
		return frontResponse.Page[frontResponse.RecommendationSummary]{}, appErrors.FrontInternal.Wrap(err, "list recommendations")
	}
	list := make([]frontResponse.RecommendationSummary, 0, len(recommendations))
	for _, recommendation := range recommendations {
		item, err := service.recommendationSummary(ctx, userID, recommendation)
		if err != nil {
			return frontResponse.Page[frontResponse.RecommendationSummary]{}, err
		}
		list = append(list, item)
	}
	return frontResponse.Page[frontResponse.RecommendationSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

func (service *ContentService) recommendationSummary(
	ctx context.Context,
	userID string,
	recommendation orderfoodModel.PlatformRecommendation,
) (frontResponse.RecommendationSummary, error) {
	var copyCount int64
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.RecommendationCopy{}).
		Where("recommendation_id = ? AND user_id = ?", recommendation.ID, userID).
		Count(&copyCount).Error; err != nil {
		return frontResponse.RecommendationSummary{}, appErrors.FrontInternal.Wrap(err, "check recommendation copy")
	}
	author := frontResponse.AuthorSummary{Type: recommendation.SourceType, Name: "平台官方"}
	var summary frontResponse.DishSummary
	switch recommendation.SourceType {
	case "creator":
		if recommendation.DishID == nil {
			return frontResponse.RecommendationSummary{}, appErrors.FrontInternal.DefaultMsg()
		}
		var dish orderfoodModel.UserDish
		if err := preloadDish(service.database().WithContext(ctx)).
			First(&dish, "id = ?", *recommendation.DishID).Error; err != nil {
			return frontResponse.RecommendationSummary{}, appErrors.FrontInternal.Wrap(err, "load creator recommendation dish")
		}
		summary = dishSummary(dish)
		id := dish.OwnerID
		author.ID = &id
		author.Name = dish.Owner.Nickname
		author.AvatarURL = dish.Owner.AvatarURL
	case "official":
		if recommendation.OfficialDishID == nil {
			return frontResponse.RecommendationSummary{}, appErrors.FrontInternal.DefaultMsg()
		}
		var dish orderfoodModel.OfficialDish
		if err := preloadOfficialDish(service.database().WithContext(ctx)).
			First(&dish, "id = ? AND status = ?", *recommendation.OfficialDishID, orderfoodModel.DishStatusUsable).Error; err != nil {
			return frontResponse.RecommendationSummary{}, appErrors.FrontInternal.Wrap(err, "load official recommendation dish")
		}
		summary = officialDishSummary(dish)
	default:
		return frontResponse.RecommendationSummary{}, appErrors.FrontInternal.DefaultMsg()
	}
	return frontResponse.RecommendationSummary{
		DishSummary: summary, RecommendationID: recommendation.ID,
		SourceType: recommendation.SourceType, Author: author, Copied: copyCount > 0,
	}, nil
}

// Recommendation 获取平台推荐菜详情。
func (service *ContentService) Recommendation(
	ctx context.Context,
	userID string,
	recommendationID string,
) (frontResponse.Recommendation, error) {
	var recommendation orderfoodModel.PlatformRecommendation
	if err := service.database().WithContext(ctx).
		First(&recommendation, "id = ? AND selected = ?", recommendationID, true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.Recommendation{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.Recommendation{}, appErrors.FrontInternal.Wrap(err, "load recommendation")
	}
	summary, err := service.recommendationSummary(ctx, userID, recommendation)
	if err != nil {
		return frontResponse.Recommendation{}, err
	}
	var detail frontResponse.Dish
	switch recommendation.SourceType {
	case "creator":
		if recommendation.DishID == nil {
			return frontResponse.Recommendation{}, appErrors.FrontInternal.DefaultMsg()
		}
		var dish orderfoodModel.UserDish
		if err := preloadDish(service.database().WithContext(ctx)).
			First(&dish, *recommendation.DishID).Error; err != nil {
			return frontResponse.Recommendation{}, appErrors.FrontInternal.Wrap(err, "load creator recommendation details")
		}
		detail = dishResponse(dish)
	case "official":
		if recommendation.OfficialDishID == nil {
			return frontResponse.Recommendation{}, appErrors.FrontInternal.DefaultMsg()
		}
		var dish orderfoodModel.OfficialDish
		if err := preloadOfficialDish(service.database().WithContext(ctx)).
			First(&dish, "id = ? AND status = ?", *recommendation.OfficialDishID, orderfoodModel.DishStatusUsable).Error; err != nil {
			return frontResponse.Recommendation{}, appErrors.FrontInternal.Wrap(err, "load official recommendation details")
		}
		detail = officialDishResponse(dish)
	default:
		return frontResponse.Recommendation{}, appErrors.FrontInternal.DefaultMsg()
	}
	return frontResponse.Recommendation{
		RecommendationSummary: summary, Description: stringValue(detail.Description),
		Ingredients: detail.Ingredients, Steps: detail.Steps,
	}, nil
}

// CopyRecommendation 复制推荐。
func (service *ContentService) CopyRecommendation(
	ctx context.Context,
	userID string,
	recommendationID string,
) (frontResponse.Dish, bool, error) {
	db := service.database()
	var copiedID string
	already := false
	preferenceEnabled := preferenceUpdatesEnabled(ctx)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing orderfoodModel.RecommendationCopy
		if err := tx.First(&existing, "recommendation_id = ? AND user_id = ?", recommendationID, userID).Error; err == nil {
			var dish orderfoodModel.UserDish
			if err := tx.First(&dish, existing.CopiedDishID).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "load copied recommendation")
			}
			copiedID = dish.PublicID
			already = true
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.FrontInternal.Wrap(err, "check copied recommendation")
		}
		var recommendation orderfoodModel.PlatformRecommendation
		if err := tx.First(&recommendation, "id = ? AND selected = ?", recommendationID, true).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load recommendation for copy")
		}
		var copied orderfoodModel.UserDish
		var sourceIngredients []orderfoodModel.DishIngredient
		var sourceSteps []orderfoodModel.DishStep
		var sourceTags []orderfoodModel.ContentTag
		operationSourceType := "recommendation"
		switch recommendation.SourceType {
		case "creator":
			if recommendation.DishID == nil {
				return appErrors.FrontInternal.DefaultMsg()
			}
			var source orderfoodModel.UserDish
			if err := preloadDish(tx).First(&source, *recommendation.DishID).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "load creator recommendation source dish")
			}
			copied = source
			copied.ID = 0
			copied.CreatedAt = time.Time{}
			copied.UpdatedAt = time.Time{}
			copied.DeletedAt = gorm.DeletedAt{}
			copied.Owner = orderfoodModel.MiniAppUser{}
			copied.SourceType = orderfoodModel.SourceTypeCreatorCopy
			copied.DirectSourceDishID = &source.ID
			rootID := source.ID
			if source.RootSourceDishID != nil {
				rootID = *source.RootSourceDishID
			}
			copied.RootSourceDishID = &rootID
			originalAuthorID := source.OwnerID
			copied.OriginalAuthorID = &originalAuthorID
			copied.ChainDepth = source.ChainDepth + 1
			sourceIngredients = source.Ingredients
			sourceSteps = source.Steps
			sourceTags = source.Tags
		case "official":
			if recommendation.OfficialDishID == nil {
				return appErrors.FrontInternal.DefaultMsg()
			}
			var source orderfoodModel.OfficialDish
			if err := preloadOfficialDish(tx).
				First(&source, "id = ? AND status = ?", *recommendation.OfficialDishID, orderfoodModel.DishStatusUsable).Error; err != nil {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			coverURL := source.CoverURL
			copied = orderfoodModel.UserDish{
				PublicID: orderfoodModel.NewID(), CoverFileID: source.CoverFileID,
				CoverURL: &coverURL, Name: source.Name, CategoryID: source.CategoryID,
				Status: orderfoodModel.DishStatusUsable, SourceType: orderfoodModel.SourceTypeOfficialCopy,
				SourceLocked: true, Description: source.Description, Serving: source.Serving,
				MediaReviewStatus:          orderfoodModel.MediaReviewNotRequired,
				DirectSourceOfficialDishID: &source.ID, RootSourceOfficialDishID: &source.ID,
				ChainDepth: 1, Version: 1,
			}
			sourceIngredients = make([]orderfoodModel.DishIngredient, 0, len(source.Ingredients))
			for _, ingredient := range source.Ingredients {
				sourceIngredients = append(sourceIngredients, orderfoodModel.DishIngredient{
					Name: ingredient.Name, Quantity: ingredient.Quantity,
					UnitID: ingredient.UnitID, Note: ingredient.Note, SortOrder: ingredient.SortOrder,
				})
			}
			sourceSteps = make([]orderfoodModel.DishStep, 0, len(source.Steps))
			for _, step := range source.Steps {
				sourceSteps = append(sourceSteps, orderfoodModel.DishStep{
					Description: step.Description, ImageURL: step.ImageURL, SortOrder: step.SortOrder,
				})
			}
			sourceTags = source.Tags
		default:
			return appErrors.FrontInternal.DefaultMsg()
		}
		copied.PublicID = orderfoodModel.NewID()
		copied.OwnerID = userID
		copied.Status = orderfoodModel.DishStatusUsable
		copied.Discoverable = false
		copied.DiscoverableAt = nil
		copied.SourceLocked = true
		copied.OperationSourceType = &operationSourceType
		copied.OperationSourceID = &recommendation.ID
		copied.Ingredients = nil
		copied.Steps = nil
		copied.Tags = nil
		if err := tx.Create(&copied).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "copy recommendation dish")
		}
		for _, ingredient := range sourceIngredients {
			ingredient.ID = 0
			ingredient.CreatedAt = time.Time{}
			ingredient.UpdatedAt = time.Time{}
			ingredient.DeletedAt = gorm.DeletedAt{}
			ingredient.DishID = copied.ID
			ingredient.Unit = nil
			if err := tx.Create(&ingredient).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "copy recommendation ingredients")
			}
		}
		for _, step := range sourceSteps {
			step.ID = 0
			step.CreatedAt = time.Time{}
			step.UpdatedAt = time.Time{}
			step.DeletedAt = gorm.DeletedAt{}
			step.DishID = copied.ID
			if err := tx.Create(&step).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "copy recommendation steps")
			}
		}
		for _, tag := range sourceTags {
			if err := tx.Create(&orderfoodModel.UserDishTag{
				DishID: copied.ID, TagID: tag.ID, CreatedAt: service.now(),
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "copy recommendation tags")
			}
		}
		if err := tx.Create(&orderfoodModel.RecommendationCopy{
			ID: orderfoodModel.NewID(), RecommendationID: recommendation.ID,
			UserID: userID, CopiedDishID: copied.ID, CreatedAt: service.now(),
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "record recommendation copy")
		}
		if preferenceEnabled {
			if err := ServiceGroupApp.PreferenceService.queueDishEvidenceTx(
				ctx, tx, userID,
				orderfoodModel.PreferenceSourceRecommendationAdopted, recommendation.ID,
				"positive", 3, copied.ID, service.now(),
			); err != nil {
				return err
			}
		}
		copiedID = copied.PublicID
		return nil
	})
	if err != nil {
		return frontResponse.Dish{}, false, err
	}
	result, err := service.Dish(ctx, userID, copiedID)
	return result, already, err
}

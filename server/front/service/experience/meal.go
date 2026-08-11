package experience

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MealService 提供饭局与采购清单业务能力。
type MealService struct {
	DB  *gorm.DB         // 业务数据库
	Now func() time.Time // 可注入的当前时间函数
}

func (service *MealService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *MealService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

func shareDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func secureToken(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func mealCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	value := make([]byte, 6)
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	for index := range value {
		value[index] = alphabet[int(random[index])%len(alphabet)]
	}
	return string(value), nil
}

func (service *MealService) closeExpired(ctx context.Context, meal *mealModel.FrontMeal) error {
	return service.closeExpiredWithSource(
		ctx,
		meal,
		mealModel.MealCloseSourceServiceGuard,
	)
}

// closeExpiredWithSource 按指定触发来源关闭到期饭局。
func (service *MealService) closeExpiredWithSource(
	ctx context.Context,
	meal *mealModel.FrontMeal,
	source mealModel.MealCloseSource,
) error {
	if meal.Status != mealModel.MealCollecting || service.now().Before(meal.DeadlineAt) {
		return nil
	}
	return service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return service.closeExpiredWithDBSource(ctx, tx, meal, source)
	})
}

// closeExpiredWithDB 在调用方事务中关闭到期饭局并生成站内通知。
func (service *MealService) closeExpiredWithDB(
	ctx context.Context,
	tx *gorm.DB,
	meal *mealModel.FrontMeal,
) error {
	return service.closeExpiredWithDBSource(
		ctx,
		tx,
		meal,
		mealModel.MealCloseSourceServiceGuard,
	)
}

// closeExpiredWithDBSource 在调用方事务中按指定来源关闭到期饭局并生成通知。
func (service *MealService) closeExpiredWithDBSource(
	ctx context.Context,
	tx *gorm.DB,
	meal *mealModel.FrontMeal,
	source mealModel.MealCloseSource,
) error {
	if meal.Status != mealModel.MealCollecting || service.now().Before(meal.DeadlineAt) {
		return nil
	}
	if source != mealModel.MealCloseSourceMinuteScan &&
		source != mealModel.MealCloseSourceServiceGuard {
		return appErrors.FrontBadRequest.DefaultMsg()
	}
	reason := "deadline"
	now := service.now()
	var current mealModel.FrontMeal
	if err := tx.WithContext(ctx).
		First(&current, "id = ?", meal.ID).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "load expired meal")
	}
	if current.Status != mealModel.MealCollecting || now.Before(current.DeadlineAt) {
		*meal = current
		return nil
	}
	closedAt := current.DeadlineAt.UTC()
	if err := tx.Model(&current).Updates(map[string]interface{}{
		"status": mealModel.MealClosed, "close_reason": reason,
		"close_source": source, "closed_at": closedAt, "updated_at": now,
	}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "close expired meal")
	}
	if err := service.writeMealClosedNotificationsTx(tx, current, now); err != nil {
		return err
	}
	current.Status = mealModel.MealClosed
	current.CloseReason = &reason
	current.CloseSource = &source
	current.ClosedAt = &closedAt
	current.UpdatedAt = now
	*meal = current
	return nil
}

// writeMealClosedNotificationsTx 为饭局全部参与者写入一次关闭点餐站内通知。
func (service *MealService) writeMealClosedNotificationsTx(
	tx *gorm.DB,
	meal mealModel.FrontMeal,
	now time.Time,
) error {
	var participants []mealModel.MealParticipant
	if err := tx.Where("meal_id = ?", meal.ID).Find(&participants).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "load expired meal participants")
	}
	targetType := "meal"
	targetID := meal.ID
	for _, participant := range participants {
		if err := tx.Create(&engagementModel.UserNotification{
			ID: commonModel.NewID(), UserID: participant.UserID,
			Type: "meal", Title: "饭局已关闭点单",
			Content:    meal.Name + "已关闭点单，可以确认最终菜单了",
			TargetType: &targetType, TargetID: &targetID, CreatedAt: now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create expired meal notification")
		}
	}
	return nil
}

// ProcessExpiredMeals 分批关闭达到点单截止时间的饭局并幂等生成站内通知。
func (service *MealService) ProcessExpiredMeals(ctx context.Context, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	var rows []mealModel.FrontMeal
	if err := service.database().WithContext(ctx).
		Where("status = ? AND deadline_at <= ?", mealModel.MealCollecting, service.now()).
		Order("deadline_at asc, id asc").Limit(batchSize).Find(&rows).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "list expired meals")
	}
	var firstErr error
	for index := range rows {
		if err := service.closeExpiredWithSource(
			ctx,
			&rows[index],
			mealModel.MealCloseSourceMinuteScan,
		); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (service *MealService) canAccess(
	ctx context.Context,
	meal mealModel.FrontMeal,
	userID string,
) error {
	if meal.CreatorID == userID {
		return nil
	}
	var count int64
	if err := service.database().WithContext(ctx).Model(&mealModel.MealParticipant{}).
		Where("meal_id = ? AND user_id = ?", meal.ID, userID).Count(&count).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "check meal participant")
	}
	if count == 0 {
		return appErrors.FrontNoPermission.DefaultMsg()
	}
	return nil
}

func (service *MealService) loadMealRow(
	ctx context.Context,
	userID string,
	mealID string,
) (mealModel.FrontMeal, error) {
	var meal mealModel.FrontMeal
	if err := service.database().WithContext(ctx).First(&meal, "id = ?", mealID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return meal, appErrors.FrontNotFound.DefaultMsg()
		}
		return meal, appErrors.FrontInternal.Wrap(err, "load meal")
	}
	if err := service.closeExpired(ctx, &meal); err != nil {
		return meal, err
	}
	if err := service.canAccess(ctx, meal, userID); err != nil {
		return meal, err
	}
	return meal, nil
}

func (service *MealService) mealSummary(
	ctx context.Context,
	meal mealModel.FrontMeal,
) (frontResponse.MealSummary, error) {
	db := service.database().WithContext(ctx)
	var participantCount int64
	var candidateCount int64
	var finalCount int64
	if err := db.Model(&mealModel.MealParticipant{}).Where("meal_id = ?", meal.ID).Count(&participantCount).Error; err != nil {
		return frontResponse.MealSummary{}, appErrors.FrontInternal.Wrap(err, "count meal participants")
	}
	if err := db.Model(&mealModel.FrontMealCandidate{}).
		Where(
			"meal_id = ? AND (unavailable_reason IS NULL OR unavailable_reason <> ?)",
			meal.ID,
			"removed",
		).
		Count(&candidateCount).Error; err != nil {
		return frontResponse.MealSummary{}, appErrors.FrontInternal.Wrap(err, "count meal candidates")
	}
	if err := db.Model(&mealModel.FrontMealFinalDish{}).Where("meal_id = ?", meal.ID).Count(&finalCount).Error; err != nil {
		return frontResponse.MealSummary{}, appErrors.FrontInternal.Wrap(err, "count meal final dishes")
	}
	var totalServings int64
	if err := db.Model(&mealModel.FrontMealFinalDish{}).Where("meal_id = ?", meal.ID).
		Select("COALESCE(SUM(final_servings), 0)").Scan(&totalServings).Error; err != nil {
		return frontResponse.MealSummary{}, appErrors.FrontInternal.Wrap(err, "sum meal servings")
	}
	var coverURL *string
	var candidate mealModel.FrontMealCandidate
	if err := db.Order("sort_order asc").First(&candidate, "meal_id = ?", meal.ID).Error; err == nil {
		var dish contentModel.UserDish
		if err := db.First(&dish, candidate.DishID).Error; err == nil {
			coverURL = dish.CoverURL
		}
	}
	var cancelReason *string
	if meal.Status == mealModel.MealCancelled {
		value := "manual"
		if meal.CancelledReason != nil && *meal.CancelledReason == "creator_disabled" {
			value = "creator_disabled"
		}
		cancelReason = &value
	}
	return frontResponse.MealSummary{
		ID: meal.ID, Name: meal.Name, Code: meal.Code, Status: string(meal.Status),
		CloseReason: meal.CloseReason, CancelReason: cancelReason, DeadlineAt: meal.DeadlineAt,
		ParticipantCount: int(participantCount), CandidateCount: int(candidateCount),
		CoverURL: coverURL, FinalDishCount: int(finalCount), TotalServings: int(totalServings),
		CreatedAt: meal.CreatedAt, ConfirmedAt: meal.ConfirmedAt, CompletedAt: meal.CompletedAt,
		CancelledAt: meal.CancelledAt,
	}, nil
}

// candidateDishAvailability 返回候选源菜品及当前可用性，并幂等固化源菜品失效状态。
func (service *MealService) candidateDishAvailability(
	db *gorm.DB,
	row *mealModel.FrontMealCandidate,
) (contentModel.UserDish, bool, error) {
	var dish contentModel.UserDish
	err := preloadDish(db.Unscoped()).First(&dish, row.DishID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dish, false, appErrors.FrontInternal.Wrap(err, "load meal candidate dish")
	}
	sourceAvailable := err == nil &&
		!dish.DeletedAt.Valid &&
		dish.Status == contentModel.DishStatusUsable
	if sourceAvailable && row.Available {
		return dish, true, nil
	}

	// 已由创建者移除的候选保留 removed 原因；这里只补齐源菜品失效状态。
	if row.Available {
		reason := "source_deleted"
		update := db.Model(&mealModel.FrontMealCandidate{}).
			Where("id = ? AND available = ?", row.ID, true).
			Updates(map[string]interface{}{
				"available": false, "unavailable_reason": reason,
			})
		if update.Error != nil {
			return dish, false, appErrors.FrontInternal.Wrap(update.Error, "invalidate meal candidate")
		}
		row.Available = false
		row.UnavailableReason = &reason
	}
	return dish, false, nil
}

func (service *MealService) candidates(
	ctx context.Context,
	mealID string,
	userID string,
	category string,
) ([]frontResponse.MealCandidate, error) {
	db := service.database().WithContext(ctx)
	statement := db.Model(&mealModel.FrontMealCandidate{}).Where("meal_id = ?", mealID)
	if category != "" {
		statement = statement.Where("dish_id IN (?)",
			db.Model(&contentModel.UserDish{}).Select("id").Where("category_id IN (?)",
				db.Model(&dishModel.ContentCategory{}).Select("id").
					Where("public_id = ? OR name = ?", category, category),
			),
		)
	}
	var rows []mealModel.FrontMealCandidate
	if err := statement.Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "list meal candidates")
	}
	result := make([]frontResponse.MealCandidate, 0, len(rows))
	for _, row := range rows {
		// 创建者主动移除的候选只保留审计记录，不再向小程序任何参与者展示。
		if row.UnavailableReason != nil && *row.UnavailableReason == "removed" {
			continue
		}
		dish, available, err := service.candidateDishAvailability(db, &row)
		if err != nil {
			return nil, err
		}
		var voteCount int64
		var selectedCount int64
		if available {
			if err := db.Model(&mealModel.FrontMealVote{}).Where("candidate_id = ?", row.ID).Count(&voteCount).Error; err != nil {
				return nil, appErrors.FrontInternal.Wrap(err, "count meal candidate votes")
			}
			if err := db.Model(&mealModel.FrontMealVote{}).
				Where("candidate_id = ? AND user_id = ?", row.ID, userID).Count(&selectedCount).Error; err != nil {
				return nil, appErrors.FrontInternal.Wrap(err, "load current meal vote")
			}
		}
		if dish.ID == 0 {
			dish.Name = "菜品已失效"
		}
		summary := dishSummary(dish)
		result = append(result, frontResponse.MealCandidate{
			ID: row.ID, DishID: dish.PublicID, Name: dish.Name, CoverURL: summary.CoverURL,
			Category: summary.Category, Tags: summary.Tags, VoteCount: int(voteCount),
			SelectedByMe: selectedCount > 0, Available: available,
			UnavailableReason: row.UnavailableReason, SortOrder: row.SortOrder,
		})
	}
	return result, nil
}

// Meal 获取饭局详情，并按当前用户身份组装候选菜和最终菜单。
func (service *MealService) Meal(
	ctx context.Context,
	userID string,
	mealID string,
) (frontResponse.Meal, error) {
	row, err := service.loadMealRow(ctx, userID, mealID)
	if err != nil {
		return frontResponse.Meal{}, err
	}
	summary, err := service.mealSummary(ctx, row)
	if err != nil {
		return frontResponse.Meal{}, err
	}
	candidates, err := service.candidates(ctx, row.ID, userID, "")
	if err != nil {
		return frontResponse.Meal{}, err
	}
	result := frontResponse.Meal{
		MealSummary: summary, Candidates: candidates, CandidateIDs: []string{},
		FinalDishes: []frontResponse.MealFinalDishSnapshot{}, ShoppingListID: row.ShoppingListID,
		CreatedByMe: row.CreatorID == userID,
	}
	if row.CancelledFrom != nil {
		value := string(*row.CancelledFrom)
		result.CancelledFromStatus = &value
	}
	if row.CancelledReason != nil && *row.CancelledReason == "creator_disabled" {
		reason := "creator_disabled"
		result.ReadOnly = true
		result.ReadOnlyReason = &reason
	}
	if !result.CreatedByMe {
		var acceptedCount int64
		if err := service.database().WithContext(ctx).
			Model(&engagementModel.MealFinalResultSubscription{}).
			Where("meal_id = ? AND user_id = ? AND accepted = ? AND consumed_at IS NULL",
				row.ID, userID, true).
			Count(&acceptedCount).Error; err != nil {
			return frontResponse.Meal{}, appErrors.FrontInternal.Wrap(
				err,
				"load active meal result subscription",
			)
		}
		result.FinalResultSubscriptionAccepted = acceptedCount > 0
	}
	for _, candidate := range candidates {
		result.CandidateIDs = append(result.CandidateIDs, candidate.ID)
	}
	var finalRows []mealModel.FrontMealFinalDish
	if err := service.database().WithContext(ctx).Where("meal_id = ?", row.ID).Order("created_at asc").Find(&finalRows).Error; err != nil {
		return frontResponse.Meal{}, appErrors.FrontInternal.Wrap(err, "load meal final dishes")
	}
	for _, final := range finalRows {
		var dish contentModel.UserDish
		if err := service.database().WithContext(ctx).Unscoped().First(&dish, final.DishID).Error; err != nil {
			return frontResponse.Meal{}, appErrors.FrontInternal.Wrap(err, "load final dish identity")
		}
		result.FinalDishes = append(result.FinalDishes, frontResponse.MealFinalDishSnapshot{
			CandidateID: final.CandidateID, DishID: dish.PublicID, Name: final.Name,
			CoverURL: final.CoverURL, FinalServing: final.FinalServings,
		})
	}
	return result, nil
}

// Create 创建饭局与采购清单。
func (service *MealService) Create(
	ctx context.Context,
	userID string,
	input frontRequest.MealCreateInput,
) (frontResponse.Meal, error) {
	deadline, err := time.Parse(time.RFC3339, input.DeadlineAt)
	if err != nil || !deadline.After(service.now()) {
		return frontResponse.Meal{}, appErrors.FrontBadRequest.DefaultMsg()
	}
	db := service.database()
	var row mealModel.FrontMeal
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务内校验创建者状态和当前进行中的饭局。
		var creator userModel.MiniAppUser
		if err := tx.
			Select("id", "status").
			First(&creator, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontLoginExpired.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal creator")
		}
		if creator.Status != userModel.UserStatusNormal {
			return appErrors.FrontUserDisabled.DefaultMsg()
		}
		var activeCount int64
		if err := tx.Model(&mealModel.FrontMeal{}).
			Where("creator_id = ? AND status IN ?", userID,
				[]mealModel.MealStatus{mealModel.MealCollecting, mealModel.MealClosed, mealModel.MealConfirmed}).
			Count(&activeCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "count active creator meals")
		}
		if activeCount > 0 {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		var code string
		for attempt := 0; attempt < 8; attempt++ {
			code, err = mealCode()
			if err != nil {
				return appErrors.FrontInternal.Wrap(err, "generate meal code")
			}
			var count int64
			if err := tx.Model(&mealModel.FrontMeal{}).Where("code = ?", code).Count(&count).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "check meal code")
			}
			if count == 0 {
				break
			}
		}
		if code == "" {
			return appErrors.FrontInternal.DefaultMsg()
		}
		now := service.now()
		row = mealModel.FrontMeal{
			ID: commonModel.NewID(), CreatorID: userID, Name: strings.TrimSpace(input.Name),
			Code: code, Status: mealModel.MealCollecting, DeadlineAt: deadline.UTC(),
			SourceRecipeID: input.SourceRecipeID, CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create meal")
		}
		if err := tx.Create(&mealModel.MealParticipant{
			ID: commonModel.NewID(), MealID: row.ID, UserID: userID, JoinedAt: now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "join creator to meal")
		}
		for index, dishID := range input.CandidateDishIDs {
			var dish contentModel.UserDish
			if err := tx.First(&dish,
				"public_id = ? AND owner_id = ? AND status = ?", dishID, userID, contentModel.DishStatusUsable,
			).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return appErrors.FrontBadRequest.DefaultMsg()
				}
				return appErrors.FrontInternal.Wrap(err, "resolve meal candidate dish")
			}
			if err := tx.Create(&mealModel.FrontMealCandidate{
				ID: commonModel.NewID(), MealID: row.ID, DishID: dish.ID,
				Available: true, SortOrder: index + 1, CreatedAt: now,
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create meal candidate")
			}
		}
		return nil
	})
	if err != nil {
		return frontResponse.Meal{}, err
	}
	return service.Meal(ctx, userID, row.ID)
}

// List 分页查询饭局与采购清单。
func (service *MealService) List(
	ctx context.Context,
	userID string,
	query frontRequest.MealListQuery,
) (frontResponse.Page[frontResponse.MealSummary], error) {
	query.PageQuery.Defaults()
	statuses := []mealModel.MealStatus{mealModel.MealCollecting, mealModel.MealClosed, mealModel.MealConfirmed}
	if query.Scope == "history" {
		statuses = []mealModel.MealStatus{mealModel.MealCompleted, mealModel.MealCancelled}
	}
	db := service.database().WithContext(ctx)
	statement := db.Model(&mealModel.FrontMeal{}).
		Where("status IN ? AND (creator_id = ? OR id IN (?))", statuses, userID,
			db.Model(&mealModel.MealParticipant{}).Select("meal_id").Where("user_id = ?", userID))
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return frontResponse.Page[frontResponse.MealSummary]{}, appErrors.FrontInternal.Wrap(err, "count meals")
	}
	var rows []mealModel.FrontMeal
	if err := statement.Order("created_at desc, id desc").
		Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error; err != nil {
		return frontResponse.Page[frontResponse.MealSummary]{}, appErrors.FrontInternal.Wrap(err, "list meals")
	}
	list := make([]frontResponse.MealSummary, 0, len(rows))
	for index := range rows {
		if err := service.closeExpired(ctx, &rows[index]); err != nil {
			return frontResponse.Page[frontResponse.MealSummary]{}, err
		}
		summary, err := service.mealSummary(ctx, rows[index])
		if err != nil {
			return frontResponse.Page[frontResponse.MealSummary]{}, err
		}
		list = append(list, summary)
	}
	return frontResponse.Page[frontResponse.MealSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// Current 获取当前饭局与采购清单。
func (service *MealService) Current(
	ctx context.Context,
	userID string,
) (*frontResponse.Meal, error) {
	var row mealModel.FrontMeal
	err := service.database().WithContext(ctx).
		Where("creator_id = ? AND status IN ?", userID,
			[]mealModel.MealStatus{mealModel.MealCollecting, mealModel.MealClosed, mealModel.MealConfirmed}).
		Order("created_at desc").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "load current meal")
	}
	result, err := service.Meal(ctx, userID, row.ID)
	return &result, err
}

// Lookup 使用点餐码查询可加入的饭局。
func (service *MealService) Lookup(
	ctx context.Context,
	code string,
) (frontResponse.MealJoinPreview, error) {
	var row mealModel.FrontMeal
	if err := service.database().WithContext(ctx).First(&row, "code = ?", strings.ToUpper(code)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.MealJoinPreview{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.MealJoinPreview{}, appErrors.FrontInternal.Wrap(err, "lookup meal")
	}
	if err := service.closeExpired(ctx, &row); err != nil {
		return frontResponse.MealJoinPreview{}, err
	}
	summary, err := service.mealSummary(ctx, row)
	if err != nil {
		return frontResponse.MealJoinPreview{}, err
	}
	return frontResponse.MealJoinPreview{
		ID: row.ID, Name: row.Name, Status: string(row.Status), DeadlineAt: row.DeadlineAt,
		ParticipantCount: summary.ParticipantCount, CandidateCount: summary.CandidateCount,
	}, nil
}

// Join 使用点餐码加入饭局，并保证同一用户不会重复加入。
func (service *MealService) Join(
	ctx context.Context,
	userID string,
	code string,
) (frontResponse.Meal, error) {
	var row mealModel.FrontMeal
	db := service.database()
	closedByDeadline := false
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			First(&row, "code = ?", strings.ToUpper(code)).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal to join")
		}
		wasCollecting := row.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &row); err != nil {
			return err
		}
		if wasCollecting && row.Status == mealModel.MealClosed &&
			row.CloseReason != nil && *row.CloseReason == "deadline" {
			closedByDeadline = true
			return nil
		}
		if row.Status != mealModel.MealCollecting {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		participant := mealModel.MealParticipant{
			ID: commonModel.NewID(), MealID: row.ID, UserID: userID, JoinedAt: service.now(),
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&participant)
		if result.Error != nil {
			return appErrors.FrontInternal.Wrap(result.Error, "join meal")
		}
		return nil
	})
	if err != nil {
		return frontResponse.Meal{}, err
	}
	if closedByDeadline {
		return frontResponse.Meal{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	return service.Meal(ctx, userID, row.ID)
}

// RecordFinalResultSubscription 登记参与者主动申请的一次饭局最终结果微信订阅授权。
func (service *MealService) RecordFinalResultSubscription(
	ctx context.Context,
	userID string,
	mealID string,
	wechatTemplateID string,
	authorizationResult string,
) (frontResponse.MealFinalResultSubscription, error) {
	switch authorizationResult {
	case engagementModel.MealAuthorizationAccept,
		engagementModel.MealAuthorizationReject,
		engagementModel.MealAuthorizationBan:
	default:
		return frontResponse.MealFinalResultSubscription{}, appErrors.FrontBadRequest.DefaultMsg()
	}

	wechatTemplateID = strings.TrimSpace(wechatTemplateID)
	if wechatTemplateID == "" {
		return frontResponse.MealFinalResultSubscription{}, appErrors.FrontBadRequest.DefaultMsg()
	}

	db := service.database().WithContext(ctx)
	var result frontResponse.MealFinalResultSubscription
	closedByDeadline := false
	err := db.Transaction(func(tx *gorm.DB) error {
		var meal mealModel.FrontMeal
		if err := tx.
			First(&meal, "id = ?", mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal subscription target")
		}
		wasCollecting := meal.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &meal); err != nil {
			return err
		}
		if wasCollecting && meal.Status == mealModel.MealClosed &&
			meal.CloseReason != nil && *meal.CloseReason == "deadline" {
			closedByDeadline = true
			return nil
		}
		if meal.CreatorID == userID {
			return appErrors.FrontNoPermission.DefaultMsg()
		}
		if meal.Status != mealModel.MealCollecting && meal.Status != mealModel.MealClosed {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		var participantCount int64
		if err := tx.Model(&mealModel.MealParticipant{}).
			Where("meal_id = ? AND user_id = ?", mealID, userID).
			Count(&participantCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "check meal subscription participant")
		}
		if participantCount == 0 {
			return appErrors.FrontNoPermission.DefaultMsg()
		}
		var template engagementModel.SubscribeMessageTemplate
		if err := tx.Where("wechat_template_id = ? AND scene = ? AND enabled = ?",
			wechatTemplateID, engagementModel.SubscribeSceneMealStatus, true).
			First(&template).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontStateConflict.New("本次订阅模板已停用，请刷新页面后重试")
			}
			return appErrors.FrontInternal.Wrap(err, "load authorized meal subscription template")
		}

		now := service.now()
		accepted := authorizationResult == engagementModel.MealAuthorizationAccept
		if accepted {
			var existing engagementModel.MealFinalResultSubscription
			if err := tx.Where(
				"meal_id = ? AND user_id = ? AND accepted = ? AND consumed_at IS NULL",
				mealID,
				userID,
				true,
			).Order("recorded_at asc, id asc").First(&existing).Error; err == nil {
				result = frontResponse.MealFinalResultSubscription{
					MealID: mealID, AuthorizationResult: existing.AuthorizationResult,
					Accepted: true, RecordedAt: existing.RecordedAt,
				}
				return nil
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontInternal.Wrap(err, "load active meal result subscription")
			}
		}

		row := engagementModel.MealFinalResultSubscription{
			ID: commonModel.NewID(), MealID: mealID, UserID: userID,
			TemplateID: template.ID, AuthorizationResult: authorizationResult,
			Accepted: accepted, RecordedAt: now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "record meal result subscription")
		}
		result = frontResponse.MealFinalResultSubscription{
			MealID: mealID, AuthorizationResult: authorizationResult,
			Accepted: accepted, RecordedAt: now,
		}
		return nil
	})
	if err != nil {
		return frontResponse.MealFinalResultSubscription{}, err
	}
	if closedByDeadline {
		return service.RecordFinalResultSubscription(
			ctx,
			userID,
			mealID,
			wechatTemplateID,
			authorizationResult,
		)
	}
	return result, nil
}

// writeMealFinalResultNotifications 在最终状态事务内写站内通知、消费授权并创建待发送记录。
func (service *MealService) writeMealFinalResultNotifications(
	tx *gorm.DB,
	meal mealModel.FrontMeal,
	resultStatus mealModel.MealStatus,
	actorID string,
	allowWechatSubscription bool,
) error {
	var participants []mealModel.MealParticipant
	if err := tx.Where("meal_id = ?", meal.ID).Order("joined_at asc").Find(&participants).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "load meal final notification participants")
	}

	title := "饭局菜单已确认"
	content := meal.Name + "的最终菜单已经确认"
	if resultStatus == mealModel.MealCancelled {
		title = "饭局已取消"
		content = meal.Name + "已取消"
	}
	targetType := "meal"
	targetID := meal.ID
	targetPage := "/pages/meal/history?mealId=" + meal.ID
	now := service.now()

	for _, participant := range participants {
		// 操作者已经知道结果，不重复向本人生成通知或外部提醒。
		if participant.UserID == actorID {
			continue
		}
		var subscription engagementModel.MealFinalResultSubscription
		subscriptionErr := tx.
			Where("meal_id = ? AND user_id = ? AND accepted = ? AND consumed_at IS NULL",
				meal.ID, participant.UserID, true).
			Order("recorded_at asc, id asc").
			First(&subscription).Error
		if subscriptionErr != nil && !errors.Is(subscriptionErr, gorm.ErrRecordNotFound) {
			return appErrors.FrontInternal.Wrap(subscriptionErr, "load meal final subscription")
		}

		var subscribeLogID *string
		subscribeRequired := false
		if subscriptionErr == nil {
			consumedResult := string(resultStatus)
			if err := tx.Model(&engagementModel.MealFinalResultSubscription{}).
				Where("meal_id = ? AND user_id = ? AND accepted = ? AND consumed_at IS NULL",
					meal.ID, participant.UserID, true).
				Updates(map[string]interface{}{
					"consumed_result": consumedResult, "consumed_at": now,
				}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "consume meal final subscriptions")
			}

			var enabledTemplateCount int64
			if allowWechatSubscription {
				if err := tx.Model(&engagementModel.SubscribeMessageTemplate{}).
					Where("id = ? AND scene = ? AND enabled = ?",
						subscription.TemplateID, engagementModel.SubscribeSceneMealStatus, true).
					Count(&enabledTemplateCount).Error; err != nil {
					return appErrors.FrontInternal.Wrap(err, "validate meal subscription template")
				}
			}
			if allowWechatSubscription && enabledTemplateCount == 1 {
				payload, err := json.Marshal(map[string]string{
					"mealId": meal.ID, "mealName": meal.Name,
					"result": string(resultStatus), "resultAt": now.Format(time.RFC3339),
				})
				if err != nil {
					return appErrors.FrontInternal.Wrap(err, "encode meal subscription payload summary")
				}
				logID := commonModel.NewID()
				if err := tx.Create(&engagementModel.SubscribeMessageLog{
					ID: logID, UserID: participant.UserID, TemplateID: subscription.TemplateID,
					SubscriptionID: subscription.ID, Scene: engagementModel.SubscribeSceneMealStatus,
					Status: engagementModel.SubscribeLogPending, RequestID: commonModel.NewID(),
					AuthorizationChecked: true, AuthorizationAvailable: true,
					TargetPage: &targetPage, SafePayloadSummary: datatypes.JSON(payload),
					CreatedAt: now,
				}).Error; err != nil {
					return appErrors.FrontInternal.Wrap(err, "create meal subscription send log")
				}
				subscribeLogID = &logID
				subscribeRequired = true
			}
		}

		if err := tx.Create(&engagementModel.UserNotification{
			ID: commonModel.NewID(), UserID: participant.UserID,
			Type: "meal", Title: title, Content: content,
			TargetType: &targetType, TargetID: &targetID,
			SubscribeRequired: subscribeRequired, SubscribeLogID: subscribeLogID,
			CreatedAt: now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create meal final notification")
		}
	}
	return nil
}

func (service *MealService) transition(
	ctx context.Context,
	userID string,
	mealID string,
	from []mealModel.MealStatus,
	to mealModel.MealStatus,
	extra map[string]interface{},
) (frontResponse.Meal, error) {
	db := service.database()
	var row mealModel.FrontMeal
	if err := db.WithContext(ctx).First(&row, "id = ?", mealID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.Meal{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.Meal{}, appErrors.FrontInternal.Wrap(err, "load meal transition")
	}
	if row.CreatorID != userID {
		return frontResponse.Meal{}, appErrors.FrontNoPermission.DefaultMsg()
	}
	valid := false
	for _, status := range from {
		if row.Status == status {
			valid = true
		}
	}
	if !valid {
		return frontResponse.Meal{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	updates := map[string]interface{}{"status": to, "updated_at": service.now()}
	for key, value := range extra {
		updates[key] = value
	}
	result := db.WithContext(ctx).Model(&mealModel.FrontMeal{}).
		Where("id = ? AND status = ?", mealID, row.Status).Updates(updates)
	if result.Error != nil {
		return frontResponse.Meal{}, appErrors.FrontInternal.Wrap(result.Error, "transition meal")
	}
	if result.RowsAffected != 1 {
		return frontResponse.Meal{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	return service.Meal(ctx, userID, mealID)
}

// Close 关闭饭局点单，饭局仍保持进行中以等待确认采购。
func (service *MealService) Close(ctx context.Context, userID, mealID string) (frontResponse.Meal, error) {
	db := service.database()
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meal mealModel.FrontMeal
		if err := tx.
			First(&meal, "id = ?", mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal for close")
		}
		if meal.CreatorID != userID {
			return appErrors.FrontNoPermission.DefaultMsg()
		}
		wasCollecting := meal.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &meal); err != nil {
			return err
		}
		if wasCollecting && meal.Status == mealModel.MealClosed {
			return nil
		}
		if meal.Status != mealModel.MealCollecting {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		reason := "manual"
		source := mealModel.MealCloseSourceCreatorAction
		now := service.now()
		if err := tx.Model(&meal).Updates(map[string]interface{}{
			"status": mealModel.MealClosed, "close_reason": reason,
			"close_source": source, "closed_at": now, "updated_at": now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "close meal manually")
		}
		return service.writeMealClosedNotificationsTx(tx, meal, now)
	})
	if err != nil {
		return frontResponse.Meal{}, err
	}
	return service.Meal(ctx, userID, mealID)
}

// Cancel 取消饭局并终止后续点单、确认和采购流程。
func (service *MealService) Cancel(
	ctx context.Context,
	userID string,
	mealID string,
	reason *string,
) (frontResponse.Meal, error) {
	db := service.database()
	now := service.now()
	closedByDeadline := false
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meal mealModel.FrontMeal
		if err := tx.
			First(&meal, "id = ?", mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal for cancel")
		}
		if meal.CreatorID != userID {
			return appErrors.FrontNoPermission.DefaultMsg()
		}
		wasCollecting := meal.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &meal); err != nil {
			return err
		}
		if wasCollecting && meal.Status == mealModel.MealClosed &&
			meal.CloseReason != nil && *meal.CloseReason == "deadline" {
			closedByDeadline = true
			return nil
		}
		if meal.Status != mealModel.MealCollecting && meal.Status != mealModel.MealClosed {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		cancelledFrom := meal.Status
		if err := tx.Model(&meal).Updates(map[string]interface{}{
			"status": mealModel.MealCancelled, "cancelled_at": now,
			"cancelled_reason": reason, "cancelled_from_status": cancelledFrom,
			"updated_at": now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "cancel meal")
		}
		meal.Status = mealModel.MealCancelled
		meal.CancelledAt = &now
		meal.CancelledReason = reason
		meal.CancelledFrom = &cancelledFrom
		return service.writeMealFinalResultNotifications(
			tx, meal, mealModel.MealCancelled, userID, true,
		)
	})
	if err != nil {
		return frontResponse.Meal{}, err
	}
	if closedByDeadline {
		return service.Cancel(ctx, userID, mealID, reason)
	}
	return service.Meal(ctx, userID, mealID)
}

// CandidateList 返回饭局候选菜及当前用户的点选状态。
func (service *MealService) CandidateList(
	ctx context.Context,
	userID string,
	mealID string,
	category string,
) ([]frontResponse.CategoryCount, []frontResponse.MealCandidate, error) {
	if _, err := service.loadMealRow(ctx, userID, mealID); err != nil {
		return nil, nil, err
	}
	all, err := service.candidates(ctx, mealID, userID, "")
	if err != nil {
		return nil, nil, err
	}
	counts := make(map[string]int)
	for _, candidate := range all {
		counts[candidate.Category]++
	}
	categories := make([]frontResponse.CategoryCount, 0, len(counts))
	for name, count := range counts {
		categories = append(categories, frontResponse.CategoryCount{Category: name, Count: count})
	}
	filtered := all
	if category != "" {
		filtered = []frontResponse.MealCandidate{}
		for _, candidate := range all {
			if candidate.Category == category {
				filtered = append(filtered, candidate)
			}
		}
	}
	return categories, filtered, nil
}

// RemoveCandidate 由饭局创建者在收集阶段移除候选菜并保留原点选记录。
func (service *MealService) RemoveCandidate(
	ctx context.Context,
	userID string,
	mealID string,
	candidateID string,
) (time.Time, error) {
	var removedAt time.Time
	closedByDeadline := false
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meal mealModel.FrontMeal
		if err := tx.
			First(&meal, "id = ?", mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal for candidate removal")
		}
		if meal.CreatorID != userID {
			return appErrors.FrontNoPermission.DefaultMsg()
		}
		wasCollecting := meal.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &meal); err != nil {
			return err
		}
		if wasCollecting && meal.Status == mealModel.MealClosed &&
			meal.CloseReason != nil && *meal.CloseReason == "deadline" {
			closedByDeadline = true
			return nil
		}
		if meal.Status != mealModel.MealCollecting {
			return appErrors.FrontStateConflict.DefaultMsg()
		}

		var candidate mealModel.FrontMealCandidate
		if err := tx.
			First(&candidate, "id = ? AND meal_id = ?", candidateID, mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal candidate for removal")
		}
		if !candidate.Available {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		var remainingCount int64
		if err := tx.Model(&mealModel.FrontMealCandidate{}).
			Where(
				"meal_id = ? AND id <> ? AND available = ?",
				mealID,
				candidateID,
				true,
			).
			Count(&remainingCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "count remaining meal candidates")
		}
		if remainingCount == 0 {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		reason := "removed"
		removedAt = service.now()
		if err := tx.Model(&candidate).Updates(map[string]interface{}{
			"available":          false,
			"unavailable_reason": reason,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "remove meal candidate")
		}
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	if closedByDeadline {
		return time.Time{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	return removedAt, nil
}

// MyVotes 获取当前用户在饭局中的点选结果。
func (service *MealService) MyVotes(
	ctx context.Context,
	userID string,
	mealID string,
) ([]string, *time.Time, error) {
	if _, err := service.loadMealRow(ctx, userID, mealID); err != nil {
		return nil, nil, err
	}
	var rows []mealModel.FrontMealVote
	if err := service.database().WithContext(ctx).
		Joins("JOIN of_meal_candidates AS candidate ON candidate.id = of_meal_votes.candidate_id AND candidate.available = ?", true).
		Joins("JOIN of_dishes AS dish ON dish.id = candidate.dish_id AND dish.deleted_at IS NULL AND dish.status = ?", contentModel.DishStatusUsable).
		Where("of_meal_votes.meal_id = ? AND of_meal_votes.user_id = ?", mealID, userID).
		Order("of_meal_votes.updated_at asc").Find(&rows).Error; err != nil {
		return nil, nil, appErrors.FrontInternal.Wrap(err, "load meal votes")
	}
	ids := make([]string, 0, len(rows))
	var updatedAt *time.Time
	for _, row := range rows {
		ids = append(ids, row.CandidateID)
		value := row.UpdatedAt
		if updatedAt == nil || value.After(*updatedAt) {
			updatedAt = &value
		}
	}
	return ids, updatedAt, nil
}

// SaveVotes 覆盖保存当前用户的饭局点选结果。
func (service *MealService) SaveVotes(
	ctx context.Context,
	userID string,
	mealID string,
	candidateIDs []string,
) ([]string, time.Time, error) {
	preferenceEnabled := preferenceUpdatesEnabled(ctx, userID)
	var savedAt time.Time
	closedByDeadline := false
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meal mealModel.FrontMeal
		if err := tx.
			First(&meal, "id = ?", mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal for vote")
		}
		if meal.CreatorID != userID {
			var participantCount int64
			if err := tx.Model(&mealModel.MealParticipant{}).
				Where("meal_id = ? AND user_id = ?", mealID, userID).
				Count(&participantCount).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "check meal vote participant")
			}
			if participantCount == 0 {
				return appErrors.FrontNoPermission.DefaultMsg()
			}
		}
		wasCollecting := meal.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &meal); err != nil {
			return err
		}
		if wasCollecting && meal.Status == mealModel.MealClosed &&
			meal.CloseReason != nil && *meal.CloseReason == "deadline" {
			closedByDeadline = true
			return nil
		}
		if meal.Status != mealModel.MealCollecting {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		savedAt = service.now()
		var validCount int64
		if len(candidateIDs) > 0 {
			if err := tx.Model(&mealModel.FrontMealCandidate{}).
				Joins("JOIN of_dishes AS dish ON dish.id = of_meal_candidates.dish_id AND dish.deleted_at IS NULL AND dish.status = ?", contentModel.DishStatusUsable).
				Where("of_meal_candidates.meal_id = ? AND of_meal_candidates.id IN ? AND of_meal_candidates.available = ?", mealID, candidateIDs, true).
				Count(&validCount).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "validate meal vote candidates")
			}
			if validCount != int64(len(candidateIDs)) {
				return appErrors.FrontBadRequest.DefaultMsg()
			}
		}
		// 覆盖保存只替换当前仍可点选候选的记录；已移除或源菜失效的旧点选保留用于历史解释。
		validCandidateIDs := tx.Model(&mealModel.FrontMealCandidate{}).
			Select("of_meal_candidates.id").
			Joins(
				"JOIN of_dishes AS dish ON dish.id = of_meal_candidates.dish_id "+
					"AND dish.deleted_at IS NULL AND dish.status = ?",
				contentModel.DishStatusUsable,
			).
			Where(
				"of_meal_candidates.meal_id = ? AND of_meal_candidates.available = ?",
				mealID,
				true,
			)
		if err := tx.Where(
			"meal_id = ? AND user_id = ? AND candidate_id IN (?)",
			mealID,
			userID,
			validCandidateIDs,
		).Delete(&mealModel.FrontMealVote{}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "replace meal votes")
		}
		for _, candidateID := range candidateIDs {
			if err := tx.Create(&mealModel.FrontMealVote{
				ID: commonModel.NewID(), MealID: mealID, UserID: userID,
				CandidateID: candidateID, CreatedAt: savedAt, UpdatedAt: savedAt,
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create meal vote")
			}
			if preferenceEnabled {
				var candidate mealModel.FrontMealCandidate
				if err := tx.First(&candidate, "id = ? AND meal_id = ?", candidateID, mealID).Error; err != nil {
					return appErrors.FrontInternal.Wrap(err, "load preference vote candidate")
				}
				if err := ServiceGroupApp.PreferenceService.queueDishEvidenceTx(
					ctx, tx, userID,
					userModel.PreferenceSourceDishOrdered, mealID+":"+candidateID,
					"positive", 2.5, candidate.DishID, savedAt,
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, time.Time{}, err
	}
	if closedByDeadline {
		return nil, time.Time{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	return candidateIDs, savedAt, nil
}

// Stats 获取饭局候选菜点选统计。
func (service *MealService) Stats(
	ctx context.Context,
	userID string,
	mealID string,
) (frontResponse.MealSummary, []frontResponse.MealDishStat, error) {
	meal, err := service.loadMealRow(ctx, userID, mealID)
	if err != nil {
		return frontResponse.MealSummary{}, nil, err
	}
	summary, err := service.mealSummary(ctx, meal)
	if err != nil {
		return frontResponse.MealSummary{}, nil, err
	}
	candidates, err := service.candidates(ctx, mealID, userID, "")
	if err != nil {
		return frontResponse.MealSummary{}, nil, err
	}
	finalByCandidate := map[string]int{}
	var finals []mealModel.FrontMealFinalDish
	if err := service.database().WithContext(ctx).Where("meal_id = ?", mealID).Find(&finals).Error; err != nil {
		return frontResponse.MealSummary{}, nil, appErrors.FrontInternal.Wrap(err, "load final meal stats")
	}
	for _, final := range finals {
		finalByCandidate[final.CandidateID] = final.FinalServings
	}
	stats := make([]frontResponse.MealDishStat, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.Available {
			continue
		}
		var dish contentModel.UserDish
		if err := service.database().WithContext(ctx).First(&dish, "public_id = ?", candidate.DishID).Error; err != nil {
			return frontResponse.MealSummary{}, nil, appErrors.FrontInternal.Wrap(err, "load meal stat dish")
		}
		suggested := dish.Serving
		if candidate.VoteCount > suggested {
			suggested = candidate.VoteCount
		}
		final := finalByCandidate[candidate.ID]
		stats = append(stats, frontResponse.MealDishStat{
			CandidateID: candidate.ID, DishID: candidate.DishID, Name: candidate.Name,
			CoverURL: candidate.CoverURL, Category: candidate.Category, Tags: candidate.Tags,
			VoteCount: candidate.VoteCount, BaseServing: dish.Serving, SuggestedServing: suggested,
			Selected: final > 0, FinalServing: final,
		})
	}
	return summary, stats, nil
}

type ingredientSnapshot struct {
	Name     string  `json:"name"`
	Quantity string  `json:"quantity"`
	Unit     string  `json:"unit"`
	Note     *string `json:"note"`
}

type mealDishStepSnapshot struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	ImageURL    *string `json:"imageUrl,omitempty"`
	SortOrder   int     `json:"sortOrder"`
}

func shoppingAmount(quantity string, baseServing int, finalServing int) string {
	value, err := strconv.ParseFloat(strings.TrimSpace(quantity), 64)
	if err != nil || baseServing <= 0 {
		return quantity
	}
	scaled := value * float64(finalServing) / float64(baseServing)
	if math.Abs(scaled-math.Round(scaled)) < 0.0001 {
		return strconv.FormatInt(int64(math.Round(scaled)), 10)
	}
	return strconv.FormatFloat(scaled, 'f', 1, 64)
}

func mergeShoppingAmount(left, right string) string {
	lv, leftErr := strconv.ParseFloat(left, 64)
	rv, rightErr := strconv.ParseFloat(right, 64)
	if leftErr != nil || rightErr != nil {
		if left == right {
			return left
		}
		return left + " + " + right
	}
	total := lv + rv
	if math.Abs(total-math.Round(total)) < 0.0001 {
		return strconv.FormatInt(int64(math.Round(total)), 10)
	}
	return strconv.FormatFloat(total, 'f', 1, 64)
}

// Confirm 冻结最终菜单快照并生成采购清单。
func (service *MealService) Confirm(
	ctx context.Context,
	userID string,
	mealID string,
	input frontRequest.MealConfirmInput,
) (frontResponse.Meal, string, error) {
	db := service.database()
	shoppingListID := ""
	closedByDeadline := false
	// 在同一事务内冻结菜单快照并生成采购清单。
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meal mealModel.FrontMeal
		if err := tx.First(&meal, "id = ?", mealID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load meal for confirm")
		}
		if meal.CreatorID != userID {
			return appErrors.FrontNoPermission.DefaultMsg()
		}
		wasCollecting := meal.Status == mealModel.MealCollecting
		if err := service.closeExpiredWithDB(ctx, tx, &meal); err != nil {
			return err
		}
		if wasCollecting && meal.Status == mealModel.MealClosed &&
			meal.CloseReason != nil && *meal.CloseReason == "deadline" {
			closedByDeadline = true
			return nil
		}
		if meal.Status != mealModel.MealClosed {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		selectedCount := 0
		now := service.now()
		merged := map[string]*mealModel.FrontShoppingItem{}
		sourceDishNames := map[string][]string{}
		for _, inputDish := range input.Dishes {
			if inputDish.Selected == nil || !*inputDish.Selected {
				continue
			}
			if inputDish.FinalServing < 1 {
				return appErrors.FrontBadRequest.DefaultMsg()
			}
			var candidate mealModel.FrontMealCandidate
			if err := tx.First(&candidate,
				"id = ? AND meal_id = ? AND available = ?", inputDish.CandidateID, mealID, true,
			).Error; err != nil {
				return appErrors.FrontBadRequest.DefaultMsg()
			}
			var dish contentModel.UserDish
			if err := preloadDish(tx).First(
				&dish,
				"id = ? AND status = ?",
				candidate.DishID,
				contentModel.DishStatusUsable,
			).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return appErrors.FrontBadRequest.DefaultMsg()
				}
				return appErrors.FrontInternal.Wrap(err, "load final menu dish")
			}
			ingredients := make([]ingredientSnapshot, 0, len(dish.Ingredients))
			for _, ingredient := range dish.Ingredients {
				unit := ""
				if ingredient.Unit != nil {
					unit = ingredient.Unit.Name
				}
				quantity := shoppingAmount(stringValue(ingredient.Quantity), dish.Serving, inputDish.FinalServing)
				ingredients = append(ingredients, ingredientSnapshot{
					Name: ingredient.Name, Quantity: quantity, Unit: unit, Note: ingredient.Note,
				})
				key := strings.TrimSpace(ingredient.Name) + "\x00" + strings.TrimSpace(unit)
				amount := strings.TrimSpace(quantity + " " + unit)
				if existing := merged[key]; existing != nil {
					existing.Amount = strings.TrimSpace(mergeShoppingAmount(
						strings.TrimSpace(strings.TrimSuffix(existing.Amount, unit)),
						quantity,
					) + " " + unit)
				} else {
					merged[key] = &mealModel.FrontShoppingItem{
						ID: commonModel.NewID(), Name: ingredient.Name, Amount: amount,
						Note: ingredient.Note, SortOrder: len(merged) + 1, CreatedAt: now, UpdatedAt: now,
					}
				}
				foundSource := false
				for _, sourceName := range sourceDishNames[key] {
					if sourceName == dish.Name {
						foundSource = true
						break
					}
				}
				if !foundSource {
					sourceDishNames[key] = append(sourceDishNames[key], dish.Name)
				}
			}
			encoded, err := json.Marshal(ingredients)
			if err != nil {
				return appErrors.FrontInternal.Wrap(err, "encode final dish ingredients")
			}
			stepSnapshots := make([]mealDishStepSnapshot, 0, len(dish.Steps))
			for _, step := range dish.Steps {
				stepSnapshots = append(stepSnapshots, mealDishStepSnapshot{
					ID: commonModel.NewID(), Description: step.Description,
					ImageURL: step.ImageURL, SortOrder: step.SortOrder,
				})
			}
			encodedSteps, err := json.Marshal(stepSnapshots)
			if err != nil {
				return appErrors.FrontInternal.Wrap(err, "encode final dish steps")
			}
			if err := tx.Create(&mealModel.FrontMealFinalDish{
				ID: commonModel.NewID(), MealID: mealID, CandidateID: candidate.ID,
				DishID: dish.ID, Name: dish.Name, CoverURL: dish.CoverURL,
				FinalServings: inputDish.FinalServing, Ingredients: datatypes.JSON(encoded),
				Steps: datatypes.JSON(encodedSteps), CreatedAt: now,
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create final meal dish")
			}
			selectedCount++
		}
		if selectedCount == 0 {
			return appErrors.FrontBadRequest.DefaultMsg()
		}
		shareToken, err := secureToken(32)
		if err != nil {
			return appErrors.FrontInternal.Wrap(err, "create shopping share token")
		}
		shoppingListID = commonModel.NewID()
		list := mealModel.FrontShoppingList{
			ID: shoppingListID, MealID: mealID, OwnerID: userID,
			ShareToken: shareToken, ShareTokenHash: shareDigest(shareToken),
			CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.Create(&list).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create shopping list")
		}
		for key, item := range merged {
			item.ListID = list.ID
			encodedSourceNames, err := json.Marshal(sourceDishNames[key])
			if err != nil {
				return appErrors.FrontInternal.Wrap(err, "encode shopping item source dishes")
			}
			item.SourceDishNames = datatypes.JSON(encodedSourceNames)
			if err := tx.Create(item).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "create shopping item")
			}
		}
		if err := tx.Model(&meal).Updates(map[string]interface{}{
			"status": mealModel.MealConfirmed, "confirmed_at": now,
			"shopping_list_id": list.ID, "updated_at": now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "confirm meal")
		}
		meal.Status = mealModel.MealConfirmed
		meal.ConfirmedAt = &now
		meal.ShoppingListID = &list.ID
		return service.writeMealFinalResultNotifications(
			tx, meal, mealModel.MealConfirmed, userID, true,
		)
	})
	if err != nil {
		return frontResponse.Meal{}, "", err
	}
	if closedByDeadline {
		return service.Confirm(ctx, userID, mealID, input)
	}
	meal, err := service.Meal(ctx, userID, mealID)
	return meal, shoppingListID, err
}

// Complete 由创建者手动完成已进入采购阶段的饭局。
func (service *MealService) Complete(ctx context.Context, userID, mealID string) (frontResponse.Meal, error) {
	now := service.now()
	return service.transition(ctx, userID, mealID,
		[]mealModel.MealStatus{mealModel.MealConfirmed},
		mealModel.MealCompleted, map[string]interface{}{"completed_at": now})
}

func shoppingItemResponse(row mealModel.FrontShoppingItem) frontResponse.ShoppingItem {
	return frontResponse.ShoppingItem{
		ID: row.ID, Name: row.Name, Amount: row.Amount, Note: row.Note,
		Completed: row.Completed, SortOrder: row.SortOrder,
	}
}

func (service *MealService) shoppingList(
	ctx context.Context,
	list mealModel.FrontShoppingList,
	readOnly bool,
) (frontResponse.ShoppingList, error) {
	var meal mealModel.FrontMeal
	if err := service.database().WithContext(ctx).First(&meal, "id = ?", list.MealID).Error; err != nil {
		return frontResponse.ShoppingList{}, appErrors.FrontInternal.Wrap(err, "load shopping meal")
	}
	summary, err := service.mealSummary(ctx, meal)
	if err != nil {
		return frontResponse.ShoppingList{}, err
	}
	var rows []mealModel.FrontShoppingItem
	if err := service.database().WithContext(ctx).Where("list_id = ?", list.ID).
		Order("sort_order asc, created_at asc").Find(&rows).Error; err != nil {
		return frontResponse.ShoppingList{}, appErrors.FrontInternal.Wrap(err, "load shopping items")
	}
	result := frontResponse.ShoppingList{
		ID: list.ID, ShareToken: list.ShareToken, Meal: summary, Items: []frontResponse.ShoppingItem{},
		ReadOnly: readOnly,
	}
	for _, row := range rows {
		result.Items = append(result.Items, shoppingItemResponse(row))
		if row.Completed {
			result.CompletedCount++
		} else {
			result.PendingCount++
		}
	}
	return result, nil
}

// CurrentShopping 获取当前饭局的采购清单。
func (service *MealService) CurrentShopping(
	ctx context.Context,
	userID string,
) (frontResponse.ShoppingList, error) {
	var list mealModel.FrontShoppingList
	err := service.database().WithContext(ctx).
		Joins("JOIN of_meals ON of_meals.id = of_shopping_lists.meal_id").
		Where("of_shopping_lists.owner_id = ? AND of_meals.status = ?",
			userID, mealModel.MealConfirmed).
		Order("of_shopping_lists.created_at desc").
		First(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.ShoppingList{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.ShoppingList{}, appErrors.FrontInternal.Wrap(err, "load current shopping list")
	}
	return service.shoppingList(ctx, list, false)
}

// loadCurrentShoppingListTx 在写事务中读取当前采购阶段饭局及所属采购清单。
func (service *MealService) loadCurrentShoppingListTx(
	tx *gorm.DB,
	userID string,
) (mealModel.FrontShoppingList, error) {
	var meal mealModel.FrontMeal
	if err := tx.
		Where("creator_id = ? AND status = ?", userID, mealModel.MealConfirmed).
		Order("created_at desc").
		First(&meal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mealModel.FrontShoppingList{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return mealModel.FrontShoppingList{}, appErrors.FrontInternal.Wrap(
			err,
			"load current shopping meal",
		)
	}
	var list mealModel.FrontShoppingList
	if err := tx.First(
		&list,
		"meal_id = ? AND owner_id = ?",
		meal.ID,
		userID,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mealModel.FrontShoppingList{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return mealModel.FrontShoppingList{}, appErrors.FrontInternal.Wrap(
			err,
			"load current shopping list for mutation",
		)
	}
	return list, nil
}

// ShoppingByID 获取本人清单或向已加入成员开放的历史只读采购清单。
func (service *MealService) ShoppingByID(
	ctx context.Context,
	userID string,
	listID string,
) (frontResponse.ShoppingList, error) {
	var list mealModel.FrontShoppingList
	if err := service.database().WithContext(ctx).First(&list, "id = ?", listID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.ShoppingList{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.ShoppingList{}, appErrors.FrontInternal.Wrap(err, "load retained shopping list")
	}
	var meal mealModel.FrontMeal
	if err := service.database().WithContext(ctx).First(&meal, "id = ?", list.MealID).Error; err != nil {
		return frontResponse.ShoppingList{}, appErrors.FrontInternal.Wrap(err, "load retained shopping meal")
	}
	if list.OwnerID == userID {
		return service.shoppingList(
			ctx,
			list,
			meal.Status != mealModel.MealConfirmed,
		)
	}
	historicalReadable := meal.Status == mealModel.MealCompleted ||
		(meal.Status == mealModel.MealCancelled &&
			meal.CancelledReason != nil &&
			*meal.CancelledReason == "creator_disabled")
	if !historicalReadable {
		return frontResponse.ShoppingList{}, appErrors.FrontNoPermission.DefaultMsg()
	}
	var participantCount int64
	if err := service.database().WithContext(ctx).Model(&mealModel.MealParticipant{}).
		Where("meal_id = ? AND user_id = ?", meal.ID, userID).
		Count(&participantCount).Error; err != nil {
		return frontResponse.ShoppingList{}, appErrors.FrontInternal.Wrap(
			err,
			"check historical shopping participant",
		)
	}
	if participantCount == 0 {
		return frontResponse.ShoppingList{}, appErrors.FrontNoPermission.DefaultMsg()
	}
	return service.shoppingList(ctx, list, true)
}

// SharedShopping 通过分享令牌获取只读采购清单。
func (service *MealService) SharedShopping(
	ctx context.Context,
	shareToken string,
) (frontResponse.SharedShoppingList, error) {
	var list mealModel.FrontShoppingList
	err := service.database().WithContext(ctx).First(&list, "share_token_hash = ?", shareDigest(shareToken)).Error
	if err != nil || list.ShareRevokedAt != nil ||
		(list.ShareExpiresAt != nil && !list.ShareExpiresAt.After(service.now())) {
		return frontResponse.SharedShoppingList{}, appErrors.FrontNotFound.DefaultMsg()
	}
	internal, err := service.shoppingList(ctx, list, true)
	if err != nil {
		return frontResponse.SharedShoppingList{}, err
	}
	return frontResponse.SharedShoppingList{
		ID: internal.ID, Meal: internal.Meal, Items: internal.Items,
		PendingCount: internal.PendingCount, CompletedCount: internal.CompletedCount, ReadOnly: true,
	}, nil
}

// CreateShoppingItem 创建采购清单项目。
func (service *MealService) CreateShoppingItem(
	ctx context.Context,
	userID string,
	input frontRequest.ShoppingItemCreateInput,
) (frontResponse.ShoppingItem, error) {
	var row mealModel.FrontShoppingItem
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		list, err := service.loadCurrentShoppingListTx(tx, userID)
		if err != nil {
			return err
		}
		var maxSort int
		if err := tx.Model(&mealModel.FrontShoppingItem{}).
			Where("list_id = ?", list.ID).
			Select("COALESCE(MAX(sort_order), 0)").
			Scan(&maxSort).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "load shopping sort order")
		}
		now := service.now()
		row = mealModel.FrontShoppingItem{
			ID: commonModel.NewID(), ListID: list.ID, Name: input.Name,
			Amount: input.Amount, Note: input.Note, SortOrder: maxSort + 1,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create shopping item")
		}
		return nil
	})
	if err != nil {
		return frontResponse.ShoppingItem{}, err
	}
	return shoppingItemResponse(row), nil
}

// UpdateShoppingItem 更新采购清单项目。
func (service *MealService) UpdateShoppingItem(
	ctx context.Context,
	userID string,
	itemID string,
	input frontRequest.ShoppingItemUpdateInput,
) (frontResponse.ShoppingItem, error) {
	var row mealModel.FrontShoppingItem
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		list, err := service.loadCurrentShoppingListTx(tx, userID)
		if err != nil {
			return err
		}
		if err := tx.First(&row, "id = ? AND list_id = ?", itemID, list.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load shopping item")
		}
		updates := map[string]interface{}{"updated_at": service.now()}
		if input.Name != nil {
			updates["name"] = *input.Name
		}
		if input.Amount != nil {
			updates["amount"] = *input.Amount
		}
		if input.Note != nil {
			updates["note"] = input.Note
		}
		if input.Completed != nil {
			updates["completed"] = *input.Completed
		}
		if input.SortOrder != nil {
			updates["sort_order"] = *input.SortOrder
		}
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "update shopping item")
		}
		if err := tx.First(&row, "id = ?", itemID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "reload shopping item")
		}
		return nil
	})
	if err != nil {
		return frontResponse.ShoppingItem{}, err
	}
	return shoppingItemResponse(row), nil
}

// DeleteShoppingItem 删除采购清单项目。
func (service *MealService) DeleteShoppingItem(
	ctx context.Context,
	userID string,
	itemID string,
) error {
	return service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		list, err := service.loadCurrentShoppingListTx(tx, userID)
		if err != nil {
			return err
		}
		result := tx.
			Where("id = ? AND list_id = ?", itemID, list.ID).
			Delete(&mealModel.FrontShoppingItem{})
		if result.Error != nil {
			return appErrors.FrontInternal.Wrap(result.Error, "delete shopping item")
		}
		if result.RowsAffected == 0 {
			return appErrors.FrontNotFound.DefaultMsg()
		}
		return nil
	})
}

// ExportShopping 生成采购清单的可复制文本。
func (service *MealService) ExportShopping(
	ctx context.Context,
	userID string,
	listID string,
) (string, string, time.Time, error) {
	var list mealModel.FrontShoppingList
	if err := service.database().WithContext(ctx).
		First(&list, "id = ? AND owner_id = ?", listID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", time.Time{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return "", "", time.Time{}, appErrors.FrontInternal.Wrap(err, "load shopping export")
	}
	result, err := service.shoppingList(ctx, list, false)
	if err != nil {
		return "", "", time.Time{}, err
	}
	var builder strings.Builder
	builder.WriteString(result.Meal.Name + "采购清单\n")
	for _, item := range result.Items {
		marker := "□"
		if item.Completed {
			marker = "✓"
		}
		builder.WriteString(fmt.Sprintf("%s %s %s", marker, item.Name, item.Amount))
		if item.Note != nil && strings.TrimSpace(*item.Note) != "" {
			builder.WriteString("（" + *item.Note + "）")
		}
		builder.WriteByte('\n')
	}
	return result.Meal.Name + "采购清单", strings.TrimSpace(builder.String()), list.UpdatedAt, nil
}

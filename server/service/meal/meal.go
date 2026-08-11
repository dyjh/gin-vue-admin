package meal

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	mealModel "github.com/dyjh/order-food-mini-app/server/model/meal"
	mealRequest "github.com/dyjh/order-food-mini-app/server/model/meal/request"
	mealResponse "github.com/dyjh/order-food-mini-app/server/model/meal/response"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PermissionMealRead 表示读取饭局管理数据所需的权限。
const PermissionMealRead = "orderfood:meal:read"

// MealService 提供管理端饭局只读业务能力。
type MealService struct {
	DB         *gorm.DB                         // 数据库连接
	Permission *serviceCommon.PermissionService // 管理端权限服务
	Now        func() time.Time                 // 当前时间函数
}

// NewMealService 创建管理端饭局查询服务实例。
func NewMealService(db *gorm.DB, permission *serviceCommon.PermissionService) *MealService {
	if permission == nil {
		permission = serviceCommon.NewPermissionService(db)
	}
	return &MealService{DB: db, Permission: permission, Now: time.Now}
}

// database 返回服务使用的数据库连接。
func (service *MealService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回服务使用的UTC时间。
func (service *MealService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// validateMealAdminListQuery 校验服务层饭局列表参数并阻止不安全排序表达式。
func validateMealAdminListQuery(query mealRequest.MealAdminListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.Keyword))) > 60 ||
		len([]rune(strings.TrimSpace(query.CreatorID))) > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, ok := map[string]struct{}{
		"": {}, "collecting": {}, "closed": {}, "confirmed": {},
		"completed": {}, "cancelled": {},
	}[query.Status]; !ok {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, ok := map[string]struct{}{"": {}, "manual": {}, "deadline": {}}[query.CloseReason]; !ok {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, ok := map[string]struct{}{
		"": {}, "manual": {}, "creator_disabled": {},
	}[query.CancelReason]; !ok {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, ok := map[string]struct{}{
		"createdAt": {}, "deadlineAt": {}, "confirmedAt": {}, "completedAt": {},
	}[query.SortBy]; !ok {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.SortOrder != "asc" && query.SortOrder != "desc" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	for _, pair := range [][2]*time.Time{
		{query.CreatedFrom, query.CreatedTo},
		{query.DeadlineFrom, query.DeadlineTo},
		{query.JoinedFrom, query.JoinedTo},
		{query.ConfirmedFrom, query.ConfirmedTo},
		{query.CompletedFrom, query.CompletedTo},
	} {
		if pair[0] != nil && pair[1] != nil && pair[0].After(*pair[1]) {
			return appErrors.AdminBadRequest.DefaultMsg()
		}
	}
	return nil
}

// ListMeals 分页查询饭局。
func (service *MealService) ListMeals(
	ctx context.Context,
	query mealRequest.MealAdminListQuery,
	actor commonRequest.AdminActor,
) (commonResponse.Page[mealResponse.MealAdminSummary], error) {
	query.ApplyDefaults()
	if err := validateMealAdminListQuery(query); err != nil {
		return commonResponse.Page[mealResponse.MealAdminSummary]{}, err
	}
	if service == nil || service.Permission == nil {
		return commonResponse.Page[mealResponse.MealAdminSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionMealRead); err != nil {
		return commonResponse.Page[mealResponse.MealAdminSummary]{}, err
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[mealResponse.MealAdminSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&mealModel.FrontMeal{})
	if value := strings.TrimSpace(query.Keyword); value != "" {
		keyword := "%" + value + "%"
		statement = statement.Where("id LIKE ? OR name LIKE ?", keyword, keyword)
	}
	for _, filter := range []struct {
		column string // 数据库字段
		value  string // 查询值
	}{
		{"creator_id", query.CreatorID},
		{"status", query.Status},
		{"close_reason", query.CloseReason},
	} {
		if value := strings.TrimSpace(filter.value); value != "" {
			statement = statement.Where(filter.column+" = ?", value)
		}
	}
	switch query.CancelReason {
	case "":
	case "creator_disabled":
		statement = statement.Where("status = ? AND cancelled_reason = ?", mealModel.MealCancelled, "creator_disabled")
	case "manual":
		statement = statement.Where(
			"status = ? AND (cancelled_reason IS NULL OR cancelled_reason <> ?)",
			mealModel.MealCancelled, "creator_disabled",
		)
	default:
		return commonResponse.Page[mealResponse.MealAdminSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", query.CreatedFrom.UTC())
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", query.CreatedTo.UTC())
	}
	if query.DeadlineFrom != nil {
		statement = statement.Where("deadline_at >= ?", query.DeadlineFrom.UTC())
	}
	if query.DeadlineTo != nil {
		statement = statement.Where("deadline_at <= ?", query.DeadlineTo.UTC())
	}
	if query.ConfirmedFrom != nil {
		statement = statement.Where("confirmed_at >= ?", query.ConfirmedFrom.UTC())
	}
	if query.ConfirmedTo != nil {
		statement = statement.Where("confirmed_at <= ?", query.ConfirmedTo.UTC())
	}
	if query.CompletedFrom != nil {
		statement = statement.Where("completed_at >= ?", query.CompletedFrom.UTC())
	}
	if query.CompletedTo != nil {
		statement = statement.Where("completed_at <= ?", query.CompletedTo.UTC())
	}
	if query.JoinedFrom != nil || query.JoinedTo != nil {
		memberTable := (mealModel.MealParticipant{}).TableName()
		mealTable := (mealModel.FrontMeal{}).TableName()
		joinedMeals := db.WithContext(ctx).Table(memberTable + " AS members").
			Select("members.meal_id").
			Joins("JOIN " + mealTable + " AS joined_meals ON joined_meals.id = members.meal_id").
			Where("members.user_id <> joined_meals.creator_id")
		if query.JoinedFrom != nil {
			joinedMeals = joinedMeals.Where("members.joined_at >= ?", query.JoinedFrom.UTC())
		}
		if query.JoinedTo != nil {
			joinedMeals = joinedMeals.Where("members.joined_at <= ?", query.JoinedTo.UTC())
		}
		statement = statement.Where("id IN (?)", joinedMeals)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[mealResponse.MealAdminSummary]{},
			appErrors.AdminInternal.Wrap(err, "count admin meals")
	}
	sortColumn := map[string]string{
		"createdAt": "created_at", "deadlineAt": "deadline_at",
		"confirmedAt": "confirmed_at", "completedAt": "completed_at",
	}[query.SortBy]
	if sortColumn == "" {
		return commonResponse.Page[mealResponse.MealAdminSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var meals []mealModel.FrontMeal
	if err := statement.
		Order(sortColumn + " " + query.SortOrder).
		Order("id " + query.SortOrder).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&meals).Error; err != nil {
		return commonResponse.Page[mealResponse.MealAdminSummary]{},
			appErrors.AdminInternal.Wrap(err, "list admin meals")
	}
	list := make([]mealResponse.MealAdminSummary, 0, len(meals))
	for index := range meals {
		summary, err := service.mealSummary(ctx, db, meals[index])
		if err != nil {
			return commonResponse.Page[mealResponse.MealAdminSummary]{}, err
		}
		list = append(list, summary)
	}
	return commonResponse.Page[mealResponse.MealAdminSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetMeal 获取饭局只读详情。
func (service *MealService) GetMeal(
	ctx context.Context,
	mealID string,
	actor commonRequest.AdminActor,
) (mealResponse.MealAdminDetail, error) {
	if service == nil || service.Permission == nil {
		return mealResponse.MealAdminDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionMealRead); err != nil {
		return mealResponse.MealAdminDetail{}, err
	}
	db := service.database()
	if db == nil {
		return mealResponse.MealAdminDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var meal mealModel.FrontMeal
	if err := db.WithContext(ctx).First(&meal, "id = ?", strings.TrimSpace(mealID)).Error; err != nil {
		return mealResponse.MealAdminDetail{}, mealAdminLookupError(err)
	}
	summary, err := service.mealSummary(ctx, db, meal)
	if err != nil {
		return mealResponse.MealAdminDetail{}, err
	}
	participants, timeline, err := service.mealParticipantsAndTimeline(ctx, db, meal)
	if err != nil {
		return mealResponse.MealAdminDetail{}, err
	}
	candidates, snapshots, err := service.mealCandidatesAndSnapshots(ctx, db, meal)
	if err != nil {
		return mealResponse.MealAdminDetail{}, err
	}
	var shoppingSummary *mealResponse.ShoppingListAdminSummary
	if meal.ShoppingListID != nil {
		var shoppingList mealModel.FrontShoppingList
		if err := db.WithContext(ctx).First(&shoppingList, "id = ?", *meal.ShoppingListID).Error; err == nil {
			value, err := buildShoppingListSummary(ctx, db, shoppingList, service.now())
			if err != nil {
				return mealResponse.MealAdminDetail{}, err
			}
			shoppingSummary = &value
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return mealResponse.MealAdminDetail{},
				appErrors.AdminInternal.Wrap(err, "load meal shopping summary")
		}
	}
	return mealResponse.MealAdminDetail{
		MealAdminSummary: summary, Participants: participants, Candidates: candidates,
		Timeline: timeline, FinalDishSnapshots: snapshots,
		ShoppingListSummary:   shoppingSummary,
		ShareRevocationStatus: mealShareRevocationStatus(meal, shoppingSummary),
	}, nil
}

// mealSummary 将饭局记录转换为管理端摘要。
func (service *MealService) mealSummary(
	ctx context.Context,
	db *gorm.DB,
	meal mealModel.FrontMeal,
) (mealResponse.MealAdminSummary, error) {
	var creator userModel.MiniAppUser
	if err := db.WithContext(ctx).First(&creator, "id = ?", meal.CreatorID).Error; err != nil {
		return mealResponse.MealAdminSummary{}, mealAdminLookupError(err)
	}
	var participantCount int64
	if err := db.WithContext(ctx).Model(&mealModel.MealParticipant{}).
		Where("meal_id = ?", meal.ID).Count(&participantCount).Error; err != nil {
		return mealResponse.MealAdminSummary{},
			appErrors.AdminInternal.Wrap(err, "count admin meal participants")
	}
	var candidateCount int64
	if err := db.WithContext(ctx).Model(&mealModel.FrontMealCandidate{}).
		Where("meal_id = ?", meal.ID).Count(&candidateCount).Error; err != nil {
		return mealResponse.MealAdminSummary{},
			appErrors.AdminInternal.Wrap(err, "count admin meal candidates")
	}
	var finalDishCount int64
	if err := db.WithContext(ctx).Model(&mealModel.FrontMealFinalDish{}).
		Where("meal_id = ?", meal.ID).Count(&finalDishCount).Error; err != nil {
		return mealResponse.MealAdminSummary{},
			appErrors.AdminInternal.Wrap(err, "count admin meal final dishes")
	}
	return mealResponse.MealAdminSummary{
		ID: meal.ID, Name: meal.Name, Creator: serviceCommon.UserReference(creator),
		CodeMasked: maskMealCode(meal.Code), Status: string(meal.Status),
		CloseReason: meal.CloseReason, CloseSource: mealCloseSource(meal),
		CancelReason:        adminCancelReason(meal),
		CancelledFromStatus: mealCancelledFromStatus(meal),
		ParticipantCount:    int(participantCount), CandidateCount: int(candidateCount),
		FinalDishCount: int(finalDishCount),
		ShoppingListID: meal.ShoppingListID, CreatedAt: meal.CreatedAt,
		DeadlineAt: meal.DeadlineAt, ClosedAt: meal.ClosedAt,
		ConfirmedAt: meal.ConfirmedAt, CompletedAt: meal.CompletedAt, CancelledAt: meal.CancelledAt,
	}, nil
}

// mealParticipantsAndTimeline 构建饭局参与者和状态时间线。
func (service *MealService) mealParticipantsAndTimeline(
	ctx context.Context,
	db *gorm.DB,
	meal mealModel.FrontMeal,
) ([]commonResponse.UserReference, []mealResponse.MealTimelineEvent, error) {
	var members []mealModel.MealParticipant
	if err := db.WithContext(ctx).Where("meal_id = ?", meal.ID).
		Order("joined_at ASC, id ASC").Find(&members).Error; err != nil {
		return nil, nil, appErrors.AdminInternal.Wrap(err, "list admin meal participants")
	}
	participants := make([]commonResponse.UserReference, 0, len(members))
	timeline := []mealResponse.MealTimelineEvent{{
		Type: "created", OccurredAt: meal.CreatedAt, ActorType: "user",
		ActorID: stringAddress(meal.CreatorID), Summary: "创建饭局",
	}}
	for index := range members {
		var user userModel.MiniAppUser
		if err := db.WithContext(ctx).First(&user, "id = ?", members[index].UserID).Error; err != nil {
			return nil, nil, mealAdminLookupError(err)
		}
		participants = append(participants, serviceCommon.UserReference(user))
		if members[index].UserID != meal.CreatorID || members[index].JoinedAt.After(meal.CreatedAt) {
			timeline = append(timeline, mealResponse.MealTimelineEvent{
				Type: "joined", OccurredAt: members[index].JoinedAt, ActorType: "user",
				ActorID: stringAddress(members[index].UserID), Summary: "加入饭局",
			})
		}
	}
	if meal.ClosedAt != nil {
		event := mealResponse.MealTimelineEvent{
			Type: "ordering_closed", OccurredAt: *meal.ClosedAt, Summary: "关闭点单",
			ActorType: "system",
		}
		if meal.CloseReason != nil && *meal.CloseReason == "manual" {
			event.ActorType = "user"
			event.ActorID = stringAddress(meal.CreatorID)
		}
		timeline = append(timeline, event)
	}
	for _, event := range []struct {
		eventType string     // 事件类型
		at        *time.Time // 发生时间
		summary   string     // 摘要
	}{
		{"menu_confirmed", meal.ConfirmedAt, "确认最终菜单并生成采购清单"},
		{"completed", meal.CompletedAt, "完成饭局"},
		{"cancelled", meal.CancelledAt, "取消饭局"},
	} {
		if event.at == nil {
			continue
		}
		timelineEvent := mealResponse.MealTimelineEvent{
			Type: event.eventType, OccurredAt: *event.at, ActorType: "user",
			ActorID: stringAddress(meal.CreatorID), Summary: event.summary,
		}
		if event.eventType == "cancelled" &&
			meal.CancelledReason != nil &&
			*meal.CancelledReason == "creator_disabled" {
			timelineEvent.ActorType = "system"
			timelineEvent.ActorID = nil
			timelineEvent.Summary = "发起人账号禁用，系统取消饭局"
		}
		timeline = append(timeline, timelineEvent)
	}
	sort.SliceStable(timeline, func(left, right int) bool {
		return timeline[left].OccurredAt.Before(timeline[right].OccurredAt)
	})
	return participants, timeline, nil
}

// mealCandidatesAndSnapshots 构建候选菜统计和最终菜单快照。
func (service *MealService) mealCandidatesAndSnapshots(
	ctx context.Context,
	db *gorm.DB,
	meal mealModel.FrontMeal,
) ([]mealResponse.MealCandidateAdmin, []mealResponse.MealFinalDishAdminSnapshot, error) {
	var finals []mealModel.FrontMealFinalDish
	if err := db.WithContext(ctx).Where("meal_id = ?", meal.ID).
		Order("created_at ASC, id ASC").Find(&finals).Error; err != nil {
		return nil, nil, appErrors.AdminInternal.Wrap(err, "list admin final dish snapshots")
	}
	finalByCandidate := make(map[string]mealModel.FrontMealFinalDish, len(finals))
	snapshots := make([]mealResponse.MealFinalDishAdminSnapshot, 0, len(finals))
	for index := range finals {
		finalByCandidate[finals[index].CandidateID] = finals[index]
		dishPublicID := ""
		var dish contentModel.UserDish
		if err := db.WithContext(ctx).Unscoped().First(&dish, finals[index].DishID).Error; err == nil {
			dishPublicID = dish.PublicID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, appErrors.AdminInternal.Wrap(err, "load final snapshot dish")
		}
		ingredients, err := decodeMealSnapshot(finals[index].Ingredients)
		if err != nil {
			return nil, nil, err
		}
		steps, err := decodeMealSnapshot(finals[index].Steps)
		if err != nil {
			return nil, nil, err
		}
		snapshots = append(snapshots, mealResponse.MealFinalDishAdminSnapshot{
			ID: finals[index].ID, CandidateID: finals[index].CandidateID, DishID: dishPublicID,
			Name: finals[index].Name, CoverURL: finals[index].CoverURL,
			FinalServings: finals[index].FinalServings, Ingredients: ingredients,
			Steps: steps, CreatedAt: finals[index].CreatedAt,
		})
	}
	var rows []mealModel.FrontMealCandidate
	if err := db.WithContext(ctx).Where("meal_id = ?", meal.ID).
		Order("sort_order ASC, created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, nil, appErrors.AdminInternal.Wrap(err, "list admin meal candidates")
	}
	candidates := make([]mealResponse.MealCandidateAdmin, 0, len(rows))
	for index := range rows {
		var dish contentModel.UserDish
		if err := db.WithContext(ctx).Unscoped().First(&dish, rows[index].DishID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, appErrors.AdminInternal.Wrap(err, "load admin candidate dish")
			}
		}
		var voteCount int64
		if err := db.WithContext(ctx).Model(&mealModel.FrontMealVote{}).
			Where("candidate_id = ?", rows[index].ID).Count(&voteCount).Error; err != nil {
			return nil, nil, appErrors.AdminInternal.Wrap(err, "count admin candidate votes")
		}
		status := "active"
		if !rows[index].Available {
			status = "removed"
			if rows[index].UnavailableReason != nil && *rows[index].UnavailableReason == "source_deleted" {
				status = "source_deleted"
			}
		}
		selected := false
		var finalServings *int
		if final, ok := finalByCandidate[rows[index].ID]; ok {
			selected = true
			value := final.FinalServings
			finalServings = &value
			// 来源菜品被删除时仍使用最终菜单快照名称，确保历史饭局可审计。
			if dish.Name == "" {
				dish.Name = final.Name
			}
		}
		candidates = append(candidates, mealResponse.MealCandidateAdmin{
			ID: rows[index].ID, DishID: dish.PublicID, Name: dish.Name,
			Status: status, VoteCount: int(voteCount), Selected: selected,
			FinalServings: finalServings,
		})
	}
	return candidates, snapshots, nil
}

// decodeMealSnapshot 解码最终菜单JSON快照。
func decodeMealSnapshot(raw datatypes.JSON) (interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []interface{}{}, nil
	}
	var snapshot interface{}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "decode final meal snapshot")
	}
	return snapshot, nil
}

// adminCancelReason 将用户填写的取消说明归类为稳定管理端枚举。
func adminCancelReason(meal mealModel.FrontMeal) *string {
	if meal.Status != mealModel.MealCancelled {
		return nil
	}
	value := "manual"
	if meal.CancelledReason != nil && *meal.CancelledReason == "creator_disabled" {
		value = "creator_disabled"
	}
	return &value
}

// mealCloseSource 返回饭局关闭点单的可诊断触发来源。
func mealCloseSource(meal mealModel.FrontMeal) *string {
	if meal.CloseReason == nil {
		return nil
	}
	if meal.CloseSource != nil {
		value := string(*meal.CloseSource)
		return &value
	}
	value := string(mealModel.MealCloseSourceLegacy)
	if *meal.CloseReason == "manual" {
		value = string(mealModel.MealCloseSourceCreatorAction)
	}
	return &value
}

// mealCancelledFromStatus 返回饭局取消前的生命周期状态。
func mealCancelledFromStatus(meal mealModel.FrontMeal) *string {
	if meal.CancelledFrom == nil {
		return nil
	}
	value := string(*meal.CancelledFrom)
	return &value
}

// mealShareRevocationStatus 返回发起人禁用取消饭局时的采购分享撤销结果。
func mealShareRevocationStatus(
	meal mealModel.FrontMeal,
	shoppingSummary *mealResponse.ShoppingListAdminSummary,
) *string {
	if meal.CancelledReason == nil || *meal.CancelledReason != "creator_disabled" {
		return nil
	}
	value := "no_share"
	if shoppingSummary != nil {
		value = "not_revoked"
		if shoppingSummary.ShareStatus == "revoked" {
			value = "revoked"
		}
	}
	return &value
}

// maskMealCode 返回脱敏后的饭局邀请码。
func maskMealCode(code string) string {
	code = strings.TrimSpace(code)
	if len(code) <= 2 {
		return "******"
	}
	return code[:2] + "****"
}

// stringAddress 返回字符串指针。
func stringAddress(value string) *string {
	return &value
}

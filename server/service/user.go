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
	PermissionUserRead           = "orderfood:user:read"
	PermissionUserDisable        = "orderfood:user:disable"
	PermissionUserPreferenceRead = "orderfood:user:preference:read"
)

var preferenceEvidenceSourceTypes = []string{
	"checkin_image",
	"recommendation_adopted",
	"dish_saved",
	"dish_ordered",
	"reshuffle",
	"skip_for_now",
	"explicit_setting",
}

// CapabilityPolicySnapshot 表示影响小程序用户能力状态的平台策略快照。
type CapabilityPolicySnapshot struct {
	PlatformDefaultEnabled bool // 平台AI总开关
	EmergencyDisabled      bool // 是否处于紧急停用状态
}

// CapabilityPolicyProvider 定义能力策略供应商所需的业务能力。
type CapabilityPolicyProvider interface {
	Snapshot(ctx context.Context) (CapabilityPolicySnapshot, error)
}

// StaticCapabilityPolicyProvider 为测试或降级场景提供固定能力策略。
type StaticCapabilityPolicyProvider struct {
	Policy CapabilityPolicySnapshot // 固定返回的能力策略
}

// Snapshot 获取当前生效的静态能力策略供应商快照。
func (provider StaticCapabilityPolicyProvider) Snapshot(context.Context) (CapabilityPolicySnapshot, error) {
	return provider.Policy, nil
}

// UserService 提供小程序用户业务能力。
type UserService struct {
	DB          *gorm.DB                 // 业务数据库
	Audit       *AccessAuditService      // 敏感数据访问审计服务
	Idempotency *IdempotencyService      // 幂等执行服务
	Policy      CapabilityPolicyProvider // 平台整体能力策略提供器
	Now         func() time.Time         // 可注入的当前时间函数
}

// NewUserService 创建用户服务实例。
func NewUserService(
	db *gorm.DB,
	audit *AccessAuditService,
	idempotency *IdempotencyService,
	policy CapabilityPolicyProvider,
) *UserService {
	if audit == nil {
		audit = NewAccessAuditService(db)
	}
	if idempotency == nil {
		idempotency = NewIdempotencyService(db)
	}
	if policy == nil {
		policy = StaticCapabilityPolicyProvider{
			Policy: CapabilityPolicySnapshot{PlatformDefaultEnabled: false},
		}
	}
	return &UserService{
		DB:          db,
		Audit:       audit,
		Idempotency: idempotency,
		Policy:      policy,
		Now:         time.Now,
	}
}

func (service *UserService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *UserService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

func (service *UserService) policy(ctx context.Context) (CapabilityPolicySnapshot, error) {
	if service != nil && service.Policy != nil {
		policy, err := service.Policy.Snapshot(ctx)
		if err != nil {
			return CapabilityPolicySnapshot{}, appErrors.AdminInternal.Wrap(err, "read capability policy")
		}
		return policy, nil
	}
	return CapabilityPolicySnapshot{PlatformDefaultEnabled: false}, nil
}

func capabilityState(policy CapabilityPolicySnapshot) (effective string, source string) {
	if policy.EmergencyDisabled {
		return "disabled", "emergency"
	}
	if policy.PlatformDefaultEnabled {
		return "enabled", "platform_default"
	}
	return "disabled", "platform_default"
}

func capabilityEffects(effective string) orderfoodResponse.CapabilityEffects {
	enabled := effective == "enabled"
	return orderfoodResponse.CapabilityEffects{
		PointsVisible:             enabled,
		PointEntriesVisible:       enabled,
		CheckinRewardEnabled:      enabled,
		PreferenceAnalysisEnabled: enabled,
	}
}

// activeCreatedMealImpact 查询用户当前发起的唯一进行中饭局及禁用时需要保留的数据量。
func activeCreatedMealImpact(
	db *gorm.DB,
	userID string,
) (*orderfoodResponse.UserActiveMealImpact, error) {
	var meal orderfoodModel.FrontMeal
	err := db.Where(
		"creator_id = ? AND status IN ?",
		userID,
		[]orderfoodModel.MealStatus{
			orderfoodModel.MealCollecting,
			orderfoodModel.MealClosed,
			orderfoodModel.MealConfirmed,
		},
	).Order("created_at desc, id desc").First(&meal).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "load active created meal impact")
	}

	impact := &orderfoodResponse.UserActiveMealImpact{
		MealID: meal.ID,
		Status: string(meal.Status),
	}
	if err := db.Model(&orderfoodModel.MealParticipant{}).
		Where("meal_id = ?", meal.ID).
		Count(&impact.ParticipantCount).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "count active meal participants")
	}
	if err := db.Model(&orderfoodModel.FrontMealFinalDish{}).
		Where("meal_id = ?", meal.ID).
		Count(&impact.FinalSnapshotCount).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "count active meal snapshots")
	}
	if err := db.Model(&orderfoodModel.FrontShoppingList{}).
		Where("meal_id = ?", meal.ID).
		Count(&impact.ShoppingListCount).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "count active meal shopping lists")
	}
	return impact, nil
}

// List 分页查询小程序用户。
func (service *UserService) List(
	ctx context.Context,
	query orderfoodRequest.UserListQuery,
) (orderfoodResponse.Page[orderfoodResponse.UserSummary], error) {
	query.ApplyDefaults()
	db := service.database()
	if db == nil {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	policy, err := service.policy(ctx)
	if err != nil {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, err
	}
	if query.MinPoints != nil && query.MaxPoints != nil && *query.MinPoints > *query.MaxPoints {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}

	statement := db.WithContext(ctx).Model(&orderfoodModel.MiniAppUser{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		search := "%" + keyword + "%"
		statement = statement.Where("id LIKE ? OR nickname LIKE ?", search, search)
	}
	if query.Status != "" {
		statement = statement.Where("status = ?", query.Status)
	}
	if query.MinPoints != nil {
		statement = statement.Where("points >= ?", *query.MinPoints)
	}
	if query.MaxPoints != nil {
		statement = statement.Where("points <= ?", *query.MaxPoints)
	}
	for _, dateFilter := range []struct {
		value  string
		column string
		op     string
	}{
		{query.RegisteredFrom, "registered_at", ">="},
		{query.RegisteredTo, "registered_at", "<="},
		{query.LastLoginFrom, "last_login_at", ">="},
		{query.LastLoginTo, "last_login_at", "<="},
	} {
		if dateFilter.value == "" {
			continue
		}
		parsed, parseErr := time.Parse(time.RFC3339, dateFilter.value)
		if parseErr != nil {
			return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		statement = statement.Where(
			dateFilter.column+" "+dateFilter.op+" ?",
			parsed.UTC(),
		)
	}
	activeFrom, err := userActivityDate(query.ActiveFrom)
	if err != nil {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, err
	}
	activeTo, err := userActivityDate(query.ActiveTo)
	if err != nil {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, err
	}
	if activeFrom != "" || activeTo != "" {
		activity := db.WithContext(ctx).Model(&orderfoodModel.FrontUserActivityDay{}).
			Select("user_id")
		if activeFrom != "" {
			activity = activity.Where("active_date >= ?", activeFrom)
		}
		if activeTo != "" {
			activity = activity.Where("active_date <= ?", activeTo)
		}
		statement = statement.Where("id IN (?)", activity)
	}
	statement = applyCapabilityFilter(statement, query.CapabilityEffective, policy)

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminInternal.Wrap(err, "count users")
	}
	sortColumns := map[string]string{
		"createdAt":   "registered_at",
		"lastLoginAt": "last_login_at",
		"points":      "points",
	}
	sortColumn, ok := sortColumns[query.SortBy]
	if !ok {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	sortOrder := strings.ToLower(query.SortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}

	var users []orderfoodModel.MiniAppUser
	if err := statement.
		Order(sortColumn + " " + sortOrder).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&users).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.UserSummary]{}, appErrors.AdminInternal.Wrap(err, "list users")
	}
	list := make([]orderfoodResponse.UserSummary, 0, len(users))
	for index := range users {
		list = append(list, toUserSummary(users[index], policy))
	}
	return orderfoodResponse.Page[orderfoodResponse.UserSummary]{
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
		List:     list,
	}, nil
}

// userActivityDate 将管理端时间筛选转换为上海自然日字符串。
func userActivityDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", appErrors.AdminBadRequest.DefaultMsg()
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return parsed.In(location).Format("2006-01-02"), nil
}

func applyCapabilityFilter(
	statement *gorm.DB,
	effective string,
	policy CapabilityPolicySnapshot,
) *gorm.DB {
	if effective == "" {
		return statement
	}
	if policy.EmergencyDisabled {
		if effective == "enabled" {
			return statement.Where("1 = 0")
		}
		return statement
	}
	platformEffective := "disabled"
	if policy.PlatformDefaultEnabled {
		platformEffective = "enabled"
	}
	if effective != platformEffective {
		return statement.Where("1 = 0")
	}
	return statement
}

// Detail 获取小程序用户详情。
func (service *UserService) Detail(
	ctx context.Context,
	userID string,
) (orderfoodResponse.UserDetail, error) {
	db := service.database()
	if db == nil || userID == "" {
		if userID == "" {
			return orderfoodResponse.UserDetail{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		return orderfoodResponse.UserDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	policy, err := service.policy(ctx)
	if err != nil {
		return orderfoodResponse.UserDetail{}, err
	}
	var user orderfoodModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodResponse.UserDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return orderfoodResponse.UserDetail{}, appErrors.AdminInternal.Wrap(err, "get user")
	}
	summary := toUserSummary(user, policy)
	mealImpact, err := activeCreatedMealImpact(db.WithContext(ctx), userID)
	if err != nil {
		return orderfoodResponse.UserDetail{}, err
	}
	return orderfoodResponse.UserDetail{
		UserSummary:             summary,
		WechatIdentityMasked:    maskWechatIdentity(user.OpenIDHash),
		DisabledReason:          user.DisabledReason,
		DisabledAt:              user.DisabledAt,
		CapabilityEffects:       capabilityEffects(summary.CapabilityEffective),
		RecipeCount:             user.RecipeCount,
		CheckinCount:            user.CheckinCount,
		ActiveCreatedMealImpact: mealImpact,
	}, nil
}

func toUserSummary(
	user orderfoodModel.MiniAppUser,
	policy CapabilityPolicySnapshot,
) orderfoodResponse.UserSummary {
	effective, source := capabilityState(policy)
	return orderfoodResponse.UserSummary{
		ID:                  user.ID,
		AvatarURL:           user.AvatarURL,
		Nickname:            user.Nickname,
		Points:              user.Points,
		CapabilityEffective: effective,
		CapabilitySource:    source,
		CheckinDayCount:     user.CheckinDayCount,
		DishCount:           user.DishCount,
		MealCount:           user.MealCount,
		Status:              string(user.Status),
		Version:             user.Version,
		RegisteredAt:        user.RegisteredAt,
		LastLoginAt:         user.LastLoginAt,
	}
}

func maskWechatIdentity(openIDHash string) *string {
	if openIDHash == "" {
		return nil
	}
	maskedLength := 12
	if len(openIDHash) < maskedLength {
		maskedLength = len(openIDHash)
	}
	masked := "sha256:" + openIDHash[:maskedLength]
	return &masked
}

func validateReason(reason string) error {
	length := utf8.RuneCountInString(strings.TrimSpace(reason))
	if length < 4 || length > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

func validateAdminActor(actor orderfoodRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

type userDisableMealEffects struct {
	cancelledMealCount         int64
	notifiedParticipantCount   int64
	preservedSnapshotCount     int64
	preservedShoppingListCount int64
}

// cancelCreatorMealsForDisabledUser 在用户禁用事务内取消其发起的进行中饭局并保留历史快照。
func cancelCreatorMealsForDisabledUser(
	tx *gorm.DB,
	userID string,
	now time.Time,
) (userDisableMealEffects, error) {
	effects := userDisableMealEffects{}
	var meals []orderfoodModel.FrontMeal
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"creator_id = ? AND status IN ?",
			userID,
			[]orderfoodModel.MealStatus{
				orderfoodModel.MealCollecting,
				orderfoodModel.MealClosed,
				orderfoodModel.MealConfirmed,
			},
		).
		Order("created_at asc, id asc").
		Find(&meals).Error; err != nil {
		return effects, appErrors.AdminInternal.Wrap(err, "lock active meals for disabled creator")
	}

	cancelledResult := string(orderfoodModel.MealCancelled)
	for _, meal := range meals {
		var snapshotCount int64
		if err := tx.Model(&orderfoodModel.FrontMealFinalDish{}).
			Where("meal_id = ?", meal.ID).
			Count(&snapshotCount).Error; err != nil {
			return effects, appErrors.AdminInternal.Wrap(err, "count preserved meal snapshots")
		}
		effects.preservedSnapshotCount += snapshotCount

		var shoppingListCount int64
		if err := tx.Model(&orderfoodModel.FrontShoppingList{}).
			Where("meal_id = ?", meal.ID).
			Count(&shoppingListCount).Error; err != nil {
			return effects, appErrors.AdminInternal.Wrap(err, "count preserved shopping lists")
		}
		effects.preservedShoppingListCount += shoppingListCount
		if shoppingListCount > 0 {
			if err := tx.Model(&orderfoodModel.FrontShoppingList{}).
				Where("meal_id = ?", meal.ID).
				Updates(map[string]interface{}{
					"share_revoked_at": now,
					"share_expires_at": now,
					"updated_at":       now,
				}).Error; err != nil {
				return effects, appErrors.AdminInternal.Wrap(err, "revoke disabled creator shopping shares")
			}
		}

		// 用户禁用不发送微信订阅消息，但饭局已到最终状态，剩余的一次性授权不可再使用。
		if err := tx.Model(&orderfoodModel.MealFinalResultSubscription{}).
			Where("meal_id = ? AND accepted = ? AND consumed_at IS NULL", meal.ID, true).
			Updates(map[string]interface{}{
				"consumed_result": cancelledResult,
				"consumed_at":     now,
			}).Error; err != nil {
			return effects, appErrors.AdminInternal.Wrap(err, "consume disabled creator meal subscriptions")
		}

		var participants []orderfoodModel.MealParticipant
		if err := tx.Where("meal_id = ?", meal.ID).
			Order("joined_at asc, id asc").
			Find(&participants).Error; err != nil {
			return effects, appErrors.AdminInternal.Wrap(err, "load disabled creator meal participants")
		}
		targetType := "meal"
		targetID := meal.ID
		for _, participant := range participants {
			if participant.UserID == userID {
				continue
			}
			if err := tx.Create(&orderfoodModel.UserNotification{
				ID: orderfoodModel.NewID(), UserID: participant.UserID,
				Type: "meal", Title: "饭局已取消",
				Content:           "发起人账号不可用，本饭局已取消",
				TargetType:        &targetType,
				TargetID:          &targetID,
				SubscribeRequired: false,
				CreatedAt:         now,
			}).Error; err != nil {
				return effects, appErrors.AdminInternal.Wrap(err, "notify disabled creator meal participants")
			}
			effects.notifiedParticipantCount++
		}

		cancelReason := "creator_disabled"
		result := tx.Model(&orderfoodModel.FrontMeal{}).
			Where("id = ? AND status = ?", meal.ID, meal.Status).
			Updates(map[string]interface{}{
				"status":                orderfoodModel.MealCancelled,
				"cancelled_reason":      cancelReason,
				"cancelled_from_status": meal.Status,
				"cancelled_at":          now,
				"updated_at":            now,
			})
		if result.Error != nil {
			return effects, appErrors.AdminInternal.Wrap(result.Error, "cancel disabled creator meal")
		}
		if result.RowsAffected != 1 {
			return effects, appErrors.AdminStateConflict.DefaultMsg()
		}
		effects.cancelledMealCount++
	}
	return effects, nil
}

// UpdateStatus 更新状态。
func (service *UserService) UpdateStatus(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	userID string,
	idempotencyKey string,
	body orderfoodRequest.UpdateUserStatusBody,
) (orderfoodResponse.UserStatusResult, bool, error) {
	if err := validateReason(body.Reason); err != nil {
		return orderfoodResponse.UserStatusResult{}, false, err
	}
	if err := validateAdminActor(actor); err != nil {
		return orderfoodResponse.UserStatusResult{}, false, err
	}
	payload := struct {
		UserID string                                `json:"userId"`
		Body   orderfoodRequest.UpdateUserStatusBody `json:"body"`
	}{UserID: userID, Body: body}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"user_status_update",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			return service.updateStatusWithDB(ctx, tx, actor, userID, idempotencyKey, body)
		},
	)
	if err != nil {
		return orderfoodResponse.UserStatusResult{}, false, err
	}
	var result orderfoodResponse.UserStatusResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return orderfoodResponse.UserStatusResult{}, false, appErrors.AdminInternal.Wrap(err, "decode status response")
	}
	return result, replayed, nil
}

func (service *UserService) updateStatusWithDB(
	ctx context.Context,
	tx *gorm.DB,
	actor orderfoodRequest.AdminActor,
	userID string,
	idempotencyKey string,
	body orderfoodRequest.UpdateUserStatusBody,
) (orderfoodResponse.UserStatusResult, error) {
	var current orderfoodModel.MiniAppUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&current, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodResponse.UserStatusResult{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return orderfoodResponse.UserStatusResult{}, appErrors.AdminInternal.Wrap(err, "get user for status update")
	}
	if current.Version != body.ExpectedVersion || current.Status == body.Status {
		return orderfoodResponse.UserStatusResult{}, appErrors.AdminStateConflict.DefaultMsg()
	}
	now := service.now()
	updates := map[string]interface{}{
		"status":     body.Status,
		"version":    gorm.Expr("version + 1"),
		"updated_at": now,
	}
	if body.Status == orderfoodModel.UserStatusDisabled {
		reason := strings.TrimSpace(body.Reason)
		updates["disabled_reason"] = &reason
		updates["disabled_at"] = &now
	} else {
		updates["disabled_reason"] = nil
		updates["disabled_at"] = nil
	}
	result := tx.Model(&orderfoodModel.MiniAppUser{}).
		Where("id = ? AND version = ?", userID, body.ExpectedVersion).
		Updates(updates)
	if result.Error != nil {
		return orderfoodResponse.UserStatusResult{}, appErrors.AdminInternal.Wrap(result.Error, "update user status")
	}
	if result.RowsAffected != 1 {
		return orderfoodResponse.UserStatusResult{}, appErrors.AdminStateConflict.DefaultMsg()
	}
	mealEffects := userDisableMealEffects{}
	if body.Status == orderfoodModel.UserStatusDisabled {
		var mealErr error
		mealEffects, mealErr = cancelCreatorMealsForDisabledUser(tx, userID, now)
		if mealErr != nil {
			return orderfoodResponse.UserStatusResult{}, mealErr
		}
	}
	var updated orderfoodModel.MiniAppUser
	if err := tx.First(&updated, "id = ?", userID).Error; err != nil {
		return orderfoodResponse.UserStatusResult{}, appErrors.AdminInternal.Wrap(err, "reload user status")
	}
	responseResult := orderfoodResponse.UserStatusResult{
		UserID:                     updated.ID,
		Status:                     string(updated.Status),
		DisabledReason:             updated.DisabledReason,
		CancelledMealCount:         mealEffects.cancelledMealCount,
		NotifiedParticipantCount:   mealEffects.notifiedParticipantCount,
		PreservedSnapshotCount:     mealEffects.preservedSnapshotCount,
		PreservedShoppingListCount: mealEffects.preservedShoppingListCount,
		Version:                    updated.Version,
		UpdatedAt:                  updated.UpdatedAt,
	}
	if err := service.writeUserMutationAudit(
		ctx,
		tx,
		actor,
		"update_user_status",
		current,
		strings.TrimSpace(body.Reason),
		idempotencyKey,
		map[string]interface{}{
			"status":  current.Status,
			"version": current.Version,
		},
		map[string]interface{}{
			"status":                     updated.Status,
			"version":                    updated.Version,
			"cancelledMealCount":         mealEffects.cancelledMealCount,
			"notifiedParticipantCount":   mealEffects.notifiedParticipantCount,
			"preservedSnapshotCount":     mealEffects.preservedSnapshotCount,
			"preservedShoppingListCount": mealEffects.preservedShoppingListCount,
		},
	); err != nil {
		return orderfoodResponse.UserStatusResult{}, err
	}
	return responseResult, nil
}

func (service *UserService) writeUserMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor orderfoodRequest.AdminActor,
	action string,
	user orderfoodModel.MiniAppUser,
	reason string,
	idempotencyKey string,
	beforeSummary map[string]interface{},
	afterSummary map[string]interface{},
) error {
	beforeJSON, err := json.Marshal(beforeSummary)
	if err != nil {
		return appErrors.AdminInternal.Wrap(err, "encode user mutation before summary")
	}
	afterJSON, err := json.Marshal(afterSummary)
	if err != nil {
		return appErrors.AdminInternal.Wrap(err, "encode user mutation after summary")
	}
	targetLabel := user.Nickname
	reasonCopy := reason
	idempotencyKeyCopy := idempotencyKey
	audit := orderfoodModel.AdminAuditLog{
		PublicID:              orderfoodModel.NewID(),
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: strings.TrimSpace(actor.Username),
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            "user",
		TargetID:              user.ID,
		TargetLabel:           &targetLabel,
		Reason:                &reasonCopy,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             strings.TrimSpace(actor.RequestID),
		IdempotencyKey:        &idempotencyKeyCopy,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	if err := tx.WithContext(ctx).Create(&audit).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "write user mutation audit")
	}
	return nil
}

// PreferenceProfile 查询用户打卡形成的偏好画像。
func (service *UserService) PreferenceProfile(
	ctx context.Context,
	userID string,
	access SensitiveAccess,
) (orderfoodResponse.UserPreferenceProfile, error) {
	db := service.database()
	if db == nil {
		return orderfoodResponse.UserPreferenceProfile{}, appErrors.AdminInternal.DefaultMsg()
	}
	if userID == "" ||
		access.Permission != PermissionUserPreferenceRead ||
		access.AdministratorID == 0 {
		return orderfoodResponse.UserPreferenceProfile{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	policy, err := service.policy(ctx)
	if err != nil {
		return orderfoodResponse.UserPreferenceProfile{}, err
	}

	var result orderfoodResponse.UserPreferenceProfile
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user orderfoodModel.MiniAppUser
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.AdminNotFound.DefaultMsg()
			}
			return appErrors.AdminInternal.Wrap(err, "get preference profile user")
		}

		var stored orderfoodModel.UserPreferenceProfile
		profileErr := tx.First(&stored, "user_id = ?", userID).Error
		if profileErr != nil && !errors.Is(profileErr, gorm.ErrRecordNotFound) {
			return appErrors.AdminInternal.Wrap(profileErr, "get preference profile")
		}
		built, buildErr := buildPreferenceProfile(user, stored, profileErr == nil, policy)
		if buildErr != nil {
			return buildErr
		}
		result = built

		var evidence []orderfoodModel.PreferenceEvidenceAggregate
		if err := tx.Where("user_id = ?", userID).Find(&evidence).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "get preference evidence aggregates")
		}
		result.EvidenceSources = buildEvidenceSummaries(evidence)

		access.TargetType = "user_preference_profile"
		access.TargetID = userID
		access.Action = "read_preference_profile"
		auditID, auditErr := service.Audit.RecordWithDB(ctx, tx, access)
		if auditErr != nil {
			return auditErr
		}
		result.AccessAuditID = auditID
		return nil
	})
	if err != nil {
		// Never return a partially built sensitive response when its audit row
		// could not be committed.
		return orderfoodResponse.UserPreferenceProfile{}, err
	}
	return result, nil
}

func buildPreferenceProfile(
	user orderfoodModel.MiniAppUser,
	stored orderfoodModel.UserPreferenceProfile,
	found bool,
	policy CapabilityPolicySnapshot,
) (orderfoodResponse.UserPreferenceProfile, error) {
	effective, _ := capabilityState(policy)
	result := orderfoodResponse.UserPreferenceProfile{
		User: orderfoodResponse.UserReference{
			ID:        user.ID,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
			Status:    string(user.Status),
		},
		HasProfile:      found && stored.HasProfile,
		UpdateState:     string(orderfoodModel.PreferenceUpdatePending),
		UpdateEnabled:   effective == "enabled",
		CheckinDayCount: user.CheckinDayCount,
		EvidenceSources: []orderfoodResponse.PreferenceEvidenceSourceSummary{},
	}
	if !found {
		if effective != "enabled" {
			setCapabilityPaused(&result)
		}
		return result, nil
	}
	result.UpdateState = string(stored.UpdateState)
	result.UpdateEnabled = stored.UpdateEnabled && effective == "enabled"
	result.PausedReasonCode = stored.PausedReasonCode
	result.PausedReasonSummary = stored.PausedReasonSummary
	result.LastEvidenceAt = stored.LastEvidenceAt
	result.LastAggregatedAt = stored.LastAggregatedAt
	result.ProfileUpdatedAt = stored.ProfileUpdatedAt
	if stored.HasProfile && len(stored.ProfileJSON) > 0 {
		content := orderfoodResponse.PreferenceProfileContent{
			CommonDishTags:            []orderfoodResponse.PreferenceTermSummary{},
			FrequentlyOrderedDishTags: []orderfoodResponse.PreferenceTermSummary{},
			CommonIngredients:         []orderfoodResponse.PreferenceTermSummary{},
			UserSettings:              []orderfoodResponse.PreferenceUserSettingSummary{},
		}
		if err := json.Unmarshal(stored.ProfileJSON, &content); err != nil {
			return orderfoodResponse.UserPreferenceProfile{},
				appErrors.AdminInternal.Wrap(err, "decode preference profile")
		}
		result.Profile = &content
	}
	if len(stored.LatestFailureJSON) > 0 {
		var failure orderfoodResponse.PreferenceProfileFailureSummary
		if err := json.Unmarshal(stored.LatestFailureJSON, &failure); err != nil {
			return orderfoodResponse.UserPreferenceProfile{},
				appErrors.AdminInternal.Wrap(err, "decode preference failure summary")
		}
		result.LatestFailure = &failure
	}
	if effective != "enabled" {
		setCapabilityPaused(&result)
	}
	return result, nil
}

func setCapabilityPaused(result *orderfoodResponse.UserPreferenceProfile) {
	code := "capability_disabled"
	summary := "整体增强能力已关闭，历史画像保留但暂停更新"
	result.UpdateState = string(orderfoodModel.PreferenceUpdatePaused)
	result.UpdateEnabled = false
	result.PausedReasonCode = &code
	result.PausedReasonSummary = &summary
}

func buildEvidenceSummaries(
	aggregates []orderfoodModel.PreferenceEvidenceAggregate,
) []orderfoodResponse.PreferenceEvidenceSourceSummary {
	byType := make(map[string]orderfoodModel.PreferenceEvidenceAggregate, len(aggregates))
	for index := range aggregates {
		byType[aggregates[index].SourceType] = aggregates[index]
	}
	result := make([]orderfoodResponse.PreferenceEvidenceSourceSummary, 0, len(preferenceEvidenceSourceTypes))
	for _, sourceType := range preferenceEvidenceSourceTypes {
		aggregate := byType[sourceType]
		result = append(result, orderfoodResponse.PreferenceEvidenceSourceSummary{
			SourceType:       sourceType,
			TotalCount:       aggregate.TotalCount,
			AggregatedCount:  aggregate.AggregatedCount,
			PendingCount:     aggregate.PendingCount,
			LatestOccurredAt: aggregate.LatestOccurredAt,
		})
	}
	return result
}

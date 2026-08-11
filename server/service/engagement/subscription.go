package engagement

import (
	"context"
	"encoding/json"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"sort"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	engagementRequest "github.com/dyjh/order-food-mini-app/server/model/engagement/request"
	engagementResponse "github.com/dyjh/order-food-mini-app/server/model/engagement/response"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	// PermissionNotificationRead 表示读取站内通知所需的权限。
	PermissionNotificationRead = "orderfood:notification:read"
	// PermissionSubscribeSceneRead 表示读取固定订阅场景所需的权限。
	PermissionSubscribeSceneRead = "orderfood:subscribe-scene:read"
	// PermissionSubscribeSceneUpdate 表示配置固定订阅场景所需的权限。
	PermissionSubscribeSceneUpdate = "orderfood:subscribe-scene:update"
	// PermissionSubscribeSceneStatus 表示启停固定订阅场景所需的权限。
	PermissionSubscribeSceneStatus = "orderfood:subscribe-scene:status"
	// PermissionSubscribeLogRead 表示读取订阅消息发送记录所需的权限。
	PermissionSubscribeLogRead = "orderfood:subscribe-log:read"
)

var requiredMealStatusMappings = []string{"mealName", "result", "resultAt"}

type subscribeSceneDefinition struct {
	Scene                string
	Name                 string
	Purpose              string
	TriggerDescription   string
	RecipientDescription string
	RequiredFields       []engagementResponse.SubscribeSceneFieldDefinition
}

var fixedSubscribeSceneDefinitions = []subscribeSceneDefinition{
	{
		Scene:                engagementModel.SubscribeSceneMealStatus,
		Name:                 "饭局最终结果",
		Purpose:              "通知参与者饭局最终确认或取消结果",
		TriggerDescription:   "饭局确认完成或被取消后触发",
		RecipientDescription: "已主动授权本场饭局最终结果提醒的参与者",
		RequiredFields: []engagementResponse.SubscribeSceneFieldDefinition{
			{Key: "mealName", Label: "饭局名称", Description: "本次饭局的名称"},
			{Key: "result", Label: "最终结果", Description: "饭局已确认或已取消的结果"},
			{Key: "resultAt", Label: "结果时间", Description: "最终结果产生的时间"},
		},
	},
}

func fixedSubscribeSceneDefinition(scene string) (subscribeSceneDefinition, bool) {
	scene = strings.TrimSpace(scene)
	for _, definition := range fixedSubscribeSceneDefinitions {
		if definition.Scene == scene {
			return definition, true
		}
	}
	return subscribeSceneDefinition{}, false
}

// SubscriptionService 提供站内通知、订阅消息模板和发送记录的管理能力。
type SubscriptionService struct {
	DB            *gorm.DB                          // 业务数据库
	Idempotency   *serviceCommon.IdempotencyService // 幂等执行服务
	MutationAudit serviceCommon.MutationAuditWriter // 配置变更审计写入器
	Now           func() time.Time                  // 可注入的当前时间函数
}

// NewSubscriptionService 创建消息中心服务实例。
func NewSubscriptionService(
	db *gorm.DB,
	idempotency *serviceCommon.IdempotencyService,
) *SubscriptionService {
	if idempotency == nil {
		idempotency = serviceCommon.NewIdempotencyService(db)
	}
	return &SubscriptionService{
		DB:            db,
		Idempotency:   idempotency,
		MutationAudit: serviceCommon.GormMutationAuditWriter{},
		Now:           time.Now,
	}
}

// database 返回当前服务使用的数据库连接。
func (service *SubscriptionService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回统一使用的UTC当前时间。
func (service *SubscriptionService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// subscriptionUserReference 将用户模型转换为安全的用户引用。
func subscriptionUserReference(user userModel.MiniAppUser) commonResponse.UserReference {
	return commonResponse.UserReference{
		ID:        user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Status:    string(user.Status),
	}
}

// notificationAdminSummary 将站内通知模型转换为管理端摘要。
func notificationAdminSummary(
	row engagementModel.UserNotification,
) engagementResponse.NotificationAdminSummary {
	return engagementResponse.NotificationAdminSummary{
		ID:         row.ID,
		User:       subscriptionUserReference(row.User),
		Type:       row.Type,
		Title:      row.Title,
		Read:       row.ReadAt != nil,
		TargetType: row.TargetType,
		TargetID:   row.TargetID,
		CreatedAt:  row.CreatedAt,
		ReadAt:     row.ReadAt,
	}
}

// notificationGenerationSource 将通知类型转换为管理端可读的生成来源。
func notificationGenerationSource(notificationType string) string {
	switch notificationType {
	case "governance":
		return "违规处理"
	case "discoverability":
		return "公开状态变更"
	case "points":
		return "积分业务"
	case "feature_refund":
		return "功能处理失败退款"
	case "meal":
		return "饭局状态变更"
	default:
		return "业务系统"
	}
}

// validateNotificationAdminListQuery 校验站内通知列表筛选、分页和排序参数。
func validateNotificationAdminListQuery(query engagementRequest.NotificationAdminListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.UserID))) > 64 ||
		len([]rune(strings.TrimSpace(query.TargetType))) > 40 ||
		len([]rune(strings.TrimSpace(query.TargetID))) > 64 ||
		(query.SortBy != "createdAt" && query.SortBy != "readAt") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	switch query.Type {
	case "", "governance", "discoverability", "points", "feature_refund", "meal":
	default:
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// validateNotificationIdentity 校验站内通知公开ID。
func validateNotificationIdentity(notificationID string) error {
	length := len([]rune(strings.TrimSpace(notificationID)))
	if length < 1 || length > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// ListNotifications 分页查询站内通知投递记录。
func (service *SubscriptionService) ListNotifications(
	ctx context.Context,
	query engagementRequest.NotificationAdminListQuery,
) (commonResponse.Page[engagementResponse.NotificationAdminSummary], error) {
	query.ApplyDefaults()
	if err := validateNotificationAdminListQuery(query); err != nil {
		return commonResponse.Page[engagementResponse.NotificationAdminSummary]{}, err
	}
	sortColumns := map[string]string{
		"createdAt": "of_notifications.created_at",
		"readAt":    "of_notifications.read_at",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[engagementResponse.NotificationAdminSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[engagementResponse.NotificationAdminSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&engagementModel.UserNotification{})
	if value := strings.TrimSpace(query.UserID); value != "" {
		statement = statement.Where("of_notifications.user_id = ?", value)
	}
	if query.Type != "" {
		statement = statement.Where("of_notifications.type = ?", query.Type)
	}
	if query.Read != nil {
		if *query.Read {
			statement = statement.Where("of_notifications.read_at IS NOT NULL")
		} else {
			statement = statement.Where("of_notifications.read_at IS NULL")
		}
	}
	if value := strings.TrimSpace(query.TargetType); value != "" {
		statement = statement.Where("of_notifications.target_type = ?", value)
	}
	if value := strings.TrimSpace(query.TargetID); value != "" {
		statement = statement.Where("of_notifications.target_id = ?", value)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("of_notifications.created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("of_notifications.created_at <= ?", *query.CreatedTo)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[engagementResponse.NotificationAdminSummary]{}, appErrors.AdminInternal.Wrap(err, "count notifications")
	}
	rows := make([]engagementModel.UserNotification, 0)
	if err := statement.Preload("User").
		Order(column + " " + query.SortOrder).
		Order("of_notifications.id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[engagementResponse.NotificationAdminSummary]{}, appErrors.AdminInternal.Wrap(err, "list notifications")
	}
	list := make([]engagementResponse.NotificationAdminSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, notificationAdminSummary(row))
	}
	return commonResponse.Page[engagementResponse.NotificationAdminSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetNotification 获取站内通知投递详情，不改变用户已读状态。
func (service *SubscriptionService) GetNotification(
	ctx context.Context,
	notificationID string,
) (engagementResponse.NotificationAdminDetail, error) {
	if err := validateNotificationIdentity(notificationID); err != nil {
		return engagementResponse.NotificationAdminDetail{}, err
	}
	db := service.database()
	if db == nil {
		return engagementResponse.NotificationAdminDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var row engagementModel.UserNotification
	if err := db.WithContext(ctx).Preload("User").
		First(&row, "id = ?", strings.TrimSpace(notificationID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return engagementResponse.NotificationAdminDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return engagementResponse.NotificationAdminDetail{}, appErrors.AdminInternal.Wrap(err, "load notification")
	}
	return engagementResponse.NotificationAdminDetail{
		NotificationAdminSummary: notificationAdminSummary(row),
		Content:                  row.Content,
		GenerationSource:         notificationGenerationSource(row.Type),
		SubscribeRequired:        row.SubscribeRequired,
		SubscribeLogID:           row.SubscribeLogID,
	}, nil
}

// maskWechatTemplateID 对微信模板ID进行中间部分脱敏。
func maskWechatTemplateID(templateID string) string {
	value := strings.TrimSpace(templateID)
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}

// decodeTemplateMappings 解析模板字段映射JSON。
func decodeTemplateMappings(raw datatypes.JSON) (map[string]string, error) {
	result := make(map[string]string)
	if len(raw) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "decode subscription template field mappings")
	}
	return result, nil
}

// subscribeTemplateDetail 将订阅消息模板模型转换为详情响应。
func subscribeTemplateDetail(
	db *gorm.DB,
	row engagementModel.SubscribeMessageTemplate,
) (engagementResponse.SubscribeTemplateDetail, error) {
	mappings, err := decodeTemplateMappings(row.FieldMappings)
	if err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, err
	}
	var logCount int64
	if err := db.Model(&engagementModel.SubscribeMessageLog{}).
		Where("template_id = ?", row.ID).
		Count(&logCount).Error; err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, appErrors.AdminInternal.Wrap(
			err,
			"count subscription template logs",
		)
	}
	mappingSummary := make([]string, 0, len(mappings))
	for key, value := range mappings {
		mappingSummary = append(mappingSummary, key+"→"+value)
	}
	sort.Strings(mappingSummary)
	return engagementResponse.SubscribeTemplateDetail{
		SubscribeTemplateSummary: engagementResponse.SubscribeTemplateSummary{
			ID:                     row.ID,
			Name:                   row.Name,
			WechatTemplateIDMasked: maskWechatTemplateID(row.WechatTemplateID),
			Scene:                  row.Scene,
			Purpose:                row.Purpose,
			FieldMappingSummary:    mappingSummary,
			Enabled:                row.Enabled,
			LogCount:               logCount,
			SendCount:              row.SendCount,
			FailureCount:           row.FailureCount,
			Version:                row.Version,
			UpdatedAt:              row.UpdatedAt,
		},
		WechatTemplateID: row.WechatTemplateID,
		FieldMappings:    mappings,
		CreatedAt:        row.CreatedAt,
	}, nil
}

// subscribeSceneDetail 将固定场景定义与可选的当前模板绑定合并为响应。
func subscribeSceneDetail(
	db *gorm.DB,
	definition subscribeSceneDefinition,
	row *engagementModel.SubscribeMessageTemplate,
) (engagementResponse.SubscribeSceneDetail, error) {
	summary := engagementResponse.SubscribeSceneSummary{
		Scene:                definition.Scene,
		Name:                 definition.Name,
		Purpose:              definition.Purpose,
		TriggerDescription:   definition.TriggerDescription,
		RecipientDescription: definition.RecipientDescription,
		RequiredFields: append(
			[]engagementResponse.SubscribeSceneFieldDefinition(nil),
			definition.RequiredFields...,
		),
		FieldMappingSummary: make([]string, 0),
	}
	detail := engagementResponse.SubscribeSceneDetail{
		SubscribeSceneSummary: summary,
		FieldMappings:         make(map[string]string),
	}
	if row == nil {
		return detail, nil
	}
	templateDetail, err := subscribeTemplateDetail(db, *row)
	if err != nil {
		return engagementResponse.SubscribeSceneDetail{}, err
	}
	templateID := row.ID
	maskedTemplateID := templateDetail.WechatTemplateIDMasked
	updatedAt := row.UpdatedAt
	detail.Configured = true
	detail.TemplateID = &templateID
	detail.WechatTemplateIDMasked = &maskedTemplateID
	detail.FieldMappingSummary = templateDetail.FieldMappingSummary
	detail.Enabled = row.Enabled
	detail.LogCount = templateDetail.LogCount
	detail.SendCount = row.SendCount
	detail.FailureCount = row.FailureCount
	detail.Version = row.Version
	detail.UpdatedAt = &updatedAt
	detail.WechatTemplateID = row.WechatTemplateID
	detail.FieldMappings = templateDetail.FieldMappings
	return detail, nil
}

// ListSubscribeScenes 返回代码定义的固定订阅消息场景及当前绑定。
func (service *SubscriptionService) ListSubscribeScenes(
	ctx context.Context,
) ([]engagementResponse.SubscribeSceneSummary, error) {
	db := service.database()
	if db == nil {
		return nil, appErrors.AdminInternal.DefaultMsg()
	}
	result := make([]engagementResponse.SubscribeSceneSummary, 0, len(fixedSubscribeSceneDefinitions))
	for _, definition := range fixedSubscribeSceneDefinitions {
		var row engagementModel.SubscribeMessageTemplate
		err := db.WithContext(ctx).First(&row, "scene = ?", definition.Scene).Error
		var rowPointer *engagementModel.SubscribeMessageTemplate
		if err == nil {
			rowPointer = &row
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.AdminInternal.Wrap(err, "load subscription scene binding")
		}
		detail, err := subscribeSceneDetail(db.WithContext(ctx), definition, rowPointer)
		if err != nil {
			return nil, err
		}
		result = append(result, detail.SubscribeSceneSummary)
	}
	return result, nil
}

// GetSubscribeScene 返回一个固定订阅消息场景及当前完整绑定。
func (service *SubscriptionService) GetSubscribeScene(
	ctx context.Context,
	scene string,
) (engagementResponse.SubscribeSceneDetail, error) {
	definition, ok := fixedSubscribeSceneDefinition(scene)
	if !ok {
		return engagementResponse.SubscribeSceneDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return engagementResponse.SubscribeSceneDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var row engagementModel.SubscribeMessageTemplate
	err := db.WithContext(ctx).First(&row, "scene = ?", definition.Scene).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return subscribeSceneDetail(db.WithContext(ctx), definition, nil)
	}
	if err != nil {
		return engagementResponse.SubscribeSceneDetail{}, appErrors.AdminInternal.Wrap(
			err,
			"load subscription scene binding",
		)
	}
	return subscribeSceneDetail(db.WithContext(ctx), definition, &row)
}

// normalizeSubscribeSceneInput 清理固定场景模板绑定输入。
func normalizeSubscribeSceneInput(
	input engagementRequest.SubscribeSceneConfigureInput,
) (engagementRequest.SubscribeSceneConfigureInput, error) {
	input.WechatTemplateID = strings.TrimSpace(input.WechatTemplateID)
	normalized := make(map[string]string, len(input.FieldMappings))
	for key, value := range input.FieldMappings {
		normalizedKey := strings.TrimSpace(key)
		if _, exists := normalized[normalizedKey]; exists {
			return input, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		normalized[normalizedKey] = strings.TrimSpace(value)
	}
	input.FieldMappings = normalized
	return input, nil
}

// validateSubscribeSceneInput 校验固定场景及其模板绑定。
func validateSubscribeSceneInput(
	scene string,
	input engagementRequest.SubscribeSceneConfigureInput,
) error {
	if _, ok := fixedSubscribeSceneDefinition(scene); !ok {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	length := len([]rune(input.WechatTemplateID))
	if length < 1 || length > 100 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return validateTemplateMappings(strings.TrimSpace(scene), input.FieldMappings)
}

// normalizeTemplateInput 清理模板输入中的首尾空白。
func normalizeTemplateInput(
	input engagementRequest.SubscribeTemplateCreateInput,
) (engagementRequest.SubscribeTemplateCreateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.WechatTemplateID = strings.TrimSpace(input.WechatTemplateID)
	input.Scene = strings.TrimSpace(input.Scene)
	input.Purpose = strings.TrimSpace(input.Purpose)
	normalized := make(map[string]string, len(input.FieldMappings))
	for key, value := range input.FieldMappings {
		normalizedKey := strings.TrimSpace(key)
		if _, exists := normalized[normalizedKey]; exists {
			return input, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		normalized[normalizedKey] = strings.TrimSpace(value)
	}
	input.FieldMappings = normalized
	return input, nil
}

// validateTemplateMappings 校验业务必需字段、空值和重复微信字段。
func validateTemplateMappings(scene string, mappings map[string]string) error {
	if scene != engagementModel.SubscribeSceneMealStatus ||
		len(mappings) != len(requiredMealStatusMappings) {
		return appErrors.AdminInvalidConfig.DefaultMsg()
	}
	usedWechatFields := make(map[string]struct{}, len(mappings))
	for key, value := range mappings {
		if key == "" || value == "" ||
			len([]rune(key)) > 64 ||
			len([]rune(value)) > 40 {
			return appErrors.AdminInvalidConfig.DefaultMsg()
		}
		if _, exists := usedWechatFields[value]; exists {
			return appErrors.AdminInvalidConfig.DefaultMsg()
		}
		usedWechatFields[value] = struct{}{}
	}
	for _, required := range requiredMealStatusMappings {
		if strings.TrimSpace(mappings[required]) == "" {
			return appErrors.AdminInvalidConfig.DefaultMsg()
		}
	}
	return nil
}

// validateSubscribeTemplateListQuery 校验订阅消息模板列表参数。
func validateSubscribeTemplateListQuery(query engagementRequest.SubscribeTemplateListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.Keyword))) > 60 ||
		(query.Scene != "" && query.Scene != engagementModel.SubscribeSceneMealStatus) ||
		(query.SortBy != "updatedAt" && query.SortBy != "createdAt" && query.SortBy != "scene") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// validateSubscribeTemplateIdentity 校验订阅消息模板公开ID。
func validateSubscribeTemplateIdentity(templateID string) error {
	length := len([]rune(strings.TrimSpace(templateID)))
	if length < 1 || length > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// validateSubscribeTemplateInput 校验订阅消息模板基础配置。
func validateSubscribeTemplateInput(input engagementRequest.SubscribeTemplateCreateInput) error {
	nameLength := len([]rune(input.Name))
	wechatTemplateIDLength := len([]rune(input.WechatTemplateID))
	purposeLength := len([]rune(input.Purpose))
	if nameLength < 1 || nameLength > 60 ||
		wechatTemplateIDLength < 1 || wechatTemplateIDLength > 100 ||
		purposeLength < 1 || purposeLength > 160 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return validateTemplateMappings(input.Scene, input.FieldMappings)
}

// validateSubscriptionReason 校验订阅消息模板启停或删除原因。
func validateSubscriptionReason(reason string) error {
	length := len([]rune(strings.TrimSpace(reason)))
	if length < 4 || length > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// encodeTemplateMappings 将字段映射编码为数据库JSON。
func encodeTemplateMappings(mappings map[string]string) (datatypes.JSON, error) {
	raw, err := json.Marshal(mappings)
	if err != nil {
		return nil, appErrors.AdminBadRequest.Wrap(err, "encode subscription template field mappings")
	}
	return datatypes.JSON(raw), nil
}

// ConfigureSubscribeScene 创建或更新固定场景的唯一当前模板绑定。
func (service *SubscriptionService) ConfigureSubscribeScene(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	scene string,
	input engagementRequest.SubscribeSceneConfigureInput,
) (engagementResponse.SubscribeSceneDetail, bool, error) {
	if err := validateSubscriptionActor(actor); err != nil {
		return engagementResponse.SubscribeSceneDetail{}, false, err
	}
	definition, ok := fixedSubscribeSceneDefinition(scene)
	if !ok {
		return engagementResponse.SubscribeSceneDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	var err error
	input, err = normalizeSubscribeSceneInput(input)
	if err != nil {
		return engagementResponse.SubscribeSceneDetail{}, false, err
	}
	if err := validateSubscribeSceneInput(definition.Scene, input); err != nil {
		return engagementResponse.SubscribeSceneDetail{}, false, err
	}
	mappings, err := encodeTemplateMappings(input.FieldMappings)
	if err != nil {
		return engagementResponse.SubscribeSceneDetail{}, false, err
	}
	payload := struct {
		Scene string
		Input engagementRequest.SubscribeSceneConfigureInput
	}{Scene: definition.Scene, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"subscribe_scene_configure",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var row engagementModel.SubscribeMessageTemplate
			loadErr := tx.
				First(&row, "scene = ?", definition.Scene).Error
			if loadErr != nil && !errors.Is(loadErr, gorm.ErrRecordNotFound) {
				return nil, appErrors.AdminInternal.Wrap(loadErr, "load subscription scene binding")
			}
			var duplicateCount int64
			duplicateQuery := tx.Model(&engagementModel.SubscribeMessageTemplate{}).
				Where("wechat_template_id = ?", input.WechatTemplateID)
			if loadErr == nil {
				duplicateQuery = duplicateQuery.Where("id <> ?", row.ID)
			}
			if err := duplicateQuery.Count(&duplicateCount).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check duplicate subscription template id")
			}
			if duplicateCount > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}

			if errors.Is(loadErr, gorm.ErrRecordNotFound) {
				if input.ExpectedVersion != nil {
					return nil, appErrors.AdminStateConflict.DefaultMsg()
				}
				now := service.now()
				row = engagementModel.SubscribeMessageTemplate{
					ID:               serviceCommon.NewPublicID("subtpl"),
					Name:             definition.Name,
					WechatTemplateID: input.WechatTemplateID,
					Scene:            definition.Scene,
					Purpose:          definition.Purpose,
					FieldMappings:    mappings,
					Enabled:          false,
					Version:          1,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				if err := tx.Create(&row).Error; err != nil {
					return nil, appErrors.AdminStateConflict.Wrap(err, "create subscription scene binding")
				}
				after, err := subscribeSceneDetail(tx, definition, &row)
				if err != nil {
					return nil, err
				}
				if err := service.writeMutationAudit(
					ctx,
					tx,
					actor,
					"subscribe_scene_configure",
					definition.Scene,
					"首次配置固定订阅场景",
					idempotencyKey,
					nil,
					after,
				); err != nil {
					return nil, err
				}
				return after, nil
			}

			if input.ExpectedVersion == nil || row.Version != *input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			before, err := subscribeSceneDetail(tx, definition, &row)
			if err != nil {
				return nil, err
			}
			if err := tx.Model(&row).Updates(map[string]interface{}{
				"name":               definition.Name,
				"wechat_template_id": input.WechatTemplateID,
				"purpose":            definition.Purpose,
				"field_mappings":     mappings,
				"version":            row.Version + 1,
				"updated_at":         service.now(),
			}).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update subscription scene binding")
			}
			if err := tx.First(&row, "id = ?", row.ID).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "reload subscription scene binding")
			}
			after, err := subscribeSceneDetail(tx, definition, &row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"subscribe_scene_configure",
				definition.Scene,
				"更新固定订阅场景模板绑定",
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[engagementResponse.SubscribeSceneDetail](raw, replayed, err)
}

// UpdateSubscribeSceneStatus 启用或停用固定场景的当前模板绑定。
func (service *SubscriptionService) UpdateSubscribeSceneStatus(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	scene string,
	input engagementRequest.SubscribeSceneStatusInput,
) (engagementResponse.SubscribeSceneDetail, bool, error) {
	if err := validateSubscriptionActor(actor); err != nil {
		return engagementResponse.SubscribeSceneDetail{}, false, err
	}
	definition, ok := fixedSubscribeSceneDefinition(scene)
	if !ok || input.Enabled == nil || input.ExpectedVersion < 1 {
		return engagementResponse.SubscribeSceneDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateSubscriptionReason(input.Reason); err != nil {
		return engagementResponse.SubscribeSceneDetail{}, false, err
	}
	payload := struct {
		Scene string
		Input engagementRequest.SubscribeSceneStatusInput
	}{Scene: definition.Scene, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"subscribe_scene_status_update",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var row engagementModel.SubscribeMessageTemplate
			if err := tx.
				First(&row, "scene = ?", definition.Scene).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminInvalidConfig.New("请先配置该订阅场景的微信模板")
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load subscription scene binding status")
			}
			if row.Version != input.ExpectedVersion || row.Enabled == *input.Enabled {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if *input.Enabled {
				mappings, err := decodeTemplateMappings(row.FieldMappings)
				if err != nil {
					return nil, err
				}
				if err := validateTemplateMappings(definition.Scene, mappings); err != nil {
					return nil, err
				}
			}
			before, err := subscribeSceneDetail(tx, definition, &row)
			if err != nil {
				return nil, err
			}
			if err := tx.Model(&row).Updates(map[string]interface{}{
				"enabled":    *input.Enabled,
				"version":    row.Version + 1,
				"updated_at": service.now(),
			}).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update subscription scene status")
			}
			if err := tx.First(&row, "id = ?", row.ID).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "reload subscription scene status")
			}
			after, err := subscribeSceneDetail(tx, definition, &row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"subscribe_scene_status_update",
				definition.Scene,
				input.Reason,
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[engagementResponse.SubscribeSceneDetail](raw, replayed, err)
}

// ListSubscribeTemplates 分页查询订阅消息模板配置。
func (service *SubscriptionService) ListSubscribeTemplates(
	ctx context.Context,
	query engagementRequest.SubscribeTemplateListQuery,
) (commonResponse.Page[engagementResponse.SubscribeTemplateSummary], error) {
	query.ApplyDefaults()
	if err := validateSubscribeTemplateListQuery(query); err != nil {
		return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{}, err
	}
	sortColumns := map[string]string{
		"updatedAt": "updated_at",
		"createdAt": "created_at",
		"scene":     "scene",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&engagementModel.SubscribeMessageTemplate{})
	if value := strings.TrimSpace(query.Keyword); value != "" {
		search := "%" + value + "%"
		statement = statement.Where("name LIKE ? OR purpose LIKE ?", search, search)
	}
	if query.Scene != "" {
		statement = statement.Where("scene = ?", query.Scene)
	}
	if query.Enabled != nil {
		statement = statement.Where("enabled = ?", *query.Enabled)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{}, appErrors.AdminInternal.Wrap(err, "count subscription templates")
	}
	rows := make([]engagementModel.SubscribeMessageTemplate, 0)
	if err := statement.Order(column + " " + query.SortOrder).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{}, appErrors.AdminInternal.Wrap(err, "list subscription templates")
	}
	list := make([]engagementResponse.SubscribeTemplateSummary, 0, len(rows))
	for _, row := range rows {
		detail, err := subscribeTemplateDetail(db, row)
		if err != nil {
			return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{}, err
		}
		list = append(list, detail.SubscribeTemplateSummary)
	}
	return commonResponse.Page[engagementResponse.SubscribeTemplateSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// GetSubscribeTemplate 获取订阅消息模板详情。
func (service *SubscriptionService) GetSubscribeTemplate(
	ctx context.Context,
	templateID string,
) (engagementResponse.SubscribeTemplateDetail, error) {
	if err := validateSubscribeTemplateIdentity(templateID); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, err
	}
	db := service.database()
	if db == nil {
		return engagementResponse.SubscribeTemplateDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var row engagementModel.SubscribeMessageTemplate
	if err := db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(templateID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return engagementResponse.SubscribeTemplateDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return engagementResponse.SubscribeTemplateDetail{}, appErrors.AdminInternal.Wrap(err, "load subscription template")
	}
	return subscribeTemplateDetail(db, row)
}

// validateSubscriptionActor 校验管理操作上下文。
func validateSubscriptionActor(actor commonRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// writeMutationAudit 在同一事务内记录订阅模板变更审计。
func (service *SubscriptionService) writeMutationAudit(
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
	reasonValue := serviceCommon.TruncateRunes(strings.TrimSpace(reason), 200)
	var reasonPointer *string
	if reasonValue != "" {
		reasonPointer = &reasonValue
	}
	key := strings.TrimSpace(idempotencyKey)
	audit := auditModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            "subscribe_scene",
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

// CreateSubscribeTemplate 创建默认停用的订阅消息模板。
func (service *SubscriptionService) CreateSubscribeTemplate(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	input engagementRequest.SubscribeTemplateCreateInput,
) (engagementResponse.SubscribeTemplateDetail, bool, error) {
	if err := validateSubscriptionActor(actor); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	var err error
	input, err = normalizeTemplateInput(input)
	if err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	if err := validateSubscribeTemplateInput(input); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	mappings, err := encodeTemplateMappings(input.FieldMappings)
	if err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "subscribe_template_create", strings.TrimSpace(idempotencyKey), input,
		func(tx *gorm.DB) (interface{}, error) {
			var count int64
			if err := tx.Model(&engagementModel.SubscribeMessageTemplate{}).
				Where("wechat_template_id = ?", input.WechatTemplateID).
				Count(&count).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check subscription template id")
			}
			if count > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			now := service.now()
			row := engagementModel.SubscribeMessageTemplate{
				ID:               serviceCommon.NewPublicID("subtpl"),
				Name:             input.Name,
				WechatTemplateID: input.WechatTemplateID,
				Scene:            input.Scene,
				Purpose:          input.Purpose,
				FieldMappings:    mappings,
				Enabled:          false,
				Version:          1,
				CreatedAt:        now,
				UpdatedAt:        now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create subscription template")
			}
			detail, err := subscribeTemplateDetail(tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "subscribe_template_create", row.ID,
				"创建订阅消息模板", idempotencyKey, nil, detail,
			); err != nil {
				return nil, err
			}
			return detail, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[engagementResponse.SubscribeTemplateDetail](raw, replayed, err)
}

// UpdateSubscribeTemplate 编辑订阅消息模板，但不改变启用状态。
func (service *SubscriptionService) UpdateSubscribeTemplate(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	templateID string,
	input engagementRequest.SubscribeTemplateUpdateInput,
) (engagementResponse.SubscribeTemplateDetail, bool, error) {
	if err := validateSubscriptionActor(actor); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	if err := validateSubscribeTemplateIdentity(templateID); err != nil ||
		input.ExpectedVersion < 1 {
		return engagementResponse.SubscribeTemplateDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	templateID = strings.TrimSpace(templateID)
	var err error
	input.SubscribeTemplateCreateInput, err = normalizeTemplateInput(input.SubscribeTemplateCreateInput)
	if err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	if err := validateSubscribeTemplateInput(input.SubscribeTemplateCreateInput); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	mappings, err := encodeTemplateMappings(input.FieldMappings)
	if err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	payload := struct {
		TemplateID string
		Input      engagementRequest.SubscribeTemplateUpdateInput
	}{TemplateID: templateID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "subscribe_template_update", strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			var row engagementModel.SubscribeMessageTemplate
			if err := tx.
				First(&row, "id = ?", templateID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load subscription template")
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			var duplicateCount int64
			if err := tx.Model(&engagementModel.SubscribeMessageTemplate{}).
				Where("wechat_template_id = ? AND id <> ?", input.WechatTemplateID, row.ID).
				Count(&duplicateCount).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "check duplicate subscription template id")
			}
			if duplicateCount > 0 {
				return nil, appErrors.AdminAlreadyExists.DefaultMsg()
			}
			before, err := subscribeTemplateDetail(tx, row)
			if err != nil {
				return nil, err
			}
			if err := tx.Model(&row).Updates(map[string]interface{}{
				"name":               input.Name,
				"wechat_template_id": input.WechatTemplateID,
				"scene":              input.Scene,
				"purpose":            input.Purpose,
				"field_mappings":     mappings,
				"version":            row.Version + 1,
				"updated_at":         service.now(),
			}).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update subscription template")
			}
			if err := tx.First(&row, "id = ?", row.ID).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "reload subscription template")
			}
			after, err := subscribeTemplateDetail(tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "subscribe_template_update", row.ID,
				"编辑订阅消息模板", idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[engagementResponse.SubscribeTemplateDetail](raw, replayed, err)
}

// UpdateSubscribeTemplateStatus 启用或停用订阅消息模板。
func (service *SubscriptionService) UpdateSubscribeTemplateStatus(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	templateID string,
	input engagementRequest.SubscribeTemplateStatusInput,
) (engagementResponse.SubscribeTemplateDetail, bool, error) {
	if err := validateSubscriptionActor(actor); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	if err := validateSubscribeTemplateIdentity(templateID); err != nil ||
		input.Enabled == nil ||
		input.ExpectedVersion < 1 {
		return engagementResponse.SubscribeTemplateDetail{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateSubscriptionReason(input.Reason); err != nil {
		return engagementResponse.SubscribeTemplateDetail{}, false, err
	}
	templateID = strings.TrimSpace(templateID)
	payload := struct {
		TemplateID string
		Input      engagementRequest.SubscribeTemplateStatusInput
	}{TemplateID: templateID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "subscribe_template_status_update", strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			var row engagementModel.SubscribeMessageTemplate
			if err := tx.
				First(&row, "id = ?", templateID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load subscription template")
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if row.Enabled == *input.Enabled {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if *input.Enabled {
				mappings, err := decodeTemplateMappings(row.FieldMappings)
				if err != nil {
					return nil, err
				}
				if err := validateTemplateMappings(row.Scene, mappings); err != nil {
					return nil, err
				}
				var enabledCount int64
				if err := tx.Model(&engagementModel.SubscribeMessageTemplate{}).
					Where("scene = ? AND enabled = ? AND id <> ?", row.Scene, true, row.ID).
					Count(&enabledCount).Error; err != nil {
					return nil, appErrors.AdminInternal.Wrap(err, "check enabled subscription template")
				}
				if enabledCount > 0 {
					return nil, appErrors.AdminInvalidConfig.New("同一订阅场景只能启用一个模板")
				}
			}
			before, err := subscribeTemplateDetail(tx, row)
			if err != nil {
				return nil, err
			}
			if err := tx.Model(&row).Updates(map[string]interface{}{
				"enabled":    *input.Enabled,
				"version":    row.Version + 1,
				"updated_at": service.now(),
			}).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update subscription template status")
			}
			if err := tx.First(&row, "id = ?", row.ID).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "reload subscription template status")
			}
			after, err := subscribeTemplateDetail(tx, row)
			if err != nil {
				return nil, err
			}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "subscribe_template_status_update", row.ID,
				input.Reason, idempotencyKey, before, after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[engagementResponse.SubscribeTemplateDetail](raw, replayed, err)
}

// DeleteSubscribeTemplate 删除未启用且没有发送记录引用的订阅消息模板。
func (service *SubscriptionService) DeleteSubscribeTemplate(
	ctx context.Context,
	actor commonRequest.AdminActor,
	idempotencyKey string,
	templateID string,
	input engagementRequest.SubscribeTemplateDeleteInput,
) (aiResponse.DeletedResult, bool, error) {
	if err := validateSubscriptionActor(actor); err != nil {
		return aiResponse.DeletedResult{}, false, err
	}
	if err := validateSubscribeTemplateIdentity(templateID); err != nil ||
		input.ExpectedVersion < 1 {
		return aiResponse.DeletedResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateSubscriptionReason(input.Reason); err != nil {
		return aiResponse.DeletedResult{}, false, err
	}
	templateID = strings.TrimSpace(templateID)
	payload := struct {
		TemplateID string
		Input      engagementRequest.SubscribeTemplateDeleteInput
	}{TemplateID: templateID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx, actor.AdministratorID, "subscribe_template_delete", strings.TrimSpace(idempotencyKey), payload,
		func(tx *gorm.DB) (interface{}, error) {
			var row engagementModel.SubscribeMessageTemplate
			if err := tx.
				First(&row, "id = ?", templateID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "load subscription template for deletion")
			}
			if row.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			if row.Enabled {
				return nil, appErrors.AdminStateConflict.New("启用中的订阅消息模板不能删除")
			}
			var logCount int64
			if err := tx.Model(&engagementModel.SubscribeMessageLog{}).
				Where("template_id = ?", row.ID).
				Count(&logCount).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "count subscription template logs")
			}
			if logCount > 0 {
				return nil, appErrors.AdminResourceInUse.DefaultMsg()
			}
			before, err := subscribeTemplateDetail(tx, row)
			if err != nil {
				return nil, err
			}
			deletion := tx.Where("id = ? AND version = ?", row.ID, input.ExpectedVersion).
				Delete(&engagementModel.SubscribeMessageTemplate{})
			if deletion.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(deletion.Error, "delete subscription template")
			}
			if deletion.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			result := aiResponse.DeletedResult{Deleted: true}
			if err := service.writeMutationAudit(
				ctx, tx, actor, "subscribe_template_delete", row.ID,
				input.Reason, idempotencyKey, before, result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	return serviceCommon.DecodeIdempotentResult[aiResponse.DeletedResult](raw, replayed, err)
}

type subscribeLogRelations struct {
	users     map[string]userModel.MiniAppUser                    // 用户ID到用户模型的映射
	templates map[string]engagementModel.SubscribeMessageTemplate // 模板ID到模板模型的映射
}

// loadSubscribeLogRelations 批量加载发送记录所需的用户和模板数据。
func (service *SubscriptionService) loadSubscribeLogRelations(
	ctx context.Context,
	logs []engagementModel.SubscribeMessageLog,
) (subscribeLogRelations, error) {
	result := subscribeLogRelations{
		users:     make(map[string]userModel.MiniAppUser),
		templates: make(map[string]engagementModel.SubscribeMessageTemplate),
	}
	if len(logs) == 0 {
		return result, nil
	}
	userIDs := make([]string, 0, len(logs))
	templateIDs := make([]string, 0, len(logs))
	for _, row := range logs {
		userIDs = append(userIDs, row.UserID)
		templateIDs = append(templateIDs, row.TemplateID)
	}
	var users []userModel.MiniAppUser
	if err := service.database().WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "load subscription log users")
	}
	for _, user := range users {
		result.users[user.ID] = user
	}
	var templates []engagementModel.SubscribeMessageTemplate
	if err := service.database().WithContext(ctx).Where("id IN ?", templateIDs).Find(&templates).Error; err != nil {
		return result, appErrors.AdminInternal.Wrap(err, "load subscription log templates")
	}
	for _, template := range templates {
		result.templates[template.ID] = template
	}
	return result, nil
}

// subscribeLogSummary 将订阅消息发送记录转换为管理端摘要。
func subscribeLogSummary(
	row engagementModel.SubscribeMessageLog,
	relations subscribeLogRelations,
) engagementResponse.SubscribeLogSummary {
	user := relations.users[row.UserID]
	template := relations.templates[row.TemplateID]
	return engagementResponse.SubscribeLogSummary{
		ID:              row.ID,
		User:            subscriptionUserReference(user),
		TemplateID:      row.TemplateID,
		TemplateName:    template.Name,
		Scene:           row.Scene,
		Status:          row.Status,
		RetryCount:      row.RetryCount,
		WechatErrorCode: row.WechatErrorCode,
		RequestID:       row.RequestID,
		CreatedAt:       row.CreatedAt,
		SentAt:          row.SentAt,
	}
}

// validateSubscribeLogListQuery 校验订阅消息发送记录列表参数。
func validateSubscribeLogListQuery(query engagementRequest.SubscribeLogListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.UserID))) > 64 ||
		len([]rune(strings.TrimSpace(query.TemplateID))) > 64 ||
		len([]rune(strings.TrimSpace(query.RequestID))) > 128 ||
		(query.Scene != "" && query.Scene != engagementModel.SubscribeSceneMealStatus) ||
		(query.SortBy != "createdAt" && query.SortBy != "sentAt" && query.SortBy != "retryCount") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	switch query.Status {
	case "", engagementModel.SubscribeLogPending, engagementModel.SubscribeLogSending,
		engagementModel.SubscribeLogSent, engagementModel.SubscribeLogFailed:
	default:
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// validateSubscribeLogIdentity 校验订阅消息发送记录公开ID。
func validateSubscribeLogIdentity(logID string) error {
	length := len([]rune(strings.TrimSpace(logID)))
	if length < 1 || length > 64 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// ListSubscribeLogs 分页查询订阅消息发送记录。
func (service *SubscriptionService) ListSubscribeLogs(
	ctx context.Context,
	query engagementRequest.SubscribeLogListQuery,
) (commonResponse.Page[engagementResponse.SubscribeLogSummary], error) {
	query.ApplyDefaults()
	if err := validateSubscribeLogListQuery(query); err != nil {
		return commonResponse.Page[engagementResponse.SubscribeLogSummary]{}, err
	}
	sortColumns := map[string]string{
		"createdAt":  "created_at",
		"sentAt":     "sent_at",
		"retryCount": "retry_count",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[engagementResponse.SubscribeLogSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[engagementResponse.SubscribeLogSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&engagementModel.SubscribeMessageLog{})
	if value := strings.TrimSpace(query.UserID); value != "" {
		statement = statement.Where("user_id = ?", value)
	}
	if value := strings.TrimSpace(query.TemplateID); value != "" {
		statement = statement.Where("template_id = ?", value)
	}
	if query.Scene != "" {
		statement = statement.Where("scene = ?", query.Scene)
	}
	if query.Status != "" {
		statement = statement.Where("status = ?", query.Status)
	}
	if value := strings.TrimSpace(query.RequestID); value != "" {
		statement = statement.Where("request_id = ?", value)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", *query.CreatedTo)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[engagementResponse.SubscribeLogSummary]{}, appErrors.AdminInternal.Wrap(err, "count subscription logs")
	}
	rows := make([]engagementModel.SubscribeMessageLog, 0)
	if err := statement.Order(column + " " + query.SortOrder).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[engagementResponse.SubscribeLogSummary]{}, appErrors.AdminInternal.Wrap(err, "list subscription logs")
	}
	relations, err := service.loadSubscribeLogRelations(ctx, rows)
	if err != nil {
		return commonResponse.Page[engagementResponse.SubscribeLogSummary]{}, err
	}
	list := make([]engagementResponse.SubscribeLogSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, subscribeLogSummary(row, relations))
	}
	return commonResponse.Page[engagementResponse.SubscribeLogSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// decodeSafePayloadSummary 解析允许管理端查看的脱敏载荷摘要。
func decodeSafePayloadSummary(raw datatypes.JSON) (map[string]interface{}, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "decode safe subscription payload summary")
	}
	return result, nil
}

// GetSubscribeLog 获取订阅消息发送详情及各次尝试记录。
func (service *SubscriptionService) GetSubscribeLog(
	ctx context.Context,
	logID string,
) (engagementResponse.SubscribeLogDetail, error) {
	if err := validateSubscribeLogIdentity(logID); err != nil {
		return engagementResponse.SubscribeLogDetail{}, err
	}
	db := service.database()
	if db == nil {
		return engagementResponse.SubscribeLogDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var row engagementModel.SubscribeMessageLog
	if err := db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(logID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return engagementResponse.SubscribeLogDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return engagementResponse.SubscribeLogDetail{}, appErrors.AdminInternal.Wrap(err, "load subscription log")
	}
	relations, err := service.loadSubscribeLogRelations(ctx, []engagementModel.SubscribeMessageLog{row})
	if err != nil {
		return engagementResponse.SubscribeLogDetail{}, err
	}
	payload, err := decodeSafePayloadSummary(row.SafePayloadSummary)
	if err != nil {
		return engagementResponse.SubscribeLogDetail{}, err
	}
	attemptRows := make([]engagementModel.SubscribeMessageAttempt, 0)
	if err := db.WithContext(ctx).Where("log_id = ?", row.ID).
		Order("attempt asc").Find(&attemptRows).Error; err != nil {
		return engagementResponse.SubscribeLogDetail{}, appErrors.AdminInternal.Wrap(err, "list subscription attempts")
	}
	attempts := make([]engagementResponse.SubscribeAttempt, 0, len(attemptRows))
	for _, attempt := range attemptRows {
		attempts = append(attempts, engagementResponse.SubscribeAttempt{
			Attempt:         attempt.Attempt,
			Status:          attempt.Status,
			WechatErrorCode: attempt.WechatErrorCode,
			ErrorSummary:    attempt.ErrorSummary,
			StartedAt:       attempt.StartedAt,
			FinishedAt:      attempt.FinishedAt,
		})
	}
	var relatedMealID *string
	if value, ok := payload["mealId"].(string); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			relatedMealID = &value
		}
	}
	return engagementResponse.SubscribeLogDetail{
		SubscribeLogSummary:    subscribeLogSummary(row, relations),
		AuthorizationChecked:   row.AuthorizationChecked,
		AuthorizationAvailable: row.AuthorizationAvailable,
		RelatedMealID:          relatedMealID,
		TargetPage:             row.TargetPage,
		SafePayloadSummary:     payload,
		ErrorSummary:           row.ErrorSummary,
		Attempts:               attempts,
	}, nil
}

package orderfood

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// PermissionAIUsageRead 表示读取AI调用记录所需权限。
	PermissionAIUsageRead = "orderfood:ai-usage:read"
	// PermissionAIUsageSensitiveRead 表示读取AI调用敏感内容所需权限。
	PermissionAIUsageSensitiveRead = "orderfood:ai-usage:sensitive-read"
	// PermissionAIUsageSensitiveDelete 表示清除AI调用敏感内容所需权限。
	PermissionAIUsageSensitiveDelete = "orderfood:ai-usage:sensitive-delete"
)

var aiUsageRetainedAuditFields = []string{
	"id", "userId", "capabilityCode", "providerName", "modelName",
	"requestId", "idempotencyKey", "executionStatus", "billingStatus",
	"promptMode", "promptHash", "durationMs", "inputTokens", "outputTokens",
	"imageCount", "pointCost", "estimatedCostCny", "createdAt", "finishedAt",
}

// AIUsageService 提供AI调用记录查询和敏感内容清除能力。
type AIUsageService struct {
	DB          *gorm.DB            // 数据库连接
	Permission  *PermissionService  // 管理端权限服务
	Audit       *AccessAuditService // 敏感访问审计服务
	Idempotency *IdempotencyService // 管理端幂等服务
	Now         func() time.Time    // 当前时间函数
}

// NewAIUsageService 创建AI调用记录服务实例。
func NewAIUsageService(
	db *gorm.DB,
	permission *PermissionService,
	idempotency *IdempotencyService,
) *AIUsageService {
	if permission == nil {
		permission = NewPermissionService(db)
	}
	if idempotency == nil {
		idempotency = NewIdempotencyService(db)
	}
	return &AIUsageService{
		DB: db, Permission: permission, Audit: NewAccessAuditService(db),
		Idempotency: idempotency, Now: time.Now,
	}
}

// database 返回服务使用的数据库连接。
func (service *AIUsageService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回服务使用的UTC时间。
func (service *AIUsageService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// List 分页查询AI调用记录。
func (service *AIUsageService) List(
	ctx context.Context,
	query orderfoodRequest.AIUsageListQuery,
	actor orderfoodRequest.AdminActor,
) (orderfoodResponse.Page[orderfoodResponse.AIUsageSummary], error) {
	query.ApplyDefaults()
	if err := validateAIUsageListQuery(query); err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{}, err
	}
	if service == nil || service.Permission == nil {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionAIUsageRead); err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{}, err
	}
	db := service.database()
	if db == nil {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&orderfoodModel.FrontFeatureUsage{})
	for _, filter := range []struct {
		column string // 数据库字段
		value  string // 查询值
	}{
		{"user_id", query.UserID},
		{"provider_id", query.ProviderID},
		{"model_id", query.ModelID},
		{"request_id", query.RequestID},
		{"idempotency_key", query.IdempotencyKey},
		{"execution_status", query.ExecutionStatus},
		{"billing_status", query.BillingStatus},
	} {
		if value := strings.TrimSpace(filter.value); value != "" {
			statement = statement.Where(filter.column+" = ?", value)
		}
	}
	if value := strings.TrimSpace(query.CapabilityCode); value != "" {
		// 兼容扩展字段上线前仅保存feature的历史记录。
		statement = statement.Where(
			"capability_code = ? OR (capability_code = '' AND feature = ?)", value, value,
		)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", query.CreatedFrom.UTC())
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", query.CreatedTo.UTC())
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{},
			appErrors.AdminInternal.Wrap(err, "count AI usage records")
	}
	orderColumns := map[string]string{
		"createdAt": "created_at", "durationMs": "duration_ms",
		"estimatedCostCny": "estimated_cost_cny",
	}
	orderColumn, ok := orderColumns[query.SortBy]
	if !ok {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var rows []orderfoodModel.FrontFeatureUsage
	if err := statement.
		Order(orderColumn + " " + query.SortOrder).
		Order("id " + query.SortOrder).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{},
			appErrors.AdminInternal.Wrap(err, "list AI usage records")
	}
	list := make([]orderfoodResponse.AIUsageSummary, 0, len(rows))
	for index := range rows {
		summary, err := service.summary(ctx, db, rows[index])
		if err != nil {
			return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{}, err
		}
		list = append(list, summary)
	}
	return orderfoodResponse.Page[orderfoodResponse.AIUsageSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// Detail 获取AI调用记录详情。
func (service *AIUsageService) Detail(
	ctx context.Context,
	usageID string,
	includeSensitive bool,
	actor orderfoodRequest.AdminActor,
) (orderfoodResponse.AIUsageDetail, error) {
	usageID = strings.TrimSpace(usageID)
	if usageID == "" || len([]rune(usageID)) > 64 {
		return orderfoodResponse.AIUsageDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	if service == nil || service.Permission == nil {
		return orderfoodResponse.AIUsageDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionAIUsageRead); err != nil {
		return orderfoodResponse.AIUsageDetail{}, err
	}
	if includeSensitive {
		if err := service.Permission.Require(
			ctx,
			actor.AuthorityID,
			PermissionAIUsageSensitiveRead,
		); err != nil {
			return orderfoodResponse.AIUsageDetail{}, err
		}
	}
	db := service.database()
	if db == nil {
		return orderfoodResponse.AIUsageDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var row orderfoodModel.FrontFeatureUsage
	if err := db.WithContext(ctx).First(&row, "id = ?", usageID).Error; err != nil {
		return orderfoodResponse.AIUsageDetail{}, aiUsageLookupError(err)
	}
	summary, err := service.summary(ctx, db, row)
	if err != nil {
		return orderfoodResponse.AIUsageDetail{}, err
	}
	pointEntries, err := usagePointEntries(ctx, db, row.ID)
	if err != nil {
		return orderfoodResponse.AIUsageDetail{}, err
	}
	chargedID, refundID := usagePointEntryIDs(pointEntries)
	result := orderfoodResponse.AIUsageDetail{
		AIUsageSummary: summary, FailureCategory: row.FailureCategory,
		FailureSummary: row.FailureSummary, ChargedPointEntryID: chargedID,
		RefundPointEntryID: refundID, ClearedAt: row.ClearedAt,
		ClearedBy: clearedByAdministrator(row), Timeline: aiUsageTimeline(row, pointEntries),
		Version: row.Version,
	}
	if includeSensitive && !row.SensitiveContentCleared {
		if err := decodeOptionalJSON(row.OriginalInputJSON, &result.OriginalInput); err != nil {
			return orderfoodResponse.AIUsageDetail{}, err
		}
		if err := decodeOptionalJSON(row.ImageMetadataJSON, &result.ImageMetadata); err != nil {
			return orderfoodResponse.AIUsageDetail{}, err
		}
		if err := decodeOptionalJSON(row.ModelOutputJSON, &result.ModelOutput); err != nil {
			return orderfoodResponse.AIUsageDetail{}, err
		}
		if service.Audit == nil {
			service.Audit = NewAccessAuditService(db)
		}
		if _, err := service.Audit.Record(ctx, SensitiveAccess{
			AdministratorID: actor.AdministratorID, AdministratorUsername: actor.Username,
			AdministratorNickname: actor.Nickname, AuthorityID: actor.AuthorityID,
			Permission: PermissionAIUsageSensitiveRead, Action: "read_ai_usage_sensitive_content",
			TargetType: "ai_usage", TargetID: row.ID, RequestID: actor.RequestID,
			SourceIPMasked: actor.SourceIPMasked, UserAgentSummary: actor.UserAgentSummary,
		}); err != nil {
			return orderfoodResponse.AIUsageDetail{}, err
		}
	}
	return result, nil
}

// ClearSensitiveContent 永久清除AI调用记录中的原始输入和模型输出。
func (service *AIUsageService) ClearSensitiveContent(
	ctx context.Context,
	usageID string,
	input orderfoodRequest.AIUsageSensitiveDeleteInput,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
) (orderfoodResponse.AIUsageSensitiveDeleteResult, bool, error) {
	if service == nil || service.Permission == nil || service.Idempotency == nil {
		return orderfoodResponse.AIUsageSensitiveDeleteResult{}, false, appErrors.AdminInternal.DefaultMsg()
	}
	usageID = strings.TrimSpace(usageID)
	if usageID == "" || len([]rune(usageID)) > 64 || input.ExpectedVersion < 1 {
		return orderfoodResponse.AIUsageSensitiveDeleteResult{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := validateAdminActor(actor); err != nil {
		return orderfoodResponse.AIUsageSensitiveDeleteResult{}, false, err
	}
	if err := validateReason(input.Reason); err != nil {
		return orderfoodResponse.AIUsageSensitiveDeleteResult{}, false, err
	}
	if err := service.Permission.Require(
		ctx,
		actor.AuthorityID,
		PermissionAIUsageSensitiveDelete,
	); err != nil {
		return orderfoodResponse.AIUsageSensitiveDeleteResult{}, false, err
	}
	payload := struct {
		UsageID string                                       `json:"usageId"` // 调用记录ID
		Input   orderfoodRequest.AIUsageSensitiveDeleteInput `json:"input"`   // 清除参数
	}{UsageID: usageID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"ai_usage_sensitive_content_delete",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			var row orderfoodModel.FrontFeatureUsage
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&row, "id = ?", usageID).Error; err != nil {
				return nil, aiUsageLookupError(err)
			}
			if row.Version != input.ExpectedVersion || row.SensitiveContentCleared {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			now := service.now()
			if err := tx.Model(&row).Updates(map[string]interface{}{
				"original_input_json": nil, "image_metadata_json": nil, "model_output_json": nil,
				"sensitive_cleared": true, "cleared_at": now,
				"cleared_by_id": actor.AdministratorID, "cleared_by_username": actor.Username,
				"cleared_by_nickname": actor.Nickname, "version": gorm.Expr("version + 1"),
				"updated_at": now,
			}).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "clear AI usage sensitive content")
			}
			beforeJSON, _ := json.Marshal(map[string]interface{}{
				"sensitiveContentCleared": false, "version": row.Version,
			})
			afterJSON, _ := json.Marshal(map[string]interface{}{
				"sensitiveContentCleared": true, "version": row.Version + 1,
			})
			reason := strings.TrimSpace(input.Reason)
			key := strings.TrimSpace(idempotencyKey)
			audit := orderfoodModel.AdminAuditLog{
				AdministratorID: actor.AdministratorID, AdministratorUsername: actor.Username,
				AdministratorNickname: actor.Nickname, Action: "clear_ai_usage_sensitive_content",
				TargetType: "ai_usage", TargetID: row.ID, Reason: &reason,
				BeforeSummary: datatypes.JSON(beforeJSON), AfterSummary: datatypes.JSON(afterJSON),
				RequestID: actor.RequestID, IdempotencyKey: &key,
			}
			audit.CreatedAt = now
			audit.UpdatedAt = now
			if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
				audit.SourceIPMasked = &value
			}
			if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
				audit.UserAgentSummary = &value
			}
			if err := (gormMutationAuditWriter{}).WriteMutationAudit(ctx, tx, &audit); err != nil {
				return nil, err
			}
			return orderfoodResponse.AIUsageSensitiveDeleteResult{
				Cleared: true, ClearedAt: now,
				RetainedAuditFields: append([]string(nil), aiUsageRetainedAuditFields...),
			}, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.AIUsageSensitiveDeleteResult](raw, replayed, err)
}

// summary 将AI调用记录转换为管理端摘要。
func (service *AIUsageService) summary(
	ctx context.Context,
	db *gorm.DB,
	row orderfoodModel.FrontFeatureUsage,
) (orderfoodResponse.AIUsageSummary, error) {
	var user orderfoodModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", row.UserID).Error; err != nil {
		return orderfoodResponse.AIUsageSummary{}, aiUsageLookupError(err)
	}
	capabilityCode := row.CapabilityCode
	if capabilityCode == "" {
		capabilityCode = row.Feature
	}
	promptMode := row.PromptMode
	if promptMode == "" || promptMode == string(orderfoodModel.AIPromptPreset) {
		promptMode = "default"
	}
	return orderfoodResponse.AIUsageSummary{
		ID: row.ID, User: userReference(user), CapabilityCode: capabilityCode,
		ProviderID: row.ProviderID, ProviderName: row.ProviderName,
		ModelID: row.ModelID, ModelName: row.ModelName,
		RequestID: row.RequestID, IdempotencyKey: row.IdempotencyKey,
		ExecutionStatus: row.ExecutionStatus, BillingStatus: row.BillingStatus,
		PromptMode: promptMode, PromptHash: row.PromptHash, DurationMS: row.DurationMS,
		InputTokens: row.InputTokens, OutputTokens: row.OutputTokens, ImageCount: row.ImageCount,
		PointCost: row.PointCost, EstimatedCostCNY: row.EstimatedCostCNY, Currency: "CNY",
		SensitiveContentCleared: row.SensitiveContentCleared,
		CreatedAt:               row.CreatedAt, FinishedAt: row.FinishedAt,
	}, nil
}

// usagePointEntries 查询AI调用关联的扣积分和退款流水。
func usagePointEntries(
	ctx context.Context,
	db *gorm.DB,
	usageID string,
) ([]orderfoodModel.FrontPointEntry, error) {
	var rows []orderfoodModel.FrontPointEntry
	if err := db.WithContext(ctx).
		Where("related_object_type = ? AND related_object_id = ?", "feature_usage", usageID).
		Order("created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "list AI usage point entries")
	}
	return rows, nil
}

// usagePointEntryIDs 从关联流水中提取扣积分和退款流水ID。
func usagePointEntryIDs(rows []orderfoodModel.FrontPointEntry) (*string, *string) {
	var chargedID *string
	var refundID *string
	for index := range rows {
		value := rows[index].ID
		switch rows[index].Type {
		case orderfoodModel.PointSpent:
			chargedID = &value
		case orderfoodModel.PointRefund:
			refundID = &value
		}
	}
	return chargedID, refundID
}

// aiUsageTimeline 构建AI调用执行、计费、退款和敏感内容清除时间线。
func aiUsageTimeline(
	row orderfoodModel.FrontFeatureUsage,
	pointEntries []orderfoodModel.FrontPointEntry,
) []orderfoodResponse.AIUsageTimelineEvent {
	timeline := []orderfoodResponse.AIUsageTimelineEvent{{
		Type: "created", Label: "创建调用", OccurredAt: row.CreatedAt,
	}}
	for index := range pointEntries {
		switch pointEntries[index].Type {
		case orderfoodModel.PointSpent:
			timeline = append(timeline, orderfoodResponse.AIUsageTimelineEvent{
				Type: "points_charged", Label: "扣除积分", OccurredAt: pointEntries[index].CreatedAt,
			})
		case orderfoodModel.PointRefund:
			timeline = append(timeline, orderfoodResponse.AIUsageTimelineEvent{
				Type: "points_refunded", Label: "退还积分", OccurredAt: pointEntries[index].CreatedAt,
			})
		}
	}
	if row.FinishedAt != nil {
		label := "调用结束"
		eventType := "execution_finished"
		switch row.ExecutionStatus {
		case orderfoodModel.FeatureExecutionSucceeded:
			label = "调用成功"
			eventType = "execution_succeeded"
		case orderfoodModel.FeatureExecutionFailed:
			label = "调用失败"
			eventType = "execution_failed"
		}
		timeline = append(timeline, orderfoodResponse.AIUsageTimelineEvent{
			Type: eventType, Label: label, OccurredAt: *row.FinishedAt,
		})
	}
	if row.BillingStatus == orderfoodModel.FeatureBillingRefundPending {
		timeline = append(timeline, orderfoodResponse.AIUsageTimelineEvent{
			Type: "refund_pending", Label: "积分等待退还", OccurredAt: row.UpdatedAt,
		})
	}
	if row.ClearedAt != nil {
		timeline = append(timeline, orderfoodResponse.AIUsageTimelineEvent{
			Type: "sensitive_content_cleared", Label: "敏感内容已清除", OccurredAt: *row.ClearedAt,
		})
	}
	sort.SliceStable(timeline, func(left, right int) bool {
		return timeline[left].OccurredAt.Before(timeline[right].OccurredAt)
	})
	return timeline
}

// validateAIUsageListQuery 校验调用记录筛选、分页和排序参数。
func validateAIUsageListQuery(query orderfoodRequest.AIUsageListQuery) error {
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len([]rune(strings.TrimSpace(query.UserID))) > 64 ||
		len([]rune(strings.TrimSpace(query.CapabilityCode))) > 64 ||
		len([]rune(strings.TrimSpace(query.ProviderID))) > 64 ||
		len([]rune(strings.TrimSpace(query.ModelID))) > 64 ||
		len([]rune(strings.TrimSpace(query.RequestID))) > 128 ||
		len([]rune(strings.TrimSpace(query.IdempotencyKey))) > 128 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	for _, enum := range []struct {
		value   string              // 待校验枚举值
		allowed map[string]struct{} // 允许值集合
	}{
		{
			strings.TrimSpace(query.ExecutionStatus),
			map[string]struct{}{"": {}, "pending": {}, "processing": {}, "succeeded": {}, "failed": {}},
		},
		{
			strings.TrimSpace(query.BillingStatus),
			map[string]struct{}{"": {}, "not_charged": {}, "charged": {}, "refund_pending": {}, "refunded": {}},
		},
	} {
		if _, exists := enum.allowed[enum.value]; !exists {
			return appErrors.AdminBadRequest.DefaultMsg()
		}
	}
	if _, exists := map[string]struct{}{
		"createdAt": {}, "durationMs": {}, "estimatedCostCny": {},
	}[query.SortBy]; !exists {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.SortOrder != "asc" && query.SortOrder != "desc" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// decodeOptionalJSON 解码可为空的AI调用敏感JSON字段。
func decodeOptionalJSON(raw datatypes.JSON, target interface{}) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return appErrors.AdminInternal.Wrap(err, "decode AI usage sensitive content")
	}
	return nil
}

// clearedByAdministrator 构建敏感内容清除管理员摘要。
func clearedByAdministrator(
	row orderfoodModel.FrontFeatureUsage,
) *orderfoodResponse.AdministratorSummary {
	if row.ClearedByID == nil {
		return nil
	}
	username := ""
	if row.ClearedByUsername != nil {
		username = *row.ClearedByUsername
	}
	return &orderfoodResponse.AdministratorSummary{
		ID:       strconv.FormatUint(uint64(*row.ClearedByID), 10),
		Username: username, Nickname: row.ClearedByNickname,
	}
}

// aiUsageLookupError 统一转换AI调用记录查询错误。
func aiUsageLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.AdminNotFound.DefaultMsg()
	}
	return appErrors.AdminInternal.Wrap(err, "query AI usage record")
}

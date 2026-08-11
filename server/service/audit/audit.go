package audit

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	auditRequest "github.com/dyjh/order-food-mini-app/server/model/audit/request"
	auditResponse "github.com/dyjh/order-food-mini-app/server/model/audit/response"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const PermissionAuditRead = "orderfood:audit:read"

// AuditService 提供管理操作审计业务能力。
type AuditService struct {
	DB *gorm.DB // 业务数据库
}

// NewAuditService 创建审计服务实例。
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{DB: db}
}

func (service *AuditService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// List 分页查询管理操作审计。
func (service *AuditService) List(
	ctx context.Context,
	query auditRequest.AuditLogListQuery,
) (commonResponse.Page[auditResponse.AuditLog], error) {
	query.ApplyDefaults()
	db := service.database()
	if db == nil {
		return commonResponse.Page[auditResponse.AuditLog]{}, appErrors.AdminInternal.DefaultMsg()
	}

	statement := db.WithContext(ctx).Model(&auditModel.AdminAuditLog{})
	if query.AdministratorID != "" {
		administratorID, err := strconv.ParseUint(query.AdministratorID, 10, 64)
		if err != nil || administratorID == 0 {
			return commonResponse.Page[auditResponse.AuditLog]{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		statement = statement.Where("administrator_id = ?", uint(administratorID))
	}
	for _, filter := range []struct {
		column string
		value  string
	}{
		{"target_type", query.TargetType},
		{"target_id", query.TargetID},
		{"action", query.Action},
		{"request_id", query.RequestID},
	} {
		if value := strings.TrimSpace(filter.value); value != "" {
			statement = statement.Where(filter.column+" = ?", value)
		}
	}
	for _, filter := range []struct {
		value string
		op    string
	}{
		{query.CreatedFrom, ">="},
		{query.CreatedTo, "<="},
	} {
		if filter.value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, filter.value)
		if err != nil {
			return commonResponse.Page[auditResponse.AuditLog]{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		statement = statement.Where("created_at "+filter.op+" ?", parsed.UTC())
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[auditResponse.AuditLog]{}, appErrors.AdminInternal.Wrap(err, "count audit logs")
	}
	sortOrder := strings.ToLower(query.SortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		return commonResponse.Page[auditResponse.AuditLog]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var records []auditModel.AdminAuditLog
	if err := statement.
		Order("created_at " + sortOrder).
		Order("id " + sortOrder).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&records).Error; err != nil {
		return commonResponse.Page[auditResponse.AuditLog]{}, appErrors.AdminInternal.Wrap(err, "list audit logs")
	}
	list := make([]auditResponse.AuditLog, 0, len(records))
	for index := range records {
		list = append(list, toAuditLog(records[index]))
	}
	return commonResponse.Page[auditResponse.AuditLog]{
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
		List:     list,
	}, nil
}

// Detail 获取管理操作审计详情。
func (service *AuditService) Detail(
	ctx context.Context,
	auditLogID string,
) (auditResponse.AuditLogDetail, error) {
	db := service.database()
	if db == nil {
		return auditResponse.AuditLogDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	auditLogID = strings.TrimSpace(auditLogID)
	if auditLogID == "" {
		return auditResponse.AuditLogDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var record auditModel.AdminAuditLog
	if err := db.WithContext(ctx).First(&record, "public_id = ?", auditLogID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auditResponse.AuditLogDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return auditResponse.AuditLogDetail{}, appErrors.AdminInternal.Wrap(err, "get audit log")
	}
	beforeSummary, err := decodeAuditSummary(record.BeforeSummary)
	if err != nil {
		return auditResponse.AuditLogDetail{}, err
	}
	afterSummary, err := decodeAuditSummary(record.AfterSummary)
	if err != nil {
		return auditResponse.AuditLogDetail{}, err
	}
	return auditResponse.AuditLogDetail{
		AuditLog:         toAuditLog(record),
		BeforeSummary:    beforeSummary,
		AfterSummary:     afterSummary,
		SourceIPMasked:   record.SourceIPMasked,
		UserAgentSummary: record.UserAgentSummary,
	}, nil
}

func toAuditLog(record auditModel.AdminAuditLog) auditResponse.AuditLog {
	return auditResponse.AuditLog{
		ID: record.PublicID,
		Administrator: commonResponse.AdministratorSummary{
			ID:       strconv.FormatUint(uint64(record.AdministratorID), 10),
			Username: record.AdministratorUsername,
			Nickname: record.AdministratorNickname,
		},
		Action:         record.Action,
		TargetType:     record.TargetType,
		TargetID:       record.TargetID,
		TargetLabel:    record.TargetLabel,
		Reason:         record.Reason,
		RequestID:      record.RequestID,
		IdempotencyKey: record.IdempotencyKey,
		CreatedAt:      record.CreatedAt,
	}
}

func decodeAuditSummary(raw datatypes.JSON) (interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var summary map[string]interface{}
	if err := json.Unmarshal(raw, &summary); err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "decode audit summary")
	}
	return summary, nil
}

package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	// PermissionGovernanceRead 表示读取违规处理记录所需的权限。
	PermissionGovernanceRead = "orderfood:governance:read"
	// PermissionGovernanceRetry 表示重试违规处理任务所需的权限。
	PermissionGovernanceRetry = "orderfood:governance:retry"
	// PermissionGovernanceCascade 表示处理严重违规复制链所需的权限。
	PermissionGovernanceCascade = "orderfood:governance:cascade"
)

// GovernanceService 提供违规处理记录和异步任务业务能力。
type GovernanceService struct {
	DB            *gorm.DB                          // 数据库连接
	Permission    *serviceCommon.PermissionService  // 管理端权限服务
	Idempotency   *serviceCommon.IdempotencyService // 幂等请求服务
	MutationAudit serviceCommon.MutationAuditWriter // 变更审计写入器
	Now           func() time.Time                  // 当前时间函数
}

// NewGovernanceService 创建违规处理记录服务实例。
func NewGovernanceService(
	db *gorm.DB,
	permission *serviceCommon.PermissionService,
	idempotency *serviceCommon.IdempotencyService,
) *GovernanceService {
	if permission == nil {
		permission = serviceCommon.NewPermissionService(db)
	}
	if idempotency == nil {
		idempotency = serviceCommon.NewIdempotencyService(db)
	}
	return &GovernanceService{
		DB:            db,
		Permission:    permission,
		Idempotency:   idempotency,
		MutationAudit: serviceCommon.GormMutationAuditWriter{},
		Now:           time.Now,
	}
}

// database 返回服务使用的数据库连接。
func (service *GovernanceService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// now 返回服务使用的UTC时间。
func (service *GovernanceService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// ListRecords 分页查询违规处理记录。
func (service *GovernanceService) ListRecords(
	ctx context.Context,
	query contentRequest.GovernanceRecordListQuery,
	actor commonRequest.AdminActor,
) (commonResponse.Page[contentResponse.GovernanceRecordSummary], error) {
	query.ApplyDefaults()
	if err := validateGovernanceRecordListQuery(query); err != nil {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, err
	}
	if service == nil || service.Permission == nil {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionGovernanceRead); err != nil {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, err
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&contentModel.GovernanceRecord{})
	for _, filter := range []struct {
		column string // 数据库字段
		value  string // 查询值
	}{
		{"target_type", query.TargetType},
		{"target_id", query.TargetID},
		{"violation_type", query.ViolationType},
	} {
		if value := strings.TrimSpace(filter.value); value != "" {
			statement = statement.Where(filter.column+" = ?", value)
		}
	}
	if query.JobStatus == "abnormal" {
		statement = statement.Where(
			"job_status IN ?",
			[]string{
				contentModel.GovernanceJobProcessing,
				contentModel.GovernanceJobPartiallySucceeded,
				contentModel.GovernanceJobFailed,
			},
		)
	} else if value := strings.TrimSpace(query.JobStatus); value != "" {
		statement = statement.Where("job_status = ?", value)
	}
	if value := strings.TrimSpace(query.Action); value != "" {
		encoded, _ := json.Marshal(value)
		statement = statement.Where("JSON_CONTAINS(actions, ?)", string(encoded))
	}
	if value := strings.TrimSpace(query.AdministratorID); value != "" {
		administratorID, err := strconv.ParseUint(value, 10, 64)
		if err != nil || administratorID == 0 {
			return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		statement = statement.Where("administrator_id = ?", administratorID)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", query.CreatedFrom.UTC())
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", query.CreatedTo.UTC())
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{},
			appErrors.AdminInternal.Wrap(err, "count governance records")
	}
	sortColumn := map[string]string{"createdAt": "created_at", "updatedAt": "updated_at"}[query.SortBy]
	if sortColumn == "" {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var records []contentModel.GovernanceRecord
	if err := statement.
		Order(sortColumn + " " + query.SortOrder).
		Order("id " + query.SortOrder).
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&records).Error; err != nil {
		return commonResponse.Page[contentResponse.GovernanceRecordSummary]{},
			appErrors.AdminInternal.Wrap(err, "list governance records")
	}
	list := make([]contentResponse.GovernanceRecordSummary, 0, len(records))
	recordIDs := make([]string, 0, len(records))
	for index := range records {
		summary, err := governanceRecordSummary(records[index])
		if err != nil {
			return commonResponse.Page[contentResponse.GovernanceRecordSummary]{}, err
		}
		list = append(list, summary)
		recordIDs = append(recordIDs, records[index].PublicID)
	}
	if len(recordIDs) > 0 {
		var jobs []contentModel.GovernanceJob
		if err := db.WithContext(ctx).
			Where("record_id IN ?", recordIDs).
			Find(&jobs).Error; err != nil {
			return commonResponse.Page[contentResponse.GovernanceRecordSummary]{},
				appErrors.AdminInternal.Wrap(err, "list governance job progress")
		}
		progressByRecordID := make(map[string]contentResponse.GovernanceJobProgress, len(jobs))
		for _, job := range jobs {
			progressByRecordID[job.RecordID] = governanceJobProgress(job)
		}
		for index := range list {
			if progress, exists := progressByRecordID[list[index].ID]; exists {
				value := progress
				list[index].JobProgress = &value
			}
		}
	}
	return commonResponse.Page[contentResponse.GovernanceRecordSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// validateGovernanceRecordListQuery 校验服务层直接调用时的分页、枚举和排序边界。
func validateGovernanceRecordListQuery(query contentRequest.GovernanceRecordListQuery) error {
	allowedTargetTypes := map[string]struct{}{
		"": {}, "dish": {}, "official_dish": {}, "recipe": {}, "checkin": {},
	}
	allowedActions := map[string]struct{}{
		"": {}, "disable_discoverability": {}, "soft_delete_dish": {},
		"soft_delete_official_dish": {}, "soft_delete_recipe": {},
		"soft_delete_checkin": {}, "delete_copy_chain": {},
	}
	allowedJobStatuses := map[string]struct{}{
		"": {}, "pending": {}, "processing": {}, "partially_succeeded": {},
		"succeeded": {}, "failed": {}, "abnormal": {},
	}
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		utf8.RuneCountInString(strings.TrimSpace(query.TargetID)) > 64 ||
		utf8.RuneCountInString(strings.TrimSpace(query.ViolationType)) > 80 ||
		utf8.RuneCountInString(strings.TrimSpace(query.AdministratorID)) > 32 ||
		(query.SortBy != "createdAt" && query.SortBy != "updatedAt") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, exists := allowedTargetTypes[strings.TrimSpace(query.TargetType)]; !exists {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, exists := allowedActions[strings.TrimSpace(query.Action)]; !exists {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, exists := allowedJobStatuses[strings.TrimSpace(query.JobStatus)]; !exists {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// GetRecord 获取违规处理记录详情。
func (service *GovernanceService) GetRecord(
	ctx context.Context,
	recordID string,
	actor commonRequest.AdminActor,
) (contentResponse.GovernanceRecordDetail, error) {
	if service == nil || service.Permission == nil {
		return contentResponse.GovernanceRecordDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionGovernanceRead); err != nil {
		return contentResponse.GovernanceRecordDetail{}, err
	}
	db := service.database()
	if db == nil {
		return contentResponse.GovernanceRecordDetail{}, appErrors.AdminInternal.DefaultMsg()
	}
	var record contentModel.GovernanceRecord
	if err := db.WithContext(ctx).First(&record, "public_id = ?", strings.TrimSpace(recordID)).Error; err != nil {
		return contentResponse.GovernanceRecordDetail{}, governanceLookupError(err)
	}
	return service.governanceRecordDetail(ctx, db, record)
}

// GetJob 获取违规处理异步任务进度。
func (service *GovernanceService) GetJob(
	ctx context.Context,
	jobID string,
	actor commonRequest.AdminActor,
) (contentResponse.GovernanceJob, error) {
	if service == nil || service.Permission == nil {
		return contentResponse.GovernanceJob{}, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionGovernanceRead); err != nil {
		return contentResponse.GovernanceJob{}, err
	}
	db := service.database()
	if db == nil {
		return contentResponse.GovernanceJob{}, appErrors.AdminInternal.DefaultMsg()
	}
	var job contentModel.GovernanceJob
	if err := db.WithContext(ctx).First(&job, "id = ?", strings.TrimSpace(jobID)).Error; err != nil {
		return contentResponse.GovernanceJob{}, governanceLookupError(err)
	}
	return governanceJobDetail(ctx, db, job)
}

// ProcessPendingJobs 分批执行待处理的复制链治理任务。
func (service *GovernanceService) ProcessPendingJobs(ctx context.Context, batchSize int) error {
	if service == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	var jobIDs []string
	if err := db.WithContext(ctx).Model(&contentModel.GovernanceJob{}).
		Where("status IN ?", []string{
			contentModel.GovernanceJobPending,
			contentModel.GovernanceJobProcessing,
		}).
		Order("created_at ASC").
		Limit(10).
		Pluck("id", &jobIDs).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "list pending governance jobs")
	}
	var firstError error
	for _, jobID := range jobIDs {
		if err := service.processPendingJob(ctx, db, jobID, batchSize); err != nil &&
			firstError == nil {
			firstError = err
		}
	}
	return firstError
}

// processPendingJob 在事务中处理一个复制链任务批次并刷新任务进度。
func (service *GovernanceService) processPendingJob(
	ctx context.Context,
	db *gorm.DB,
	jobID string,
	batchSize int,
) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job contentModel.GovernanceJob
		if err := tx.
			First(&job, "id = ?", jobID).Error; err != nil {
			return governanceLookupError(err)
		}
		if job.Status != contentModel.GovernanceJobPending &&
			job.Status != contentModel.GovernanceJobProcessing {
			return nil
		}
		now := service.now()
		if job.Status == contentModel.GovernanceJobPending {
			update := tx.Model(&contentModel.GovernanceJob{}).
				Where("id = ? AND version = ?", job.ID, job.Version).
				Updates(map[string]interface{}{
					"status":     contentModel.GovernanceJobProcessing,
					"started_at": now,
					"version":    gorm.Expr("version + 1"),
					"updated_at": now,
				})
			if update.Error != nil {
				return appErrors.AdminInternal.Wrap(update.Error, "start governance job")
			}
			if update.RowsAffected != 1 {
				return appErrors.AdminStateConflict.DefaultMsg()
			}
			job.Status = contentModel.GovernanceJobProcessing
			job.StartedAt = &now
			job.Version++
		}
		var record contentModel.GovernanceRecord
		if err := tx.First(&record, "public_id = ?", job.RecordID).Error; err != nil {
			return governanceLookupError(err)
		}
		var items []contentModel.GovernanceJobItem
		if err := tx.
			Where("job_id = ? AND status = ?", job.ID, contentModel.GovernanceJobItemPending).
			Order("id ASC").
			Limit(batchSize).
			Find(&items).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "list pending governance job items")
		}
		for index := range items {
			savepoint := fmt.Sprintf("gov_process_%d", index)
			if err := tx.SavePoint(savepoint).Error; err != nil {
				return appErrors.AdminInternal.Wrap(err, "create governance job item savepoint")
			}
			if err := service.processJobItem(ctx, tx, &items[index], record, now); err != nil {
				if rollbackErr := tx.RollbackTo(savepoint).Error; rollbackErr != nil {
					return appErrors.AdminInternal.Wrap(rollbackErr, "rollback governance job item")
				}
				if updateErr := tx.Model(&contentModel.GovernanceJobItem{}).
					Where("id = ? AND status = ?", items[index].ID, contentModel.GovernanceJobItemPending).
					Updates(map[string]interface{}{
						"status":             contentModel.GovernanceJobItemFailed,
						"attempt_count":      gorm.Expr("attempt_count + 1"),
						"last_error_summary": "任务项处理失败",
						"processed_at":       now,
						"updated_at":         now,
					}).Error; updateErr != nil {
					return appErrors.AdminInternal.Wrap(updateErr, "mark governance job item failed")
				}
			}
		}
		if err := refreshGovernanceJobProgress(ctx, tx, &job, now, job.Version); err != nil {
			return err
		}
		return syncGovernanceRecordProgress(ctx, tx, record.PublicID, job)
	})
}

// RetryJob 重试违规处理任务中的失败项。
func (service *GovernanceService) RetryJob(
	ctx context.Context,
	jobID string,
	input contentRequest.GovernanceJobRetryInput,
	actor commonRequest.AdminActor,
	idempotencyKey string,
) (contentResponse.GovernanceJob, bool, error) {
	var result contentResponse.GovernanceJob
	if service == nil || service.Permission == nil || service.Idempotency == nil {
		return result, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := serviceCommon.ValidateAdminActor(actor); err != nil {
		return result, false, err
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionGovernanceRetry); err != nil {
		return result, false, err
	}
	if err := service.Permission.Require(ctx, actor.AuthorityID, PermissionGovernanceCascade); err != nil {
		return result, false, err
	}
	jobID = strings.TrimSpace(jobID)
	input.Reason = strings.TrimSpace(input.Reason)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if jobID == "" || input.ExpectedVersion < 1 ||
		utf8.RuneCountInString(input.Reason) < 4 ||
		utf8.RuneCountInString(input.Reason) > 200 {
		return result, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	payload := struct {
		JobID string                                 `json:"jobId"` // 任务ID
		Input contentRequest.GovernanceJobRetryInput `json:"input"` // 重试参数
	}{JobID: jobID, Input: input}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"governance_job_retry",
		idempotencyKey,
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			return service.retryJobOnce(ctx, tx, jobID, input, actor, idempotencyKey)
		},
	)
	if err != nil {
		return result, false, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode governance job retry result")
	}
	return result, replayed, nil
}

// retryJobOnce 在幂等事务中重试任务失败项并刷新任务进度。
func (service *GovernanceService) retryJobOnce(
	ctx context.Context,
	tx *gorm.DB,
	jobID string,
	input contentRequest.GovernanceJobRetryInput,
	actor commonRequest.AdminActor,
	idempotencyKey string,
) (contentResponse.GovernanceJob, error) {
	var job contentModel.GovernanceJob
	if err := tx.WithContext(ctx).
		First(&job, "id = ?", jobID).Error; err != nil {
		return contentResponse.GovernanceJob{}, governanceLookupError(err)
	}
	if job.Version != input.ExpectedVersion || job.FailedCount == 0 {
		return contentResponse.GovernanceJob{}, appErrors.AdminStateConflict.DefaultMsg()
	}
	var record contentModel.GovernanceRecord
	if err := tx.WithContext(ctx).First(&record, "public_id = ?", job.RecordID).Error; err != nil {
		return contentResponse.GovernanceJob{}, governanceLookupError(err)
	}
	var items []contentModel.GovernanceJobItem
	if err := tx.WithContext(ctx).
		Where("job_id = ? AND status = ?", job.ID, contentModel.GovernanceJobItemFailed).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return contentResponse.GovernanceJob{}, appErrors.AdminInternal.Wrap(err, "list failed governance job items")
	}
	if len(items) == 0 {
		return contentResponse.GovernanceJob{}, appErrors.AdminStateConflict.DefaultMsg()
	}
	now := service.now()
	for index := range items {
		savepoint := fmt.Sprintf("gov_retry_%d", index)
		if err := tx.SavePoint(savepoint).Error; err != nil {
			return contentResponse.GovernanceJob{}, appErrors.AdminInternal.Wrap(err, "create governance retry savepoint")
		}
		if err := service.processJobItem(ctx, tx, &items[index], record, now); err != nil {
			if rollbackErr := tx.RollbackTo(savepoint).Error; rollbackErr != nil {
				return contentResponse.GovernanceJob{}, appErrors.AdminInternal.Wrap(rollbackErr, "rollback governance retry item")
			}
			errorSummary := "任务项处理失败"
			if updateErr := tx.Model(&contentModel.GovernanceJobItem{}).
				Where("id = ?", items[index].ID).
				Updates(map[string]interface{}{
					"status":             contentModel.GovernanceJobItemFailed,
					"attempt_count":      gorm.Expr("attempt_count + 1"),
					"last_error_summary": errorSummary,
					"processed_at":       now,
					"updated_at":         now,
				}).Error; updateErr != nil {
				return contentResponse.GovernanceJob{}, appErrors.AdminInternal.Wrap(updateErr, "mark governance retry item failed")
			}
		}
	}
	before := governanceJobResponse(job)
	if err := refreshGovernanceJobProgress(ctx, tx, &job, now, input.ExpectedVersion); err != nil {
		return contentResponse.GovernanceJob{}, err
	}
	if err := syncGovernanceRecordProgress(ctx, tx, record.PublicID, job); err != nil {
		return contentResponse.GovernanceJob{}, err
	}
	afterSummary := governanceJobResponse(job)
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(afterSummary)
	targetLabel := "违规处理任务"
	reason := input.Reason
	idempotencyKeyCopy := idempotencyKey
	audit := auditModel.AdminAuditLog{
		AdministratorID: actor.AdministratorID, AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname, Action: "retry_governance_job",
		TargetType: "governance_job", TargetID: job.ID, TargetLabel: &targetLabel,
		Reason: &reason, BeforeSummary: datatypes.JSON(beforeJSON), AfterSummary: datatypes.JSON(afterJSON),
		RequestID: actor.RequestID, IdempotencyKey: &idempotencyKeyCopy,
	}
	if service.MutationAudit == nil {
		service.MutationAudit = serviceCommon.GormMutationAuditWriter{}
	}
	if err := service.MutationAudit.WriteMutationAudit(ctx, tx, &audit); err != nil {
		return contentResponse.GovernanceJob{}, err
	}
	return governanceJobDetail(ctx, tx, job)
}

// processJobItem 处理一个复制链菜品任务项，成功项不会再次进入此方法。
func (service *GovernanceService) processJobItem(
	ctx context.Context,
	tx *gorm.DB,
	item *contentModel.GovernanceJobItem,
	record contentModel.GovernanceRecord,
	now time.Time,
) error {
	var dish contentModel.UserDish
	if err := tx.WithContext(ctx).Unscoped().First(&dish, item.DishID).Error; err != nil {
		return governanceLookupError(err)
	}
	notificationID := item.NotificationID
	if !dish.DeletedAt.Valid {
		if err := removeGovernedDishLiveReferences(ctx, tx, dish.ID); err != nil {
			return err
		}
		offlineReason := "严重违规复制链处理"
		if err := tx.WithContext(ctx).Model(&dishModel.PlatformRecommendation{}).
			Where("status = ?", serviceCommon.RecommendationStatusPublished).
			Where(
				"dish_id = ? OR (source_type = ? AND source_dish_id = ?)",
				dish.ID,
				"creator",
				dish.PublicID,
			).
			Updates(map[string]interface{}{
				"status": serviceCommon.RecommendationStatusOffline, "selected": false,
				"offline_at": now, "offline_reason": offlineReason, "version": gorm.Expr("version + 1"),
				"updated_by_id":       record.AdministratorID,
				"updated_by_username": record.AdministratorUsername,
				"updated_by_nickname": record.AdministratorNickname,
				"updated_at":          now,
			}).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "offline governed copy recommendations")
		}
		if err := tx.WithContext(ctx).Model(&dish).Updates(map[string]interface{}{
			"discoverable": false, "discoverable_at": nil,
			"deleted_reason": record.Reason, "deleted_violation_type": record.ViolationType,
			"deleted_by_admin_id": record.AdministratorID,
			"deleted_by_username": record.AdministratorUsername,
			"deleted_by_nickname": record.AdministratorNickname,
			"version":             gorm.Expr("version + 1"),
		}).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "mark governed copy deleted")
		}
		if err := tx.WithContext(ctx).Delete(&dish).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "soft delete governed copy")
		}
	}
	if notificationID == nil {
		targetType := contentModel.GovernanceTargetDish
		targetID := dish.PublicID
		value := serviceCommon.NewPublicID("notice")
		notification := engagementModel.UserNotification{
			ID: value, UserID: dish.OwnerID, Type: "governance",
			Title:      "菜品处理通知",
			Content:    fmt.Sprintf("“%s”已被平台处理。原因：%s", dish.Name, record.Reason),
			TargetType: &targetType, TargetID: &targetID,
			SubscribeRequired: false, CreatedAt: now,
		}
		if err := tx.WithContext(ctx).Create(&notification).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "create governed copy notification")
		}
		notificationID = &value
	}
	return tx.WithContext(ctx).Model(&contentModel.GovernanceJobItem{}).
		Where("id = ? AND status IN ?", item.ID, []string{
			contentModel.GovernanceJobItemPending,
			contentModel.GovernanceJobItemFailed,
		}).
		Updates(map[string]interface{}{
			"status": contentModel.GovernanceJobItemSucceeded, "attempt_count": gorm.Expr("attempt_count + 1"),
			"last_error_summary": nil, "notification_id": notificationID,
			"processed_at": now, "updated_at": now,
		}).Error
}

// syncGovernanceRecordProgress 同步异步任务状态、成功数量和通知ID到治理记录。
func syncGovernanceRecordProgress(
	ctx context.Context,
	tx *gorm.DB,
	recordID string,
	job contentModel.GovernanceJob,
) error {
	var notificationIDs []string
	if err := tx.WithContext(ctx).Model(&contentModel.GovernanceJobItem{}).
		Where("job_id = ? AND status = ? AND notification_id IS NOT NULL", job.ID, contentModel.GovernanceJobItemSucceeded).
		Order("id ASC").
		Pluck("notification_id", &notificationIDs).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "list governance job notifications")
	}
	notificationJSON, err := json.Marshal(notificationIDs)
	if err != nil {
		return appErrors.AdminInternal.Wrap(err, "encode governance job notifications")
	}
	afterJSON, err := json.Marshal(governanceJobResponse(job))
	if err != nil {
		return appErrors.AdminInternal.Wrap(err, "encode governance job progress")
	}
	var record contentModel.GovernanceRecord
	if err := tx.WithContext(ctx).Where("public_id = ?", recordID).First(&record).Error; err != nil {
		return governanceLookupError(err)
	}
	affectedCount := job.SucceededCount
	if record.TargetType == contentModel.GovernanceTargetOfficialDish {
		// 官方源菜品在任务创建时已同步处理，任务进度只统计其用户复制菜品。
		affectedCount++
	}
	if err := tx.WithContext(ctx).Model(&record).
		Where("public_id = ?", recordID).
		Updates(map[string]interface{}{
			"job_status":        job.Status,
			"affected_count":    affectedCount,
			"notification_i_ds": notificationJSON,
			"after_summary":     afterJSON,
		}).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "sync governance record progress")
	}
	return nil
}

// refreshGovernanceJobProgress 根据任务项状态刷新任务汇总。
func refreshGovernanceJobProgress(
	ctx context.Context,
	tx *gorm.DB,
	job *contentModel.GovernanceJob,
	now time.Time,
	expectedVersion int64,
) error {
	type statusCount struct {
		Status string // 任务项状态
		Count  int    // 数量
	}
	var counts []statusCount
	if err := tx.WithContext(ctx).Model(&contentModel.GovernanceJobItem{}).
		Select("status, COUNT(*) AS count").
		Where("job_id = ?", job.ID).
		Group("status").
		Scan(&counts).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "count governance job items")
	}
	job.SucceededCount, job.FailedCount, job.PendingCount = 0, 0, 0
	for _, count := range counts {
		switch count.Status {
		case contentModel.GovernanceJobItemSucceeded:
			job.SucceededCount = count.Count
		case contentModel.GovernanceJobItemFailed:
			job.FailedCount = count.Count
		default:
			job.PendingCount += count.Count
		}
	}
	job.TotalCount = job.SucceededCount + job.FailedCount + job.PendingCount
	job.FinishedAt = nil
	switch {
	case job.PendingCount > 0:
		job.Status = contentModel.GovernanceJobProcessing
	case job.FailedCount == 0:
		job.Status = contentModel.GovernanceJobSucceeded
		job.FinishedAt = &now
	case job.SucceededCount == 0:
		job.Status = contentModel.GovernanceJobFailed
		job.FinishedAt = &now
	default:
		job.Status = contentModel.GovernanceJobPartiallySucceeded
		job.FinishedAt = &now
	}
	job.LastErrorSummary = nil
	if job.FailedCount > 0 {
		value := "存在处理失败的任务项"
		job.LastErrorSummary = &value
	}
	update := tx.WithContext(ctx).Model(&contentModel.GovernanceJob{}).
		Where("id = ? AND version = ?", job.ID, expectedVersion).
		Updates(map[string]interface{}{
			"status": job.Status, "total_count": job.TotalCount,
			"succeeded_count": job.SucceededCount, "failed_count": job.FailedCount,
			"pending_count": job.PendingCount, "last_error_summary": job.LastErrorSummary,
			"finished_at": job.FinishedAt, "version": gorm.Expr("version + 1"), "updated_at": now,
		})
	if update.Error != nil {
		return appErrors.AdminInternal.Wrap(update.Error, "update governance job progress")
	}
	if update.RowsAffected != 1 {
		return appErrors.AdminStateConflict.DefaultMsg()
	}
	job.Version = expectedVersion + 1
	job.UpdatedAt = now
	return nil
}

// governanceRecordSummary 将持久化记录转换为列表摘要。
func governanceRecordSummary(
	record contentModel.GovernanceRecord,
) (contentResponse.GovernanceRecordSummary, error) {
	var actions []string
	if err := json.Unmarshal(record.Actions, &actions); err != nil {
		return contentResponse.GovernanceRecordSummary{},
			appErrors.AdminInternal.Wrap(err, "decode governance actions")
	}
	return contentResponse.GovernanceRecordSummary{
		ID: record.PublicID, TargetType: record.TargetType, TargetID: record.TargetID,
		TargetLabel: record.TargetLabel, Actions: actions, ViolationType: record.ViolationType,
		Severity: record.Severity, Reason: record.Reason,
		Administrator: commonResponse.AdministratorSummary{
			ID:       strconv.FormatUint(uint64(record.AdministratorID), 10),
			Username: record.AdministratorUsername, Nickname: record.AdministratorNickname,
		},
		AffectedCount: record.AffectedCount, JobStatus: record.JobStatus, CreatedAt: record.CreatedAt,
	}, nil
}

// governanceRecordDetail 将持久化记录转换为详情响应。
func (service *GovernanceService) governanceRecordDetail(
	ctx context.Context,
	db *gorm.DB,
	record contentModel.GovernanceRecord,
) (contentResponse.GovernanceRecordDetail, error) {
	summary, err := governanceRecordSummary(record)
	if err != nil {
		return contentResponse.GovernanceRecordDetail{}, err
	}
	var impact contentResponse.GovernanceImpactPreview
	if err := json.Unmarshal(record.ImpactSnapshot, &impact); err != nil {
		return contentResponse.GovernanceRecordDetail{},
			appErrors.AdminInternal.Wrap(err, "decode governance impact snapshot")
	}
	before, err := decodeGovernanceSummary(record.BeforeSummary)
	if err != nil {
		return contentResponse.GovernanceRecordDetail{}, err
	}
	after, err := decodeGovernanceSummary(record.AfterSummary)
	if err != nil {
		return contentResponse.GovernanceRecordDetail{}, err
	}
	notificationIDs := []string{}
	if len(record.NotificationIDs) > 0 && string(record.NotificationIDs) != "null" {
		if err := json.Unmarshal(record.NotificationIDs, &notificationIDs); err != nil {
			return contentResponse.GovernanceRecordDetail{},
				appErrors.AdminInternal.Wrap(err, "decode governance notification ids")
		}
	}
	var jobResponse *contentResponse.GovernanceJob
	var job contentModel.GovernanceJob
	if err := db.WithContext(ctx).First(&job, "record_id = ?", record.PublicID).Error; err == nil {
		value, detailErr := governanceJobDetail(ctx, db, job)
		if detailErr != nil {
			return contentResponse.GovernanceRecordDetail{}, detailErr
		}
		jobResponse = &value
		progress := governanceJobProgress(job)
		summary.JobProgress = &progress
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return contentResponse.GovernanceRecordDetail{},
			appErrors.AdminInternal.Wrap(err, "load governance job")
	}
	targetExists, targetVersion, err := governanceTargetState(
		ctx, db, record.TargetType, record.TargetID,
	)
	if err != nil {
		return contentResponse.GovernanceRecordDetail{}, err
	}
	return contentResponse.GovernanceRecordDetail{
		GovernanceRecordSummary: summary,
		ImpactSnapshot:          impact,
		BeforeSummary:           before,
		AfterSummary:            after,
		Job:                     jobResponse,
		NotificationIDs:         notificationIDs,
		RequestID:               record.RequestID,
		IdempotencyKey:          record.IdempotencyKey,
		TargetExists:            targetExists,
		TargetVersion:           targetVersion,
	}, nil
}

// governanceTargetState 返回违规处理目标是否仍存在及其当前版本。
func governanceTargetState(
	ctx context.Context,
	db *gorm.DB,
	targetType string,
	targetID string,
) (bool, int, error) {
	var state struct {
		Version int `gorm:"column:version"` // 当前目标版本
	}
	var result *gorm.DB
	switch targetType {
	case contentModel.GovernanceTargetDish:
		result = db.WithContext(ctx).Model(&contentModel.UserDish{}).
			Select("version").Where("public_id = ?", targetID).Take(&state)
	case contentModel.GovernanceTargetOfficialDish:
		result = db.WithContext(ctx).Model(&dishModel.OfficialDish{}).
			Select("version").Where("public_id = ?", targetID).Take(&state)
	case contentModel.GovernanceTargetRecipe:
		result = db.WithContext(ctx).Model(&contentModel.UserRecipe{}).
			Select("version").Where("public_id = ?", targetID).Take(&state)
	case contentModel.GovernanceTargetCheckin:
		result = db.WithContext(ctx).Model(&engagementModel.FrontCheckin{}).
			Select("version").Where("id = ?", targetID).Take(&state)
	default:
		return false, 0, nil
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, 0, nil
	}
	if result.Error != nil {
		return false, 0, appErrors.AdminInternal.Wrap(result.Error, "check governance target state")
	}
	return true, state.Version, nil
}

// governanceJobResponse 将持久化任务转换为响应数据。
func governanceJobResponse(job contentModel.GovernanceJob) contentResponse.GovernanceJob {
	return contentResponse.GovernanceJob{
		ID: job.ID, RecordID: job.RecordID, Status: job.Status,
		TotalCount: job.TotalCount, SucceededCount: job.SucceededCount,
		FailedCount: job.FailedCount, PendingCount: job.PendingCount,
		LastErrorSummary: job.LastErrorSummary, Version: job.Version,
		StartedAt: job.StartedAt, FinishedAt: job.FinishedAt, UpdatedAt: job.UpdatedAt,
	}
}

// governanceJobProgress 将持久化任务转换为列表使用的数量进度。
func governanceJobProgress(job contentModel.GovernanceJob) contentResponse.GovernanceJobProgress {
	return contentResponse.GovernanceJobProgress{
		TotalCount: job.TotalCount, SucceededCount: job.SucceededCount,
		FailedCount: job.FailedCount, PendingCount: job.PendingCount,
	}
}

// governanceJobDetail 查询任务项并组装完整的异步任务详情。
func governanceJobDetail(
	ctx context.Context,
	db *gorm.DB,
	job contentModel.GovernanceJob,
) (contentResponse.GovernanceJob, error) {
	var rows []contentModel.GovernanceJobItem
	if err := db.WithContext(ctx).
		Where("job_id = ?", job.ID).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return contentResponse.GovernanceJob{},
			appErrors.AdminInternal.Wrap(err, "list governance job items")
	}
	result := governanceJobResponse(job)
	result.Items = make([]contentResponse.GovernanceJobItem, 0, len(rows))
	for _, row := range rows {
		result.Items = append(result.Items, contentResponse.GovernanceJobItem{
			ID: row.ID, DishID: row.DishPublicID, UserID: row.UserID,
			Status: row.Status, AttemptCount: row.AttemptCount,
			LastErrorSummary: row.LastErrorSummary, NotificationID: row.NotificationID,
			ProcessedAt: row.ProcessedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return result, nil
}

// decodeGovernanceSummary 解码违规处理前后摘要。
func decodeGovernanceSummary(raw datatypes.JSON) (interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var summary map[string]interface{}
	if err := json.Unmarshal(raw, &summary); err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "decode governance summary")
	}
	return summary, nil
}

// governanceLookupError 统一转换违规处理资源查询错误。
func governanceLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.AdminNotFound.DefaultMsg()
	}
	return appErrors.AdminInternal.Wrap(err, "query governance resource")
}

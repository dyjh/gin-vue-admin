package content

import (
	"context"
	"encoding/json"
	"errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"path/filepath"
	"strconv"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	engagementModel "github.com/dyjh/order-food-mini-app/server/model/engagement"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"gorm.io/gorm"
)

const (
	// PermissionMediaRead 表示读取图片资源所需的权限。
	PermissionMediaRead = "orderfood:media:read"
)

// MediaService 提供图片资源和图片审核记录查询能力。
type MediaService struct {
	DB *gorm.DB // 数据库连接
}

// NewMediaService 创建图片资源查询服务实例。
func NewMediaService(db *gorm.DB) *MediaService {
	return &MediaService{DB: db}
}

// database 返回服务使用的数据库连接。
func (service *MediaService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// mediaBinding 表示图片当前绑定的业务对象。
type mediaBinding struct {
	ObjectType *string // 对象类型
	ObjectID   *string // 对象ID
	Version    *int    // 对象数据版本
	Deleted    bool    // 对象是否已软删除
}

// loadMediaBinding 查询图片当前绑定的第一个业务对象。
func loadMediaBinding(
	ctx context.Context,
	db *gorm.DB,
	fileID string,
) (mediaBinding, error) {
	var official dishModel.OfficialDish
	if err := db.WithContext(ctx).Unscoped().
		Where("cover_file_id = ?", fileID).
		Order("id ASC").
		First(&official).Error; err == nil {
		objectType := "official_dish"
		version := int(official.Version)
		return mediaBinding{
			ObjectType: &objectType, ObjectID: &official.PublicID,
			Version: &version, Deleted: official.DeletedAt.Valid,
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return mediaBinding{}, appErrors.AdminInternal.Wrap(err, "load official dish media binding")
	}
	var dish contentModel.UserDish
	if err := db.WithContext(ctx).Unscoped().
		Where("cover_file_id = ?", fileID).
		Order("id ASC").
		First(&dish).Error; err == nil {
		objectType := "user_dish"
		version := dish.Version
		return mediaBinding{
			ObjectType: &objectType, ObjectID: &dish.PublicID,
			Version: &version, Deleted: dish.DeletedAt.Valid,
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return mediaBinding{}, appErrors.AdminInternal.Wrap(err, "load user dish media binding")
	}
	var checkin engagementModel.FrontCheckin
	if err := db.WithContext(ctx).Unscoped().
		Where("image_file_id = ?", fileID).
		Order("created_at ASC").
		First(&checkin).Error; err == nil {
		objectType := "checkin"
		version := checkin.Version
		return mediaBinding{
			ObjectType: &objectType, ObjectID: &checkin.ID,
			Version: &version, Deleted: checkin.DeletedAt.Valid,
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return mediaBinding{}, appErrors.AdminInternal.Wrap(err, "load checkin media binding")
	}
	return mediaBinding{}, nil
}

// mediaObjectKeyMasked 返回不暴露存储目录的对象键。
func mediaObjectKeyMasked(storagePath string) string {
	key := filepath.Base(strings.TrimSpace(storagePath))
	if key == "." || key == "" {
		return "***"
	}
	if len(key) <= 12 {
		return "***" + key
	}
	return "***" + key[len(key)-12:]
}

// mediaUploader 查询并转换图片上传者摘要。
func mediaUploader(
	ctx context.Context,
	db *gorm.DB,
	asset contentModel.FrontMediaAsset,
) (interface{}, error) {
	if asset.AdministratorID != nil {
		username := ""
		if asset.AdministratorUsername != nil {
			username = *asset.AdministratorUsername
		}
		return serviceCommon.AdministratorSummary(
			*asset.AdministratorID,
			username,
			asset.AdministratorNickname,
		), nil
	}
	if strings.TrimSpace(asset.UserID) == "" {
		return nil, nil
	}
	var user userModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", asset.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "load media uploader")
	}
	return serviceCommon.UserReference(user), nil
}

// mediaSummary 将媒体文件转换为管理端摘要。
func mediaSummary(
	ctx context.Context,
	db *gorm.DB,
	asset contentModel.FrontMediaAsset,
) (contentResponse.MediaSummary, error) {
	binding, err := loadMediaBinding(ctx, db, asset.ID)
	if err != nil {
		return contentResponse.MediaSummary{}, err
	}
	uploader, err := mediaUploader(ctx, db, asset)
	if err != nil {
		return contentResponse.MediaSummary{}, err
	}
	var url *string
	if asset.ResourceStatus == serviceCommon.MediaResourceActive && asset.DeletedAt == nil &&
		strings.TrimSpace(asset.URL) != "" {
		value := asset.URL
		url = &value
	}
	status := asset.ResourceStatus
	if status == "" {
		status = serviceCommon.MediaResourceActive
	}
	uploadSource := asset.UploadSource
	if uploadSource == "" {
		uploadSource = "user_upload"
	}
	var width *int
	if asset.Width > 0 {
		value := asset.Width
		width = &value
	}
	var height *int
	if asset.Height > 0 {
		value := asset.Height
		height = &value
	}
	return contentResponse.MediaSummary{
		ID:                 asset.ID,
		ObjectKeyMasked:    mediaObjectKeyMasked(asset.StoragePath),
		URL:                url,
		SourceScene:        asset.Scene,
		UploadSource:       uploadSource,
		Uploader:           uploader,
		BoundObjectType:    binding.ObjectType,
		BoundObjectID:      binding.ObjectID,
		BoundObjectVersion: binding.Version,
		BoundObjectDeleted: binding.Deleted,
		ReviewStatus:       asset.ReviewStatus,
		ResourceStatus:     status,
		Width:              width,
		Height:             height,
		FileSize:           asset.SizeBytes,
		MimeType:           asset.ContentType,
		CreatedAt:          asset.CreatedAt,
	}, nil
}

// ListMedia 分页查询图片资源。
func (service *MediaService) ListMedia(
	ctx context.Context,
	query contentRequest.MediaListQuery,
) (commonResponse.Page[contentResponse.MediaSummary], error) {
	query.ApplyDefaults()
	if err := validateMediaListQuery(query); err != nil {
		return commonResponse.Page[contentResponse.MediaSummary]{}, err
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[contentResponse.MediaSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&contentModel.FrontMediaAsset{})
	if value := strings.TrimSpace(query.FileID); value != "" {
		statement = statement.Where("id = ?", value)
	}
	if value := strings.TrimSpace(query.UploaderUserID); value != "" {
		statement = statement.Where("user_id = ?", value)
	}
	if value := strings.TrimSpace(query.UploaderAdministratorID); value != "" {
		administratorID, err := strconv.ParseUint(value, 10, 64)
		if err != nil || administratorID == 0 {
			return commonResponse.Page[contentResponse.MediaSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
		}
		statement = statement.Where("administrator_id = ?", administratorID)
	}
	if query.UploadSource != "" {
		statement = statement.Where("upload_source = ?", query.UploadSource)
	}
	if query.SourceScene != "" {
		statement = statement.Where("scene = ?", query.SourceScene)
	}
	if query.ReviewStatus != "" {
		statement = statement.Where("review_status = ?", query.ReviewStatus)
	}
	if query.ResourceStatus != "" {
		statement = statement.Where("resource_status = ?", query.ResourceStatus)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", *query.CreatedTo)
	}
	if value := strings.TrimSpace(query.BoundObjectType); value != "" {
		statement = filterMediaByBoundObject(statement, db, value, strings.TrimSpace(query.BoundObjectID))
	} else if value := strings.TrimSpace(query.BoundObjectID); value != "" {
		statement = filterMediaByAnyBoundObject(statement, db, value)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[contentResponse.MediaSummary]{}, appErrors.AdminInternal.Wrap(err, "count media")
	}
	sortColumns := map[string]string{"createdAt": "created_at", "fileSize": "size_bytes"}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[contentResponse.MediaSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	rows := make([]contentModel.FrontMediaAsset, 0)
	if err := statement.Order(column + " " + query.SortOrder).
		Order("id ASC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return commonResponse.Page[contentResponse.MediaSummary]{}, appErrors.AdminInternal.Wrap(err, "list media")
	}
	list := make([]contentResponse.MediaSummary, 0, len(rows))
	for _, asset := range rows {
		item, err := mediaSummary(ctx, db, asset)
		if err != nil {
			return commonResponse.Page[contentResponse.MediaSummary]{}, err
		}
		list = append(list, item)
	}
	return commonResponse.Page[contentResponse.MediaSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// validateMediaListQuery 校验服务层直接调用时的图片列表查询边界。
func validateMediaListQuery(query contentRequest.MediaListQuery) error {
	allowedUploadSources := map[string]struct{}{
		"": {}, "user_upload": {}, "admin_upload": {}, "ai_generated": {},
	}
	allowedScenes := map[string]struct{}{
		"": {}, "profile_avatar": {}, "dish_cover": {}, "dish_step": {},
		"dish_extract": {}, "checkin": {}, "generated_cover": {},
		"official_dish_cover": {},
	}
	allowedBindings := map[string]struct{}{
		"": {}, "user_dish": {}, "official_dish": {}, "checkin": {},
	}
	allowedReviewStatuses := map[string]struct{}{
		"": {}, "pending": {}, "passed": {}, "rejected": {}, "failed": {},
		"not_required": {},
	}
	allowedResourceStatuses := map[string]struct{}{
		"": {}, "active": {}, "missing": {}, "deleted": {},
	}
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len(strings.TrimSpace(query.FileID)) > 128 ||
		len(strings.TrimSpace(query.UploaderUserID)) > 64 ||
		len(strings.TrimSpace(query.UploaderAdministratorID)) > 32 ||
		len(strings.TrimSpace(query.BoundObjectID)) > 80 ||
		(query.SortBy != "createdAt" && query.SortBy != "fileSize") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	for _, enum := range []struct {
		value   string              // 待校验枚举值
		allowed map[string]struct{} // 允许值集合
	}{
		{strings.TrimSpace(query.UploadSource), allowedUploadSources},
		{strings.TrimSpace(query.SourceScene), allowedScenes},
		{strings.TrimSpace(query.BoundObjectType), allowedBindings},
		{strings.TrimSpace(query.ReviewStatus), allowedReviewStatuses},
		{strings.TrimSpace(query.ResourceStatus), allowedResourceStatuses},
	} {
		if _, exists := enum.allowed[enum.value]; !exists {
			return appErrors.AdminBadRequest.DefaultMsg()
		}
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// filterMediaByBoundObject 按指定类型和可选对象ID筛选绑定图片。
func filterMediaByBoundObject(
	statement *gorm.DB,
	db *gorm.DB,
	objectType string,
	objectID string,
) *gorm.DB {
	switch objectType {
	case "official_dish":
		subquery := db.Unscoped().Model(&dishModel.OfficialDish{}).Select("cover_file_id")
		if objectID != "" {
			subquery = subquery.Where("public_id = ?", objectID)
		}
		return statement.Where("id IN (?)", subquery)
	case "user_dish":
		subquery := db.Unscoped().Model(&contentModel.UserDish{}).Select("cover_file_id")
		if objectID != "" {
			subquery = subquery.Where("public_id = ?", objectID)
		}
		return statement.Where("id IN (?)", subquery)
	case "checkin":
		subquery := db.Unscoped().Model(&engagementModel.FrontCheckin{}).Select("image_file_id")
		if objectID != "" {
			subquery = subquery.Where("id = ?", objectID)
		}
		return statement.Where("id IN (?)", subquery)
	default:
		return statement.Where("1 = 0")
	}
}

// filterMediaByAnyBoundObject 按对象ID在所有支持的绑定类型中筛选图片。
func filterMediaByAnyBoundObject(
	statement *gorm.DB,
	db *gorm.DB,
	objectID string,
) *gorm.DB {
	return statement.Where(
		"id IN (?) OR id IN (?) OR id IN (?)",
		db.Unscoped().Model(&dishModel.OfficialDish{}).
			Select("cover_file_id").Where("public_id = ?", objectID),
		db.Unscoped().Model(&contentModel.UserDish{}).
			Select("cover_file_id").Where("public_id = ?", objectID),
		db.Unscoped().Model(&engagementModel.FrontCheckin{}).
			Select("image_file_id").Where("id = ?", objectID),
	)
}

// GetMedia 获取图片资源详情。
func (service *MediaService) GetMedia(
	ctx context.Context,
	fileID string,
) (contentResponse.MediaDetail, error) {
	db := service.database()
	fileID = strings.TrimSpace(fileID)
	if db == nil || fileID == "" {
		return contentResponse.MediaDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var asset contentModel.FrontMediaAsset
	if err := db.WithContext(ctx).First(&asset, "id = ?", fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contentResponse.MediaDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return contentResponse.MediaDetail{}, appErrors.AdminInternal.Wrap(err, "load media")
	}
	summary, err := mediaSummary(ctx, db, asset)
	if err != nil {
		return contentResponse.MediaDetail{}, err
	}
	var checksum *string
	if value := strings.TrimSpace(asset.Checksum); value != "" {
		checksum = &value
	}
	var latest *contentResponse.ModerationRecordSummary
	var record contentModel.ImageModerationRecord
	if err := db.WithContext(ctx).Where("file_id = ?", asset.ID).
		Order("created_at DESC, id DESC").First(&record).Error; err == nil {
		value, summaryErr := ModerationRecordSummary(ctx, db, record)
		if summaryErr != nil {
			return contentResponse.MediaDetail{}, summaryErr
		}
		latest = &value
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return contentResponse.MediaDetail{}, appErrors.AdminInternal.Wrap(err, "load latest moderation record")
	}
	return contentResponse.MediaDetail{
		MediaSummary:           summary,
		Checksum:               checksum,
		LatestModerationRecord: latest,
		DeletedAt:              asset.DeletedAt,
		DeleteReason:           asset.DeleteReason,
	}, nil
}

// moderationRiskLabels 解析图片审核风险标签。
func moderationRiskLabels(raw []byte) []string {
	result := make([]string, 0)
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &result)
	}
	return result
}

// ModerationRecordSummary 将图片审核模型转换为摘要。
func ModerationRecordSummary(
	ctx context.Context,
	db *gorm.DB,
	record contentModel.ImageModerationRecord,
) (contentResponse.ModerationRecordSummary, error) {
	var user *commonResponse.UserReference
	var asset contentModel.FrontMediaAsset
	if err := db.WithContext(ctx).First(&asset, "id = ?", record.FileID).Error; err == nil {
		if strings.TrimSpace(asset.UserID) != "" {
			var model userModel.MiniAppUser
			if userErr := db.WithContext(ctx).First(&model, "id = ?", asset.UserID).Error; userErr == nil {
				value := serviceCommon.UserReference(model)
				user = &value
			} else if !errors.Is(userErr, gorm.ErrRecordNotFound) {
				return contentResponse.ModerationRecordSummary{}, appErrors.AdminInternal.Wrap(
					userErr,
					"load moderation record user",
				)
			}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return contentResponse.ModerationRecordSummary{}, appErrors.AdminInternal.Wrap(
			err,
			"load moderation record media",
		)
	}
	return contentResponse.ModerationRecordSummary{
		ID:                record.ID,
		RequestID:         record.RequestID,
		FileID:            record.FileID,
		User:              user,
		ObjectType:        record.ObjectType,
		ObjectID:          record.ObjectID,
		Status:            string(record.Status),
		RiskLabels:        moderationRiskLabels(record.RiskLabelsJSON),
		RiskLevel:         record.RiskLevel,
		DurationMS:        record.DurationMS,
		ProviderRequestID: record.ProviderRequestID,
		ErrorCode:         record.ErrorCode,
		CreatedAt:         record.CreatedAt,
	}, nil
}

// ListModerationRecords 分页查询图片审核记录。
func (service *MediaService) ListModerationRecords(
	ctx context.Context,
	query contentRequest.ModerationRecordListQuery,
) (commonResponse.Page[contentResponse.ModerationRecordSummary], error) {
	query.ApplyDefaults()
	if err := validateModerationRecordListQuery(query); err != nil {
		return commonResponse.Page[contentResponse.ModerationRecordSummary]{}, err
	}
	db := service.database()
	if db == nil {
		return commonResponse.Page[contentResponse.ModerationRecordSummary]{}, appErrors.AdminInternal.DefaultMsg()
	}
	table := (contentModel.ImageModerationRecord{}).TableName()
	statement := db.WithContext(ctx).Table(table + " AS records")
	if value := strings.TrimSpace(query.RequestID); value != "" {
		statement = statement.Where("records.request_id = ?", value)
	}
	if value := strings.TrimSpace(query.FileID); value != "" {
		statement = statement.Where("records.file_id = ?", value)
	}
	if value := strings.TrimSpace(query.UserID); value != "" {
		statement = statement.Joins(
			"JOIN "+(contentModel.FrontMediaAsset{}).TableName()+
				" AS media ON media.id = records.file_id",
		).Where("media.user_id = ?", value)
	}
	if value := strings.TrimSpace(query.ObjectType); value != "" {
		statement = statement.Where("records.object_type = ?", value)
	}
	if query.Status == "rejected_or_failed" {
		statement = statement.Where(
			"records.status IN ?",
			[]string{
				string(contentModel.ModerationStatusRejected),
				string(contentModel.ModerationStatusFailed),
			},
		)
	} else if query.Status != "" {
		statement = statement.Where("records.status = ?", query.Status)
	}
	if value := strings.TrimSpace(query.RiskLabel); value != "" {
		encoded, _ := json.Marshal(value)
		statement = statement.Where("JSON_CONTAINS(records.risk_labels_json, ?)", string(encoded))
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("records.created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("records.created_at <= ?", *query.CreatedTo)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return commonResponse.Page[contentResponse.ModerationRecordSummary]{}, appErrors.AdminInternal.Wrap(err, "count moderation records")
	}
	sortColumns := map[string]string{
		"createdAt":  "records.created_at",
		"durationMs": "records.duration_ms",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return commonResponse.Page[contentResponse.ModerationRecordSummary]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	rows := make([]contentModel.ImageModerationRecord, 0)
	if err := statement.Select("records.*").
		Order(column + " " + query.SortOrder).
		Order("records.id ASC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Scan(&rows).Error; err != nil {
		return commonResponse.Page[contentResponse.ModerationRecordSummary]{}, appErrors.AdminInternal.Wrap(err, "list moderation records")
	}
	list := make([]contentResponse.ModerationRecordSummary, 0, len(rows))
	for _, record := range rows {
		summary, err := ModerationRecordSummary(ctx, db, record)
		if err != nil {
			return commonResponse.Page[contentResponse.ModerationRecordSummary]{}, err
		}
		list = append(list, summary)
	}
	return commonResponse.Page[contentResponse.ModerationRecordSummary]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// validateModerationRecordListQuery 校验服务层直接调用时的审核记录查询边界。
func validateModerationRecordListQuery(
	query contentRequest.ModerationRecordListQuery,
) error {
	allowedStatuses := map[string]struct{}{
		"": {}, "pending": {}, "passed": {}, "rejected": {}, "failed": {},
		"not_required": {}, "rejected_or_failed": {},
	}
	if query.Page < 1 ||
		(query.PageSize != 20 && query.PageSize != 50 && query.PageSize != 100) ||
		len(strings.TrimSpace(query.RequestID)) > 128 ||
		len(strings.TrimSpace(query.FileID)) > 128 ||
		len(strings.TrimSpace(query.UserID)) > 64 ||
		len(strings.TrimSpace(query.ObjectType)) > 60 ||
		len(strings.TrimSpace(query.RiskLabel)) > 80 ||
		(query.SortBy != "createdAt" && query.SortBy != "durationMs") ||
		(query.SortOrder != "asc" && query.SortOrder != "desc") {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if _, exists := allowedStatuses[strings.TrimSpace(query.Status)]; !exists {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if query.CreatedFrom != nil && query.CreatedTo != nil &&
		query.CreatedFrom.After(*query.CreatedTo) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// GetModerationRecord 获取图片审核记录详情，并按权限决定是否包含供应商摘要。
func (service *MediaService) GetModerationRecord(
	ctx context.Context,
	recordID string,
	includeSensitive bool,
) (contentResponse.ModerationRecordDetail, error) {
	db := service.database()
	recordID = strings.TrimSpace(recordID)
	if db == nil || recordID == "" {
		return contentResponse.ModerationRecordDetail{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	var record contentModel.ImageModerationRecord
	if err := db.WithContext(ctx).First(&record, "id = ?", recordID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contentResponse.ModerationRecordDetail{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return contentResponse.ModerationRecordDetail{}, appErrors.AdminInternal.Wrap(err, "load moderation record")
	}
	var providerSummary map[string]interface{}
	if includeSensitive && len(record.ProviderSummaryJSON) > 0 {
		_ = json.Unmarshal(record.ProviderSummaryJSON, &providerSummary)
	}
	var endpointLabel *string
	if record.ConfigVersion != nil {
		var version contentModel.ModerationConfig
		if err := db.WithContext(ctx).
			First(&version, "config_version = ?", *record.ConfigVersion).Error; err == nil {
			label := string(version.Provider) + "/" + version.Region
			endpointLabel = &label
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return contentResponse.ModerationRecordDetail{}, appErrors.AdminInternal.Wrap(err, "load moderation endpoint label")
		}
	}
	summary, err := ModerationRecordSummary(ctx, db, record)
	if err != nil {
		return contentResponse.ModerationRecordDetail{}, err
	}
	return contentResponse.ModerationRecordDetail{
		ModerationRecordSummary: summary,
		EndpointLabel:           endpointLabel,
		ErrorSummary:            record.ErrorSummary,
		ProviderResponseSummary: providerSummary,
	}, nil
}

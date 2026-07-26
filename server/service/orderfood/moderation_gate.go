package orderfood

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	moderationCategoryUnconfigured = "unconfigured"
	moderationCategoryDisabled     = "disabled"
	moderationGateMaxImageBytes    = 10 << 20
)

// ModerationGateInput 表示用户图片进入业务对象前所需的审核上下文。
type ModerationGateInput struct {
	RequestID     string                         // 请求ID
	StagingFileID string                         // 待审核临时文件ID
	Scene         orderfoodModel.ModerationScene // 图片使用场景
	ContentType   string                         // 图片MIME类型
	Bytes         []byte                         // 图片原始字节
	ObjectType    *string                        // 可选的业务对象类型
	ObjectID      *string                        // 可选的业务对象ID
}

// PassedImagePermit can only be produced by this package after an explicit
// provider pass (or the immutable generated-cover exclusion).
type PassedImagePermit struct {
	ModerationRecordID string                          // 图片审核记录ID
	ReviewStatus       orderfoodModel.ModerationStatus // 审核通过后的资源状态
	ConfigVersion      *int64                          // 本次审核使用的配置版本
	Scene              orderfoodModel.ModerationScene  // 图片使用场景
}

// PersistPassedImage 定义审核通过后持久化图片业务数据的回调。
type PersistPassedImage func(*gorm.DB, PassedImagePermit) (interface{}, error)

// ImageModerationGate is the integration boundary for mini-program uploads.
// Callers provide their business-table write as a callback; the callback is
// never invoked for review/block/error/timeout/unconfigured/disabled results.
type ImageModerationGate interface {
	AcceptUserImage(
		context.Context,
		ModerationGateInput,
		PersistPassedImage,
	) (interface{}, orderfoodResponse.ImageModerationDecision, error)
}

var _ ImageModerationGate = (*ModerationService)(nil)

// AcceptUserImage 审核用户图片并在同一事务中绑定业务数据。
func (service *ModerationService) AcceptUserImage(
	ctx context.Context,
	input ModerationGateInput,
	persist PersistPassedImage,
) (interface{}, orderfoodResponse.ImageModerationDecision, error) {
	var decision orderfoodResponse.ImageModerationDecision
	db := service.database()
	if db == nil || service == nil || persist == nil {
		return nil, decision, appErrors.FrontInvalidImage.DefaultMsg()
	}
	input.RequestID = strings.TrimSpace(input.RequestID)
	input.StagingFileID = strings.TrimSpace(input.StagingFileID)
	input.ContentType = strings.TrimSpace(input.ContentType)
	if input.RequestID == "" || input.StagingFileID == "" || len(input.Bytes) == 0 ||
		!validModerationScene(input.Scene) ||
		!validModerationImagePayload(input.Bytes, input.ContentType) {
		return nil, decision, appErrors.FrontInvalidImage.DefaultMsg()
	}

	if input.Scene == orderfoodModel.ModerationSceneGeneratedCover {
		record := moderationGateRecord(
			input,
			nil,
			ModerationProviderResult{
				Status:      orderfoodModel.ModerationStatusNotRequired,
				Category:    string(orderfoodModel.ModerationStatusNotRequired),
				RiskLabels:  []string{},
				SafeMessage: "AI 生成图片无需上传阶段自动审核",
			},
			service.now(),
		)
		persisted, err := service.persistAllowedImage(ctx, db, &record, persist)
		if err != nil {
			return nil, decision, err
		}
		return persisted, moderationDecisionFromRecord(record, true), nil
	}

	var current orderfoodModel.ModerationConfig
	err := db.WithContext(ctx).
		Where("singleton_key = ?", moderationConfigSingleton).
		First(&current).Error
	if err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, decision, appErrors.FrontInvalidImage.Wrap(
				err,
				"query effective moderation config",
			)
		}
		failure := moderationFailedProviderResult(
			moderationCategoryUnconfigured,
			"图片审核尚未配置，请稍后重试",
			0,
		)
		record := moderationGateRecord(input, nil, failure, service.now())
		if writeErr := db.WithContext(ctx).Create(&record).Error; writeErr != nil {
			return nil, decision, appErrors.FrontInvalidImage.Wrap(
				writeErr,
				"record unconfigured moderation rejection",
			)
		}
		return nil, moderationDecisionFromRecord(record, false), appErrors.FrontInvalidImage.DefaultMsg()
	}
	if !current.Enabled {
		failure := moderationFailedProviderResult(
			moderationCategoryDisabled,
			"图片审核当前不可用，请稍后重试",
			0,
		)
		record := moderationGateRecord(input, &current, failure, service.now())
		if writeErr := db.WithContext(ctx).Create(&record).Error; writeErr != nil {
			return nil, decision, appErrors.FrontInvalidImage.Wrap(
				writeErr,
				"record disabled moderation rejection",
			)
		}
		return nil, moderationDecisionFromRecord(record, false), appErrors.FrontInvalidImage.DefaultMsg()
	}
	if current.Provider != orderfoodModel.ModerationProviderAliyun ||
		strings.TrimSpace(current.CredentialRef) == "" ||
		service.Provider == nil {
		failure := moderationFailedProviderResult(
			ModerationCategoryUnavailable,
			"图片审核服务暂不可用",
			0,
		)
		record := moderationGateRecord(input, &current, failure, service.now())
		if writeErr := service.persistRejectedDecision(ctx, db, &record, true); writeErr != nil {
			return nil, decision, writeErr
		}
		return nil, moderationDecisionFromRecord(record, false), appErrors.FrontInvalidImage.DefaultMsg()
	}

	providerResult := service.Provider.ModerateImage(
		ctx,
		moderationProviderConfig(current),
		ModerationProviderImage{
			Bytes: input.Bytes, ContentType: input.ContentType, Scene: input.Scene,
		},
	)
	// The only accepting condition is an explicit pass mapping. Any unknown or
	// contradictory provider output is forced to failed.
	explicitPass := providerResult.Status == orderfoodModel.ModerationStatusPassed &&
		providerResult.Category == ModerationCategoryPassed
	if !explicitPass &&
		providerResult.Status != orderfoodModel.ModerationStatusRejected &&
		providerResult.Status != orderfoodModel.ModerationStatusFailed {
		providerResult.Status = orderfoodModel.ModerationStatusFailed
		providerResult.Category = ModerationCategoryInvalidResponse
		providerResult.SafeMessage = "图片审核服务返回无效结果"
	}
	record := moderationGateRecord(input, &current, providerResult, service.now())
	if explicitPass {
		persisted, persistErr := service.persistAllowedImage(ctx, db, &record, persist)
		if persistErr != nil {
			return nil, decision, persistErr
		}
		return persisted, moderationDecisionFromRecord(record, true), nil
	}

	providerFailure := record.Status == orderfoodModel.ModerationStatusFailed
	if err := service.persistRejectedDecision(ctx, db, &record, providerFailure); err != nil {
		return nil, decision, err
	}
	decision = moderationDecisionFromRecord(record, false)
	if record.Status == orderfoodModel.ModerationStatusRejected {
		return nil, decision, appErrors.FrontImageRejected.DefaultMsg()
	}
	return nil, decision, appErrors.FrontInvalidImage.DefaultMsg()
}

func (service *ModerationService) persistAllowedImage(
	ctx context.Context,
	db *gorm.DB,
	record *orderfoodModel.ImageModerationRecord,
	persist PersistPassedImage,
) (interface{}, error) {
	var persisted interface{}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return appErrors.FrontInvalidImage.Wrap(err, "record passed moderation")
		}
		permit := PassedImagePermit{
			ModerationRecordID: record.ID,
			ReviewStatus:       record.Status,
			ConfigVersion:      record.ConfigVersion,
			Scene:              record.Scene,
		}
		result, err := persist(tx, permit)
		if err != nil {
			return err
		}
		persisted = result
		if record.Status == orderfoodModel.ModerationStatusPassed {
			if err := service.markModerationSuccess(tx, record.ConfigVersion); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return persisted, nil
}

func (service *ModerationService) persistRejectedDecision(
	ctx context.Context,
	db *gorm.DB,
	record *orderfoodModel.ImageModerationRecord,
	providerFailure bool,
) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return appErrors.FrontInvalidImage.Wrap(err, "record failed moderation")
		}
		if providerFailure {
			if err := service.markModerationFailure(
				tx,
				record.ConfigVersion,
				moderationStringValue(record.ErrorSummary),
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func moderationGateRecord(
	input ModerationGateInput,
	version *orderfoodModel.ModerationConfig,
	result ModerationProviderResult,
	now time.Time,
) orderfoodModel.ImageModerationRecord {
	riskLabels := result.RiskLabels
	if riskLabels == nil {
		riskLabels = []string{}
	}
	riskLabelsJSON, _ := json.Marshal(riskLabels)
	summaryJSON, _ := json.Marshal(moderationSanitizeProviderSummary(result.ProviderResponseSummary))
	record := orderfoodModel.ImageModerationRecord{
		ID: orderfoodModel.NewID(), RequestID: input.RequestID,
		FileID: input.StagingFileID, Scene: input.Scene, Status: result.Status,
		RiskLabelsJSON: datatypes.JSON(riskLabelsJSON),
		RiskLevel:      result.RiskLevel, ProviderRequestID: result.ProviderRequestID,
		ProviderSummaryJSON: datatypes.JSON(summaryJSON),
		ObjectType:          input.ObjectType, ObjectID: input.ObjectID, CreatedAt: now,
	}
	duration := result.DurationMS
	record.DurationMS = &duration
	if result.Status == orderfoodModel.ModerationStatusFailed {
		category := strings.TrimSpace(result.Category)
		if category == "" {
			category = ModerationCategoryUnavailable
		}
		record.ErrorCode = &category
		summary := strings.TrimSpace(result.SafeMessage)
		if summary == "" {
			summary = "图片审核服务暂不可用"
		}
		record.ErrorSummary = &summary
	}
	if version != nil {
		value := version.ConfigVersion
		record.ConfigVersion = &value
	}
	return record
}

func moderationDecisionFromRecord(
	record orderfoodModel.ImageModerationRecord,
	allowed bool,
) orderfoodResponse.ImageModerationDecision {
	var riskLabels []string
	if err := json.Unmarshal(record.RiskLabelsJSON, &riskLabels); err != nil {
		riskLabels = []string{}
	}
	return orderfoodResponse.ImageModerationDecision{
		RecordID: record.ID, Status: string(record.Status), Allowed: allowed,
		RiskLabels: riskLabels, RiskLevel: record.RiskLevel,
	}
}

func (service *ModerationService) markModerationSuccess(
	tx *gorm.DB,
	configVersion *int64,
) error {
	now := service.now()
	health := orderfoodModel.ModerationConfigHealth{
		SingletonKey:              moderationConfigSingleton,
		Status:                    orderfoodModel.ModerationHealthHealthy,
		EffectiveConfigVersion:    configVersion,
		LastModerationSucceededAt: &now,
		ConsecutiveFailureCount:   0, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "singleton_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"status", "effective_config_version",
			"last_moderation_succeeded_at", "consecutive_failure_count",
			"last_failure_at", "last_failure_safe_summary", "updated_at",
		}),
	}).Create(&health).Error; err != nil {
		return appErrors.FrontInvalidImage.Wrap(err, "update moderation success health")
	}
	return nil
}

func (service *ModerationService) markModerationFailure(
	tx *gorm.DB,
	configVersion *int64,
	safeSummary string,
) error {
	now := service.now()
	var current orderfoodModel.ModerationConfigHealth
	err := tx.Where("singleton_key = ?", moderationConfigSingleton).First(&current).Error
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.FrontInvalidImage.Wrap(err, "query moderation failure health")
	}
	failureCount := current.ConsecutiveFailureCount + 1
	status := orderfoodModel.ModerationHealthDegraded
	if failureCount >= 3 {
		status = orderfoodModel.ModerationHealthUnavailable
	}
	safeSummary = strings.TrimSpace(safeSummary)
	if safeSummary == "" {
		safeSummary = "图片审核服务暂不可用"
	}
	health := orderfoodModel.ModerationConfigHealth{
		ID: current.ID, SingletonKey: moderationConfigSingleton,
		Status: status, EffectiveConfigVersion: configVersion,
		LastConnectionTestID:      current.LastConnectionTestID,
		LastModerationSucceededAt: current.LastModerationSucceededAt,
		LastFailureAt:             &now, LastFailureSafeSummary: &safeSummary,
		ConsecutiveFailureCount: failureCount,
		CreatedAt:               current.CreatedAt, UpdatedAt: now,
	}
	if health.CreatedAt.IsZero() {
		health.CreatedAt = now
	}
	if err := tx.Save(&health).Error; err != nil {
		return appErrors.FrontInvalidImage.Wrap(err, "update moderation failure health")
	}
	return nil
}

func validModerationScene(scene orderfoodModel.ModerationScene) bool {
	for _, mandatory := range mandatoryModerationScenes {
		if scene == mandatory {
			return true
		}
	}
	return scene == orderfoodModel.ModerationSceneGeneratedCover
}

func moderationStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validModerationImagePayload(payload []byte, declaredContentType string) bool {
	if len(payload) == 0 || len(payload) > moderationGateMaxImageBytes {
		return false
	}
	detected := http.DetectContentType(payload)
	switch detected {
	case "image/jpeg", "image/png", "image/gif":
	default:
		return false
	}
	declaredContentType = strings.ToLower(strings.TrimSpace(declaredContentType))
	if declaredContentType != "" {
		if semicolon := strings.IndexByte(declaredContentType, ';'); semicolon >= 0 {
			declaredContentType = strings.TrimSpace(declaredContentType[:semicolon])
		}
		if declaredContentType == "image/jpg" {
			declaredContentType = "image/jpeg"
		}
		if declaredContentType != detected {
			return false
		}
	}
	_, _, err := image.DecodeConfig(bytes.NewReader(payload))
	return err == nil
}

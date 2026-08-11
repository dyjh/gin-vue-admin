package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var credentialReferencePattern = regexp.MustCompile(`^env://[A-Za-z_][A-Za-z0-9_]*$`)

const (
	PermissionModerationConfigRead            = "orderfood:moderation-config:read"
	PermissionModerationConfigUpdate          = "orderfood:moderation-config:update"
	PermissionModerationConfigCredentialWrite = "orderfood:moderation-config:credential-write"
	PermissionModerationConfigTest            = "orderfood:moderation-config:test"
	PermissionModerationSensitiveRead         = "orderfood:moderation:sensitive-read"

	moderationConfigSingleton = "default"
	moderationTestValidity    = 30 * time.Minute
)

var mandatoryModerationScenes = []contentModel.ModerationScene{
	contentModel.ModerationSceneProfileAvatar,
	contentModel.ModerationSceneDishCover,
	contentModel.ModerationSceneDishStep,
	contentModel.ModerationSceneDishExtract,
	contentModel.ModerationSceneCheckin,
}

var excludedModerationScenes = []contentModel.ModerationScene{
	contentModel.ModerationSceneGeneratedCover,
}

// ModerationService 提供图片审核配置和测试能力。
type ModerationService struct {
	DB            *gorm.DB                          // 数据库连接
	Provider      ModerationProvider                // 图片审核供应商
	Permission    *serviceCommon.PermissionService  // 管理端权限服务
	AccessAudit   *serviceCommon.AccessAuditService // 敏感数据访问审计服务
	Idempotency   *serviceCommon.IdempotencyService // 幂等请求服务
	MutationAudit serviceCommon.MutationAuditWriter // 配置变更审计写入器
	Now           func() time.Time                  // 当前时间函数
}

// ModerationServiceOption 定义图片审核服务配置选项。
type ModerationServiceOption func(*ModerationService)

// WithModerationMutationAuditWriter 配置图片审核变更审计写入器。
func WithModerationMutationAuditWriter(
	writer serviceCommon.MutationAuditWriter,
) ModerationServiceOption {
	return func(service *ModerationService) {
		service.MutationAudit = writer
	}
}

// NewModerationService 创建图片审核服务实例。
func NewModerationService(
	db *gorm.DB,
	provider ModerationProvider,
	permission *serviceCommon.PermissionService,
	accessAudit *serviceCommon.AccessAuditService,
	idempotency *serviceCommon.IdempotencyService,
	options ...ModerationServiceOption,
) *ModerationService {
	if provider == nil {
		provider = NewProductionAliyunModerationProvider()
	}
	if permission == nil {
		permission = serviceCommon.NewPermissionService(db)
	}
	if accessAudit == nil {
		accessAudit = serviceCommon.NewAccessAuditService(db)
	}
	if idempotency == nil {
		idempotency = serviceCommon.NewIdempotencyService(db)
	}
	service := &ModerationService{
		DB: db, Provider: provider, Permission: permission,
		AccessAudit: accessAudit, Idempotency: idempotency,
		MutationAudit: serviceCommon.GormMutationAuditWriter{}, Now: time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (service *ModerationService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *ModerationService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// ensureDefaultConfig 初始化可直接编辑的关闭状态默认配置。
func (service *ModerationService) ensureDefaultConfig(ctx context.Context) error {
	db := service.database()
	if db == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	now := service.now()
	config := contentModel.ModerationConfig{
		SingletonKey:      moderationConfigSingleton,
		ConfigVersion:     1,
		Provider:          contentModel.ModerationProviderAliyun,
		Enabled:           false,
		Region:            "cn-shanghai",
		Endpoint:          "https://green-cip.cn-shanghai.aliyuncs.com",
		ServiceCode:       "baselineCheck",
		TimeoutMS:         3000,
		RetryCount:        1,
		RetryBackoffMS:    200,
		AppliedByUsername: "system",
		AppliedAt:         now,
		Reason:            "系统默认图片审核配置",
	}
	config.ConfigHash = moderationConfigHash(config)
	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&config).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "seed default moderation config")
	}
	health := contentModel.ModerationConfigHealth{
		SingletonKey:            moderationConfigSingleton,
		Status:                  contentModel.ModerationHealthDisabled,
		EffectiveConfigVersion:  &config.ConfigVersion,
		ConsecutiveFailureCount: 0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&health).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "seed default moderation health")
	}
	return nil
}

// GetConfig 获取当前生效的图片审核配置。
func (service *ModerationService) GetConfig(
	ctx context.Context,
	actor commonRequest.AdminActor,
) (contentResponse.ModerationConfig, error) {
	var result contentResponse.ModerationConfig
	if err := service.require(ctx, actor, PermissionModerationConfigRead); err != nil {
		return result, err
	}
	if err := service.ensureDefaultConfig(ctx); err != nil {
		return result, err
	}
	current, err := loadCurrentModerationConfig(ctx, service.database())
	if err != nil {
		return result, err
	}
	return service.buildConfig(ctx, service.database(), current)
}

// UpdateConfig 保存并立即应用图片审核配置。
func (service *ModerationService) UpdateConfig(
	ctx context.Context,
	input contentRequest.ModerationConfigUpdateInput,
	actor commonRequest.AdminActor,
	idempotencyKey string,
) (contentResponse.ModerationConfig, bool, error) {
	var result contentResponse.ModerationConfig
	if err := service.require(ctx, actor, PermissionModerationConfigUpdate); err != nil {
		return result, false, err
	}
	if input.CredentialRef != nil {
		if err := service.require(ctx, actor, PermissionModerationConfigCredentialWrite); err != nil {
			return result, false, err
		}
	}
	if err := validateModerationConfigInput(input); err != nil {
		return result, false, err
	}
	if service.Idempotency == nil {
		return result, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.ensureDefaultConfig(ctx); err != nil {
		return result, false, err
	}

	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"moderation_config_update",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			current, loadErr := loadCurrentModerationConfig(ctx, tx)
			if loadErr != nil {
				return nil, loadErr
			}
			if current.ConfigVersion != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			now := service.now()
			credentialRef := current.CredentialRef
			credentialUpdatedAt := current.CredentialUpdatedAt
			if input.CredentialRef != nil {
				credentialRef = strings.TrimSpace(*input.CredentialRef)
				if credentialRef == "" {
					return nil, appErrors.AdminInvalidConfig.DefaultMsg()
				}
				credentialUpdatedAt = &now
			}
			if input.Enabled && credentialRef == "" {
				return nil, appErrors.AdminInvalidConfig.DefaultMsg()
			}
			next := contentModel.ModerationConfig{
				SingletonKey:        moderationConfigSingleton,
				ConfigVersion:       current.ConfigVersion + 1,
				Provider:            contentModel.ModerationProviderAliyun,
				Enabled:             input.Enabled,
				Region:              strings.TrimSpace(input.Region),
				Endpoint:            strings.TrimSpace(input.Endpoint),
				ServiceCode:         strings.TrimSpace(input.ServiceCode),
				TimeoutMS:           input.TimeoutMS,
				RetryCount:          input.RetryCount,
				RetryBackoffMS:      input.RetryBackoffMS,
				CredentialRef:       credentialRef,
				CredentialUpdatedAt: credentialUpdatedAt,
				AppliedByID:         actor.AdministratorID,
				AppliedByUsername:   actor.Username,
				AppliedByNickname:   actor.Nickname,
				AppliedAt:           now,
				Reason:              strings.TrimSpace(input.Reason),
			}
			next.ConfigHash = moderationConfigHash(next)
			update := tx.Model(&contentModel.ModerationConfig{}).
				Where("singleton_key = ? AND config_version = ?", moderationConfigSingleton, input.ExpectedVersion).
				Select("*").Omit("singleton_key").Updates(&next)
			if update.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(update.Error, "save moderation config")
			}
			if update.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			health, healthErr := service.nextConfigHealth(tx, current, next, now)
			if healthErr != nil {
				return nil, healthErr
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "singleton_key"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"status", "effective_config_version", "last_connection_test_id",
					"last_moderation_succeeded_at", "last_failure_at",
					"last_failure_safe_summary", "consecutive_failure_count", "updated_at",
				}),
			}).Create(&health).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "update moderation health")
			}
			if err := service.writeModerationMutationAudit(
				ctx,
				tx,
				actor,
				"update_moderation_config",
				strconv.FormatInt(next.ConfigVersion, 10),
				next.Reason,
				idempotencyKey,
				moderationConfigAuditSummary(current, true),
				moderationConfigAuditSummary(next, true),
			); err != nil {
				return nil, err
			}
			return service.buildConfig(ctx, tx, next)
		},
	)
	if err != nil {
		return result, false, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode moderation config response")
	}
	return result, replayed, nil
}

// TestConnection 测试当前图片审核配置连接。
func (service *ModerationService) TestConnection(
	ctx context.Context,
	input contentRequest.ModerationConfigTestInput,
	actor commonRequest.AdminActor,
	idempotencyKey string,
) (contentResponse.ModerationConnectionTestResult, bool, error) {
	var result contentResponse.ModerationConnectionTestResult
	if err := service.require(ctx, actor, PermissionModerationConfigTest); err != nil {
		return result, false, err
	}
	if service.Idempotency == nil || service.Provider == nil {
		return result, false, appErrors.AdminInternal.DefaultMsg()
	}
	if err := service.ensureDefaultConfig(ctx); err != nil {
		return result, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"moderation_config_connection_test",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			current, loadErr := loadCurrentModerationConfig(ctx, tx)
			if loadErr != nil {
				return nil, loadErr
			}
			if input.ExpectedVersion != nil && current.ConfigVersion != *input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			providerResult := service.Provider.TestConnection(
				ctx,
				moderationProviderConfig(current),
			)
			now := service.now()
			test := contentModel.ModerationConnectionTest{
				ID:               commonModel.NewID(),
				ConfigVersion:    current.ConfigVersion,
				ConfigHash:       current.ConfigHash,
				Success:          providerResult.Category == ModerationCategoryOK,
				Category:         providerResult.Category,
				RequestID:        actor.RequestID,
				DurationMS:       providerResult.DurationMS,
				TestedByID:       actor.AdministratorID,
				TestedByUsername: actor.Username,
				TestedByNickname: actor.Nickname,
				TestedAt:         now,
				SafeMessage:      providerResult.SafeMessage,
			}
			if test.Success {
				validUntil := now.Add(moderationTestValidity)
				test.ValidUntil = &validUntil
			}
			if err := tx.Create(&test).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create moderation connection test")
			}
			if err := service.attachConnectionTestToHealth(tx, test, now); err != nil {
				return nil, err
			}
			if err := service.writeModerationMutationAudit(
				ctx,
				tx,
				actor,
				"test_moderation_connection",
				test.ID,
				"测试当前图片审核配置连接",
				idempotencyKey,
				moderationConfigAuditSummary(current, true),
				moderationConnectionTestAuditSummary(test),
			); err != nil {
				return nil, err
			}
			return moderationConnectionTestResponse(test), nil
		},
	)
	if err != nil {
		return result, false, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode connection test response")
	}
	return result, replayed, nil
}

// TestImage 测试当前图片审核配置的结果映射。
func (service *ModerationService) TestImage(
	ctx context.Context,
	input contentRequest.ModerationConfigTestInput,
	image ModerationProviderImage,
	actor commonRequest.AdminActor,
	idempotencyKey string,
) (contentResponse.ModerationImageTestResult, bool, error) {
	var result contentResponse.ModerationImageTestResult
	if err := service.require(ctx, actor, PermissionModerationConfigTest); err != nil {
		return result, false, err
	}
	if len(image.Bytes) == 0 || service.Idempotency == nil || service.Provider == nil {
		return result, false, appErrors.AdminInvalidImage.DefaultMsg()
	}
	if err := service.ensureDefaultConfig(ctx); err != nil {
		return result, false, err
	}
	image.Scene = contentModel.ModerationSceneDishCover
	payload := struct {
		ExpectedVersion *int64 `json:"expectedVersion"`
		ImageHash       string `json:"imageHash"`
		ContentType     string `json:"contentType"`
	}{
		ExpectedVersion: input.ExpectedVersion,
		ImageHash:       moderationBytesHash(image.Bytes),
		ContentType:     image.ContentType,
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"moderation_config_moderation_test",
		strings.TrimSpace(idempotencyKey),
		payload,
		func(tx *gorm.DB) (interface{}, error) {
			current, loadErr := loadCurrentModerationConfig(ctx, tx)
			if loadErr != nil {
				return nil, loadErr
			}
			if input.ExpectedVersion != nil && current.ConfigVersion != *input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			providerResult := service.Provider.ModerateImage(
				ctx,
				moderationProviderConfig(current),
				image,
			)
			now := service.now()
			riskLabelsJSON, _ := json.Marshal(providerResult.RiskLabels)
			summaryJSON, _ := json.Marshal(
				moderationSanitizeProviderSummary(providerResult.ProviderResponseSummary),
			)
			test := contentModel.ModerationImageTest{
				ID:                   commonModel.NewID(),
				ConfigVersion:        current.ConfigVersion,
				MappedStatus:         providerResult.Status,
				RiskLabelsJSON:       datatypes.JSON(riskLabelsJSON),
				RiskLevel:            providerResult.RiskLevel,
				Category:             providerResult.Category,
				ProviderRequestID:    providerResult.ProviderRequestID,
				RequestID:            actor.RequestID,
				DurationMS:           providerResult.DurationMS,
				SafeMessage:          providerResult.SafeMessage,
				ProviderSummaryJSON:  datatypes.JSON(summaryJSON),
				TemporaryFileCleaned: true,
				TestedByID:           actor.AdministratorID,
				TestedAt:             now,
			}
			if err := tx.Create(&test).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create moderation image test")
			}
			responseResult := moderationImageTestResponse(test, providerResult.RiskLabels, nil)
			canReadSummary, permissionErr := service.Permission.HasPermission(
				ctx,
				actor.AuthorityID,
				PermissionModerationSensitiveRead,
			)
			if permissionErr != nil {
				return nil, permissionErr
			}
			if canReadSummary && len(providerResult.ProviderResponseSummary) > 0 {
				if service.AccessAudit == nil {
					return nil, appErrors.AdminInternal.DefaultMsg()
				}
				if _, accessErr := service.AccessAudit.RecordWithDB(
					ctx,
					tx,
					serviceCommon.SensitiveAccess{
						AdministratorID:  actor.AdministratorID,
						AuthorityID:      actor.AuthorityID,
						Permission:       PermissionModerationSensitiveRead,
						Action:           "read_moderation_test_provider_summary",
						TargetType:       "moderation_image_test",
						TargetID:         test.ID,
						RequestID:        actor.RequestID,
						SourceIPMasked:   actor.SourceIPMasked,
						UserAgentSummary: actor.UserAgentSummary,
					},
				); accessErr != nil {
					return nil, accessErr
				}
				responseResult.ProviderResponseSummary = moderationSanitizeProviderSummary(
					providerResult.ProviderResponseSummary,
				)
			}
			if err := service.writeModerationMutationAudit(
				ctx,
				tx,
				actor,
				"test_moderation_image",
				test.ID,
				"测试图片审核结果映射",
				idempotencyKey,
				moderationConfigAuditSummary(current, true),
				moderationImageTestAuditSummary(test, providerResult.RiskLabels),
			); err != nil {
				return nil, err
			}
			return responseResult, nil
		},
	)
	if err != nil {
		return result, false, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, false, appErrors.AdminInternal.Wrap(err, "decode moderation image test response")
	}
	return result, replayed, nil
}

func (service *ModerationService) require(
	ctx context.Context,
	actor commonRequest.AdminActor,
	permission string,
) error {
	if service == nil || service.Permission == nil ||
		actor.AdministratorID == 0 || actor.AuthorityID == 0 {
		return appErrors.AdminLoginExpired.DefaultMsg()
	}
	return service.Permission.Require(ctx, actor.AuthorityID, permission)
}

func validateModerationConfigInput(
	input contentRequest.ModerationConfigUpdateInput,
) error {
	reason := strings.TrimSpace(input.Reason)
	if strings.TrimSpace(input.Region) == "" ||
		strings.TrimSpace(input.ServiceCode) == "" ||
		input.TimeoutMS < 1000 || input.TimeoutMS > 10000 ||
		input.RetryCount < 0 || input.RetryCount > 3 ||
		input.RetryBackoffMS < 0 || input.RetryBackoffMS > 2000 ||
		input.ExpectedVersion < 1 ||
		len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	endpoint, err := url.ParseRequestURI(strings.TrimSpace(input.Endpoint))
	if err != nil ||
		!strings.EqualFold(endpoint.Scheme, "https") ||
		endpoint.Host == "" ||
		endpoint.User != nil ||
		(endpoint.Path != "" && endpoint.Path != "/") ||
		endpoint.RawQuery != "" ||
		endpoint.Fragment != "" {
		return appErrors.AdminInvalidConfig.DefaultMsg()
	}
	if input.CredentialRef != nil {
		credentialRef := strings.TrimSpace(*input.CredentialRef)
		if credentialRef == "" ||
			!credentialReferencePattern.MatchString(credentialRef) ||
			!strings.HasPrefix(credentialRef, "env://") {
			return appErrors.AdminInvalidConfig.DefaultMsg()
		}
	}
	return nil
}

func loadCurrentModerationConfig(
	ctx context.Context,
	db *gorm.DB,
) (contentModel.ModerationConfig, error) {
	var current contentModel.ModerationConfig
	if db == nil {
		return current, appErrors.AdminInternal.DefaultMsg()
	}
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", moderationConfigSingleton).
		First(&current).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return current, appErrors.AdminNotFound.DefaultMsg()
		}
		return current, appErrors.AdminInternal.Wrap(err, "load current moderation config")
	}
	return current, nil
}

func (service *ModerationService) nextConfigHealth(
	tx *gorm.DB,
	current contentModel.ModerationConfig,
	next contentModel.ModerationConfig,
	now time.Time,
) (contentModel.ModerationConfigHealth, error) {
	status := contentModel.ModerationHealthDegraded
	if !next.Enabled {
		status = contentModel.ModerationHealthDisabled
	}
	health := contentModel.ModerationConfigHealth{
		SingletonKey:            moderationConfigSingleton,
		Status:                  status,
		EffectiveConfigVersion:  &next.ConfigVersion,
		ConsecutiveFailureCount: 0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if !sameModerationConnectionMaterial(current, next) {
		return health, nil
	}
	var existing contentModel.ModerationConfigHealth
	if err := tx.Where("singleton_key = ?", moderationConfigSingleton).
		First(&existing).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return health, nil
		}
		return health, appErrors.AdminInternal.Wrap(err, "load moderation health")
	}
	health.LastConnectionTestID = existing.LastConnectionTestID
	health.LastModerationSucceededAt = existing.LastModerationSucceededAt
	health.LastFailureAt = existing.LastFailureAt
	health.LastFailureSafeSummary = existing.LastFailureSafeSummary
	health.ConsecutiveFailureCount = existing.ConsecutiveFailureCount
	if next.Enabled {
		switch existing.Status {
		case contentModel.ModerationHealthHealthy,
			contentModel.ModerationHealthDegraded,
			contentModel.ModerationHealthUnavailable:
			health.Status = existing.Status
		}
	}
	return health, nil
}

func (service *ModerationService) buildConfig(
	ctx context.Context,
	db *gorm.DB,
	config contentModel.ModerationConfig,
) (contentResponse.ModerationConfig, error) {
	health, err := service.buildHealth(ctx, db)
	if err != nil {
		return contentResponse.ModerationConfig{}, err
	}
	return contentResponse.ModerationConfig{
		ModerationConfigValues: moderationConfigValues(
			config.Provider,
			config.Enabled,
			config.Region,
			config.Endpoint,
			config.ServiceCode,
			config.TimeoutMS,
			config.RetryCount,
			config.RetryBackoffMS,
			config.CredentialRef,
			config.CredentialUpdatedAt,
		),
		ConfigVersion:      config.ConfigVersion,
		ConfigHash:         config.ConfigHash,
		Health:             health,
		LastConnectionTest: health.LastConnectionTest,
		UpdatedBy: moderationAdministratorSummary(
			config.AppliedByID,
			config.AppliedByUsername,
			config.AppliedByNickname,
		),
		UpdatedAt: config.AppliedAt,
		Reason:    config.Reason,
	}, nil
}

func (service *ModerationService) buildHealth(
	ctx context.Context,
	db *gorm.DB,
) (contentResponse.ModerationConfigHealth, error) {
	result := contentResponse.ModerationConfigHealth{
		Status:    string(contentModel.ModerationHealthUnconfigured),
		UpdatedAt: service.now(),
	}
	var health contentModel.ModerationConfigHealth
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", moderationConfigSingleton).
		First(&health).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return result, nil
		}
		return result, appErrors.AdminInternal.Wrap(err, "load moderation health")
	}
	result.Status = string(health.Status)
	result.EffectiveConfigVersion = health.EffectiveConfigVersion
	result.LastModerationSucceededAt = health.LastModerationSucceededAt
	result.LastFailureAt = health.LastFailureAt
	result.LastFailureSafeSummary = health.LastFailureSafeSummary
	result.ConsecutiveFailureCount = health.ConsecutiveFailureCount
	result.UpdatedAt = health.UpdatedAt
	if health.LastConnectionTestID != nil {
		test, err := moderationFindConnectionTest(ctx, db, *health.LastConnectionTestID)
		if err != nil {
			return result, err
		}
		if test != nil {
			responseTest := moderationConnectionTestResponse(*test)
			result.LastConnectionTest = &responseTest
		}
	}
	return result, nil
}

func (service *ModerationService) attachConnectionTestToHealth(
	tx *gorm.DB,
	test contentModel.ModerationConnectionTest,
	now time.Time,
) error {
	status := contentModel.ModerationHealthHealthy
	if !test.Success {
		status = contentModel.ModerationHealthUnavailable
	}
	configVersion := test.ConfigVersion
	health := contentModel.ModerationConfigHealth{
		SingletonKey:            moderationConfigSingleton,
		Status:                  status,
		EffectiveConfigVersion:  &configVersion,
		LastConnectionTestID:    &test.ID,
		ConsecutiveFailureCount: 0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if !test.Success {
		health.LastFailureAt = &now
		message := serviceCommon.TruncateRunes(strings.TrimSpace(test.SafeMessage), 240)
		health.LastFailureSafeSummary = &message
		health.ConsecutiveFailureCount = 1
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "singleton_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"status", "effective_config_version", "last_connection_test_id",
			"last_failure_at", "last_failure_safe_summary",
			"consecutive_failure_count", "updated_at",
		}),
	}).Create(&health).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "attach moderation connection test to health")
	}
	return nil
}

func moderationFindConnectionTest(
	ctx context.Context,
	db *gorm.DB,
	id string,
) (*contentModel.ModerationConnectionTest, error) {
	var test contentModel.ModerationConnectionTest
	if err := db.WithContext(ctx).Where("id = ?", id).First(&test).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "load moderation connection test")
	}
	return &test, nil
}

func moderationConnectionTestResponse(
	test contentModel.ModerationConnectionTest,
) contentResponse.ModerationConnectionTestResult {
	return contentResponse.ModerationConnectionTestResult{
		TestID:        test.ID,
		ConfigVersion: test.ConfigVersion,
		ConfigHash:    test.ConfigHash,
		Success:       test.Success,
		Category:      test.Category,
		RequestID:     test.RequestID,
		DurationMS:    test.DurationMS,
		TestedBy: moderationAdministratorSummary(
			test.TestedByID,
			test.TestedByUsername,
			test.TestedByNickname,
		),
		TestedAt:    test.TestedAt,
		ValidUntil:  test.ValidUntil,
		SafeMessage: test.SafeMessage,
	}
}

func moderationImageTestResponse(
	test contentModel.ModerationImageTest,
	riskLabels []string,
	providerSummary map[string]interface{},
) contentResponse.ModerationImageTestResult {
	if riskLabels == nil {
		riskLabels = []string{}
	}
	return contentResponse.ModerationImageTestResult{
		TestID:                  test.ID,
		ConfigVersion:           test.ConfigVersion,
		MappedStatus:            string(test.MappedStatus),
		RiskLabels:              riskLabels,
		RiskLevel:               test.RiskLevel,
		Category:                test.Category,
		ProviderRequestID:       test.ProviderRequestID,
		RequestID:               test.RequestID,
		DurationMS:              test.DurationMS,
		SafeMessage:             test.SafeMessage,
		ProviderResponseSummary: providerSummary,
		TemporaryFileCleaned:    test.TemporaryFileCleaned,
		TestedAt:                test.TestedAt,
	}
}

// moderationConfigValues 组装图片审核当前配置项并隐藏供应商凭证明文。
func moderationConfigValues(
	provider contentModel.ModerationProviderType,
	enabled bool,
	region string,
	endpoint string,
	serviceCode string,
	timeoutMS int,
	retryCount int,
	retryBackoffMS int,
	credentialRef string,
	credentialUpdatedAt *time.Time,
) contentResponse.ModerationConfigValues {
	return contentResponse.ModerationConfigValues{
		Provider:        string(provider),
		Enabled:         enabled,
		Region:          region,
		Endpoint:        endpoint,
		ServiceCode:     serviceCode,
		TimeoutMS:       timeoutMS,
		RetryCount:      retryCount,
		RetryBackoffMS:  retryBackoffMS,
		WorstCaseWaitMS: moderationWorstCaseWait(timeoutMS, retryCount, retryBackoffMS),
		Credential: contentResponse.ModerationCredentialStatus{
			Configured: strings.TrimSpace(credentialRef) != "",
			UpdatedAt:  credentialUpdatedAt,
		},
		SafetyPolicy: moderationSafetyPolicy(),
	}
}

func moderationSafetyPolicy() contentResponse.ModerationSafetyPolicy {
	return contentResponse.ModerationSafetyPolicy{
		MandatoryScenes:         moderationSceneStrings(mandatoryModerationScenes),
		ExcludedScenes:          moderationSceneStrings(excludedModerationScenes),
		ExecutionMode:           "synchronous_before_accept",
		FailurePolicy:           "fail_closed",
		DisabledBehavior:        "reject_upload",
		UserUploadSuccessStatus: "passed",
		ProviderPassedMapping:   "passed",
		ProviderRiskMapping:     "rejected",
		ProviderFailureMapping:  "failed",
	}
}

func moderationWorstCaseWait(timeoutMS int, retryCount int, retryBackoffMS int) int {
	return timeoutMS*(retryCount+1) + retryBackoffMS*retryCount
}

func moderationConfigHash(config contentModel.ModerationConfig) string {
	material := struct {
		Provider       contentModel.ModerationProviderType `json:"provider"`
		Enabled        bool                                `json:"enabled"`
		Region         string                              `json:"region"`
		Endpoint       string                              `json:"endpoint"`
		ServiceCode    string                              `json:"serviceCode"`
		TimeoutMS      int                                 `json:"timeoutMs"`
		RetryCount     int                                 `json:"retryCount"`
		RetryBackoffMS int                                 `json:"retryBackoffMs"`
		CredentialRef  string                              `json:"credentialRef"`
	}{
		Provider:       config.Provider,
		Enabled:        config.Enabled,
		Region:         config.Region,
		Endpoint:       config.Endpoint,
		ServiceCode:    config.ServiceCode,
		TimeoutMS:      config.TimeoutMS,
		RetryCount:     config.RetryCount,
		RetryBackoffMS: config.RetryBackoffMS,
		CredentialRef:  config.CredentialRef,
	}
	encoded, _ := json.Marshal(material)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func sameModerationConnectionMaterial(
	left contentModel.ModerationConfig,
	right contentModel.ModerationConfig,
) bool {
	return left.Provider == right.Provider &&
		left.Region == right.Region &&
		left.Endpoint == right.Endpoint &&
		left.ServiceCode == right.ServiceCode &&
		left.TimeoutMS == right.TimeoutMS &&
		left.RetryCount == right.RetryCount &&
		left.RetryBackoffMS == right.RetryBackoffMS &&
		left.CredentialRef == right.CredentialRef &&
		moderationOptionalTimesEqual(left.CredentialUpdatedAt, right.CredentialUpdatedAt)
}

// moderationOptionalTimesEqual 判断两个可选时间是否表示同一时刻。
func moderationOptionalTimesEqual(left *time.Time, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func moderationSceneStrings(input []contentModel.ModerationScene) []string {
	result := make([]string, 0, len(input))
	for _, item := range input {
		result = append(result, string(item))
	}
	return result
}

func moderationProviderConfig(
	version contentModel.ModerationConfig,
) ModerationProviderConfig {
	return ModerationProviderConfig{
		Provider: version.Provider, Region: version.Region,
		Endpoint: version.Endpoint, ServiceCode: version.ServiceCode,
		TimeoutMS: version.TimeoutMS, RetryCount: version.RetryCount,
		RetryBackoffMS: version.RetryBackoffMS,
		CredentialRef:  version.CredentialRef,
		ConfigVersion:  version.ConfigVersion,
	}
}

func moderationAdministratorSummary(
	id uint,
	username string,
	nickname *string,
) commonResponse.AdministratorSummary {
	return commonResponse.AdministratorSummary{
		ID:       strconv.FormatUint(uint64(id), 10),
		Username: username,
		Nickname: nickname,
	}
}

func moderationBytesHash(input []byte) string {
	sum := sha256.Sum256(input)
	return hex.EncodeToString(sum[:])
}

func (service *ModerationService) writeModerationMutationAudit(
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
	requestID := serviceCommon.TruncateRunes(strings.TrimSpace(actor.RequestID), 96)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if actor.AdministratorID == 0 || strings.TrimSpace(actor.Username) == "" ||
		requestID == "" || idempotencyKey == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	beforeJSON, err := serviceCommon.SafeAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := serviceCommon.SafeAuditJSON(after)
	if err != nil {
		return err
	}
	var reasonPointer *string
	if value := serviceCommon.TruncateRunes(strings.TrimSpace(reason), 200); value != "" {
		reasonPointer = &value
	}
	audit := auditModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            "moderation_config",
		TargetID:              targetID,
		Reason:                reasonPointer,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             requestID,
		IdempotencyKey:        &idempotencyKey,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := serviceCommon.TruncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

func moderationConfigAuditSummary(
	version contentModel.ModerationConfig,
	exists bool,
) interface{} {
	if !exists {
		return nil
	}
	return map[string]interface{}{
		"configVersion":        version.ConfigVersion,
		"provider":             version.Provider,
		"enabled":              version.Enabled,
		"region":               version.Region,
		"endpoint":             version.Endpoint,
		"serviceCode":          version.ServiceCode,
		"timeoutMs":            version.TimeoutMS,
		"retryCount":           version.RetryCount,
		"retryBackoffMs":       version.RetryBackoffMS,
		"credentialConfigured": strings.TrimSpace(version.CredentialRef) != "",
		"configHash":           version.ConfigHash,
	}
}

func moderationConnectionTestAuditSummary(
	test contentModel.ModerationConnectionTest,
) interface{} {
	return map[string]interface{}{
		"testId":        test.ID,
		"configVersion": test.ConfigVersion,
		"configHash":    test.ConfigHash,
		"success":       test.Success,
		"category":      test.Category,
		"durationMs":    test.DurationMS,
	}
}

func moderationImageTestAuditSummary(
	test contentModel.ModerationImageTest,
	riskLabels []string,
) interface{} {
	return map[string]interface{}{
		"testId":        test.ID,
		"configVersion": test.ConfigVersion,
		"mappedStatus":  test.MappedStatus,
		"riskLabels":    riskLabels,
		"riskLevel":     test.RiskLevel,
		"category":      test.Category,
		"durationMs":    test.DurationMS,
	}
}

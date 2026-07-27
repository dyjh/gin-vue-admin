package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	PermissionModerationConfigRead            = "orderfood:moderation-config:read"
	PermissionModerationConfigUpdate          = "orderfood:moderation-config:update"
	PermissionModerationConfigCredentialWrite = "orderfood:moderation-config:credential-write"
	PermissionModerationConfigTest            = "orderfood:moderation-config:test"
	PermissionModerationSensitiveRead         = "orderfood:moderation:sensitive-read"

	moderationConfigSingleton = "default"
	moderationTestValidity    = 30 * time.Minute
)

var mandatoryModerationScenes = []orderfoodModel.ModerationScene{
	orderfoodModel.ModerationSceneProfileAvatar,
	orderfoodModel.ModerationSceneDishCover,
	orderfoodModel.ModerationSceneDishStep,
	orderfoodModel.ModerationSceneDishExtract,
	orderfoodModel.ModerationSceneCheckin,
}

var excludedModerationScenes = []orderfoodModel.ModerationScene{
	orderfoodModel.ModerationSceneGeneratedCover,
}

// ModerationService 提供图片审核配置和测试能力。
type ModerationService struct {
	DB            *gorm.DB            // 数据库连接
	Provider      ModerationProvider  // 图片审核供应商
	Permission    *PermissionService  // 管理端权限服务
	AccessAudit   *AccessAuditService // 敏感数据访问审计服务
	Idempotency   *IdempotencyService // 幂等请求服务
	MutationAudit MutationAuditWriter // 配置变更审计写入器
	Now           func() time.Time    // 当前时间函数
}

// ModerationServiceOption 定义图片审核服务配置选项。
type ModerationServiceOption func(*ModerationService)

// WithModerationMutationAuditWriter 配置图片审核变更审计写入器。
func WithModerationMutationAuditWriter(
	writer MutationAuditWriter,
) ModerationServiceOption {
	return func(service *ModerationService) {
		service.MutationAudit = writer
	}
}

// NewModerationService 创建图片审核服务实例。
func NewModerationService(
	db *gorm.DB,
	provider ModerationProvider,
	permission *PermissionService,
	accessAudit *AccessAuditService,
	idempotency *IdempotencyService,
	options ...ModerationServiceOption,
) *ModerationService {
	if provider == nil {
		provider = NewProductionAliyunModerationProvider()
	}
	if permission == nil {
		permission = NewPermissionService(db)
	}
	if accessAudit == nil {
		accessAudit = NewAccessAuditService(db)
	}
	if idempotency == nil {
		idempotency = NewIdempotencyService(db)
	}
	service := &ModerationService{
		DB: db, Provider: provider, Permission: permission,
		AccessAudit: accessAudit, Idempotency: idempotency,
		MutationAudit: gormMutationAuditWriter{}, Now: time.Now,
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
	config := orderfoodModel.ModerationConfig{
		SingletonKey:      moderationConfigSingleton,
		ConfigVersion:     1,
		Provider:          orderfoodModel.ModerationProviderAliyun,
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
	health := orderfoodModel.ModerationConfigHealth{
		SingletonKey:            moderationConfigSingleton,
		Status:                  orderfoodModel.ModerationHealthDisabled,
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
	actor orderfoodRequest.AdminActor,
) (orderfoodResponse.ModerationConfig, error) {
	var result orderfoodResponse.ModerationConfig
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
	input orderfoodRequest.ModerationConfigUpdateInput,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
) (orderfoodResponse.ModerationConfig, bool, error) {
	var result orderfoodResponse.ModerationConfig
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
			current, loadErr := loadCurrentModerationConfigWithLock(tx)
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
			next := orderfoodModel.ModerationConfig{
				SingletonKey:        moderationConfigSingleton,
				ConfigVersion:       current.ConfigVersion + 1,
				Provider:            orderfoodModel.ModerationProviderAliyun,
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
			if err := tx.Save(&next).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "save moderation config")
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
	input orderfoodRequest.ModerationConfigTestInput,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
) (orderfoodResponse.ModerationConnectionTestResult, bool, error) {
	var result orderfoodResponse.ModerationConnectionTestResult
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
			test := orderfoodModel.ModerationConnectionTest{
				ID:               orderfoodModel.NewID(),
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
	input orderfoodRequest.ModerationConfigTestInput,
	image ModerationProviderImage,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
) (orderfoodResponse.ModerationImageTestResult, bool, error) {
	var result orderfoodResponse.ModerationImageTestResult
	if err := service.require(ctx, actor, PermissionModerationConfigTest); err != nil {
		return result, false, err
	}
	if len(image.Bytes) == 0 || service.Idempotency == nil || service.Provider == nil {
		return result, false, appErrors.AdminInvalidImage.DefaultMsg()
	}
	if err := service.ensureDefaultConfig(ctx); err != nil {
		return result, false, err
	}
	image.Scene = orderfoodModel.ModerationSceneDishCover
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
			test := orderfoodModel.ModerationImageTest{
				ID:                   orderfoodModel.NewID(),
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
					SensitiveAccess{
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
	actor orderfoodRequest.AdminActor,
	permission string,
) error {
	if service == nil || service.Permission == nil ||
		actor.AdministratorID == 0 || actor.AuthorityID == 0 {
		return appErrors.AdminLoginExpired.DefaultMsg()
	}
	return service.Permission.Require(ctx, actor.AuthorityID, permission)
}

func validateModerationConfigInput(
	input orderfoodRequest.ModerationConfigUpdateInput,
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
) (orderfoodModel.ModerationConfig, error) {
	var current orderfoodModel.ModerationConfig
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

func loadCurrentModerationConfigWithLock(
	tx *gorm.DB,
) (orderfoodModel.ModerationConfig, error) {
	var current orderfoodModel.ModerationConfig
	if tx == nil {
		return current, appErrors.AdminInternal.DefaultMsg()
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("singleton_key = ?", moderationConfigSingleton).
		First(&current).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return current, appErrors.AdminNotFound.DefaultMsg()
		}
		return current, appErrors.AdminInternal.Wrap(err, "lock current moderation config")
	}
	return current, nil
}

func (service *ModerationService) nextConfigHealth(
	tx *gorm.DB,
	current orderfoodModel.ModerationConfig,
	next orderfoodModel.ModerationConfig,
	now time.Time,
) (orderfoodModel.ModerationConfigHealth, error) {
	status := orderfoodModel.ModerationHealthDegraded
	if !next.Enabled {
		status = orderfoodModel.ModerationHealthDisabled
	}
	health := orderfoodModel.ModerationConfigHealth{
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
	var existing orderfoodModel.ModerationConfigHealth
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
		case orderfoodModel.ModerationHealthHealthy,
			orderfoodModel.ModerationHealthDegraded,
			orderfoodModel.ModerationHealthUnavailable:
			health.Status = existing.Status
		}
	}
	return health, nil
}

func (service *ModerationService) buildConfig(
	ctx context.Context,
	db *gorm.DB,
	config orderfoodModel.ModerationConfig,
) (orderfoodResponse.ModerationConfig, error) {
	health, err := service.buildHealth(ctx, db)
	if err != nil {
		return orderfoodResponse.ModerationConfig{}, err
	}
	return orderfoodResponse.ModerationConfig{
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
) (orderfoodResponse.ModerationConfigHealth, error) {
	result := orderfoodResponse.ModerationConfigHealth{
		Status:    string(orderfoodModel.ModerationHealthUnconfigured),
		UpdatedAt: service.now(),
	}
	var health orderfoodModel.ModerationConfigHealth
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
	test orderfoodModel.ModerationConnectionTest,
	now time.Time,
) error {
	status := orderfoodModel.ModerationHealthHealthy
	if !test.Success {
		status = orderfoodModel.ModerationHealthUnavailable
	}
	configVersion := test.ConfigVersion
	health := orderfoodModel.ModerationConfigHealth{
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
		message := truncateRunes(strings.TrimSpace(test.SafeMessage), 240)
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
) (*orderfoodModel.ModerationConnectionTest, error) {
	var test orderfoodModel.ModerationConnectionTest
	if err := db.WithContext(ctx).Where("id = ?", id).First(&test).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.AdminInternal.Wrap(err, "load moderation connection test")
	}
	return &test, nil
}

func moderationConnectionTestResponse(
	test orderfoodModel.ModerationConnectionTest,
) orderfoodResponse.ModerationConnectionTestResult {
	return orderfoodResponse.ModerationConnectionTestResult{
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
	test orderfoodModel.ModerationImageTest,
	riskLabels []string,
	providerSummary map[string]interface{},
) orderfoodResponse.ModerationImageTestResult {
	if riskLabels == nil {
		riskLabels = []string{}
	}
	return orderfoodResponse.ModerationImageTestResult{
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
	provider orderfoodModel.ModerationProviderType,
	enabled bool,
	region string,
	endpoint string,
	serviceCode string,
	timeoutMS int,
	retryCount int,
	retryBackoffMS int,
	credentialRef string,
	credentialUpdatedAt *time.Time,
) orderfoodResponse.ModerationConfigValues {
	return orderfoodResponse.ModerationConfigValues{
		Provider:        string(provider),
		Enabled:         enabled,
		Region:          region,
		Endpoint:        endpoint,
		ServiceCode:     serviceCode,
		TimeoutMS:       timeoutMS,
		RetryCount:      retryCount,
		RetryBackoffMS:  retryBackoffMS,
		WorstCaseWaitMS: moderationWorstCaseWait(timeoutMS, retryCount, retryBackoffMS),
		Credential: orderfoodResponse.ModerationCredentialStatus{
			Configured: strings.TrimSpace(credentialRef) != "",
			UpdatedAt:  credentialUpdatedAt,
		},
		SafetyPolicy: moderationSafetyPolicy(),
	}
}

func moderationSafetyPolicy() orderfoodResponse.ModerationSafetyPolicy {
	return orderfoodResponse.ModerationSafetyPolicy{
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

func moderationConfigHash(config orderfoodModel.ModerationConfig) string {
	material := struct {
		Provider       orderfoodModel.ModerationProviderType `json:"provider"`
		Enabled        bool                                  `json:"enabled"`
		Region         string                                `json:"region"`
		Endpoint       string                                `json:"endpoint"`
		ServiceCode    string                                `json:"serviceCode"`
		TimeoutMS      int                                   `json:"timeoutMs"`
		RetryCount     int                                   `json:"retryCount"`
		RetryBackoffMS int                                   `json:"retryBackoffMs"`
		CredentialRef  string                                `json:"credentialRef"`
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
	left orderfoodModel.ModerationConfig,
	right orderfoodModel.ModerationConfig,
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

func moderationSceneStrings(input []orderfoodModel.ModerationScene) []string {
	result := make([]string, 0, len(input))
	for _, item := range input {
		result = append(result, string(item))
	}
	return result
}

func moderationProviderConfig(
	version orderfoodModel.ModerationConfig,
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
) orderfoodResponse.AdministratorSummary {
	return orderfoodResponse.AdministratorSummary{
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
	actor orderfoodRequest.AdminActor,
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
	requestID := truncateRunes(strings.TrimSpace(actor.RequestID), 96)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if actor.AdministratorID == 0 || strings.TrimSpace(actor.Username) == "" ||
		requestID == "" || idempotencyKey == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	beforeJSON, err := safeAIAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := safeAIAuditJSON(after)
	if err != nil {
		return err
	}
	var reasonPointer *string
	if value := truncateRunes(strings.TrimSpace(reason), 200); value != "" {
		reasonPointer = &value
	}
	audit := orderfoodModel.AdminAuditLog{
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
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

func moderationConfigAuditSummary(
	version orderfoodModel.ModerationConfig,
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
	test orderfoodModel.ModerationConnectionTest,
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
	test orderfoodModel.ModerationImageTest,
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

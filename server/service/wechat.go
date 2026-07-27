package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	wechatConfigSingletonKey = "current"

	// PermissionWeChatConfigRead 表示查看微信小程序配置所需权限。
	PermissionWeChatConfigRead = "orderfood:wechat-config:read"
	// PermissionWeChatConfigUpdate 表示保存微信小程序配置所需权限。
	PermissionWeChatConfigUpdate = "orderfood:wechat-config:update"
)

var (
	wechatAppIDPattern     = regexp.MustCompile(`^wx[0-9A-Za-z]{16}$`)
	wechatAppSecretPattern = regexp.MustCompile(`^[0-9A-Za-z]{32}$`)
)

// WeChatCredentials 表示仅供服务端调用微信接口使用的明文凭据。
type WeChatCredentials struct {
	AppID     string // 微信小程序AppID
	AppSecret string // 微信小程序AppSecret
}

// WeChatConfigService 提供微信小程序当前配置管理能力。
type WeChatConfigService struct {
	DB            *gorm.DB            // 业务数据库
	Permission    *PermissionService  // GVA按钮权限服务
	Idempotency   *IdempotencyService // 管理端幂等服务
	MutationAudit MutationAuditWriter // 配置变更审计写入器
	Now           func() time.Time    // 可注入的当前时间函数
	Random        io.Reader           // AppSecret加密随机源
}

// NewWeChatConfigService 创建微信小程序配置服务。
func NewWeChatConfigService(
	db *gorm.DB,
	permission *PermissionService,
	idempotency *IdempotencyService,
) *WeChatConfigService {
	return &WeChatConfigService{
		DB:            db,
		Permission:    permission,
		Idempotency:   idempotency,
		MutationAudit: gormMutationAuditWriter{},
		Now:           time.Now,
		Random:        rand.Reader,
	}
}

// database 返回当前服务使用的数据库。
func (service *WeChatConfigService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return nil
}

// now 返回统一使用的UTC当前时间。
func (service *WeChatConfigService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// random 返回AppSecret加密所需的随机源。
func (service *WeChatConfigService) random() io.Reader {
	if service != nil && service.Random != nil {
		return service.Random
	}
	return rand.Reader
}

// validateWeChatActor 校验微信配置管理操作的管理员上下文。
func validateWeChatActor(actor orderfoodRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		actor.AuthorityID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// require 校验管理员是否拥有指定微信配置权限。
func (service *WeChatConfigService) require(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	permission string,
) error {
	if service == nil || service.Permission == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	return service.Permission.Require(ctx, actor.AuthorityID, permission)
}

// LoadIdentityEncryptionKey 从config.yaml读取微信敏感字段共用的32字节服务端加密密钥。
func LoadIdentityEncryptionKey() ([]byte, error) {
	encoded := strings.TrimSpace(global.GVA_CONFIG.OrderFood.IdentityKey)
	if encoded == "" {
		return nil, appErrors.AdminIdentityKeyMissing.New("orderfood.identity-key is empty")
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		if err == nil {
			err = errors.New("identity encryption key must contain 32 bytes")
		}
		return nil, appErrors.AdminIdentityKeyInvalid.Wrap(err, "load identity encryption key from config.yaml")
	}
	return key, nil
}

// encryptWeChatSecret 使用AES-GCM加密微信AppSecret。
func encryptWeChatSecret(secret string, randomSource io.Reader) (string, error) {
	key, err := LoadIdentityEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "create wechat secret cipher")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "create wechat secret gcm")
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(randomSource, nonce); err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "create wechat secret nonce")
	}
	payload := append(nonce, aead.Seal(nil, nonce, []byte(secret), nil)...)
	return base64.RawStdEncoding.EncodeToString(payload), nil
}

// decryptWeChatSecret 解密仅供服务端运行时使用的微信AppSecret。
func decryptWeChatSecret(encoded string) (string, error) {
	key, err := LoadIdentityEncryptionKey()
	if err != nil {
		return "", err
	}
	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", appErrors.AdminInvalidConfig.Wrap(err, "decode wechat secret")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "create wechat secret cipher")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "create wechat secret gcm")
	}
	if len(payload) < aead.NonceSize() {
		return "", appErrors.AdminInvalidConfig.New("encrypted wechat secret is incomplete")
	}
	plaintext, err := aead.Open(
		nil,
		payload[:aead.NonceSize()],
		payload[aead.NonceSize():],
		nil,
	)
	if err != nil {
		return "", appErrors.AdminInvalidConfig.Wrap(err, "decrypt wechat secret")
	}
	return string(plaintext), nil
}

// LoadCurrentWeChatCredentials 读取并解密当前微信服务端凭据。
func LoadCurrentWeChatCredentials(
	ctx context.Context,
	db *gorm.DB,
) (WeChatCredentials, error) {
	if db == nil {
		return WeChatCredentials{}, appErrors.AdminInternal.DefaultMsg()
	}
	var current orderfoodModel.WeChatConfig
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", wechatConfigSingletonKey).
		First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return WeChatCredentials{}, appErrors.AdminInvalidConfig.DefaultMsg()
		}
		return WeChatCredentials{}, appErrors.AdminInternal.Wrap(err, "load current wechat configuration")
	}
	secret, err := decryptWeChatSecret(current.AppSecretEncrypted)
	if err != nil {
		return WeChatCredentials{}, err
	}
	if !wechatAppIDPattern.MatchString(current.AppID) ||
		!wechatAppSecretPattern.MatchString(secret) {
		return WeChatCredentials{}, appErrors.AdminInvalidConfig.DefaultMsg()
	}
	return WeChatCredentials{AppID: current.AppID, AppSecret: secret}, nil
}

// weChatConfigResponse 组装不包含AppSecret明文或密文的安全响应。
func weChatConfigResponse(
	current orderfoodModel.WeChatConfig,
	exists bool,
) orderfoodResponse.WeChatConfig {
	if !exists {
		return orderfoodResponse.WeChatConfig{}
	}
	updatedAt := current.AppliedAt
	administrator := aiAdministratorSummary(
		current.AppliedByID,
		current.AppliedByUsername,
		current.AppliedByNickname,
	)
	return orderfoodResponse.WeChatConfig{
		Configured:          true,
		AppID:               current.AppID,
		AppSecretConfigured: strings.TrimSpace(current.AppSecretEncrypted) != "",
		Version:             current.Version,
		UpdatedBy:           &administrator,
		UpdatedAt:           &updatedAt,
		SecretUpdatedAt:     current.SecretUpdatedAt,
		Reason:              current.Reason,
	}
}

// GetConfig 获取可安全展示的微信小程序当前配置。
func (service *WeChatConfigService) GetConfig(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
) (orderfoodResponse.WeChatConfig, error) {
	if err := validateWeChatActor(actor); err != nil {
		return orderfoodResponse.WeChatConfig{}, err
	}
	if err := service.require(ctx, actor, PermissionWeChatConfigRead); err != nil {
		return orderfoodResponse.WeChatConfig{}, err
	}
	db := service.database()
	if db == nil {
		return orderfoodResponse.WeChatConfig{}, appErrors.AdminInternal.DefaultMsg()
	}
	var current orderfoodModel.WeChatConfig
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", wechatConfigSingletonKey).
		First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodResponse.WeChatConfig{}, nil
		}
		return orderfoodResponse.WeChatConfig{}, appErrors.AdminInternal.Wrap(err, "get wechat configuration")
	}
	return weChatConfigResponse(current, true), nil
}

// validateWeChatConfigInput 校验AppID、AppSecret和版本组合。
func validateWeChatConfigInput(
	input orderfoodRequest.WeChatConfigUpdateInput,
	current orderfoodModel.WeChatConfig,
	exists bool,
) error {
	if !wechatAppIDPattern.MatchString(strings.TrimSpace(input.AppID)) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	reasonLength := len([]rune(strings.TrimSpace(input.Reason)))
	if reasonLength < 2 || reasonLength > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	if (!exists && input.ExpectedVersion != 0) ||
		(exists && current.Version != input.ExpectedVersion) {
		return appErrors.AdminStateConflict.DefaultMsg()
	}
	if input.AppSecret == nil {
		if !exists || current.AppID != strings.TrimSpace(input.AppID) {
			return appErrors.AdminInvalidConfig.DefaultMsg()
		}
		return nil
	}
	if !wechatAppSecretPattern.MatchString(strings.TrimSpace(*input.AppSecret)) {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// writeWeChatConfigAudit 在同一事务内记录微信配置变更，摘要不包含AppSecret。
func (service *WeChatConfigService) writeWeChatConfigAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	reason string,
	before orderfoodResponse.WeChatConfig,
	after orderfoodResponse.WeChatConfig,
) error {
	if service == nil || service.MutationAudit == nil || tx == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	beforeJSON, err := safeAIAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := safeAIAuditJSON(after)
	if err != nil {
		return err
	}
	reasonValue := strings.TrimSpace(reason)
	key := strings.TrimSpace(idempotencyKey)
	audit := orderfoodModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                "update_wechat_config",
		TargetType:            "wechat_config",
		TargetID:              wechatConfigSingletonKey,
		Reason:                &reasonValue,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             truncateRunes(strings.TrimSpace(actor.RequestID), 96),
		IdempotencyKey:        &key,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

// UpdateConfig 保存并立即应用微信小程序配置。
func (service *WeChatConfigService) UpdateConfig(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.WeChatConfigUpdateInput,
) (orderfoodResponse.WeChatConfig, bool, error) {
	if err := validateWeChatActor(actor); err != nil {
		return orderfoodResponse.WeChatConfig{}, false, err
	}
	if err := service.require(ctx, actor, PermissionWeChatConfigUpdate); err != nil {
		return orderfoodResponse.WeChatConfig{}, false, err
	}
	if service == nil || service.Idempotency == nil {
		return orderfoodResponse.WeChatConfig{}, false, appErrors.AdminInternal.DefaultMsg()
	}
	input.AppID = strings.TrimSpace(input.AppID)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.AppSecret != nil {
		secret := strings.TrimSpace(*input.AppSecret)
		input.AppSecret = &secret
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"wechat_config_update",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			var current orderfoodModel.WeChatConfig
			findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("singleton_key = ?", wechatConfigSingletonKey).
				First(&current).Error
			exists := findErr == nil
			if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return nil, appErrors.AdminInternal.Wrap(findErr, "lock wechat configuration")
			}
			if err := validateWeChatConfigInput(input, current, exists); err != nil {
				return nil, err
			}
			before := weChatConfigResponse(current, exists)
			now := service.now()
			next := current
			next.SingletonKey = wechatConfigSingletonKey
			next.Version = current.Version + 1
			next.AppID = input.AppID
			next.AppliedByID = actor.AdministratorID
			next.AppliedByUsername = actor.Username
			next.AppliedByNickname = actor.Nickname
			next.AppliedAt = now
			next.Reason = input.Reason
			if input.AppSecret != nil {
				encrypted, encryptErr := encryptWeChatSecret(*input.AppSecret, service.random())
				if encryptErr != nil {
					return nil, encryptErr
				}
				next.AppSecretEncrypted = encrypted
				next.SecretUpdatedAt = &now
			}
			if err := tx.Save(&next).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "save wechat configuration")
			}
			after := weChatConfigResponse(next, true)
			if err := service.writeWeChatConfigAudit(
				ctx,
				tx,
				actor,
				idempotencyKey,
				input.Reason,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.WeChatConfig](raw, replayed, err)
}

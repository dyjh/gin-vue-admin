package user

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/user"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	miniAppSessionTTL = 7 * 24 * time.Hour
	loginCodeTTL      = 10 * time.Minute
	code2SessionURL   = "https://api.weixin.qq.com/sns/jscode2session"
)

// WeChatSession 表示微信登录凭证兑换得到的服务端会话信息。
type WeChatSession struct {
	OpenID     string // 微信用户OpenID
	UnionID    string // 微信开放平台UnionID
	SessionKey string // 微信会话密钥
}

// WeChatSessionExchanger 定义微信会话兑换器所需的业务能力。
type WeChatSessionExchanger interface {
	Exchange(ctx context.Context, code string) (WeChatSession, error)
}

// HTTPWeChatSessionExchanger 通过微信服务端接口兑换一次性登录凭证。
type HTTPWeChatSessionExchanger struct {
	Client    *http.Client // 可注入的HTTP客户端
	AppID     string       // 微信小程序AppID
	AppSecret string       // 微信小程序AppSecret
	Endpoint  string       // jscode2session接口地址
	DB        *gorm.DB     // 微信配置数据库
}

type code2SessionResponse struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrorCode  int    `json:"errcode"`
}

// Exchange 调用微信 jscode2session 接口并校验返回的身份信息。
func (exchanger HTTPWeChatSessionExchanger) Exchange(
	ctx context.Context,
	code string,
) (WeChatSession, error) {
	appID := strings.TrimSpace(exchanger.AppID)
	appSecret := strings.TrimSpace(exchanger.AppSecret)
	if appID == "" && appSecret == "" {
		db := exchanger.DB
		if db == nil {
			db = global.GVA_DB
		}
		credentials, err := orderfoodService.LoadCurrentWeChatCredentials(ctx, db)
		if err != nil {
			return WeChatSession{}, appErrors.FrontInternal.Wrap(err, "load wechat credentials")
		}
		appID = credentials.AppID
		appSecret = credentials.AppSecret
	}
	if appID == "" || appSecret == "" {
		return WeChatSession{}, appErrors.FrontInternal.New("wechat credentials are incomplete")
	}
	endpoint := strings.TrimSpace(exchanger.Endpoint)
	if endpoint == "" {
		endpoint = code2SessionURL
	}
	query := url.Values{
		"appid":      []string{appID},
		"secret":     []string{appSecret},
		"js_code":    []string{code},
		"grant_type": []string{"authorization_code"},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return WeChatSession{}, appErrors.FrontInternal.Wrap(err, "create code2session request")
	}
	client := exchanger.Client
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	httpResponse, err := client.Do(request)
	if err != nil {
		return WeChatSession{}, appErrors.FrontWxLoginInvalid.Wrap(err, "exchange wx login code")
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(httpResponse.Body, 64<<10))
		return WeChatSession{}, appErrors.FrontWxLoginInvalid.DefaultMsg()
	}
	var result code2SessionResponse
	decoder := json.NewDecoder(io.LimitReader(httpResponse.Body, 64<<10))
	if err := decoder.Decode(&result); err != nil {
		return WeChatSession{}, appErrors.FrontWxLoginInvalid.Wrap(err, "decode code2session response")
	}
	if result.ErrorCode != 0 ||
		strings.TrimSpace(result.OpenID) == "" ||
		strings.TrimSpace(result.SessionKey) == "" {
		return WeChatSession{}, appErrors.FrontWxLoginInvalid.DefaultMsg()
	}
	return WeChatSession{
		OpenID:     result.OpenID,
		UnionID:    result.UnionID,
		SessionKey: result.SessionKey,
	}, nil
}

// LoginResult 表示业务令牌、用户资料和运行时配置组成的登录结果。
type LoginResult struct {
	AccessToken string                // 小程序业务访问令牌
	ExpiresIn   int64                 // 令牌有效秒数
	IsNewUser   bool                  // 是否为首次登录用户
	User        userModel.MiniAppUser // 当前用户
}

// AuthService 提供小程序认证业务能力。
type AuthService struct {
	DB          *gorm.DB               // 业务数据库
	Exchanger   WeChatSessionExchanger // 微信会话兑换器
	IdentityKey []byte                 // 微信身份字段加密密钥
	Now         func() time.Time       // 可注入的当前时间函数
	Random      io.Reader              // 加密与令牌生成随机源
}

func (service *AuthService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *AuthService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

func (service *AuthService) random() io.Reader {
	if service != nil && service.Random != nil {
		return service.Random
	}
	return rand.Reader
}

func (service *AuthService) exchanger() WeChatSessionExchanger {
	if service != nil && service.Exchanger != nil {
		return service.Exchanger
	}
	return HTTPWeChatSessionExchanger{}
}

func (service *AuthService) identityKey() ([]byte, error) {
	if service != nil && len(service.IdentityKey) > 0 {
		if len(service.IdentityKey) != 32 {
			return nil, appErrors.FrontInternal.New("identity key must contain 32 bytes")
		}
		return append([]byte(nil), service.IdentityKey...), nil
	}
	key, err := orderfoodService.LoadIdentityEncryptionKey()
	if err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "load identity encryption key")
	}
	return key, nil
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func encryptIdentity(key []byte, plaintext string, random io.Reader) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(random, nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, sealed...)
	return base64.RawStdEncoding.EncodeToString(payload), nil
}

// decryptIdentity 解密仅保存在服务端的微信身份字段。
func decryptIdentity(key []byte, encoded string) (string, error) {
	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < aead.NonceSize() {
		return "", errors.New("encrypted identity payload is too short")
	}
	plaintext, err := aead.Open(
		nil,
		payload[:aead.NonceSize()],
		payload[aead.NonceSize():],
		nil,
	)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func randomToken(random io.Reader) (string, error) {
	value := make([]byte, 32)
	if _, err := io.ReadFull(random, value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

// Login 兑换微信身份、创建或更新用户，并签发业务访问令牌。
func (service *AuthService) Login(
	ctx context.Context,
	code string,
) (LoginResult, error) {
	db := service.database()
	if db == nil {
		return LoginResult{}, appErrors.FrontInternal.DefaultMsg()
	}
	key, err := service.identityKey()
	if err != nil {
		return LoginResult{}, err
	}
	wechatSession, err := service.exchanger().Exchange(ctx, code)
	if err != nil {
		if appErrors.GetType(err) == appErrors.AdminInternal {
			return LoginResult{}, appErrors.FrontInternal.Wrap(err, "exchange wx login code")
		}
		return LoginResult{}, err
	}
	openIDEncrypted, err := encryptIdentity(key, wechatSession.OpenID, service.random())
	if err != nil {
		return LoginResult{}, appErrors.FrontInternal.Wrap(err, "encrypt openid")
	}
	sessionKeyEncrypted, err := encryptIdentity(key, wechatSession.SessionKey, service.random())
	if err != nil {
		return LoginResult{}, appErrors.FrontInternal.Wrap(err, "encrypt session key")
	}
	var unionIDHash *string
	var unionIDEncrypted *string
	if wechatSession.UnionID != "" {
		hash := digest(wechatSession.UnionID)
		encrypted, encryptErr := encryptIdentity(key, wechatSession.UnionID, service.random())
		if encryptErr != nil {
			return LoginResult{}, appErrors.FrontInternal.Wrap(encryptErr, "encrypt unionid")
		}
		unionIDHash = &hash
		unionIDEncrypted = &encrypted
	}
	accessToken, err := randomToken(service.random())
	if err != nil {
		return LoginResult{}, appErrors.FrontInternal.Wrap(err, "create miniapp access token")
	}

	now := service.now()
	result := LoginResult{
		AccessToken: accessToken,
		ExpiresIn:   int64(miniAppSessionTTL / time.Second),
	}
	transactionErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		codeRecord := userModel.MiniAppLoginCode{
			CodeHash:   digest(code),
			RedeemedAt: now,
			ExpiresAt:  now.Add(loginCodeTTL),
		}
		insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&codeRecord)
		if insert.Error != nil {
			return appErrors.FrontInternal.Wrap(insert.Error, "store redeemed wx login code")
		}
		if insert.RowsAffected != 1 {
			return appErrors.FrontWxLoginInvalid.DefaultMsg()
		}

		openIDHash := digest(wechatSession.OpenID)
		var user userModel.MiniAppUser
		findErr := tx.First(&user, "open_id_hash = ?", openIDHash).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			user = userModel.MiniAppUser{
				ID:                  commonModel.NewID(),
				OpenIDHash:          openIDHash,
				OpenIDEncrypted:     openIDEncrypted,
				UnionIDHash:         unionIDHash,
				UnionIDEncrypted:    unionIDEncrypted,
				SessionKeyEncrypted: &sessionKeyEncrypted,
				Nickname:            "微信用户",
				Status:              userModel.UserStatusNormal,
				RegisteredAt:        now,
				LastLoginAt:         &now,
				CreatedAt:           now,
				UpdatedAt:           now,
			}
			create := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "open_id_hash"}},
				DoNothing: true,
			}).Create(&user)
			if create.Error != nil {
				return appErrors.FrontInternal.Wrap(create.Error, "create miniapp user")
			}
			if create.RowsAffected == 1 {
				result.IsNewUser = true
			} else if err := tx.First(&user, "open_id_hash = ?", openIDHash).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "load concurrent miniapp user")
			}
		} else if findErr != nil {
			return appErrors.FrontInternal.Wrap(findErr, "load miniapp user")
		}
		if user.Status == userModel.UserStatusDisabled {
			return appErrors.FrontUserDisabled.DefaultMsg()
		}
		updates := map[string]interface{}{
			"open_id_encrypted":     openIDEncrypted,
			"session_key_encrypted": sessionKeyEncrypted,
			"last_login_at":         now,
			"updated_at":            now,
		}
		if unionIDHash != nil {
			updates["union_id_hash"] = *unionIDHash
			updates["union_id_encrypted"] = *unionIDEncrypted
		}
		if err := tx.Model(&userModel.MiniAppUser{}).
			Where("id = ?", user.ID).
			Updates(updates).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "update miniapp login identity")
		}
		user.OpenIDEncrypted = openIDEncrypted
		user.SessionKeyEncrypted = &sessionKeyEncrypted
		user.UnionIDHash = unionIDHash
		user.UnionIDEncrypted = unionIDEncrypted
		user.LastLoginAt = &now

		loginSession := userModel.MiniAppSession{
			ID:        commonModel.NewID(),
			TokenHash: digest(accessToken),
			UserID:    user.ID,
			ExpiresAt: now.Add(miniAppSessionTTL),
			CreatedAt: now,
		}
		if err := tx.Create(&loginSession).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create miniapp login session")
		}
		result.User = user
		return nil
	})
	if transactionErr != nil {
		return LoginResult{}, transactionErr
	}
	return result, nil
}

// Authenticate 校验业务令牌并返回当前可用的小程序用户。
func (service *AuthService) Authenticate(
	ctx context.Context,
	accessToken string,
) (userModel.MiniAppUser, error) {
	db := service.database()
	if db == nil || strings.TrimSpace(accessToken) == "" {
		if db == nil {
			return userModel.MiniAppUser{}, appErrors.FrontInternal.DefaultMsg()
		}
		return userModel.MiniAppUser{}, appErrors.FrontLoginExpired.DefaultMsg()
	}
	var loginSession userModel.MiniAppSession
	if err := db.WithContext(ctx).
		First(&loginSession, "token_hash = ?", digest(accessToken)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return userModel.MiniAppUser{}, appErrors.FrontLoginExpired.DefaultMsg()
		}
		return userModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "load miniapp session")
	}
	now := service.now()
	if loginSession.RevokedAt != nil || !loginSession.ExpiresAt.After(now) {
		return userModel.MiniAppUser{}, appErrors.FrontLoginExpired.DefaultMsg()
	}
	var user userModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", loginSession.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return userModel.MiniAppUser{}, appErrors.FrontLoginExpired.DefaultMsg()
		}
		return userModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "load miniapp session user")
	}
	if user.Status == userModel.UserStatusDisabled {
		return userModel.MiniAppUser{}, appErrors.FrontUserDisabled.DefaultMsg()
	}
	if err := db.WithContext(ctx).Model(&userModel.MiniAppSession{}).
		Where("id = ?", loginSession.ID).
		Update("last_used_at", now).Error; err != nil {
		return userModel.MiniAppUser{}, appErrors.FrontInternal.Wrap(err, "touch miniapp session")
	}
	return user, nil
}

package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	wechatAccessTokenURL     = "https://api.weixin.qq.com/cgi-bin/token"
	wechatSubscribeSendURL   = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"
	wechatInvalidTokenCode   = 40014
	wechatExpiredTokenCode   = 42001
	subscribeMaxSendAttempts = 3
)

// WeChatSubscribeRequest 表示一次微信订阅消息投递所需的最小数据。
type WeChatSubscribeRequest struct {
	OpenID     string                       // 接收用户微信OpenID
	TemplateID string                       // 微信订阅模板ID
	Page       string                       // 点击消息后进入的小程序页面
	Data       map[string]map[string]string // 微信模板字段和值
}

// WeChatSubscribeSender 定义微信订阅消息发送器。
type WeChatSubscribeSender interface {
	Send(ctx context.Context, input WeChatSubscribeRequest) error
}

// WeChatDeliveryError 表示可安全落库的微信投递失败信息。
type WeChatDeliveryError struct {
	Code    string // 微信错误码或内部稳定分类
	Summary string // 不含凭据和身份信息的失败摘要
}

// Error 返回安全的微信投递失败说明。
func (deliveryError WeChatDeliveryError) Error() string {
	return deliveryError.Summary
}

// HTTPWeChatSubscribeSender 通过微信服务端接口投递订阅消息。
type HTTPWeChatSubscribeSender struct {
	Client        *http.Client // 可注入的HTTP客户端
	AppID         string       // 小程序AppID
	AppSecret     string       // 小程序AppSecret
	TokenEndpoint string       // 微信AccessToken接口地址
	SendEndpoint  string       // 微信订阅消息发送接口地址
	DB            *gorm.DB     // 微信配置数据库

	mu                    sync.Mutex // AccessToken缓存互斥锁
	credentialFingerprint string     // 当前缓存令牌对应的凭据摘要
	accessToken           string     // 仅保存在进程内存中的AccessToken
	tokenExpiresAt        time.Time  // AccessToken提前一分钟失效的时间
}

type wechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"` // 微信AccessToken
	ExpiresIn   int    `json:"expires_in"`   // AccessToken有效秒数
	ErrorCode   int    `json:"errcode"`      // 微信错误码
	ErrorMsg    string `json:"errmsg"`       // 微信错误信息
}

type wechatSubscribeResponse struct {
	ErrorCode int    `json:"errcode"` // 微信错误码
	ErrorMsg  string `json:"errmsg"`  // 微信错误信息
}

// client 返回带有合理超时的HTTP客户端。
func (sender *HTTPWeChatSubscribeSender) client() *http.Client {
	if sender != nil && sender.Client != nil {
		return sender.Client
	}
	return &http.Client{Timeout: 8 * time.Second}
}

// credentials 返回数据库当前配置或测试注入字段中的微信服务端凭据。
func (sender *HTTPWeChatSubscribeSender) credentials(
	ctx context.Context,
) (string, string, error) {
	appID := strings.TrimSpace(sender.AppID)
	appSecret := strings.TrimSpace(sender.AppSecret)
	if appID == "" && appSecret == "" {
		db := sender.DB
		if db == nil {
			db = global.GVA_DB
		}
		credentials, err := orderfoodService.LoadCurrentWeChatCredentials(ctx, db)
		if err != nil {
			return "", "", WeChatDeliveryError{
				Code: "credentials_unavailable", Summary: "微信服务端凭据不可用",
			}
		}
		appID = credentials.AppID
		appSecret = credentials.AppSecret
	}
	if appID == "" || appSecret == "" {
		return "", "", WeChatDeliveryError{
			Code: "credentials_missing", Summary: "微信服务端凭据未配置",
		}
	}
	return appID, appSecret, nil
}

// token 获取并缓存微信AccessToken，凭据和Token不会写入日志或数据库。
func (sender *HTTPWeChatSubscribeSender) token(ctx context.Context) (string, error) {
	appID, appSecret, err := sender.credentials(ctx)
	if err != nil {
		return "", err
	}
	fingerprintBytes := sha256.Sum256([]byte(appID + "\x00" + appSecret))
	fingerprint := fmt.Sprintf("%x", fingerprintBytes[:])
	sender.mu.Lock()
	defer sender.mu.Unlock()
	now := time.Now().UTC()
	if sender.credentialFingerprint == "" {
		sender.credentialFingerprint = fingerprint
	} else if sender.credentialFingerprint != fingerprint {
		sender.accessToken = ""
		sender.tokenExpiresAt = time.Time{}
		sender.credentialFingerprint = fingerprint
	}
	if sender.accessToken != "" && now.Before(sender.tokenExpiresAt) {
		return sender.accessToken, nil
	}
	endpoint := strings.TrimSpace(sender.TokenEndpoint)
	if endpoint == "" {
		endpoint = wechatAccessTokenURL
	}
	query := url.Values{
		"grant_type": []string{"client_credential"},
		"appid":      []string{appID},
		"secret":     []string{appSecret},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return "", WeChatDeliveryError{Code: "token_request_invalid", Summary: "微信令牌请求创建失败"}
	}
	response, err := sender.client().Do(request)
	if err != nil {
		return "", WeChatDeliveryError{Code: "token_transport_failed", Summary: "微信令牌接口暂时不可用"}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return "", WeChatDeliveryError{
			Code:    "token_http_" + strconv.Itoa(response.StatusCode),
			Summary: "微信令牌接口返回异常",
		}
	}
	var result wechatAccessTokenResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return "", WeChatDeliveryError{Code: "token_response_invalid", Summary: "微信令牌响应无法解析"}
	}
	if result.ErrorCode != 0 || strings.TrimSpace(result.AccessToken) == "" {
		return "", WeChatDeliveryError{
			Code: strconv.Itoa(result.ErrorCode), Summary: "微信令牌接口拒绝请求",
		}
	}
	expiresIn := result.ExpiresIn
	if expiresIn <= 120 {
		expiresIn = 120
	}
	sender.accessToken = result.AccessToken
	sender.tokenExpiresAt = now.Add(time.Duration(expiresIn-60) * time.Second)
	return sender.accessToken, nil
}

// invalidateToken 仅清除本次请求实际使用的旧令牌，避免并发请求误删刚刷新的令牌。
func (sender *HTTPWeChatSubscribeSender) invalidateToken(usedToken string) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.accessToken != usedToken {
		return
	}
	sender.accessToken = ""
	sender.tokenExpiresAt = time.Time{}
}

// sendWithToken 使用指定AccessToken发送一次订阅消息并返回微信业务响应。
func (sender *HTTPWeChatSubscribeSender) sendWithToken(
	ctx context.Context,
	endpoint string,
	accessToken string,
	encoded []byte,
) (wechatSubscribeResponse, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint+"?access_token="+url.QueryEscape(accessToken),
		bytes.NewReader(encoded),
	)
	if err != nil {
		return wechatSubscribeResponse{}, WeChatDeliveryError{
			Code: "send_request_invalid", Summary: "订阅消息请求创建失败",
		}
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := sender.client().Do(request)
	if err != nil {
		return wechatSubscribeResponse{}, WeChatDeliveryError{
			Code: "send_transport_failed", Summary: "微信订阅消息接口暂时不可用",
		}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return wechatSubscribeResponse{}, WeChatDeliveryError{
			Code:    "send_http_" + strconv.Itoa(response.StatusCode),
			Summary: "微信订阅消息接口返回异常",
		}
	}
	var result wechatSubscribeResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return wechatSubscribeResponse{}, WeChatDeliveryError{
			Code: "send_response_invalid", Summary: "微信订阅消息响应无法解析",
		}
	}
	return result, nil
}

// Send 调用微信订阅消息接口；令牌失效时刷新一次，并只返回可安全持久化的错误分类。
func (sender *HTTPWeChatSubscribeSender) Send(
	ctx context.Context,
	input WeChatSubscribeRequest,
) error {
	endpoint := strings.TrimSpace(sender.SendEndpoint)
	if endpoint == "" {
		endpoint = wechatSubscribeSendURL
	}
	payload := map[string]interface{}{
		"touser": input.OpenID, "template_id": input.TemplateID,
		"page": input.Page, "data": input.Data,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return WeChatDeliveryError{Code: "payload_invalid", Summary: "订阅消息载荷无法编码"}
	}
	for attempt := 0; attempt < 2; attempt++ {
		accessToken, tokenErr := sender.token(ctx)
		if tokenErr != nil {
			return tokenErr
		}
		result, sendErr := sender.sendWithToken(ctx, endpoint, accessToken, encoded)
		if sendErr != nil {
			return sendErr
		}
		if result.ErrorCode == 0 {
			return nil
		}
		if attempt == 0 &&
			(result.ErrorCode == wechatInvalidTokenCode ||
				result.ErrorCode == wechatExpiredTokenCode) {
			// 微信可能在本地缓存到期前提前使令牌失效；只刷新并补发一次。
			sender.invalidateToken(accessToken)
			continue
		}
		return WeChatDeliveryError{
			Code: strconv.Itoa(result.ErrorCode), Summary: "微信订阅消息发送失败",
		}
	}
	return WeChatDeliveryError{Code: "send_failed", Summary: "微信订阅消息发送失败"}
}

// SubscriptionDeliveryService 提供微信订阅消息待发送记录的后台投递能力。
type SubscriptionDeliveryService struct {
	DB     *gorm.DB              // 业务数据库
	Sender WeChatSubscribeSender // 微信订阅消息发送器
	Now    func() time.Time      // 可注入的当前时间
}

// database 返回当前服务使用的数据库。
func (service *SubscriptionDeliveryService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// sender 返回注入发送器或进程级默认微信发送器。
func (service *SubscriptionDeliveryService) sender() WeChatSubscribeSender {
	if service != nil && service.Sender != nil {
		return service.Sender
	}
	return defaultWeChatSubscribeSender
}

// now 返回统一使用的UTC当前时间。
func (service *SubscriptionDeliveryService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

var defaultWeChatSubscribeSender = &HTTPWeChatSubscribeSender{}

// decodeMealSubscriptionPayload 解码饭局结果发送记录中的安全业务摘要。
func decodeMealSubscriptionPayload(raw datatypes.JSON) (map[string]string, error) {
	var payload map[string]string
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "decode meal subscription payload")
	}
	for _, key := range []string{"mealName", "result", "resultAt"} {
		if strings.TrimSpace(payload[key]) == "" {
			return nil, appErrors.FrontInternal.New("meal subscription payload is incomplete")
		}
	}
	switch payload["result"] {
	case string(orderfoodModel.MealConfirmed):
		payload["result"] = "菜单已确认"
	case string(orderfoodModel.MealCancelled):
		payload["result"] = "饭局已取消"
	}
	if resultAt, err := time.Parse(time.RFC3339, payload["resultAt"]); err == nil {
		location, locationErr := time.LoadLocation("Asia/Shanghai")
		if locationErr != nil {
			location = time.FixedZone("Asia/Shanghai", 8*60*60)
		}
		payload["resultAt"] = resultAt.In(location).Format("2006-01-02 15:04")
	}
	return payload, nil
}

// subscriptionTemplateData 根据后台字段映射生成微信模板字段载荷。
func subscriptionTemplateData(
	template orderfoodModel.SubscribeMessageTemplate,
	payload map[string]string,
) (map[string]map[string]string, error) {
	var mappings map[string]string
	if err := json.Unmarshal(template.FieldMappings, &mappings); err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "decode subscription field mappings")
	}
	result := make(map[string]map[string]string, len(mappings))
	for businessField, wechatField := range mappings {
		value := strings.TrimSpace(payload[businessField])
		if value == "" {
			continue
		}
		result[strings.TrimSpace(wechatField)] = map[string]string{"value": value}
	}
	if len(result) < 3 {
		return nil, appErrors.FrontInternal.New("subscription field mappings are incomplete")
	}
	return result, nil
}

// deliveryErrorFields 提取可安全落库的微信错误码和摘要。
func deliveryErrorFields(err error) (*string, *string) {
	var deliveryError WeChatDeliveryError
	if errors.As(err, &deliveryError) {
		code := deliveryError.Code
		summary := deliveryError.Summary
		return &code, &summary
	}
	code := "internal_delivery_failed"
	summary := "订阅消息投递失败"
	return &code, &summary
}

// deliverLog 在数据库行锁内完成单条消息投递和尝试记录，进程异常时事务会回滚为待发送。
func (service *SubscriptionDeliveryService) deliverLog(
	ctx context.Context,
	logID string,
) error {
	return service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row orderfoodModel.SubscribeMessageLog
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&row, "id = ?", logID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "lock subscription send log")
		}
		if row.Status != orderfoodModel.SubscribeLogPending {
			return nil
		}
		var attemptCount int64
		if err := tx.Model(&orderfoodModel.SubscribeMessageAttempt{}).
			Where("log_id = ?", row.ID).Count(&attemptCount).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "count subscription send attempts")
		}
		if attemptCount >= subscribeMaxSendAttempts {
			return tx.Model(&row).Update("status", orderfoodModel.SubscribeLogFailed).Error
		}
		var user orderfoodModel.MiniAppUser
		if err := tx.First(&user, "id = ?", row.UserID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "load subscription recipient")
		}
		var template orderfoodModel.SubscribeMessageTemplate
		if err := tx.First(&template, "id = ?", row.TemplateID).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "load subscription template")
		}
		key, err := (&AuthService{}).identityKey()
		if err != nil {
			return err
		}
		openID, err := decryptIdentity(key, user.OpenIDEncrypted)
		if err != nil {
			return appErrors.FrontInternal.Wrap(err, "decrypt subscription recipient")
		}
		payload, err := decodeMealSubscriptionPayload(row.SafePayloadSummary)
		if err != nil {
			return err
		}
		data, err := subscriptionTemplateData(template, payload)
		if err != nil {
			return err
		}
		page := ""
		if row.TargetPage != nil {
			page = *row.TargetPage
		}
		startedAt := service.now()
		sendErr := service.sender().Send(ctx, WeChatSubscribeRequest{
			OpenID: openID, TemplateID: template.WechatTemplateID, Page: page, Data: data,
		})
		finishedAt := service.now()
		attempt := int(attemptCount) + 1
		attemptStatus := orderfoodModel.SubscribeLogSent
		if sendErr != nil {
			attemptStatus = orderfoodModel.SubscribeLogFailed
		}
		errorCode, errorSummary := deliveryErrorFields(sendErr)
		if sendErr == nil {
			errorCode = nil
			errorSummary = nil
		}
		if err := tx.Create(&orderfoodModel.SubscribeMessageAttempt{
			LogID: row.ID, Attempt: attempt, Status: attemptStatus,
			WechatErrorCode: errorCode, ErrorSummary: errorSummary,
			StartedAt: startedAt, FinishedAt: finishedAt,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "create subscription send attempt")
		}
		retryCount := attempt - 1
		if sendErr == nil {
			if err := tx.Model(&row).Updates(map[string]interface{}{
				"status": orderfoodModel.SubscribeLogSent, "retry_count": retryCount,
				"wechat_error_code": nil, "error_summary": nil, "sent_at": finishedAt,
				"authorization_available": false,
			}).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "complete subscription send log")
			}
			if err := tx.Model(&template).
				Update("send_count", gorm.Expr("send_count + 1")).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "increment subscription send count")
			}
			return nil
		}
		status := orderfoodModel.SubscribeLogPending
		if attempt >= subscribeMaxSendAttempts {
			status = orderfoodModel.SubscribeLogFailed
		}
		if err := tx.Model(&row).Updates(map[string]interface{}{
			"status": status, "retry_count": retryCount,
			"wechat_error_code": errorCode, "error_summary": errorSummary,
			"authorization_available": status != orderfoodModel.SubscribeLogFailed,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "record subscription send failure")
		}
		if status == orderfoodModel.SubscribeLogFailed {
			if err := tx.Model(&template).
				Update("failure_count", gorm.Expr("failure_count + 1")).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "increment subscription failure count")
			}
		}
		return nil
	})
}

// ProcessPendingMessages 分批投递待发送微信订阅消息，单条失败不阻断同批其他记录。
func (service *SubscriptionDeliveryService) ProcessPendingMessages(
	ctx context.Context,
	batchSize int,
) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	var logIDs []string
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.SubscribeMessageLog{}).
		Where("status = ?", orderfoodModel.SubscribeLogPending).
		Order("created_at asc, id asc").Limit(batchSize).Pluck("id", &logIDs).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "list pending subscription messages")
	}
	var firstErr error
	for _, logID := range logIDs {
		if err := service.deliverLog(ctx, logID); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// String 返回适合诊断的发送器描述，不包含任何凭据。
func (sender *HTTPWeChatSubscribeSender) String() string {
	return fmt.Sprintf("wechat subscription sender tokenEndpoint=%t sendEndpoint=%t",
		strings.TrimSpace(sender.TokenEndpoint) != "",
		strings.TrimSpace(sender.SendEndpoint) != "",
	)
}

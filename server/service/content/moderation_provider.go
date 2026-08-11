package content

import (
	"context"
	"errors"
	"strings"
	"time"

	contentModel "github.com/dyjh/order-food-mini-app/server/model/content"
)

const (
	ModerationCategoryOK                   = "ok"
	ModerationCategoryPassed               = "passed"
	ModerationCategoryContentRisk          = "content_risk"
	ModerationCategoryAuthenticationFailed = "authentication_failed"
	ModerationCategoryRateLimited          = "rate_limited"
	ModerationCategoryTimeout              = "timeout"
	ModerationCategoryUnavailable          = "unavailable"
	ModerationCategoryInvalidResponse      = "invalid_response"
)

// ModerationProviderConfig is an internal runtime snapshot. CredentialRef is
// intentionally absent from all API response DTOs.
type ModerationProviderConfig struct {
	Provider       contentModel.ModerationProviderType // 供应商类型
	Region         string                              // 接入地域
	Endpoint       string                              // 接入地址
	ServiceCode    string                              // 图片审核服务编码
	TimeoutMS      int                                 // 单次调用超时时间（毫秒）
	RetryCount     int                                 // 失败重试次数
	RetryBackoffMS int                                 // 重试间隔（毫秒）
	CredentialRef  string                              // 只在服务端解析的凭据引用
	ConfigVersion  int64                               // 当前配置版本
}

// ModerationProviderImage 表示提交给图片审核供应商的图片及业务场景。
type ModerationProviderImage struct {
	Bytes       []byte                       // 图片字节
	ContentType string                       // 图片 MIME 类型
	Scene       contentModel.ModerationScene // 业务审核场景
}

// ModerationProviderResult 表示供应商审核结果归一化后的内部数据。
type ModerationProviderResult struct {
	Status                  contentModel.ModerationStatus // 平台审核状态
	Category                string                        // 稳定结果类别
	RiskLabels              []string                      // 供应商风险标签
	RiskLevel               *string                       // 可选风险等级
	ProviderRequestID       *string                       // 供应商请求 ID
	DurationMS              int                           // 调用耗时（毫秒）
	SafeMessage             string                        // 可展示的安全描述
	ProviderResponseSummary map[string]interface{}        // 脱敏后的供应商摘要
}

// ModerationProvider 定义图片审核供应商所需的业务能力。
type ModerationProvider interface {
	TestConnection(context.Context, ModerationProviderConfig) ModerationProviderResult
	ModerateImage(context.Context, ModerationProviderConfig, ModerationProviderImage) ModerationProviderResult
}

// AliyunCredential is deliberately write-only. It must never be marshalled,
// logged or included in an error.
type AliyunCredential struct {
	AccessKeyID     string `json:"-"` // RAM 用户 AccessKey ID
	AccessKeySecret string `json:"-"` // RAM 用户 AccessKey Secret
	SecurityToken   string `json:"-"` // 可选临时安全令牌
}

// AliyunCredentialResolver keeps secret storage outside this package. A
// production implementation may resolve a KMS/Vault/config-center reference;
// the database stores only that opaque reference.
type AliyunCredentialResolver interface {
	ResolveAliyunCredential(context.Context, string) (AliyunCredential, error)
}

// AliyunImageAuditRequest 表示调用阿里云图片审核适配器的请求。
type AliyunImageAuditRequest struct {
	Region      string           // 接入地域
	Endpoint    string           // 接入地址
	ServiceCode string           // 图片审核服务编码
	Credential  AliyunCredential // 本次调用凭据
	Image       []byte           // 图片字节
	ContentType string           // 图片 MIME 类型
	Scene       string           // 业务审核场景
}

// AliyunImageAuditResponse 表示阿里云图片审核适配器的归一化响应。
type AliyunImageAuditResponse struct {
	Suggestion        string                 // 归一化建议：pass、review 或 block
	RiskLabels        []string               // 风险标签
	RiskLevel         *string                // 可选风险等级
	ProviderRequestID string                 // 阿里云请求 ID
	SafeSummary       map[string]interface{} // 可供后台查看的脱敏摘要
}

// AliyunImageAuditClient is the narrow boundary around the official Aliyun
// SDK. The concrete SDK adapter owns request signing and must return only a
// sanitized SafeSummary.
type AliyunImageAuditClient interface {
	TestConnection(context.Context, AliyunImageAuditRequest) (string, error)
	ModerateImage(context.Context, AliyunImageAuditRequest) (AliyunImageAuditResponse, error)
}

// AliyunProviderError 表示可安全映射为业务错误的阿里云调用失败。
type AliyunProviderError struct {
	Category    string // 平台错误类别
	SafeMessage string // 可展示的安全描述
	Cause       error  // 仅用于内部错误链的原始错误
}

// Error 返回供应商错误的安全描述。
func (err *AliyunProviderError) Error() string {
	if err == nil || strings.TrimSpace(err.SafeMessage) == "" {
		return "aliyun moderation provider failed"
	}
	return err.SafeMessage
}

// Unwrap 返回可用于错误链判断的原始错误。
func (err *AliyunProviderError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

// AliyunModerationProvider 实现阿里云图片审核供应商适配。
type AliyunModerationProvider struct {
	Credentials AliyunCredentialResolver // 凭据引用解析器
	Client      AliyunImageAuditClient   // 阿里云官方 SDK 适配器
}

var _ ModerationProvider = (*AliyunModerationProvider)(nil)

// NewAliyunModerationProvider 创建阿里云图片审核供应商实例。
func NewAliyunModerationProvider(
	resolver AliyunCredentialResolver,
	client AliyunImageAuditClient,
) *AliyunModerationProvider {
	return &AliyunModerationProvider{
		Credentials: resolver,
		Client:      client,
	}
}

// TestConnection 测试连接。
func (provider *AliyunModerationProvider) TestConnection(
	ctx context.Context,
	config ModerationProviderConfig,
) ModerationProviderResult {
	startedAt := time.Now()
	if config.Provider != contentModel.ModerationProviderAliyun {
		return moderationFailedProviderResult(
			ModerationCategoryInvalidResponse,
			"不支持的图片审核供应商",
			moderationElapsedMilliseconds(startedAt),
		)
	}
	credential, failure := provider.resolveCredential(ctx, config.CredentialRef)
	if failure != nil {
		failure.DurationMS = moderationElapsedMilliseconds(startedAt)
		return *failure
	}
	if provider == nil || provider.Client == nil {
		return moderationFailedProviderResult(
			ModerationCategoryUnavailable,
			"图片审核服务暂不可用",
			moderationElapsedMilliseconds(startedAt),
		)
	}

	request := AliyunImageAuditRequest{
		Region: config.Region, Endpoint: config.Endpoint,
		ServiceCode: config.ServiceCode, Credential: credential,
	}
	var providerRequestID string
	callResult := provider.callWithRetry(ctx, config, func(callCtx context.Context) error {
		requestID, err := provider.Client.TestConnection(callCtx, request)
		if err == nil {
			providerRequestID = strings.TrimSpace(requestID)
		}
		return err
	})
	callResult.DurationMS = moderationElapsedMilliseconds(startedAt)
	if callResult.Category != ModerationCategoryOK {
		return callResult
	}
	if providerRequestID != "" {
		callResult.ProviderRequestID = &providerRequestID
	}
	callResult.Status = contentModel.ModerationStatusPassed
	callResult.SafeMessage = "连接测试通过"
	return callResult
}

// ModerateImage 审核图片。
func (provider *AliyunModerationProvider) ModerateImage(
	ctx context.Context,
	config ModerationProviderConfig,
	image ModerationProviderImage,
) ModerationProviderResult {
	startedAt := time.Now()
	if config.Provider != contentModel.ModerationProviderAliyun {
		return moderationFailedProviderResult(
			ModerationCategoryInvalidResponse,
			"不支持的图片审核供应商",
			moderationElapsedMilliseconds(startedAt),
		)
	}
	credential, failure := provider.resolveCredential(ctx, config.CredentialRef)
	if failure != nil {
		failure.DurationMS = moderationElapsedMilliseconds(startedAt)
		return *failure
	}
	if provider == nil || provider.Client == nil {
		return moderationFailedProviderResult(
			ModerationCategoryUnavailable,
			"图片审核服务暂不可用",
			moderationElapsedMilliseconds(startedAt),
		)
	}

	request := AliyunImageAuditRequest{
		Region: config.Region, Endpoint: config.Endpoint,
		ServiceCode: config.ServiceCode, Credential: credential,
		Image: image.Bytes, ContentType: image.ContentType, Scene: string(image.Scene),
	}
	var raw AliyunImageAuditResponse
	callResult := provider.callWithRetry(ctx, config, func(callCtx context.Context) error {
		var err error
		raw, err = provider.Client.ModerateImage(callCtx, request)
		return err
	})
	callResult.DurationMS = moderationElapsedMilliseconds(startedAt)
	if callResult.Category != ModerationCategoryOK {
		return callResult
	}

	requestID := strings.TrimSpace(raw.ProviderRequestID)
	if requestID != "" {
		callResult.ProviderRequestID = &requestID
	}
	callResult.RiskLabels = append([]string(nil), raw.RiskLabels...)
	callResult.RiskLevel = raw.RiskLevel
	callResult.ProviderResponseSummary = moderationSanitizeProviderSummary(raw.SafeSummary)
	switch strings.ToLower(strings.TrimSpace(raw.Suggestion)) {
	case "pass":
		callResult.Status = contentModel.ModerationStatusPassed
		callResult.Category = ModerationCategoryPassed
		callResult.SafeMessage = "图片审核通过"
	case "review", "block":
		callResult.Status = contentModel.ModerationStatusRejected
		callResult.Category = ModerationCategoryContentRisk
		callResult.SafeMessage = "图片内容审核未通过"
	default:
		callResult.Status = contentModel.ModerationStatusFailed
		callResult.Category = ModerationCategoryInvalidResponse
		callResult.SafeMessage = "图片审核服务返回无效结果"
	}
	return callResult
}

// resolveCredential 解析并校验本次阿里云调用所需的凭据。
func (provider *AliyunModerationProvider) resolveCredential(
	ctx context.Context,
	reference string,
) (AliyunCredential, *ModerationProviderResult) {
	if provider == nil || provider.Credentials == nil || strings.TrimSpace(reference) == "" {
		failure := moderationFailedProviderResult(
			ModerationCategoryAuthenticationFailed,
			"图片审核凭据未配置或不可用",
			0,
		)
		return AliyunCredential{}, &failure
	}
	credential, err := provider.Credentials.ResolveAliyunCredential(ctx, reference)
	if err != nil ||
		strings.TrimSpace(credential.AccessKeyID) == "" ||
		strings.TrimSpace(credential.AccessKeySecret) == "" {
		failure := moderationFailedProviderResult(
			ModerationCategoryAuthenticationFailed,
			"图片审核凭据未配置或不可用",
			0,
		)
		return AliyunCredential{}, &failure
	}
	return credential, nil
}

// callWithRetry 按当前配置执行限次重试，并归一化每次调用错误。
func (provider *AliyunModerationProvider) callWithRetry(
	ctx context.Context,
	config ModerationProviderConfig,
	call func(context.Context) error,
) ModerationProviderResult {
	attempts := config.RetryCount + 1
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		timeout := time.Duration(config.TimeoutMS) * time.Millisecond
		if timeout <= 0 {
			timeout = time.Second
		}
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		err := call(callCtx)
		callContextErr := callCtx.Err()
		cancel()
		if err == nil {
			return ModerationProviderResult{
				Status:     contentModel.ModerationStatusPassed,
				Category:   ModerationCategoryOK,
				RiskLabels: []string{},
			}
		}

		category, safeMessage := classifyAliyunProviderError(err, callContextErr)
		if attempt+1 >= attempts || !retryableModerationCategory(category) {
			return moderationFailedProviderResult(category, safeMessage, 0)
		}
		backoff := time.Duration(config.RetryBackoffMS) * time.Millisecond
		if backoff <= 0 {
			continue
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return moderationFailedProviderResult(
				ModerationCategoryTimeout,
				"图片审核服务调用超时",
				0,
			)
		case <-timer.C:
		}
	}
	return moderationFailedProviderResult(
		ModerationCategoryUnavailable,
		"图片审核服务暂不可用",
		0,
	)
}

// classifyAliyunProviderError 将内部适配器错误转换为稳定业务类别和安全描述。
func classifyAliyunProviderError(err error, callContextErr error) (string, string) {
	if errors.Is(callContextErr, context.DeadlineExceeded) ||
		errors.Is(err, context.DeadlineExceeded) {
		return ModerationCategoryTimeout, "图片审核服务调用超时"
	}
	var providerErr *AliyunProviderError
	if errors.As(err, &providerErr) {
		switch providerErr.Category {
		case ModerationCategoryAuthenticationFailed:
			return providerErr.Category, "图片审核凭据无效"
		case ModerationCategoryRateLimited:
			return providerErr.Category, "图片审核服务繁忙"
		case ModerationCategoryTimeout:
			return providerErr.Category, "图片审核服务调用超时"
		case ModerationCategoryInvalidResponse:
			return providerErr.Category, "图片审核服务返回无效结果"
		default:
			return ModerationCategoryUnavailable, "图片审核服务暂不可用"
		}
	}
	return ModerationCategoryUnavailable, "图片审核服务暂不可用"
}

// retryableModerationCategory 判断供应商错误是否允许重试。
func retryableModerationCategory(category string) bool {
	return category == ModerationCategoryRateLimited ||
		category == ModerationCategoryTimeout ||
		category == ModerationCategoryUnavailable
}

// moderationFailedProviderResult 构造统一的供应商失败结果。
func moderationFailedProviderResult(category string, safeMessage string, durationMS int) ModerationProviderResult {
	return ModerationProviderResult{
		Status:      contentModel.ModerationStatusFailed,
		Category:    category,
		RiskLabels:  []string{},
		DurationMS:  durationMS,
		SafeMessage: safeMessage,
	}
}

// moderationElapsedMilliseconds 计算非负的毫秒耗时。
func moderationElapsedMilliseconds(start time.Time) int {
	elapsed := time.Since(start).Milliseconds()
	if elapsed < 0 {
		return 0
	}
	return int(elapsed)
}

// moderationSanitizeProviderSummary accepts only shallow scalar values and scalar
// arrays. Credentials, tokens, headers and full provider payloads must never
// cross this boundary.
func moderationSanitizeProviderSummary(input map[string]interface{}) map[string]interface{} {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]interface{}, len(input))
	for key, value := range input {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "" ||
			strings.Contains(normalizedKey, "credential") ||
			strings.Contains(normalizedKey, "accesskey") ||
			strings.Contains(normalizedKey, "secret") ||
			strings.Contains(normalizedKey, "token") ||
			strings.Contains(normalizedKey, "authorization") {
			continue
		}
		switch typed := value.(type) {
		case string:
			if len(typed) <= 240 {
				output[key] = typed
			}
		case bool, int, int32, int64, float32, float64:
			output[key] = typed
		case []string:
			if len(typed) <= 20 {
				output[key] = append([]string(nil), typed...)
			}
		}
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

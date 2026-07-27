package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	green "github.com/alibabacloud-go/green-20220302/v3/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

const (
	aliyunModerationNoRiskLabel = "nonlabel"
	aliyunModerationCleanupWait = time.Second
)

// EnvironmentAliyunCredentialResolver 从环境变量读取阿里云图片审核凭据。
type EnvironmentAliyunCredentialResolver struct{}

// aliyunEnvironmentCredentialValue 表示环境变量内保存的阿里云凭据 JSON。
type aliyunEnvironmentCredentialValue struct {
	AccessKeyID     string `json:"accessKeyId"`     // RAM 用户 AccessKey ID
	AccessKeySecret string `json:"accessKeySecret"` // RAM 用户 AccessKey Secret
	SecurityToken   string `json:"securityToken"`   // 可选的临时安全令牌
}

// ResolveAliyunCredential 解析 env://变量名，并且只返回进程内使用的凭据。
func (EnvironmentAliyunCredentialResolver) ResolveAliyunCredential(
	ctx context.Context,
	reference string,
) (AliyunCredential, error) {
	var credential AliyunCredential
	if err := ctx.Err(); err != nil {
		return credential, err
	}
	reference = strings.TrimSpace(reference)
	if !strings.HasPrefix(reference, "env://") {
		return credential, errors.New("unsupported aliyun credential reference")
	}
	name := strings.TrimSpace(strings.TrimPrefix(reference, "env://"))
	if name == "" {
		return credential, errors.New("empty aliyun credential reference")
	}
	raw, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(raw) == "" {
		return credential, errors.New("aliyun credential reference is unavailable")
	}
	var value aliyunEnvironmentCredentialValue
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return credential, errors.New("aliyun credential reference is invalid")
	}
	credential = AliyunCredential{
		AccessKeyID:     strings.TrimSpace(value.AccessKeyID),
		AccessKeySecret: strings.TrimSpace(value.AccessKeySecret),
		SecurityToken:   strings.TrimSpace(value.SecurityToken),
	}
	if credential.AccessKeyID == "" || credential.AccessKeySecret == "" {
		return AliyunCredential{}, errors.New("aliyun credential reference is incomplete")
	}
	return credential, nil
}

// AliyunSDKImageAuditClient 使用阿里云官方 SDK 调用图片审核增强版。
type AliyunSDKImageAuditClient struct{}

var _ AliyunImageAuditClient = (*AliyunSDKImageAuditClient)(nil)

// NewProductionAliyunModerationProvider 创建可执行真实阿里云调用的生产供应商。
func NewProductionAliyunModerationProvider() *AliyunModerationProvider {
	return NewAliyunModerationProvider(
		EnvironmentAliyunCredentialResolver{},
		&AliyunSDKImageAuditClient{},
	)
}

// TestConnection 通过获取临时上传令牌验证接入地址、凭据和图片审核权限。
func (*AliyunSDKImageAuditClient) TestConnection(
	ctx context.Context,
	request AliyunImageAuditRequest,
) (string, error) {
	client, runtime, err := newAliyunGreenClient(ctx, request)
	if err != nil {
		return "", err
	}
	response, err := client.DescribeUploadTokenWithOptions(runtime)
	if err != nil {
		return "", normalizeAliyunCallError(err)
	}
	token, requestID, err := parseAliyunUploadToken(response)
	if err != nil {
		return "", err
	}
	if token.BucketName == "" || token.OSSInternetEndpoint == "" {
		return "", aliyunInvalidResponseError(nil)
	}
	return requestID, nil
}

// ModerateImage 上传本地图片到阿里云临时 OSS，并调用同步图片审核接口。
func (*AliyunSDKImageAuditClient) ModerateImage(
	ctx context.Context,
	request AliyunImageAuditRequest,
) (AliyunImageAuditResponse, error) {
	var result AliyunImageAuditResponse
	if len(request.Image) == 0 {
		return result, aliyunInvalidResponseError(nil)
	}
	client, runtime, err := newAliyunGreenClient(ctx, request)
	if err != nil {
		return result, err
	}
	tokenResponse, err := client.DescribeUploadTokenWithOptions(runtime)
	if err != nil {
		return result, normalizeAliyunCallError(err)
	}
	token, _, err := parseAliyunUploadToken(tokenResponse)
	if err != nil {
		return result, err
	}
	objectName, cleanup, err := uploadAliyunModerationImage(ctx, token, request)
	if err != nil {
		return result, err
	}
	defer cleanup()

	// 审核接口通过临时 OSS 定位本次图片，dataId 只用于单次请求追踪。
	serviceParameters, err := json.Marshal(map[string]string{
		"ossBucketName": token.BucketName,
		"ossObjectName": objectName,
		"dataId":        uuid.NewString(),
	})
	if err != nil {
		return result, aliyunInvalidResponseError(err)
	}
	serviceCode := strings.TrimSpace(request.ServiceCode)
	imageRequest := &green.ImageModerationRequest{
		Service:           &serviceCode,
		ServiceParameters: aliyunStringPointer(string(serviceParameters)),
	}
	response, err := client.ImageModerationWithContext(ctx, imageRequest, runtime)
	if err != nil {
		return result, normalizeAliyunCallError(err)
	}
	return mapAliyunModerationResponse(response)
}

// aliyunTemporaryUploadToken 保存阿里云临时 OSS 上传所需的短期凭据。
type aliyunTemporaryUploadToken struct {
	AccessKeyID         string // 临时 AccessKey ID
	AccessKeySecret     string // 临时 AccessKey Secret
	SecurityToken       string // 临时安全令牌
	BucketName          string // 临时存储桶名称
	FileNamePrefix      string // 服务端限定的对象名前缀
	OSSInternetEndpoint string // OSS 公网接入地址
}

// newAliyunGreenClient 创建绑定本次凭据、地域、地址和超时的官方 SDK 客户端。
func newAliyunGreenClient(
	ctx context.Context,
	request AliyunImageAuditRequest,
) (*green.Client, *dara.RuntimeOptions, error) {
	endpoint, err := normalizeAliyunModerationEndpoint(request.Endpoint)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(request.Credential.AccessKeyID) == "" ||
		strings.TrimSpace(request.Credential.AccessKeySecret) == "" {
		return nil, nil, aliyunAuthenticationError(nil)
	}
	timeoutMS := aliyunRequestTimeoutMilliseconds(ctx)
	region := strings.TrimSpace(request.Region)
	protocol := "HTTPS"
	config := &openapi.Config{
		AccessKeyId:     aliyunStringPointer(strings.TrimSpace(request.Credential.AccessKeyID)),
		AccessKeySecret: aliyunStringPointer(strings.TrimSpace(request.Credential.AccessKeySecret)),
		Endpoint:        aliyunStringPointer(endpoint),
		RegionId:        aliyunStringPointer(region),
		Protocol:        &protocol,
		ReadTimeout:     &timeoutMS,
		ConnectTimeout:  &timeoutMS,
	}
	if token := strings.TrimSpace(request.Credential.SecurityToken); token != "" {
		config.SecurityToken = &token
	}
	client, err := green.NewClient(config)
	if err != nil {
		return nil, nil, normalizeAliyunCallError(err)
	}
	autoRetry := false
	runtime := &dara.RuntimeOptions{
		Autoretry:      &autoRetry,
		ReadTimeout:    &timeoutMS,
		ConnectTimeout: &timeoutMS,
	}
	return client, runtime, nil
}

// normalizeAliyunModerationEndpoint 将后台保存的 HTTPS URL 转为 SDK 所需的主机地址。
func normalizeAliyunModerationEndpoint(raw string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil ||
		!strings.EqualFold(parsed.Scheme, "https") ||
		parsed.Host == "" ||
		parsed.User != nil ||
		(parsed.Path != "" && parsed.Path != "/") ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return "", aliyunInvalidResponseError(err)
	}
	return parsed.Host, nil
}

// aliyunRequestTimeoutMilliseconds 从业务上下文计算本次 SDK 调用超时。
func aliyunRequestTimeoutMilliseconds(ctx context.Context) int {
	const defaultTimeoutMS = 10000
	deadline, ok := ctx.Deadline()
	if !ok {
		return defaultTimeoutMS
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return 1
	}
	timeoutMS := int(remaining.Milliseconds())
	if timeoutMS < 1 {
		return 1
	}
	if timeoutMS > defaultTimeoutMS {
		return defaultTimeoutMS
	}
	return timeoutMS
}

// parseAliyunUploadToken 校验并提取临时 OSS 上传令牌，避免任何令牌进入摘要或日志。
func parseAliyunUploadToken(
	response *green.DescribeUploadTokenResponse,
) (aliyunTemporaryUploadToken, string, error) {
	var token aliyunTemporaryUploadToken
	if response == nil || response.StatusCode == nil || *response.StatusCode != 200 ||
		response.Body == nil || response.Body.Code == nil || *response.Body.Code != 200 {
		statusCode, bodyCode := aliyunUploadTokenResponseCodes(response)
		return token, "", aliyunResponseCodeError(statusCode, bodyCode)
	}
	data := response.Body.Data
	if data == nil {
		return token, "", aliyunInvalidResponseError(nil)
	}
	token = aliyunTemporaryUploadToken{
		AccessKeyID:         aliyunStringValue(data.AccessKeyId),
		AccessKeySecret:     aliyunStringValue(data.AccessKeySecret),
		SecurityToken:       aliyunStringValue(data.SecurityToken),
		BucketName:          aliyunStringValue(data.BucketName),
		FileNamePrefix:      aliyunStringValue(data.FileNamePrefix),
		OSSInternetEndpoint: aliyunStringValue(data.OssInternetEndPoint),
	}
	if token.AccessKeyID == "" || token.AccessKeySecret == "" ||
		token.SecurityToken == "" || token.BucketName == "" ||
		token.FileNamePrefix == "" || token.OSSInternetEndpoint == "" {
		return aliyunTemporaryUploadToken{}, "", aliyunInvalidResponseError(nil)
	}
	return token, aliyunStringValue(response.Body.RequestId), nil
}

// aliyunUploadTokenResponseCodes 安全读取临时上传令牌响应的状态码。
func aliyunUploadTokenResponseCodes(response *green.DescribeUploadTokenResponse) (int, int) {
	var statusCode int
	var bodyCode int
	if response != nil && response.StatusCode != nil {
		statusCode = int(*response.StatusCode)
	}
	if response != nil && response.Body != nil && response.Body.Code != nil {
		bodyCode = int(*response.Body.Code)
	}
	return statusCode, bodyCode
}

// uploadAliyunModerationImage 使用临时凭据上传图片，并返回尽力清理临时对象的函数。
func uploadAliyunModerationImage(
	ctx context.Context,
	token aliyunTemporaryUploadToken,
	request AliyunImageAuditRequest,
) (string, func(), error) {
	timeoutSeconds := int64((aliyunRequestTimeoutMilliseconds(ctx) + 999) / 1000)
	if timeoutSeconds < 1 {
		timeoutSeconds = 1
	}
	client, err := oss.New(
		token.OSSInternetEndpoint,
		token.AccessKeyID,
		token.AccessKeySecret,
		oss.SecurityToken(token.SecurityToken),
		oss.Timeout(timeoutSeconds, timeoutSeconds),
	)
	if err != nil {
		return "", func() {}, normalizeAliyunCallError(err)
	}
	bucket, err := client.Bucket(token.BucketName)
	if err != nil {
		return "", func() {}, normalizeAliyunCallError(err)
	}
	objectName := token.FileNamePrefix + uuid.NewString() + aliyunImageExtension(request.ContentType)
	if err := bucket.PutObject(
		objectName,
		bytes.NewReader(request.Image),
		oss.ContentType(strings.TrimSpace(request.ContentType)),
		oss.WithContext(ctx),
	); err != nil {
		return "", func() {}, normalizeAliyunCallError(err)
	}
	cleanup := func() {
		cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), aliyunModerationCleanupWait)
		defer cancel()
		_ = bucket.DeleteObject(objectName, oss.WithContext(cleanupContext))
	}
	return objectName, cleanup, nil
}

// aliyunImageExtension 返回受支持图片类型的安全扩展名。
func aliyunImageExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}

// mapAliyunModerationResponse 将阿里云响应映射为平台统一的通过或拒绝结果。
func mapAliyunModerationResponse(
	response *green.ImageModerationResponse,
) (AliyunImageAuditResponse, error) {
	var mapped AliyunImageAuditResponse
	if response == nil || response.StatusCode == nil || *response.StatusCode != 200 ||
		response.Body == nil || response.Body.Code == nil || *response.Body.Code != 200 ||
		response.Body.Data == nil {
		statusCode, bodyCode := aliyunModerationResponseCodes(response)
		return mapped, aliyunResponseCodeError(statusCode, bodyCode)
	}
	data := response.Body.Data
	labelSet := make(map[string]struct{})
	var riskLevel string
	for _, item := range data.Result {
		if item == nil {
			continue
		}
		label := strings.TrimSpace(aliyunStringValue(item.Label))
		if label != "" && !aliyunNoRiskLabel(label) {
			labelSet[label] = struct{}{}
		}
		if riskLevel == "" {
			riskLevel = strings.TrimSpace(aliyunStringValue(item.RiskLevel))
		}
	}
	if value := strings.TrimSpace(aliyunStringValue(data.RiskLevel)); value != "" {
		riskLevel = value
	}
	labels := make([]string, 0, len(labelSet))
	for label := range labelSet {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	suggestion := "pass"
	if len(labels) > 0 {
		suggestion = "block"
	}
	mapped = AliyunImageAuditResponse{
		Suggestion:        suggestion,
		RiskLabels:        labels,
		ProviderRequestID: aliyunStringValue(response.Body.RequestId),
		SafeSummary: map[string]interface{}{
			"httpStatus":  int(*response.StatusCode),
			"code":        int(*response.Body.Code),
			"service":     "ImageModeration",
			"resultCount": len(data.Result),
			"riskLabels":  labels,
		},
	}
	if riskLevel != "" {
		mapped.RiskLevel = &riskLevel
		mapped.SafeSummary["riskLevel"] = riskLevel
	}
	return mapped, nil
}

// aliyunModerationResponseCodes 安全读取图片审核响应的状态码。
func aliyunModerationResponseCodes(response *green.ImageModerationResponse) (int, int) {
	var statusCode int
	var bodyCode int
	if response != nil && response.StatusCode != nil {
		statusCode = int(*response.StatusCode)
	}
	if response != nil && response.Body != nil && response.Body.Code != nil {
		bodyCode = int(*response.Body.Code)
	}
	return statusCode, bodyCode
}

// aliyunNoRiskLabel 判断供应商标签是否明确表示未命中风险。
func aliyunNoRiskLabel(label string) bool {
	normalized := strings.ToLower(strings.TrimSpace(label))
	return normalized == aliyunModerationNoRiskLabel
}

// normalizeAliyunCallError 将 SDK、OSS 和上下文错误归一化为安全类别。
func normalizeAliyunCallError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return &AliyunProviderError{
			Category:    ModerationCategoryTimeout,
			SafeMessage: "图片审核服务调用超时",
			Cause:       err,
		}
	}
	var responseError dara.ResponseError
	if errors.As(err, &responseError) {
		return aliyunCodeProviderError(
			aliyunIntValue(responseError.GetStatusCode()),
			aliyunStringValue(responseError.GetCode()),
			err,
		)
	}
	var ossError oss.ServiceError
	if errors.As(err, &ossError) {
		return aliyunCodeProviderError(ossError.StatusCode, ossError.Code, err)
	}
	return &AliyunProviderError{
		Category:    ModerationCategoryUnavailable,
		SafeMessage: "图片审核服务暂不可用",
		Cause:       err,
	}
}

// aliyunResponseCodeError 根据 HTTP 与业务状态码构造安全供应商错误。
func aliyunResponseCodeError(statusCode int, bodyCode int) error {
	code := ""
	if bodyCode > 0 {
		code = fmt.Sprintf("%d", bodyCode)
	}
	if statusCode == 0 {
		statusCode = bodyCode
	}
	return aliyunCodeProviderError(statusCode, code, nil)
}

// aliyunCodeProviderError 将阿里云状态码和错误码映射为平台错误类别。
func aliyunCodeProviderError(statusCode int, code string, cause error) error {
	normalizedCode := strings.ToLower(strings.TrimSpace(code))
	category := ModerationCategoryUnavailable
	safeMessage := "图片审核服务暂不可用"
	switch {
	case statusCode == 401 || statusCode == 403 ||
		strings.Contains(normalizedCode, "accesskey") ||
		strings.Contains(normalizedCode, "signature") ||
		strings.Contains(normalizedCode, "securitytoken") ||
		strings.Contains(normalizedCode, "unauthorized") ||
		strings.Contains(normalizedCode, "forbidden"):
		category = ModerationCategoryAuthenticationFailed
		safeMessage = "图片审核凭据无效"
	case statusCode == 429 ||
		strings.Contains(normalizedCode, "throttl") ||
		strings.Contains(normalizedCode, "ratelimit") ||
		strings.Contains(normalizedCode, "quota"):
		category = ModerationCategoryRateLimited
		safeMessage = "图片审核服务繁忙"
	case statusCode == 408 || statusCode == 504 ||
		strings.Contains(normalizedCode, "timeout"):
		category = ModerationCategoryTimeout
		safeMessage = "图片审核服务调用超时"
	case statusCode == 400 || statusCode == 422 ||
		strings.Contains(normalizedCode, "invalidparameter") ||
		strings.Contains(normalizedCode, "invalidservice"):
		category = ModerationCategoryInvalidResponse
		safeMessage = "图片审核配置或请求无效"
	}
	return &AliyunProviderError{
		Category:    category,
		SafeMessage: safeMessage,
		Cause:       cause,
	}
}

// aliyunAuthenticationError 返回安全的鉴权失败错误。
func aliyunAuthenticationError(cause error) error {
	return &AliyunProviderError{
		Category:    ModerationCategoryAuthenticationFailed,
		SafeMessage: "图片审核凭据无效",
		Cause:       cause,
	}
}

// aliyunInvalidResponseError 返回安全的配置或响应无效错误。
func aliyunInvalidResponseError(cause error) error {
	return &AliyunProviderError{
		Category:    ModerationCategoryInvalidResponse,
		SafeMessage: "图片审核配置或响应无效",
		Cause:       cause,
	}
}

// aliyunStringPointer 返回字符串指针。
func aliyunStringPointer(value string) *string {
	return &value
}

// aliyunStringValue 安全读取字符串指针。
func aliyunStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

// aliyunIntValue 安全读取整数指针。
func aliyunIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

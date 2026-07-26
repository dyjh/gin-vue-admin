package orderfood

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	green "github.com/alibabacloud-go/green-20220302/v3/client"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
)

// TestEnvironmentAliyunCredentialResolver 验证环境变量凭据解析和错误输入拒绝逻辑。
func TestEnvironmentAliyunCredentialResolver(t *testing.T) {
	resolver := EnvironmentAliyunCredentialResolver{}
	raw, err := json.Marshal(aliyunEnvironmentCredentialValue{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
		SecurityToken:   "test-token",
	})
	if err != nil {
		t.Fatalf("marshal credential: %v", err)
	}
	t.Setenv("ORDERFOOD_ALIYUN_MODERATION_TEST", string(raw))

	credential, err := resolver.ResolveAliyunCredential(
		context.Background(),
		"env://ORDERFOOD_ALIYUN_MODERATION_TEST",
	)
	if err != nil {
		t.Fatalf("resolve credential: %v", err)
	}
	if credential.AccessKeyID != "test-id" ||
		credential.AccessKeySecret != "test-secret" ||
		credential.SecurityToken != "test-token" {
		t.Fatalf("unexpected credential: %+v", credential)
	}

	if _, err := resolver.ResolveAliyunCredential(
		context.Background(),
		"secret://ORDERFOOD_ALIYUN_MODERATION_TEST",
	); err == nil {
		t.Fatal("expected unsupported reference to fail")
	}
	t.Setenv("ORDERFOOD_ALIYUN_MODERATION_INVALID", `{"accessKeyId":"only-id"}`)
	if _, err := resolver.ResolveAliyunCredential(
		context.Background(),
		"env://ORDERFOOD_ALIYUN_MODERATION_INVALID",
	); err == nil {
		t.Fatal("expected incomplete credential to fail")
	}
}

// TestNormalizeAliyunModerationEndpoint 验证只接受没有路径参数的 HTTPS 接入地址。
func TestNormalizeAliyunModerationEndpoint(t *testing.T) {
	endpoint, err := normalizeAliyunModerationEndpoint(
		"https://green-cip.cn-shanghai.aliyuncs.com",
	)
	if err != nil {
		t.Fatalf("normalize endpoint: %v", err)
	}
	if endpoint != "green-cip.cn-shanghai.aliyuncs.com" {
		t.Fatalf("unexpected endpoint: %s", endpoint)
	}
	for _, invalid := range []string{
		"http://green-cip.cn-shanghai.aliyuncs.com",
		"https://green-cip.cn-shanghai.aliyuncs.com/path",
		"https://green-cip.cn-shanghai.aliyuncs.com?debug=1",
	} {
		if _, err := normalizeAliyunModerationEndpoint(invalid); err == nil {
			t.Fatalf("expected endpoint to fail: %s", invalid)
		}
	}
}

// TestValidateModerationConfigInputRequiresExecutableCredentialReference 验证保存时拒绝运行时不支持的凭据引用和带路径地址。
func TestValidateModerationConfigInputRequiresExecutableCredentialReference(t *testing.T) {
	validReference := "env://ORDERFOOD_ALIYUN_MODERATION"
	input := orderfoodRequest.ModerationConfigUpdateInput{
		Enabled:         true,
		Region:          "cn-shanghai",
		Endpoint:        "https://green-cip.cn-shanghai.aliyuncs.com",
		ServiceCode:     "baselineCheck",
		TimeoutMS:       3000,
		RetryCount:      1,
		RetryBackoffMS:  200,
		CredentialRef:   &validReference,
		Reason:          "启用图片审核",
		ExpectedVersion: 1,
	}
	if err := validateModerationConfigInput(input); err != nil {
		t.Fatalf("validate executable input: %v", err)
	}
	unsupportedReference := "secret://aliyun/orderfood"
	input.CredentialRef = &unsupportedReference
	if err := validateModerationConfigInput(input); err == nil {
		t.Fatal("expected unsupported credential reference to fail")
	}
	input.CredentialRef = &validReference
	input.Endpoint = "https://green-cip.cn-shanghai.aliyuncs.com/path"
	if err := validateModerationConfigInput(input); err == nil {
		t.Fatal("expected endpoint path to fail")
	}
}

// TestMapAliyunModerationResponse 验证未命中风险通过、命中风险拒绝和无效响应失败。
func TestMapAliyunModerationResponse(t *testing.T) {
	httpOK := int32(200)
	bodyOK := int32(200)
	requestID := "aliyun-request-id"
	nonLabel := "nonLabel"
	passResponse := &green.ImageModerationResponse{
		StatusCode: &httpOK,
		Body: &green.ImageModerationResponseBody{
			Code:      &bodyOK,
			RequestId: &requestID,
			Data: &green.ImageModerationResponseBodyData{
				Result: []*green.ImageModerationResponseBodyDataResult{
					{Label: &nonLabel},
				},
			},
		},
	}
	passed, err := mapAliyunModerationResponse(passResponse)
	if err != nil {
		t.Fatalf("map pass response: %v", err)
	}
	if passed.Suggestion != "pass" || len(passed.RiskLabels) != 0 ||
		passed.ProviderRequestID != requestID {
		t.Fatalf("unexpected pass mapping: %+v", passed)
	}

	riskLabel := "violent_explosion"
	riskLevel := "high"
	riskResponse := &green.ImageModerationResponse{
		StatusCode: &httpOK,
		Body: &green.ImageModerationResponseBody{
			Code:      &bodyOK,
			RequestId: &requestID,
			Data: &green.ImageModerationResponseBodyData{
				RiskLevel: &riskLevel,
				Result: []*green.ImageModerationResponseBodyDataResult{
					{Label: &riskLabel, RiskLevel: &riskLevel},
					{Label: &riskLabel, RiskLevel: &riskLevel},
				},
			},
		},
	}
	rejected, err := mapAliyunModerationResponse(riskResponse)
	if err != nil {
		t.Fatalf("map risk response: %v", err)
	}
	if rejected.Suggestion != "block" ||
		len(rejected.RiskLabels) != 1 ||
		rejected.RiskLabels[0] != riskLabel ||
		rejected.RiskLevel == nil ||
		*rejected.RiskLevel != riskLevel {
		t.Fatalf("unexpected risk mapping: %+v", rejected)
	}

	if _, err := mapAliyunModerationResponse(nil); err == nil {
		t.Fatal("expected nil response to fail")
	}
}

// TestAliyunProviderErrorClassification 验证鉴权、限流和超时错误使用稳定安全类别。
func TestAliyunProviderErrorClassification(t *testing.T) {
	testCases := []struct {
		name       string // 测试名称
		statusCode int    // HTTP 状态码
		code       string // 供应商错误码
		category   string // 期望平台类别
	}{
		{name: "auth", statusCode: 403, code: "Forbidden", category: ModerationCategoryAuthenticationFailed},
		{name: "rate", statusCode: 429, code: "Throttling.User", category: ModerationCategoryRateLimited},
		{name: "timeout", statusCode: 504, code: "Timeout", category: ModerationCategoryTimeout},
		{name: "invalid", statusCode: 400, code: "InvalidParameter", category: ModerationCategoryInvalidResponse},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := aliyunCodeProviderError(testCase.statusCode, testCase.code, errors.New("provider detail"))
			var providerError *AliyunProviderError
			if !errors.As(err, &providerError) {
				t.Fatalf("expected AliyunProviderError, got %T", err)
			}
			if providerError.Category != testCase.category {
				t.Fatalf("unexpected category: got=%s want=%s", providerError.Category, testCase.category)
			}
			if providerError.Error() == "provider detail" {
				t.Fatal("provider detail leaked through safe error")
			}
		})
	}
}

// TestNewModerationServiceUsesProductionAliyunProvider 验证默认服务不再装配空供应商。
func TestNewModerationServiceUsesProductionAliyunProvider(t *testing.T) {
	service := NewModerationService(nil, nil, nil, nil, nil)
	provider, ok := service.Provider.(*AliyunModerationProvider)
	if !ok {
		t.Fatalf("unexpected provider type: %T", service.Provider)
	}
	if provider.Credentials == nil || provider.Client == nil {
		t.Fatal("production provider dependencies must be configured")
	}
}

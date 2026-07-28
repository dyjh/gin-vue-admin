package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
)

// AIConnectionResult 表示 AI 供应商连接测试的归一化结果。
type AIConnectionResult struct {
	Success     bool   // 连接是否成功
	Category    string // 稳定结果分类
	DurationMS  int    // 调用耗时（毫秒）
	SafeMessage string // 可安全展示的结果说明
}

// AIPromptRunRequest 表示一次 AI 提示词测试调用。
type AIPromptRunRequest struct {
	ModelCapability orderfoodModel.AIModelCapability // 模型能力类型
	SystemPrompt    string                           // 系统提示词
	UserPrompt      string                           // 用户提示词
}

// AIPromptRunResult 表示 AI 提示词测试调用的归一化结果。
type AIPromptRunResult struct {
	Success      bool   // 调用是否成功
	DurationMS   int    // 调用耗时（毫秒）
	InputTokens  *int   // 输入Token数量
	OutputTokens *int   // 输出Token数量
	SafeMessage  string // 可安全展示的结果说明
}

// AIConnector 定义AI连接器所需的业务能力。
type AIConnector interface {
	ListModels(
		ctx context.Context,
		provider orderfoodModel.AIProvider,
		timeout time.Duration,
	) ([]string, error)
	TestConnection(
		ctx context.Context,
		provider orderfoodModel.AIProvider,
		timeout time.Duration,
	) (AIConnectionResult, error)
	TestPrompt(
		ctx context.Context,
		provider orderfoodModel.AIProvider,
		model orderfoodModel.AIModel,
		request AIPromptRunRequest,
		timeout time.Duration,
	) (AIPromptRunResult, error)
}

// HTTPAIConnector 通过供应商 HTTP 接口执行连接和提示词测试。
type HTTPAIConnector struct {
	Client *http.Client // 可注入的HTTP客户端
}

// NewHTTPAIConnector 创建HTTP AI连接器实例。
func NewHTTPAIConnector(client *http.Client) *HTTPAIConnector {
	if client == nil {
		client = &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return &HTTPAIConnector{Client: client}
}

func (connector *HTTPAIConnector) credential(
	ctx context.Context,
	provider orderfoodModel.AIProvider,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return "", errors.New("provider API key is unavailable")
	}
	return apiKey, nil
}

func aiProviderEndpoint(baseURL string, path string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + path
}

func aiHTTPResult(
	status int,
	duration time.Duration,
) AIConnectionResult {
	result := AIConnectionResult{DurationMS: int(duration.Milliseconds())}
	switch {
	case status >= 200 && status < 300:
		result.Success = true
		result.Category = "ok"
		result.SafeMessage = "连接成功"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		result.Category = "authentication_failed"
		result.SafeMessage = "凭据校验失败"
	case status == http.StatusTooManyRequests:
		result.Category = "rate_limited"
		result.SafeMessage = "供应商请求频率受限"
	case status >= 500:
		result.Category = "unavailable"
		result.SafeMessage = "供应商服务暂不可用"
	default:
		result.Category = "invalid_response"
		result.SafeMessage = "供应商返回了无法识别的响应"
	}
	return result
}

const maxDiscoveredAIModels = 1000

// normalizeDiscoveredAIModelKeys 仅保留可安全作为模型标识的供应商返回值。
func normalizeDiscoveredAIModelKeys(modelKeys []string) []string {
	unique := make(map[string]struct{}, len(modelKeys))
	for _, rawModelKey := range modelKeys {
		modelKey := strings.TrimSpace(rawModelKey)
		length := len([]rune(modelKey))
		if length < 1 || length > 120 {
			continue
		}
		invalid := false
		for _, character := range modelKey {
			if unicode.IsSpace(character) || unicode.IsControl(character) {
				invalid = true
				break
			}
		}
		if invalid {
			continue
		}
		unique[modelKey] = struct{}{}
		if len(unique) >= maxDiscoveredAIModels {
			break
		}
	}
	result := make([]string, 0, len(unique))
	for modelKey := range unique {
		result = append(result, modelKey)
	}
	sort.Strings(result)
	return result
}

// ListModels 从供应商兼容接口读取可用模型标识，不返回原始响应或凭据。
func (connector *HTTPAIConnector) ListModels(
	ctx context.Context,
	provider orderfoodModel.AIProvider,
	timeout time.Duration,
) ([]string, error) {
	credential, err := connector.credential(ctx, provider)
	if err != nil {
		return nil, err
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(
		requestContext,
		http.MethodGet,
		aiProviderEndpoint(provider.BaseURL, "/models"),
		nil,
	)
	if err != nil {
		return nil, errors.New("provider model catalog URL is invalid")
	}
	request.Header.Set("Authorization", "Bearer "+credential)
	request.Header.Set("Accept", "application/json")
	response, err := connector.Client.Do(request)
	if err != nil {
		if errors.Is(requestContext.Err(), context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}
		return nil, errors.New("provider model catalog request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 8192))
		return nil, errors.New("provider model catalog request was rejected")
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, errors.New("provider model catalog response is invalid")
	}
	modelKeys := make([]string, 0, len(payload.Data))
	for _, model := range payload.Data {
		modelKeys = append(modelKeys, model.ID)
	}
	return normalizeDiscoveredAIModelKeys(modelKeys), nil
}

// TestConnection 测试连接。
func (connector *HTTPAIConnector) TestConnection(
	ctx context.Context,
	provider orderfoodModel.AIProvider,
	timeout time.Duration,
) (AIConnectionResult, error) {
	startedAt := time.Now()
	credential, err := connector.credential(ctx, provider)
	if err != nil {
		return AIConnectionResult{
			Category:    "unavailable",
			DurationMS:  int(time.Since(startedAt).Milliseconds()),
			SafeMessage: "API Key 当前不可用",
		}, nil
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(
		requestContext,
		http.MethodGet,
		aiProviderEndpoint(provider.BaseURL, "/models"),
		nil,
	)
	if err != nil {
		return AIConnectionResult{
			Category:    "invalid_response",
			DurationMS:  int(time.Since(startedAt).Milliseconds()),
			SafeMessage: "供应商地址无效",
		}, nil
	}
	request.Header.Set("Authorization", "Bearer "+credential)
	request.Header.Set("Accept", "application/json")
	response, err := connector.Client.Do(request)
	duration := time.Since(startedAt)
	if err != nil {
		category := "unavailable"
		safeMessage := "供应商连接失败"
		if errors.Is(requestContext.Err(), context.DeadlineExceeded) {
			category = "timeout"
			safeMessage = "供应商连接超时"
		}
		return AIConnectionResult{
			Category:    category,
			DurationMS:  int(duration.Milliseconds()),
			SafeMessage: safeMessage,
		}, nil
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 8192))
	return aiHTTPResult(response.StatusCode, duration), nil
}

// TestPrompt 测试提示词。
func (connector *HTTPAIConnector) TestPrompt(
	ctx context.Context,
	provider orderfoodModel.AIProvider,
	model orderfoodModel.AIModel,
	promptRequest AIPromptRunRequest,
	timeout time.Duration,
) (AIPromptRunResult, error) {
	startedAt := time.Now()
	credential, err := connector.credential(ctx, provider)
	if err != nil {
		return AIPromptRunResult{
			SafeMessage: "API Key 当前不可用",
		}, nil
	}

	endpoint := aiProviderEndpoint(provider.BaseURL, "/chat/completions")
	payload := map[string]interface{}{
		"model": model.ModelKey,
		"messages": []map[string]string{
			{"role": "system", "content": promptRequest.SystemPrompt},
			{"role": "user", "content": promptRequest.UserPrompt},
		},
		"temperature": 0,
	}
	if promptRequest.ModelCapability == orderfoodModel.AIModelImageGeneration {
		endpoint = aiProviderEndpoint(provider.BaseURL, "/images/generations")
		payload = map[string]interface{}{
			"model":  model.ModelKey,
			"prompt": promptRequest.SystemPrompt + "\n\n" + promptRequest.UserPrompt,
			"n":      1,
		}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return AIPromptRunResult{SafeMessage: "测试请求无法生成"}, nil
	}

	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(
		requestContext,
		http.MethodPost,
		endpoint,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return AIPromptRunResult{SafeMessage: "供应商地址无效"}, nil
	}
	request.Header.Set("Authorization", "Bearer "+credential)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := connector.Client.Do(request)
	duration := time.Since(startedAt)
	if err != nil {
		safeMessage := "供应商测试失败"
		if errors.Is(requestContext.Err(), context.DeadlineExceeded) {
			safeMessage = "供应商测试超时"
		}
		return AIPromptRunResult{
			DurationMS:  int(duration.Milliseconds()),
			SafeMessage: safeMessage,
		}, nil
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 65536))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		connectionResult := aiHTTPResult(response.StatusCode, duration)
		return AIPromptRunResult{
			DurationMS:  connectionResult.DurationMS,
			SafeMessage: connectionResult.SafeMessage,
		}, nil
	}

	var envelope struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			InputTokens      int `json:"input_tokens"`
			OutputTokens     int `json:"output_tokens"`
		} `json:"usage"`
	}
	_ = json.Unmarshal(body, &envelope)
	inputTokens := envelope.Usage.PromptTokens
	if inputTokens == 0 {
		inputTokens = envelope.Usage.InputTokens
	}
	outputTokens := envelope.Usage.CompletionTokens
	if outputTokens == 0 {
		outputTokens = envelope.Usage.OutputTokens
	}
	var inputTokenPointer *int
	if inputTokens > 0 {
		inputTokenPointer = &inputTokens
	}
	var outputTokenPointer *int
	if outputTokens > 0 {
		outputTokenPointer = &outputTokens
	}
	return AIPromptRunResult{
		Success:      true,
		DurationMS:   int(duration.Milliseconds()),
		InputTokens:  inputTokenPointer,
		OutputTokens: outputTokenPointer,
		SafeMessage:  "测试成功",
	}, nil
}

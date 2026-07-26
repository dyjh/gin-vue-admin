package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	adminResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AssistService 提供小程序增强功能业务能力。
type AssistService struct {
	DB         *gorm.DB           // 业务数据库
	Content    *ContentService    // 菜品与菜谱服务
	Engagement *EngagementService // 积分扣除与退款服务
	Upload     *UploadService     // AI生成图片持久化服务
	Client     *http.Client       // 可注入的HTTP客户端
	Now        func() time.Time   // 可注入的当前时间函数
}

func (service *AssistService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return ServiceGroupApp.ContentService.database()
}

func (service *AssistService) content() *ContentService {
	if service != nil && service.Content != nil {
		return service.Content
	}
	return &ServiceGroupApp.ContentService
}

func (service *AssistService) engagement() *EngagementService {
	if service != nil && service.Engagement != nil {
		return service.Engagement
	}
	return &ServiceGroupApp.EngagementService
}

func (service *AssistService) upload() *UploadService {
	if service != nil && service.Upload != nil {
		return service.Upload
	}
	return &ServiceGroupApp.UploadService
}

func (service *AssistService) client() *http.Client {
	if service != nil && service.Client != nil {
		return service.Client
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (service *AssistService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

type capabilityRuntime struct {
	Config   orderfoodModel.AICapabilityConfig
	Model    orderfoodModel.AIModel
	Provider orderfoodModel.AIProvider
}

func (service *AssistService) capability(
	ctx context.Context,
	code string,
) (capabilityRuntime, error) {
	db := service.database().WithContext(ctx)
	var definition orderfoodModel.AICapabilityDefinition
	if err := db.First(&definition, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return capabilityRuntime{}, appErrors.FrontFeatureDisabled.DefaultMsg()
		}
		return capabilityRuntime{}, appErrors.FrontInternal.Wrap(err, "load feature capability")
	}
	var config orderfoodModel.AICapabilityConfig
	if err := db.First(&config, "capability_code = ?", code).Error; err != nil {
		return capabilityRuntime{}, appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	var model orderfoodModel.AIModel
	if err := db.First(&model, "id = ? AND enabled = ?", config.PrimaryModelID, true).Error; err != nil {
		return capabilityRuntime{}, appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	var provider orderfoodModel.AIProvider
	if err := db.First(&provider, "id = ? AND enabled = ?", model.ProviderID, true).Error; err != nil {
		return capabilityRuntime{}, appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	return capabilityRuntime{Config: config, Model: model, Provider: provider}, nil
}

func renderPrompt(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

func trimJSONFence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```JSON")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}

type compatibleChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type compatibleImageResponse struct {
	Data []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

func (service *AssistService) runJSONCapability(
	ctx context.Context,
	code string,
	variables map[string]string,
	output interface{},
	usageID string,
) (runErr error) {
	startedAt := service.now()
	runtime, err := service.capability(ctx, code)
	if err != nil {
		return err
	}
	if err := service.recordFeatureUsageRuntime(
		ctx,
		usageID,
		code,
		runtime,
		variables,
		0,
	); err != nil {
		return err
	}
	// 失败追踪只保存安全摘要；供应商原始响应和完整提示词不会进入调用记录。
	defer func() {
		if runErr != nil {
			_ = service.recordFeatureUsageFailure(ctx, usageID, startedAt)
		}
	}()
	credential, err := (orderfoodService.EnvironmentAICredentialResolver{}).
		Resolve(ctx, runtime.Provider.CredentialRef)
	if err != nil {
		return appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	userContent := interface{}(renderPrompt(runtime.Config.UserPromptTemplate, variables))
	if imageURL := strings.TrimSpace(variables["image_content"]); imageURL != "" &&
		(code == orderfoodModel.AICapabilityRecipeImageExtract ||
			code == orderfoodModel.AICapabilityCheckinImageAnalyze) {
		textVariables := make(map[string]string, len(variables))
		for key, value := range variables {
			textVariables[key] = value
		}
		textVariables["image_content"] = "见随消息提交的图片"
		userContent = []map[string]interface{}{
			{"type": "text", "text": renderPrompt(runtime.Config.UserPromptTemplate, textVariables)},
			{"type": "image_url", "image_url": map[string]string{"url": imageURL}},
		}
	}
	payload := map[string]interface{}{
		"model": runtime.Model.ModelKey,
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": runtime.Config.SystemPrompt +
					"\n只输出符合约定字段的 JSON，不要输出 Markdown。",
			},
			{"role": "user", "content": userContent},
		},
		"temperature": 0,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return appErrors.FrontInternal.Wrap(err, "encode feature request")
	}
	timeout := time.Duration(runtime.Config.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(
		requestContext,
		http.MethodPost,
		strings.TrimRight(runtime.Provider.BaseURL, "/")+"/chat/completions",
		bytes.NewReader(encoded),
	)
	if err != nil {
		return appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	request.Header.Set("Authorization", "Bearer "+credential)
	request.Header.Set("Content-Type", "application/json")
	httpResponse, err := service.client().Do(request)
	if err != nil {
		return appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(httpResponse.Body, 64<<10))
		return appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	var result compatibleChatResponse
	if err := json.NewDecoder(io.LimitReader(httpResponse.Body, 2<<20)).Decode(&result); err != nil ||
		len(result.Choices) == 0 {
		return appErrors.FrontResultInvalid.DefaultMsg()
	}
	if err := json.Unmarshal([]byte(trimJSONFence(result.Choices[0].Message.Content)), output); err != nil {
		return appErrors.FrontResultInvalid.Wrap(err, "decode feature result")
	}
	if err := service.recordFeatureUsageSuccess(
		ctx,
		usageID,
		startedAt,
		runtime.Model,
		result.Usage.PromptTokens,
		result.Usage.CompletionTokens,
		0,
		output,
	); err != nil {
		return err
	}
	return nil
}

func (service *AssistService) runImageCapability(
	ctx context.Context,
	variables map[string]string,
	usageID string,
) (content []byte, contentType string, runErr error) {
	startedAt := service.now()
	runtime, err := service.capability(ctx, orderfoodModel.AICapabilityDishCoverCreate)
	if err != nil {
		return nil, "", err
	}
	if err := service.recordFeatureUsageRuntime(
		ctx,
		usageID,
		orderfoodModel.AICapabilityDishCoverCreate,
		runtime,
		variables,
		1,
	); err != nil {
		return nil, "", err
	}
	defer func() {
		if runErr != nil {
			_ = service.recordFeatureUsageFailure(ctx, usageID, startedAt)
		}
	}()
	credential, err := (orderfoodService.EnvironmentAICredentialResolver{}).
		Resolve(ctx, runtime.Provider.CredentialRef)
	if err != nil {
		return nil, "", appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	payload := map[string]interface{}{
		"model": runtime.Model.ModelKey,
		"prompt": runtime.Config.SystemPrompt + "\n\n" +
			renderPrompt(runtime.Config.UserPromptTemplate, variables),
		"n": 1,
	}
	encoded, _ := json.Marshal(payload)
	timeout := time.Duration(runtime.Config.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(
		requestContext, http.MethodPost,
		strings.TrimRight(runtime.Provider.BaseURL, "/")+"/images/generations",
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, "", appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	request.Header.Set("Authorization", "Bearer "+credential)
	request.Header.Set("Content-Type", "application/json")
	httpResponse, err := service.client().Do(request)
	if err != nil {
		return nil, "", appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(httpResponse.Body, 64<<10))
		return nil, "", appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	var result compatibleImageResponse
	responseLimit := int64(maxFrontImageBytes)*4/3 + 64*1024
	if err := json.NewDecoder(io.LimitReader(httpResponse.Body, responseLimit)).Decode(&result); err != nil ||
		len(result.Data) == 0 {
		return nil, "", appErrors.FrontResultInvalid.DefaultMsg()
	}
	if result.Data[0].B64JSON != "" {
		content, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
		if err != nil {
			return nil, "", appErrors.FrontResultInvalid.DefaultMsg()
		}
		if err := service.recordFeatureUsageSuccess(
			ctx,
			usageID,
			startedAt,
			runtime.Model,
			0,
			0,
			1,
			map[string]interface{}{"imageGenerated": true},
		); err != nil {
			return nil, "", err
		}
		return content, http.DetectContentType(content), nil
	}
	imageRequest, err := http.NewRequestWithContext(
		requestContext,
		http.MethodGet,
		strings.TrimSpace(result.Data[0].URL),
		nil,
	)
	if err != nil ||
		imageRequest.URL == nil ||
		imageRequest.URL.Scheme != "https" ||
		imageRequest.URL.Host == "" ||
		imageRequest.URL.User != nil {
		return nil, "", appErrors.FrontResultInvalid.DefaultMsg()
	}
	imageResponse, err := service.client().Do(imageRequest)
	if err != nil {
		return nil, "", appErrors.FrontExecutionRefunded.DefaultMsg()
	}
	defer imageResponse.Body.Close()
	if imageResponse.StatusCode < 200 || imageResponse.StatusCode >= 300 ||
		imageResponse.ContentLength > maxFrontImageBytes {
		_, _ = io.Copy(io.Discard, io.LimitReader(imageResponse.Body, 64<<10))
		return nil, "", appErrors.FrontResultInvalid.DefaultMsg()
	}
	content, err = io.ReadAll(io.LimitReader(imageResponse.Body, maxFrontImageBytes+1))
	if err != nil || len(content) == 0 || len(content) > maxFrontImageBytes {
		return nil, "", appErrors.FrontResultInvalid.DefaultMsg()
	}
	if err := service.recordFeatureUsageSuccess(
		ctx,
		usageID,
		startedAt,
		runtime.Model,
		0,
		0,
		1,
		map[string]interface{}{"imageGenerated": true},
	); err != nil {
		return nil, "", err
	}
	return content, http.DetectContentType(content), nil
}

// recordFeatureUsageRuntime 保存实际供应商、模型、提示词摘要和敏感输入快照。
func (service *AssistService) recordFeatureUsageRuntime(
	ctx context.Context,
	usageID string,
	code string,
	runtime capabilityRuntime,
	variables map[string]string,
	imageCount int,
) error {
	if strings.TrimSpace(usageID) == "" {
		return nil
	}
	input := make(map[string]string, len(variables))
	imageMetadata := make([]map[string]string, 0, 1)
	for key, value := range variables {
		if key == "image_content" {
			imageMetadata = append(imageMetadata, map[string]string{"source": value})
			continue
		}
		input[key] = value
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return appErrors.FrontInternal.Wrap(err, "encode feature usage input")
	}
	var imageJSON datatypes.JSON
	if len(imageMetadata) > 0 {
		encoded, encodeErr := json.Marshal(imageMetadata)
		if encodeErr != nil {
			return appErrors.FrontInternal.Wrap(encodeErr, "encode feature usage image metadata")
		}
		imageJSON = datatypes.JSON(encoded)
	}
	promptMode := string(runtime.Config.PromptMode)
	if runtime.Config.PromptMode == orderfoodModel.AIPromptPreset {
		promptMode = "default"
	}
	updates := map[string]interface{}{
		"capability_code":     code,
		"provider_id":         runtime.Provider.ID,
		"provider_name":       runtime.Provider.Name,
		"model_id":            runtime.Model.ID,
		"model_name":          runtime.Model.Name,
		"prompt_mode":         promptMode,
		"prompt_hash":         runtime.Config.PromptHash,
		"image_count":         imageCount,
		"original_input_json": datatypes.JSON(inputJSON),
		"image_metadata_json": imageJSON,
		"updated_at":          service.now(),
	}
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontFeatureUsage{}).
		Where("id = ?", usageID).Updates(updates).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "record feature usage runtime")
	}
	return nil
}

// recordFeatureUsageSuccess 保存模型返回摘要、Token和估算成本。
func (service *AssistService) recordFeatureUsageSuccess(
	ctx context.Context,
	usageID string,
	startedAt time.Time,
	model orderfoodModel.AIModel,
	inputTokens int,
	outputTokens int,
	imageCount int,
	output interface{},
) error {
	if strings.TrimSpace(usageID) == "" {
		return nil
	}
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return appErrors.FrontInternal.Wrap(err, "encode feature usage output")
	}
	updates := map[string]interface{}{
		"duration_ms":        int(service.now().Sub(startedAt).Milliseconds()),
		"estimated_cost_cny": estimateRuntimeCost(model, inputTokens, outputTokens, imageCount),
		"model_output_json":  datatypes.JSON(outputJSON),
		"failure_category":   nil,
		"failure_summary":    nil,
		"updated_at":         service.now(),
	}
	if inputTokens > 0 {
		updates["input_tokens"] = inputTokens
	}
	if outputTokens > 0 {
		updates["output_tokens"] = outputTokens
	}
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontFeatureUsage{}).
		Where("id = ?", usageID).Updates(updates).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "record successful feature usage")
	}
	return nil
}

// recordFeatureUsageFailure 保存不含供应商原文的失败分类和摘要。
func (service *AssistService) recordFeatureUsageFailure(
	ctx context.Context,
	usageID string,
	startedAt time.Time,
) error {
	if strings.TrimSpace(usageID) == "" {
		return nil
	}
	category := "execution_failed"
	summary := "模型调用失败或返回结果未通过校验"
	if err := service.database().WithContext(ctx).Model(&orderfoodModel.FrontFeatureUsage{}).
		Where("id = ?", usageID).Updates(map[string]interface{}{
		"duration_ms":      int(service.now().Sub(startedAt).Milliseconds()),
		"failure_category": category, "failure_summary": summary,
		"updated_at": service.now(),
	}).Error; err != nil {
		return appErrors.FrontInternal.Wrap(err, "record failed feature usage")
	}
	return nil
}

// estimateRuntimeCost 根据实际Token或图片数生成六位小数的人民币估算成本。
func estimateRuntimeCost(
	model orderfoodModel.AIModel,
	inputTokens int,
	outputTokens int,
	imageCount int,
) string {
	total := new(big.Rat)
	addTokenCost := func(tokens int, price *string) {
		if tokens <= 0 || price == nil {
			return
		}
		rate, ok := new(big.Rat).SetString(*price)
		if !ok {
			return
		}
		total.Add(total, new(big.Rat).Mul(
			rate,
			new(big.Rat).SetFrac64(int64(tokens), 1_000_000),
		))
	}
	addTokenCost(inputTokens, model.InputPricePerMillionTokensCNY)
	addTokenCost(outputTokens, model.OutputPricePerMillionTokensCNY)
	if imageCount > 0 && model.ImagePricePerUnitCNY != nil {
		if price, ok := new(big.Rat).SetString(*model.ImagePricePerUnitCNY); ok {
			total.Add(total, new(big.Rat).Mul(price, big.NewRat(int64(imageCount), 1)))
		}
	}
	return total.FloatString(6)
}

// ExtractDish 解析菜品。
func (service *AssistService) ExtractDish(
	ctx context.Context,
	userID string,
	input frontRequest.DishExtractionInput,
) (frontRequest.DishDraftInput, frontResponse.FeatureUsageResult, error) {
	variables := map[string]string{"locale": "zh-CN"}
	code := orderfoodModel.AICapabilityDishTextExtract
	if input.Text != nil {
		variables["input_text"] = *input.Text
	}
	if input.ImageFileID != nil {
		var asset orderfoodModel.FrontMediaAsset
		if err := service.database().WithContext(ctx).First(&asset,
			"id = ? AND user_id = ? AND scene = ? AND review_status = ?",
			*input.ImageFileID, userID, "dish_extract", orderfoodModel.MediaReviewPassed,
		).Error; err != nil {
			return frontRequest.DishDraftInput{}, frontResponse.FeatureUsageResult{}, appErrors.FrontInvalidImage.DefaultMsg()
		}
		code = orderfoodModel.AICapabilityRecipeImageExtract
		variables["image_content"] = asset.URL
	}
	if len(variables) == 1 {
		return frontRequest.DishDraftInput{}, frontResponse.FeatureUsageResult{}, appErrors.FrontBadRequest.DefaultMsg()
	}
	usage, _, err := service.engagement().BeginFeatureUsage(ctx, userID, code, true)
	if err != nil {
		return frontRequest.DishDraftInput{}, frontResponse.FeatureUsageResult{}, err
	}
	var draft frontRequest.DishDraftInput
	err = service.runJSONCapability(ctx, code, variables, &draft, usage.ID)
	usageResult, finishErr := service.engagement().FinishFeatureUsage(ctx, usage, err == nil)
	if finishErr != nil {
		return frontRequest.DishDraftInput{}, frontResponse.FeatureUsageResult{}, finishErr
	}
	if err != nil {
		return frontRequest.DishDraftInput{}, usageResult, err
	}
	if draft.Serving < 1 {
		draft.Serving = 1
	}
	if draft.Status == "" {
		draft.Status = orderfoodModel.DishStatusDraft
	}
	return draft, usageResult, nil
}

// CreateCover 创建封面。
func (service *AssistService) CreateCover(
	ctx context.Context,
	userID string,
	input frontRequest.DishCoverInput,
	requestID string,
) (frontResponse.ImageUploadResult, frontResponse.FeatureUsageResult, error) {
	usage, _, err := service.engagement().BeginFeatureUsage(
		ctx,
		userID,
		orderfoodModel.AICapabilityDishCoverCreate,
		true,
	)
	if err != nil {
		return frontResponse.ImageUploadResult{}, frontResponse.FeatureUsageResult{}, err
	}
	variables := map[string]string{
		"dish_name":   input.Name,
		"description": stringValue(input.Description),
		"ingredients": "",
	}
	content, contentType, err := service.runImageCapability(ctx, variables, usage.ID)
	if err == nil {
		var image frontResponse.ImageUploadResult
		image, err = service.upload().UploadGeneratedImage(ctx, userID, contentType, content, requestID)
		usageResult, finishErr := service.engagement().FinishFeatureUsage(ctx, usage, err == nil)
		if finishErr != nil {
			return frontResponse.ImageUploadResult{}, frontResponse.FeatureUsageResult{}, finishErr
		}
		return image, usageResult, err
	}
	usageResult, finishErr := service.engagement().FinishFeatureUsage(ctx, usage, false)
	if finishErr != nil {
		return frontResponse.ImageUploadResult{}, frontResponse.FeatureUsageResult{}, finishErr
	}
	return frontResponse.ImageUploadResult{}, usageResult, err
}

// SuggestionStatus 获取推荐菜功能解锁、额度和积分状态。
func (service *AssistService) SuggestionStatus(
	ctx context.Context,
	userID string,
) (bool, bool, int, int, int, *int64, *int, error) {
	runtime, err := service.engagement().runtimeService().Current(ctx)
	if err != nil {
		return false, false, 0, 7, 0, nil, nil, err
	}
	if !runtime.EnhancedFeaturesEnabled {
		return false, false, 0, 7, 0, nil, nil, appErrors.FrontFeatureDisabled.DefaultMsg()
	}
	var user orderfoodModel.MiniAppUser
	if err := service.database().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return false, false, 0, 7, 0, nil, nil, appErrors.FrontInternal.Wrap(err, "load suggestion status user")
	}
	freeQuota, pointCost, err := service.engagement().featureRemainingQuota(
		ctx,
		userID,
		orderfoodModel.AICapabilityMealSuggest,
	)
	if err != nil {
		return false, false, 0, 7, 0, nil, nil, err
	}
	balance := user.Points
	return true, user.CheckinDayCount >= 7, user.CheckinDayCount, 7, freeQuota, &balance, &pointCost, nil
}

type suggestionRecallCandidate struct {
	Dish               orderfoodModel.UserDish
	OfficialDishID     *uint
	RecommendationID   *string
	SourceLabel        string
	Score              float64
	RecommendationSort int
	Copied             bool
	CopiedDishID       *uint
	CopiedPublicID     *string
}

type suggestionDishSnapshot struct {
	Dish                 frontResponse.SuggestionDish
	SourceDishID         *uint
	SourceOfficialDishID *uint
	CopiedDishID         *uint
	NormalizedName       string
}

type suggestionModelIngredient struct {
	Name   string  `json:"name"`
	Amount string  `json:"amount"`
	Unit   string  `json:"unit"`
	Note   *string `json:"note"`
}

type suggestionModelStep struct {
	Text string `json:"text"`
}

type suggestionModelDish struct {
	Name        string                      `json:"name"`
	Cuisine     string                      `json:"cuisine"`
	Category    string                      `json:"category"`
	Tags        []string                    `json:"tags"`
	Serving     int                         `json:"serving"`
	Description string                      `json:"description"`
	Ingredients []suggestionModelIngredient `json:"ingredients"`
	Steps       []suggestionModelStep       `json:"steps"`
}

type suggestionModelOutput struct {
	Reason string                `json:"reason"`
	Dishes []suggestionModelDish `json:"dishes"`
}

type suggestionValidationContext struct {
	Categories          map[string]orderfoodModel.ContentCategory
	Tags                map[string]orderfoodModel.ContentTag
	Units               map[string]orderfoodModel.ContentUnit
	StandardDishes      map[string]orderfoodModel.StandardDishIndex
	StandardIngredients map[string]orderfoodModel.StandardIngredient
}

func suggestionTargetCount(people int) int {
	if people <= 1 {
		return 1
	}
	count := people/2 + 2
	if count > 6 {
		return 6
	}
	return count
}

func normalizeSuggestionText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Join(strings.Fields(value), "")
}

func suggestionTextLengthBetween(value string, minimum int, maximum int) bool {
	length := len([]rune(strings.TrimSpace(value)))
	return length >= minimum && length <= maximum
}

// hardConditionKeyword 提取一条用户硬条件中用于菜品排除的关键词。
func hardConditionKeyword(value string) string {
	value = normalizeSuggestionText(value)
	if value == "" || strings.Contains(value, "无特殊忌口") {
		return ""
	}
	for _, prefix := range []string{"我不吃", "不吃", "忌口", "忌", "对"} {
		value = strings.TrimPrefix(value, prefix)
	}
	value = strings.TrimSuffix(value, "过敏")
	return strings.Trim(value, "，。；、· ")
}

// dishViolatesHardConditions 判断菜品名称、标签或食材是否命中用户硬条件。
func dishViolatesHardConditions(
	dish orderfoodModel.UserDish,
	hardConditions []string,
) bool {
	searchable := normalizeSuggestionText(dish.Name)
	for _, tag := range dish.Tags {
		searchable += normalizeSuggestionText(tag.Name)
	}
	for _, ingredient := range dish.Ingredients {
		searchable += normalizeSuggestionText(ingredient.Name)
	}
	for _, condition := range hardConditions {
		keyword := hardConditionKeyword(condition)
		if keyword != "" && strings.Contains(searchable, keyword) {
			return true
		}
	}
	return false
}

// prepareSuggestionPreferences 将已聚合画像和本次明确选择注入推荐筛选上下文。
func (service *AssistService) prepareSuggestionPreferences(
	ctx context.Context,
	userID string,
	input frontRequest.MealSuggestionInput,
) (frontRequest.MealSuggestionInput, error) {
	var profile orderfoodModel.UserPreferenceProfile
	err := service.database().WithContext(ctx).First(&profile, "user_id = ? AND has_profile = ?", userID, true).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return input, appErrors.FrontInternal.Wrap(err, "load suggestion preference profile")
	}
	if err == nil && len(profile.ProfileJSON) > 0 {
		var content adminResponse.PreferenceProfileContent
		if err := json.Unmarshal(profile.ProfileJSON, &content); err != nil {
			return input, appErrors.FrontInternal.Wrap(err, "decode suggestion preference profile")
		}
		for _, term := range append(content.CommonDishTags, content.FrequentlyOrderedDishTags...) {
			input.Tags = append(input.Tags, term.Name)
		}
		for _, setting := range content.UserSettings {
			switch setting.Type {
			case "allergy", "avoidance", "dietary_restriction":
				input.HardConditions = append(input.HardConditions, setting.Value)
			}
		}
		encoded, err := json.Marshal(content)
		if err != nil {
			return input, appErrors.FrontInternal.Wrap(err, "encode suggestion preference context")
		}
		input.PreferenceContext = string(encoded)
	}
	if input.Preference != nil {
		value := strings.TrimSpace(*input.Preference)
		if hardConditionKeyword(value) != "" &&
			(strings.Contains(value, "不吃") ||
				strings.Contains(value, "忌") ||
				strings.Contains(value, "过敏")) {
			input.HardConditions = append(input.HardConditions, value)
		}
	}
	input.Tags = normalizePreferenceTerms(input.Tags, 10)
	input.HardConditions = normalizePreferenceTerms(input.HardConditions, 20)
	return input, nil
}

func (service *AssistService) excludedSuggestionNames(
	ctx context.Context,
	userID string,
	input frontRequest.MealSuggestionInput,
) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(input.ExcludedSuggestionDishIDs) == 0 {
		return result, nil
	}
	var rows []orderfoodModel.FrontMealSuggestionDish
	err := service.database().WithContext(ctx).
		Table(orderfoodModel.FrontMealSuggestionDish{}.TableName()+" AS dishes").
		Select("dishes.*").
		Joins(
			"JOIN "+orderfoodModel.FrontMealSuggestion{}.TableName()+
				" AS suggestions ON suggestions.id = dishes.suggestion_id",
		).
		Where("dishes.id IN ? AND suggestions.user_id = ?", input.ExcludedSuggestionDishIDs, userID).
		Find(&rows).Error
	if err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "load excluded suggestion dishes")
	}
	for _, row := range rows {
		result[row.NormalizedName] = struct{}{}
	}
	return result, nil
}

func (service *AssistService) recallSuggestionCandidates(
	ctx context.Context,
	userID string,
	input frontRequest.MealSuggestionInput,
	excludedNames map[string]struct{},
) ([]suggestionRecallCandidate, error) {
	db := service.database().WithContext(ctx)
	var recommendations []orderfoodModel.PlatformRecommendation
	if err := db.Where("selected = ?", true).Find(&recommendations).Error; err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "load selected suggestion recommendations")
	}
	type recommendationReference struct {
		ID        string
		SortOrder int
	}
	recommendationByDishID := make(map[uint]recommendationReference, len(recommendations))
	recommendationByOfficialDishID := make(map[uint]recommendationReference, len(recommendations))
	recommendationDishIDs := make([]uint, 0, len(recommendations))
	recommendationOfficialDishIDs := make([]uint, 0, len(recommendations))
	for _, recommendation := range recommendations {
		if recommendation.DishID != nil {
			recommendationByDishID[*recommendation.DishID] = recommendationReference{
				ID: recommendation.ID, SortOrder: recommendation.SortOrder,
			}
			recommendationDishIDs = append(recommendationDishIDs, *recommendation.DishID)
		}
		if recommendation.OfficialDishID != nil {
			recommendationByOfficialDishID[*recommendation.OfficialDishID] = recommendationReference{
				ID: recommendation.ID, SortOrder: recommendation.SortOrder,
			}
			recommendationOfficialDishIDs = append(
				recommendationOfficialDishIDs, *recommendation.OfficialDishID,
			)
		}
	}
	if len(recommendationDishIDs) == 0 {
		recommendationDishIDs = []uint{0}
	}
	var dishes []orderfoodModel.UserDish
	if err := preloadDish(db).
		Where("status = ? AND category_id IN (?)",
			orderfoodModel.DishStatusUsable,
			db.Model(&orderfoodModel.ContentCategory{}).Select("id").Where("enabled = ?", true),
		).
		Where("(owner_id = ? OR discoverable = ? OR id IN ?)", userID, true, recommendationDishIDs).
		Find(&dishes).Error; err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "recall suggestion candidate dishes")
	}
	excludedRecommendations := make(map[string]struct{}, len(input.ExcludedRecommendationIDs))
	for _, id := range input.ExcludedRecommendationIDs {
		excludedRecommendations[id] = struct{}{}
	}
	requestedTags := make(map[string]struct{}, len(input.Tags))
	for _, tag := range input.Tags {
		requestedTags[normalizeSuggestionText(tag)] = struct{}{}
	}
	candidates := make([]suggestionRecallCandidate, 0, len(dishes))
	for _, dish := range dishes {
		normalizedName := normalizeSuggestionText(dish.Name)
		if _, excluded := excludedNames[normalizedName]; excluded {
			continue
		}
		if len(dish.Ingredients) == 0 || len(dish.Steps) == 0 {
			continue
		}
		if dishViolatesHardConditions(dish, input.HardConditions) {
			continue
		}
		reference, recommended := recommendationByDishID[dish.ID]
		if recommended {
			if _, excluded := excludedRecommendations[reference.ID]; excluded {
				continue
			}
		}
		score := 0.50
		sourceLabel := "大家的菜品"
		if dish.OwnerID == userID {
			score = 0.68
			sourceLabel = "我的菜品"
		}
		var recommendationID *string
		recommendationSort := 1 << 30
		if recommended {
			score = 0.70
			sourceLabel = "平台推荐"
			id := reference.ID
			recommendationID = &id
			recommendationSort = reference.SortOrder
		}
		matches := 0
		for _, tag := range dish.Tags {
			if _, matched := requestedTags[normalizeSuggestionText(tag.Name)]; matched {
				matches++
			}
			if _, matched := requestedTags[normalizeSuggestionText(tag.PublicID)]; matched {
				matches++
			}
		}
		if matches > 2 {
			matches = 2
		}
		score += float64(matches) * 0.15
		if score > 1 {
			score = 1
		}
		copied := dish.OwnerID == userID
		var copiedDishID *uint
		var copiedPublicID *string
		if copied {
			id := dish.ID
			publicID := dish.PublicID
			copiedDishID = &id
			copiedPublicID = &publicID
		}
		if recommendationID != nil && !copied {
			var existingCopy orderfoodModel.RecommendationCopy
			copyErr := db.First(
				&existingCopy,
				"recommendation_id = ? AND user_id = ?",
				*recommendationID,
				userID,
			).Error
			if copyErr == nil {
				var copiedDish orderfoodModel.UserDish
				if err := db.First(&copiedDish, existingCopy.CopiedDishID).Error; err != nil {
					return nil, appErrors.FrontInternal.Wrap(err, "load existing suggestion recommendation copy")
				}
				copied = true
				id := copiedDish.ID
				publicID := copiedDish.PublicID
				copiedDishID = &id
				copiedPublicID = &publicID
			} else if !errors.Is(copyErr, gorm.ErrRecordNotFound) {
				return nil, appErrors.FrontInternal.Wrap(copyErr, "check suggestion recommendation copy")
			}
		}
		candidates = append(candidates, suggestionRecallCandidate{
			Dish: dish, RecommendationID: recommendationID, SourceLabel: sourceLabel,
			Score: score, RecommendationSort: recommendationSort, Copied: copied,
			CopiedDishID: copiedDishID, CopiedPublicID: copiedPublicID,
		})
	}
	if len(recommendationOfficialDishIDs) > 0 {
		var officialDishes []orderfoodModel.OfficialDish
		if err := preloadOfficialDish(db).
			Where("id IN ? AND status = ?", recommendationOfficialDishIDs, orderfoodModel.DishStatusUsable).
			Find(&officialDishes).Error; err != nil {
			return nil, appErrors.FrontInternal.Wrap(err, "recall official suggestion candidate dishes")
		}
		for _, officialDish := range officialDishes {
			normalizedName := normalizeSuggestionText(officialDish.Name)
			if _, excluded := excludedNames[normalizedName]; excluded {
				continue
			}
			reference, exists := recommendationByOfficialDishID[officialDish.ID]
			if !exists {
				continue
			}
			if _, excluded := excludedRecommendations[reference.ID]; excluded {
				continue
			}
			converted := officialDishAsUserDish(officialDish)
			if dishViolatesHardConditions(converted, input.HardConditions) {
				continue
			}
			matches := 0
			for _, tag := range officialDish.Tags {
				if _, matched := requestedTags[normalizeSuggestionText(tag.Name)]; matched {
					matches++
				}
				if _, matched := requestedTags[normalizeSuggestionText(tag.PublicID)]; matched {
					matches++
				}
			}
			if matches > 2 {
				matches = 2
			}
			score := 0.70 + float64(matches)*0.15
			if score > 1 {
				score = 1
			}
			recommendationID := reference.ID
			officialDishID := officialDish.ID
			var copiedDishID *uint
			var copiedPublicID *string
			var existingCopy orderfoodModel.RecommendationCopy
			copyErr := db.First(
				&existingCopy,
				"recommendation_id = ? AND user_id = ?",
				recommendationID,
				userID,
			).Error
			copied := copyErr == nil
			if copyErr == nil {
				var copiedDish orderfoodModel.UserDish
				if err := db.First(&copiedDish, existingCopy.CopiedDishID).Error; err != nil {
					return nil, appErrors.FrontInternal.Wrap(err, "load copied official suggestion dish")
				}
				copiedDishID = &copiedDish.ID
				copiedPublicID = &copiedDish.PublicID
			} else if !errors.Is(copyErr, gorm.ErrRecordNotFound) {
				return nil, appErrors.FrontInternal.Wrap(copyErr, "check official suggestion recommendation copy")
			}
			candidates = append(candidates, suggestionRecallCandidate{
				Dish: converted, OfficialDishID: &officialDishID,
				RecommendationID: &recommendationID, SourceLabel: "平台推荐",
				Score: score, RecommendationSort: reference.SortOrder, Copied: copied,
				CopiedDishID: copiedDishID, CopiedPublicID: copiedPublicID,
			})
		}
	}
	sort.SliceStable(candidates, func(left int, right int) bool {
		if candidates[left].Score != candidates[right].Score {
			return candidates[left].Score > candidates[right].Score
		}
		if candidates[left].RecommendationSort != candidates[right].RecommendationSort {
			return candidates[left].RecommendationSort < candidates[right].RecommendationSort
		}
		return candidates[left].Dish.UpdatedAt.After(candidates[right].Dish.UpdatedAt)
	})
	seenNames := make(map[string]struct{}, len(candidates))
	deduplicated := make([]suggestionRecallCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		name := normalizeSuggestionText(candidate.Dish.Name)
		if _, exists := seenNames[name]; exists {
			continue
		}
		seenNames[name] = struct{}{}
		deduplicated = append(deduplicated, candidate)
	}
	return deduplicated, nil
}

func officialDishAsUserDish(dish orderfoodModel.OfficialDish) orderfoodModel.UserDish {
	coverURL := dish.CoverURL
	ingredients := make([]orderfoodModel.DishIngredient, 0, len(dish.Ingredients))
	for _, ingredient := range dish.Ingredients {
		ingredients = append(ingredients, orderfoodModel.DishIngredient{
			Name: ingredient.Name, Quantity: ingredient.Quantity,
			UnitID: ingredient.UnitID, Unit: ingredient.Unit,
			Note: ingredient.Note, SortOrder: ingredient.SortOrder,
		})
	}
	steps := make([]orderfoodModel.DishStep, 0, len(dish.Steps))
	for _, step := range dish.Steps {
		steps = append(steps, orderfoodModel.DishStep{
			Description: step.Description, ImageURL: step.ImageURL, SortOrder: step.SortOrder,
		})
	}
	return orderfoodModel.UserDish{
		GVA_MODEL: dish.GVA_MODEL, PublicID: dish.PublicID,
		CoverFileID: dish.CoverFileID, CoverURL: &coverURL,
		Name: dish.Name, CategoryID: dish.CategoryID, Category: dish.Category,
		Status: dish.Status, SourceType: orderfoodModel.SourceTypeOfficialCopy,
		SourceLocked: true, Description: dish.Description, Serving: dish.Serving,
		MediaReviewStatus: orderfoodModel.MediaReviewNotRequired,
		Tags:              dish.Tags, Ingredients: ingredients, Steps: steps,
		Version: int(dish.Version),
	}
}

func suggestionSnapshotFromCandidate(
	candidate suggestionRecallCandidate,
) suggestionDishSnapshot {
	detail := dishResponse(candidate.Dish)
	dishID := candidate.Dish.PublicID
	if candidate.CopiedPublicID != nil {
		dishID = *candidate.CopiedPublicID
	}
	snapshotID := orderfoodModel.NewID()
	return suggestionDishSnapshot{
		Dish: frontResponse.SuggestionDish{
			ID: snapshotID, Source: "library", SourceLabel: candidate.SourceLabel,
			RecommendationID: candidate.RecommendationID, DishID: &dishID,
			Name: candidate.Dish.Name, Category: candidate.Dish.Category.Name,
			Tags: detail.Tags, CoverURL: stringValue(candidate.Dish.CoverURL),
			Serving: candidate.Dish.Serving, Description: stringValue(candidate.Dish.Description),
			Ingredients: detail.Ingredients, Steps: detail.Steps, Copied: candidate.Copied,
		},
		SourceDishID: func() *uint {
			if candidate.OfficialDishID != nil {
				return nil
			}
			return &candidate.Dish.ID
		}(),
		SourceOfficialDishID: candidate.OfficialDishID,
		CopiedDishID:         candidate.CopiedDishID,
		NormalizedName:       normalizeSuggestionText(candidate.Dish.Name),
	}
}

func (service *AssistService) loadSuggestionValidationContext(
	ctx context.Context,
	withCatalog bool,
) (suggestionValidationContext, error) {
	db := service.database().WithContext(ctx)
	result := suggestionValidationContext{
		Categories:          make(map[string]orderfoodModel.ContentCategory),
		Tags:                make(map[string]orderfoodModel.ContentTag),
		Units:               make(map[string]orderfoodModel.ContentUnit),
		StandardDishes:      make(map[string]orderfoodModel.StandardDishIndex),
		StandardIngredients: make(map[string]orderfoodModel.StandardIngredient),
	}
	var categories []orderfoodModel.ContentCategory
	var tags []orderfoodModel.ContentTag
	var units []orderfoodModel.ContentUnit
	if err := db.Where("enabled = ?", true).Find(&categories).Error; err != nil {
		return result, appErrors.FrontInternal.Wrap(err, "load suggestion categories")
	}
	if err := db.Where("enabled = ?", true).Find(&tags).Error; err != nil {
		return result, appErrors.FrontInternal.Wrap(err, "load suggestion tags")
	}
	if err := db.Where("enabled = ?", true).Find(&units).Error; err != nil {
		return result, appErrors.FrontInternal.Wrap(err, "load suggestion units")
	}
	for _, category := range categories {
		result.Categories[normalizeSuggestionText(category.Name)] = category
		result.Categories[normalizeSuggestionText(category.PublicID)] = category
	}
	for _, tag := range tags {
		result.Tags[normalizeSuggestionText(tag.Name)] = tag
		result.Tags[normalizeSuggestionText(tag.PublicID)] = tag
	}
	for _, unit := range units {
		result.Units[normalizeSuggestionText(unit.Name)] = unit
		result.Units[normalizeSuggestionText(unit.PublicID)] = unit
	}
	if !withCatalog {
		return result, nil
	}
	var ingredients []orderfoodModel.StandardIngredient
	if err := db.Where("enabled = ?", true).Find(&ingredients).Error; err != nil {
		return result, appErrors.FrontInternal.Wrap(err, "load standard ingredient lexicon")
	}
	for _, ingredient := range ingredients {
		result.StandardIngredients[normalizeSuggestionText(ingredient.Name)] = ingredient
		for _, alias := range decodeSuggestionAliases(ingredient.AliasesJSON) {
			result.StandardIngredients[normalizeSuggestionText(alias)] = ingredient
		}
	}
	var dishes []orderfoodModel.StandardDishIndex
	if err := db.Where("enabled = ?", true).
		Preload("Category").
		Preload("Ingredients.Ingredient").
		Find(&dishes).Error; err != nil {
		return result, appErrors.FrontInternal.Wrap(err, "load standard dish indexes")
	}
	for _, dish := range dishes {
		result.StandardDishes[normalizeSuggestionText(dish.Name)] = dish
		for _, alias := range decodeSuggestionAliases(dish.AliasesJSON) {
			result.StandardDishes[normalizeSuggestionText(alias)] = dish
		}
	}
	return result, nil
}

func decodeSuggestionAliases(raw datatypes.JSON) []string {
	var aliases []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &aliases)
	}
	return aliases
}

func suggestionPromptCatalog(
	validation suggestionValidationContext,
	withCatalog bool,
) map[string]interface{} {
	categories := make([]string, 0)
	seenCategories := make(map[uint]struct{})
	for _, category := range validation.Categories {
		if _, exists := seenCategories[category.ID]; exists {
			continue
		}
		seenCategories[category.ID] = struct{}{}
		categories = append(categories, category.Name)
	}
	tags := make([]string, 0)
	seenTags := make(map[uint]struct{})
	for _, tag := range validation.Tags {
		if _, exists := seenTags[tag.ID]; exists {
			continue
		}
		seenTags[tag.ID] = struct{}{}
		tags = append(tags, tag.Name)
	}
	units := make([]string, 0)
	seenUnits := make(map[uint]struct{})
	for _, unit := range validation.Units {
		if _, exists := seenUnits[unit.ID]; exists {
			continue
		}
		seenUnits[unit.ID] = struct{}{}
		units = append(units, unit.Name)
	}
	sort.Strings(categories)
	sort.Strings(tags)
	sort.Strings(units)
	result := map[string]interface{}{
		"categories": categories,
		"tags":       tags,
		"units":      units,
	}
	if withCatalog {
		standardDishes := make([]map[string]interface{}, 0)
		seenDishes := make(map[string]struct{})
		for _, dish := range validation.StandardDishes {
			if _, exists := seenDishes[dish.ID]; exists {
				continue
			}
			seenDishes[dish.ID] = struct{}{}
			ingredients := make([]map[string]interface{}, 0, len(dish.Ingredients))
			for _, relation := range dish.Ingredients {
				ingredients = append(ingredients, map[string]interface{}{
					"name": relation.Ingredient.Name, "required": relation.Required,
				})
			}
			standardDishes = append(standardDishes, map[string]interface{}{
				"name": dish.Name, "aliases": decodeSuggestionAliases(dish.AliasesJSON),
				"cuisine": dish.Cuisine, "category": dish.Category.Name,
				"ingredients": ingredients,
			})
		}
		sort.Slice(standardDishes, func(left int, right int) bool {
			return fmt.Sprint(standardDishes[left]["name"]) < fmt.Sprint(standardDishes[right]["name"])
		})
		result["standardDishes"] = standardDishes
	}
	return result
}

func (service *AssistService) generateSuggestionDishes(
	ctx context.Context,
	input frontRequest.MealSuggestionInput,
	targetCount int,
	excludedNames map[string]struct{},
	candidates []suggestionRecallCandidate,
	catalogValidationEnabled bool,
) ([]suggestionDishSnapshot, string, error) {
	return service.generateSuggestionDishesWithUsage(
		ctx, input, "", targetCount, excludedNames, candidates, catalogValidationEnabled,
	)
}

// generateSuggestionDishesWithUsage 生成候选菜并关联AI调用记录。
func (service *AssistService) generateSuggestionDishesWithUsage(
	ctx context.Context,
	input frontRequest.MealSuggestionInput,
	usageID string,
	targetCount int,
	excludedNames map[string]struct{},
	candidates []suggestionRecallCandidate,
	catalogValidationEnabled bool,
) ([]suggestionDishSnapshot, string, error) {
	validation, err := service.loadSuggestionValidationContext(ctx, catalogValidationEnabled)
	if err != nil {
		return nil, "", err
	}
	candidatePrompt := make([]map[string]interface{}, 0, len(candidates))
	for index, candidate := range candidates {
		if index == 10 {
			break
		}
		candidatePrompt = append(candidatePrompt, map[string]interface{}{
			"name": candidate.Dish.Name, "category": candidate.Dish.Category.Name,
			"tags": dishSummary(candidate.Dish).Tags, "score": candidate.Score,
		})
	}
	constraints := map[string]interface{}{
		"preferredTags": input.Tags, "timeLimitMinutes": input.TimeLimitMinutes,
		"explicitPreference": input.Preference,
		"preferenceProfile":  input.PreferenceContext, "hardConditions": input.HardConditions,
		"targetDishCount": targetCount, "excludedNames": excludedNames,
		"catalogValidationEnabled": catalogValidationEnabled,
	}
	candidateJSON, _ := json.Marshal(candidatePrompt)
	constraintJSON, _ := json.Marshal(constraints)
	metadataJSON, _ := json.Marshal(suggestionPromptCatalog(validation, catalogValidationEnabled))
	attempts := 1
	if catalogValidationEnabled {
		attempts += orderfoodService.SuggestionValidationRetryCount
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		var output suggestionModelOutput
		previousFailure := ""
		if attempt > 1 {
			previousFailure = "首次结果未通过，请重新生成并严格遵守 JSON 字段、可选值和菜品索引约束。"
		}
		runErr := service.runJSONCapability(ctx, orderfoodModel.AICapabilityMealSuggest, map[string]string{
			"candidate_dishes": string(candidateJSON),
			"constraints":      string(constraintJSON),
			"servings":         fmt.Sprintf("%d", input.People),
			"metadata":         string(metadataJSON),
			"target_count":     fmt.Sprintf("%d", targetCount),
			"attempt":          fmt.Sprintf("%d", attempt),
			"previous_failure": previousFailure,
		}, &output, usageID)
		if runErr != nil {
			lastErr = runErr
			continue
		}
		dishes, validationErr := validateGeneratedSuggestion(
			output, input.People, targetCount, excludedNames, input.HardConditions,
			validation, catalogValidationEnabled,
		)
		if validationErr != nil {
			lastErr = validationErr
			continue
		}
		reason := strings.TrimSpace(output.Reason)
		if reason == "" {
			reason = fmt.Sprintf("按 %d 人用餐规模生成的菜品组合", input.People)
		}
		return dishes, reason, nil
	}
	if lastErr == nil {
		lastErr = appErrors.FrontResultInvalid.DefaultMsg()
	}
	return nil, "", lastErr
}

func validateGeneratedSuggestion(
	output suggestionModelOutput,
	people int,
	targetCount int,
	excludedNames map[string]struct{},
	hardConditions []string,
	validation suggestionValidationContext,
	catalogValidationEnabled bool,
) ([]suggestionDishSnapshot, error) {
	if len(output.Dishes) != targetCount ||
		(strings.TrimSpace(output.Reason) != "" && !suggestionTextLengthBetween(output.Reason, 1, 300)) {
		return nil, appErrors.FrontResultInvalid.DefaultMsg()
	}
	seenNames := make(map[string]struct{}, len(output.Dishes))
	result := make([]suggestionDishSnapshot, 0, len(output.Dishes))
	for _, generated := range output.Dishes {
		name := strings.TrimSpace(generated.Name)
		normalizedName := normalizeSuggestionText(name)
		if !suggestionTextLengthBetween(name, 1, 40) ||
			!suggestionTextLengthBetween(generated.Cuisine, 1, 80) ||
			!suggestionTextLengthBetween(generated.Category, 1, 80) ||
			!suggestionTextLengthBetween(generated.Description, 1, 180) ||
			generated.Serving < 1 || generated.Serving > 20 ||
			len(generated.Tags) > 3 ||
			len(generated.Ingredients) < 1 || len(generated.Ingredients) > 30 ||
			len(generated.Steps) < 1 || len(generated.Steps) > 20 {
			return nil, appErrors.FrontResultInvalid.DefaultMsg()
		}
		if _, exists := excludedNames[normalizedName]; exists {
			return nil, appErrors.FrontResultInvalid.DefaultMsg()
		}
		if _, exists := seenNames[normalizedName]; exists {
			return nil, appErrors.FrontResultInvalid.DefaultMsg()
		}
		searchable := normalizedName +
			normalizeSuggestionText(generated.Cuisine) +
			normalizeSuggestionText(strings.Join(generated.Tags, ""))
		for _, ingredient := range generated.Ingredients {
			searchable += normalizeSuggestionText(ingredient.Name)
		}
		for _, condition := range hardConditions {
			keyword := hardConditionKeyword(condition)
			if keyword != "" && strings.Contains(searchable, keyword) {
				return nil, appErrors.FrontResultInvalid.DefaultMsg()
			}
		}
		seenNames[normalizedName] = struct{}{}
		category, exists := validation.Categories[normalizeSuggestionText(generated.Category)]
		if !exists {
			return nil, appErrors.FrontResultInvalid.DefaultMsg()
		}
		tags := make([]string, 0, len(generated.Tags))
		seenTagIDs := make(map[uint]struct{}, len(generated.Tags))
		for _, generatedTag := range generated.Tags {
			tag, exists := validation.Tags[normalizeSuggestionText(generatedTag)]
			if !exists {
				return nil, appErrors.FrontResultInvalid.DefaultMsg()
			}
			if _, duplicated := seenTagIDs[tag.ID]; duplicated {
				return nil, appErrors.FrontResultInvalid.DefaultMsg()
			}
			seenTagIDs[tag.ID] = struct{}{}
			tags = append(tags, tag.Name)
		}
		stepsText := strings.Builder{}
		steps := make([]frontResponse.DishStep, 0, len(generated.Steps))
		for index, generatedStep := range generated.Steps {
			text := strings.TrimSpace(generatedStep.Text)
			if !suggestionTextLengthBetween(text, 1, 500) {
				return nil, appErrors.FrontResultInvalid.DefaultMsg()
			}
			stepsText.WriteString(normalizeSuggestionText(text))
			steps = append(steps, frontResponse.DishStep{
				ID: orderfoodModel.NewID(), Text: text, SortOrder: index + 1,
			})
		}
		ingredients := make([]frontResponse.Ingredient, 0, len(generated.Ingredients))
		generatedStandardIngredientIDs := make(map[string]struct{})
		for index, generatedIngredient := range generated.Ingredients {
			ingredientName := strings.TrimSpace(generatedIngredient.Name)
			amount := strings.TrimSpace(generatedIngredient.Amount)
			unit, unitExists := validation.Units[normalizeSuggestionText(generatedIngredient.Unit)]
			if !suggestionTextLengthBetween(ingredientName, 1, 30) ||
				!suggestionTextLengthBetween(amount, 1, 20) ||
				!unitExists ||
				(generatedIngredient.Note != nil &&
					len([]rune(strings.TrimSpace(*generatedIngredient.Note))) > 80) ||
				!strings.Contains(stepsText.String(), normalizeSuggestionText(ingredientName)) {
				return nil, appErrors.FrontResultInvalid.DefaultMsg()
			}
			if catalogValidationEnabled {
				standardIngredient, exists :=
					validation.StandardIngredients[normalizeSuggestionText(ingredientName)]
				if !exists {
					return nil, appErrors.FrontResultInvalid.DefaultMsg()
				}
				generatedStandardIngredientIDs[standardIngredient.ID] = struct{}{}
				ingredientName = standardIngredient.Name
			}
			var note *string
			if generatedIngredient.Note != nil {
				value := strings.TrimSpace(*generatedIngredient.Note)
				if value != "" {
					note = &value
				}
			}
			ingredients = append(ingredients, frontResponse.Ingredient{
				ID: orderfoodModel.NewID(), Name: ingredientName, Amount: amount,
				Unit: unit.Name, Note: note, SortOrder: index + 1,
			})
		}
		cuisine := strings.TrimSpace(generated.Cuisine)
		if catalogValidationEnabled {
			standardDish, exists := validation.StandardDishes[normalizedName]
			if !exists ||
				normalizeSuggestionText(standardDish.Cuisine) != normalizeSuggestionText(cuisine) ||
				standardDish.CategoryID != category.ID {
				return nil, appErrors.FrontResultInvalid.DefaultMsg()
			}
			allowed := make(map[string]struct{}, len(standardDish.Ingredients))
			for _, relation := range standardDish.Ingredients {
				allowed[relation.IngredientID] = struct{}{}
				if relation.Required {
					if _, included := generatedStandardIngredientIDs[relation.IngredientID]; !included {
						return nil, appErrors.FrontResultInvalid.DefaultMsg()
					}
				}
			}
			for ingredientID := range generatedStandardIngredientIDs {
				if _, allowedIngredient := allowed[ingredientID]; !allowedIngredient {
					return nil, appErrors.FrontResultInvalid.DefaultMsg()
				}
			}
			name = standardDish.Name
			normalizedName = normalizeSuggestionText(name)
			cuisine = standardDish.Cuisine
		}
		description := strings.TrimSpace(generated.Description)
		snapshotID := orderfoodModel.NewID()
		result = append(result, suggestionDishSnapshot{
			Dish: frontResponse.SuggestionDish{
				ID: snapshotID, Source: "generated", SourceLabel: "生成建议",
				Name: name, Category: category.Name, Cuisine: cuisine, Tags: tags,
				Serving: generated.Serving, Description: description,
				Ingredients: ingredients, Steps: steps, Copied: false,
			},
			NormalizedName: normalizedName,
		})
	}
	_ = people
	return result, nil
}

func (service *AssistService) persistSuggestion(
	ctx context.Context,
	userID string,
	usageID string,
	people int,
	source string,
	sourceLabel string,
	reason string,
	dishes []suggestionDishSnapshot,
) (orderfoodModel.FrontMealSuggestion, []frontResponse.SuggestionDish, error) {
	if len(dishes) == 0 {
		return orderfoodModel.FrontMealSuggestion{}, nil, appErrors.FrontResultInvalid.DefaultMsg()
	}
	now := service.now()
	ids := make([]string, 0, len(dishes))
	responseDishes := make([]frontResponse.SuggestionDish, 0, len(dishes))
	for _, snapshot := range dishes {
		ids = append(ids, snapshot.Dish.ID)
		responseDishes = append(responseDishes, snapshot.Dish)
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return orderfoodModel.FrontMealSuggestion{}, nil,
			appErrors.FrontInternal.Wrap(err, "encode suggestion dish ids")
	}
	row := orderfoodModel.FrontMealSuggestion{
		ID: orderfoodModel.NewID(), UserID: userID, Source: source, SourceLabel: sourceLabel,
		Reason: strings.TrimSpace(reason), People: people, DishIDsJSON: datatypes.JSON(idsJSON),
		UsageID: usageID, CreatedAt: now, UpdatedAt: now,
	}
	err = service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "save meal suggestion")
		}
		for index, snapshot := range dishes {
			encoded, err := json.Marshal(snapshot.Dish)
			if err != nil {
				return appErrors.FrontInternal.Wrap(err, "encode meal suggestion dish snapshot")
			}
			snapshotRow := orderfoodModel.FrontMealSuggestionDish{
				ID: snapshot.Dish.ID, SuggestionID: row.ID, SortOrder: index + 1,
				Source: snapshot.Dish.Source, SourceLabel: snapshot.Dish.SourceLabel,
				SourceDishID: snapshot.SourceDishID, SourceOfficialDishID: snapshot.SourceOfficialDishID,
				RecommendationID: snapshot.Dish.RecommendationID,
				Name:             snapshot.Dish.Name, NormalizedName: snapshot.NormalizedName,
				SnapshotJSON: datatypes.JSON(encoded), CopiedDishID: snapshot.CopiedDishID,
				CreatedAt: now,
			}
			if err := tx.Create(&snapshotRow).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "save meal suggestion dish snapshot")
			}
		}
		return nil
	})
	if err != nil {
		return orderfoodModel.FrontMealSuggestion{}, nil, err
	}
	return row, responseDishes, nil
}

func suggestionResponse(
	row orderfoodModel.FrontMealSuggestion,
	dishes []frontResponse.SuggestionDish,
	usage frontResponse.FeatureUsageResult,
) frontResponse.MealSuggestion {
	return frontResponse.MealSuggestion{
		ID: row.ID, Source: row.Source, SourceLabel: row.SourceLabel, Reason: row.Reason,
		People: row.People, SuggestedDishCount: len(dishes), Dish: dishes[0],
		Dishes: dishes, Usage: usage,
	}
}

// CreateSuggestion 按人数和偏好生成推荐菜结果。
func (service *AssistService) CreateSuggestion(
	ctx context.Context,
	userID string,
	input frontRequest.MealSuggestionInput,
) (frontResponse.MealSuggestion, error) {
	var user orderfoodModel.MiniAppUser
	if err := service.database().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return frontResponse.MealSuggestion{}, appErrors.FrontInternal.Wrap(err, "load suggestion user")
	}
	if user.CheckinDayCount < 7 {
		return frontResponse.MealSuggestion{}, appErrors.FrontFeatureLocked.DefaultMsg()
	}
	input, err := service.prepareSuggestionPreferences(ctx, userID, input)
	if err != nil {
		return frontResponse.MealSuggestion{}, err
	}
	usage, _, err := service.engagement().BeginFeatureUsage(
		ctx,
		userID,
		orderfoodModel.AICapabilityMealSuggest,
		input.UseFreeQuota,
	)
	if err != nil {
		return frontResponse.MealSuggestion{}, err
	}
	targetCount := suggestionTargetCount(input.People)
	excludedNames, err := service.excludedSuggestionNames(ctx, userID, input)
	if err != nil {
		_, _ = service.engagement().FinishFeatureUsage(ctx, usage, false)
		return frontResponse.MealSuggestion{}, err
	}
	candidates, err := service.recallSuggestionCandidates(ctx, userID, input, excludedNames)
	if err != nil {
		_, _ = service.engagement().FinishFeatureUsage(ctx, usage, false)
		return frontResponse.MealSuggestion{}, err
	}

	var dishes []suggestionDishSnapshot
	reason := fmt.Sprintf("按 %d 人用餐规模，从现有菜品中组合", input.People)
	source := "library"
	sourceLabel := "现有菜品"
	maxScore := 0.0
	if len(candidates) > 0 {
		maxScore = candidates[0].Score
	}
	// 优先使用现有菜品库的高质量召回；样本不足时才调用生成能力，减少成本并提高结果可验证性。
	if len(candidates) >= 3 && maxScore >= 0.65 && len(candidates) >= targetCount {
		dishes = make([]suggestionDishSnapshot, 0, targetCount)
		for _, candidate := range candidates[:targetCount] {
			dishes = append(dishes, suggestionSnapshotFromCandidate(candidate))
		}
	} else {
		catalogValidationEnabled, policyErr :=
			orderfoodService.CurrentSuggestionCatalogValidation(ctx, service.database())
		if policyErr != nil {
			_, _ = service.engagement().FinishFeatureUsage(ctx, usage, false)
			return frontResponse.MealSuggestion{}, policyErr
		}
		dishes, reason, err = service.generateSuggestionDishesWithUsage(
			ctx,
			input,
			usage.ID,
			targetCount,
			excludedNames,
			candidates,
			catalogValidationEnabled,
		)
		if err != nil {
			_, finishErr := service.engagement().FinishFeatureUsage(ctx, usage, false)
			if finishErr != nil {
				return frontResponse.MealSuggestion{}, finishErr
			}
			return frontResponse.MealSuggestion{}, err
		}
		source = "generated"
		sourceLabel = "生成建议"
	}

	row, responseDishes, err := service.persistSuggestion(
		ctx, userID, usage.ID, input.People, source, sourceLabel, reason, dishes,
	)
	if err != nil {
		_, _ = service.engagement().FinishFeatureUsage(ctx, usage, false)
		return frontResponse.MealSuggestion{}, err
	}
	usageResult, finishErr := service.engagement().FinishFeatureUsage(ctx, usage, true)
	if finishErr != nil {
		// 计费状态未能落为成功时删除刚保存的结果，再按失败路径退款，避免用户看到不可追溯的成功结果。
		_ = service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("suggestion_id = ?", row.ID).
				Delete(&orderfoodModel.FrontMealSuggestionDish{}).Error; err != nil {
				return err
			}
			return tx.Delete(&row).Error
		})
		_, _ = service.engagement().FinishFeatureUsage(ctx, usage, false)
		return frontResponse.MealSuggestion{}, finishErr
	}
	if input.Preference != nil && strings.TrimSpace(*input.Preference) != "" {
		value := strings.TrimSpace(*input.Preference)
		settingType := "preference"
		switch {
		case strings.Contains(value, "过敏"):
			settingType = "allergy"
		case strings.Contains(value, "不吃"), strings.Contains(value, "忌"):
			settingType = "avoidance"
		case strings.Contains(value, "素食"), strings.Contains(value, "清真"):
			settingType = "dietary_restriction"
		}
		_ = service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return ServiceGroupApp.PreferenceService.queueEvidenceTx(
				ctx, tx, userID,
				orderfoodModel.PreferenceSourceExplicitSetting,
				digest(settingType+"\x00"+value),
				"positive", 4, 1,
				preferenceFacts{Settings: []preferenceSetting{{
					Type: settingType, Value: value, UpdatedAt: service.now(),
				}}},
				service.now(),
			)
		})
	}
	return suggestionResponse(row, responseDishes, usageResult), nil
}

// Suggestion 获取推荐菜结果详情。
func (service *AssistService) Suggestion(
	ctx context.Context,
	userID string,
	suggestionID string,
) (frontResponse.MealSuggestion, error) {
	var row orderfoodModel.FrontMealSuggestion
	if err := service.database().WithContext(ctx).First(&row, "id = ? AND user_id = ?", suggestionID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.MealSuggestion{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.MealSuggestion{}, appErrors.FrontInternal.Wrap(err, "load meal suggestion")
	}
	var snapshotRows []orderfoodModel.FrontMealSuggestionDish
	if err := service.database().WithContext(ctx).
		Where("suggestion_id = ?", row.ID).
		Order("sort_order asc").
		Find(&snapshotRows).Error; err != nil {
		return frontResponse.MealSuggestion{}, appErrors.FrontInternal.Wrap(err, "load meal suggestion snapshots")
	}
	dishes := make([]frontResponse.SuggestionDish, 0, len(snapshotRows))
	for _, snapshotRow := range snapshotRows {
		var snapshot frontResponse.SuggestionDish
		if err := json.Unmarshal(snapshotRow.SnapshotJSON, &snapshot); err != nil {
			return frontResponse.MealSuggestion{}, appErrors.FrontInternal.Wrap(err, "decode meal suggestion snapshot")
		}
		snapshot.ID = snapshotRow.ID
		snapshot.Copied = snapshotRow.CopiedDishID != nil || snapshot.Copied
		if snapshotRow.CopiedDishID != nil {
			var copied orderfoodModel.UserDish
			if err := service.database().WithContext(ctx).First(&copied, *snapshotRow.CopiedDishID).Error; err == nil {
				id := copied.PublicID
				snapshot.DishID = &id
			}
		}
		dishes = append(dishes, snapshot)
	}
	if len(dishes) == 0 {
		return frontResponse.MealSuggestion{}, appErrors.FrontNotFound.DefaultMsg()
	}
	var usage orderfoodModel.FrontFeatureUsage
	if err := service.database().WithContext(ctx).First(&usage, "id = ?", row.UsageID).Error; err != nil {
		return frontResponse.MealSuggestion{}, appErrors.FrontInternal.Wrap(err, "load suggestion usage")
	}
	var user orderfoodModel.MiniAppUser
	_ = service.database().WithContext(ctx).First(&user, "id = ?", userID).Error
	return suggestionResponse(row, dishes, frontResponse.FeatureUsageResult{
		UsageID: usage.ID, PointCost: usage.PointCost, BillingStatus: usage.BillingStatus,
		PointBalance: user.Points,
	}), nil
}

// CopySuggestion 将推荐结果中的菜品复制到个人菜品库。
func (service *AssistService) CopySuggestion(
	ctx context.Context,
	userID string,
	suggestionID string,
	suggestionDishID string,
) (frontResponse.Dish, bool, error) {
	var copiedPublicID string
	recorded := false
	preferenceEnabled := preferenceUpdatesEnabled(ctx)
	// 锁定推荐菜快照并复用已复制结果，使重复点击不会生成多个个人菜品副本。
	err := service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var suggestion orderfoodModel.FrontMealSuggestion
		if err := tx.First(&suggestion, "id = ? AND user_id = ?", suggestionID, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load suggestion for copy")
		}
		var snapshotRow orderfoodModel.FrontMealSuggestionDish
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&snapshotRow, "id = ? AND suggestion_id = ?", suggestionDishID, suggestionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			return appErrors.FrontInternal.Wrap(err, "load suggestion dish for copy")
		}
		if snapshotRow.CopiedDishID != nil {
			var copied orderfoodModel.UserDish
			if err := tx.First(&copied, *snapshotRow.CopiedDishID).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "load adopted suggestion dish")
			}
			copiedPublicID = copied.PublicID
			return nil
		}
		var snapshot frontResponse.SuggestionDish
		if err := json.Unmarshal(snapshotRow.SnapshotJSON, &snapshot); err != nil {
			return appErrors.FrontInternal.Wrap(err, "decode suggestion dish for copy")
		}
		var copied orderfoodModel.UserDish
		if snapshotRow.Source == "library" && snapshotRow.SourceDishID != nil {
			var source orderfoodModel.UserDish
			if err := preloadDish(tx).First(&source, *snapshotRow.SourceDishID).Error; err != nil {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			if source.OwnerID == userID {
				copied = source
			} else {
				var err error
				copied, err = service.copySuggestionLibraryDish(tx, userID, source, snapshotRow)
				if err != nil {
					return err
				}
			}
		} else if snapshotRow.Source == "library" && snapshotRow.SourceOfficialDishID != nil {
			var source orderfoodModel.OfficialDish
			if err := preloadOfficialDish(tx).
				First(&source, "id = ? AND status = ?", *snapshotRow.SourceOfficialDishID, orderfoodModel.DishStatusUsable).Error; err != nil {
				return appErrors.FrontNotFound.DefaultMsg()
			}
			var err error
			copied, err = service.copySuggestionOfficialDish(tx, userID, source, snapshotRow)
			if err != nil {
				return err
			}
		} else if snapshotRow.Source == "generated" {
			var err error
			copied, err = service.copyGeneratedSuggestionDish(tx, userID, snapshot, snapshotRow)
			if err != nil {
				return err
			}
		} else {
			return appErrors.FrontStateConflict.DefaultMsg()
		}
		now := service.now()
		if err := tx.Model(&snapshotRow).Updates(map[string]interface{}{
			"copied_dish_id": copied.ID,
			"adopted_at":     now,
		}).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "record suggestion dish adoption")
		}
		if snapshotRow.RecommendationID != nil {
			copyRecord := orderfoodModel.RecommendationCopy{
				ID: orderfoodModel.NewID(), RecommendationID: *snapshotRow.RecommendationID,
				UserID: userID, CopiedDishID: copied.ID, CreatedAt: now,
			}
			if err := tx.Where(
				"recommendation_id = ? AND user_id = ?",
				copyRecord.RecommendationID,
				userID,
			).FirstOrCreate(&copyRecord).Error; err != nil {
				return appErrors.FrontInternal.Wrap(err, "record suggestion recommendation copy")
			}
		}
		if preferenceEnabled {
			if err := ServiceGroupApp.PreferenceService.queueDishEvidenceTx(
				ctx, tx, userID,
				orderfoodModel.PreferenceSourceRecommendationAdopted, snapshotRow.ID,
				"positive", 3, copied.ID, now,
			); err != nil {
				return err
			}
		}
		copiedPublicID = copied.PublicID
		recorded = true
		return nil
	})
	if err != nil {
		return frontResponse.Dish{}, false, err
	}
	dish, err := service.content().Dish(ctx, userID, copiedPublicID)
	return dish, recorded, err
}

func (service *AssistService) copySuggestionOfficialDish(
	tx *gorm.DB,
	userID string,
	source orderfoodModel.OfficialDish,
	snapshotRow orderfoodModel.FrontMealSuggestionDish,
) (orderfoodModel.UserDish, error) {
	coverURL := source.CoverURL
	operationSourceType := "meal_suggestion"
	copied := orderfoodModel.UserDish{
		PublicID: orderfoodModel.NewID(), OwnerID: userID,
		CoverFileID: source.CoverFileID, CoverURL: &coverURL,
		Name: source.Name, CategoryID: source.CategoryID,
		Status: orderfoodModel.DishStatusUsable, Discoverable: false,
		SourceType: orderfoodModel.SourceTypeMealSuggestionCopy, SourceLocked: true,
		Description: source.Description, Serving: source.Serving,
		MediaReviewStatus:   orderfoodModel.MediaReviewNotRequired,
		OperationSourceType: &operationSourceType, OperationSourceID: &snapshotRow.ID,
		DirectSourceOfficialDishID: &source.ID, RootSourceOfficialDishID: &source.ID,
		ChainDepth: 1, Version: 1,
	}
	if err := tx.Create(&copied).Error; err != nil {
		return copied, appErrors.FrontInternal.Wrap(err, "copy official suggestion dish")
	}
	for _, ingredient := range source.Ingredients {
		if err := tx.Create(&orderfoodModel.DishIngredient{
			DishID: copied.ID, Name: ingredient.Name, Quantity: ingredient.Quantity,
			UnitID: ingredient.UnitID, Note: ingredient.Note, SortOrder: ingredient.SortOrder,
		}).Error; err != nil {
			return copied, appErrors.FrontInternal.Wrap(err, "copy official suggestion ingredients")
		}
	}
	for _, step := range source.Steps {
		if err := tx.Create(&orderfoodModel.DishStep{
			DishID: copied.ID, Description: step.Description,
			ImageURL: step.ImageURL, SortOrder: step.SortOrder,
		}).Error; err != nil {
			return copied, appErrors.FrontInternal.Wrap(err, "copy official suggestion steps")
		}
	}
	for _, tag := range source.Tags {
		if err := tx.Create(&orderfoodModel.UserDishTag{
			DishID: copied.ID, TagID: tag.ID, CreatedAt: service.now(),
		}).Error; err != nil {
			return copied, appErrors.FrontInternal.Wrap(err, "copy official suggestion tags")
		}
	}
	return copied, nil
}

func (service *AssistService) copySuggestionLibraryDish(
	tx *gorm.DB,
	userID string,
	source orderfoodModel.UserDish,
	snapshotRow orderfoodModel.FrontMealSuggestionDish,
) (orderfoodModel.UserDish, error) {
	copied := source
	copied.ID = 0
	copied.CreatedAt = time.Time{}
	copied.UpdatedAt = time.Time{}
	copied.DeletedAt = gorm.DeletedAt{}
	copied.PublicID = orderfoodModel.NewID()
	copied.OwnerID = userID
	copied.Owner = orderfoodModel.MiniAppUser{}
	copied.Discoverable = false
	copied.DiscoverableAt = nil
	copied.SourceLocked = true
	copied.SourceType = orderfoodModel.SourceTypeMealSuggestionCopy
	operationSourceType := "meal_suggestion"
	copied.OperationSourceType = &operationSourceType
	copied.OperationSourceID = &snapshotRow.ID
	copied.DirectSourceDishID = &source.ID
	rootID := source.ID
	if source.RootSourceDishID != nil {
		rootID = *source.RootSourceDishID
	}
	copied.RootSourceDishID = &rootID
	originalAuthorID := source.OwnerID
	copied.OriginalAuthorID = &originalAuthorID
	copied.ChainDepth = source.ChainDepth + 1
	copied.Ingredients = nil
	copied.Steps = nil
	copied.Tags = nil
	if err := tx.Create(&copied).Error; err != nil {
		return copied, appErrors.FrontInternal.Wrap(err, "copy suggestion library dish")
	}
	for _, ingredient := range source.Ingredients {
		ingredient.ID = 0
		ingredient.CreatedAt = time.Time{}
		ingredient.UpdatedAt = time.Time{}
		ingredient.DeletedAt = gorm.DeletedAt{}
		ingredient.DishID = copied.ID
		ingredient.Unit = nil
		if err := tx.Create(&ingredient).Error; err != nil {
			return copied, appErrors.FrontInternal.Wrap(err, "copy suggestion library ingredients")
		}
	}
	for _, step := range source.Steps {
		step.ID = 0
		step.CreatedAt = time.Time{}
		step.UpdatedAt = time.Time{}
		step.DeletedAt = gorm.DeletedAt{}
		step.DishID = copied.ID
		if err := tx.Create(&step).Error; err != nil {
			return copied, appErrors.FrontInternal.Wrap(err, "copy suggestion library steps")
		}
	}
	for _, tag := range source.Tags {
		if err := tx.Create(&orderfoodModel.UserDishTag{
			DishID: copied.ID, TagID: tag.ID, CreatedAt: service.now(),
		}).Error; err != nil {
			return copied, appErrors.FrontInternal.Wrap(err, "copy suggestion library tags")
		}
	}
	return copied, nil
}

func (service *AssistService) copyGeneratedSuggestionDish(
	tx *gorm.DB,
	userID string,
	snapshot frontResponse.SuggestionDish,
	snapshotRow orderfoodModel.FrontMealSuggestionDish,
) (orderfoodModel.UserDish, error) {
	var category orderfoodModel.ContentCategory
	if err := tx.First(
		&category,
		"name = ? AND enabled = ?",
		snapshot.Category,
		true,
	).Error; err != nil {
		return orderfoodModel.UserDish{}, appErrors.FrontStateConflict.Wrap(
			err, "resolve generated suggestion category",
		)
	}
	var description *string
	if value := strings.TrimSpace(snapshot.Description); value != "" {
		description = &value
	}
	operationSourceType := "meal_suggestion"
	row := orderfoodModel.UserDish{
		PublicID: orderfoodModel.NewID(), OwnerID: userID,
		CoverFileID: "", CoverURL: nil, Name: strings.TrimSpace(snapshot.Name),
		CategoryID: category.ID, Status: orderfoodModel.DishStatusDraft,
		Discoverable: false, SourceType: orderfoodModel.SourceTypeGeneratedSuggestion,
		SourceLocked: false, Description: description, Serving: snapshot.Serving,
		MediaReviewStatus:   orderfoodModel.MediaReviewNotRequired,
		OperationSourceType: &operationSourceType, OperationSourceID: &snapshotRow.ID,
		Version: 1,
	}
	if err := tx.Create(&row).Error; err != nil {
		return row, appErrors.FrontInternal.Wrap(err, "save generated suggestion dish")
	}
	for _, snapshotIngredient := range snapshot.Ingredients {
		var unit orderfoodModel.ContentUnit
		if err := tx.First(
			&unit,
			"name = ? AND enabled = ?",
			snapshotIngredient.Unit,
			true,
		).Error; err != nil {
			return row, appErrors.FrontStateConflict.Wrap(err, "resolve generated suggestion unit")
		}
		quantity := strings.TrimSpace(snapshotIngredient.Amount)
		if err := tx.Create(&orderfoodModel.DishIngredient{
			DishID: row.ID, Name: strings.TrimSpace(snapshotIngredient.Name),
			Quantity: &quantity, UnitID: &unit.ID, Note: snapshotIngredient.Note,
			SortOrder: snapshotIngredient.SortOrder,
		}).Error; err != nil {
			return row, appErrors.FrontInternal.Wrap(err, "save generated suggestion ingredient")
		}
	}
	for _, snapshotStep := range snapshot.Steps {
		if err := tx.Create(&orderfoodModel.DishStep{
			DishID: row.ID, Description: strings.TrimSpace(snapshotStep.Text),
			SortOrder: snapshotStep.SortOrder,
		}).Error; err != nil {
			return row, appErrors.FrontInternal.Wrap(err, "save generated suggestion step")
		}
	}
	seenTags := make(map[uint]struct{}, len(snapshot.Tags))
	for _, snapshotTag := range snapshot.Tags {
		var tag orderfoodModel.ContentTag
		if err := tx.First(&tag, "name = ? AND enabled = ?", snapshotTag, true).Error; err != nil {
			return row, appErrors.FrontStateConflict.Wrap(err, "resolve generated suggestion tag")
		}
		if _, duplicated := seenTags[tag.ID]; duplicated {
			continue
		}
		seenTags[tag.ID] = struct{}{}
		if err := tx.Create(&orderfoodModel.UserDishTag{
			DishID: row.ID, TagID: tag.ID, CreatedAt: service.now(),
		}).Error; err != nil {
			return row, appErrors.FrontInternal.Wrap(err, "save generated suggestion tag")
		}
	}
	return row, nil
}

// Feedback 记录用户对增强功能结果的采用或不采用反馈。
func (service *AssistService) Feedback(
	ctx context.Context,
	userID string,
	kind string,
	id string,
	action string,
) error {
	var model interface{}
	if kind == "suggestion" {
		model = &orderfoodModel.FrontMealSuggestion{}
	} else {
		model = &orderfoodModel.FrontPrepPlan{}
	}
	preferenceEnabled := kind == "suggestion" && preferenceUpdatesEnabled(ctx)
	now := service.now()
	return service.database().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(model).
			Where("id = ? AND user_id = ?", id, userID).
			Updates(map[string]interface{}{"feedback": action, "updated_at": now})
		if result.Error != nil {
			return appErrors.FrontInternal.Wrap(result.Error, "save feature feedback")
		}
		if result.RowsAffected == 0 {
			return appErrors.FrontNotFound.DefaultMsg()
		}
		if !preferenceEnabled {
			return nil
		}
		var rows []orderfoodModel.FrontMealSuggestionDish
		if err := tx.Where("suggestion_id = ?", id).
			Order("sort_order asc").Find(&rows).Error; err != nil {
			return appErrors.FrontInternal.Wrap(err, "load suggestion feedback evidence")
		}
		facts := preferenceFacts{}
		for _, row := range rows {
			var dish frontResponse.SuggestionDish
			if err := json.Unmarshal(row.SnapshotJSON, &dish); err != nil {
				return appErrors.FrontInternal.Wrap(err, "decode suggestion feedback evidence")
			}
			facts.DishNames = append(facts.DishNames, dish.Name)
			facts.Tags = append(facts.Tags, dish.Tags...)
			if dish.Cuisine != "" {
				facts.Cuisines = append(facts.Cuisines, dish.Cuisine)
			}
			for _, ingredient := range dish.Ingredients {
				facts.Ingredients = append(facts.Ingredients, ingredient.Name)
			}
		}
		sourceType := orderfoodModel.PreferenceSourceSkipForNow
		weight := 1.5
		if action == "regenerated" {
			sourceType = orderfoodModel.PreferenceSourceReshuffle
			weight = 0.75
		}
		return ServiceGroupApp.PreferenceService.queueEvidenceTx(
			ctx, tx, userID, sourceType, id,
			"negative", weight, 1, facts, now,
		)
	})
}

// PrepQuote 获取备菜顺序生成所需积分报价。
func (service *AssistService) PrepQuote(
	ctx context.Context,
	userID string,
	mealID string,
) (int, int64, bool, error) {
	if _, err := ServiceGroupApp.MealService.loadMealRow(ctx, userID, mealID); err != nil {
		return 0, 0, false, err
	}
	_, cost, _, _, err := service.engagement().featurePolicy(
		ctx,
		orderfoodModel.AICapabilityPrepSequence,
	)
	if err != nil {
		return 0, 0, false, err
	}
	var user orderfoodModel.MiniAppUser
	if err := service.database().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return 0, 0, false, appErrors.FrontInternal.Wrap(err, "load prep quote balance")
	}
	return cost, user.Points, user.Points >= int64(cost), nil
}

type prepSourceStep struct {
	ID             string `json:"id"`
	DishSnapshotID string `json:"dishSnapshotId"`
	DishName       string `json:"dishName"`
	Instruction    string `json:"instruction"`
	SortOrder      int    `json:"sortOrder"`
}

type prepModelStep struct {
	SourceStepID          string   `json:"sourceStepId"`
	ParallelSourceStepIDs []string `json:"parallelSourceStepIds"`
}

type prepModelOutput struct {
	EstimatedMinutes int             `json:"estimatedMinutes"`
	Steps            []prepModelStep `json:"steps"`
}

type prepExcludedPlan struct {
	PlanID  string   `json:"planId"`
	StepIDs []string `json:"stepIds"`
}

type prepMealDishPrompt struct {
	DishSnapshotID string `json:"dishSnapshotId"`
	Name           string `json:"name"`
	FinalServings  int    `json:"finalServings"`
}

func prepSourcesFromFinals(
	finals []orderfoodModel.FrontMealFinalDish,
) ([]prepSourceStep, error) {
	sources := make([]prepSourceStep, 0)
	seen := make(map[string]struct{})
	for _, final := range finals {
		if len(final.Steps) == 0 {
			return nil, appErrors.FrontResultInvalid.New("confirmed meal is missing its step snapshot")
		}
		var snapshots []mealDishStepSnapshot
		if err := json.Unmarshal(final.Steps, &snapshots); err != nil {
			return nil, appErrors.FrontResultInvalid.Wrap(err, "decode confirmed meal step snapshot")
		}
		for _, snapshot := range snapshots {
			id := strings.TrimSpace(snapshot.ID)
			instruction := strings.TrimSpace(snapshot.Description)
			if id == "" || instruction == "" {
				return nil, appErrors.FrontResultInvalid.New("confirmed meal contains an invalid step snapshot")
			}
			if _, duplicated := seen[id]; duplicated {
				return nil, appErrors.FrontResultInvalid.New("confirmed meal contains duplicate step snapshots")
			}
			seen[id] = struct{}{}
			sources = append(sources, prepSourceStep{
				ID: id, DishSnapshotID: final.ID, DishName: final.Name,
				Instruction: instruction, SortOrder: snapshot.SortOrder,
			})
		}
	}
	if len(sources) == 0 {
		return nil, appErrors.FrontResultInvalid.New("confirmed meal has no preparation steps")
	}
	return sources, nil
}

func sameStepOrder(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validatePrepModelOutput(
	output prepModelOutput,
	sources []prepSourceStep,
	excluded []prepExcludedPlan,
) ([]frontResponse.PrepStep, error) {
	if len(output.Steps) != len(sources) {
		return nil, appErrors.FrontResultInvalid.New("preparation result omitted or added steps")
	}
	minutesUpperBound := len(sources) * 60
	if minutesUpperBound > 1440 {
		minutesUpperBound = 1440
	}
	if output.EstimatedMinutes < len(sources) || output.EstimatedMinutes > minutesUpperBound {
		return nil, appErrors.FrontResultInvalid.New("preparation estimate is outside the allowed range")
	}
	sourceByID := make(map[string]prepSourceStep, len(sources))
	for _, source := range sources {
		sourceByID[source.ID] = source
	}
	seen := make(map[string]struct{}, len(sources))
	orderedIDs := make([]string, 0, len(sources))
	steps := make([]frontResponse.PrepStep, 0, len(sources))
	for index, modelStep := range output.Steps {
		sourceID := strings.TrimSpace(modelStep.SourceStepID)
		source, exists := sourceByID[sourceID]
		if !exists {
			return nil, appErrors.FrontResultInvalid.New("preparation result referenced an unknown step")
		}
		if _, duplicated := seen[sourceID]; duplicated {
			return nil, appErrors.FrontResultInvalid.New("preparation result repeated a step")
		}
		seen[sourceID] = struct{}{}
		orderedIDs = append(orderedIDs, sourceID)

		parallelSeen := make(map[string]struct{}, len(modelStep.ParallelSourceStepIDs))
		parallelActions := make([]string, 0, len(modelStep.ParallelSourceStepIDs))
		for _, parallelIDValue := range modelStep.ParallelSourceStepIDs {
			parallelID := strings.TrimSpace(parallelIDValue)
			parallel, valid := sourceByID[parallelID]
			if !valid || parallelID == sourceID {
				return nil, appErrors.FrontResultInvalid.New("preparation result contains an invalid parallel step")
			}
			if _, duplicated := parallelSeen[parallelID]; duplicated {
				return nil, appErrors.FrontResultInvalid.New("preparation result repeated a parallel step")
			}
			parallelSeen[parallelID] = struct{}{}
			parallelActions = append(
				parallelActions,
				strings.TrimSpace(parallel.DishName)+"："+parallel.Instruction,
			)
		}
		steps = append(steps, frontResponse.PrepStep{
			ID: source.ID, Order: index + 1, Title: source.DishName,
			Instruction: source.Instruction, ParallelActions: parallelActions,
			CompletionHint: "完成「" + source.Instruction + "」后继续",
		})
	}
	for _, previous := range excluded {
		if sameStepOrder(orderedIDs, previous.StepIDs) {
			return nil, appErrors.FrontResultInvalid.New("preparation result repeated an excluded plan")
		}
	}
	return steps, nil
}

func (service *AssistService) excludedPrepPlans(
	ctx context.Context,
	userID string,
	mealID string,
	ids []string,
) ([]prepExcludedPlan, error) {
	if len(ids) == 0 {
		return []prepExcludedPlan{}, nil
	}
	unique := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, appErrors.FrontBadRequest.DefaultMsg()
		}
		if _, duplicated := unique[id]; duplicated {
			return nil, appErrors.FrontBadRequest.DefaultMsg()
		}
		unique[id] = struct{}{}
	}
	var rows []orderfoodModel.FrontPrepPlan
	if err := service.database().WithContext(ctx).
		Where("user_id = ? AND meal_id = ? AND id IN ?", userID, mealID, ids).
		Find(&rows).Error; err != nil {
		return nil, appErrors.FrontInternal.Wrap(err, "load excluded preparation plans")
	}
	if len(rows) != len(ids) {
		return nil, appErrors.FrontBadRequest.DefaultMsg()
	}
	rowByID := make(map[string]orderfoodModel.FrontPrepPlan, len(rows))
	for _, row := range rows {
		rowByID[row.ID] = row
	}
	result := make([]prepExcludedPlan, 0, len(ids))
	for _, id := range ids {
		row := rowByID[id]
		var steps []frontResponse.PrepStep
		if err := json.Unmarshal(row.StepsJSON, &steps); err != nil || len(steps) == 0 {
			return nil, appErrors.FrontBadRequest.DefaultMsg()
		}
		stepIDs := make([]string, 0, len(steps))
		for _, step := range steps {
			if strings.TrimSpace(step.ID) == "" {
				return nil, appErrors.FrontBadRequest.DefaultMsg()
			}
			stepIDs = append(stepIDs, step.ID)
		}
		result = append(result, prepExcludedPlan{PlanID: id, StepIDs: stepIDs})
	}
	return result, nil
}

func (service *AssistService) failPrepUsage(
	ctx context.Context,
	usage orderfoodModel.FrontFeatureUsage,
	cause error,
) error {
	if _, err := service.engagement().FinishFeatureUsage(ctx, usage, false); err != nil {
		return appErrors.FrontInternal.Wrap(err, "refund failed preparation usage")
	}
	return cause
}

// GeneratePrepPlan 根据已确认饭局生成备菜顺序。
func (service *AssistService) GeneratePrepPlan(
	ctx context.Context,
	userID string,
	input frontRequest.PrepPlanInput,
) (frontResponse.PrepPlan, error) {
	meal, err := ServiceGroupApp.MealService.loadMealRow(ctx, userID, input.MealID)
	if err != nil {
		return frontResponse.PrepPlan{}, err
	}
	if meal.Status != orderfoodModel.MealConfirmed {
		return frontResponse.PrepPlan{}, appErrors.FrontStateConflict.DefaultMsg()
	}
	var finals []orderfoodModel.FrontMealFinalDish
	if err := service.database().WithContext(ctx).
		Where("meal_id = ?", input.MealID).
		Order("created_at asc, id asc").
		Find(&finals).Error; err != nil {
		return frontResponse.PrepPlan{}, appErrors.FrontInternal.Wrap(err, "load prep final dishes")
	}
	sources, err := prepSourcesFromFinals(finals)
	if err != nil {
		return frontResponse.PrepPlan{}, err
	}
	excluded, err := service.excludedPrepPlans(
		ctx, userID, input.MealID, input.ExcludedPlanIDs,
	)
	if err != nil {
		return frontResponse.PrepPlan{}, err
	}

	promptDishes := make([]prepMealDishPrompt, 0, len(finals))
	for _, final := range finals {
		promptDishes = append(promptDishes, prepMealDishPrompt{
			DishSnapshotID: final.ID, Name: final.Name, FinalServings: final.FinalServings,
		})
	}
	mealSnapshotJSON, err := json.Marshal(struct {
		MealID        string               `json:"mealId"`
		Dishes        []prepMealDishPrompt `json:"dishes"`
		ExcludedPlans []prepExcludedPlan   `json:"excludedPlans"`
	}{
		MealID: input.MealID, Dishes: promptDishes, ExcludedPlans: excluded,
	})
	if err != nil {
		return frontResponse.PrepPlan{}, appErrors.FrontInternal.Wrap(err, "encode preparation meal snapshot")
	}
	stepContractJSON, err := json.Marshal(struct {
		Rules []string         `json:"rules"`
		Steps []prepSourceStep `json:"steps"`
	}{
		Rules: []string{
			"只输出 estimatedMinutes 和 steps",
			"steps 中每个 sourceStepId 必须来自输入且恰好出现一次",
			"parallelSourceStepIds 只能引用输入中的其他步骤，可为空数组",
			"不得输出或改写菜名、做法、食材、设备、等待分钟数或完成文案",
		},
		Steps: sources,
	})
	if err != nil {
		return frontResponse.PrepPlan{}, appErrors.FrontInternal.Wrap(err, "encode preparation step contract")
	}
	totalServings := 0
	for _, final := range finals {
		totalServings += final.FinalServings
	}
	usage, _, err := service.engagement().BeginFeatureUsage(
		ctx, userID, orderfoodModel.AICapabilityPrepSequence, true,
	)
	if err != nil {
		return frontResponse.PrepPlan{}, err
	}
	var modelOutput prepModelOutput
	err = service.runJSONCapability(
		ctx,
		orderfoodModel.AICapabilityPrepSequence,
		map[string]string{
			"meal_snapshot": string(mealSnapshotJSON),
			"dish_steps":    string(stepContractJSON),
			"servings":      fmt.Sprintf("%d", totalServings),
		},
		&modelOutput,
		usage.ID,
	)
	if err != nil {
		return frontResponse.PrepPlan{}, service.failPrepUsage(ctx, usage, err)
	}
	steps, err := validatePrepModelOutput(modelOutput, sources, excluded)
	if err != nil {
		return frontResponse.PrepPlan{}, service.failPrepUsage(ctx, usage, err)
	}
	encoded, err := json.Marshal(steps)
	if err != nil {
		return frontResponse.PrepPlan{}, service.failPrepUsage(
			ctx, usage, appErrors.FrontInternal.Wrap(err, "encode preparation result"),
		)
	}
	now := service.now()
	row := orderfoodModel.FrontPrepPlan{
		ID: orderfoodModel.NewID(), UserID: userID, MealID: input.MealID,
		EstimatedMinutes: modelOutput.EstimatedMinutes,
		StepsJSON:        datatypes.JSON(encoded), UsageID: usage.ID,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := service.database().WithContext(ctx).Create(&row).Error; err != nil {
		return frontResponse.PrepPlan{}, service.failPrepUsage(
			ctx, usage, appErrors.FrontInternal.Wrap(err, "save prep plan"),
		)
	}
	usageResult, err := service.engagement().FinishFeatureUsage(ctx, usage, true)
	if err != nil {
		_ = service.database().WithContext(ctx).Delete(&row).Error
		return frontResponse.PrepPlan{}, service.failPrepUsage(ctx, usage, err)
	}
	return frontResponse.PrepPlan{
		ID: row.ID, MealID: row.MealID, EstimatedMinutes: row.EstimatedMinutes,
		Steps: steps, Usage: usageResult,
	}, nil
}

// PrepPlan 获取备菜顺序详情。
func (service *AssistService) PrepPlan(
	ctx context.Context,
	userID string,
	planID string,
) (frontResponse.PrepPlan, error) {
	var row orderfoodModel.FrontPrepPlan
	if err := service.database().WithContext(ctx).First(&row, "id = ? AND user_id = ?", planID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return frontResponse.PrepPlan{}, appErrors.FrontNotFound.DefaultMsg()
		}
		return frontResponse.PrepPlan{}, appErrors.FrontInternal.Wrap(err, "load prep plan")
	}
	var steps []frontResponse.PrepStep
	if err := json.Unmarshal(row.StepsJSON, &steps); err != nil {
		return frontResponse.PrepPlan{}, appErrors.FrontInternal.Wrap(err, "decode prep plan")
	}
	var usage orderfoodModel.FrontFeatureUsage
	var user orderfoodModel.MiniAppUser
	_ = service.database().WithContext(ctx).First(&usage, "id = ?", row.UsageID).Error
	_ = service.database().WithContext(ctx).First(&user, "id = ?", userID).Error
	return frontResponse.PrepPlan{
		ID: row.ID, MealID: row.MealID, EstimatedMinutes: row.EstimatedMinutes, Steps: steps,
		Usage: frontResponse.FeatureUsageResult{
			UsageID: usage.ID, PointCost: usage.PointCost, BillingStatus: usage.BillingStatus,
			PointBalance: user.Points,
		},
	}, nil
}

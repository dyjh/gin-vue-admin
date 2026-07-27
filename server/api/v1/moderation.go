package v1

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/gin-gonic/gin"
)

const moderationTestImageMaxBytes = 10 << 20

// ModerationApi 提供图片审核配置接口处理能力。
type ModerationApi struct {
	service *orderfoodService.ModerationService
}

// NewModerationApi 创建图片审核API实例。
func NewModerationApi(service *orderfoodService.ModerationService) *ModerationApi {
	return &ModerationApi{service: service}
}

// GetConfig 获取当前图片审核配置
// @Tags OrderFoodModerationConfig
// @Summary 获取当前图片审核配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=orderfoodResponse.ModerationConfig,msg=string} "获取成功"
// @Router /orderfood/moderation-config [get]
func (api *ModerationApi) GetConfig(c *gin.Context) {
	var query orderfoodRequest.ModerationEmptyQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.ModerationConfig
	result, err := api.service.GetConfig(c.Request.Context(), actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdateConfig 保存并立即应用图片审核配置
// @Tags OrderFoodModerationConfig
// @Summary 保存并立即应用图片审核配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.ModerationConfigUpdateInput true "图片审核配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.ModerationConfig,msg=string} "操作成功"
// @Router /orderfood/moderation-config [put]
func (api *ModerationApi) UpdateConfig(c *gin.Context) {
	header, ok := bindModerationIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.ModerationConfigUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.Region = strings.TrimSpace(body.Region)
	body.Endpoint = strings.TrimSpace(body.Endpoint)
	body.ServiceCode = strings.TrimSpace(body.ServiceCode)
	if body.CredentialRef != nil {
		value := strings.TrimSpace(*body.CredentialRef)
		body.CredentialRef = &value
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	result, replayed, err := api.service.UpdateConfig(
		c.Request.Context(),
		body,
		actor,
		header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setModerationReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// TestConnection 测试当前图片审核配置连接
// @Tags OrderFoodModerationConfig
// @Summary 测试当前图片审核配置连接
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.ModerationConfigTestInput true "可选并发校验参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.ModerationConnectionTestResult,msg=string} "操作成功"
// @Router /orderfood/moderation-config/connection-tests [post]
func (api *ModerationApi) TestConnection(c *gin.Context) {
	header, ok := bindModerationIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.ModerationConfigTestInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	result, replayed, err := api.service.TestConnection(
		c.Request.Context(),
		body,
		actor,
		header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setModerationReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// TestImage 测试图片审核结果映射
// @Tags OrderFoodModerationConfig
// @Summary 测试图片审核结果映射
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param expectedVersion formData integer false "当前配置版本"
// @Param file formData file true "临时测试图片"
// @Success 200 {object} response.Response{data=orderfoodResponse.ModerationImageTestResult,msg=string} "操作成功"
// @Router /orderfood/moderation-config/moderation-tests [post]
func (api *ModerationApi) TestImage(c *gin.Context) {
	header, ok := bindModerationIdempotencyHeader(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		moderationTestImageMaxBytes+(1<<20),
	)
	var form orderfoodRequest.ModerationConfigTestInput
	if !bindAndVerify(c, &form, c.ShouldBind) {
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader.Size <= 0 || fileHeader.Size > moderationTestImageMaxBytes {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	defer file.Close()
	imageBytes, err := io.ReadAll(io.LimitReader(file, moderationTestImageMaxBytes+1))
	if err != nil || len(imageBytes) == 0 || len(imageBytes) > moderationTestImageMaxBytes {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	contentType := http.DetectContentType(imageBytes)
	if !strings.HasPrefix(contentType, "image/") {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	if _, _, err := image.DecodeConfig(bytes.NewReader(imageBytes)); err != nil {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	result, replayed, err := api.service.TestImage(
		c.Request.Context(),
		form,
		orderfoodService.ModerationProviderImage{
			Bytes: imageBytes, ContentType: contentType,
		},
		actor,
		header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setModerationReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

func (api *ModerationApi) available(c *gin.Context) bool {
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return false
	}
	return true
}

func bindModerationIdempotencyHeader(
	c *gin.Context,
) (orderfoodRequest.IdempotencyHeader, bool) {
	var header orderfoodRequest.IdempotencyHeader
	if !bindAndVerify(c, &header, c.ShouldBindHeader) {
		return orderfoodRequest.IdempotencyHeader{}, false
	}
	header.Key = strings.TrimSpace(header.Key)
	if header.Key == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.IdempotencyHeader{}, false
	}
	return header, true
}

func setModerationReplayHeader(c *gin.Context, replayed bool) {
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
}

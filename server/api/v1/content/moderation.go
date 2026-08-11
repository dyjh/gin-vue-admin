package content

import (
	"bytes"
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/content"
	"github.com/gin-gonic/gin"
)

const moderationTestImageMaxBytes = 10 << 20

// ModerationApi 提供图片审核配置接口处理能力。
type ModerationApi struct{}

// GetConfig 获取当前图片审核配置
// @Tags OrderFoodModerationConfig
// @Summary 获取当前图片审核配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=contentResponse.ModerationConfig,msg=string} "获取成功"
// @Router /orderfood/moderation-config [get]
func (api *ModerationApi) GetConfig(c *gin.Context) {
	var query contentRequest.ModerationEmptyQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result contentResponse.ModerationConfig
	result, err := moderationService.GetConfig(c.Request.Context(), actor)
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
// @Param data body contentRequest.ModerationConfigUpdateInput true "图片审核配置"
// @Success 200 {object} response.Response{data=contentResponse.ModerationConfig,msg=string} "操作成功"
// @Router /orderfood/moderation-config [put]
func (api *ModerationApi) UpdateConfig(c *gin.Context) {
	header, ok := bindModerationIdempotencyHeader(c)
	if !ok {
		return
	}
	var body contentRequest.ModerationConfigUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.Region = strings.TrimSpace(body.Region)
	body.Endpoint = strings.TrimSpace(body.Endpoint)
	body.ServiceCode = strings.TrimSpace(body.ServiceCode)
	if body.CredentialRef != nil {
		value := strings.TrimSpace(*body.CredentialRef)
		body.CredentialRef = &value
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := moderationService.UpdateConfig(
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
// @Param data body contentRequest.ModerationConfigTestInput true "可选并发校验参数"
// @Success 200 {object} response.Response{data=contentResponse.ModerationConnectionTestResult,msg=string} "操作成功"
// @Router /orderfood/moderation-config/connection-tests [post]
func (api *ModerationApi) TestConnection(c *gin.Context) {
	header, ok := bindModerationIdempotencyHeader(c)
	if !ok {
		return
	}
	var body contentRequest.ModerationConfigTestInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := moderationService.TestConnection(
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
// @Success 200 {object} response.Response{data=contentResponse.ModerationImageTestResult,msg=string} "操作成功"
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
	var form contentRequest.ModerationConfigTestInput
	if !apiCommon.BindAndVerify(c, &form, c.ShouldBind) {
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
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := moderationService.TestImage(
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

func bindModerationIdempotencyHeader(
	c *gin.Context,
) (commonRequest.IdempotencyHeader, bool) {
	var header commonRequest.IdempotencyHeader
	if !apiCommon.BindAndVerify(c, &header, c.ShouldBindHeader) {
		return commonRequest.IdempotencyHeader{}, false
	}
	header.Key = strings.TrimSpace(header.Key)
	if header.Key == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return commonRequest.IdempotencyHeader{}, false
	}
	return header, true
}

func setModerationReplayHeader(c *gin.Context, replayed bool) {
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
}

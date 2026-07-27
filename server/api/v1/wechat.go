package v1

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/gin-gonic/gin"
)

// WeChatConfigApi 提供微信小程序配置接口处理能力。
type WeChatConfigApi struct {
	service *orderfoodService.WeChatConfigService // 微信小程序配置服务
}

// NewWeChatConfigApi 创建微信小程序配置API实例。
func NewWeChatConfigApi(service *orderfoodService.WeChatConfigService) *WeChatConfigApi {
	return &WeChatConfigApi{service: service}
}

// GetConfig 获取微信小程序配置
// @Tags OrderFoodWeChatConfig
// @Summary 获取微信小程序配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=orderfoodResponse.WeChatConfig,msg=string} "获取成功"
// @Router /orderfood/wechat-config [get]
func (api *WeChatConfigApi) GetConfig(c *gin.Context) {
	var query orderfoodRequest.WeChatConfigQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.WeChatConfig
	result, err := api.service.GetConfig(c.Request.Context(), actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdateConfig 保存并立即应用微信小程序配置
// @Tags OrderFoodWeChatConfig
// @Summary 保存并立即应用微信小程序配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.WeChatConfigUpdateInput true "微信小程序配置"
// @Success 200 {object} response.Response{data=orderfoodResponse.WeChatConfig,msg=string} "操作成功"
// @Router /orderfood/wechat-config [put]
func (api *WeChatConfigApi) UpdateConfig(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.WeChatConfigUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.AppID = strings.TrimSpace(body.AppID)
	body.Reason = strings.TrimSpace(body.Reason)
	if body.AppSecret != nil {
		secret := strings.TrimSpace(*body.AppSecret)
		body.AppSecret = &secret
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.WeChatConfig
	result, replayed, err := api.service.UpdateConfig(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// available 检查微信小程序配置API是否已注入业务服务。
func (api *WeChatConfigApi) available(c *gin.Context) bool {
	if api != nil && api.service != nil {
		return true
	}
	response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
	return false
}

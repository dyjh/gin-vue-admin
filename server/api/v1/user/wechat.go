package user

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	userRequest "github.com/dyjh/order-food-mini-app/server/model/user/request"
	userResponse "github.com/dyjh/order-food-mini-app/server/model/user/response"
	"github.com/gin-gonic/gin"
)

// WeChatConfigApi 提供微信小程序配置接口处理能力。
type WeChatConfigApi struct{}

// GetConfig 获取微信小程序配置
// @Tags OrderFoodWeChatConfig
// @Summary 获取微信小程序配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=userResponse.WeChatConfig,msg=string} "获取成功"
// @Router /orderfood/wechat-config [get]
func (api *WeChatConfigApi) GetConfig(c *gin.Context) {
	var query userRequest.WeChatConfigQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result userResponse.WeChatConfig
	result, err := weChatConfigService.GetConfig(c.Request.Context(), actor)
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
// @Param data body userRequest.WeChatConfigUpdateInput true "微信小程序配置"
// @Success 200 {object} response.Response{data=userResponse.WeChatConfig,msg=string} "操作成功"
// @Router /orderfood/wechat-config [put]
func (api *WeChatConfigApi) UpdateConfig(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body userRequest.WeChatConfigUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.AppID = strings.TrimSpace(body.AppID)
	body.Reason = strings.TrimSpace(body.Reason)
	if body.AppSecret != nil {
		secret := strings.TrimSpace(*body.AppSecret)
		body.AppSecret = &secret
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result userResponse.WeChatConfig
	result, replayed, err := weChatConfigService.UpdateConfig(
		c.Request.Context(),
		actor,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

package engagement

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	engagementRequest "github.com/dyjh/order-food-mini-app/server/model/engagement/request"
	engagementResponse "github.com/dyjh/order-food-mini-app/server/model/engagement/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/engagement"
	"github.com/gin-gonic/gin"
)

// SubscriptionApi 提供管理端消息中心接口处理能力。
type SubscriptionApi struct{}

// ListNotifications 分页查询站内通知投递
// @Tags OrderFoodMessageCenter
// @Summary 分页查询站内通知投递
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query engagementRequest.NotificationAdminListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]engagementResponse.NotificationAdminSummary},msg=string} "获取成功"
// @Router /orderfood/notifications [get]
func (api *SubscriptionApi) ListNotifications(c *gin.Context) {
	var query engagementRequest.NotificationAdminListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionNotificationRead) {
		return
	}
	var result response.Page[engagementResponse.NotificationAdminSummary]
	result, err := subscriptionService.ListNotifications(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetNotification 获取站内通知投递详情
// @Tags OrderFoodMessageCenter
// @Summary 获取站内通知投递详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param notificationId path string true "站内通知 ID"
// @Success 200 {object} response.Response{data=engagementResponse.NotificationAdminDetail,msg=string} "获取成功"
// @Router /orderfood/notifications/{notificationId} [get]
func (api *SubscriptionApi) GetNotification(c *gin.Context) {
	var path engagementRequest.NotificationAdminPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionNotificationRead) {
		return
	}
	result, err := subscriptionService.GetNotification(c.Request.Context(), path.NotificationID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListSubscribeScenes 查询固定订阅消息场景
// @Tags OrderFoodMessageCenter
// @Summary 查询固定订阅消息场景
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=[]engagementResponse.SubscribeSceneSummary,msg=string} "获取成功"
// @Router /orderfood/subscribe-scenes [get]
func (api *SubscriptionApi) ListSubscribeScenes(c *gin.Context) {
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionSubscribeSceneRead) {
		return
	}
	result, err := subscriptionService.ListSubscribeScenes(c.Request.Context())
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetSubscribeScene 获取固定订阅消息场景详情
// @Tags OrderFoodMessageCenter
// @Summary 获取固定订阅消息场景详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param scene path string true "固定业务场景" Enums(meal_status)
// @Success 200 {object} response.Response{data=engagementResponse.SubscribeSceneDetail,msg=string} "获取成功"
// @Router /orderfood/subscribe-scenes/{scene} [get]
func (api *SubscriptionApi) GetSubscribeScene(c *gin.Context) {
	var path engagementRequest.SubscribeScenePath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSubscribeSceneRead) {
		return
	}
	result, err := subscriptionService.GetSubscribeScene(c.Request.Context(), path.Scene)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ConfigureSubscribeScene 配置固定场景的微信模板绑定
// @Tags OrderFoodMessageCenter
// @Summary 配置固定场景的微信模板绑定
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param scene path string true "固定业务场景" Enums(meal_status)
// @Param data body engagementRequest.SubscribeSceneConfigureInput true "模板绑定参数"
// @Success 200 {object} response.Response{data=engagementResponse.SubscribeSceneDetail,msg=string} "操作成功"
// @Router /orderfood/subscribe-scenes/{scene} [put]
func (api *SubscriptionApi) ConfigureSubscribeScene(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path engagementRequest.SubscribeScenePath
	var body engagementRequest.SubscribeSceneConfigureInput
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSubscribeSceneUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := subscriptionService.ConfigureSubscribeScene(
		c.Request.Context(), actor, header.Key, path.Scene, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateSubscribeSceneStatus 启用或停用固定订阅场景
// @Tags OrderFoodMessageCenter
// @Summary 启用或停用固定订阅场景
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param scene path string true "固定业务场景" Enums(meal_status)
// @Param data body engagementRequest.SubscribeSceneStatusInput true "启停参数"
// @Success 200 {object} response.Response{data=engagementResponse.SubscribeSceneDetail,msg=string} "操作成功"
// @Router /orderfood/subscribe-scenes/{scene}/status [put]
func (api *SubscriptionApi) UpdateSubscribeSceneStatus(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path engagementRequest.SubscribeScenePath
	var body engagementRequest.SubscribeSceneStatusInput
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSubscribeSceneStatus) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := subscriptionService.UpdateSubscribeSceneStatus(
		c.Request.Context(), actor, header.Key, path.Scene, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListSubscribeLogs 分页查询订阅消息发送记录
// @Tags OrderFoodMessageCenter
// @Summary 分页查询订阅消息发送记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query engagementRequest.SubscribeLogListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]engagementResponse.SubscribeLogSummary},msg=string} "获取成功"
// @Router /orderfood/subscribe-logs [get]
func (api *SubscriptionApi) ListSubscribeLogs(c *gin.Context) {
	var query engagementRequest.SubscribeLogListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSubscribeLogRead) {
		return
	}
	result, err := subscriptionService.ListSubscribeLogs(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetSubscribeLog 获取订阅消息发送详情
// @Tags OrderFoodMessageCenter
// @Summary 获取订阅消息发送详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param logId path string true "订阅消息发送记录 ID"
// @Success 200 {object} response.Response{data=engagementResponse.SubscribeLogDetail,msg=string} "获取成功"
// @Router /orderfood/subscribe-logs/{logId} [get]
func (api *SubscriptionApi) GetSubscribeLog(c *gin.Context) {
	var path engagementRequest.SubscribeLogPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionSubscribeLogRead) {
		return
	}
	result, err := subscriptionService.GetSubscribeLog(c.Request.Context(), path.LogID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

package v1

import (
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/gin-gonic/gin"
)

// SubscriptionApi 提供管理端消息中心接口处理能力。
type SubscriptionApi struct {
	service *orderfoodService.SubscriptionService
}

// NewSubscriptionApi 创建管理端消息中心API实例。
func NewSubscriptionApi(service *orderfoodService.SubscriptionService) *SubscriptionApi {
	return &SubscriptionApi{service: service}
}

// ListNotifications 分页查询站内通知投递
// @Tags OrderFoodMessageCenter
// @Summary 分页查询站内通知投递
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.NotificationAdminListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.NotificationAdminSummary},msg=string} "获取成功"
// @Router /orderfood/notifications [get]
func (api *SubscriptionApi) ListNotifications(c *gin.Context) {
	var query orderfoodRequest.NotificationAdminListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionNotificationRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.NotificationAdminSummary]
	result, err := api.service.ListNotifications(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.NotificationAdminDetail,msg=string} "获取成功"
// @Router /orderfood/notifications/{notificationId} [get]
func (api *SubscriptionApi) GetNotification(c *gin.Context) {
	var path orderfoodRequest.NotificationAdminPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!requirePermission(c, orderfoodService.PermissionNotificationRead) {
		return
	}
	result, err := api.service.GetNotification(c.Request.Context(), path.NotificationID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListSubscribeTemplates 分页查询订阅消息模板
// @Tags OrderFoodMessageCenter
// @Summary 分页查询订阅消息模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.SubscribeTemplateListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.SubscribeTemplateSummary},msg=string} "获取成功"
// @Router /orderfood/subscribe-templates [get]
func (api *SubscriptionApi) ListSubscribeTemplates(c *gin.Context) {
	var query orderfoodRequest.SubscribeTemplateListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeTemplateRead) {
		return
	}
	result, err := api.service.ListSubscribeTemplates(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetSubscribeTemplate 获取订阅消息模板详情
// @Tags OrderFoodMessageCenter
// @Summary 获取订阅消息模板详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param templateId path string true "订阅消息模板 ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.SubscribeTemplateDetail,msg=string} "获取成功"
// @Router /orderfood/subscribe-templates/{templateId} [get]
func (api *SubscriptionApi) GetSubscribeTemplate(c *gin.Context) {
	var path orderfoodRequest.SubscribeTemplatePath
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeTemplateRead) {
		return
	}
	result, err := api.service.GetSubscribeTemplate(c.Request.Context(), path.TemplateID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateSubscribeTemplate 创建订阅消息模板
// @Tags OrderFoodMessageCenter
// @Summary 创建订阅消息模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.SubscribeTemplateCreateInput true "订阅消息模板"
// @Success 200 {object} response.Response{data=orderfoodResponse.SubscribeTemplateDetail,msg=string} "操作成功"
// @Router /orderfood/subscribe-templates [post]
func (api *SubscriptionApi) CreateSubscribeTemplate(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.SubscribeTemplateCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeTemplateCreate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateSubscribeTemplate(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateSubscribeTemplate 编辑订阅消息模板
// @Tags OrderFoodMessageCenter
// @Summary 编辑订阅消息模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param templateId path string true "订阅消息模板 ID"
// @Param data body orderfoodRequest.SubscribeTemplateUpdateInput true "订阅消息模板"
// @Success 200 {object} response.Response{data=orderfoodResponse.SubscribeTemplateDetail,msg=string} "操作成功"
// @Router /orderfood/subscribe-templates/{templateId} [put]
func (api *SubscriptionApi) UpdateSubscribeTemplate(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path orderfoodRequest.SubscribeTemplatePath
	var body orderfoodRequest.SubscribeTemplateUpdateInput
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeTemplateUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateSubscribeTemplate(
		c.Request.Context(), actor, header.Key, path.TemplateID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateSubscribeTemplateStatus 启用或停用订阅消息模板
// @Tags OrderFoodMessageCenter
// @Summary 启用或停用订阅消息模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param templateId path string true "订阅消息模板 ID"
// @Param data body orderfoodRequest.SubscribeTemplateStatusInput true "启停参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.SubscribeTemplateDetail,msg=string} "操作成功"
// @Router /orderfood/subscribe-templates/{templateId}/status [put]
func (api *SubscriptionApi) UpdateSubscribeTemplateStatus(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path orderfoodRequest.SubscribeTemplatePath
	var body orderfoodRequest.SubscribeTemplateStatusInput
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeTemplateStatus) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateSubscribeTemplateStatus(
		c.Request.Context(), actor, header.Key, path.TemplateID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// DeleteSubscribeTemplate 删除未使用的订阅消息模板
// @Tags OrderFoodMessageCenter
// @Summary 删除未使用的订阅消息模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param templateId path string true "订阅消息模板 ID"
// @Param data body orderfoodRequest.SubscribeTemplateDeleteInput true "删除参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.DeletedResult,msg=string} "操作成功"
// @Router /orderfood/subscribe-templates/{templateId} [delete]
func (api *SubscriptionApi) DeleteSubscribeTemplate(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var path orderfoodRequest.SubscribeTemplatePath
	var body orderfoodRequest.SubscribeTemplateDeleteInput
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeTemplateDelete) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteSubscribeTemplate(
		c.Request.Context(), actor, header.Key, path.TemplateID, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListSubscribeLogs 分页查询订阅消息发送记录
// @Tags OrderFoodMessageCenter
// @Summary 分页查询订阅消息发送记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.SubscribeLogListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.SubscribeLogSummary},msg=string} "获取成功"
// @Router /orderfood/subscribe-logs [get]
func (api *SubscriptionApi) ListSubscribeLogs(c *gin.Context) {
	var query orderfoodRequest.SubscribeLogListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeLogRead) {
		return
	}
	result, err := api.service.ListSubscribeLogs(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.SubscribeLogDetail,msg=string} "获取成功"
// @Router /orderfood/subscribe-logs/{logId} [get]
func (api *SubscriptionApi) GetSubscribeLog(c *gin.Context) {
	var path orderfoodRequest.SubscribeLogPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) ||
		!requirePermission(c, orderfoodService.PermissionSubscribeLogRead) {
		return
	}
	result, err := api.service.GetSubscribeLog(c.Request.Context(), path.LogID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

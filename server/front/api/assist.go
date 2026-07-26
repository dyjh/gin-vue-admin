package api

import (
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	response "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// AssistApi 提供小程序增强功能接口处理能力。
type AssistApi struct{}

// DishExtraction 解析菜品文本并生成可编辑草稿
// @Tags 小程序增强功能
// @Summary 解析菜品文本并生成可编辑草稿
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.DishExtractionInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.DishExtractionResult,msg=string} "操作成功"
// @Router /miniapp/v1/assist/dish-extraction [post]
func (*AssistApi) DishExtraction(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.DishExtractionInput
	if !bindJSON(c, &input) {
		return
	}
	draft, usage, err := assistService.ExtractDish(c.Request.Context(), user.ID, input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.DishExtractionResult{
		DishDraft: draft, Usage: usage,
	}, c)
}

// DishCover 为菜品生成封面图片
// @Tags 小程序增强功能
// @Summary 为菜品生成封面图片
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.DishCoverInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.DishCoverResult,msg=string} "操作成功"
// @Router /miniapp/v1/assist/dish-covers [post]
func (*AssistApi) DishCover(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.DishCoverInput
	if !bindJSON(c, &input) {
		return
	}
	image, usage, err := assistService.CreateCover(
		c.Request.Context(), user.ID, input, middleware.GetRequestID(c),
	)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.DishCoverResult{
		FileID: image.FileID, URL: image.URL, Usage: usage,
	}, c)
}

// SuggestionStatus 获取推荐菜功能解锁、额度和积分状态
// @Tags 小程序增强功能
// @Summary 获取推荐菜功能解锁、额度和积分状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.SuggestionStatusResult,msg=string} "获取成功"
// @Router /miniapp/v1/meal-suggestions/status [get]
func (*AssistApi) SuggestionStatus(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	enabled, unlocked, days, unlockDays, quota, balance, cost, err :=
		assistService.SuggestionStatus(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.SuggestionStatusResult{
		Enabled: enabled, Unlocked: unlocked, CheckinDays: days,
		UnlockDays: unlockDays, FreeQuota: quota, PointBalance: balance, PointCost: cost,
	}, c)
}

// CreateSuggestion 按人数和偏好生成推荐菜结果
// @Tags 小程序增强功能
// @Summary 按人数和偏好生成推荐菜结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.MealSuggestionInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.MealSuggestion,msg=string} "操作成功"
// @Router /miniapp/v1/meal-suggestions [post]
func (*AssistApi) CreateSuggestion(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealSuggestionInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := assistService.CreateSuggestion(c.Request.Context(), user.ID, input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Suggestion 获取推荐菜结果详情
// @Tags 小程序增强功能
// @Summary 获取推荐菜结果详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param suggestionId path string true "推荐菜ID"
// @Success 200 {object} response.Envelope{data=response.MealSuggestion,msg=string} "获取成功"
// @Router /miniapp/v1/meal-suggestions/{suggestionId} [get]
func (*AssistApi) Suggestion(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := assistService.Suggestion(c.Request.Context(), user.ID, c.Param("suggestionId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CopySuggestion 将推荐结果中的菜品复制到个人菜品库
// @Tags 小程序增强功能
// @Summary 将推荐结果中的菜品复制到个人菜品库
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param suggestionId path string true "推荐菜ID"
// @Param data body frontRequest.SuggestionDishCopyInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.SuggestionCopyResult,msg=string} "操作成功"
// @Router /miniapp/v1/meal-suggestions/{suggestionId}/copy [post]
func (*AssistApi) CopySuggestion(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.SuggestionDishCopyInput
	if !bindJSON(c, &input) {
		return
	}
	dish, recorded, err := assistService.CopySuggestion(
		c.Request.Context(), user.ID, c.Param("suggestionId"), input.SuggestionDishID,
	)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.SuggestionCopyResult{
		Dish: dish, AdoptionRecorded: recorded,
	}, c)
}

// SuggestionFeedback 记录用户对推荐菜结果的反馈
// @Tags 小程序增强功能
// @Summary 记录用户对推荐菜结果的反馈
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param suggestionId path string true "推荐菜ID"
// @Param data body frontRequest.FeedbackInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.RecordedResult,msg=string} "操作成功"
// @Router /miniapp/v1/meal-suggestions/{suggestionId}/feedback [post]
func (*AssistApi) SuggestionFeedback(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.FeedbackInput
	if !bindJSON(c, &input) {
		return
	}
	if err := assistService.Feedback(
		c.Request.Context(), user.ID, "suggestion", c.Param("suggestionId"), input.Action,
	); err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.RecordedResult{Recorded: true}, c)
}

// PrepQuote 获取备菜顺序生成所需积分报价
// @Tags 小程序增强功能
// @Summary 获取备菜顺序生成所需积分报价
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.PrepPlanQuoteQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.PrepQuoteResult,msg=string} "获取成功"
// @Router /miniapp/v1/prep-plans/quote [get]
func (*AssistApi) PrepQuote(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var query frontRequest.PrepPlanQuoteQuery
	if !bindQuery(c, &query) {
		return
	}
	cost, balance, can, err := assistService.PrepQuote(c.Request.Context(), user.ID, query.MealID)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.PrepQuoteResult{
		PointCost: cost, PointBalance: balance, CanGenerate: can,
	}, c)
}

// GeneratePrepPlan 根据已确认饭局生成备菜顺序
// @Tags 小程序增强功能
// @Summary 根据已确认饭局生成备菜顺序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.PrepPlanInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.PrepPlan,msg=string} "操作成功"
// @Router /miniapp/v1/prep-plans [post]
func (*AssistApi) GeneratePrepPlan(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.PrepPlanInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := assistService.GeneratePrepPlan(c.Request.Context(), user.ID, input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// PrepPlan 获取备菜顺序详情
// @Tags 小程序增强功能
// @Summary 获取备菜顺序详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param planId path string true "计划ID"
// @Success 200 {object} response.Envelope{data=response.PrepPlan,msg=string} "获取成功"
// @Router /miniapp/v1/prep-plans/{planId} [get]
func (*AssistApi) PrepPlan(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := assistService.PrepPlan(c.Request.Context(), user.ID, c.Param("planId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// PrepFeedback 记录用户对备菜顺序的反馈
// @Tags 小程序增强功能
// @Summary 记录用户对备菜顺序的反馈
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param planId path string true "计划ID"
// @Param data body frontRequest.FeedbackInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.RecordedResult,msg=string} "操作成功"
// @Router /miniapp/v1/prep-plans/{planId}/feedback [post]
func (*AssistApi) PrepFeedback(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.FeedbackInput
	if !bindJSON(c, &input) {
		return
	}
	if err := assistService.Feedback(
		c.Request.Context(), user.ID, "prep", c.Param("planId"), input.Action,
	); err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.RecordedResult{Recorded: true}, c)
}

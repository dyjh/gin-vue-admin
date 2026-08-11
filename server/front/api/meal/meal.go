package meal

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/front/api/common"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	response "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/gin-gonic/gin"
)

// MealApi 提供饭局与采购清单接口处理能力。
type MealApi struct{}

// Meals 分页查询本人创建的饭局
// @Tags 饭局与采购清单
// @Summary 分页查询本人创建的饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.MealListQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.MealSummary],msg=string} "获取成功"
// @Router /miniapp/v1/meals [get]
func (*MealApi) Meals(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.MealListQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	result, err := mealService.List(c.Request.Context(), user.ID, query)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CurrentMeal 获取当前用户创建的进行中饭局
// @Tags 饭局与采购清单
// @Summary 获取当前用户创建的进行中饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.CurrentMealResult,msg=string} "获取成功"
// @Router /miniapp/v1/meals/current [get]
func (*MealApi) CurrentMeal(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	meal, err := mealService.Current(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.CurrentMealResult{Meal: meal}, c)
}

// CreateMeal 创建饭局
// @Tags 饭局与采购清单
// @Summary 创建饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.MealCreateInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Meal,msg=string} "操作成功"
// @Router /miniapp/v1/meals [post]
func (*MealApi) CreateMeal(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealCreateInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.Create(c.Request.Context(), user.ID, input)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Meal 获取饭局详情
// @Tags 饭局与采购清单
// @Summary 获取饭局详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param mealId path string true "饭局ID"
// @Success 200 {object} response.Envelope{data=response.Meal,msg=string} "获取成功"
// @Router /miniapp/v1/meals/{mealId} [get]
func (*MealApi) Meal(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	result, err := mealService.Meal(c.Request.Context(), user.ID, c.Param("mealId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Lookup 使用点餐码查询可加入的饭局
// @Tags 饭局与采购清单
// @Summary 使用点餐码查询可加入的饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.MealCodeInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.MealJoinPreview,msg=string} "操作成功"
// @Router /miniapp/v1/meals/lookup [post]
func (*MealApi) Lookup(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	if _, ok := apiCommon.CurrentUser(c); !ok {
		return
	}
	var input frontRequest.MealCodeInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.Lookup(c.Request.Context(), input.Code)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Join 加入点餐码对应的饭局
// @Tags 饭局与采购清单
// @Summary 加入点餐码对应的饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.MealCodeInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Meal,msg=string} "操作成功"
// @Router /miniapp/v1/meals/join [post]
func (*MealApi) Join(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealCodeInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.Join(c.Request.Context(), user.ID, input.Code)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// RecordFinalResultSubscription 登记饭局最终结果订阅授权
// @Tags 饭局与采购清单
// @Summary 登记饭局最终结果订阅授权
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Param data body frontRequest.MealFinalResultSubscriptionInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.MealFinalResultSubscription,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/final-result-subscriptions [post]
func (*MealApi) RecordFinalResultSubscription(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealFinalResultSubscriptionInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.RecordFinalResultSubscription(
		c.Request.Context(),
		user.ID,
		c.Param("mealId"),
		input.TemplateID,
		input.AuthorizationResult,
	)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Close 关闭饭局点单
// @Tags 饭局与采购清单
// @Summary 关闭饭局点单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Success 200 {object} response.Envelope{data=response.Meal,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/close [post]
func (*MealApi) Close(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	result, err := mealService.Close(c.Request.Context(), user.ID, c.Param("mealId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Cancel 取消饭局
// @Tags 饭局与采购清单
// @Summary 取消饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Param data body frontRequest.MealCancelInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Meal,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/cancel [post]
func (*MealApi) Cancel(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealCancelInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.Cancel(c.Request.Context(), user.ID, c.Param("mealId"), input.Reason)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Candidates 获取饭局候选菜列表
// @Tags 饭局与采购清单
// @Summary 获取饭局候选菜列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param mealId path string true "饭局ID"
// @Param data query frontRequest.MealCandidatesQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.MealCandidatesResult,msg=string} "获取成功"
// @Router /miniapp/v1/meals/{mealId}/candidates [get]
func (*MealApi) Candidates(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.MealCandidatesQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	categories, list, err := mealService.CandidateList(
		c.Request.Context(), user.ID, c.Param("mealId"), query.Category,
	)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.MealCandidatesResult{
		Categories: categories, List: list,
	}, c)
}

// RemoveCandidate 移除饭局候选菜
// @Tags 饭局与采购清单
// @Summary 移除饭局候选菜
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Param candidateId path string true "候选菜ID"
// @Success 200 {object} response.Envelope{data=response.MealCandidateRemovalResult,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/candidates/{candidateId} [delete]
func (*MealApi) RemoveCandidate(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var path frontRequest.MealCandidatePath
	if !apiCommon.BindURI(c, &path) {
		return
	}
	removedAt, err := mealService.RemoveCandidate(
		c.Request.Context(),
		user.ID,
		path.MealID,
		path.CandidateID,
	)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.MealCandidateRemovalResult{
		CandidateID: path.CandidateID,
		RemovedAt:   removedAt,
	}, c)
}

// MyVotes 获取当前用户在饭局中的点选结果
// @Tags 饭局与采购清单
// @Summary 获取当前用户在饭局中的点选结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param mealId path string true "饭局ID"
// @Success 200 {object} response.Envelope{data=response.MealVotesResult,msg=string} "获取成功"
// @Router /miniapp/v1/meals/{mealId}/votes/me [get]
func (*MealApi) MyVotes(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	ids, at, err := mealService.MyVotes(c.Request.Context(), user.ID, c.Param("mealId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.MealVotesResult{
		CandidateIDs: ids, UpdatedAt: at,
	}, c)
}

// SaveVotes 覆盖保存当前用户的饭局点选结果
// @Tags 饭局与采购清单
// @Summary 覆盖保存当前用户的饭局点选结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Param data body frontRequest.MealVoteInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.MealVotesSavedResult,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/votes/me [put]
func (*MealApi) SaveVotes(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealVoteInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	ids, at, err := mealService.SaveVotes(c.Request.Context(), user.ID, c.Param("mealId"), input.CandidateIDs)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.MealVotesSavedResult{
		CandidateIDs: ids, SavedAt: at,
	}, c)
}

// Stats 获取饭局候选菜点选统计
// @Tags 饭局与采购清单
// @Summary 获取饭局候选菜点选统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param mealId path string true "饭局ID"
// @Success 200 {object} response.Envelope{data=response.MealStatsResult,msg=string} "获取成功"
// @Router /miniapp/v1/meals/{mealId}/stats [get]
func (*MealApi) Stats(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	meal, dishes, err := mealService.Stats(c.Request.Context(), user.ID, c.Param("mealId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.MealStatsResult{Meal: meal, Dishes: dishes}, c)
}

// Confirm 确认最终菜单并生成采购清单
// @Tags 饭局与采购清单
// @Summary 确认最终菜单并生成采购清单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Param data body frontRequest.MealConfirmInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.MealConfirmResult,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/confirm [post]
func (*MealApi) Confirm(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.MealConfirmInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	meal, listID, err := mealService.Confirm(c.Request.Context(), user.ID, c.Param("mealId"), input)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.MealConfirmResult{
		Meal: meal, ShoppingListID: listID,
	}, c)
}

// Complete 手动完成饭局
// @Tags 饭局与采购清单
// @Summary 手动完成饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param mealId path string true "饭局ID"
// @Success 200 {object} response.Envelope{data=response.Meal,msg=string} "操作成功"
// @Router /miniapp/v1/meals/{mealId}/complete [post]
func (*MealApi) Complete(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	result, err := mealService.Complete(c.Request.Context(), user.ID, c.Param("mealId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CurrentShopping 获取当前饭局的采购清单
// @Tags 饭局与采购清单
// @Summary 获取当前饭局的采购清单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.ShoppingList,msg=string} "获取成功"
// @Router /miniapp/v1/shopping-lists/current [get]
func (*MealApi) CurrentShopping(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	result, err := mealService.CurrentShopping(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// ShoppingByID 获取本人或参与者保留的采购清单
// @Tags 饭局与采购清单
// @Summary 获取指定采购清单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param listId path string true "采购清单ID"
// @Success 200 {object} response.Envelope{data=response.ShoppingList,msg=string} "获取成功"
// @Router /miniapp/v1/shopping-lists/{listId} [get]
func (*MealApi) ShoppingByID(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	result, err := mealService.ShoppingByID(
		c.Request.Context(),
		user.ID,
		c.Param("listId"),
	)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// SharedShopping 通过分享令牌获取只读采购清单
// @Tags 饭局与采购清单
// @Summary 通过分享令牌获取只读采购清单
// @Security NoAuth
// @accept application/json
// @Produce application/json
// @Param shareToken path string true "分享Token"
// @Success 200 {object} response.Envelope{data=response.SharedShoppingList,msg=string} "获取成功"
// @Router /miniapp/v1/public/shopping-lists/{shareToken} [get]
func (*MealApi) SharedShopping(c *gin.Context) {
	result, err := mealService.SharedShopping(c.Request.Context(), c.Param("shareToken"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CreateShoppingItem 创建采购清单项
// @Tags 饭局与采购清单
// @Summary 创建采购清单项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.ShoppingItemCreateInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.ShoppingItem,msg=string} "操作成功"
// @Router /miniapp/v1/shopping-lists/current/items [post]
func (*MealApi) CreateShoppingItem(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.ShoppingItemCreateInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.CreateShoppingItem(c.Request.Context(), user.ID, input)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// UpdateShoppingItem 编辑采购清单项
// @Tags 饭局与采购清单
// @Summary 编辑采购清单项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param itemId path string true "项目ID"
// @Param data body frontRequest.ShoppingItemUpdateInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.ShoppingItem,msg=string} "操作成功"
// @Router /miniapp/v1/shopping-lists/current/items/{itemId} [put]
func (*MealApi) UpdateShoppingItem(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.ShoppingItemUpdateInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	result, err := mealService.UpdateShoppingItem(c.Request.Context(), user.ID, c.Param("itemId"), input)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// DeleteShoppingItem 删除采购清单项
// @Tags 饭局与采购清单
// @Summary 删除采购清单项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param itemId path string true "项目ID"
// @Success 200 {object} response.Envelope{data=response.DeletedResult,msg=string} "操作成功"
// @Router /miniapp/v1/shopping-lists/current/items/{itemId} [delete]
func (*MealApi) DeleteShoppingItem(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	if err := mealService.DeleteShoppingItem(c.Request.Context(), user.ID, c.Param("itemId")); err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.DeletedResult{Deleted: true}, c)
}

// ExportShopping 生成采购清单的可复制文本
// @Tags 饭局与采购清单
// @Summary 生成采购清单的可复制文本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param listId path string true "列表ID"
// @Success 200 {object} response.Envelope{data=response.ShoppingExportResult,msg=string} "获取成功"
// @Router /miniapp/v1/shopping-lists/{listId}/export-text [get]
func (*MealApi) ExportShopping(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	title, text, at, err := mealService.ExportShopping(c.Request.Context(), user.ID, c.Param("listId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.ShoppingExportResult{
		Title: title, Text: text, UpdatedAt: at,
	}, c)
}

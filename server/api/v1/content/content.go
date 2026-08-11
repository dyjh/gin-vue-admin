package content

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	"github.com/gin-gonic/gin"
)

// ContentApi 提供内容管理接口处理能力。
type ContentApi struct{}

// ListUserDishes 分页查询全部用户菜品
// @Tags OrderFoodUserDish
// @Summary 分页查询全部用户菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query contentRequest.UserDishSearch true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]contentResponse.UserDishAdminSummary},msg=string} "获取成功"
// @Router /orderfood/user-dishes [get]
func (api *ContentApi) ListUserDishes(c *gin.Context) {
	var query contentRequest.UserDishSearch
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	var result response.Page[contentResponse.UserDishAdminSummary]
	result, err := contentService.ListUserDishes(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetUserDishDetail 获取用户菜品私有详情
// @Tags OrderFoodUserDish
// @Summary 获取用户菜品私有详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param dishId path string true "菜品 ID"
// @Success 200 {object} response.Response{data=contentResponse.UserDishAdminDetail,msg=string} "获取成功"
// @Router /orderfood/user-dishes/{dishId} [get]
func (api *ContentApi) GetUserDishDetail(c *gin.Context) {
	path, ok := apiCommon.BindDishIDPath(c)
	if !ok {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := contentService.GetUserDishDetail(c.Request.Context(), path.DishID, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListUserDishReferences 分页查询用户菜品引用位置
// @Tags OrderFoodUserDish
// @Summary 分页查询用户菜品引用位置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param dishId path string true "菜品 ID"
// @Param data query contentRequest.UserDishReferenceSearch true "查询条件"
// @Success 200 {object} response.Response{data=response.AuditedPage[contentResponse.UserDishReference],msg=string} "获取成功"
// @Router /orderfood/user-dishes/{dishId}/references [get]
func (api *ContentApi) ListUserDishReferences(c *gin.Context) {
	path, ok := apiCommon.BindDishIDPath(c)
	if !ok {
		return
	}
	var query contentRequest.UserDishReferenceSearch
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := contentService.ListUserDishReferences(c.Request.Context(), path.DishID, query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListUserRecipes 分页查询全部用户菜谱
// @Tags OrderFoodUserRecipe
// @Summary 分页查询全部用户菜谱
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query contentRequest.UserRecipeSearch true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]contentResponse.UserRecipeAdminSummary},msg=string} "获取成功"
// @Router /orderfood/user-recipes [get]
func (api *ContentApi) ListUserRecipes(c *gin.Context) {
	var query contentRequest.UserRecipeSearch
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := contentService.ListUserRecipes(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetUserRecipeDetail 获取用户菜谱私有详情
// @Tags OrderFoodUserRecipe
// @Summary 获取用户菜谱私有详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recipeId path string true "菜谱 ID"
// @Success 200 {object} response.Response{data=contentResponse.UserRecipeAdminDetail,msg=string} "获取成功"
// @Router /orderfood/user-recipes/{recipeId} [get]
func (api *ContentApi) GetUserRecipeDetail(c *gin.Context) {
	path, ok := bindRecipeIDPath(c)
	if !ok {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := contentService.GetUserRecipeDetail(c.Request.Context(), path.RecipeID, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// PreviewGovernanceAction 预览内容源头违规处理影响
// @Tags OrderFoodGovernance
// @Summary 预览用户菜品或菜谱违规处理影响
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body contentRequest.GovernanceActionPreviewInput true "违规处理预览"
// @Success 200 {object} response.Response{data=contentResponse.GovernanceImpactPreview,msg=string} "操作成功"
// @Router /orderfood/governance-actions/preview [post]
func (api *ContentApi) PreviewGovernanceAction(c *gin.Context) {
	var body contentRequest.GovernanceActionPreviewInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.TargetID = strings.TrimSpace(body.TargetID)
	body.Reason = strings.TrimSpace(body.Reason)
	body.ViolationType = strings.TrimSpace(body.ViolationType)
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := contentService.PreviewGovernance(c.Request.Context(), body, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ExecuteGovernanceAction 执行内容源头违规处理
// @Tags OrderFoodGovernance
// @Summary 执行用户菜品或菜谱违规处理
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body contentRequest.GovernanceActionExecuteInput true "违规处理执行参数"
// @Success 200 {object} response.Response{data=contentResponse.GovernanceExecutionResult,msg=string} "操作成功"
// @Router /orderfood/governance-actions [post]
func (api *ContentApi) ExecuteGovernanceAction(c *gin.Context) {
	var header commonRequest.IdempotencyHeader
	if !apiCommon.BindAndVerify(c, &header, c.ShouldBindHeader) {
		return
	}
	header.Key = strings.TrimSpace(header.Key)
	var body contentRequest.GovernanceActionExecuteInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.TargetID = strings.TrimSpace(body.TargetID)
	body.Reason = strings.TrimSpace(body.Reason)
	body.ViolationType = strings.TrimSpace(body.ViolationType)
	body.PreviewToken = strings.TrimSpace(body.PreviewToken)
	body.ConfirmText = strings.TrimSpace(body.ConfirmText)
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || contentService == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := contentService.ExecuteGovernance(c.Request.Context(), body, actor, header.Key)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

func bindRecipeIDPath(c *gin.Context) (contentRequest.UserRecipeIDPath, bool) {
	var path contentRequest.UserRecipeIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return contentRequest.UserRecipeIDPath{}, false
	}
	path.RecipeID = strings.TrimSpace(path.RecipeID)
	if path.RecipeID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return contentRequest.UserRecipeIDPath{}, false
	}
	return path, true
}

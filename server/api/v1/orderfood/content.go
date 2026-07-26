package orderfood

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// ContentApi 提供内容管理接口处理能力。
type ContentApi struct {
	service *orderfoodService.ContentService
}

// NewContentApi 创建内容API实例。
func NewContentApi(service *orderfoodService.ContentService) *ContentApi {
	return &ContentApi{service: service}
}

// ListUserDishes 分页查询全部用户菜品
// @Tags OrderFoodUserDish
// @Summary 分页查询全部用户菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.UserDishSearch true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.UserDishAdminSummary},msg=string} "获取成功"
// @Router /orderfood/user-dishes [get]
func (api *ContentApi) ListUserDishes(c *gin.Context) {
	var query orderfoodRequest.UserDishSearch
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.UserDishAdminSummary]
	result, err := api.service.ListUserDishes(c.Request.Context(), query, actor)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.UserDishAdminDetail,msg=string} "获取成功"
// @Router /orderfood/user-dishes/{dishId} [get]
func (api *ContentApi) GetUserDishDetail(c *gin.Context) {
	path, ok := bindDishIDPath(c)
	if !ok {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := api.service.GetUserDishDetail(c.Request.Context(), path.DishID, actor)
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
// @Param data query orderfoodRequest.UserDishReferenceSearch true "查询条件"
// @Success 200 {object} response.Response{data=orderfoodResponse.AuditedPage[orderfoodResponse.UserDishReference],msg=string} "获取成功"
// @Router /orderfood/user-dishes/{dishId}/references [get]
func (api *ContentApi) ListUserDishReferences(c *gin.Context) {
	path, ok := bindDishIDPath(c)
	if !ok {
		return
	}
	var query orderfoodRequest.UserDishReferenceSearch
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := api.service.ListUserDishReferences(c.Request.Context(), path.DishID, query, actor)
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
// @Param data query orderfoodRequest.UserRecipeSearch true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.UserRecipeAdminSummary},msg=string} "获取成功"
// @Router /orderfood/user-recipes [get]
func (api *ContentApi) ListUserRecipes(c *gin.Context) {
	var query orderfoodRequest.UserRecipeSearch
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := api.service.ListUserRecipes(c.Request.Context(), query, actor)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.UserRecipeAdminDetail,msg=string} "获取成功"
// @Router /orderfood/user-recipes/{recipeId} [get]
func (api *ContentApi) GetUserRecipeDetail(c *gin.Context) {
	path, ok := bindRecipeIDPath(c)
	if !ok {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := api.service.GetUserRecipeDetail(c.Request.Context(), path.RecipeID, actor)
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
// @Param data body orderfoodRequest.GovernanceActionPreviewInput true "违规处理预览"
// @Success 200 {object} response.Response{data=orderfoodResponse.GovernanceImpactPreview,msg=string} "操作成功"
// @Router /orderfood/governance-actions/preview [post]
func (api *ContentApi) PreviewGovernanceAction(c *gin.Context) {
	var body orderfoodRequest.GovernanceActionPreviewInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.TargetID = strings.TrimSpace(body.TargetID)
	body.Reason = strings.TrimSpace(body.Reason)
	body.ViolationType = strings.TrimSpace(body.ViolationType)
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := api.service.PreviewGovernance(c.Request.Context(), body, actor)
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
// @Param data body orderfoodRequest.GovernanceActionExecuteInput true "违规处理执行参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.GovernanceExecutionResult,msg=string} "操作成功"
// @Router /orderfood/governance-actions [post]
func (api *ContentApi) ExecuteGovernanceAction(c *gin.Context) {
	var header orderfoodRequest.IdempotencyHeader
	if !bindAndVerify(c, &header, c.ShouldBindHeader) {
		return
	}
	header.Key = strings.TrimSpace(header.Key)
	var body orderfoodRequest.GovernanceActionExecuteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	body.TargetID = strings.TrimSpace(body.TargetID)
	body.Reason = strings.TrimSpace(body.Reason)
	body.ViolationType = strings.TrimSpace(body.ViolationType)
	body.PreviewToken = strings.TrimSpace(body.PreviewToken)
	body.ConfirmText = strings.TrimSpace(body.ConfirmText)
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	if api == nil || api.service == nil {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return
	}
	result, err := api.service.ExecuteGovernance(c.Request.Context(), body, actor, header.Key)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

func bindDishIDPath(c *gin.Context) (orderfoodRequest.UserDishIDPath, bool) {
	var path orderfoodRequest.UserDishIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.UserDishIDPath{}, false
	}
	path.DishID = strings.TrimSpace(path.DishID)
	if path.DishID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.UserDishIDPath{}, false
	}
	return path, true
}

func bindRecipeIDPath(c *gin.Context) (orderfoodRequest.UserRecipeIDPath, bool) {
	var path orderfoodRequest.UserRecipeIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.UserRecipeIDPath{}, false
	}
	path.RecipeID = strings.TrimSpace(path.RecipeID)
	if path.RecipeID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.UserRecipeIDPath{}, false
	}
	return path, true
}

func contentAdminActor(c *gin.Context) (orderfoodRequest.AdminActor, bool) {
	claims := utils.GetUserInfo(c)
	if claims == nil || claims.BaseClaims.ID == 0 {
		response.FailWithBusinessError(appErrors.AdminLoginExpired.DefaultMsg(), c)
		return orderfoodRequest.AdminActor{}, false
	}
	requestID := middleware.GetRequestID(c)
	if requestID == "" {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return orderfoodRequest.AdminActor{}, false
	}
	var nickname *string
	if value := strings.TrimSpace(claims.NickName); value != "" {
		nickname = &value
	}
	return orderfoodRequest.AdminActor{
		AdministratorID:  claims.BaseClaims.ID,
		AuthorityID:      claims.AuthorityId,
		Username:         claims.Username,
		Nickname:         nickname,
		RequestID:        requestID,
		SourceIPMasked:   orderfoodService.MaskIP(c.ClientIP()),
		UserAgentSummary: c.Request.UserAgent(),
	}, true
}

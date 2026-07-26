package orderfood

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/gin-gonic/gin"
)

// MealAdminApi 提供饭局和采购清单管理端只读接口处理能力。
type MealAdminApi struct {
	service *orderfoodService.MealAdminService // 饭局管理业务服务
}

// NewMealAdminApi 创建饭局管理API实例。
func NewMealAdminApi(service *orderfoodService.MealAdminService) *MealAdminApi {
	return &MealAdminApi{service: service}
}

// ListMeals 分页查询饭局
// @Tags OrderFoodMeal
// @Summary 分页查询饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.MealAdminListQuery false "分页和筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.MealAdminSummary},msg=string} "获取成功"
// @Router /orderfood/meals [get]
func (api *MealAdminApi) ListMeals(c *gin.Context) {
	var query orderfoodRequest.MealAdminListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.MealAdminSummary]
	result, err := api.service.ListMeals(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetMeal 获取饭局详情
// @Tags OrderFoodMeal
// @Summary 获取饭局详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param mealId path string true "饭局ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.MealAdminDetail,msg=string} "获取成功"
// @Router /orderfood/meals/{mealId} [get]
func (api *MealAdminApi) GetMeal(c *gin.Context) {
	path, ok := bindMealAdminIDPath(c)
	if !ok {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.MealAdminDetail
	result, err := api.service.GetMeal(c.Request.Context(), path.MealID, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListShoppingLists 分页查询采购清单
// @Tags OrderFoodShopping
// @Summary 分页查询采购清单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.ShoppingListAdminListQuery false "分页和筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.ShoppingListAdminSummary},msg=string} "获取成功"
// @Router /orderfood/shopping-lists [get]
func (api *MealAdminApi) ListShoppingLists(c *gin.Context) {
	var query orderfoodRequest.ShoppingListAdminListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.ShoppingListAdminSummary]
	result, err := api.service.ListShoppingLists(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetShoppingList 获取采购清单详情
// @Tags OrderFoodShopping
// @Summary 获取采购清单详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param shoppingListId path string true "采购清单ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.ShoppingListAdminDetail,msg=string} "获取成功"
// @Router /orderfood/shopping-lists/{shoppingListId} [get]
func (api *MealAdminApi) GetShoppingList(c *gin.Context) {
	path, ok := bindShoppingListAdminIDPath(c)
	if !ok {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.ShoppingListAdminDetail
	result, err := api.service.GetShoppingList(
		c.Request.Context(),
		path.ShoppingListID,
		actor,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// available 检查饭局管理API是否已注入业务服务。
func (api *MealAdminApi) available(c *gin.Context) bool {
	if api != nil && api.service != nil {
		return true
	}
	response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
	return false
}

// bindMealAdminIDPath 绑定并规范化饭局ID路径参数。
func bindMealAdminIDPath(c *gin.Context) (orderfoodRequest.MealAdminIDPath, bool) {
	var path orderfoodRequest.MealAdminIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.MealAdminIDPath{}, false
	}
	path.MealID = strings.TrimSpace(path.MealID)
	if path.MealID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.MealAdminIDPath{}, false
	}
	return path, true
}

// bindShoppingListAdminIDPath 绑定并规范化采购清单ID路径参数。
func bindShoppingListAdminIDPath(
	c *gin.Context,
) (orderfoodRequest.ShoppingListAdminIDPath, bool) {
	var path orderfoodRequest.ShoppingListAdminIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.ShoppingListAdminIDPath{}, false
	}
	path.ShoppingListID = strings.TrimSpace(path.ShoppingListID)
	if path.ShoppingListID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.ShoppingListAdminIDPath{}, false
	}
	return path, true
}

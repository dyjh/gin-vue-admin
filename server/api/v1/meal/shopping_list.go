package meal

import (
	"strings"

	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	mealRequest "github.com/dyjh/order-food-mini-app/server/model/meal/request"
	mealResponse "github.com/dyjh/order-food-mini-app/server/model/meal/response"
	"github.com/gin-gonic/gin"
)

// ShoppingListApi 提供采购清单管理端只读接口处理能力。
type ShoppingListApi struct{}

// ListShoppingLists 分页查询采购清单
// @Tags OrderFoodShopping
// @Summary 分页查询采购清单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query mealRequest.ShoppingListAdminListQuery false "分页和筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]mealResponse.ShoppingListAdminSummary},msg=string} "获取成功"
// @Router /orderfood/shopping-lists [get]
func (api *ShoppingListApi) ListShoppingLists(c *gin.Context) {
	var query mealRequest.ShoppingListAdminListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result response.Page[mealResponse.ShoppingListAdminSummary]
	result, err := shoppingListService.ListShoppingLists(c.Request.Context(), query, actor)
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
// @Success 200 {object} response.Response{data=mealResponse.ShoppingListAdminDetail,msg=string} "获取成功"
// @Router /orderfood/shopping-lists/{shoppingListId} [get]
func (api *ShoppingListApi) GetShoppingList(c *gin.Context) {
	path, ok := bindShoppingListIDPath(c)
	if !ok {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result mealResponse.ShoppingListAdminDetail
	result, err := shoppingListService.GetShoppingList(
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

// bindShoppingListIDPath 绑定并规范化采购清单ID路径参数。
func bindShoppingListIDPath(c *gin.Context) (mealRequest.ShoppingListAdminIDPath, bool) {
	var path mealRequest.ShoppingListAdminIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return mealRequest.ShoppingListAdminIDPath{}, false
	}
	path.ShoppingListID = strings.TrimSpace(path.ShoppingListID)
	if path.ShoppingListID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return mealRequest.ShoppingListAdminIDPath{}, false
	}
	return path, true
}

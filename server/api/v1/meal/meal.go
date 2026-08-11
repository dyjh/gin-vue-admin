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

// MealApi 提供饭局管理端只读接口处理能力。
type MealApi struct{}

// ListMeals 分页查询饭局
// @Tags OrderFoodMeal
// @Summary 分页查询饭局
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query mealRequest.MealAdminListQuery false "分页和筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]mealResponse.MealAdminSummary},msg=string} "获取成功"
// @Router /orderfood/meals [get]
func (api *MealApi) ListMeals(c *gin.Context) {
	var query mealRequest.MealAdminListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result response.Page[mealResponse.MealAdminSummary]
	result, err := mealService.ListMeals(c.Request.Context(), query, actor)
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
// @Success 200 {object} response.Response{data=mealResponse.MealAdminDetail,msg=string} "获取成功"
// @Router /orderfood/meals/{mealId} [get]
func (api *MealApi) GetMeal(c *gin.Context) {
	path, ok := bindMealIDPath(c)
	if !ok {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result mealResponse.MealAdminDetail
	result, err := mealService.GetMeal(c.Request.Context(), path.MealID, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// bindMealIDPath 绑定并规范化饭局ID路径参数。
func bindMealIDPath(c *gin.Context) (mealRequest.MealAdminIDPath, bool) {
	var path mealRequest.MealAdminIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return mealRequest.MealAdminIDPath{}, false
	}
	path.MealID = strings.TrimSpace(path.MealID)
	if path.MealID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return mealRequest.MealAdminIDPath{}, false
	}
	return path, true
}

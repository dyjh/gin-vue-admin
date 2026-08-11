package system

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/front/api/common"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/gin-gonic/gin"
)

// SystemApi 提供小程序基础配置接口处理能力。
type SystemApi struct{}

// RuntimeConfig 获取当前生效的小程序运行时配置
// @Tags 小程序基础配置
// @Summary 获取当前生效的小程序运行时配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} frontResponse.Envelope{data=frontResponse.RuntimeConfig,msg=string} "获取成功"
// @Router /miniapp/v1/runtime-config [get]
func (*SystemApi) RuntimeConfig(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	runtime, err := runtimeService.Current(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	frontResponse.OkWithData(runtime, c)
}

// Bootstrap 获取小程序首页启动所需的聚合数据
// @Tags 小程序基础配置
// @Summary 获取小程序首页启动所需的聚合数据
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} frontResponse.Envelope{data=frontResponse.BootstrapData,msg=string} "获取成功"
// @Router /miniapp/v1/bootstrap [get]
func (*SystemApi) Bootstrap(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	runtime, err := runtimeService.Current(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	dishes, err := contentService.ListDishes(c.Request.Context(), user.ID, frontRequest.DishListQuery{
		PageQuery: frontRequest.PageQuery{Page: 1, PageSize: 3},
	})
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	recommendations, err := contentService.Recommendations(
		c.Request.Context(), user.ID,
		frontRequest.RecommendationListQuery{PageQuery: frontRequest.PageQuery{Page: 1, PageSize: 5}},
	)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	activeMeal, err := mealService.Current(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	unread, _, err := engagementService.NotificationSummary(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	var activeSummary *frontResponse.MealSummary
	if activeMeal != nil {
		summary := activeMeal.MealSummary
		activeSummary = &summary
	}
	frontResponse.OkWithData(frontResponse.BootstrapData{
		Profile: apiCommon.ServiceProfile(user, runtime), RuntimeConfig: runtime,
		Dishes: dishes.List, DishTotal: dishes.Total, Recommendations: recommendations.List,
		ActiveMeal: activeSummary, UnreadCount: unread,
	}, c)
}

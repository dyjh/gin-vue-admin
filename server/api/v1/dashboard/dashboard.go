package dashboard

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	dashboardRequest "github.com/dyjh/order-food-mini-app/server/model/dashboard/request"
	dashboardResponse "github.com/dyjh/order-food-mini-app/server/model/dashboard/response"
	"github.com/gin-gonic/gin"
)

// DashboardApi 提供管理端运营概览接口处理能力。
type DashboardApi struct{}

// GetDashboard 获取运营概览
// @Tags OrderFoodDashboard
// @Summary 获取运营概览
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query dashboardRequest.DashboardQuery false "统计范围"
// @Success 200 {object} response.Response{data=dashboardResponse.DashboardData,msg=string} "获取成功"
// @Router /orderfood/dashboard [get]
func (api *DashboardApi) GetDashboard(c *gin.Context) {
	var query dashboardRequest.DashboardQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result dashboardResponse.DashboardData
	result, err := dashboardService.GetDashboard(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

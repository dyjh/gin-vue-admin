package dashboard

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	DashboardRouter
}

var dashboardApi = api.ApiGroupApp.DashboardApiGroup.DashboardApi

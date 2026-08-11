package dashboard

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合运营概览接口。
type ApiGroup struct {
	DashboardApi // 运营概览接口
}

var dashboardService = service.ServiceGroupApp.DashboardServiceGroup.Dashboard

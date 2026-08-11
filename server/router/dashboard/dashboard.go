package dashboard

import "github.com/gin-gonic/gin"

// DashboardRouter registers the administrator dashboard route.
type DashboardRouter struct{}

// InitDashboardRouter initializes the dashboard route.
func (*DashboardRouter) InitDashboardRouter(orderFoodGroup *gin.RouterGroup) {
	orderFoodGroup.GET("/dashboard", dashboardApi.GetDashboard)
}

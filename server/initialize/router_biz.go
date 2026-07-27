package initialize

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/dyjh/order-food-mini-app/server/front"
	"github.com/dyjh/order-food-mini-app/server/router"
	"github.com/gin-gonic/gin"
)

// 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	holder(publicGroup, privateGroup)
	front.Register(publicGroup)
	businessRouter := router.RouterGroupApp
	businessRouter.InitUserRouter(privateGroup)
	businessRouter.InitAuditRouter(privateGroup)
	businessRouter.InitContentRouter(privateGroup)
	businessRouter.InitPointsRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.PointsApi,
	)
	businessRouter.InitCatalogRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.CatalogApi,
	)
	businessRouter.InitRecommendationRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.RecommendationApi,
	)
	businessRouter.InitOfficialDishRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.OfficialDishApi,
	)
	businessRouter.InitMediaRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.MediaApi,
	)
	businessRouter.InitGovernanceRouter(privateGroup)
	businessRouter.InitMealAdminRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.MealAdminApi,
	)
	businessRouter.InitOperationsRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.DashboardApi,
		orderfoodApi.ApiGroupApp.AIUsageApi,
	)
	businessRouter.InitModerationRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.ModerationApi,
	)
	businessRouter.InitAIRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.AIApi,
	)
	businessRouter.InitSuggestionCatalogRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.SuggestionCatalogApi,
	)
	businessRouter.InitSubscriptionRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.SubscriptionApi,
	)
	businessRouter.InitWeChatConfigRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.WeChatConfigApi,
	)

}

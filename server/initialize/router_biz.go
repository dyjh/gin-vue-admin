package initialize

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
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
	orderFoodRouter := router.RouterGroupApp.OrderFood
	orderFoodRouter.InitUserRouter(privateGroup)
	orderFoodRouter.InitAuditRouter(privateGroup)
	orderFoodRouter.InitContentRouter(privateGroup)
	orderFoodRouter.InitPointsRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.PointsApi,
	)
	orderFoodRouter.InitCatalogRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.CatalogApi,
	)
	orderFoodRouter.InitRecommendationRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.RecommendationApi,
	)
	orderFoodRouter.InitOfficialDishRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.OfficialDishApi,
	)
	orderFoodRouter.InitMediaRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.MediaApi,
	)
	orderFoodRouter.InitGovernanceRouter(privateGroup)
	orderFoodRouter.InitMealAdminRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.MealAdminApi,
	)
	orderFoodRouter.InitOperationsRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.DashboardApi,
		orderfoodApi.ApiGroupApp.AIUsageApi,
	)
	orderFoodRouter.InitModerationRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.ModerationApi,
	)
	orderFoodRouter.InitAIRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.AIApi,
	)
	orderFoodRouter.InitSuggestionCatalogRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.SuggestionCatalogApi,
	)
	orderFoodRouter.InitSubscriptionRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.SubscriptionApi,
	)
	orderFoodRouter.InitWeChatConfigRouter(
		privateGroup,
		orderfoodApi.ApiGroupApp.WeChatConfigApi,
	)

}

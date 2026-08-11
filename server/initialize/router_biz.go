package initialize

import (
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
	orderFoodGroup := privateGroup.Group("/orderfood")
	userRouter := router.RouterGroupApp.User
	engagementRouter := router.RouterGroupApp.Engagement
	contentRouter := router.RouterGroupApp.Content
	dishRouter := router.RouterGroupApp.Dish
	mealRouter := router.RouterGroupApp.Meal
	aiRouter := router.RouterGroupApp.AI
	dashboardRouter := router.RouterGroupApp.Dashboard
	auditRouter := router.RouterGroupApp.Audit

	userRouter.InitUserRouter(orderFoodGroup)
	auditRouter.InitAuditRouter(orderFoodGroup)
	contentRouter.InitContentRouter(orderFoodGroup)
	engagementRouter.InitPointsRouter(orderFoodGroup)
	dishRouter.InitCatalogRouter(orderFoodGroup)
	dishRouter.InitRecommendationRouter(orderFoodGroup)
	dishRouter.InitOfficialDishRouter(orderFoodGroup)
	contentRouter.InitMediaRouter(orderFoodGroup)
	contentRouter.InitGovernanceRouter(orderFoodGroup)
	mealRouter.InitMealRouter(orderFoodGroup)
	mealRouter.InitShoppingListRouter(orderFoodGroup)
	dashboardRouter.InitDashboardRouter(orderFoodGroup)
	aiRouter.InitAIUsageRouter(orderFoodGroup)
	contentRouter.InitModerationRouter(orderFoodGroup)
	aiRouter.InitAIRouter(orderFoodGroup)
	dishRouter.InitSuggestionCatalogRouter(orderFoodGroup)
	engagementRouter.InitSubscriptionRouter(orderFoodGroup)
	userRouter.InitWeChatConfigRouter(orderFoodGroup)

}

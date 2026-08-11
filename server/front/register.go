package front

import (
	frontMiddleware "github.com/dyjh/order-food-mini-app/server/front/middleware"
	"github.com/dyjh/order-food-mini-app/server/front/router"
	"github.com/gin-gonic/gin"
)

func Register(publicGroup *gin.RouterGroup) {
	// The mini-program is an independent front module. It deliberately does
	// not inherit administrator JWT or Casbin middleware.
	miniAppRoot := publicGroup.Group("api/miniapp/v1")
	miniAppPublic := miniAppRoot.Group("")
	miniAppPrivate := miniAppRoot.Group("")
	miniAppPrivate.Use((frontMiddleware.AuthMiddleware{}).Handle())
	miniAppPrivate.Use((frontMiddleware.IdempotencyMiddleware{}).Handle())

	authRouter := router.RouterGroupApp.Auth
	systemRouter := router.RouterGroupApp.System
	profileRouter := router.RouterGroupApp.Profile
	contentRouter := router.RouterGroupApp.Content
	engagementRouter := router.RouterGroupApp.Engagement
	mealRouter := router.RouterGroupApp.Meal
	assistRouter := router.RouterGroupApp.Assist

	authRouter.InitAuthRouter(miniAppPublic)
	systemRouter.InitSystemRouter(miniAppPrivate)
	profileRouter.InitProfileRouter(miniAppPrivate)
	contentRouter.InitContentRouter(miniAppPrivate)
	engagementRouter.InitEngagementRouter(miniAppPrivate)
	mealRouter.InitMealRouter(miniAppPrivate, miniAppPublic)
	assistRouter.InitAssistRouter(miniAppPrivate)
}

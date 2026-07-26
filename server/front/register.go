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

	router.RouterGroupApp.InitAuthRouter(miniAppPublic)
	router.RouterGroupApp.InitSystemRouter(miniAppPrivate)
	router.RouterGroupApp.InitProfileRouter(miniAppPrivate)
	router.RouterGroupApp.InitContentRouter(miniAppPrivate)
	router.RouterGroupApp.InitEngagementRouter(miniAppPrivate)
	router.RouterGroupApp.InitMealRouter(miniAppPrivate, miniAppPublic)
	router.RouterGroupApp.InitAssistRouter(miniAppPrivate)
}

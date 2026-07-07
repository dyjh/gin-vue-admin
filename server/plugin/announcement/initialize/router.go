package initialize

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/dyjh/order-food-mini-app/server/plugin/announcement/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.Router.Info.Init(public, private)
}

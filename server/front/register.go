package front

import (
	"github.com/dyjh/order-food-mini-app/server/front/router"
	"github.com/gin-gonic/gin"
)

func Register(publicGroup *gin.RouterGroup) {
	frontGroup := publicGroup.Group("front")
	router.RouterGroupApp.InitTestRouter(frontGroup)
}

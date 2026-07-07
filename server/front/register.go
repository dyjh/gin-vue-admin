package front

import (
	"github.com/flipped-aurora/gin-vue-admin/server/front/router"
	"github.com/gin-gonic/gin"
)

func Register(publicGroup *gin.RouterGroup) {
	frontGroup := publicGroup.Group("front")
	router.RouterGroupApp.InitTestRouter(frontGroup)
}

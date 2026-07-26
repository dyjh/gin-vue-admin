package router

import (
	"github.com/dyjh/order-food-mini-app/server/front/api"
	"github.com/gin-gonic/gin"
)

type TestRouter struct{}

func (r *TestRouter) InitTestRouter(publicGroup *gin.RouterGroup) {
	testRouter := publicGroup.Group("test")
	testRouter.GET("", (&api.TestApi{}).Ping)
}

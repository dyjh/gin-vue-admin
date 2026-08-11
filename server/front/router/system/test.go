package system

import "github.com/gin-gonic/gin"

type TestRouter struct{}

func (r *TestRouter) InitTestRouter(publicGroup *gin.RouterGroup) {
	testRouter := publicGroup.Group("test")
	testRouter.GET("", testApi.Ping)
}

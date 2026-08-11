package profile

import "github.com/gin-gonic/gin"

type ProfileRouter struct{}

func (*ProfileRouter) InitProfileRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/profile", profileApi.Get)
	privateGroup.PUT("/profile", profileApi.Update)
}

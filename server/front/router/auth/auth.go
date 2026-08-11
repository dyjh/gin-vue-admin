package auth

import "github.com/gin-gonic/gin"

type AuthRouter struct{}

func (*AuthRouter) InitAuthRouter(publicGroup *gin.RouterGroup) {
	publicGroup.POST("/auth/wx-login", authApi.WxLogin)
}

package user

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/business/api/v1"
	"github.com/gin-gonic/gin"
)

type AuthRouter struct {
}

func (e *AuthRouter) InitAuthRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("auth")
	userAuthApi := v1.ApiGroupApp.AuthApiGroup.AuthApi
	{
		userRouter.POST("login", userAuthApi.Login) // 创建客户
	}
}

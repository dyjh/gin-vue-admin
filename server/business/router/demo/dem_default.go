package demo

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/business/api/v1"
	"github.com/gin-gonic/gin"
)

type DefaultRouter struct {
}

func (e *DefaultRouter) InitDefaultRouter(Router *gin.RouterGroup) {
	defaultRouter := Router.Group("default")
	demDefaultApi := v1.ApiGroupApp.DefaultApiGroup.DefaultApi
	{
		defaultRouter.GET("index", demDefaultApi.Index) // 创建客户
	}
}

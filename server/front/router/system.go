package router

import "github.com/gin-gonic/gin"

type SystemRouter struct{}

func (*SystemRouter) InitSystemRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/bootstrap", systemApi.Bootstrap)
	privateGroup.GET("/runtime-config", systemApi.RuntimeConfig)
	privateGroup.GET("/metadata", contentApi.Metadata)
}

package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (router *UserRouter) InitUserRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("/orderfood")
	api := orderfoodApi.ApiGroupApp.UserApi
	{
		group.GET("/users", api.List)
		group.GET("/users/:userId", api.Detail)
		group.PUT("/users/:userId/status", api.UpdateStatus)
		group.PUT("/users/:userId/capability", api.UpdateCapability)
		group.GET("/users/:userId/preference-profile", api.PreferenceProfile)
	}
}

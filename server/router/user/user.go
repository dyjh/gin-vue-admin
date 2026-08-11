package user

import "github.com/gin-gonic/gin"

type UserRouter struct{}

func (router *UserRouter) InitUserRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	{
		group.GET("/users", userApi.List)
		group.GET("/users/:userId", userApi.Detail)
		group.PUT("/users/:userId/status", userApi.UpdateStatus)
		group.PUT("/users/:userId/capability", userApi.UpdateCapability)
		group.GET("/users/:userId/preference-profile", userApi.PreferenceProfile)
	}
}

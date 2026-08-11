package dish

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// OfficialDishRouter 提供官方菜品和官方封面路由注册能力。
type OfficialDishRouter struct{}

// InitOfficialDishRouter 注册官方菜品和官方封面管理路由。
func (router *OfficialDishRouter) InitOfficialDishRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/official-dishes", officialDishApi.ListOfficialDishes)
		group.GET("/official-dishes/:dishId", officialDishApi.GetOfficialDish)
		group.GET("/official-dish-covers", officialDishApi.ListOfficialDishCovers)
		mutationGroup.POST("/official-dishes", officialDishApi.CreateOfficialDish)
		mutationGroup.PUT("/official-dishes/:dishId", officialDishApi.UpdateOfficialDish)
		mutationGroup.DELETE("/official-dishes/:dishId", officialDishApi.DeleteOfficialDish)
		mutationGroup.POST("/official-dish-covers", officialDishApi.UploadOfficialDishCover)
	}
}

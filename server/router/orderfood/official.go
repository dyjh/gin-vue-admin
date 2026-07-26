package orderfood

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1/orderfood"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// OfficialDishRouter 提供官方菜品和官方封面路由注册能力。
type OfficialDishRouter struct{}

// InitOfficialDishRouter 注册官方菜品和官方封面管理路由。
func (router *OfficialDishRouter) InitOfficialDishRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.OfficialDishApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/official-dishes", api.ListOfficialDishes)
		group.GET("/official-dishes/:dishId", api.GetOfficialDish)
		group.GET("/official-dish-covers", api.ListOfficialDishCovers)
		mutationGroup.POST("/official-dishes", api.CreateOfficialDish)
		mutationGroup.PUT("/official-dishes/:dishId", api.UpdateOfficialDish)
		mutationGroup.DELETE("/official-dishes/:dishId", api.DeleteOfficialDish)
		mutationGroup.POST("/official-dish-covers", api.UploadOfficialDishCover)
	}
}

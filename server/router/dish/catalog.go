package dish

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// CatalogRouter 提供分类、标签和单位的管理端路由注册能力。
type CatalogRouter struct{}

// InitCatalogRouter 注册分类、标签和单位的查询与变更路由。
func (router *CatalogRouter) InitCatalogRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/categories", catalogApi.ListCategories)
		mutationGroup.POST("/categories", catalogApi.CreateCategory)
		mutationGroup.PUT("/categories/sort-order", catalogApi.UpdateCategorySortOrder)
		mutationGroup.PUT("/categories/:categoryId", catalogApi.UpdateCategory)
		mutationGroup.DELETE("/categories/:categoryId", catalogApi.DeleteCategory)

		group.GET("/tags", catalogApi.ListTags)
		mutationGroup.POST("/tags", catalogApi.CreateTag)
		mutationGroup.PUT("/tags/sort-order", catalogApi.UpdateTagSortOrder)
		mutationGroup.PUT("/tags/:tagId", catalogApi.UpdateTag)
		mutationGroup.DELETE("/tags/:tagId", catalogApi.DeleteTag)

		group.GET("/units", catalogApi.ListUnits)
		mutationGroup.POST("/units", catalogApi.CreateUnit)
		mutationGroup.PUT("/units/sort-order", catalogApi.UpdateUnitSortOrder)
		mutationGroup.PUT("/units/:unitId", catalogApi.UpdateUnit)
		mutationGroup.DELETE("/units/:unitId", catalogApi.DeleteUnit)
	}
}

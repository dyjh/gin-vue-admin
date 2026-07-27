package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// CatalogRouter 提供分类、标签和单位的管理端路由注册能力。
type CatalogRouter struct{}

// InitCatalogRouter 注册分类、标签和单位的查询与变更路由。
func (router *CatalogRouter) InitCatalogRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.CatalogApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/categories", api.ListCategories)
		mutationGroup.POST("/categories", api.CreateCategory)
		mutationGroup.PUT("/categories/sort-order", api.UpdateCategorySortOrder)
		mutationGroup.PUT("/categories/:categoryId", api.UpdateCategory)
		mutationGroup.DELETE("/categories/:categoryId", api.DeleteCategory)

		group.GET("/tags", api.ListTags)
		mutationGroup.POST("/tags", api.CreateTag)
		mutationGroup.PUT("/tags/sort-order", api.UpdateTagSortOrder)
		mutationGroup.PUT("/tags/:tagId", api.UpdateTag)
		mutationGroup.DELETE("/tags/:tagId", api.DeleteTag)

		group.GET("/units", api.ListUnits)
		mutationGroup.POST("/units", api.CreateUnit)
		mutationGroup.PUT("/units/sort-order", api.UpdateUnitSortOrder)
		mutationGroup.PUT("/units/:unitId", api.UpdateUnit)
		mutationGroup.DELETE("/units/:unitId", api.DeleteUnit)
	}
}

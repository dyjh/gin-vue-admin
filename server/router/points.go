package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

type PointsRouter struct{}

func (router *PointsRouter) InitPointsRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.PointsApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/point-entries", api.ListPointEntries)
		group.GET("/point-rules", api.GetPointRule)
		group.POST("/point-adjustments/preview", api.PreviewPointAdjustment)
		mutationGroup.PUT("/point-rules", api.UpdatePointRule)
		mutationGroup.POST("/point-adjustments", api.CreatePointAdjustment)
	}
}

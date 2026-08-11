package engagement

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

type PointsRouter struct{}

func (router *PointsRouter) InitPointsRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/point-entries", pointsApi.ListPointEntries)
		group.GET("/point-rules", pointsApi.GetPointRule)
		group.POST("/point-adjustments/preview", pointsApi.PreviewPointAdjustment)
		mutationGroup.PUT("/point-rules", pointsApi.UpdatePointRule)
		mutationGroup.POST("/point-adjustments", pointsApi.CreatePointAdjustment)
	}
}

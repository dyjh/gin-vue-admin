package dish

import (
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// RecommendationRouter 提供可发现候选菜品和推荐精选路由注册能力。
type RecommendationRouter struct{}

// InitRecommendationRouter 注册可发现候选菜品和推荐精选管理路由。
func (router *RecommendationRouter) InitRecommendationRouter(orderFoodGroup *gin.RouterGroup) {
	group := orderFoodGroup
	mutationGroup := orderFoodGroup.Group("").Use(middleware.OperationRecord())
	{
		group.GET("/discoverable-dishes", recommendationApi.ListDiscoverableDishes)
		group.GET("/discoverable-dishes/:dishId", recommendationApi.GetDiscoverableDish)

		group.GET("/recommendations", recommendationApi.ListRecommendations)
		group.GET("/recommendations/:recommendationId", recommendationApi.GetRecommendation)
		mutationGroup.POST("/recommendations", recommendationApi.CreateRecommendation)
		mutationGroup.PUT("/recommendations/sort-order", recommendationApi.UpdateRecommendationSortOrder)
		mutationGroup.PUT("/recommendations/:recommendationId", recommendationApi.UpdateRecommendation)
		mutationGroup.DELETE("/recommendations/:recommendationId", recommendationApi.DeleteRecommendation)
		mutationGroup.POST("/recommendations/:recommendationId/publish", recommendationApi.PublishRecommendation)
		mutationGroup.POST("/recommendations/:recommendationId/offline", recommendationApi.OfflineRecommendation)
	}
}

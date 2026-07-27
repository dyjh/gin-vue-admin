package router

import (
	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// RecommendationRouter 提供可发现候选菜品和推荐精选路由注册能力。
type RecommendationRouter struct{}

// InitRecommendationRouter 注册可发现候选菜品和推荐精选管理路由。
func (router *RecommendationRouter) InitRecommendationRouter(
	privateGroup *gin.RouterGroup,
	api *orderfoodApi.RecommendationApi,
) {
	group := privateGroup.Group("/orderfood")
	mutationGroup := privateGroup.Group("/orderfood").Use(middleware.OperationRecord())
	{
		group.GET("/discoverable-dishes", api.ListDiscoverableDishes)
		group.GET("/discoverable-dishes/:dishId", api.GetDiscoverableDish)

		group.GET("/recommendations", api.ListRecommendations)
		group.GET("/recommendations/:recommendationId", api.GetRecommendation)
		mutationGroup.POST("/recommendations", api.CreateRecommendation)
		mutationGroup.PUT("/recommendations/sort-order", api.UpdateRecommendationSortOrder)
		mutationGroup.PUT("/recommendations/:recommendationId", api.UpdateRecommendation)
		mutationGroup.DELETE("/recommendations/:recommendationId", api.DeleteRecommendation)
		mutationGroup.POST("/recommendations/:recommendationId/publish", api.PublishRecommendation)
		mutationGroup.POST("/recommendations/:recommendationId/offline", api.OfflineRecommendation)
	}
}

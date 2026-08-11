package assist

import "github.com/gin-gonic/gin"

type AssistRouter struct{}

func (*AssistRouter) InitAssistRouter(privateGroup *gin.RouterGroup) {
	privateGroup.POST("/assist/dish-extraction", assistApi.DishExtraction)
	privateGroup.POST("/assist/dish-covers", assistApi.DishCover)
	privateGroup.GET("/meal-suggestions/status", assistApi.SuggestionStatus)
	privateGroup.POST("/meal-suggestions", assistApi.CreateSuggestion)
	privateGroup.GET("/meal-suggestions/:suggestionId", assistApi.Suggestion)
	privateGroup.POST("/meal-suggestions/:suggestionId/copy", assistApi.CopySuggestion)
	privateGroup.POST("/meal-suggestions/:suggestionId/feedback", assistApi.SuggestionFeedback)
	privateGroup.GET("/prep-plans/quote", assistApi.PrepQuote)
	privateGroup.POST("/prep-plans", assistApi.GeneratePrepPlan)
	privateGroup.GET("/prep-plans/:planId", assistApi.PrepPlan)
	privateGroup.POST("/prep-plans/:planId/feedback", assistApi.PrepFeedback)
}

package v1

import (
	"github.com/dyjh/order-food-mini-app/server/api/v1/example"
	"github.com/dyjh/order-food-mini-app/server/api/v1/system"
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	"github.com/dyjh/order-food-mini-app/server/service"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// ApiGroup aggregates system, example, and business APIs.
type ApiGroup struct {
	SystemApiGroup  system.ApiGroup
	ExampleApiGroup example.ApiGroup

	UserApi
	AuditApi
	ContentApi           *ContentApi
	PointsApi            *PointsApi
	CatalogApi           *CatalogApi
	RecommendationApi    *RecommendationApi
	OfficialDishApi      *OfficialDishApi
	MediaApi             *MediaApi
	GovernanceApi        *GovernanceApi
	MealAdminApi         *MealAdminApi
	DashboardApi         *DashboardApi
	AIUsageApi           *AIUsageApi
	ModerationApi        *ModerationApi
	AIApi                *AIApi
	SuggestionCatalogApi *SuggestionCatalogApi
	SubscriptionApi      *SubscriptionApi
	WeChatConfigApi      *WeChatConfigApi
}

var (
	orderFoodServiceGroup = service.ServiceGroupApp
	ApiGroupApp           = &ApiGroup{
		ContentApi:           NewContentApi(orderFoodServiceGroup.Content),
		PointsApi:            NewPointsApi(orderFoodServiceGroup.Points),
		CatalogApi:           NewCatalogApi(orderFoodServiceGroup.Catalog),
		RecommendationApi:    NewRecommendationApi(orderFoodServiceGroup.Recommendation),
		OfficialDishApi:      NewOfficialDishApi(orderFoodServiceGroup.OfficialDish),
		MediaApi:             NewMediaApi(orderFoodServiceGroup.Media),
		GovernanceApi:        NewGovernanceApi(orderFoodServiceGroup.Governance),
		MealAdminApi:         NewMealAdminApi(orderFoodServiceGroup.MealAdmin),
		DashboardApi:         NewDashboardApi(orderFoodServiceGroup.Dashboard),
		AIUsageApi:           NewAIUsageApi(orderFoodServiceGroup.AIUsage),
		ModerationApi:        NewModerationApi(orderFoodServiceGroup.Moderation),
		AIApi:                NewAIApi(orderFoodServiceGroup.AI),
		SuggestionCatalogApi: NewSuggestionCatalogApi(orderFoodServiceGroup.SuggestionCatalog),
		SubscriptionApi:      NewSubscriptionApi(orderFoodServiceGroup.Subscription),
		WeChatConfigApi:      NewWeChatConfigApi(orderFoodServiceGroup.WeChatConfig),
	}
)

func bindAndVerify(c *gin.Context, value interface{}, bind func(interface{}) error) bool {
	if err := bind(value); err != nil {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return false
	}
	if err := utils.VerifyAll(value); err != nil {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return false
	}
	return true
}

func requirePermission(c *gin.Context, permission string) bool {
	err := orderFoodServiceGroup.Permission.Require(
		c.Request.Context(),
		utils.GetUserAuthorityId(c),
		permission,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return false
	}
	return true
}

package orderfood

import (
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// ApiGroup 聚合当前模块的 API 处理器。
type ApiGroup struct {
	UserApi                                    // 用户接口
	AuditApi                                   // 审计接口
	ContentApi           *ContentApi           // 用户内容接口
	PointsApi            *PointsApi            // 积分接口
	CatalogApi           *CatalogApi           // 基础数据接口
	RecommendationApi    *RecommendationApi    // 推荐精选接口
	OfficialDishApi      *OfficialDishApi      // 官方菜品接口
	MediaApi             *MediaApi             // 图片资源接口
	GovernanceApi        *GovernanceApi        // 违规处理记录接口
	MealAdminApi         *MealAdminApi         // 饭局和采购清单管理接口
	DashboardApi         *DashboardApi         // 运营概览接口
	AIUsageApi           *AIUsageApi           // AI调用记录接口
	ModerationApi        *ModerationApi        // 图片审核配置接口
	AIApi                *AIApi                // AI配置接口
	SuggestionCatalogApi *SuggestionCatalogApi // 标准菜品索引接口
	SubscriptionApi      *SubscriptionApi      // 订阅消息接口
	WeChatConfigApi      *WeChatConfigApi      // 微信小程序配置接口
}

var (
	orderFoodServiceGroup = orderfoodService.ServiceGroupApp
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

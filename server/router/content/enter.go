package content

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	ContentRouter
	MediaRouter
	GovernanceRouter
	ModerationRouter
}

var (
	contentApi    = api.ApiGroupApp.ContentApiGroup.ContentApi
	mediaApi      = api.ApiGroupApp.ContentApiGroup.MediaApi
	governanceApi = api.ApiGroupApp.ContentApiGroup.GovernanceApi
	moderationApi = api.ApiGroupApp.ContentApiGroup.ModerationApi
)

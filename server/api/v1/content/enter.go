package content

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合内容、媒体、治理与审核接口。
type ApiGroup struct {
	ContentApi    // 用户内容管理接口
	MediaApi      // 媒体资源接口
	GovernanceApi // 内容治理接口
	ModerationApi // 图片审核接口
}

var (
	contentService    = service.ServiceGroupApp.ContentServiceGroup.Content
	mediaService      = service.ServiceGroupApp.ContentServiceGroup.Media
	governanceService = service.ServiceGroupApp.ContentServiceGroup.Governance
	moderationService = service.ServiceGroupApp.ContentServiceGroup.Moderation
)

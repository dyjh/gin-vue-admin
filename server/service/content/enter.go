package content

// ServiceGroup aggregates content-management services.
type ServiceGroup struct {
	Content    *ContentService    // 用户内容管理服务
	Media      *MediaService      // 媒体资源服务
	Governance *GovernanceService // 内容治理服务
	Moderation *ModerationService // 图片审核服务
}

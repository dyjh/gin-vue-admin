package content

const (
	MediaReviewPending     = "pending"
	MediaReviewPassed      = "passed"
	MediaReviewRejected    = "rejected"
	MediaReviewFailed      = "failed"
	MediaReviewNotRequired = "not_required"
)

// UsableMediaReviewStatuses 返回允许绑定到业务对象的图片审核状态。
func UsableMediaReviewStatuses() []string {
	return []string{MediaReviewPassed, MediaReviewNotRequired}
}

// IsMediaReviewUsable 判断图片是否已审核通过或按当前配置无需审核。
func IsMediaReviewUsable(status string) bool {
	return status == MediaReviewPassed || status == MediaReviewNotRequired
}

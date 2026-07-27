package model

import "testing"

// TestMediaReviewUsability 验证审核通过和无需审核的图片都可绑定业务对象。
func TestMediaReviewUsability(t *testing.T) {
	for _, status := range []string{MediaReviewPassed, MediaReviewNotRequired} {
		if !IsMediaReviewUsable(status) {
			t.Fatalf("status %q should be usable", status)
		}
	}
	for _, status := range []string{MediaReviewPending, MediaReviewRejected, MediaReviewFailed, ""} {
		if IsMediaReviewUsable(status) {
			t.Fatalf("status %q should not be usable", status)
		}
	}
}

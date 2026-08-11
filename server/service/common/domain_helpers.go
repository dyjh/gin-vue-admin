package common

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
)

const (
	PlatformPolicySingletonKey    = "platform"
	RecommendationStatusPublished = "published"
	RecommendationStatusOffline   = "offline"
	RecommendationPositionHome    = "home_featured"
	MediaResourceActive           = "active"
)

// ValidateReason validates the shared administrator reason contract.
func ValidateReason(reason string) error {
	length := utf8.RuneCountInString(strings.TrimSpace(reason))
	if length < 4 || length > 200 {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// ValidateAdminActor validates the shared administrator request context.
func ValidateAdminActor(actor commonRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

// AdministratorSummary converts persisted administrator fields to an API summary.
func AdministratorSummary(
	id uint,
	username string,
	nickname *string,
) commonResponse.AdministratorSummary {
	return commonResponse.AdministratorSummary{
		ID:       strconv.FormatUint(uint64(id), 10),
		Username: username,
		Nickname: nickname,
	}
}

// UserReference converts a mini-app user to the shared response reference.
func UserReference(user userModel.MiniAppUser) commonResponse.UserReference {
	return commonResponse.UserReference{
		ID: user.ID, Nickname: user.Nickname, AvatarURL: user.AvatarURL, Status: string(user.Status),
	}
}

// ValidTimeRange reports whether an optional time range is ordered.
func ValidTimeRange(from, to *time.Time) bool {
	return from == nil || to == nil || !from.After(*to)
}

// ValidPagination validates the shared administrator pagination sizes.
func ValidPagination(page, pageSize int) bool {
	if page < 0 {
		return false
	}
	switch pageSize {
	case 0, 20, 50, 100:
		return true
	default:
		return false
	}
}

// DecodeIdempotentResult decodes a typed response saved by the idempotency service.
func DecodeIdempotentResult[T any](
	raw json.RawMessage,
	replayed bool,
	err error,
) (T, bool, error) {
	var zero T
	if err != nil {
		return zero, false, err
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return zero, false, appErrors.AdminInternal.Wrap(err, "decode idempotent AI response")
	}
	return result, replayed, nil
}

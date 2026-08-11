package common

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	"github.com/dyjh/order-food-mini-app/server/service"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// BindAndVerify binds an HTTP value and applies the shared validator rules.
func BindAndVerify(c *gin.Context, value interface{}, bind func(interface{}) error) bool {
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

// RequirePermission checks the current GVA authority against a business permission.
func RequirePermission(c *gin.Context, permission string) bool {
	err := service.ServiceGroupApp.CommonServiceGroup.Permission.Require(
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

// HasPermission reports whether the current authority owns a permission
// without writing an error response.
func HasPermission(c *gin.Context, permission string) bool {
	return service.ServiceGroupApp.CommonServiceGroup.Permission.Require(
		c.Request.Context(),
		utils.GetUserAuthorityId(c),
		permission,
	) == nil
}

// CurrentAdminActor builds the current administrator actor from the Gin context.
func CurrentAdminActor(c *gin.Context) (commonRequest.AdminActor, bool) {
	claims := utils.GetUserInfo(c)
	if claims == nil || claims.BaseClaims.ID == 0 {
		response.FailWithBusinessError(appErrors.AdminLoginExpired.DefaultMsg(), c)
		return commonRequest.AdminActor{}, false
	}
	requestID := middleware.GetRequestID(c)
	if requestID == "" {
		response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
		return commonRequest.AdminActor{}, false
	}
	var nickname *string
	if value := strings.TrimSpace(claims.NickName); value != "" {
		nickname = &value
	}
	return commonRequest.AdminActor{
		AdministratorID:  claims.BaseClaims.ID,
		AuthorityID:      claims.AuthorityId,
		Username:         claims.Username,
		Nickname:         nickname,
		RequestID:        requestID,
		SourceIPMasked:   serviceCommon.MaskIP(c.ClientIP()),
		UserAgentSummary: c.Request.UserAgent(),
	}, true
}

// BindIdempotencyHeader binds and normalizes the shared idempotency header.
func BindIdempotencyHeader(c *gin.Context) (commonRequest.IdempotencyHeader, bool) {
	var header commonRequest.IdempotencyHeader
	if !BindAndVerify(c, &header, c.ShouldBindHeader) {
		return commonRequest.IdempotencyHeader{}, false
	}
	header.Key = strings.TrimSpace(header.Key)
	if header.Key == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return commonRequest.IdempotencyHeader{}, false
	}
	return header, true
}

// BindDishIDPath binds and normalizes a user-dish path identifier.
func BindDishIDPath(c *gin.Context) (contentRequest.UserDishIDPath, bool) {
	var path contentRequest.UserDishIDPath
	if !BindAndVerify(c, &path, c.ShouldBindUri) {
		return contentRequest.UserDishIDPath{}, false
	}
	path.DishID = strings.TrimSpace(path.DishID)
	if path.DishID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return contentRequest.UserDishIDPath{}, false
	}
	return path, true
}

// SetReplayHeader marks a replayed idempotent response.
func SetReplayHeader(c *gin.Context, replayed bool) {
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
}

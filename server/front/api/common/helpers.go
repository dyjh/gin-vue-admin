package common

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontMiddleware "github.com/dyjh/order-food-mini-app/server/front/middleware"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	frontService "github.com/dyjh/order-food-mini-app/server/front/service/experience"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// ServiceProfile 按当前运行时配置生成对外个人资料。
func ServiceProfile(
	user userModel.MiniAppUser,
	runtime frontResponse.RuntimeConfig,
) frontResponse.Profile {
	return frontService.ProfileForRuntime(user, runtime)
}

// Fail 记录业务错误并中止当前请求。
func Fail(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}

// RequireIdempotencyKey 读取并校验请求幂等键。
func RequireIdempotencyKey(c *gin.Context) (string, bool) {
	key := strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
	if key == "" || len(key) > 128 {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return "", false
	}
	return key, true
}

// CurrentUser 读取认证中间件写入的当前小程序用户。
func CurrentUser(c *gin.Context) (userModel.MiniAppUser, bool) {
	user, ok := frontMiddleware.CurrentUser(c)
	if !ok {
		Fail(c, appErrors.FrontLoginExpired.DefaultMsg())
		return userModel.MiniAppUser{}, false
	}
	return user, true
}

// BindJSON 绑定并校验 JSON 请求体。
func BindJSON(c *gin.Context, input interface{}) bool {
	if err := c.ShouldBindJSON(input); err != nil {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	if err := utils.VerifyAll(input); err != nil {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	return true
}

// BindQuery 绑定并校验查询参数。
func BindQuery(c *gin.Context, input interface{}) bool {
	if err := c.ShouldBindQuery(input); err != nil {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	if err := utils.VerifyAll(input); err != nil {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	return true
}

// BindURI 绑定并校验路径参数。
func BindURI(c *gin.Context, input interface{}) bool {
	if err := c.ShouldBindUri(input); err != nil {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	if err := utils.VerifyAll(input); err != nil {
		Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	return true
}

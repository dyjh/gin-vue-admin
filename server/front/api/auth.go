package api

import (
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// AuthApi 提供小程序认证接口处理能力。
type AuthApi struct{}

// WxLogin 使用微信一次性登录凭证换取业务令牌
// @Tags 小程序认证
// @Summary 使用微信一次性登录凭证换取业务令牌
// @Security NoAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.WxLoginInput true "请求参数"
// @Success 200 {object} frontResponse.Envelope{data=frontResponse.WxLoginResult,msg=string} "操作成功"
// @Router /miniapp/v1/auth/wx-login [post]
func (*AuthApi) WxLogin(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	var input frontRequest.WxLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return
	}
	if err := utils.VerifyAll(input); err != nil {
		fail(c, appErrors.AddErrorContext(
			appErrors.FrontBadRequest.DefaultMsg(),
			"code",
			"微信登录凭证不能为空",
		))
		return
	}
	login, err := authService.Login(c.Request.Context(), input.Code)
	if err != nil {
		fail(c, err)
		return
	}
	runtime, err := runtimeService.Current(c.Request.Context(), login.User.ID)
	if err != nil {
		fail(c, err)
		return
	}
	frontResponse.OK(c, frontResponse.WxLoginResult{
		AccessToken:   login.AccessToken,
		ExpiresIn:     login.ExpiresIn,
		IsNewUser:     login.IsNewUser,
		Profile:       serviceProfile(login.User, runtime),
		RuntimeConfig: runtime,
	})
}

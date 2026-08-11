package profile

import (
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	apiCommon "github.com/dyjh/order-food-mini-app/server/front/api/common"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// ProfileApi 提供小程序个人资料接口处理能力。
type ProfileApi struct{}

// Get 获取小程序个人资料
// @Tags 小程序个人资料
// @Summary 获取小程序个人资料
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} frontResponse.Envelope{data=frontResponse.Profile,msg=string} "获取成功"
// @Router /miniapp/v1/profile [get]
func (*ProfileApi) Get(c *gin.Context) {
	sessionUser, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	user, err := profileService.Get(c.Request.Context(), sessionUser.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	runtime, err := runtimeService.Current(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	frontResponse.OK(c, apiCommon.ServiceProfile(user, runtime))
}

// Update 更新小程序个人资料
// @Tags 小程序个人资料
// @Summary 更新小程序个人资料
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.ProfileUpdateInput true "请求参数"
// @Success 200 {object} frontResponse.Envelope{data=frontResponse.Profile,msg=string} "操作成功"
// @Router /miniapp/v1/profile [put]
func (*ProfileApi) Update(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	sessionUser, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.ProfileUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiCommon.Fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return
	}
	if err := utils.VerifyAll(input); err != nil {
		apiCommon.Fail(c, appErrors.AddErrorContext(
			appErrors.FrontBadRequest.DefaultMsg(),
			"nickname",
			"昵称为必填且不能超过 30 个字符",
		))
		return
	}
	user, err := profileService.Update(c.Request.Context(), sessionUser.ID, input)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	runtime, err := runtimeService.Current(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	frontResponse.OK(c, apiCommon.ServiceProfile(user, runtime))
}

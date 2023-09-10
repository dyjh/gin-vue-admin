package user

import (
	userReq "github.com/flipped-aurora/gin-vue-admin/server/business/request/user/auth"
	"github.com/flipped-aurora/gin-vue-admin/server/business/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram/config"
)

type AuthApi struct {
}

func (e *AuthApi) Login(c *gin.Context) {
	var loginForm userReq.LoginForm
	_ = c.BindJSON(&loginForm)
	err := utils.Validate(loginForm)
	if err != "" {
		response.FailWithMessage(err, c)
		return
	}
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &config.Config{
		AppID:     "xxx",
		AppSecret: "xxx",
		Cache:     memory,
	}
	miniProgram := wc.GetMiniProgram(cfg)
	session, err2 := miniProgram.GetAuth().Code2Session(loginForm.Code)
	if err2 != nil {
		response.FailWithMessage(err, c)
		return
	}
	global.GVA_LOG.Info(session.OpenID)
	response.OkWithMessage("hollow world!", c)
}

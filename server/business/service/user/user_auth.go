package user

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
)

type AuthService struct{}

func (exa *AuthService) Login(e example.ExaCustomer) (err error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &config.Config{
		AppID:     "wxe67232d535598db3",
		AppSecret: "181cd75975b501e24fea20ab78c61b90",
		Cache:     memory,
	}
	miniProgram := wc.GetMiniProgram(cfg)
	session, err2 := miniProgram.GetAuth().Code2Session(loginForm.Code)
	if err2 != nil {
		response.FailWithMessage(err2.Error(), c)
		return
	}
	global.GVA_LOG.Info("调用微信over2")
	global.GVA_LOG.Info(session.OpenID)
	openId := session.OpenID
}

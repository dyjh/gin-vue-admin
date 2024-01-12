package user

import (
	"fmt"
	userReq "github.com/flipped-aurora/gin-vue-admin/server/business/request/user/auth"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/business"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram/config"
	"math/rand"
)

type AuthService struct{}

func (exa *AuthService) Login(loginForm userReq.LoginForm) (Member *business.Members, err error) {
	isTest := global.GVA_CONFIG.Wechat.Test
	openid := ""
	if !isTest {
		wechatService := wechat.NewWechat()
		memory := cache.NewMemory()
		cfg := &config.Config{
			AppID:     global.GVA_CONFIG.Wechat.AppId,
			AppSecret: global.GVA_CONFIG.Wechat.AppSecret,
			Cache:     memory,
		}
		global.GVA_LOG.Info("开始调用微信")
		miniProgram := wechatService.GetMiniProgram(cfg)
		session, errWechat := miniProgram.GetAuth().Code2Session(loginForm.Code)
		if errWechat != nil {
			return nil, errWechat
		}
		global.GVA_LOG.Info("调用微信over")
		openid = session.OpenID
	} else {
		openid = loginForm.Code
	}
	global.GVA_LOG.Info(fmt.Sprintf("微信openid：%s", openid))
	member := &business.Members{}
	errSelectMember := global.GVA_DB.Where("openid = ?", openid).First(member).Error // 根据id查询api记录
	if errSelectMember != nil {
		if errSelectMember.Error() != "record not found" {
			return nil, errSelectMember
		}
		if errSelectMember.Error() == "record not found" {
			// 生成 100 到 999 之间的随机数
			randomNumber := rand.Intn(900) + 100
			result := fmt.Sprintf("用户%d", randomNumber)
			member.Nickname = result
			member.Openid = openid
			errInsertMember := global.GVA_DB.Create(&member).Error
			if errInsertMember != nil {
				return nil, errInsertMember
			}
		}
	}

	return member, nil
}

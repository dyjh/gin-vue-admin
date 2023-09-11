package user

import (
	"fmt"
	userReq "github.com/flipped-aurora/gin-vue-admin/server/business/request/user/auth"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/user"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram/config"
	"math/rand"
	"time"
)

type AuthService struct{}

func (exa *AuthService) Login(loginForm userReq.LoginForm, openid string) (Member *user.Members, err error) {
	if openid == "" {
		wc := wechat.NewWechat()
		memory := cache.NewMemory()
		cfg := &config.Config{
			AppID:     "xxxxx",
			AppSecret: "xxxxxxx",
			Cache:     memory,
		}
		global.GVA_LOG.Info("开始调用微信")
		miniProgram := wc.GetMiniProgram(cfg)
		session, errWechat := miniProgram.GetAuth().Code2Session(loginForm.Code)
		if errWechat != nil {
			return nil, errWechat
		}
		global.GVA_LOG.Info("调用微信over2")
		global.GVA_LOG.Info(session.OpenID)
		openid = session.OpenID
	}
	member := &user.Members{}
	errSelectMember := global.GVA_DB.Where("openid = ?", openid).First(member).Error // 根据id查询api记录
	if errSelectMember != nil {
		return nil, errSelectMember
	}
	if member == nil {
		rand.Seed(time.Now().UnixNano())
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
	return member, nil
}

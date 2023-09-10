package user

import (
	"errors"
	"fmt"
	userReq "github.com/flipped-aurora/gin-vue-admin/server/business/request/user/auth"
	"github.com/flipped-aurora/gin-vue-admin/server/business/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math/rand"
	"time"
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
	global.GVA_LOG.Info("开始调用微信")
	/*wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &config.Config{
		AppID:     "wxe67232d535598db3",
		AppSecret: "181cd75975b501e24fea20ab78c61b90",
		Cache:     memory,
	}*/
	//miniProgram := wc.GetMiniProgram(cfg)
	//session, err2 := miniProgram.GetAuth().Code2Session(loginForm.Code)
	//if err2 != nil {
	//	response.FailWithMessage(err2.Error(), c)
	//	return
	//}
	//global.GVA_LOG.Info("调用微信over2")
	//global.GVA_LOG.Info(session.OpenID)
	//openId := session.OpenID
	openId := "111231231"

	err3 := global.GVA_DB.Where("openid = ?", openId).First(&user.Members{}).Error // 根据id查询api记录
	if !errors.Is(err3, gorm.ErrRecordNotFound) {                                  // api记录不存在
		response.FailWithMessage("用户已注册", c)
		return
	}

	member := &user.Members{}
	rand.Seed(time.Now().UnixNano())

	// 生成 100 到 999 之间的随机数
	randomNumber := rand.Intn(900) + 100

	result := fmt.Sprintf("用户%d", randomNumber)
	member.Nickname = result
	member.Openid = openId

	err4 := global.GVA_DB.Create(&member).Error
	if err4 != nil {
		response.FailWithMessage(err4.Error(), c)
	}

	response.OkWithMessage("hollow world!", c)
}

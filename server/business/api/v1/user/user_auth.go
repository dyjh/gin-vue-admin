package user

import (
	userReq "github.com/flipped-aurora/gin-vue-admin/server/business/request/user/auth"
	userRes "github.com/flipped-aurora/gin-vue-admin/server/business/response/user"
	"github.com/flipped-aurora/gin-vue-admin/server/business/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/model/user"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AuthApi struct {
}

func (a *AuthApi) Login(c *gin.Context) {
	var loginForm userReq.LoginForm
	_ = c.BindJSON(&loginForm)
	err := utils.Validate(loginForm)
	if err != "" {
		response.FailWithMessage(err, c)
		return
	}

	member, errLogin := authService.Login(loginForm, "111231231")
	if errLogin != nil {
		response.FailWithMessage(errLogin.Error(), c)
		return
	}
	a.TokenNext(c, *member)
	return
}

// TokenNext 登录以后签发jwt
func (a *AuthApi) TokenNext(c *gin.Context, user user.Members) {
	j := &utils.BusinessJWT{SigningKey: []byte(global.GVA_CONFIG.JWT.SigningKey)} // 唯一签名
	claims := j.CreateClaims(userReq.BaseClaims{
		ID:       user.ID,
		Nickname: user.Nickname,
		Openid:   user.Openid,
	})
	token, err := j.CreateToken(claims)
	if err != nil {
		global.GVA_LOG.Error("获取token失败!", zap.Error(err))
		response.FailWithMessage("获取token失败", c)
		return
	}
	if !global.GVA_CONFIG.System.UseMultipoint {
		response.OkWithDetailed(userRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
		return
	}

	if jwtStr, err := jwtService.GetRedisJWT(user.Nickname); err == redis.Nil {
		if err := jwtService.SetRedisJWT(token, user.Nickname); err != nil {
			global.GVA_LOG.Error("设置登录状态失败!", zap.Error(err))
			response.FailWithMessage("设置登录状态失败", c)
			return
		}
		response.OkWithDetailed(userRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
	} else if err != nil {
		global.GVA_LOG.Error("设置登录状态失败!", zap.Error(err))
		response.FailWithMessage("设置登录状态失败", c)
	} else {
		var blackJWT system.JwtBlacklist
		blackJWT.Jwt = jwtStr

		if err := jwtService.SetRedisJWT(token, user.Nickname); err != nil {
			response.FailWithMessage("设置登录状态失败", c)
			return
		}
		response.OkWithDetailed(userRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
	}
}

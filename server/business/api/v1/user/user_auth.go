package user

import (
	userReq "github.com/flipped-aurora/gin-vue-admin/server/business/request/user/auth"
	userRes "github.com/flipped-aurora/gin-vue-admin/server/business/response/user"
	"github.com/flipped-aurora/gin-vue-admin/server/business/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/business"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	o_utils "github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type AuthApi struct {
}

// Login
// @Tags     Auth
// @Summary  用户登录/注册
// @Produce   application/json
// @Param    data  body      userReq.LoginForm                                         true  "微信js-code"
// @Success  200   {object}  response.Response{data=userRes.LoginResponse,msg=string}  "返回包括用户信息,token,过期时间"
// @Router   /auth/login [post]
func (a *AuthApi) Login(c *gin.Context) {
	var loginForm userReq.LoginForm
	_ = c.BindJSON(&loginForm)
	err := utils.Validate(loginForm)
	if err != "" {
		response.FailWithMessage(err, c)
		return
	}
	member, errLogin := authService.Login(loginForm)
	if errLogin != nil {
		response.FailWithMessage(errLogin.Error(), c)
		return
	}
	a.TokenNext(c, *member)
	return
}

// TokenNext 登录以后签发jwt
func (a *AuthApi) TokenNext(c *gin.Context, user business.Members) {
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
		o_utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
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
		o_utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		response.OkWithDetailed(userRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
	} else if err != nil {
		global.GVA_LOG.Error("设置登录状态失败!", zap.Error(err))
		response.FailWithMessage("设置登录状态失败", c)
	} else {
		var blackJWT business.BusJwtBlacklist
		blackJWT.Jwt = jwtStr
		if err := jwtService.JsonInBlacklist(blackJWT); err != nil {
			response.FailWithMessage("jwt作废失败", c)
			return
		}
		if err := jwtService.SetRedisJWT(token, user.Nickname); err != nil {
			response.FailWithMessage("设置登录状态失败", c)
			return
		}
		o_utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		response.OkWithDetailed(userRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
	}
}
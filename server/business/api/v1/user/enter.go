package user

import "github.com/flipped-aurora/gin-vue-admin/server/business/service"

type ApiGroup struct {
	AuthApi
}

var (
	authService = service.BusinessServiceGroupApp.UserServiceGroup.AuthService
	jwtService  = service.BusinessServiceGroupApp.UserServiceGroup.JwtService
)

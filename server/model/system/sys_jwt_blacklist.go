package system

import (
	"github.com/dyjh/order-food-mini-app/server/global"
)

type JwtBlacklist struct {
	global.GVA_MODEL
	Jwt string `gorm:"type:text;comment:jwt"`
}

package business

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type BusJwtBlacklist struct {
	global.GVA_MODEL
	Jwt string `gorm:"type:text;comment:jwt"`
}

func (BusJwtBlacklist) TableName() string {
	return "bus_jwt_blacklists"
}

package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/business/service/user"
)

type ServiceGroup struct {
	UserServiceGroup user.ServiceGroup
}

var BusinessServiceGroupApp = new(ServiceGroup)

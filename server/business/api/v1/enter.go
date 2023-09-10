package v1

import (
	"github.com/flipped-aurora/gin-vue-admin/server/business/api/v1/demo"
	"github.com/flipped-aurora/gin-vue-admin/server/business/api/v1/user"
)

type ApiGroup struct {
	DefaultApiGroup demo.ApiGroup
	AuthApiGroup    user.ApiGroup
}

var ApiGroupApp = new(ApiGroup)

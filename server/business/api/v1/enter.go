package v1

import (
	"github.com/flipped-aurora/gin-vue-admin/server/business/api/v1/demo"
)

type ApiGroup struct {
	DefaultApiGroup demo.ApiGroup
}

var ApiGroupApp = new(ApiGroup)

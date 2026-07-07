package request

import (
	"github.com/dyjh/order-food-mini-app/server/model/common/request"
	"github.com/dyjh/order-food-mini-app/server/model/system"
)

type SysApiTokenSearch struct {
	system.SysApiToken
	request.PageInfo
	Status *bool `json:"status" form:"status"`
}

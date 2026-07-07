package request

import (
	"github.com/dyjh/order-food-mini-app/server/model/common/request"
	"github.com/dyjh/order-food-mini-app/server/model/system"
)

type SysLoginLogSearch struct {
	system.SysLoginLog
	request.PageInfo
}

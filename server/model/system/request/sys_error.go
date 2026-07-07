package request

import (
	"github.com/dyjh/order-food-mini-app/server/model/common/request"
	"time"
)

type SysErrorSearch struct {
	CreatedAtRange []time.Time `json:"createdAtRange" form:"createdAtRange[]"`
	Form           *string     `json:"form" form:"form"`
	Info           *string     `json:"info" form:"info"`
	request.PageInfo
}

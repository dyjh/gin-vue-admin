package response

import (
	"github.com/dyjh/order-food-mini-app/server/model/system/request"
)

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}

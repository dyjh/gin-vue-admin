package response

import "github.com/dyjh/order-food-mini-app/server/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}

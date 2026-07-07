package system

import (
	"github.com/dyjh/order-food-mini-app/server/config"
)

// 配置文件结构体
type System struct {
	Config config.Server `json:"config"`
}

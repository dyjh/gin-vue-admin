package system

import frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"

// TestService 提供前台测试业务能力。
type TestService struct{}

// Ping 检查前台模块服务是否可用。
func (s *TestService) Ping() frontResponse.TestResponse {
	return frontResponse.TestResponse{Message: "pong"}
}

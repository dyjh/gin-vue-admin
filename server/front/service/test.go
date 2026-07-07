package service

import frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"

type TestService struct{}

func (s *TestService) Ping() frontResponse.TestResponse {
	return frontResponse.TestResponse{Message: "pong"}
}

package service

import frontResponse "github.com/flipped-aurora/gin-vue-admin/server/front/response"

type TestService struct{}

func (s *TestService) Ping() frontResponse.TestResponse {
	return frontResponse.TestResponse{Message: "pong"}
}

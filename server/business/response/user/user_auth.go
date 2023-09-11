package user

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/user"
)

type LoginResponse struct {
	User      user.Members `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt int64        `json:"expiresAt"`
}

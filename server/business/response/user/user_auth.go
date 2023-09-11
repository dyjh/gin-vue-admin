package user

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/business"
)

type LoginResponse struct {
	User      business.Members `json:"user"`
	Token     string           `json:"token"`
	ExpiresAt int64            `json:"expiresAt"`
}

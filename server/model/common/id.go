package common

import (
	"github.com/google/uuid"
)

// NewID 生成业务公开ID。
func NewID() string {
	return uuid.NewString()
}

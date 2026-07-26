package middleware

import (
	"regexp"

	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDContextKey = "request_id"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9_-]{8,128}$`)

// RequestID guarantees a safe trace identifier for every request. Untrusted
// values outside the frozen character and length limits are replaced.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if !validRequestID.MatchString(requestID) {
			requestID = uuid.NewString()
		}
		c.Set(RequestIDContextKey, requestID)
		// 同步写入标准请求上下文，供不依赖 Gin 的 Service 层记录调用链路。
		c.Request = c.Request.WithContext(utils.WithRequestID(c.Request.Context(), requestID))
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}

// GetRequestID 从Gin上下文读取请求追踪ID。
func GetRequestID(c *gin.Context) string {
	if value, ok := c.Get(RequestIDContextKey); ok {
		if requestID, valid := value.(string); valid {
			return requestID
		}
	}
	return ""
}

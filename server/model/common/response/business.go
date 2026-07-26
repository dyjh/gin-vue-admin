package response

import (
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/gin-gonic/gin"
)

// FailWithBusinessError writes a stable business error envelope. Wrapped SQL,
// provider and infrastructure messages are intentionally not returned.
func FailWithBusinessError(err error, c *gin.Context) {
	code := appErrors.GetType(err)
	var data interface{}
	if contextItems := appErrors.GetErrorContext(err); len(contextItems) > 0 {
		data = gin.H{"fields": contextItems}
	}
	Result(int(code), data, appErrors.SafeMessage(err), c)
}

func FailWithBusinessCode(code appErrors.ErrorType, c *gin.Context) {
	Result(int(code), nil, appErrors.Message(code), c)
}

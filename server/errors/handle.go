package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dyjh/order-food-mini-app/server/global"
	"github.com/gin-gonic/gin"
)

// ErrorHandlingMiddleware serializes errors attached through c.Error. Business
// handlers may alternatively use response.FailWithBusinessError directly.
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		status, code, trace, contextItems := handleError(c.Errors[0].Err)
		data := gin.H{}
		if len(contextItems) > 0 {
			data["fields"] = contextItems
		}
		if global.GVA_CONFIG.System.Env != "prod" && len(trace) > 0 {
			data["trace"] = trace
		}

		c.AbortWithStatusJSON(status, gin.H{
			"code": int(code),
			"data": data,
			"msg":  SafeMessage(c.Errors[0].Err),
		})
	}
}

func handleError(err error) (int, ErrorType, []string, []ErrorContext) {
	errorType := GetType(err)
	httpStatus := httpStatusFor(errorType)

	var trace []string
	if global.GVA_CONFIG.System.Env != "prod" {
		formatted := fmt.Sprintf("%+v", err)
		formatted = strings.ReplaceAll(formatted, "\n\t", " ")
		trace = strings.Split(formatted, "\n")
	}

	return httpStatus, errorType, trace, GetErrorContext(err)
}

func httpStatusFor(errorType ErrorType) int {
	switch errorType {
	case AdminBadRequest, FrontBadRequest:
		return http.StatusBadRequest
	case AdminLoginExpired, FrontLoginExpired, FrontWxLoginInvalid:
		return http.StatusUnauthorized
	case AdminNoPermission, AdminAccountUnavailable, FrontNoPermission, FrontUserDisabled,
		FrontFeatureDisabled, FrontPointsDisabled, FrontFeatureLocked:
		return http.StatusForbidden
	case AdminNotFound, FrontNotFound:
		return http.StatusNotFound
	case AdminStateConflict, AdminIdempotencyConflict, AdminAlreadyExists, AdminRequestProcessing,
		AdminResourceInUse, FrontStateConflict, FrontIdempotencyConflict, FrontAlreadyExists,
		FrontRequestProcessing:
		return http.StatusConflict
	case AdminImageRejected, AdminInvalidImage, AdminNegativePoints, AdminInvalidConfig,
		AdminInvalidGovernance, AdminIdentityKeyMissing, AdminIdentityKeyInvalid,
		FrontImageRejected, FrontInvalidImage, FrontInsufficientPoints, FrontResultInvalid:
		return http.StatusUnprocessableEntity
	case AdminRateLimited, FrontRateLimited, FrontQuotaExceeded:
		return http.StatusTooManyRequests
	case AdminProviderFailed, FrontExecutionRefunded:
		return http.StatusBadGateway
	case FrontRefundPending:
		return http.StatusServiceUnavailable
	case AdminProviderTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

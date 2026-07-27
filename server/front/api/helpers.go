package api

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontMiddleware "github.com/dyjh/order-food-mini-app/server/front/middleware"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

func fail(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}

func requireIdempotencyKey(c *gin.Context) (string, bool) {
	key := strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
	if key == "" || len(key) > 128 {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return "", false
	}
	return key, true
}

func currentUser(c *gin.Context) (orderfoodModel.MiniAppUser, bool) {
	user, ok := frontMiddleware.CurrentUser(c)
	if !ok {
		fail(c, appErrors.FrontLoginExpired.DefaultMsg())
		return orderfoodModel.MiniAppUser{}, false
	}
	return user, true
}

func bindJSON(c *gin.Context, input interface{}) bool {
	if err := c.ShouldBindJSON(input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	if err := utils.VerifyAll(input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	return true
}

func bindQuery(c *gin.Context, input interface{}) bool {
	if err := c.ShouldBindQuery(input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	if err := utils.VerifyAll(input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	return true
}

func bindURI(c *gin.Context, input interface{}) bool {
	if err := c.ShouldBindUri(input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	if err := utils.VerifyAll(input); err != nil {
		fail(c, appErrors.FrontBadRequest.DefaultMsg())
		return false
	}
	return true
}

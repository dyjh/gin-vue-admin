package demo

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type DefaultApi struct {
}

func (e *DefaultApi) Index(c *gin.Context) {

	response.OkWithMessage("hollow world!", c)
}

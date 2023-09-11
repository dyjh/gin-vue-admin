package demo

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type DefaultApi struct {
}

// Index
// @Tags     Default
// @Summary  demo接口
// @Produce   application/json
// @Success  200   {object}  response.Response{data=nil,msg=string}  "返回包括用户信息,token,过期时间"
// @Router   /default/index [get]
func (e *DefaultApi) Index(c *gin.Context) {

	response.OkWithMessage("hollow world!", c)
}

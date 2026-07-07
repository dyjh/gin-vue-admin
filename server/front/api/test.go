package api

import (
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type TestApi struct{}

// Ping
// @Tags      FrontTest
// @Summary   前台测试接口
// @Produce   application/json
// @Success   200  {object}  response.TestResponseEnvelope  "返回 pong"
// @Router    /front/test [get]
func (t *TestApi) Ping(c *gin.Context) {
	var result frontResponse.TestResponse = testService.Ping()
	response.OkWithDetailed(result, "成功", c)
}

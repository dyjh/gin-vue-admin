package system

import (
	frontResponse "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// TestApi 提供前台测试接口处理能力。
type TestApi struct{}

// Ping 检查前台模块服务是否可用
// @Tags 前台测试
// @Summary 检查前台模块服务是否可用
// @Security NoAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} frontResponse.Envelope{data=frontResponse.TestResponse,msg=string} "获取成功"
// @Router /front/test [get]
func (t *TestApi) Ping(c *gin.Context) {
	result := frontResponse.TestResponse{Message: "pong"}
	response.OkWithDetailed(result, "成功", c)
}

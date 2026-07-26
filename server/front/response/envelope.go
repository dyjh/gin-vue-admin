package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope 表示小程序接口统一响应结构。
type Envelope struct {
	Code int         `json:"code"` // 编码
	Data interface{} `json:"data"` // 数据
	Msg  string      `json:"msg"`  // 消息
}

// OK 使用小程序统一响应结构返回成功数据。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Data: data, Msg: "ok"})
}

// OkWithData 返回业务数据并保持与 GVA 相同的成功码语义。
func OkWithData(data interface{}, c *gin.Context) {
	OK(c, data)
}

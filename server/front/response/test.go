package response

// TestResponse 表示测试响应响应数据。
type TestResponse struct {
	Message string `json:"message"` // 消息
}

// TestResponseEnvelope 表示测试响应统一响应响应数据。
type TestResponseEnvelope struct {
	Code int          `json:"code"` // 编码
	Data TestResponse `json:"data"` // 数据
	Msg  string       `json:"msg"`  // 消息
}

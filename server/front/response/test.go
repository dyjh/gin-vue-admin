package response

type TestResponse struct {
	Message string `json:"message"`
}

type TestResponseEnvelope struct {
	Code int          `json:"code"`
	Data TestResponse `json:"data"`
	Msg  string       `json:"msg"`
}

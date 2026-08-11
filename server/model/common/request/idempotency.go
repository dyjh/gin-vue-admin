package request

// IdempotencyHeader 表示幂等请求头请求参数。
type IdempotencyHeader struct {
	Key string `header:"X-Idempotency-Key" json:"idempotencyKey" binding:"required,max=128"` // 键
}

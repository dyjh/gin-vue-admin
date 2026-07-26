package utils

import "context"

type requestIDContextKey struct{}
type idempotencyKeyContextKey struct{}

// WithRequestID 将请求追踪ID写入标准请求上下文。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

// RequestIDFromContext 从标准请求上下文读取请求追踪ID。
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

// WithIdempotencyKey 将已校验的幂等键写入标准请求上下文。
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyContextKey{}, key)
}

// IdempotencyKeyFromContext 从标准请求上下文读取已校验的幂等键。
func IdempotencyKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	key, _ := ctx.Value(idempotencyKeyContextKey{}).(string)
	return key
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDPreservesValidValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-Id", "request_12345678")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Body.String() != "request_12345678" {
		t.Fatalf("request id = %q", recorder.Body.String())
	}
	if recorder.Header().Get("X-Request-Id") != "request_12345678" {
		t.Fatal("response header did not preserve request id")
	}
}

func TestRequestIDReplacesUnsafeValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-Id", "bad\nvalue")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Body.String() == "" || recorder.Body.String() == "bad\nvalue" {
		t.Fatalf("unsafe request id was not replaced: %q", recorder.Body.String())
	}
}

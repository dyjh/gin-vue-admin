package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestHTTPWeChatSubscribeSenderRefreshesRejectedCachedToken 验证微信提前拒绝缓存令牌时只刷新并补发一次。
func TestHTTPWeChatSubscribeSenderRefreshesRejectedCachedToken(t *testing.T) {
	for _, rejectedCode := range []int{wechatInvalidTokenCode, wechatExpiredTokenCode} {
		t.Run(strconv.Itoa(rejectedCode), func(t *testing.T) {
			var mu sync.Mutex
			tokenRequests := 0
			sendTokens := make([]string, 0, 2)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				writer.Header().Set("Content-Type", "application/json")
				switch request.URL.Path {
				case "/token":
					tokenRequests++
					_ = json.NewEncoder(writer).Encode(map[string]interface{}{
						"access_token": "fresh-token",
						"expires_in":   7200,
					})
				case "/send":
					token := request.URL.Query().Get("access_token")
					sendTokens = append(sendTokens, token)
					if token == "stale-token" {
						_ = json.NewEncoder(writer).Encode(map[string]interface{}{
							"errcode": rejectedCode,
							"errmsg":  "token rejected",
						})
						return
					}
					_ = json.NewEncoder(writer).Encode(map[string]interface{}{
						"errcode": 0,
						"errmsg":  "ok",
					})
				default:
					http.NotFound(writer, request)
				}
			}))
			defer server.Close()

			sender := &HTTPWeChatSubscribeSender{
				Client:         server.Client(),
				AppID:          "test-app-id",
				AppSecret:      "test-app-secret",
				TokenEndpoint:  server.URL + "/token",
				SendEndpoint:   server.URL + "/send",
				accessToken:    "stale-token",
				tokenExpiresAt: time.Now().UTC().Add(time.Hour),
			}
			err := sender.Send(context.Background(), WeChatSubscribeRequest{
				OpenID: "openid", TemplateID: "template-id", Page: "pages/meal/history",
				Data: map[string]map[string]string{"thing1": {"value": "测试饭局"}},
			})
			if err != nil {
				t.Fatalf("send after token refresh failed: %v", err)
			}
			mu.Lock()
			defer mu.Unlock()
			if tokenRequests != 1 {
				t.Fatalf("token endpoint requests = %d, want 1", tokenRequests)
			}
			if len(sendTokens) != 2 ||
				sendTokens[0] != "stale-token" ||
				sendTokens[1] != "fresh-token" {
				t.Fatalf("send tokens = %#v, want stale then fresh", sendTokens)
			}
		})
	}
}

// TestHTTPWeChatSubscribeSenderDoesNotLoopOnRejectedFreshToken 验证刷新后的令牌仍被拒绝时立即返回业务错误。
func TestHTTPWeChatSubscribeSenderDoesNotLoopOnRejectedFreshToken(t *testing.T) {
	var mu sync.Mutex
	tokenRequests := 0
	sendRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/token":
			tokenRequests++
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"access_token": "fresh-token",
				"expires_in":   7200,
			})
		case "/send":
			sendRequests++
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"errcode": wechatExpiredTokenCode,
				"errmsg":  "token expired",
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	sender := &HTTPWeChatSubscribeSender{
		Client:         server.Client(),
		AppID:          "test-app-id",
		AppSecret:      "test-app-secret",
		TokenEndpoint:  server.URL + "/token",
		SendEndpoint:   server.URL + "/send",
		accessToken:    "stale-token",
		tokenExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	err := sender.Send(context.Background(), WeChatSubscribeRequest{
		OpenID: "openid", TemplateID: "template-id", Page: "pages/meal/history",
		Data: map[string]map[string]string{"thing1": {"value": "测试饭局"}},
	})
	var deliveryError WeChatDeliveryError
	if !errors.As(err, &deliveryError) || deliveryError.Code != strconv.Itoa(wechatExpiredTokenCode) {
		t.Fatalf("send error = %#v, want delivery error %d", err, wechatExpiredTokenCode)
	}
	mu.Lock()
	defer mu.Unlock()
	if tokenRequests != 1 || sendRequests != 2 {
		t.Fatalf(
			"token requests = %d, send requests = %d, want 1 and 2",
			tokenRequests,
			sendRequests,
		)
	}
}

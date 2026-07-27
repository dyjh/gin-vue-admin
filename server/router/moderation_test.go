package router

import (
	"sort"
	"testing"

	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

// TestModerationRoutesMatchProxyStrippedContract 验证图片审核配置仅暴露读取、直接保存和测试路由。
func TestModerationRoutesMatchProxyStrippedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	new(ModerationRouter).InitModerationRouter(
		engine.Group(""),
		orderfoodApi.NewModerationApi(nil),
	)

	got := make([]string, 0, len(engine.Routes()))
	for _, route := range engine.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	want := []string{
		"GET /orderfood/moderation-config",
		"POST /orderfood/moderation-config/connection-tests",
		"POST /orderfood/moderation-config/moderation-tests",
		"PUT /orderfood/moderation-config",
	}
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("route count = %d, want %d: %v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("routes = %v, want %v", got, want)
		}
	}
}

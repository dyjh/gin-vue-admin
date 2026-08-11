package engagement

import (
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestPointsRoutesMatchProxyStrippedContract 验证积分路由与直接生效契约一致。
func TestPointsRoutesMatchProxyStrippedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	new(PointsRouter).InitPointsRouter(engine.Group("/orderfood"))

	got := make([]string, 0, len(engine.Routes()))
	for _, route := range engine.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	want := []string{
		"GET /orderfood/point-entries",
		"GET /orderfood/point-rules",
		"POST /orderfood/point-adjustments",
		"POST /orderfood/point-adjustments/preview",
		"PUT /orderfood/point-rules",
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

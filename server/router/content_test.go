package router

import (
	"sort"
	"testing"

	orderfoodApi "github.com/dyjh/order-food-mini-app/server/api/v1"
	"github.com/gin-gonic/gin"
)

func TestContentRoutesMatchProxyStrippedContractBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterContentRoutes(engine.Group(""), orderfoodApi.NewContentApi(nil))

	got := make([]string, 0)
	for _, route := range engine.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	want := []string{
		"GET /orderfood/user-dishes",
		"GET /orderfood/user-dishes/:dishId",
		"GET /orderfood/user-dishes/:dishId/references",
		"GET /orderfood/user-recipes",
		"GET /orderfood/user-recipes/:recipeId",
		"POST /orderfood/governance-actions",
		"POST /orderfood/governance-actions/preview",
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

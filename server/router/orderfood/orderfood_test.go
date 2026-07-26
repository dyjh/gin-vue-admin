package orderfood

import (
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOrderFoodRoutesMatchProxyStrippedContractBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	privateGroup := engine.Group("")
	new(UserRouter).InitUserRouter(privateGroup)
	new(AuditRouter).InitAuditRouter(privateGroup)
	new(ContentRouter).InitContentRouter(privateGroup)

	got := make([]string, 0)
	for _, route := range engine.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	want := []string{
		"GET /orderfood/audit-logs",
		"GET /orderfood/audit-logs/:auditLogId",
		"GET /orderfood/user-dishes",
		"GET /orderfood/user-dishes/:dishId",
		"GET /orderfood/user-dishes/:dishId/references",
		"GET /orderfood/user-recipes",
		"GET /orderfood/user-recipes/:recipeId",
		"GET /orderfood/users",
		"GET /orderfood/users/:userId",
		"GET /orderfood/users/:userId/preference-profile",
		"POST /orderfood/governance-actions",
		"POST /orderfood/governance-actions/preview",
		"PUT /orderfood/users/:userId/status",
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

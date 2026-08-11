package router

import (
	"sort"
	"testing"

	auditRouter "github.com/dyjh/order-food-mini-app/server/router/audit"
	contentRouter "github.com/dyjh/order-food-mini-app/server/router/content"
	userRouter "github.com/dyjh/order-food-mini-app/server/router/user"
	"github.com/gin-gonic/gin"
)

func TestOrderFoodRoutesMatchProxyStrippedContractBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	orderFoodGroup := engine.Group("/orderfood")
	new(userRouter.UserRouter).InitUserRouter(orderFoodGroup)
	new(auditRouter.AuditRouter).InitAuditRouter(orderFoodGroup)
	new(contentRouter.ContentRouter).InitContentRouter(orderFoodGroup)

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
		"PUT /orderfood/users/:userId/capability",
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

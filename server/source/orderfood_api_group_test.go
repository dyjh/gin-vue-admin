package source

import "testing"

func TestOrderFoodAdminAPIGroupMatchesFeatureDomains(t *testing.T) {
	tests := map[string]string{
		"/orderfood/dashboard":                               orderFoodAPIGroupDashboard,
		"/orderfood/users/:userId":                           orderFoodAPIGroupUsers,
		"/orderfood/point-adjustments/preview":               orderFoodAPIGroupUsers,
		"/orderfood/user-dishes/:dishId":                     orderFoodAPIGroupDishes,
		"/orderfood/suggestion-catalog/ingredients":          orderFoodAPIGroupDishes,
		"/orderfood/official-dishes/:dishId":                 orderFoodAPIGroupDishes,
		"/orderfood/moderation-config/connection-tests":      orderFoodAPIGroupSafety,
		"/orderfood/governance-jobs/:jobId":                  orderFoodAPIGroupSafety,
		"/orderfood/shopping-lists/:shoppingListId":          orderFoodAPIGroupMeals,
		"/orderfood/ai-providers/:providerId/models":         orderFoodAPIGroupAI,
		"/orderfood/ai-capabilities/:capabilityCode/prompts": orderFoodAPIGroupAI,
		"/orderfood/subscribe-scenes/:scene/status":          orderFoodAPIGroupMessages,
		"/orderfood/wechat-config":                           orderFoodAPIGroupWeChat,
	}

	for path, want := range tests {
		if got := orderFoodAdminAPIGroup(path); got != want {
			t.Fatalf("orderFoodAdminAPIGroup(%q) = %q, want %q", path, got, want)
		}
	}
	if got := orderFoodAdminAPIGroup("/orderfood/unmapped-resource"); got != "" {
		t.Fatalf("unmapped API group = %q, want empty", got)
	}
}

func TestEveryOrderFoodAdminRouteHasAFeatureGroup(t *testing.T) {
	routes := orderFoodAdminRouteSeeds("/orderfood")
	if len(routes) != 114 {
		t.Fatalf("route seed count = %d, want 114", len(routes))
	}

	groups := make(map[string]int)
	for _, route := range routes {
		group := orderFoodAdminAPIGroup(route.Path)
		if group == "" {
			t.Fatalf("route %s %s has no API group", route.Method, route.Path)
		}
		groups[group]++
	}
	if len(groups) != 8 {
		t.Fatalf("API group count = %d, want 8: %#v", len(groups), groups)
	}
}

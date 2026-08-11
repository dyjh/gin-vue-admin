package meal

import (
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMealRoutesKeepPublicContract 验证资源拆分不改变饭局域的外部路由契约。
func TestMealRoutesKeepPublicContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	orderFoodGroup := engine.Group("/orderfood")
	new(MealRouter).InitMealRouter(orderFoodGroup)
	new(ShoppingListRouter).InitShoppingListRouter(orderFoodGroup)

	got := make([]string, 0, len(engine.Routes()))
	for _, route := range engine.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	want := []string{
		"GET /orderfood/meals",
		"GET /orderfood/meals/:mealId",
		"GET /orderfood/shopping-lists",
		"GET /orderfood/shopping-lists/:shoppingListId",
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

package meal

import "github.com/gin-gonic/gin"

// ShoppingListRouter 提供采购清单管理端路由注册能力。
type ShoppingListRouter struct{}

// InitShoppingListRouter 注册采购清单管理端只读路由。
func (router *ShoppingListRouter) InitShoppingListRouter(orderFoodGroup *gin.RouterGroup) {
	orderFoodGroup.GET("/shopping-lists", shoppingListApi.ListShoppingLists)
	orderFoodGroup.GET("/shopping-lists/:shoppingListId", shoppingListApi.GetShoppingList)
}

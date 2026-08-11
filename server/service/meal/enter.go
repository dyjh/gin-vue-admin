package meal

// ServiceGroup aggregates meal and shopping-list administration services.
type ServiceGroup struct {
	Meal         *MealService         // 饭局管理服务
	ShoppingList *ShoppingListService // 采购清单管理服务
}

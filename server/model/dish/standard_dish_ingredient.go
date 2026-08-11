package dish

import (
	"time"
)

// StandardDishIngredient 表示标准菜品与食材关联。
type StandardDishIngredient struct {
	DishID       string             `json:"dishId" gorm:"column:dish_id;type:varchar(64);primaryKey;not null;comment:菜品ID;"`             // 菜品ID
	IngredientID string             `json:"ingredientId" gorm:"column:ingredient_id;type:varchar(64);primaryKey;not null;comment:食材ID;"` // 食材ID
	Ingredient   StandardIngredient `json:"ingredient" gorm:"foreignKey:IngredientID"`                                                   // 食材关联数据
	Required     bool               `json:"required" gorm:"column:required;type:tinyint(1) unsigned;not null;comment:是否必需;"`             // 是否必需
	SortOrder    int                `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;comment:排序值;"`                 // 排序值
	CreatedAt    time.Time          `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                     // 创建时间
}

// TableName 指定StandardDishIngredient对应的数据表名。
func (StandardDishIngredient) TableName() string {
	return "of_std_dish_ingredients"
}

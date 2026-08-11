package content

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	"time"
)

// RecipeDish 表示菜谱与菜品关联。
type RecipeDish struct {
	global.GVA_MODEL           // GVA基础模型字段
	PublicID         string    `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                         // 对外公开ID
	RecipeID         uint      `json:"recipeId" gorm:"column:recipe_id;type:bigint unsigned;not null;uniqueIndex:idx_recipe_dish;index;comment:菜谱ID;"` // 菜谱ID
	DishID           uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;uniqueIndex:idx_recipe_dish;index;comment:菜品ID;"`     // 菜品ID
	Dish             UserDish  `json:"dish" gorm:"foreignKey:DishID"`                                                                                  // 菜品关联数据
	SortOrder        int       `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`                              // 排序值
	AddedAt          time.Time `json:"addedAt" gorm:"column:added_at;type:datetime;not null;comment:加入菜谱时间;"`                                          // 加入菜谱时间
}

// TableName 指定RecipeDish对应的数据表名。
func (RecipeDish) TableName() string { return "of_recipe_dishes" }

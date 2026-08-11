package content

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
)

// DishIngredient 表示菜品食材。
type DishIngredient struct {
	global.GVA_MODEL                        // GVA基础模型字段
	DishID           uint                   `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`     // 菜品ID
	Name             string                 `json:"name" gorm:"column:name;type:varchar(120);not null;comment:名称;"`                     // 名称
	Quantity         *string                `json:"quantity" gorm:"column:quantity;type:varchar(40);default:null;comment:用量;"`          // 用量
	UnitID           *uint                  `json:"unitId" gorm:"column:unit_id;type:bigint unsigned;index;default:null;comment:单位ID;"` // 单位ID
	Unit             *dishModel.ContentUnit `json:"unit" gorm:"foreignKey:UnitID"`                                                      // 单位关联数据
	Note             *string                `json:"note" gorm:"column:note;type:varchar(300);default:null;comment:备注;"`                 // 备注
	SortOrder        int                    `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`  // 排序值
}

// TableName 指定DishIngredient对应的数据表名。
func (DishIngredient) TableName() string { return "of_dish_ingredients" }

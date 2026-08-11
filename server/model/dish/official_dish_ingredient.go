package dish

import (
	"github.com/dyjh/order-food-mini-app/server/global"
)

// OfficialDishIngredient 表示官方菜品食材。
type OfficialDishIngredient struct {
	global.GVA_MODEL              // GVA基础模型字段
	DishID           uint         `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:官方菜品ID;"`   // 官方菜品ID
	Name             string       `json:"name" gorm:"column:name;type:varchar(120);not null;comment:名称;"`                     // 名称
	Quantity         *string      `json:"quantity" gorm:"column:quantity;type:varchar(40);default:null;comment:用量;"`          // 用量
	UnitID           *uint        `json:"unitId" gorm:"column:unit_id;type:bigint unsigned;index;default:null;comment:单位ID;"` // 单位ID
	Unit             *ContentUnit `json:"unit" gorm:"foreignKey:UnitID"`                                                      // 单位关联数据
	Note             *string      `json:"note" gorm:"column:note;type:varchar(300);default:null;comment:备注;"`                 // 备注
	SortOrder        int          `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`  // 排序值
}

// TableName 指定OfficialDishIngredient对应的数据表名。
func (OfficialDishIngredient) TableName() string { return "of_off_ingredients" }

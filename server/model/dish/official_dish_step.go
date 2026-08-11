package dish

import (
	"github.com/dyjh/order-food-mini-app/server/global"
)

// OfficialDishStep 表示官方菜品步骤。
type OfficialDishStep struct {
	global.GVA_MODEL         // GVA基础模型字段
	DishID           uint    `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:官方菜品ID;"`  // 官方菜品ID
	Description      string  `json:"description" gorm:"column:description;type:varchar(2000);not null;comment:说明;"`     // 说明
	ImageURL         *string `json:"imageUrl" gorm:"column:image_url;type:varchar(512);default:null;comment:图片地址;"`     // 图片地址
	SortOrder        int     `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"` // 排序值
}

// TableName 指定OfficialDishStep对应的数据表名。
func (OfficialDishStep) TableName() string { return "of_off_steps" }

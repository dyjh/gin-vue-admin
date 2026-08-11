package meal

import (
	"gorm.io/datatypes"
	"time"
)

// FrontShoppingItem 表示采购清单项。
type FrontShoppingItem struct {
	ID              string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                       // 主键ID
	ListID          string         `json:"listId" gorm:"column:list_id;type:varchar(64);not null;index;comment:采购清单ID;"`                 // 采购清单ID
	Name            string         `json:"name" gorm:"column:name;type:varchar(40);not null;comment:名称;"`                                // 名称
	Amount          string         `json:"amount" gorm:"column:amount;type:varchar(30);not null;comment:积分或用量变动值;"`                      // 积分或用量变动值
	Note            *string        `json:"note" gorm:"column:note;type:varchar(100);default:null;comment:备注;"`                           // 备注
	Completed       bool           `json:"completed" gorm:"column:completed;type:tinyint(1) unsigned;not null;index;comment:是否已完成;"`     // 是否已完成
	SourceDishNames datatypes.JSON `json:"sourceDishNames" gorm:"column:source_dish_names;type:json;default:null;comment:来源菜品名称集合JSON;"` // 来源菜品名称集合JSON
	SortOrder       int            `json:"sortOrder" gorm:"column:sort_order;type:int;not null;index;comment:排序值;"`                      // 排序值
	CreatedAt       time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                      // 创建时间
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                      // 更新时间
}

// TableName 指定FrontShoppingItem对应的数据表名。
func (FrontShoppingItem) TableName() string { return "of_shopping_items" }

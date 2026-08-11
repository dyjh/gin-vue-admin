package dish

import (
	"time"
)

// OfficialDishTag 表示官方菜品与标签关联。
type OfficialDishTag struct {
	DishID    uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;primaryKey;not null;comment:官方菜品ID;"` // 官方菜品ID
	TagID     uint      `json:"tagId" gorm:"column:tag_id;type:bigint unsigned;primaryKey;not null;comment:标签ID;"`     // 标签ID
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`               // 创建时间
}

// TableName 指定OfficialDishTag对应的数据表名。
func (OfficialDishTag) TableName() string { return "of_off_tags" }

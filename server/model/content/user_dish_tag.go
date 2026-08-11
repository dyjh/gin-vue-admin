package content

import (
	"time"
)

// UserDishTag 表示菜品与标签关联。
type UserDishTag struct {
	DishID    uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;primaryKey;not null;comment:菜品ID;"` // 菜品ID
	TagID     uint      `json:"tagId" gorm:"column:tag_id;type:bigint unsigned;primaryKey;not null;comment:标签ID;"`   // 标签ID
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`             // 创建时间
}

// TableName 指定UserDishTag对应的数据表名。
func (UserDishTag) TableName() string { return "of_dish_tags" }

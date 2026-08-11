package content

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	"time"
)

const (
	ReferenceTypeRecipe                = "recipe"
	ReferenceTypeUnconfirmedMeal       = "unconfirmed_meal"
	ReferenceTypeConfirmedMealSnapshot = "confirmed_meal_snapshot"
	ReferenceTypeShoppingSnapshot      = "shopping_snapshot"
)

// DishReference 表示菜品业务引用。
type DishReference struct {
	global.GVA_MODEL             // GVA基础模型字段
	PublicID           string    `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                                // 对外公开ID
	DishID             uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`                                        // 菜品ID
	ReferenceType      string    `json:"referenceType" gorm:"column:reference_type;type:varchar(40);not null;index;comment:引用类型;"`                              // 引用类型
	ObjectID           string    `json:"objectId" gorm:"column:object_id;type:varchar(64);not null;index;comment:对象ID;"`                                        // 对象ID
	ObjectLabel        string    `json:"objectLabel" gorm:"column:object_label;type:varchar(160);not null;comment:对象名称;"`                                       // 对象名称
	ObjectStatus       string    `json:"objectStatus" gorm:"column:object_status;type:varchar(40);not null;comment:对象状态;"`                                      // 对象状态
	HistoricalSnapshot bool      `json:"historicalSnapshot" gorm:"column:historical_snapshot;type:tinyint(1) unsigned;not null;default:false;comment:是否为历史快照;"` // 是否为历史快照
	OccurredAt         time.Time `json:"occurredAt" gorm:"column:occurred_at;type:datetime;not null;index;comment:发生时间;"`                                       // 发生时间
}

// TableName 指定DishReference对应的数据表名。
func (DishReference) TableName() string { return "of_dish_refs" }

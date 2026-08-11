package meal

import (
	"time"
)

// FrontMealCandidate 表示饭局候选菜。
type FrontMealCandidate struct {
	ID                string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                     // 主键ID
	MealID            string    `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_candidate,priority:1;index;comment:饭局ID;"`     // 饭局ID
	DishID            uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;uniqueIndex:uk_of_meal_candidate,priority:2;index;comment:菜品ID;"` // 菜品ID
	Available         bool      `json:"available" gorm:"column:available;type:tinyint(1) unsigned;not null;default:true;index;comment:候选菜是否可选;"`                    // 候选菜是否可选
	UnavailableReason *string   `json:"unavailableReason" gorm:"column:unavailable_reason;type:varchar(24);default:null;comment:不可选原因;"`                            // 不可选原因
	SortOrder         int       `json:"sortOrder" gorm:"column:sort_order;type:int;not null;comment:排序值;"`                                                          // 排序值
	CreatedAt         time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                    // 创建时间
}

// TableName 指定FrontMealCandidate对应的数据表名。
func (FrontMealCandidate) TableName() string { return "of_meal_candidates" }

package meal

import (
	"gorm.io/datatypes"
	"time"
)

// FrontMealFinalDish 表示饭局最终菜品快照。
type FrontMealFinalDish struct {
	ID            string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                  // 主键ID
	MealID        string         `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_final,priority:1;index;comment:饭局ID;"`      // 饭局ID
	CandidateID   string         `json:"candidateId" gorm:"column:candidate_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_final,priority:2;comment:候选菜ID;"` // 候选菜ID
	DishID        uint           `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`                                          // 菜品ID
	Name          string         `json:"name" gorm:"column:name;type:varchar(120);not null;comment:名称;"`                                                          // 名称
	CoverURL      *string        `json:"coverUrl" gorm:"column:cover_url;type:varchar(512);default:null;comment:封面地址;"`                                           // 封面地址
	FinalServings int            `json:"finalServings" gorm:"column:final_servings;type:int;not null;comment:最终份数;"`                                              // 最终份数
	Ingredients   datatypes.JSON `json:"-" gorm:"column:ingredients;type:json;not null;comment:食材列表;"`                                                            // 食材列表
	Steps         datatypes.JSON `json:"-" gorm:"column:steps_json;type:json;default:null;comment:步骤列表;"`                                                         // 步骤列表
	CreatedAt     time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                 // 创建时间
}

// TableName 指定FrontMealFinalDish对应的数据表名。
func (FrontMealFinalDish) TableName() string { return "of_meal_final_dishes" }

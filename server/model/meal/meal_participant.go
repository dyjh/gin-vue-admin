package meal

import (
	"time"
)

// MealParticipant 表示饭局参与人。
type MealParticipant struct {
	ID       string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                   // 主键ID
	MealID   string    `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_participant,priority:1;index;comment:饭局ID;"` // 饭局ID
	UserID   string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_participant,priority:2;index;comment:用户ID;"` // 用户ID
	JoinedAt time.Time `json:"joinedAt" gorm:"column:joined_at;type:datetime;not null;comment:加入时间;"`                                                    // 加入时间
}

// TableName 指定MealParticipant对应的数据表名。
func (MealParticipant) TableName() string { return "of_meal_members" }

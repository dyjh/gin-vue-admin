package meal

import (
	"time"
)

// FrontMealVote 表示饭局点选记录。
type FrontMealVote struct {
	ID          string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                       // 主键ID
	MealID      string    `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_vote,priority:1;index;comment:饭局ID;"`            // 饭局ID
	UserID      string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_vote,priority:2;index;comment:用户ID;"`            // 用户ID
	CandidateID string    `json:"candidateId" gorm:"column:candidate_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_vote,priority:3;index;comment:候选菜ID;"` // 候选菜ID
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                      // 创建时间
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                      // 更新时间
}

// TableName 指定FrontMealVote对应的数据表名。
func (FrontMealVote) TableName() string { return "of_meal_votes" }

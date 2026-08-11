package dish

import (
	"time"
)

// RecommendationCopy 表示推荐菜复制记录。
type RecommendationCopy struct {
	ID               string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                                     // 主键ID
	RecommendationID string    `json:"recommendationId" gorm:"column:recommendation_id;type:varchar(64);not null;uniqueIndex:uk_of_recommendation_copy,priority:1;comment:推荐位ID;"` // 推荐位ID
	UserID           string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_recommendation_copy,priority:2;comment:用户ID;"`                      // 用户ID
	CopiedDishID     uint      `json:"copiedDishId" gorm:"column:copied_dish_id;type:bigint unsigned;not null;index;comment:复制生成的菜品ID;"`                                           // 复制生成的菜品ID
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                    // 创建时间
}

// TableName 指定RecommendationCopy对应的数据表名。
func (RecommendationCopy) TableName() string {
	return "of_recommendation_copies"
}

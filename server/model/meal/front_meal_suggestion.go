package meal

import (
	"gorm.io/datatypes"
	"time"
)

// FrontMealSuggestion 表示AI饭局推荐结果。
type FrontMealSuggestion struct {
	ID          string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`           // 主键ID
	UserID      string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`       // 用户ID
	Source      string         `json:"source" gorm:"column:source;type:varchar(20);not null;comment:来源;"`                // 来源
	SourceLabel string         `json:"sourceLabel" gorm:"column:source_label;type:varchar(80);not null;comment:来源说明;"`   // 来源说明
	Reason      string         `json:"reason" gorm:"column:reason;type:varchar(300);not null;comment:原因;"`               // 原因
	People      int            `json:"people" gorm:"column:people;type:int;not null;comment:用餐人数;"`                      // 用餐人数
	DishIDsJSON datatypes.JSON `json:"-" gorm:"column:dish_i_ds_json;type:json;not null;comment:菜品ID集合JSON;"`            // 菜品ID集合JSON
	UsageID     string         `json:"usageId" gorm:"column:usage_id;type:varchar(64);not null;index;comment:AI调用记录ID;"` // AI调用记录ID
	Feedback    *string        `json:"feedback" gorm:"column:feedback;type:varchar(20);default:null;comment:用户反馈;"`      // 用户反馈
	CreatedAt   time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`    // 创建时间
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`          // 更新时间
}

// TableName 指定FrontMealSuggestion对应的数据表名。
func (FrontMealSuggestion) TableName() string { return "of_meal_suggestions" }

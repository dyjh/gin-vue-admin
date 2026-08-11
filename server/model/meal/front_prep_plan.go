package meal

import (
	"gorm.io/datatypes"
	"time"
)

// FrontPrepPlan 表示AI备菜计划。
type FrontPrepPlan struct {
	ID               string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`               // 主键ID
	UserID           string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`           // 用户ID
	MealID           string         `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;index;comment:饭局ID;"`           // 饭局ID
	EstimatedMinutes int            `json:"estimatedMinutes" gorm:"column:estimated_minutes;type:int;not null;comment:预计耗时（分钟）;"` // 预计耗时（分钟）
	StepsJSON        datatypes.JSON `json:"-" gorm:"column:steps_json;type:json;not null;comment:步骤JSON;"`                        // 步骤JSON
	UsageID          string         `json:"usageId" gorm:"column:usage_id;type:varchar(64);not null;index;comment:AI调用记录ID;"`     // AI调用记录ID
	Feedback         *string        `json:"feedback" gorm:"column:feedback;type:varchar(20);default:null;comment:用户反馈;"`          // 用户反馈
	CreatedAt        time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`        // 创建时间
	UpdatedAt        time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`              // 更新时间
}

// TableName 指定FrontPrepPlan对应的数据表名。
func (FrontPrepPlan) TableName() string { return "of_prep_plans" }

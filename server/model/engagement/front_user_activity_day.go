package engagement

import (
	"time"
)

// FrontUserActivityDay 表示用户在一个上海自然日内至少完成过一次成功业务请求。
type FrontUserActivityDay struct {
	ID         uint      `json:"id" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                                       // 主键ID
	UserID     string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_user_activity,priority:1;index;comment:用户ID;"`          // 用户ID
	ActiveDate string    `json:"activeDate" gorm:"column:active_date;type:char(10);not null;uniqueIndex:uk_of_user_activity,priority:2;index;comment:上海时区活跃日期;"` // 上海时区活跃日期
	FirstAt    time.Time `json:"firstAt" gorm:"column:first_at;type:datetime;not null;comment:当日首次成功业务请求时间;"`                                                    // 当日首次成功业务请求时间
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                        // 创建时间
}

// TableName 指定FrontUserActivityDay对应的数据表名。
func (FrontUserActivityDay) TableName() string { return "of_user_activity" }

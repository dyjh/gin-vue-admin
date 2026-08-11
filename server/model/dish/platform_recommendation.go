package dish

import (
	"time"
)

// PlatformRecommendation 表示平台推荐菜。
type PlatformRecommendation struct {
	ID                string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                          // 主键ID
	DishID            *uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;index;default:null;comment:用户菜品内部ID;"`                          // 用户菜品内部ID
	OfficialDishID    *uint      `json:"officialDishId" gorm:"column:official_dish_id;type:bigint unsigned;index;default:null;comment:官方菜品内部ID;"`         // 官方菜品内部ID
	SourceType        string     `json:"sourceType" gorm:"column:source_type;type:varchar(24);not null;index;comment:来源类型;"`                              // 来源类型
	SourceDishID      string     `json:"sourceDishId" gorm:"column:source_dish_id;type:varchar(64);not null;default:'';index;comment:来源菜品公开ID;"`          // 来源菜品公开ID
	Position          string     `json:"position" gorm:"column:position;type:varchar(32);not null;default:home_featured;index;comment:推荐位置;"`             // 推荐位置
	DisplayNote       *string    `json:"displayNote" gorm:"column:display_note;type:varchar(160);default:null;comment:展示说明;"`                             // 展示说明
	Status            string     `json:"status" gorm:"column:status;type:varchar(24);not null;default:draft;index;comment:推荐状态;"`                         // 推荐状态
	Selected          bool       `json:"selected" gorm:"column:selected;type:tinyint(1) unsigned;not null;default:false;index;comment:小程序是否可见;"`          // 小程序是否可见
	SortOrder         int        `json:"sortOrder" gorm:"column:sort_order;type:int;not null;index;comment:排序值;"`                                         // 排序值
	Version           int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                      // 数据版本
	PublishedAt       *time.Time `json:"publishedAt" gorm:"column:published_at;type:datetime;index;default:null;comment:发布时间;"`                           // 发布时间
	OfflineAt         *time.Time `json:"offlineAt" gorm:"column:offline_at;type:datetime;index;default:null;comment:下线时间;"`                               // 下线时间
	OfflineReason     *string    `json:"offlineReason" gorm:"column:offline_reason;type:varchar(200);default:null;comment:下线原因;"`                         // 下线原因
	CreatedByID       uint       `json:"createdById" gorm:"column:created_by_id;type:bigint unsigned;not null;default:0;index;comment:创建管理员ID;"`          // 创建管理员ID
	CreatedByUsername string     `json:"createdByUsername" gorm:"column:created_by_username;type:varchar(120);not null;default:system;comment:创建管理员用户名;"` // 创建管理员用户名
	CreatedByNickname *string    `json:"createdByNickname" gorm:"column:created_by_nickname;type:varchar(120);default:null;comment:创建管理员昵称;"`             // 创建管理员昵称
	UpdatedByID       uint       `json:"updatedById" gorm:"column:updated_by_id;type:bigint unsigned;not null;default:0;index;comment:更新管理员ID;"`          // 更新管理员ID
	UpdatedByUsername string     `json:"updatedByUsername" gorm:"column:updated_by_username;type:varchar(120);not null;default:system;comment:更新管理员用户名;"` // 更新管理员用户名
	UpdatedByNickname *string    `json:"updatedByNickname" gorm:"column:updated_by_nickname;type:varchar(120);default:null;comment:更新管理员昵称;"`             // 更新管理员昵称
	CreatedAt         time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                         // 创建时间
	UpdatedAt         time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                   // 更新时间
}

// TableName 指定PlatformRecommendation对应的数据表名。
func (PlatformRecommendation) TableName() string {
	return "of_recommendations"
}

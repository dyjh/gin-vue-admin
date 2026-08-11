package engagement

import (
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"time"
)

// UserNotification 表示用户站内通知。
type UserNotification struct {
	ID                string                `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                               // 主键ID
	UserID            string                `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                           // 用户ID
	User              userModel.MiniAppUser `json:"user" gorm:"foreignKey:UserID;references:ID"`                                                                          // 用户关联数据
	Type              string                `json:"type" gorm:"column:type;type:varchar(32);not null;index;comment:类型;"`                                                  // 类型
	Title             string                `json:"title" gorm:"column:title;type:varchar(120);not null;comment:标题;"`                                                     // 标题
	Content           string                `json:"content" gorm:"column:content;type:varchar(1000);not null;comment:内容;"`                                                // 内容
	TargetType        *string               `json:"targetType" gorm:"column:target_type;type:varchar(40);index;default:null;comment:目标类型;"`                               // 目标类型
	TargetID          *string               `json:"targetId" gorm:"column:target_id;type:varchar(64);index;default:null;comment:目标ID;"`                                   // 目标ID
	SubscribeRequired bool                  `json:"subscribeRequired" gorm:"column:subscribe_required;type:tinyint(1) unsigned;not null;default:false;comment:是否需要订阅消息;"` // 是否需要订阅消息
	SubscribeLogID    *string               `json:"subscribeLogId" gorm:"column:subscribe_log_id;type:varchar(64);index;default:null;comment:订阅消息日志ID;"`                  // 订阅消息日志ID
	ReadAt            *time.Time            `json:"readAt" gorm:"column:read_at;type:datetime;index;default:null;comment:阅读时间;"`                                          // 阅读时间
	CreatedAt         time.Time             `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                        // 创建时间
}

// TableName 指定UserNotification对应的数据表名。
func (UserNotification) TableName() string { return "of_notifications" }

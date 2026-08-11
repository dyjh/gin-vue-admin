package engagement

import (
	"time"
)

const (
	// MealAuthorizationAccept 表示用户接受本次微信订阅授权。
	MealAuthorizationAccept = "accept"
	// MealAuthorizationReject 表示用户拒绝本次微信订阅授权。
	MealAuthorizationReject = "reject"
	// MealAuthorizationBan 表示用户在微信侧永久拒绝本订阅模板。
	MealAuthorizationBan = "ban"
)

// MealFinalResultSubscription 表示用户对单场饭局最终结果的一次微信订阅授权。
type MealFinalResultSubscription struct {
	ID                  string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                     // 主键ID
	MealID              string     `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;index:idx_of_meal_sub_user,priority:1;comment:饭局ID;"` // 饭局ID
	UserID              string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index:idx_of_meal_sub_user,priority:2;comment:用户ID;"` // 用户ID
	TemplateID          string     `json:"templateId" gorm:"column:template_id;type:varchar(64);not null;index;comment:订阅模板配置ID;"`                     // 订阅模板配置ID
	AuthorizationResult string     `json:"authorizationResult" gorm:"column:authorization_result;type:varchar(8);not null;index;comment:微信授权结果;"`      // 微信授权结果
	Accepted            bool       `json:"accepted" gorm:"column:accepted;type:tinyint(1) unsigned;not null;default:false;index;comment:是否获得一次可用授权;"`  // 是否获得一次可用授权
	ConsumedResult      *string    `json:"consumedResult" gorm:"column:consumed_result;type:varchar(16);default:null;comment:消费该授权的最终结果;"`             // 消费该授权的最终结果
	ConsumedAt          *time.Time `json:"consumedAt" gorm:"column:consumed_at;type:datetime;index;default:null;comment:授权消费时间;"`                      // 授权消费时间
	RecordedAt          time.Time  `json:"recordedAt" gorm:"column:recorded_at;type:datetime;not null;index;comment:授权登记时间;"`                          // 授权登记时间
}

// TableName 指定MealFinalResultSubscription对应的数据表名。
func (MealFinalResultSubscription) TableName() string { return "of_meal_subs" }

package user

import (
	"time"
)

// UserStatus 表示小程序用户账号状态。
type UserStatus string

const (
	UserStatusNormal   UserStatus = "normal"
	UserStatusDisabled UserStatus = "disabled"
)

// MiniAppUser keeps WeChat credentials as storage-only fields. API response
// types live in model/orderfood/response and never include OpenID or
// SessionKeyEncrypted.
type MiniAppUser struct {
	ID                  string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                         // 主键ID
	OpenIDHash          string     `json:"-" gorm:"column:open_id_hash;type:varchar(64);not null;uniqueIndex;comment:微信OpenID摘要;"`                                         // 微信OpenID摘要
	OpenIDEncrypted     string     `json:"-" gorm:"column:open_id_encrypted;type:text;not null;comment:加密后的微信OpenID;"`                                                     // 加密后的微信OpenID
	UnionIDHash         *string    `json:"-" gorm:"column:union_id_hash;type:varchar(64);index;default:null;comment:微信UnionID摘要;"`                                         // 微信UnionID摘要
	UnionIDEncrypted    *string    `json:"-" gorm:"column:union_id_encrypted;type:text;default:null;comment:加密后的微信UnionID;"`                                               // 加密后的微信UnionID
	SessionKeyEncrypted *string    `json:"-" gorm:"column:session_key_encrypted;type:text;default:null;comment:加密后的微信会话密钥;"`                                               // 加密后的微信会话密钥
	AvatarURL           *string    `json:"avatarUrl" gorm:"column:avatar_url;type:varchar(500);default:null;comment:头像地址;"`                                                // 头像地址
	Nickname            string     `json:"nickname" gorm:"column:nickname;type:varchar(30);not null;comment:用户昵称;"`                                                        // 用户昵称
	Points              int64      `json:"points" gorm:"column:points;type:bigint;not null;default:0;index;comment:当前积分;"`                                                 // 当前积分
	CheckinDayCount     int        `json:"checkinDayCount" gorm:"column:checkin_day_count;type:int;not null;default:0;comment:累计打卡天数;"`                                    // 累计打卡天数
	DishCount           int        `json:"dishCount" gorm:"column:dish_count;type:int;not null;default:0;comment:菜品数量;"`                                                   // 菜品数量
	RecipeCount         int        `json:"recipeCount" gorm:"column:recipe_count;type:int;not null;default:0;comment:菜谱数量;"`                                               // 菜谱数量
	MealCount           int        `json:"mealCount" gorm:"column:meal_count;type:int;not null;default:0;comment:饭局数量;"`                                                   // 饭局数量
	CheckinCount        int        `json:"checkinCount" gorm:"column:checkin_count;type:int;not null;default:0;comment:打卡次数;"`                                             // 打卡次数
	CapabilityDisabled  bool       `json:"capabilityDisabled" gorm:"column:capability_disabled;type:tinyint(1) unsigned;not null;default:false;index;comment:是否单独关闭AI能力;"` // 是否单独关闭AI能力
	Status              UserStatus `json:"status" gorm:"column:status;type:varchar(16);not null;default:normal;index;comment:状态;"`                                         // 状态
	DisabledReason      *string    `json:"disabledReason" gorm:"column:disabled_reason;type:varchar(200);default:null;comment:禁用原因;"`                                      // 禁用原因
	DisabledAt          *time.Time `json:"disabledAt" gorm:"column:disabled_at;type:datetime;default:null;comment:禁用时间;"`                                                  // 禁用时间
	Version             int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                     // 数据版本
	RegisteredAt        time.Time  `json:"registeredAt" gorm:"column:registered_at;type:datetime;not null;index;comment:注册时间;"`                                            // 注册时间
	LastLoginAt         *time.Time `json:"lastLoginAt" gorm:"column:last_login_at;type:datetime;index;default:null;comment:最近登录时间;"`                                       // 最近登录时间
	CreatedAt           time.Time  `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                // 创建时间
	UpdatedAt           time.Time  `json:"-" gorm:"column:updated_at;type:datetime;index;not null;comment:更新时间;"`                                                          // 更新时间
}

// TableName 指定MiniAppUser对应的数据表名。
func (MiniAppUser) TableName() string {
	return "of_users"
}

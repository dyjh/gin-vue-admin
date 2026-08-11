package common

import (
	"time"
)

// FrontIdempotencyRecord 表示小程序幂等请求记录。
type FrontIdempotencyRecord struct {
	ID             uint      `json:"-" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                    // 主键ID
	UserID         string    `json:"-" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_idempotency,priority:1;comment:用户ID;"`   // 用户ID
	Endpoint       string    `json:"-" gorm:"column:endpoint;type:varchar(180);not null;uniqueIndex:uk_of_idempotency,priority:2;comment:服务端点;"` // 服务端点
	Key            string    `json:"-" gorm:"column:key;type:varchar(128);not null;uniqueIndex:uk_of_idempotency,priority:3;comment:幂等键;"`       // 幂等键
	RequestHash    string    `json:"-" gorm:"column:request_hash;type:varchar(64);not null;comment:请求内容摘要;"`                                     // 请求内容摘要
	State          string    `json:"-" gorm:"column:state;type:varchar(20);not null;index;comment:处理状态;"`                                        // 处理状态
	ResponseStatus int       `json:"-" gorm:"column:response_status;type:int;not null;default:0;comment:响应状态码;"`                                 // 响应状态码
	ResponseBody   []byte    `json:"-" gorm:"column:response_body;type:blob;default:null;comment:响应内容;"`                                         // 响应内容
	CreatedAt      time.Time `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                            // 创建时间
	UpdatedAt      time.Time `json:"-" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                            // 更新时间
	ExpiresAt      time.Time `json:"-" gorm:"column:expires_at;type:datetime;not null;index;comment:过期时间;"`                                      // 过期时间
}

// TableName 指定FrontIdempotencyRecord对应的数据表名。
func (FrontIdempotencyRecord) TableName() string {
	return "of_front_idem"
}

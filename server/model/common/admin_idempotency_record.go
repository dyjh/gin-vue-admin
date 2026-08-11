package common

import (
	"gorm.io/datatypes"
	"time"
)

// AdminIdempotencyState 表示管理端幂等请求的处理状态。
type AdminIdempotencyState string

const (
	AdminIdempotencyProcessing AdminIdempotencyState = "processing"
	AdminIdempotencyCompleted  AdminIdempotencyState = "completed"
)

// AdminIdempotencyRecord scopes a key to administrator + endpoint. The
// request hash detects key reuse with different path/body parameters and the
// stored response allows an exact successful replay.
type AdminIdempotencyRecord struct {
	ID              uint                  `json:"id" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                                                   // 主键ID
	AdministratorID uint                  `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;not null;uniqueIndex:uk_of_admin_idempotency,priority:1;comment:管理员ID;"` // 管理员ID
	EndpointID      string                `json:"endpointId" gorm:"column:endpoint_id;type:varchar(100);not null;uniqueIndex:uk_of_admin_idempotency,priority:2;comment:接口标识;"`               // 接口标识
	IdempotencyKey  string                `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(128);not null;uniqueIndex:uk_of_admin_idempotency,priority:3;comment:幂等键;"`        // 幂等键
	RequestHash     string                `json:"requestHash" gorm:"column:request_hash;type:varchar(64);not null;comment:请求内容摘要;"`                                                           // 请求内容摘要
	State           AdminIdempotencyState `json:"state" gorm:"column:state;type:varchar(20);not null;index;comment:处理状态;"`                                                                    // 处理状态
	ResponseCode    int                   `json:"responseCode" gorm:"column:response_code;type:int;not null;default:0;comment:响应业务码;"`                                                        // 响应业务码
	ResponseJSON    datatypes.JSON        `json:"responseJson" gorm:"column:response_json;type:json;default:null;comment:响应内容JSON;"`                                                          // 响应内容JSON
	CreatedAt       time.Time             `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                    // 创建时间
	UpdatedAt       time.Time             `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                                    // 更新时间
	ExpiresAt       time.Time             `json:"expiresAt" gorm:"column:expires_at;type:datetime;not null;index;comment:过期时间;"`                                                              // 过期时间
}

// TableName 指定AdminIdempotencyRecord对应的数据表名。
func (AdminIdempotencyRecord) TableName() string {
	return "of_admin_idem"
}

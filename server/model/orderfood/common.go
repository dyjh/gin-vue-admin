package orderfood

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	AuthorityPlatformSuperAdmin  uint = 888
	AuthorityOrderFoodSuperAdmin uint = 9901
	AuthorityOrderFoodOperator   uint = 9902
	AuthorityOrderFoodGovernance uint = 9903
	AuthorityOrderFoodSupport    uint = 9904
)

// NewID 生成业务公开ID。
func NewID() string {
	return uuid.NewString()
}

// AdminAccessAudit is deliberately separate from SysOperationRecord. Every
// sensitive read must synchronously persist one of these rows before data is
// returned to an administrator.
type AdminAccessAudit struct {
	ID               string    `json:"id" gorm:"column:id;type:varchar(36);primaryKey;not null;comment:主键ID;"`                                             // 主键ID
	AdministratorID  uint      `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`                  // 管理员ID
	AuthorityID      uint      `json:"authorityId" gorm:"column:authority_id;type:bigint unsigned;not null;index;comment:角色ID;"`                           // 角色ID
	Permission       string    `json:"permission" gorm:"column:permission;type:varchar(120);not null;index;comment:业务权限编码;"`                               // 业务权限编码
	Action           string    `json:"action" gorm:"column:action;type:varchar(80);not null;index;comment:操作类型;"`                                          // 操作类型
	TargetType       string    `json:"targetType" gorm:"column:target_type;type:varchar(60);not null;index:idx_of_access_target,priority:1;comment:目标类型;"` // 目标类型
	TargetID         string    `json:"targetId" gorm:"column:target_id;type:varchar(80);not null;index:idx_of_access_target,priority:2;comment:目标ID;"`     // 目标ID
	RequestID        string    `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                                  // 请求ID
	SourceIPMasked   *string   `json:"sourceIpMasked" gorm:"column:source_ip_masked;type:varchar(80);default:null;comment:脱敏来源IP;"`                        // 脱敏来源IP
	UserAgentSummary *string   `json:"userAgentSummary" gorm:"column:user_agent_summary;type:varchar(240);default:null;comment:客户端信息摘要;"`                  // 客户端信息摘要
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                      // 创建时间
}

// TableName 指定AdminAccessAudit对应的数据表名。
func (AdminAccessAudit) TableName() string {
	return "of_access_audits"
}

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

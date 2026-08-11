package audit

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	"gorm.io/datatypes"
)

// AdminAuditLog 表示管理端业务审计日志。
type AdminAuditLog struct {
	global.GVA_MODEL                     // GVA基础模型字段
	PublicID              string         `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                   // 对外公开ID
	AdministratorID       uint           `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`        // 管理员ID
	AdministratorUsername string         `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);not null;comment:管理员用户名;"`    // 管理员用户名
	AdministratorNickname *string        `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:管理员昵称;"` // 管理员昵称
	Action                string         `json:"action" gorm:"column:action;type:varchar(80);not null;index;comment:操作类型;"`                                // 操作类型
	TargetType            string         `json:"targetType" gorm:"column:target_type;type:varchar(32);not null;index;comment:目标类型;"`                       // 目标类型
	TargetID              string         `json:"targetId" gorm:"column:target_id;type:varchar(64);not null;index;comment:目标ID;"`                           // 目标ID
	TargetLabel           *string        `json:"targetLabel" gorm:"column:target_label;type:varchar(160);default:null;comment:目标名称;"`                      // 目标名称
	Reason                *string        `json:"reason" gorm:"column:reason;type:varchar(600);default:null;comment:原因;"`                                   // 原因
	BeforeSummary         datatypes.JSON `json:"beforeSummary" gorm:"column:before_summary;type:json;default:null;comment:变更前摘要JSON;"`                     // 变更前摘要JSON
	AfterSummary          datatypes.JSON `json:"afterSummary" gorm:"column:after_summary;type:json;default:null;comment:变更后摘要JSON;"`                       // 变更后摘要JSON
	RequestID             string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                        // 请求ID
	IdempotencyKey        *string        `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(160);index;default:null;comment:幂等键;"`           // 幂等键
	SourceIPMasked        *string        `json:"sourceIpMasked" gorm:"column:source_ip_masked;type:varchar(80);default:null;comment:脱敏来源IP;"`              // 脱敏来源IP
	UserAgentSummary      *string        `json:"userAgentSummary" gorm:"column:user_agent_summary;type:varchar(240);default:null;comment:客户端信息摘要;"`        // 客户端信息摘要
}

// TableName 指定AdminAuditLog对应的数据表名。
func (AdminAuditLog) TableName() string { return "of_admin_audits" }

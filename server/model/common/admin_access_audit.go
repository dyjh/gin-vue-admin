package common

import (
	"time"
)

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

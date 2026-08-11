package ai

import (
	"time"
)

// AIConfigurationChange 表示AI配置变更记录。
type AIConfigurationChange struct {
	ID                    string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                      // 主键ID
	ResourceType          string    `json:"resourceType" gorm:"column:resource_type;type:varchar(40);not null;index:idx_of_ai_change_resource,priority:1;comment:资源类型;"` // 资源类型
	ResourceID            string    `json:"resourceId" gorm:"column:resource_id;type:varchar(80);not null;index:idx_of_ai_change_resource,priority:2;comment:资源ID;"`     // 资源ID
	Action                string    `json:"action" gorm:"column:action;type:varchar(40);not null;comment:操作类型;"`                                                         // 操作类型
	ResourceVersion       int64     `json:"resourceVersion" gorm:"column:resource_version;type:bigint;not null;comment:资源版本;"`                                           // 资源版本
	AdministratorID       uint      `json:"-" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`                                         // 管理员ID
	AdministratorUsername string    `json:"-" gorm:"column:administrator_username;type:varchar(80);not null;comment:管理员用户名;"`                                            // 管理员用户名
	Reason                *string   `json:"reason" gorm:"column:reason;type:varchar(200);default:null;comment:原因;"`                                                      // 原因
	CreatedAt             time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                               // 创建时间
}

// TableName 指定AIConfigurationChange对应的数据表名。
func (AIConfigurationChange) TableName() string {
	return "of_ai_changes"
}

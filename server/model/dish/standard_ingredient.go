package dish

import (
	"gorm.io/datatypes"
	"time"
)

// StandardIngredient 表示标准食材索引。
type StandardIngredient struct {
	ID                string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`              // 主键ID
	Name              string         `json:"name" gorm:"column:name;type:varchar(80);not null;uniqueIndex;comment:名称;"`           // 名称
	AliasesJSON       datatypes.JSON `json:"-" gorm:"column:aliases_json;type:json;not null;comment:别名集合JSON;"`                   // 别名集合JSON
	Enabled           bool           `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;index;comment:是否启用;"` // 是否启用
	Version           int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`          // 数据版本
	UpdatedByID       uint           `json:"-" gorm:"column:updated_by_id;type:bigint unsigned;not null;comment:更新管理员ID;"`        // 更新管理员ID
	UpdatedByUsername string         `json:"-" gorm:"column:updated_by_username;type:varchar(80);not null;comment:更新管理员用户名;"`     // 更新管理员用户名
	UpdatedByNickname *string        `json:"-" gorm:"column:updated_by_nickname;type:varchar(80);default:null;comment:更新管理员昵称;"`  // 更新管理员昵称
	LastChangeReason  string         `json:"-" gorm:"column:last_change_reason;type:varchar(200);not null;comment:最近变更原因;"`       // 最近变更原因
	CreatedAt         time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`             // 创建时间
	UpdatedAt         time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`       // 更新时间
}

// TableName 指定StandardIngredient对应的数据表名。
func (StandardIngredient) TableName() string {
	return "of_std_ingredients"
}

package dish

import (
	"github.com/dyjh/order-food-mini-app/server/global"
)

// ContentUnit 表示食材单位。
type ContentUnit struct {
	global.GVA_MODEL        // GVA基础模型字段
	PublicID         string `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`     // 对外公开ID
	Name             string `json:"name" gorm:"column:name;type:varchar(40);uniqueIndex:idx_of_unit_name;not null;comment:名称;"` // 名称
	SortOrder        int    `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`          // 排序值
	Enabled          bool   `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;default:true;comment:是否启用;"` // 是否启用
	Version          int64  `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                 // 数据版本
}

// TableName 指定ContentUnit对应的数据表名。
func (ContentUnit) TableName() string { return "of_units" }

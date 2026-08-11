package dish

import (
	"time"
)

// SuggestionValidationPolicy 表示当前生效的生成结果兜底校验策略。
type SuggestionValidationPolicy struct {
	SingletonKey             string    `json:"-" gorm:"column:singleton_key;type:varchar(32);primaryKey;not null;comment:单例记录键;"`                                         // 单例记录键
	Version                  int64     `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                // 数据版本
	CatalogValidationEnabled bool      `json:"catalogValidationEnabled" gorm:"column:catalog_validation_enabled;type:tinyint(1) unsigned;not null;comment:是否启用菜品索引兜底校验;"` // 是否启用菜品索引兜底校验
	AppliedByID              uint      `json:"-" gorm:"column:applied_by_id;type:bigint unsigned;not null;comment:应用管理员ID;"`                                              // 应用管理员ID
	AppliedByUsername        string    `json:"-" gorm:"column:applied_by_username;type:varchar(80);not null;comment:应用管理员用户名;"`                                           // 应用管理员用户名
	AppliedByNickname        *string   `json:"-" gorm:"column:applied_by_nickname;type:varchar(80);default:null;comment:应用管理员昵称;"`                                        // 应用管理员昵称
	AppliedAt                time.Time `json:"appliedAt" gorm:"column:applied_at;type:datetime;not null;index;comment:应用时间;"`                                             // 应用时间
	Reason                   string    `json:"reason" gorm:"column:reason;type:varchar(200);not null;comment:原因;"`                                                        // 原因
}

// TableName 指定SuggestionValidationPolicy对应的数据表名。
func (SuggestionValidationPolicy) TableName() string {
	return "of_suggest_policy"
}

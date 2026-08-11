package engagement

import (
	"gorm.io/datatypes"
	"time"
)

const (
	// SubscribeSceneMealStatus 表示饭局最终确认或取消结果订阅场景。
	SubscribeSceneMealStatus = "meal_status"
)

// SubscribeMessageTemplate 表示微信订阅消息模板配置。
type SubscribeMessageTemplate struct {
	ID               string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                 // 主键ID
	Name             string         `json:"name" gorm:"column:name;type:varchar(60);not null;index;comment:模板名称;"`                                  // 模板名称
	WechatTemplateID string         `json:"-" gorm:"column:wechat_template_id;type:varchar(100);not null;uniqueIndex;comment:微信模板ID;"`              // 微信模板ID
	Scene            string         `json:"scene" gorm:"column:scene;type:varchar(32);not null;uniqueIndex:uk_of_sub_template_scene;comment:订阅场景;"` // 订阅场景
	Purpose          string         `json:"purpose" gorm:"column:purpose;type:varchar(160);not null;comment:模板用途;"`                                 // 模板用途
	FieldMappings    datatypes.JSON `json:"fieldMappings" gorm:"column:field_mappings;type:json;not null;comment:字段映射JSON;"`                        // 字段映射JSON
	Enabled          bool           `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;default:false;index;comment:是否启用;"`      // 是否启用
	SendCount        int64          `json:"sendCount" gorm:"column:send_count;type:bigint;not null;default:0;comment:发送次数;"`                        // 发送次数
	FailureCount     int64          `json:"failureCount" gorm:"column:failure_count;type:bigint;not null;default:0;comment:失败次数;"`                  // 失败次数
	Version          int            `json:"version" gorm:"column:version;type:int;not null;default:1;comment:数据版本;"`                                // 数据版本
	CreatedAt        time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                // 创建时间
	UpdatedAt        time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                          // 更新时间
}

// TableName 指定SubscribeMessageTemplate对应的数据表名。
func (SubscribeMessageTemplate) TableName() string { return "of_sub_templates" }

package ai

import (
	"gorm.io/datatypes"
	"time"
)

// AIModel 表示AI模型配置。
type AIModel struct {
	ID                             string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                                  // 主键ID
	ProviderID                     string         `json:"providerId" gorm:"column:provider_id;type:varchar(64);not null;uniqueIndex:uk_of_ai_model_key,priority:1;index;comment:供应商ID;"`           // 供应商ID
	Name                           string         `json:"name" gorm:"column:name;type:varchar(60);not null;comment:名称;"`                                                                           // 名称
	ModelKey                       string         `json:"modelKey" gorm:"column:model_key;type:varchar(120);not null;uniqueIndex:uk_of_ai_model_key,priority:2;comment:供应商模型标识;"`                  // 供应商模型标识
	CapabilitiesJSON               datatypes.JSON `json:"-" gorm:"column:capabilities_json;type:json;not null;comment:模型能力集合JSON;"`                                                                // 模型能力集合JSON
	ContextLength                  *int           `json:"contextLength" gorm:"column:context_length;type:int;default:null;comment:上下文长度;"`                                                         // 上下文长度
	InputPricePerMillionTokensCNY  *string        `json:"inputPricePerMillionTokensCny" gorm:"column:input_price_per_million_tokens_cny;type:varchar(40);default:null;comment:每百万输入Token价格（元）;"`   // 每百万输入Token价格（元）
	OutputPricePerMillionTokensCNY *string        `json:"outputPricePerMillionTokensCny" gorm:"column:output_price_per_million_tokens_cny;type:varchar(40);default:null;comment:每百万输出Token价格（元）;"` // 每百万输出Token价格（元）
	ImagePricePerUnitCNY           *string        `json:"imagePricePerUnitCny" gorm:"column:image_price_per_unit_cny;type:varchar(40);default:null;comment:单张图片价格（元）;"`                            // 单张图片价格（元）
	Enabled                        bool           `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;index;comment:是否启用;"`                                                     // 是否启用
	Remark                         *string        `json:"remark" gorm:"column:remark;type:varchar(300);default:null;comment:备注;"`                                                                  // 备注
	Version                        int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                              // 数据版本
	UpdatedByID                    uint           `json:"-" gorm:"column:updated_by_id;type:bigint unsigned;not null;comment:更新管理员ID;"`                                                            // 更新管理员ID
	UpdatedByUsername              string         `json:"-" gorm:"column:updated_by_username;type:varchar(80);not null;comment:更新管理员用户名;"`                                                         // 更新管理员用户名
	UpdatedByNickname              *string        `json:"-" gorm:"column:updated_by_nickname;type:varchar(80);default:null;comment:更新管理员昵称;"`                                                      // 更新管理员昵称
	LastChangeReason               *string        `json:"-" gorm:"column:last_change_reason;type:varchar(200);default:null;comment:最近变更原因;"`                                                       // 最近变更原因
	CreatedAt                      time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                 // 创建时间
	UpdatedAt                      time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                                           // 更新时间
}

// TableName 指定AIModel对应的数据表名。
func (AIModel) TableName() string {
	return "of_ai_models"
}

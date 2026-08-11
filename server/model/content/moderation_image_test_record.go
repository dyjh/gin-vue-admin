package content

import (
	"gorm.io/datatypes"
	"time"
)

// ModerationImageTest 表示图片审核样例测试。
type ModerationImageTest struct {
	ID                   string           `json:"testId" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                     // 主键ID
	ConfigVersion        int64            `json:"configVersion" gorm:"column:config_version;type:bigint;not null;index;comment:配置版本;"`                            // 配置版本
	MappedStatus         ModerationStatus `json:"mappedStatus" gorm:"column:mapped_status;type:varchar(20);not null;comment:映射后的审核状态;"`                           // 映射后的审核状态
	RiskLabelsJSON       datatypes.JSON   `json:"-" gorm:"column:risk_labels_json;type:json;not null;comment:风险标签JSON;"`                                          // 风险标签JSON
	RiskLevel            *string          `json:"riskLevel" gorm:"column:risk_level;type:varchar(40);default:null;comment:风险等级;"`                                 // 风险等级
	Category             string           `json:"category" gorm:"column:category;type:varchar(40);not null;comment:分类关联数据;"`                                      // 分类关联数据
	ProviderRequestID    *string          `json:"providerRequestId" gorm:"column:provider_request_id;type:varchar(128);default:null;comment:供应商请求ID;"`            // 供应商请求ID
	RequestID            string           `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                              // 请求ID
	DurationMS           int              `json:"durationMs" gorm:"column:duration_ms;type:int;not null;comment:耗时（毫秒）;"`                                         // 耗时（毫秒）
	SafeMessage          string           `json:"safeMessage" gorm:"column:safe_message;type:varchar(240);not null;comment:可安全展示的提示;"`                            // 可安全展示的提示
	ProviderSummaryJSON  datatypes.JSON   `json:"-" gorm:"column:provider_summary_json;type:json;default:null;comment:供应商响应摘要JSON;"`                              // 供应商响应摘要JSON
	TemporaryFileCleaned bool             `json:"temporaryFileCleaned" gorm:"column:temporary_file_cleaned;type:tinyint(1) unsigned;not null;comment:临时文件是否已清理;"` // 临时文件是否已清理
	TestedByID           uint             `json:"-" gorm:"column:tested_by_id;type:bigint unsigned;not null;comment:测试管理员ID;"`                                    // 测试管理员ID
	TestedAt             time.Time        `json:"testedAt" gorm:"column:tested_at;type:datetime;not null;index;comment:测试时间;"`                                    // 测试时间
}

// TableName 指定ModerationImageTest对应的数据表名。
func (ModerationImageTest) TableName() string {
	return "of_mod_image_tests"
}

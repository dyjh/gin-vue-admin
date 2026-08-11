package ai

import (
	"gorm.io/datatypes"
	"time"
)

const (
	// FeatureExecutionProcessing 表示增强功能正在执行。
	FeatureExecutionProcessing = "processing"
	// FeatureExecutionSucceeded 表示增强功能执行成功。
	FeatureExecutionSucceeded = "succeeded"
	// FeatureExecutionFailed 表示增强功能执行失败。
	FeatureExecutionFailed = "failed"

	// FeatureBillingNotCharged 表示本次调用没有扣除积分。
	FeatureBillingNotCharged = "not_charged"
	// FeatureBillingCharged 表示本次调用已经扣除积分。
	FeatureBillingCharged = "charged"
	// FeatureBillingRefundPending 表示失败调用的积分正在等待退还。
	FeatureBillingRefundPending = "refund_pending"
	// FeatureBillingRefunded 表示失败调用的积分已经退还。
	FeatureBillingRefunded = "refunded"
)

// FrontFeatureUsage 表示AI功能调用与计费记录。
type FrontFeatureUsage struct {
	ID                      string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                           // 主键ID
	UserID                  string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                                       // 用户ID
	Feature                 string         `json:"feature" gorm:"column:feature;type:varchar(64);not null;index;comment:客户端功能编码;"`                                                   // 客户端功能编码
	CapabilityCode          string         `json:"capabilityCode" gorm:"column:capability_code;type:varchar(64);not null;default:'';index;comment:实际AI能力编码;"`                        // 实际AI能力编码
	ProviderID              string         `json:"providerId" gorm:"column:provider_id;type:varchar(64);not null;default:'';index;comment:实际供应商ID;"`                                 // 实际供应商ID
	ProviderName            string         `json:"providerName" gorm:"column:provider_name;type:varchar(60);not null;default:'';comment:实际供应商名称快照;"`                                 // 实际供应商名称快照
	ModelID                 string         `json:"modelId" gorm:"column:model_id;type:varchar(64);not null;default:'';index;comment:实际模型ID;"`                                        // 实际模型ID
	ModelName               string         `json:"modelName" gorm:"column:model_name;type:varchar(60);not null;default:'';comment:实际模型名称快照;"`                                        // 实际模型名称快照
	RequestID               string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;default:'';index;comment:请求ID;"`                                     // 请求ID
	IdempotencyKey          string         `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(128);not null;default:'';index;comment:请求幂等键;"`                          // 请求幂等键
	PromptMode              string         `json:"promptMode" gorm:"column:prompt_mode;type:varchar(16);not null;default:'';comment:提示词模式;"`                                         // 提示词模式
	PromptHash              string         `json:"promptHash" gorm:"column:prompt_hash;type:varchar(64);not null;default:'';comment:提示词内容摘要;"`                                       // 提示词内容摘要
	PointCost               int            `json:"pointCost" gorm:"column:point_cost;type:int;not null;default:0;comment:积分消耗;"`                                                     // 积分消耗
	ExecutionStatus         string         `json:"executionStatus" gorm:"column:execution_status;type:varchar(20);not null;index;comment:执行状态;"`                                     // 执行状态
	BillingStatus           string         `json:"billingStatus" gorm:"column:billing_status;type:varchar(20);not null;index;comment:计费状态;"`                                         // 计费状态
	DurationMS              *int           `json:"durationMs" gorm:"column:duration_ms;type:int;default:null;comment:调用耗时毫秒数;"`                                                      // 调用耗时毫秒数
	InputTokens             *int           `json:"inputTokens" gorm:"column:input_tokens;type:int;default:null;comment:输入Token数;"`                                                   // 输入Token数
	OutputTokens            *int           `json:"outputTokens" gorm:"column:output_tokens;type:int;default:null;comment:输出Token数;"`                                                 // 输出Token数
	ImageCount              int            `json:"imageCount" gorm:"column:image_count;type:int;not null;default:0;comment:生成或分析图片数;"`                                               // 生成或分析图片数
	EstimatedCostCNY        string         `json:"estimatedCostCny" gorm:"column:estimated_cost_cny;type:decimal(18,6);not null;default:0;comment:估算成本人民币;"`                         // 估算成本人民币
	FailureCategory         *string        `json:"failureCategory" gorm:"column:failure_category;type:varchar(64);default:null;index;comment:失败分类;"`                                 // 失败分类
	FailureSummary          *string        `json:"failureSummary" gorm:"column:failure_summary;type:varchar(240);default:null;comment:安全失败摘要;"`                                      // 安全失败摘要
	OriginalInputJSON       datatypes.JSON `json:"-" gorm:"column:original_input_json;type:json;default:null;comment:原始输入JSON;"`                                                     // 原始输入JSON
	ImageMetadataJSON       datatypes.JSON `json:"-" gorm:"column:image_metadata_json;type:json;default:null;comment:图片元信息JSON;"`                                                    // 图片元信息JSON
	ModelOutputJSON         datatypes.JSON `json:"-" gorm:"column:model_output_json;type:json;default:null;comment:模型输出JSON;"`                                                       // 模型输出JSON
	SensitiveContentCleared bool           `json:"sensitiveContentCleared" gorm:"column:sensitive_cleared;type:tinyint(1) unsigned;not null;default:false;index;comment:敏感内容是否已清除;"` // 敏感内容是否已清除
	ClearedAt               *time.Time     `json:"clearedAt" gorm:"column:cleared_at;type:datetime;default:null;comment:敏感内容清除时间;"`                                                  // 敏感内容清除时间
	ClearedByID             *uint          `json:"-" gorm:"column:cleared_by_id;type:bigint unsigned;default:null;index;comment:清除管理员ID;"`                                           // 清除管理员ID
	ClearedByUsername       *string        `json:"-" gorm:"column:cleared_by_username;type:varchar(120);default:null;comment:清除管理员用户名;"`                                             // 清除管理员用户名
	ClearedByNickname       *string        `json:"-" gorm:"column:cleared_by_nickname;type:varchar(120);default:null;comment:清除管理员昵称;"`                                              // 清除管理员昵称
	Version                 int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                       // 数据版本
	FinishedAt              *time.Time     `json:"finishedAt" gorm:"column:finished_at;type:datetime;default:null;index;comment:调用结束时间;"`                                            // 调用结束时间
	CreatedAt               time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                                    // 创建时间
	UpdatedAt               time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                          // 更新时间
}

// TableName 指定FrontFeatureUsage对应的数据表名。
func (FrontFeatureUsage) TableName() string { return "of_ai_usages" }

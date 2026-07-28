package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	// SubscribeSceneMealStatus 表示饭局最终确认或取消结果订阅场景。
	SubscribeSceneMealStatus = "meal_status"

	// MealAuthorizationAccept 表示用户接受本次微信订阅授权。
	MealAuthorizationAccept = "accept"
	// MealAuthorizationReject 表示用户拒绝本次微信订阅授权。
	MealAuthorizationReject = "reject"
	// MealAuthorizationBan 表示用户在微信侧永久拒绝本订阅模板。
	MealAuthorizationBan = "ban"

	// SubscribeLogPending 表示订阅消息等待发送。
	SubscribeLogPending = "pending"
	// SubscribeLogSending 表示订阅消息正在发送。
	SubscribeLogSending = "sending"
	// SubscribeLogSent 表示订阅消息已发送。
	SubscribeLogSent = "sent"
	// SubscribeLogFailed 表示订阅消息发送失败。
	SubscribeLogFailed = "failed"
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

// MealFinalResultSubscription 表示用户对单场饭局最终结果的一次微信订阅授权。
type MealFinalResultSubscription struct {
	ID                  string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                     // 主键ID
	MealID              string     `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;index:idx_of_meal_sub_user,priority:1;comment:饭局ID;"` // 饭局ID
	UserID              string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index:idx_of_meal_sub_user,priority:2;comment:用户ID;"` // 用户ID
	TemplateID          string     `json:"templateId" gorm:"column:template_id;type:varchar(64);not null;index;comment:订阅模板配置ID;"`                     // 订阅模板配置ID
	AuthorizationResult string     `json:"authorizationResult" gorm:"column:authorization_result;type:varchar(8);not null;index;comment:微信授权结果;"`      // 微信授权结果
	Accepted            bool       `json:"accepted" gorm:"column:accepted;type:tinyint(1) unsigned;not null;default:false;index;comment:是否获得一次可用授权;"`  // 是否获得一次可用授权
	ConsumedResult      *string    `json:"consumedResult" gorm:"column:consumed_result;type:varchar(16);default:null;comment:消费该授权的最终结果;"`             // 消费该授权的最终结果
	ConsumedAt          *time.Time `json:"consumedAt" gorm:"column:consumed_at;type:datetime;index;default:null;comment:授权消费时间;"`                      // 授权消费时间
	RecordedAt          time.Time  `json:"recordedAt" gorm:"column:recorded_at;type:datetime;not null;index;comment:授权登记时间;"`                          // 授权登记时间
}

// TableName 指定MealFinalResultSubscription对应的数据表名。
func (MealFinalResultSubscription) TableName() string { return "of_meal_subs" }

// SubscribeMessageLog 表示一次微信订阅消息投递记录。
type SubscribeMessageLog struct {
	ID                     string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                      // 主键ID
	UserID                 string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                                  // 用户ID
	TemplateID             string         `json:"templateId" gorm:"column:template_id;type:varchar(64);not null;index;comment:订阅模板配置ID;"`                                      // 订阅模板配置ID
	SubscriptionID         string         `json:"subscriptionId" gorm:"column:subscription_id;type:varchar(64);not null;uniqueIndex;comment:订阅授权ID;"`                          // 订阅授权ID
	Scene                  string         `json:"scene" gorm:"column:scene;type:varchar(32);not null;index;comment:订阅场景;"`                                                     // 订阅场景
	Status                 string         `json:"status" gorm:"column:status;type:varchar(16);not null;index;comment:发送状态;"`                                                   // 发送状态
	RetryCount             int            `json:"retryCount" gorm:"column:retry_count;type:int;not null;default:0;comment:重试次数;"`                                              // 重试次数
	WechatErrorCode        *string        `json:"wechatErrorCode" gorm:"column:wechat_error_code;type:varchar(40);default:null;comment:微信错误码;"`                                // 微信错误码
	RequestID              string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                                           // 请求ID
	AuthorizationChecked   bool           `json:"authorizationChecked" gorm:"column:authorization_checked;type:tinyint(1) unsigned;not null;default:true;comment:是否已校验授权;"`    // 是否已校验授权
	AuthorizationAvailable bool           `json:"authorizationAvailable" gorm:"column:authorization_available;type:tinyint(1) unsigned;not null;default:true;comment:授权是否可用;"` // 授权是否可用
	TargetPage             *string        `json:"targetPage" gorm:"column:target_page;type:varchar(200);default:null;comment:消息跳转页面;"`                                         // 消息跳转页面
	SafePayloadSummary     datatypes.JSON `json:"safePayloadSummary" gorm:"column:safe_payload_summary;type:json;default:null;comment:脱敏消息摘要JSON;"`                            // 脱敏消息摘要JSON
	ErrorSummary           *string        `json:"errorSummary" gorm:"column:error_summary;type:varchar(240);default:null;comment:安全失败说明;"`                                     // 安全失败说明
	CreatedAt              time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                               // 创建时间
	SentAt                 *time.Time     `json:"sentAt" gorm:"column:sent_at;type:datetime;default:null;comment:发送成功时间;"`                                                     // 发送成功时间
}

// TableName 指定SubscribeMessageLog对应的数据表名。
func (SubscribeMessageLog) TableName() string { return "of_sub_logs" }

// SubscribeMessageAttempt 表示订阅消息的一次发送尝试。
type SubscribeMessageAttempt struct {
	ID              uint      `json:"id" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                      // 主键ID
	LogID           string    `json:"logId" gorm:"column:log_id;type:varchar(64);not null;uniqueIndex:uk_of_sub_attempt,priority:1;comment:发送记录ID;"` // 发送记录ID
	Attempt         int       `json:"attempt" gorm:"column:attempt;type:int;not null;uniqueIndex:uk_of_sub_attempt,priority:2;comment:尝试序号;"`        // 尝试序号
	Status          string    `json:"status" gorm:"column:status;type:varchar(16);not null;index;comment:发送结果;"`                                     // 发送结果
	WechatErrorCode *string   `json:"wechatErrorCode" gorm:"column:wechat_error_code;type:varchar(40);default:null;comment:微信错误码;"`                  // 微信错误码
	ErrorSummary    *string   `json:"errorSummary" gorm:"column:error_summary;type:varchar(240);default:null;comment:安全失败说明;"`                       // 安全失败说明
	StartedAt       time.Time `json:"startedAt" gorm:"column:started_at;type:datetime;not null;comment:开始时间;"`                                       // 开始时间
	FinishedAt      time.Time `json:"finishedAt" gorm:"column:finished_at;type:datetime;not null;comment:结束时间;"`                                     // 结束时间
}

// TableName 指定SubscribeMessageAttempt对应的数据表名。
func (SubscribeMessageAttempt) TableName() string { return "of_sub_attempts" }

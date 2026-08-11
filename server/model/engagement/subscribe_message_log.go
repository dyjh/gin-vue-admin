package engagement

import (
	"gorm.io/datatypes"
	"time"
)

const (
	// SubscribeLogPending 表示订阅消息等待发送。
	SubscribeLogPending = "pending"
	// SubscribeLogSending 表示订阅消息正在发送。
	SubscribeLogSending = "sending"
	// SubscribeLogSent 表示订阅消息已发送。
	SubscribeLogSent = "sent"
	// SubscribeLogFailed 表示订阅消息发送失败。
	SubscribeLogFailed = "failed"
)

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

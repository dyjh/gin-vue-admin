package engagement

import (
	"time"
)

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

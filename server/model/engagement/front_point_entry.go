package engagement

import (
	"time"
)

// PointEntryType 表示积分流水的业务变动类型。
type PointEntryType string

const (
	// PointEarned 表示用户获得积分。
	PointEarned PointEntryType = "earned"
	// PointSpent 表示用户消耗积分。
	PointSpent PointEntryType = "spent"
	// PointRefund 表示系统退还已扣积分。
	PointRefund PointEntryType = "refund"
	// PointAdjustment 表示管理员人工调整积分。
	PointAdjustment PointEntryType = "adjustment"
)

// FrontPointEntry 表示积分流水。
type FrontPointEntry struct {
	ID                    string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                    // 主键ID
	UserID                string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                // 用户ID
	Type                  PointEntryType `json:"type" gorm:"column:type;type:varchar(20);not null;index;comment:类型;"`                                       // 类型
	Scene                 string         `json:"scene" gorm:"column:scene;type:varchar(40);not null;default:'';index;comment:积分业务场景;"`                      // 积分业务场景
	Title                 string         `json:"title" gorm:"column:title;type:varchar(100);not null;comment:标题;"`                                          // 标题
	Description           string         `json:"description" gorm:"column:description;type:varchar(300);not null;comment:说明;"`                              // 说明
	Amount                int64          `json:"amount" gorm:"column:amount;type:bigint;not null;comment:积分或用量变动值;"`                                        // 积分或用量变动值
	BalanceAfter          int64          `json:"balanceAfter" gorm:"column:balance_after;type:bigint;not null;default:0;comment:变动后积分余额;"`                  // 变动后积分余额
	RelatedObjectType     *string        `json:"relatedObjectType" gorm:"column:related_object_type;type:varchar(40);index;default:null;comment:关联对象类型;"`   // 关联对象类型
	RelatedObjectID       *string        `json:"relatedObjectId" gorm:"column:related_object_id;type:varchar(64);index;default:null;comment:关联对象ID;"`       // 关联对象ID
	RelatedEntryID        *string        `json:"relatedEntryId" gorm:"column:related_entry_id;type:varchar(64);index;default:null;comment:关联积分流水ID;"`       // 关联积分流水ID
	RequestID             *string        `json:"requestId" gorm:"column:request_id;type:varchar(128);index;default:null;comment:请求ID;"`                     // 请求ID
	IdempotencyKey        *string        `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(160);index;default:null;comment:幂等键;"`            // 幂等键
	AdministratorID       *uint          `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;index;default:null;comment:操作管理员ID;"`   // 操作管理员ID
	AdministratorUsername *string        `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);default:null;comment:管理员用户名;"` // 管理员用户名
	AdministratorNickname *string        `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:管理员昵称;"`  // 管理员昵称
	CreatedAt             time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                             // 创建时间
}

// TableName 指定FrontPointEntry对应的数据表名。
func (FrontPointEntry) TableName() string { return "of_point_entries" }

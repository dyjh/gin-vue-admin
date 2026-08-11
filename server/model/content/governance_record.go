package content

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	"gorm.io/datatypes"
)

const (
	GovernanceTargetDish         = "dish"
	GovernanceTargetOfficialDish = "official_dish"
	GovernanceTargetRecipe       = "recipe"
	GovernanceTargetCheckin      = "checkin"

	GovernanceActionDisableDiscoverability = "disable_discoverability"
	GovernanceActionSoftDeleteDish         = "soft_delete_dish"
	GovernanceActionSoftDeleteOfficialDish = "soft_delete_official_dish"
	GovernanceActionSoftDeleteRecipe       = "soft_delete_recipe"
	GovernanceActionSoftDeleteCheckin      = "soft_delete_checkin"
	GovernanceActionDeleteCopyChain        = "delete_copy_chain"

	GovernanceSeverityNormal  = "normal"
	GovernanceSeveritySerious = "serious"
)

// GovernanceRecord 表示内容处理记录。
type GovernanceRecord struct {
	global.GVA_MODEL                     // GVA基础模型字段
	PublicID              string         `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                   // 对外公开ID
	TargetType            string         `json:"targetType" gorm:"column:target_type;type:varchar(32);not null;index;comment:目标类型;"`                       // 目标类型
	TargetID              string         `json:"targetId" gorm:"column:target_id;type:varchar(64);not null;index;comment:目标ID;"`                           // 目标ID
	TargetLabel           string         `json:"targetLabel" gorm:"column:target_label;type:varchar(160);not null;comment:目标名称;"`                          // 目标名称
	Actions               datatypes.JSON `json:"actions" gorm:"column:actions;type:json;not null;comment:操作集合JSON;"`                                       // 操作集合JSON
	ViolationType         string         `json:"violationType" gorm:"column:violation_type;type:varchar(80);not null;index;comment:违规类型;"`                 // 违规类型
	Severity              string         `json:"severity" gorm:"column:severity;type:varchar(20);not null;comment:处理严重程度;"`                                // 处理严重程度
	Reason                string         `json:"reason" gorm:"column:reason;type:varchar(600);not null;comment:原因;"`                                       // 原因
	AdministratorID       uint           `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`        // 管理员ID
	AdministratorUsername string         `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);not null;comment:管理员用户名;"`    // 管理员用户名
	AdministratorNickname *string        `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:管理员昵称;"` // 管理员昵称
	JobStatus             *string        `json:"jobStatus" gorm:"column:job_status;type:varchar(32);default:null;comment:异步任务状态;"`                         // 异步任务状态
	ImpactSnapshot        datatypes.JSON `json:"impactSnapshot" gorm:"column:impact_snapshot;type:json;not null;comment:影响范围快照JSON;"`                      // 影响范围快照JSON
	BeforeSummary         datatypes.JSON `json:"beforeSummary" gorm:"column:before_summary;type:json;default:null;comment:变更前摘要JSON;"`                     // 变更前摘要JSON
	AfterSummary          datatypes.JSON `json:"afterSummary" gorm:"column:after_summary;type:json;default:null;comment:变更后摘要JSON;"`                       // 变更后摘要JSON
	NotificationIDs       datatypes.JSON `json:"notificationIds" gorm:"column:notification_i_ds;type:json;not null;comment:通知ID集合JSON;"`                   // 通知ID集合JSON
	RequestID             string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                        // 请求ID
	IdempotencyKey        string         `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(160);not null;index;comment:幂等键;"`               // 幂等键
	AffectedCount         int            `json:"affectedCount" gorm:"column:affected_count;type:int;not null;default:0;comment:受影响数量;"`                    // 受影响数量
}

// TableName 指定GovernanceRecord对应的数据表名。
func (GovernanceRecord) TableName() string { return "of_governance_records" }

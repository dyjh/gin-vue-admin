package meal

import (
	"time"
)

// MealStatus 表示饭局从点单到采购完成的生命周期状态。
type MealStatus string

const (
	MealCollecting MealStatus = "collecting"
	MealClosed     MealStatus = "closed"
	MealConfirmed  MealStatus = "confirmed"
	MealCompleted  MealStatus = "completed"
	MealCancelled  MealStatus = "cancelled"
)

// MealCloseSource 表示饭局关闭点单的触发来源。
type MealCloseSource string

const (
	MealCloseSourceCreatorAction MealCloseSource = "creator_action" // 发起人主动关闭
	MealCloseSourceMinuteScan    MealCloseSource = "minute_scan"    // 分钟级截止扫描关闭
	MealCloseSourceServiceGuard  MealCloseSource = "service_guard"  // 业务请求截止守卫关闭
	MealCloseSourceLegacy        MealCloseSource = "legacy"         // 历史数据未记录准确来源
)

// FrontMeal 表示饭局。
type FrontMeal struct {
	ID              string           `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                            // 主键ID
	CreatorID       string           `json:"creatorId" gorm:"column:creator_id;type:varchar(64);not null;index;comment:创建者用户ID;"`               // 创建者用户ID
	Name            string           `json:"name" gorm:"column:name;type:varchar(30);not null;comment:名称;"`                                     // 名称
	Code            string           `json:"code" gorm:"column:code;type:varchar(6);not null;uniqueIndex;comment:饭局邀请码;"`                       // 饭局邀请码
	Status          MealStatus       `json:"status" gorm:"column:status;type:varchar(20);not null;index;comment:状态;"`                           // 状态
	CloseReason     *string          `json:"closeReason" gorm:"column:close_reason;type:varchar(20);default:null;comment:关闭点单原因;"`              // 关闭点单原因
	CloseSource     *MealCloseSource `json:"-" gorm:"column:close_source;type:varchar(24);default:null;comment:关闭点单触发来源;"`                      // 关闭点单触发来源
	ClosedAt        *time.Time       `json:"closedAt" gorm:"column:closed_at;type:datetime;index;default:null;comment:关闭点单时间;"`                 // 关闭点单时间
	DeadlineAt      time.Time        `json:"deadlineAt" gorm:"column:deadline_at;type:datetime;not null;index;comment:点单截止时间;"`                 // 点单截止时间
	SourceRecipeID  *string          `json:"sourceRecipeId" gorm:"column:source_recipe_id;type:varchar(64);default:null;comment:来源菜谱ID;"`       // 来源菜谱ID
	CancelledReason *string          `json:"-" gorm:"column:cancelled_reason;type:varchar(100);default:null;comment:取消原因;"`                     // 取消原因
	CancelledFrom   *MealStatus      `json:"-" gorm:"column:cancelled_from_status;type:varchar(20);default:null;comment:取消前饭局状态;"`              // 取消前饭局状态
	ShoppingListID  *string          `json:"shoppingListId" gorm:"column:shopping_list_id;type:varchar(64);index;default:null;comment:采购清单ID;"` // 采购清单ID
	ConfirmedAt     *time.Time       `json:"confirmedAt" gorm:"column:confirmed_at;type:datetime;default:null;comment:确认时间;"`                   // 确认时间
	CompletedAt     *time.Time       `json:"completedAt" gorm:"column:completed_at;type:datetime;default:null;comment:完成时间;"`                   // 完成时间
	CancelledAt     *time.Time       `json:"cancelledAt" gorm:"column:cancelled_at;type:datetime;default:null;comment:取消时间;"`                   // 取消时间
	CreatedAt       time.Time        `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                     // 创建时间
	UpdatedAt       time.Time        `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                           // 更新时间
}

// TableName 指定FrontMeal对应的数据表名。
func (FrontMeal) TableName() string { return "of_meals" }

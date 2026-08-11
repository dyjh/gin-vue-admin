package response

import (
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	"time"
)

// ShoppingListAdminSummary 表示管理端采购清单摘要。
type ShoppingListAdminSummary struct {
	ID             string                       `json:"id"`             // 采购清单ID
	MealID         string                       `json:"mealId"`         // 饭局ID
	MealName       string                       `json:"mealName"`       // 饭局名称
	Creator        commonResponse.UserReference `json:"creator"`        // 创建者
	TotalCount     int                          `json:"totalCount"`     // 清单项总数
	PendingCount   int                          `json:"pendingCount"`   // 待采购数量
	CompletedCount int                          `json:"completedCount"` // 已采购数量
	ShareStatus    string                       `json:"shareStatus"`    // 分享状态
	ShareExpiresAt *time.Time                   `json:"shareExpiresAt"` // 分享过期时间
	ShareRevokedAt *time.Time                   `json:"shareRevokedAt"` // 分享撤销时间
	CreatedAt      time.Time                    `json:"createdAt"`      // 创建时间
	UpdatedAt      time.Time                    `json:"updatedAt"`      // 更新时间
}

// ShoppingItemAdmin 表示管理端采购清单项。
type ShoppingItemAdmin struct {
	ID              string   `json:"id"`              // 清单项ID
	Name            string   `json:"name"`            // 名称
	Amount          string   `json:"amount"`          // 用量
	Note            *string  `json:"note"`            // 备注
	Completed       bool     `json:"completed"`       // 是否已完成
	SourceDishNames []string `json:"sourceDishNames"` // 来源菜品名称列表
	SortOrder       int      `json:"sortOrder"`       // 排序值
}

// ShoppingListAdminDetail 表示管理端采购清单详情。
type ShoppingListAdminDetail struct {
	ShoppingListAdminSummary                     // 采购清单摘要
	ShareTokenMasked         *string             `json:"shareTokenMasked"`         // 脱敏分享令牌
	Items                    []ShoppingItemAdmin `json:"items"`                    // 清单项列表
	GeneratedFromSnapshotIDs []string            `json:"generatedFromSnapshotIds"` // 生成来源快照ID列表
}

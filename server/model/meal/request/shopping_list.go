package request

import (
	"time"
)

// ShoppingListAdminListQuery 表示管理端采购清单分页查询条件。
type ShoppingListAdminListQuery struct {
	Page           int        `form:"page" json:"page" binding:"omitempty,min=1"`                                      // 页码
	PageSize       int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                    // 每页数量
	ShoppingListID string     `form:"shoppingListId" json:"shoppingListId" binding:"omitempty,max=64"`                 // 采购清单ID
	MealID         string     `form:"mealId" json:"mealId" binding:"omitempty,max=64"`                                 // 饭局ID
	CreatorID      string     `form:"creatorId" json:"creatorId" binding:"omitempty,max=64"`                           // 创建者用户ID
	ShareStatus    string     `form:"shareStatus" json:"shareStatus" binding:"omitempty,oneof=active expired revoked"` // 分享状态
	CreatedFrom    *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`          // 创建起始时间
	CreatedTo      *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`              // 创建结束时间
	SortBy         string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt pendingCount"`           // 排序字段
	SortOrder      string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                   // 排序方向
}

// ApplyDefaults 补齐管理端采购清单列表的分页和排序默认值。
func (query *ShoppingListAdminListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "createdAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// ShoppingListAdminIDPath 表示管理端采购清单ID路径参数。
type ShoppingListAdminIDPath struct {
	ShoppingListID string `uri:"shoppingListId" json:"shoppingListId" binding:"required,max=64"` // 采购清单ID
}

package request

import (
	"time"
)

// MealAdminListQuery 表示管理端饭局分页查询条件。
type MealAdminListQuery struct {
	Page          int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                     // 页码
	PageSize      int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                   // 每页数量
	Keyword       string     `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                                              // 饭局名称或ID关键词
	CreatorID     string     `form:"creatorId" json:"creatorId" binding:"omitempty,max=64"`                                          // 创建者用户ID
	Status        string     `form:"status" json:"status" binding:"omitempty,oneof=collecting closed confirmed completed cancelled"` // 饭局状态
	CloseReason   string     `form:"closeReason" json:"closeReason" binding:"omitempty,oneof=manual deadline"`                       // 关闭点单原因
	CancelReason  string     `form:"cancelReason" json:"cancelReason" binding:"omitempty,oneof=manual creator_disabled"`             // 取消原因分类
	CreatedFrom   *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                         // 创建起始时间
	CreatedTo     *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                             // 创建结束时间
	DeadlineFrom  *time.Time `form:"deadlineFrom" json:"deadlineFrom" time_format:"2006-01-02T15:04:05Z07:00"`                       // 截止起始时间
	DeadlineTo    *time.Time `form:"deadlineTo" json:"deadlineTo" time_format:"2006-01-02T15:04:05Z07:00"`                           // 截止结束时间
	JoinedFrom    *time.Time `form:"joinedFrom" json:"joinedFrom" time_format:"2006-01-02T15:04:05Z07:00"`                           // 参与者加入起始时间
	JoinedTo      *time.Time `form:"joinedTo" json:"joinedTo" time_format:"2006-01-02T15:04:05Z07:00"`                               // 参与者加入结束时间
	ConfirmedFrom *time.Time `form:"confirmedFrom" json:"confirmedFrom" time_format:"2006-01-02T15:04:05Z07:00"`                     // 确认菜单起始时间
	ConfirmedTo   *time.Time `form:"confirmedTo" json:"confirmedTo" time_format:"2006-01-02T15:04:05Z07:00"`                         // 确认菜单结束时间
	CompletedFrom *time.Time `form:"completedFrom" json:"completedFrom" time_format:"2006-01-02T15:04:05Z07:00"`                     // 完成饭局起始时间
	CompletedTo   *time.Time `form:"completedTo" json:"completedTo" time_format:"2006-01-02T15:04:05Z07:00"`                         // 完成饭局结束时间
	SortBy        string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt deadlineAt confirmedAt completedAt"`    // 排序字段
	SortOrder     string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                  // 排序方向
}

// ApplyDefaults 补齐管理端饭局列表的分页和排序默认值。
func (query *MealAdminListQuery) ApplyDefaults() {
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

// MealAdminIDPath 表示管理端饭局ID路径参数。
type MealAdminIDPath struct {
	MealID string `uri:"mealId" json:"mealId" binding:"required,max=64"` // 饭局ID
}

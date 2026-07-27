package request

import "github.com/dyjh/order-food-mini-app/server/model"

// UserIDPath 表示用户ID路径参数。
type UserIDPath struct {
	UserID string `uri:"userId" json:"userId" binding:"required,max=64"` // 用户ID
}

// IdempotencyHeader 表示幂等请求头请求参数。
type IdempotencyHeader struct {
	Key string `header:"X-Idempotency-Key" json:"idempotencyKey" binding:"required,max=128"` // 键
}

// UserListQuery 表示用户列表查询条件。
type UserListQuery struct {
	Page                int    `form:"page" json:"page" binding:"omitempty,min=1"`                                                // 页码
	PageSize            int    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                              // 每页数量
	Keyword             string `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                                         // 关键词
	Status              string `form:"status" json:"status" binding:"omitempty,oneof=normal disabled"`                            // 状态
	CapabilityEffective string `form:"capabilityEffective" json:"capabilityEffective" binding:"omitempty,oneof=enabled disabled"` // 平台增强能力是否生效
	RegisteredFrom      string `form:"registeredFrom" json:"registeredFrom" binding:"omitempty"`                                  // 注册起始时间
	RegisteredTo        string `form:"registeredTo" json:"registeredTo" binding:"omitempty"`                                      // 注册结束时间
	LastLoginFrom       string `form:"lastLoginFrom" json:"lastLoginFrom" binding:"omitempty"`                                    // 最近登录起始时间
	LastLoginTo         string `form:"lastLoginTo" json:"lastLoginTo" binding:"omitempty"`                                        // 最近登录结束时间
	ActiveFrom          string `form:"activeFrom" json:"activeFrom" binding:"omitempty"`                                          // 活跃起始时间
	ActiveTo            string `form:"activeTo" json:"activeTo" binding:"omitempty"`                                              // 活跃结束时间
	MinPoints           *int64 `form:"minPoints" json:"minPoints" binding:"omitempty"`                                            // 最小积分
	MaxPoints           *int64 `form:"maxPoints" json:"maxPoints" binding:"omitempty"`                                            // 最大积分
	SortBy              string `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt lastLoginAt points"`               // 排序字段
	SortOrder           string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                             // 排序方向
}

// ApplyDefaults 补齐分页、筛选和排序默认值。
func (query *UserListQuery) ApplyDefaults() {
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

// UpdateUserStatusBody 表示更新用户状态请求正文。
type UpdateUserStatusBody struct {
	Status          model.UserStatus `json:"status" binding:"required,oneof=normal disabled"` // 状态
	Reason          string           `json:"reason" binding:"required,min=4,max=200"`         // 原因
	ExpectedVersion int64            `json:"expectedVersion" binding:"required,min=1"`        // 预期版本
}

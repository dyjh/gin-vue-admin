package request

// AuditLogListQuery 表示审计日志列表查询条件。
type AuditLogListQuery struct {
	Page            int    `form:"page" json:"page" binding:"omitempty,min=1"`                        // 页码
	PageSize        int    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`      // 每页数量
	AdministratorID string `form:"administratorId" json:"administratorId" binding:"omitempty,max=20"` // 管理员ID
	TargetType      string `form:"targetType" json:"targetType" binding:"omitempty,max=60"`           // 目标类型
	TargetID        string `form:"targetId" json:"targetId" binding:"omitempty,max=80"`               // 目标ID
	Action          string `form:"action" json:"action" binding:"omitempty,max=80"`                   // 操作
	RequestID       string `form:"requestId" json:"requestId" binding:"omitempty,max=128"`            // 请求ID
	CreatedFrom     string `form:"createdFrom" json:"createdFrom" binding:"omitempty"`                // 创建起始时间
	CreatedTo       string `form:"createdTo" json:"createdTo" binding:"omitempty"`                    // 创建结束时间
	SortBy          string `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt"`          // 排序字段
	SortOrder       string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`     // 排序方向
}

// ApplyDefaults 补齐分页、筛选和排序默认值。
func (query *AuditLogListQuery) ApplyDefaults() {
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

// AuditLogIDPath 表示审计日志ID路径参数。
type AuditLogIDPath struct {
	AuditLogID string `uri:"auditLogId" json:"auditLogId" binding:"required,max=36"` // 审计日志ID
}

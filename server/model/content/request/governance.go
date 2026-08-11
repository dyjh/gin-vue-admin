package request

import (
	"time"
)

// GovernanceRecordListQuery 表示违规处理记录分页查询条件。
type GovernanceRecordListQuery struct {
	Page            int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                                                                                                 // 页码
	PageSize        int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                                                                                               // 每页数量
	TargetType      string     `form:"targetType" json:"targetType" binding:"omitempty,oneof=dish official_dish recipe checkin"`                                                                                   // 目标类型
	TargetID        string     `form:"targetId" json:"targetId" binding:"omitempty,max=64"`                                                                                                                        // 目标ID
	Action          string     `form:"action" json:"action" binding:"omitempty,oneof=disable_discoverability soft_delete_dish soft_delete_official_dish soft_delete_recipe soft_delete_checkin delete_copy_chain"` // 处理动作
	ViolationType   string     `form:"violationType" json:"violationType" binding:"omitempty,max=80"`                                                                                                              // 违规类型
	JobStatus       string     `form:"jobStatus" json:"jobStatus" binding:"omitempty,oneof=pending processing partially_succeeded succeeded failed abnormal"`                                                      // 异步任务状态或异常汇总
	AdministratorID string     `form:"administratorId" json:"administratorId" binding:"omitempty,max=32"`                                                                                                          // 执行管理员ID
	CreatedFrom     *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                                                                                     // 创建起始时间
	CreatedTo       *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                                                                                                         // 创建结束时间
	SortBy          string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt updatedAt"`                                                                                                         // 排序字段
	SortOrder       string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                                                                                              // 排序方向
}

// ApplyDefaults 补齐违规处理记录列表的分页和排序默认值。
func (query *GovernanceRecordListQuery) ApplyDefaults() {
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

// GovernanceRecordIDPath 表示违规处理记录ID路径参数。
type GovernanceRecordIDPath struct {
	RecordID string `uri:"recordId" json:"recordId" binding:"required,max=64"` // 违规处理记录ID
}

// GovernanceJobIDPath 表示违规处理任务ID路径参数。
type GovernanceJobIDPath struct {
	JobID string `uri:"jobId" json:"jobId" binding:"required,max=64"` // 违规处理任务ID
}

// GovernanceJobRetryInput 表示违规处理任务失败项重试参数。
type GovernanceJobRetryInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200"`  // 重试原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"` // 预期任务版本
}

package request

import "time"

// PointEntryListQuery 表示管理端积分流水列表查询条件。
type PointEntryListQuery struct {
	Page              int        `form:"page" json:"page" binding:"omitempty,min=1"`                                // 页码
	PageSize          int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`              // 每页数量
	UserID            string     `form:"userId" json:"userId" binding:"omitempty,max=64"`                           // 用户ID
	Type              string     `form:"type" json:"type" binding:"omitempty,oneof=earned spent refund adjustment"` // 流水类型
	Scene             string     `form:"scene" json:"scene" binding:"omitempty,max=40"`                             // 业务场景
	RelatedObjectType string     `form:"relatedObjectType" json:"relatedObjectType" binding:"omitempty,max=40"`     // 关联对象类型
	RelatedObjectID   string     `form:"relatedObjectId" json:"relatedObjectId" binding:"omitempty,max=64"`         // 关联对象ID
	RequestID         string     `form:"requestId" json:"requestId" binding:"omitempty,max=128" checksql:"false"`   // 请求ID
	IdempotencyKey    string     `form:"idempotencyKey" json:"idempotencyKey" binding:"omitempty,max=160"`          // 幂等键
	CreatedFrom       *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`    // 创建起始时间
	CreatedTo         *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`        // 创建结束时间
	SortBy            string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt amount"`           // 排序字段
	SortOrder         string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`             // 排序方向
}

// ApplyDefaults 补齐积分流水列表的分页和排序默认值。
func (query *PointEntryListQuery) ApplyDefaults() {
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

// PointAdjustmentPreviewInput 表示人工积分调整预览参数。
type PointAdjustmentPreviewInput struct {
	UserID    string `json:"userId" binding:"required,max=64"`                         // 用户ID
	Direction string `json:"direction" binding:"required,oneof=credit debit"`          // 调整方向
	Amount    int64  `json:"amount" binding:"required,min=1,max=100000"`               // 调整积分
	Reason    string `json:"reason" binding:"required,min=4,max=200" checksql:"false"` // 调整原因
}

// PointAdjustmentCreateInput 表示人工积分调整提交参数。
type PointAdjustmentCreateInput struct {
	PreviewToken        string `json:"previewToken" binding:"required,max=1000"`                 // 调整预览令牌
	UserID              string `json:"userId" binding:"required,max=64"`                         // 用户ID
	Direction           string `json:"direction" binding:"required,oneof=credit debit"`          // 调整方向
	Amount              int64  `json:"amount" binding:"required,min=1,max=100000"`               // 调整积分
	Reason              string `json:"reason" binding:"required,min=4,max=200" checksql:"false"` // 调整原因
	ExpectedUserVersion int64  `json:"expectedUserVersion" binding:"required,min=1"`             // 预期用户版本
}

// PointRuleEmptyQuery 表示积分规则空查询条件。
type PointRuleEmptyQuery struct{}

// PointRuleUpdateInput 表示积分规则直接保存参数。
type PointRuleUpdateInput struct {
	DailyCheckinReward int64  `json:"dailyCheckinReward" binding:"required,min=1,max=100"`      // 每日首次打卡奖励积分
	Reason             string `json:"reason" binding:"required,min=2,max=200" checksql:"false"` // 修改原因
	ExpectedVersion    int64  `json:"expectedVersion" binding:"required,min=1"`                 // 当前配置版本
}

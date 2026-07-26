package response

import "time"

// PointEntry 表示管理端积分流水响应数据。
type PointEntry struct {
	ID                string                `json:"id"`                // 流水ID
	User              UserReference         `json:"user"`              // 用户
	Type              string                `json:"type"`              // 流水类型
	Scene             string                `json:"scene"`             // 业务场景
	Title             string                `json:"title"`             // 标题
	Description       string                `json:"description"`       // 说明
	Amount            int64                 `json:"amount"`            // 变动积分
	BalanceAfter      int64                 `json:"balanceAfter"`      // 变动后余额
	RelatedObjectType *string               `json:"relatedObjectType"` // 关联对象类型
	RelatedObjectID   *string               `json:"relatedObjectId"`   // 关联对象ID
	RelatedEntryID    *string               `json:"relatedEntryId"`    // 关联积分流水ID
	RequestID         *string               `json:"requestId"`         // 请求ID
	IdempotencyKey    *string               `json:"idempotencyKey"`    // 幂等键
	Administrator     *AdministratorSummary `json:"administrator"`     // 操作管理员
	CreatedAt         time.Time             `json:"createdAt"`         // 创建时间
}

// PointAdjustmentPreview 表示人工积分调整预览。
type PointAdjustmentPreview struct {
	PreviewToken        string        `json:"previewToken"`        // 调整预览令牌
	ExpiresAt           time.Time     `json:"expiresAt"`           // 令牌过期时间
	User                UserReference `json:"user"`                // 目标用户
	Direction           string        `json:"direction"`           // 调整方向
	Amount              int64         `json:"amount"`              // 调整积分
	BalanceBefore       int64         `json:"balanceBefore"`       // 调整前余额
	BalanceAfter        int64         `json:"balanceAfter"`        // 调整后余额
	ExpectedUserVersion int64         `json:"expectedUserVersion"` // 预期用户版本
}

// PointAdjustmentResult 表示人工积分调整执行结果。
type PointAdjustmentResult struct {
	UserID        string     `json:"userId"`        // 用户ID
	BalanceBefore int64      `json:"balanceBefore"` // 调整前余额
	BalanceAfter  int64      `json:"balanceAfter"`  // 调整后余额
	PointEntry    PointEntry `json:"pointEntry"`    // 新增积分流水
	UserVersion   int64      `json:"userVersion"`   // 更新后用户版本
}

// PointRuleConfig 表示当前生效的积分规则。
type PointRuleConfig struct {
	Version            int64                `json:"version"`            // 版本
	DailyCheckinReward int64                `json:"dailyCheckinReward"` // 每日首次打卡奖励积分
	UpdatedBy          AdministratorSummary `json:"updatedBy"`          // 更新人
	UpdatedAt          time.Time            `json:"updatedAt"`          // 更新时间
	Reason             string               `json:"reason"`             // 原因
}

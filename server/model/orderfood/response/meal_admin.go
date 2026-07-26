package response

import "time"

// MealAdminSummary 表示管理端饭局摘要。
type MealAdminSummary struct {
	ID                  string        `json:"id"`                  // 饭局ID
	Name                string        `json:"name"`                // 饭局名称
	Creator             UserReference `json:"creator"`             // 创建者
	CodeMasked          string        `json:"codeMasked"`          // 脱敏邀请码
	Status              string        `json:"status"`              // 饭局状态
	CloseReason         *string       `json:"closeReason"`         // 关闭点单原因
	CloseSource         *string       `json:"closeSource"`         // 关闭点单触发来源
	CancelReason        *string       `json:"cancelReason"`        // 取消原因分类
	CancelledFromStatus *string       `json:"cancelledFromStatus"` // 取消前饭局状态
	ParticipantCount    int           `json:"participantCount"`    // 参与人数
	CandidateCount      int           `json:"candidateCount"`      // 候选菜数量
	FinalDishCount      int           `json:"finalDishCount"`      // 最终菜品数量
	ShoppingListID      *string       `json:"shoppingListId"`      // 采购清单ID
	CreatedAt           time.Time     `json:"createdAt"`           // 创建时间
	DeadlineAt          time.Time     `json:"deadlineAt"`          // 点单截止时间
	ClosedAt            *time.Time    `json:"closedAt"`            // 关闭点单时间
	ConfirmedAt         *time.Time    `json:"confirmedAt"`         // 确认菜单时间
	CompletedAt         *time.Time    `json:"completedAt"`         // 完成饭局时间
	CancelledAt         *time.Time    `json:"cancelledAt"`         // 取消饭局时间
}

// MealTimelineEvent 表示饭局状态时间线事件。
type MealTimelineEvent struct {
	Type       string    `json:"type"`       // 事件类型
	OccurredAt time.Time `json:"occurredAt"` // 发生时间
	ActorType  string    `json:"actorType"`  // 操作者类型
	ActorID    *string   `json:"actorId"`    // 操作者ID
	Summary    string    `json:"summary"`    // 事件摘要
}

// MealCandidateAdmin 表示管理端饭局候选菜摘要。
type MealCandidateAdmin struct {
	ID            string `json:"id"`            // 候选菜ID
	DishID        string `json:"dishId"`        // 菜品公开ID
	Name          string `json:"name"`          // 菜品名称
	Status        string `json:"status"`        // 候选状态
	VoteCount     int    `json:"voteCount"`     // 点选人数
	Selected      bool   `json:"selected"`      // 是否进入最终菜单
	FinalServings *int   `json:"finalServings"` // 最终份数
}

// MealFinalDishAdminSnapshot 表示管理端最终菜单只读快照。
type MealFinalDishAdminSnapshot struct {
	ID            string      `json:"id"`            // 快照ID
	CandidateID   string      `json:"candidateId"`   // 候选菜ID
	DishID        string      `json:"dishId"`        // 菜品公开ID
	Name          string      `json:"name"`          // 快照菜名
	CoverURL      *string     `json:"coverUrl"`      // 快照封面
	FinalServings int         `json:"finalServings"` // 最终份数
	Ingredients   interface{} `json:"ingredients"`   // 食材快照
	Steps         interface{} `json:"steps"`         // 步骤快照
	CreatedAt     time.Time   `json:"createdAt"`     // 快照创建时间
}

// MealAdminDetail 表示管理端饭局详情。
type MealAdminDetail struct {
	MealAdminSummary                                   // 饭局摘要
	Participants          []UserReference              `json:"participants"`          // 参与者列表
	Candidates            []MealCandidateAdmin         `json:"candidates"`            // 候选菜列表
	Timeline              []MealTimelineEvent          `json:"timeline"`              // 状态时间线
	FinalDishSnapshots    []MealFinalDishAdminSnapshot `json:"finalDishSnapshots"`    // 最终菜单快照
	ShoppingListSummary   *ShoppingListAdminSummary    `json:"shoppingListSummary"`   // 采购清单摘要
	ShareRevocationStatus *string                      `json:"shareRevocationStatus"` // 发起人禁用时的采购分享撤销结果
}

// ShoppingListAdminSummary 表示管理端采购清单摘要。
type ShoppingListAdminSummary struct {
	ID             string        `json:"id"`             // 采购清单ID
	MealID         string        `json:"mealId"`         // 饭局ID
	MealName       string        `json:"mealName"`       // 饭局名称
	Creator        UserReference `json:"creator"`        // 创建者
	TotalCount     int           `json:"totalCount"`     // 清单项总数
	PendingCount   int           `json:"pendingCount"`   // 待采购数量
	CompletedCount int           `json:"completedCount"` // 已采购数量
	ShareStatus    string        `json:"shareStatus"`    // 分享状态
	ShareExpiresAt *time.Time    `json:"shareExpiresAt"` // 分享过期时间
	ShareRevokedAt *time.Time    `json:"shareRevokedAt"` // 分享撤销时间
	CreatedAt      time.Time     `json:"createdAt"`      // 创建时间
	UpdatedAt      time.Time     `json:"updatedAt"`      // 更新时间
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

package response

import (
	commonResponse "github.com/dyjh/order-food-mini-app/server/model/common/response"
	"time"
)

// CapabilityEffects 表示能力联动效果响应数据。
type CapabilityEffects struct {
	PointsVisible             bool `json:"pointsVisible"`             // 积分是否可见
	PointEntriesVisible       bool `json:"pointEntriesVisible"`       // 积分流水是否可见
	CheckinRewardEnabled      bool `json:"checkinRewardEnabled"`      // 是否启用打卡积分奖励
	PreferenceAnalysisEnabled bool `json:"preferenceAnalysisEnabled"` // 是否启用偏好画像分析
}

// UserSummary 表示用户摘要响应数据。
type UserSummary struct {
	ID                  string     `json:"id"`                  // ID
	AvatarURL           *string    `json:"avatarUrl"`           // 头像地址
	Nickname            string     `json:"nickname"`            // 昵称
	Points              int64      `json:"points"`              // 积分
	CapabilityEffective string     `json:"capabilityEffective"` // 平台增强能力是否生效
	CapabilitySource    string     `json:"capabilitySource"`    // 增强能力状态来源
	CapabilityDisabled  bool       `json:"capabilityDisabled"`  // 是否已对该用户单独关闭AI能力
	CheckinDayCount     int        `json:"checkinDayCount"`     // 打卡天数量
	DishCount           int        `json:"dishCount"`           // 菜品数量
	MealCount           int        `json:"mealCount"`           // 饭局数量
	Status              string     `json:"status"`              // 状态
	Version             int64      `json:"version"`             // 版本
	RegisteredAt        time.Time  `json:"registeredAt"`        // 注册时间
	LastLoginAt         *time.Time `json:"lastLoginAt"`         // 最近登录时间
}

// UserDetail 表示用户详情响应数据。
type UserDetail struct {
	UserSummary                                   // 用户摘要
	WechatIdentityMasked    *string               `json:"wechatIdentityMasked"`    // 脱敏后的微信身份标识
	DisabledReason          *string               `json:"disabledReason"`          // 停用原因
	DisabledAt              *time.Time            `json:"disabledAt"`              // 停用时间
	CapabilityEffects       CapabilityEffects     `json:"capabilityEffects"`       // 能力联动效果
	RecipeCount             int                   `json:"recipeCount"`             // 菜谱数量
	CheckinCount            int                   `json:"checkinCount"`            // 打卡数量
	ActiveCreatedMealImpact *UserActiveMealImpact `json:"activeCreatedMealImpact"` // 用户发起的进行中饭局影响
}

// UserActiveMealImpact 表示禁用用户前需要收口的发起饭局影响。
type UserActiveMealImpact struct {
	MealID             string `json:"mealId"`             // 饭局ID
	Status             string `json:"status"`             // 饭局当前状态
	ParticipantCount   int64  `json:"participantCount"`   // 参与者数量
	FinalSnapshotCount int64  `json:"finalSnapshotCount"` // 将保留的最终菜单快照数量
	ShoppingListCount  int64  `json:"shoppingListCount"`  // 将保留的采购清单数量
}

// UserStatusResult 表示用户状态结果响应数据。
type UserStatusResult struct {
	UserID                     string    `json:"userId"`                     // 用户ID
	Status                     string    `json:"status"`                     // 状态
	DisabledReason             *string   `json:"disabledReason"`             // 停用原因
	CancelledMealCount         int64     `json:"cancelledMealCount"`         // 因禁用而取消的饭局数量
	NotifiedParticipantCount   int64     `json:"notifiedParticipantCount"`   // 已写入站内通知的参与者数量
	PreservedSnapshotCount     int64     `json:"preservedSnapshotCount"`     // 保留的最终菜单快照数量
	PreservedShoppingListCount int64     `json:"preservedShoppingListCount"` // 保留的采购清单数量
	Version                    int64     `json:"version"`                    // 版本
	UpdatedAt                  time.Time `json:"updatedAt"`                  // 更新时间
}

// UserCapabilityResult 表示用户AI能力单独关闭状态的更新结果。
type UserCapabilityResult struct {
	UserID              string    `json:"userId"`
	CapabilityDisabled  bool      `json:"capabilityDisabled"`
	CapabilityEffective string    `json:"capabilityEffective"`
	CapabilitySource    string    `json:"capabilitySource"`
	Version             int64     `json:"version"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// PreferenceTermSummary 表示偏好画像词条摘要响应数据。
type PreferenceTermSummary struct {
	ID     *string `json:"id"`     // ID
	Name   string  `json:"name"`   // 名称
	Origin string  `json:"origin"` // 来源
}

// PreferenceTextSummary 表示偏好画像文本摘要响应数据。
type PreferenceTextSummary struct {
	Text   string `json:"text"`   // 文本
	Origin string `json:"origin"` // 来源
}

// PreferenceUserSettingSummary 表示偏好画像用户设置摘要响应数据。
type PreferenceUserSettingSummary struct {
	Type      string    `json:"type"`      // 类型
	Value     string    `json:"value"`     // 值
	UpdatedAt time.Time `json:"updatedAt"` // 更新时间
}

// PreferenceEvidenceSourceSummary 表示偏好画像证据来源摘要响应数据。
type PreferenceEvidenceSourceSummary struct {
	SourceType       string     `json:"sourceType"`       // 来源类型
	TotalCount       int        `json:"totalCount"`       // 总数量
	AggregatedCount  int        `json:"aggregatedCount"`  // 已聚合数量
	PendingCount     int        `json:"pendingCount"`     // 待处理数量
	LatestOccurredAt *time.Time `json:"latestOccurredAt"` // 最新发生时间
}

// PreferenceProfileContent 表示偏好画像的结构化内容。
type PreferenceProfileContent struct {
	CommonDishTags               []PreferenceTermSummary        `json:"commonDishTags"`               // 常用菜品标签列表
	FrequentlyOrderedDishTags    []PreferenceTermSummary        `json:"frequentlyOrderedDishTags"`    // 经常点选的菜品标签
	CommonIngredients            []PreferenceTermSummary        `json:"commonIngredients"`            // 常用食材列表
	TastePreferenceSummary       *PreferenceTextSummary         `json:"tastePreferenceSummary"`       // 口味偏好摘要
	AvoidanceOrPreferenceSummary *PreferenceTextSummary         `json:"avoidanceOrPreferenceSummary"` // 忌口或偏好摘要
	UserSettings                 []PreferenceUserSettingSummary `json:"userSettings"`                 // 用户主动设置的偏好项
}

// PreferenceProfileFailureSummary 表示最近一次画像更新失败摘要。
type PreferenceProfileFailureSummary struct {
	FailedAt        time.Time `json:"failedAt"`        // 失败时间
	StableErrorCode string    `json:"stableErrorCode"` // 稳定错误编码
	SafeMessage     string    `json:"safeMessage"`     // 安全消息
	RequestID       string    `json:"requestId"`       // 请求ID
}

// UserPreferenceProfile 表示用户偏好画像及其更新状态。
type UserPreferenceProfile struct {
	User                commonResponse.UserReference      `json:"user"`                // 用户
	HasProfile          bool                              `json:"hasProfile"`          // 是否有画像
	UpdateState         string                            `json:"updateState"`         // 偏好画像更新状态
	UpdateEnabled       bool                              `json:"updateEnabled"`       // 是否允许继续更新偏好画像
	PausedReasonCode    *string                           `json:"pausedReasonCode"`    // 偏好画像暂停原因编码
	PausedReasonSummary *string                           `json:"pausedReasonSummary"` // 偏好画像暂停原因说明
	CheckinDayCount     int                               `json:"checkinDayCount"`     // 打卡天数量
	EvidenceSources     []PreferenceEvidenceSourceSummary `json:"evidenceSources"`     // 偏好画像证据来源统计
	LastEvidenceAt      *time.Time                        `json:"lastEvidenceAt"`      // 最近证据时间
	LastAggregatedAt    *time.Time                        `json:"lastAggregatedAt"`    // 最近已聚合时间
	ProfileUpdatedAt    *time.Time                        `json:"profileUpdatedAt"`    // 画像更新时间
	Profile             *PreferenceProfileContent         `json:"profile"`             // 偏好画像内容
	LatestFailure       *PreferenceProfileFailureSummary  `json:"latestFailure"`       // 最新失败
	AccessAuditID       string                            `json:"accessAuditId"`       // 敏感访问审计记录ID
}

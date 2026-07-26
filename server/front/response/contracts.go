package response

import (
	"time"

	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
)

// Page 表示分页响应数据。
type Page[T any] struct {
	Page     int   `json:"page"`     // 页码
	PageSize int   `json:"pageSize"` // 每页数量
	Total    int64 `json:"total"`    // 总数量
	List     []T   `json:"list"`     // 列表
}

// DishCategory 表示菜品分类响应数据。
type DishCategory struct {
	ID        string `json:"id"`        // ID
	Name      string `json:"name"`      // 名称
	SortOrder int    `json:"sortOrder"` // 排序值
	Enabled   bool   `json:"enabled"`   // 是否启用
}

// DishTag 表示菜品标签响应数据。
type DishTag struct {
	ID        string `json:"id"`        // ID
	Name      string `json:"name"`      // 名称
	SortOrder int    `json:"sortOrder"` // 排序值
	Enabled   bool   `json:"enabled"`   // 是否启用
}

// Metadata 表示元数据响应数据。
type Metadata struct {
	DishCategories  []DishCategory `json:"dishCategories"`  // 菜品分类列表
	DishTags        []DishTag      `json:"dishTags"`        // 菜品标签列表
	IngredientUnits []string       `json:"ingredientUnits"` // 食材单位列表
	RuntimeConfig   RuntimeConfig  `json:"runtimeConfig"`   // 运行时配置
}

// ImageUploadResult 表示图片上传结果响应数据。
type ImageUploadResult struct {
	FileID       string  `json:"fileId"`       // 文件ID
	URL          string  `json:"url"`          // 地址
	Width        int     `json:"width"`        // 宽度
	Height       int     `json:"height"`       // 高度
	ReviewStatus string  `json:"reviewStatus"` // 审核状态
	RejectReason *string `json:"rejectReason"` // 驳回原因
}

// Ingredient 表示食材响应数据。
type Ingredient struct {
	ID        string  `json:"id"`        // ID
	Name      string  `json:"name"`      // 名称
	Amount    string  `json:"amount"`    // 数量
	Unit      string  `json:"unit"`      // 单位
	Note      *string `json:"note"`      // 备注
	SortOrder int     `json:"sortOrder"` // 排序值
}

// DishStep 表示菜品步骤响应数据。
type DishStep struct {
	ID        string  `json:"id"`        // ID
	Text      string  `json:"text"`      // 文本
	ImageURL  *string `json:"imageUrl"`  // 图片地址
	SortOrder int     `json:"sortOrder"` // 排序值
}

// DishSummary 表示菜品摘要响应数据。
type DishSummary struct {
	ID           string   `json:"id"`           // ID
	Name         string   `json:"name"`         // 名称
	Category     string   `json:"category"`     // 分类
	Tags         []string `json:"tags"`         // 标签列表
	CoverURL     string   `json:"coverUrl"`     // 封面地址
	Serving      int      `json:"serving"`      // 份数
	Status       string   `json:"status"`       // 状态
	Discoverable bool     `json:"discoverable"` // 菜品是否允许被发现
	SourceLocked bool     `json:"sourceLocked"` // 来源字段是否永久锁定
}

// Dish 表示菜品响应数据。
type Dish struct {
	DishSummary              // 菜品摘要
	Description *string      `json:"description"` // 说明
	Ingredients []Ingredient `json:"ingredients"` // 食材列表
	Steps       []DishStep   `json:"steps"`       // 步骤列表
	CreatedAt   time.Time    `json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time    `json:"updatedAt"`   // 更新时间
}

// AuthorSummary 表示作者摘要响应数据。
type AuthorSummary struct {
	ID        *string `json:"id"`        // ID
	Name      string  `json:"name"`      // 名称
	AvatarURL *string `json:"avatarUrl"` // 头像地址
	Type      string  `json:"type"`      // 类型
}

// RecommendationSummary 表示推荐摘要响应数据。
type RecommendationSummary struct {
	DishSummary                    // 菜品摘要
	RecommendationID string        `json:"recommendationId"` // 推荐ID
	SourceType       string        `json:"sourceType"`       // 来源类型
	Author           AuthorSummary `json:"author"`           // 作者
	Copied           bool          `json:"copied"`           // 是否已复制
}

// Recommendation 表示推荐响应数据。
type Recommendation struct {
	RecommendationSummary              // 推荐菜摘要
	Description           string       `json:"description"` // 说明
	Ingredients           []Ingredient `json:"ingredients"` // 食材列表
	Steps                 []DishStep   `json:"steps"`       // 步骤列表
}

// RecipeSummary 表示菜谱摘要响应数据。
type RecipeSummary struct {
	ID        string  `json:"id"`        // ID
	Name      string  `json:"name"`      // 名称
	Note      *string `json:"note"`      // 备注
	DishCount int     `json:"dishCount"` // 菜品数量
	CoverURL  *string `json:"coverUrl"`  // 封面地址
}

// Recipe 表示菜谱响应数据。
type Recipe struct {
	RecipeSummary               // 菜谱摘要
	DishIDs       []string      `json:"dishIds"`   // 菜品ID列表
	Dishes        []DishSummary `json:"dishes"`    // 菜品列表
	UpdatedAt     time.Time     `json:"updatedAt"` // 更新时间
}

// Checkin 表示打卡响应数据。
type Checkin struct {
	ID        string    `json:"id"`        // ID
	DishName  string    `json:"dishName"`  // 菜品名称
	ImageURL  string    `json:"imageUrl"`  // 图片地址
	Note      *string   `json:"note"`      // 备注
	CheckedAt time.Time `json:"checkedAt"` // 打卡时间
	Rewarded  bool      `json:"rewarded"`  // 是否已奖励
}

// MealSummary 表示饭局摘要响应数据。
type MealSummary struct {
	ID               string     `json:"id"`               // ID
	Name             string     `json:"name"`             // 名称
	Code             string     `json:"code"`             // 编码
	Status           string     `json:"status"`           // 状态
	CloseReason      *string    `json:"closeReason"`      // 关闭原因
	CancelReason     *string    `json:"cancelReason"`     // 取消原因分类
	DeadlineAt       time.Time  `json:"deadlineAt"`       // 截止时间
	ParticipantCount int        `json:"participantCount"` // 参与者数量
	CandidateCount   int        `json:"candidateCount"`   // 候选菜数量
	CoverURL         *string    `json:"coverUrl"`         // 封面地址
	FinalDishCount   int        `json:"finalDishCount"`   // 最终菜品数量
	TotalServings    int        `json:"totalServings"`    // 总份数
	CreatedAt        time.Time  `json:"createdAt"`        // 创建时间
	ConfirmedAt      *time.Time `json:"confirmedAt"`      // 已确认时间
	CompletedAt      *time.Time `json:"completedAt"`      // 是否完成时间
	CancelledAt      *time.Time `json:"cancelledAt"`      // 取消时间
}

// MealJoinPreview 表示饭局加入预览响应数据。
type MealJoinPreview struct {
	ID               string    `json:"id"`               // ID
	Name             string    `json:"name"`             // 名称
	Status           string    `json:"status"`           // 状态
	DeadlineAt       time.Time `json:"deadlineAt"`       // 截止时间
	ParticipantCount int       `json:"participantCount"` // 参与者数量
	CandidateCount   int       `json:"candidateCount"`   // 候选菜数量
}

// MealCandidate 表示饭局候选菜响应数据。
type MealCandidate struct {
	ID                string   `json:"id"`                // ID
	DishID            string   `json:"dishId"`            // 菜品ID
	Name              string   `json:"name"`              // 名称
	CoverURL          string   `json:"coverUrl"`          // 封面地址
	Category          string   `json:"category"`          // 分类
	Tags              []string `json:"tags"`              // 标签列表
	VoteCount         int      `json:"voteCount"`         // 点选数量
	SelectedByMe      bool     `json:"selectedByMe"`      // 当前用户是否已点选
	Available         bool     `json:"available"`         // 是否可用
	UnavailableReason *string  `json:"unavailableReason"` // 不可用原因
	SortOrder         int      `json:"sortOrder"`         // 排序值
}

// MealFinalDishSnapshot 表示饭局最终菜品快照响应数据。
type MealFinalDishSnapshot struct {
	CandidateID  string  `json:"candidateId"`   // 候选菜ID
	DishID       string  `json:"dishId"`        // 菜品ID
	Name         string  `json:"name"`          // 名称
	CoverURL     *string `json:"coverUrl"`      // 封面地址
	FinalServing int     `json:"finalServings"` // 最终份数
}

// Meal 表示饭局响应数据。
type Meal struct {
	MealSummary                                             // 饭局摘要
	CandidateIDs                    []string                `json:"candidateIds"`                    // 候选菜ID列表
	Candidates                      []MealCandidate         `json:"candidates"`                      // 候选菜列表
	FinalDishes                     []MealFinalDishSnapshot `json:"finalDishes"`                     // 最终菜品列表
	ShoppingListID                  *string                 `json:"shoppingListId"`                  // 采购清单列表ID
	CreatedByMe                     bool                    `json:"createdByMe"`                     // 是否由当前用户创建
	FinalResultSubscriptionAccepted bool                    `json:"finalResultSubscriptionAccepted"` // 当前用户是否已接受本饭局最终结果提醒
	ReadOnly                        bool                    `json:"readOnly"`                        // 当前用户是否只能查看保留内容
	ReadOnlyReason                  *string                 `json:"readOnlyReason"`                  // 只读原因
	CancelledFromStatus             *string                 `json:"cancelledFromStatus"`             // 取消前饭局状态
}

// MealDishStat 表示饭局菜品统计响应数据。
type MealDishStat struct {
	CandidateID      string   `json:"candidateId"`       // 候选菜ID
	DishID           string   `json:"dishId"`            // 菜品ID
	Name             string   `json:"name"`              // 名称
	CoverURL         string   `json:"coverUrl"`          // 封面地址
	Category         string   `json:"category"`          // 分类
	Tags             []string `json:"tags"`              // 标签列表
	VoteCount        int      `json:"voteCount"`         // 点选数量
	BaseServing      int      `json:"baseServing"`       // 基础份数
	SuggestedServing int      `json:"suggestedServings"` // 建议份数
	Selected         bool     `json:"selected"`          // 是否选择
	FinalServing     int      `json:"finalServings"`     // 最终份数
}

// CategoryCount 表示分类数量响应数据。
type CategoryCount struct {
	Category string `json:"category"` // 分类
	Count    int    `json:"count"`    // 数量
}

// ShoppingItem 表示采购清单项目响应数据。
type ShoppingItem struct {
	ID        string  `json:"id"`        // ID
	Name      string  `json:"name"`      // 名称
	Amount    string  `json:"amount"`    // 数量
	Note      *string `json:"note"`      // 备注
	Completed bool    `json:"completed"` // 是否完成
	SortOrder int     `json:"sortOrder"` // 排序值
}

// ShoppingList 表示采购清单列表响应数据。
type ShoppingList struct {
	ID             string         `json:"id"`             // ID
	ShareToken     string         `json:"shareToken"`     // 分享Token
	Meal           MealSummary    `json:"meal"`           // 饭局
	Items          []ShoppingItem `json:"items"`          // 项目列表
	PendingCount   int            `json:"pendingCount"`   // 待处理数量
	CompletedCount int            `json:"completedCount"` // 已购买数量
	ReadOnly       bool           `json:"readOnly"`       // 是否只读
}

// SharedShoppingList 表示分享采购清单列表响应数据。
type SharedShoppingList struct {
	ID             string         `json:"id"`             // ID
	Meal           MealSummary    `json:"meal"`           // 饭局
	Items          []ShoppingItem `json:"items"`          // 项目列表
	PendingCount   int            `json:"pendingCount"`   // 待处理数量
	CompletedCount int            `json:"completedCount"` // 已购买数量
	ReadOnly       bool           `json:"readOnly"`       // 是否只读
}

// FeatureUsageResult 表示功能调用记录结果响应数据。
type FeatureUsageResult struct {
	UsageID       string  `json:"usageId"`       // 调用记录ID
	PointCost     int     `json:"pointCost"`     // 积分成本
	BillingStatus string  `json:"billingStatus"` // 计费状态
	PointBalance  int64   `json:"pointBalance"`  // 积分余额
	UserMessage   *string `json:"userMessage"`   // 面向用户的计费状态说明
}

// MealSuggestion 表示饭局推荐菜响应数据。
type MealSuggestion struct {
	ID                 string             `json:"id"`                 // ID
	Source             string             `json:"source"`             // 来源
	SourceLabel        string             `json:"sourceLabel"`        // 来源显示名称
	Reason             string             `json:"reason"`             // 原因
	People             int                `json:"people"`             // 用餐人数
	SuggestedDishCount int                `json:"suggestedDishCount"` // 建议菜品数量
	Dish               SuggestionDish     `json:"dish"`               // 菜品
	Dishes             []SuggestionDish   `json:"dishes"`             // 菜品列表
	Usage              FeatureUsageResult `json:"usage"`              // 调用记录
}

// SuggestionDish 表示推荐菜菜品响应数据。
type SuggestionDish struct {
	ID               string       `json:"id"`               // ID
	Source           string       `json:"source"`           // 来源
	SourceLabel      string       `json:"sourceLabel"`      // 来源显示名称
	RecommendationID *string      `json:"recommendationId"` // 推荐ID
	DishID           *string      `json:"dishId"`           // 菜品ID
	Name             string       `json:"name"`             // 名称
	Category         string       `json:"category"`         // 分类
	Cuisine          string       `json:"cuisine"`          // 菜系
	Tags             []string     `json:"tags"`             // 标签列表
	CoverURL         string       `json:"coverUrl"`         // 封面地址
	Serving          int          `json:"serving"`          // 份数
	Description      string       `json:"description"`      // 说明
	Ingredients      []Ingredient `json:"ingredients"`      // 食材列表
	Steps            []DishStep   `json:"steps"`            // 步骤列表
	Copied           bool         `json:"copied"`           // 是否已复制
}

// PrepStep 表示备菜步骤响应数据。
type PrepStep struct {
	ID              string   `json:"id"`              // ID
	Order           int      `json:"order"`           // 顺序
	Title           string   `json:"title"`           // 标题
	Instruction     string   `json:"instruction"`     // 操作说明
	ParallelActions []string `json:"parallelActions"` // 可并行执行的动作
	WaitMinutes     *int     `json:"waitMinutes"`     // 等待分钟
	CompletionHint  string   `json:"completionHint"`  // 步骤完成判断提示
}

// PrepPlan 表示备菜计划响应数据。
type PrepPlan struct {
	ID               string             `json:"id"`               // ID
	MealID           string             `json:"mealId"`           // 饭局ID
	EstimatedMinutes int                `json:"estimatedMinutes"` // 预计分钟
	Steps            []PrepStep         `json:"steps"`            // 步骤列表
	Usage            FeatureUsageResult `json:"usage"`            // 调用记录
}

// PointEntry 表示积分流水响应数据。
type PointEntry struct {
	ID             string    `json:"id"`             // ID
	Type           string    `json:"type"`           // 类型
	Title          string    `json:"title"`          // 标题
	Description    string    `json:"description"`    // 说明
	Amount         int64     `json:"amount"`         // 数量
	RelatedEntryID *string   `json:"relatedEntryId"` // 关联流水ID
	CreatedAt      time.Time `json:"createdAt"`      // 创建时间
}

// FeatureUsage 表示功能调用记录响应数据。
type FeatureUsage struct {
	ID                 string    `json:"id"`                 // ID
	Feature            string    `json:"feature"`            // 功能
	PointCost          int       `json:"pointCost"`          // 积分成本
	ExecutionStatus    string    `json:"executionStatus"`    // 执行状态
	BillingStatus      string    `json:"billingStatus"`      // 计费状态
	RefundPointEntryID *string   `json:"refundPointEntryId"` // 退款积分流水ID
	CreatedAt          time.Time `json:"createdAt"`          // 创建时间
}

// Notification 表示通知响应数据。
type Notification struct {
	ID         string    `json:"id"`         // ID
	Type       string    `json:"type"`       // 类型
	Title      string    `json:"title"`      // 标题
	Content    string    `json:"content"`    // 内容
	TargetType *string   `json:"targetType"` // 目标类型
	TargetID   *string   `json:"targetId"`   // 目标ID
	Read       bool      `json:"read"`       // 是否已读
	CreatedAt  time.Time `json:"createdAt"`  // 创建时间
}

// BootstrapData 表示小程序启动所需的聚合响应数据。
type BootstrapData struct {
	Profile         Profile                 `json:"profile"`         // 用户资料
	RuntimeConfig   RuntimeConfig           `json:"runtimeConfig"`   // 运行时配置
	Dishes          []DishSummary           `json:"dishes"`          // 菜品列表
	Recommendations []RecommendationSummary `json:"recommendations"` // 推荐菜列表
	ActiveMeal      *MealSummary            `json:"activeMeal"`      // 当前进行中的饭局
	UnreadCount     int64                   `json:"unreadCount"`     // 未读数量
}

// DeletedResult 表示资源删除结果。
type DeletedResult struct {
	Deleted bool `json:"deleted"` // 是否已删除
}

// DishDeleteResult 表示菜品删除及关联影响结果。
type DishDeleteResult struct {
	Deleted             bool  `json:"deleted"`             // 是否已删除
	AffectedRecipeCount int64 `json:"affectedRecipeCount"` // 受影响菜谱数量
	AffectedMealCount   int64 `json:"affectedMealCount"`   // 受影响饭局数量
}

// RecommendationCopyResult 表示平台推荐菜复制结果。
type RecommendationCopyResult struct {
	Dish          Dish `json:"dish"`          // 复制生成的个人菜品
	AlreadyCopied bool `json:"alreadyCopied"` // 是否为此前已经复制的结果
}

// RecipeListResult 表示个人菜谱列表结果。
type RecipeListResult struct {
	List []RecipeSummary `json:"list"` // 菜谱列表
}

// CheckinCalendarResult 表示指定月份的做菜打卡日历。
type CheckinCalendarResult struct {
	Month        string   `json:"month"`        // 查询月份
	CheckedDates []string `json:"checkedDates"` // 已打卡日期列表
	CheckinDays  int      `json:"checkinDays"`  // 累计打卡天数
}

// CheckinCreateResult 表示做菜打卡保存与奖励结果。
type CheckinCreateResult struct {
	Checkin                Checkin `json:"checkin"`                // 新增的打卡记录
	RewardGranted          bool    `json:"rewardGranted"`          // 是否发放积分奖励
	RewardAmount           int64   `json:"rewardAmount"`           // 本次奖励积分
	PreferenceUpdateQueued bool    `json:"preferenceUpdateQueued"` // 是否已提交偏好画像更新任务
}

// PointsSummaryResult 表示当前积分及本月收支汇总。
type PointsSummaryResult struct {
	Balance     int64 `json:"balance"`     // 当前积分余额
	MonthEarned int64 `json:"monthEarned"` // 本月获得积分
	MonthSpent  int64 `json:"monthSpent"`  // 本月消耗积分
}

// NotificationSummaryResult 表示通知数量汇总。
type NotificationSummaryResult struct {
	UnreadCount int64 `json:"unreadCount"` // 未读通知数量
	Total       int64 `json:"total"`       // 通知总数
}

// NotificationReadResult 表示单条通知已读结果。
type NotificationReadResult struct {
	Read   bool      `json:"read"`   // 是否已读
	ReadAt time.Time `json:"readAt"` // 标记已读时间
}

// NotificationsReadAllResult 表示全部通知已读结果。
type NotificationsReadAllResult struct {
	Affected    int64 `json:"affected"`    // 本次更新的通知数量
	UnreadCount int64 `json:"unreadCount"` // 更新后的未读通知数量
}

// DishExtractionResult 表示菜品解析草稿及本次调用记录。
type DishExtractionResult struct {
	DishDraft frontRequest.DishDraftInput `json:"dishDraft"` // 可编辑的菜品草稿
	Usage     FeatureUsageResult          `json:"usage"`     // 本次增强功能调用记录
}

// DishCoverResult 表示菜品封面生成结果。
type DishCoverResult struct {
	FileID string             `json:"fileId"` // 生成图片文件ID
	URL    string             `json:"url"`    // 生成图片访问地址
	Usage  FeatureUsageResult `json:"usage"`  // 本次增强功能调用记录
}

// SuggestionStatusResult 表示推荐菜功能解锁、额度和积分状态。
type SuggestionStatusResult struct {
	Enabled      bool   `json:"enabled"`      // 推荐菜功能是否启用
	Unlocked     bool   `json:"unlocked"`     // 当前用户是否满足解锁条件
	CheckinDays  int    `json:"checkinDays"`  // 当前累计打卡天数
	UnlockDays   int    `json:"unlockDays"`   // 解锁所需打卡天数
	FreeQuota    int    `json:"freeQuota"`    // 当日剩余免费次数
	PointBalance *int64 `json:"pointBalance"` // 当前积分余额
	PointCost    *int   `json:"pointCost"`    // 超出免费次数后的积分消耗
}

// SuggestionCopyResult 表示推荐菜复制及采用记录结果。
type SuggestionCopyResult struct {
	Dish             Dish `json:"dish"`             // 复制生成的个人菜品
	AdoptionRecorded bool `json:"adoptionRecorded"` // 是否记录为已采用
}

// RecordedResult 表示反馈记录结果。
type RecordedResult struct {
	Recorded bool `json:"recorded"` // 是否已记录
}

// PrepQuoteResult 表示备菜顺序生成的积分报价。
type PrepQuoteResult struct {
	PointCost    int   `json:"pointCost"`    // 本次生成所需积分
	PointBalance int64 `json:"pointBalance"` // 当前积分余额
	CanGenerate  bool  `json:"canGenerate"`  // 当前积分是否足够生成
}

// CurrentMealResult 表示当前进行中饭局。
type CurrentMealResult struct {
	Meal *Meal `json:"meal"` // 当前饭局；没有进行中饭局时为 null
}

// MealCandidatesResult 表示饭局候选菜及分类统计。
type MealCandidatesResult struct {
	Categories []CategoryCount `json:"categories"` // 候选菜分类数量统计
	List       []MealCandidate `json:"list"`       // 候选菜列表
}

// MealCandidateRemovalResult 表示饭局候选菜移除结果。
type MealCandidateRemovalResult struct {
	CandidateID string    `json:"candidateId"` // 候选菜ID
	RemovedAt   time.Time `json:"removedAt"`   // 移除时间
}

// MealVotesResult 表示当前用户已保存的饭局点选结果。
type MealVotesResult struct {
	CandidateIDs []string   `json:"candidateIds"` // 已点选的候选菜ID列表
	UpdatedAt    *time.Time `json:"updatedAt"`    // 最近更新时间
}

// MealVotesSavedResult 表示饭局点选保存结果。
type MealVotesSavedResult struct {
	CandidateIDs []string  `json:"candidateIds"` // 已保存的候选菜ID列表
	SavedAt      time.Time `json:"savedAt"`      // 保存时间
}

// MealFinalResultSubscription 表示饭局最终结果微信订阅授权登记结果。
type MealFinalResultSubscription struct {
	MealID              string    `json:"mealId"`              // 饭局ID
	AuthorizationResult string    `json:"authorizationResult"` // 微信一次性订阅授权结果
	Accepted            bool      `json:"accepted"`            // 是否获得一次可用授权
	RecordedAt          time.Time `json:"recordedAt"`          // 服务端登记时间
}

// MealStatsResult 表示饭局候选菜点选统计。
type MealStatsResult struct {
	Meal   MealSummary    `json:"meal"`   // 饭局摘要
	Dishes []MealDishStat `json:"dishes"` // 候选菜点选统计列表
}

// MealConfirmResult 表示最终菜单确认及采购清单创建结果。
type MealConfirmResult struct {
	Meal           Meal   `json:"meal"`           // 已进入采购阶段的饭局
	ShoppingListID string `json:"shoppingListId"` // 生成的采购清单ID
}

// ShoppingExportResult 表示采购清单文本导出结果。
type ShoppingExportResult struct {
	Title     string    `json:"title"`     // 导出标题
	Text      string    `json:"text"`      // 可复制的采购清单文本
	UpdatedAt time.Time `json:"updatedAt"` // 采购清单更新时间
}

package request

// PageQuery 表示页码查询条件。
type PageQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`             // 页码
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量
}

// Defaults 补齐分页默认值。
func (query *PageQuery) Defaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
}

// DishListQuery 表示菜品列表查询条件。
type DishListQuery struct {
	PageQuery        // 分页参数
	Q         string `form:"q" binding:"omitempty,max=40" checksql:"false"` // 搜索词
	Category  string `form:"category" binding:"omitempty,max=64"`           // 分类
	Status    string `form:"status" binding:"omitempty,oneof=draft usable"` // 状态
}

// IngredientInput 表示食材输入参数。
type IngredientInput struct {
	Name      string  `json:"name" binding:"required,max=30" checksql:"false"`   // 名称
	Amount    string  `json:"amount" binding:"required,max=20" checksql:"false"` // 数量
	Unit      string  `json:"unit" binding:"required,max=10" checksql:"false"`   // 单位
	Note      *string `json:"note" binding:"omitempty,max=80" checksql:"false"`  // 备注
	SortOrder int     `json:"sortOrder" binding:"required,min=1"`                // 排序值
}

// DishStepInput 表示菜品步骤输入参数。
type DishStepInput struct {
	Text        string  `json:"text" binding:"required,max=500" checksql:"false"` // 文本
	ImageFileID *string `json:"imageFileId" binding:"omitempty,max=64"`           // 图片文件ID
	SortOrder   int     `json:"sortOrder" binding:"required,min=1"`               // 排序值
}

// DishUpsertInput 表示菜品保存输入参数。
type DishUpsertInput struct {
	Name        string            `json:"name" binding:"required,max=40" checksql:"false"`          // 名称
	Category    string            `json:"category" binding:"required,max=64"`                       // 分类
	Tags        []string          `json:"tags" binding:"max=3,dive,max=64"`                         // 标签列表
	Serving     int               `json:"serving" binding:"required,min=1,max=20"`                  // 份数
	Description *string           `json:"description" binding:"omitempty,max=180" checksql:"false"` // 说明
	CoverFileID string            `json:"coverFileId" binding:"required,max=64"`                    // 封面文件ID
	Status      string            `json:"status" binding:"required,oneof=draft usable"`             // 状态
	Ingredients []IngredientInput `json:"ingredients" binding:"dive"`                               // 食材列表
	Steps       []DishStepInput   `json:"steps" binding:"dive"`                                     // 步骤列表
}

// DishDraftInput 表示菜品草稿输入参数。
type DishDraftInput struct {
	Name        string            `json:"name"`                  // 名称
	Category    string            `json:"category"`              // 分类
	Tags        []string          `json:"tags"`                  // 标签列表
	Serving     int               `json:"serving"`               // 份数
	Description *string           `json:"description"`           // 说明
	CoverFileID *string           `json:"coverFileId,omitempty"` // 封面文件ID
	Status      string            `json:"status"`                // 状态
	Ingredients []IngredientInput `json:"ingredients"`           // 食材列表
	Steps       []DishStepInput   `json:"steps"`                 // 步骤列表
}

// DiscoverabilityInput 表示可发现状态输入参数。
type DiscoverabilityInput struct {
	Discoverable *bool `json:"discoverable" binding:"required"` // 菜品是否允许被发现
}

// DishExtractionInput 表示菜品解析输入参数。
type DishExtractionInput struct {
	Text        *string `json:"text" binding:"omitempty,max=4000" checksql:"false"` // 文本
	ImageFileID *string `json:"imageFileId" binding:"omitempty,max=64"`             // 图片文件ID
}

// DishCoverInput 表示菜品封面输入参数。
type DishCoverInput struct {
	Name        string  `json:"name" binding:"required,max=40" checksql:"false"`          // 名称
	Description *string `json:"description" binding:"omitempty,max=180" checksql:"false"` // 说明
	Category    *string `json:"category" binding:"omitempty,max=64"`                      // 分类
}

// RecommendationListQuery 表示推荐列表查询条件。
type RecommendationListQuery struct {
	PageQuery          // 分页参数
	Category  string   `form:"category" binding:"omitempty,max=64"` // 分类
	Tags      []string `form:"tags" binding:"dive,max=64"`          // 标签列表
}

// MealSuggestionInput 表示饭局推荐菜输入参数。
type MealSuggestionInput struct {
	Tags                      []string `json:"tags" binding:"max=10,dive,max=64"`                                       // 标签列表
	People                    int      `json:"people" binding:"required,min=1,max=20"`                                  // 用餐人数
	TimeLimitMinutes          *int     `json:"timeLimitMinutes" binding:"omitempty,min=1,max=1440"`                     // 时间限额分钟
	Preference                *string  `json:"preference" binding:"omitempty,max=120" checksql:"false"`                 // 用户本次明确选择的忌口或偏好
	ExcludedSuggestionDishIDs []string `json:"excludedSuggestionDishIds" binding:"max=50,unique,dive,max=64"`           // 排除推荐菜菜品ID列表
	ExcludedRecommendationIDs []string `json:"excludedRecommendationIds,omitempty" binding:"max=50,unique,dive,max=64"` // 排除推荐ID列表
	UseFreeQuota              bool     `json:"useFreeQuota"`                                                            // 使用免费额度
	PreferenceContext         string   `json:"-" swaggerignore:"true"`                                                  // 服务端注入的结构化偏好画像上下文
	HardConditions            []string `json:"-" swaggerignore:"true"`                                                  // 服务端注入的过敏忌口等硬条件
}

// FeedbackInput 表示反馈输入参数。
type FeedbackInput struct {
	Action string `json:"action" binding:"required,oneof=ignored regenerated"` // 操作
}

// SuggestionDishCopyInput 表示推荐菜菜品复制输入参数。
type SuggestionDishCopyInput struct {
	SuggestionDishID string `json:"suggestionDishId" binding:"required,max=64"` // 推荐菜菜品ID
}

// RecipeCreateInput 表示菜谱创建输入参数。
type RecipeCreateInput struct {
	Name    string   `json:"name" binding:"required,max=30" checksql:"false"`   // 名称
	Note    *string  `json:"note" binding:"omitempty,max=100" checksql:"false"` // 备注
	DishIDs []string `json:"dishIds" binding:"dive,max=64"`                     // 菜品ID列表
}

// RecipeUpdateInput 表示菜谱更新输入参数。
type RecipeUpdateInput struct {
	Name    *string   `json:"name" binding:"omitempty,max=30" checksql:"false"`  // 名称
	Note    *string   `json:"note" binding:"omitempty,max=100" checksql:"false"` // 备注
	DishIDs *[]string `json:"dishIds" binding:"omitempty,dive,max=64"`           // 菜品ID列表
}

// RecipeDishesInput 表示菜谱菜品列表输入参数。
type RecipeDishesInput struct {
	DishIDs []string `json:"dishIds" binding:"required,min=1,dive,max=64"` // 菜品ID列表
}

// CheckinCalendarQuery 表示打卡日历查询条件。
type CheckinCalendarQuery struct {
	Month string `form:"month" binding:"required,datetime=2006-01"` // 月份
}

// CheckinInput 表示打卡输入参数。
type CheckinInput struct {
	DishName    string  `json:"dishName" binding:"required,max=40" checksql:"false"` // 菜品名称
	ImageFileID string  `json:"imageFileId" binding:"required,max=64"`               // 图片文件ID
	Note        *string `json:"note" binding:"omitempty,max=120" checksql:"false"`   // 备注
}

// MealListQuery 表示饭局列表查询条件。
type MealListQuery struct {
	PageQuery        // 分页参数
	Scope     string `form:"scope" binding:"required,oneof=active history"` // 范围
}

// MealCreateInput 表示饭局创建输入参数。
type MealCreateInput struct {
	Name             string   `json:"name" binding:"required,max=30" checksql:"false"`                  // 名称
	DeadlineAt       string   `json:"deadlineAt" binding:"required,datetime=2006-01-02T15:04:05Z07:00"` // 截止时间
	CandidateDishIDs []string `json:"candidateDishIds" binding:"required,min=1,unique,dive,max=64"`     // 候选菜菜品ID列表
	SourceRecipeID   *string  `json:"sourceRecipeId" binding:"omitempty,max=64"`                        // 来源菜谱ID
}

// MealCodeInput 表示饭局编码输入参数。
type MealCodeInput struct {
	Code string `json:"code" binding:"required,len=6,alphanum"` // 编码
}

// MealCancelInput 表示饭局取消输入参数。
type MealCancelInput struct {
	Reason *string `json:"reason" binding:"omitempty,max=100" checksql:"false"` // 原因
}

// MealCandidatesQuery 表示饭局候选菜列表查询条件。
type MealCandidatesQuery struct {
	Category string `form:"category" binding:"omitempty,max=64"` // 分类
}

// MealCandidatePath 表示饭局候选菜路径参数。
type MealCandidatePath struct {
	MealID      string `uri:"mealId" binding:"required,max=64"`      // 饭局ID
	CandidateID string `uri:"candidateId" binding:"required,max=64"` // 候选菜ID
}

// MealVoteInput 表示饭局点选输入参数。
type MealVoteInput struct {
	CandidateIDs []string `json:"candidateIds" binding:"required,unique,dive,max=64"` // 候选菜ID列表
}

// MealFinalResultSubscriptionInput 表示饭局最终结果微信订阅授权登记参数。
type MealFinalResultSubscriptionInput struct {
	AuthorizationResult string `json:"authorizationResult" binding:"required,oneof=accept reject ban"` // 微信一次性订阅授权结果
	TemplateID          string `json:"templateId" binding:"required,max=64"`                           // 本次调用微信授权使用的模板ID
}

// FinalMenuDishInput 表示最终菜单菜品输入参数。
type FinalMenuDishInput struct {
	CandidateID  string `json:"candidateId" binding:"required,max=64"` // 候选菜ID
	Selected     *bool  `json:"selected" binding:"required"`           // 是否选择
	FinalServing int    `json:"finalServings" binding:"min=0,max=20"`  // 最终份数
}

// MealConfirmInput 表示饭局确认输入参数。
type MealConfirmInput struct {
	Dishes []FinalMenuDishInput `json:"dishes" binding:"required,min=1,dive"` // 菜品列表
}

// ShoppingItemCreateInput 表示采购清单项目创建输入参数。
type ShoppingItemCreateInput struct {
	Name   string  `json:"name" binding:"required,max=40" checksql:"false"`   // 名称
	Amount string  `json:"amount" binding:"required,max=30" checksql:"false"` // 数量
	Note   *string `json:"note" binding:"omitempty,max=100" checksql:"false"` // 备注
}

// ShoppingItemUpdateInput 表示采购清单项目更新输入参数。
type ShoppingItemUpdateInput struct {
	Name      *string `json:"name" binding:"omitempty,max=40" checksql:"false"`   // 名称
	Amount    *string `json:"amount" binding:"omitempty,max=30" checksql:"false"` // 数量
	Note      *string `json:"note" binding:"omitempty,max=100" checksql:"false"`  // 备注
	Completed *bool   `json:"completed"`                                          // 是否完成
	SortOrder *int    `json:"sortOrder" binding:"omitempty,min=0"`                // 排序值
}

// PrepPlanQuoteQuery 表示备菜计划报价查询条件。
type PrepPlanQuoteQuery struct {
	MealID string `form:"mealId" binding:"required,max=64"` // 饭局ID
}

// PrepPlanInput 表示备菜计划输入参数。
type PrepPlanInput struct {
	MealID          string   `json:"mealId" binding:"required,max=64"`      // 饭局ID
	ExcludedPlanIDs []string `json:"excludedPlanIds" binding:"dive,max=64"` // 排除计划ID列表
}

// PointEntriesQuery 表示积分流水列表查询条件。
type PointEntriesQuery struct {
	PageQuery        // 分页参数
	Type      string `form:"type" binding:"omitempty,oneof=all earned spent refund"` // 类型
}

// FeatureUsageQuery 表示功能调用记录查询条件。
type FeatureUsageQuery struct {
	PageQuery              // 分页参数
	Feature         string `form:"feature" binding:"omitempty,max=64"`                                            // 功能
	ExecutionStatus string `form:"executionStatus" binding:"omitempty,oneof=pending processing succeeded failed"` // 执行状态
	BillingStatus   string `form:"billingStatus" binding:"omitempty,oneof=not_charged charged refunded"`          // 计费状态
}

// NotificationListQuery 表示通知列表查询条件。
type NotificationListQuery struct {
	PageQuery        // 分页参数
	Scope     string `form:"scope" binding:"omitempty,oneof=unread all"` // 范围
}

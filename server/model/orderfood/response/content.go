package response

import "time"

// Page 表示分页响应数据。
type Page[T any] struct {
	Page     int   `json:"page"`     // 页码
	PageSize int   `json:"pageSize"` // 每页数量
	Total    int64 `json:"total"`    // 总数量
	List     []T   `json:"list"`     // 列表
}

// AuditedPage 表示包含访问审计记录ID的分页响应数据。
type AuditedPage[T any] struct {
	Page          int    `json:"page"`          // 页码
	PageSize      int    `json:"pageSize"`      // 每页数量
	Total         int64  `json:"total"`         // 总数量
	List          []T    `json:"list"`          // 列表
	AccessAuditID string `json:"accessAuditId"` // 敏感访问审计记录ID
}

// UserReference 表示用户引用响应数据。
type UserReference struct {
	ID        string  `json:"id"`        // ID
	Nickname  string  `json:"nickname"`  // 昵称
	AvatarURL *string `json:"avatarUrl"` // 头像地址
	Status    string  `json:"status"`    // 状态
}

// AdministratorSummary 表示管理员摘要响应数据。
type AdministratorSummary struct {
	ID       string  `json:"id"`       // ID
	Username string  `json:"username"` // 用户名
	Nickname *string `json:"nickname"` // 昵称
}

// CatalogReference 表示索引引用响应数据。
type CatalogReference struct {
	ID      string `json:"id"`      // ID
	Name    string `json:"name"`    // 名称
	Enabled bool   `json:"enabled"` // 是否启用
}

// DishIngredient 表示菜品食材响应数据。
type DishIngredient struct {
	Name      string            `json:"name"`      // 名称
	Quantity  *string           `json:"quantity"`  // 用量
	Unit      *CatalogReference `json:"unit"`      // 单位
	Note      *string           `json:"note"`      // 备注
	SortOrder int               `json:"sortOrder"` // 排序值
}

// DishStep 表示菜品步骤响应数据。
type DishStep struct {
	Description string  `json:"description"` // 说明
	ImageURL    *string `json:"imageUrl"`    // 图片地址
	SortOrder   int     `json:"sortOrder"`   // 排序值
}

// UserDishAdminSummary 表示用户菜品管理员摘要响应数据。
type UserDishAdminSummary struct {
	ID                            string             `json:"id"`                            // ID
	CoverURL                      *string            `json:"coverUrl"`                      // 封面地址
	Name                          string             `json:"name"`                          // 名称
	Owner                         UserReference      `json:"owner"`                         // 所有者
	Category                      CatalogReference   `json:"category"`                      // 分类
	Tags                          []CatalogReference `json:"tags"`                          // 标签列表
	Status                        string             `json:"status"`                        // 状态
	DeletionStatus                string             `json:"deletionStatus"`                // 删除状态
	Discoverable                  bool               `json:"discoverable"`                  // 菜品是否允许被发现
	DiscoverableAt                *time.Time         `json:"discoverableAt"`                // 是否允许被发现时间
	SourceType                    string             `json:"sourceType"`                    // 来源类型
	SourceLocked                  bool               `json:"sourceLocked"`                  // 来源字段是否永久锁定
	RecipeReferenceCount          int                `json:"recipeReferenceCount"`          // 菜谱引用数量
	UnconfirmedMealReferenceCount int                `json:"unconfirmedMealReferenceCount"` // 未确认饭局引用数量
	MediaReviewStatus             string             `json:"mediaReviewStatus"`             // 图片审核状态
	RecommendationCount           int                `json:"recommendationCount"`           // 推荐数量
	Version                       int                `json:"version"`                       // 版本
	CreatedAt                     time.Time          `json:"createdAt"`                     // 创建时间
	UpdatedAt                     time.Time          `json:"updatedAt"`                     // 更新时间
	DeletedAt                     *time.Time         `json:"deletedAt"`                     // 删除时间
}

// UserDishSourceNode 表示用户菜品来源节点响应数据。
type UserDishSourceNode struct {
	DishID         string         `json:"dishId"`         // 菜品ID
	DishType       string         `json:"dishType"`       // 菜品类型
	Name           string         `json:"name"`           // 名称
	Owner          *UserReference `json:"owner"`          // 所有者
	DeletionStatus string         `json:"deletionStatus"` // 删除状态
	Accessible     bool           `json:"accessible"`     // 是否可访问
}

// UserDishSourceOverview 表示用户菜品来源概览响应数据。
type UserDishSourceOverview struct {
	SourceType          string              `json:"sourceType"`          // 来源类型
	OperationSourceType *string             `json:"operationSourceType"` // 操作来源类型
	OperationSourceID   *string             `json:"operationSourceId"`   // 操作来源ID
	DirectSource        *UserDishSourceNode `json:"directSource"`        // 直接来源
	RootSource          *UserDishSourceNode `json:"rootSource"`          // 根来源
	OriginalAuthor      *UserReference      `json:"originalAuthor"`      // 原始作者
	ChainDepth          int                 `json:"chainDepth"`          // 复制链深度
	DirectCopyCount     int                 `json:"directCopyCount"`     // 直接复制数量
	DescendantCopyCount int                 `json:"descendantCopyCount"` // 下游复制数量
}

// UserDishReferenceOverview 表示用户菜品引用概览响应数据。
type UserDishReferenceOverview struct {
	ActiveRecipeCount          int `json:"activeRecipeCount"`          // 当前菜谱数量
	UnconfirmedMealCount       int `json:"unconfirmedMealCount"`       // 未确认饭局数量
	ConfirmedMealSnapshotCount int `json:"confirmedMealSnapshotCount"` // 已确认饭局快照数量
	ShoppingSnapshotCount      int `json:"shoppingSnapshotCount"`      // 采购清单快照数量
}

// ModerationRecordSummary 表示图片审核记录摘要响应数据。
type ModerationRecordSummary struct {
	ID                string         `json:"id"`                // ID
	RequestID         string         `json:"requestId"`         // 请求ID
	FileID            string         `json:"fileId"`            // 文件ID
	User              *UserReference `json:"user"`              // 上传用户摘要
	ObjectType        *string        `json:"objectType"`        // 关联对象类型
	ObjectID          *string        `json:"objectId"`          // 关联对象ID
	Status            string         `json:"status"`            // 状态
	RiskLabels        []string       `json:"riskLabels"`        // 供应商返回的风险标签
	RiskLevel         *string        `json:"riskLevel"`         // 风险等级
	DurationMS        *int           `json:"durationMs"`        // 耗时（毫秒）
	ProviderRequestID *string        `json:"providerRequestId"` // 供应商请求ID
	ErrorCode         *string        `json:"errorCode"`         // 错误编码
	CreatedAt         time.Time      `json:"createdAt"`         // 创建时间
}

// UserDishAdminDetail 表示用户菜品管理员详情响应数据。
type UserDishAdminDetail struct {
	UserDishAdminSummary                            // 用户菜品管理摘要
	CoverFileID           string                    `json:"coverFileId"`           // 封面文件ID
	Description           *string                   `json:"description"`           // 说明
	Serving               int                       `json:"serving"`               // 份数
	Ingredients           []DishIngredient          `json:"ingredients"`           // 食材列表
	Steps                 []DishStep                `json:"steps"`                 // 步骤列表
	SourceOverview        UserDishSourceOverview    `json:"sourceOverview"`        // 来源概览
	ReferenceOverview     UserDishReferenceOverview `json:"referenceOverview"`     // 引用概览
	ModerationSummary     *ModerationRecordSummary  `json:"moderationSummary"`     // 图片审核摘要
	DeletedReason         *string                   `json:"deletedReason"`         // 删除原因
	DeletedViolationType  *string                   `json:"deletedViolationType"`  // 删除时记录的违规类型
	DeletedBy             *AdministratorSummary     `json:"deletedBy"`             // 删除人
	GovernanceRecordCount int                       `json:"governanceRecordCount"` // 违规处理记录数量
	AuditLogCount         int                       `json:"auditLogCount"`         // 审计日志数量
	AccessAuditID         string                    `json:"accessAuditId"`         // 敏感访问审计记录ID
}

// UserDishReference 表示用户菜品引用响应数据。
type UserDishReference struct {
	ID                 string    `json:"id"`                 // ID
	ReferenceType      string    `json:"referenceType"`      // 引用类型
	ObjectID           string    `json:"objectId"`           // 业务对象ID
	ObjectLabel        string    `json:"objectLabel"`        // 业务对象显示名称
	ObjectStatus       string    `json:"objectStatus"`       // 业务对象状态
	HistoricalSnapshot bool      `json:"historicalSnapshot"` // 是否为历史快照
	OccurredAt         time.Time `json:"occurredAt"`         // 发生时间
}

// UserRecipeAdminSummary 表示用户菜谱管理员摘要响应数据。
type UserRecipeAdminSummary struct {
	ID                   string        `json:"id"`                   // ID
	Name                 string        `json:"name"`                 // 名称
	CoverURL             *string       `json:"coverUrl"`             // 封面地址
	Owner                UserReference `json:"owner"`                // 所有者
	DishCount            int           `json:"dishCount"`            // 菜品数量
	UnavailableDishCount int           `json:"unavailableDishCount"` // 不可用菜品数量
	HasNote              bool          `json:"hasNote"`              // 是否有备注
	DeletionStatus       string        `json:"deletionStatus"`       // 删除状态
	ContentState         string        `json:"contentState"`         // 内容状态
	Version              int           `json:"version"`              // 版本
	CreatedAt            time.Time     `json:"createdAt"`            // 创建时间
	UpdatedAt            time.Time     `json:"updatedAt"`            // 更新时间
	DeletedAt            *time.Time    `json:"deletedAt"`            // 删除时间
}

// UserRecipeDishItem 表示用户菜谱菜品项目响应数据。
type UserRecipeDishItem struct {
	RelationID        string    `json:"relationId"`        // 关系ID
	DishID            string    `json:"dishId"`            // 菜品ID
	DishName          string    `json:"dishName"`          // 菜品名称
	CoverURL          *string   `json:"coverUrl"`          // 封面地址
	DishStatus        string    `json:"dishStatus"`        // 菜品状态
	SourceType        string    `json:"sourceType"`        // 来源类型
	UnavailableReason *string   `json:"unavailableReason"` // 不可用原因
	SortOrder         int       `json:"sortOrder"`         // 排序值
	AddedAt           time.Time `json:"addedAt"`           // 添加时间
}

// UserRecipeAdminDetail 表示用户菜谱管理员详情响应数据。
type UserRecipeAdminDetail struct {
	UserRecipeAdminSummary                       // 用户菜谱管理摘要
	Note                   *string               `json:"note"`                  // 备注
	Dishes                 []UserRecipeDishItem  `json:"dishes"`                // 菜品列表
	DeletedReason          *string               `json:"deletedReason"`         // 删除原因
	DeletedViolationType   *string               `json:"deletedViolationType"`  // 删除时记录的违规类型
	DeletedBy              *AdministratorSummary `json:"deletedBy"`             // 删除人
	GovernanceRecordCount  int                   `json:"governanceRecordCount"` // 违规处理记录数量
	AuditLogCount          int                   `json:"auditLogCount"`         // 审计日志数量
	AccessAuditID          string                `json:"accessAuditId"`         // 敏感访问审计记录ID
}

// GovernanceImpactPreview 表示违规处理影响预览响应数据。
type GovernanceImpactPreview struct {
	PreviewToken          string    `json:"previewToken"`          // 预览Token
	ExpiresAt             time.Time `json:"expiresAt"`             // 过期时间
	TargetType            string    `json:"targetType"`            // 实际处理的源头类型
	TargetID              string    `json:"targetId"`              // 实际处理的源头ID
	EntryRecommendationID *string   `json:"entryRecommendationId"` // 发起处理的推荐ID
	RecommendationCount   int       `json:"recommendationCount"`   // 将联动下线的推荐数量
	SourceDishCount       int       `json:"sourceDishCount"`       // 来源菜品数量
	CopiedDishCount       int       `json:"copiedDishCount"`       // 已复制菜品数量
	AffectedUserCount     int       `json:"affectedUserCount"`     // 受影响用户数量
	NotificationCount     int       `json:"notificationCount"`     // 通知数量
	Asynchronous          bool      `json:"asynchronous"`          // 是否异步
	Warnings              []string  `json:"warnings"`              // 警告列表
	ExpectedVersion       int       `json:"expectedVersion"`       // 预期版本
}

// GovernanceExecutionResult 表示违规处理执行结果响应数据。
type GovernanceExecutionResult struct {
	RecordID      string  `json:"recordId"`      // 记录ID
	JobID         *string `json:"jobId"`         // 任务ID
	Status        string  `json:"status"`        // 状态
	AffectedCount int     `json:"affectedCount"` // 受影响数量
}

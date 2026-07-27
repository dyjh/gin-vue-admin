package model

import (
	"time"

	"github.com/dyjh/order-food-mini-app/server/global"
	"gorm.io/datatypes"
)

const (
	DishStatusDraft  = "draft"
	DishStatusUsable = "usable"

	ContentStatusNormal   = "normal"
	ContentStatusDisabled = "disabled"

	SourceTypeManual              = "manual"
	SourceTypeCreatorCopy         = "creator_copy"
	SourceTypeOfficialCopy        = "official_copy"
	SourceTypeMealSuggestionCopy  = "meal_suggestion_copy"
	SourceTypeGeneratedSuggestion = "generated_suggestion"

	MediaReviewPending     = "pending"
	MediaReviewPassed      = "passed"
	MediaReviewRejected    = "rejected"
	MediaReviewFailed      = "failed"
	MediaReviewNotRequired = "not_required"

	ReferenceTypeRecipe                = "recipe"
	ReferenceTypeUnconfirmedMeal       = "unconfirmed_meal"
	ReferenceTypeConfirmedMealSnapshot = "confirmed_meal_snapshot"
	ReferenceTypeShoppingSnapshot      = "shopping_snapshot"

	GovernanceTargetDish         = "dish"
	GovernanceTargetOfficialDish = "official_dish"
	GovernanceTargetRecipe       = "recipe"
	GovernanceTargetCheckin      = "checkin"

	GovernanceActionDisableDiscoverability = "disable_discoverability"
	GovernanceActionSoftDeleteDish         = "soft_delete_dish"
	GovernanceActionSoftDeleteOfficialDish = "soft_delete_official_dish"
	GovernanceActionSoftDeleteRecipe       = "soft_delete_recipe"
	GovernanceActionSoftDeleteCheckin      = "soft_delete_checkin"
	GovernanceActionDeleteCopyChain        = "delete_copy_chain"

	GovernanceSeverityNormal  = "normal"
	GovernanceSeveritySerious = "serious"

	GovernanceJobPending            = "pending"
	GovernanceJobProcessing         = "processing"
	GovernanceJobPartiallySucceeded = "partially_succeeded"
	GovernanceJobSucceeded          = "succeeded"
	GovernanceJobFailed             = "failed"

	GovernanceJobItemPending   = "pending"
	GovernanceJobItemSucceeded = "succeeded"
	GovernanceJobItemFailed    = "failed"
)

// UsableMediaReviewStatuses 返回允许绑定到业务对象的图片审核状态。
func UsableMediaReviewStatuses() []string {
	return []string{MediaReviewPassed, MediaReviewNotRequired}
}

// IsMediaReviewUsable 判断图片是否已审核通过或按当前配置无需审核。
func IsMediaReviewUsable(status string) bool {
	return status == MediaReviewPassed || status == MediaReviewNotRequired
}

// ContentCategory 表示菜品分类。
type ContentCategory struct {
	global.GVA_MODEL        // GVA基础模型字段
	PublicID         string `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`         // 对外公开ID
	Name             string `json:"name" gorm:"column:name;type:varchar(40);uniqueIndex:idx_of_category_name;not null;comment:名称;"` // 名称
	SortOrder        int    `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`              // 排序值
	Enabled          bool   `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;default:true;comment:是否启用;"`     // 是否启用
	Version          int64  `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                     // 数据版本
}

// TableName 指定ContentCategory对应的数据表名。
func (ContentCategory) TableName() string { return "of_categories" }

// ContentTag 表示菜品标签。
type ContentTag struct {
	global.GVA_MODEL        // GVA基础模型字段
	PublicID         string `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`     // 对外公开ID
	Name             string `json:"name" gorm:"column:name;type:varchar(40);uniqueIndex:idx_of_tag_name;not null;comment:名称;"`  // 名称
	SortOrder        int    `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`          // 排序值
	Enabled          bool   `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;default:true;comment:是否启用;"` // 是否启用
	Version          int64  `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                 // 数据版本
}

// TableName 指定ContentTag对应的数据表名。
func (ContentTag) TableName() string { return "of_tags" }

// ContentUnit 表示食材单位。
type ContentUnit struct {
	global.GVA_MODEL        // GVA基础模型字段
	PublicID         string `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`     // 对外公开ID
	Name             string `json:"name" gorm:"column:name;type:varchar(40);uniqueIndex:idx_of_unit_name;not null;comment:名称;"` // 名称
	SortOrder        int    `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`          // 排序值
	Enabled          bool   `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;default:true;comment:是否启用;"` // 是否启用
	Version          int64  `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                 // 数据版本
}

// TableName 指定ContentUnit对应的数据表名。
func (ContentUnit) TableName() string { return "of_units" }

// UserDish 表示用户菜品。
type UserDish struct {
	global.GVA_MODEL                 // GVA基础模型字段
	PublicID         string          `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                           // 对外公开ID
	OwnerID          string          `json:"ownerId" gorm:"column:owner_id;type:varchar(64);not null;index;comment:所属用户ID;"`                                   // 所属用户ID
	Owner            MiniAppUser     `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`                                                                    // 所属用户关联数据
	CoverFileID      string          `json:"coverFileId" gorm:"column:cover_file_id;type:varchar(64);not null;comment:封面文件ID;"`                                // 封面文件ID
	CoverURL         *string         `json:"coverUrl" gorm:"column:cover_url;type:varchar(512);default:null;comment:封面地址;"`                                    // 封面地址
	Name             string          `json:"name" gorm:"column:name;type:varchar(120);not null;index;comment:名称;"`                                             // 名称
	CategoryID       uint            `json:"categoryId" gorm:"column:category_id;type:bigint unsigned;not null;index;comment:分类ID;"`                           // 分类ID
	Category         ContentCategory `json:"category" gorm:"foreignKey:CategoryID"`                                                                            // 分类关联数据
	Status           string          `json:"status" gorm:"column:status;type:varchar(20);not null;index;comment:状态;"`                                          // 状态
	Discoverable     bool            `json:"discoverable" gorm:"column:discoverable;type:tinyint(1) unsigned;not null;default:false;index;comment:是否允许被发现;"`   // 是否允许被发现
	DiscoverableAt   *time.Time      `json:"discoverableAt" gorm:"column:discoverable_at;type:datetime;default:null;comment:公开时间;"`                            // 公开时间
	SourceType       string          `json:"sourceType" gorm:"column:source_type;type:varchar(32);not null;index;comment:来源类型;"`                               // 来源类型
	SourceLocked     bool            `json:"sourceLocked" gorm:"column:source_locked;type:tinyint(1) unsigned;not null;default:false;index;comment:来源字段是否锁定;"` // 来源字段是否锁定
	Description      *string         `json:"description" gorm:"column:description;type:varchar(1000);default:null;comment:说明;"`                                // 说明
	Serving          int             `json:"serving" gorm:"column:serving;type:int;not null;default:1;comment:默认份数;"`                                          // 默认份数

	MediaReviewStatus string `json:"mediaReviewStatus" gorm:"column:media_review_status;type:varchar(24);not null;default:not_required;index;comment:媒体审核状态;"` // 媒体审核状态

	OperationSourceType        *string `json:"operationSourceType" gorm:"column:operation_source_type;type:varchar(32);default:null;comment:操作来源类型;"`                               // 操作来源类型
	OperationSourceID          *string `json:"operationSourceId" gorm:"column:operation_source_id;type:varchar(64);default:null;comment:操作来源ID;"`                                   // 操作来源ID
	DirectSourceDishID         *uint   `json:"directSourceDishId" gorm:"column:direct_source_dish_id;type:bigint unsigned;index;default:null;comment:直接来源菜品ID;"`                    // 直接来源菜品ID
	RootSourceDishID           *uint   `json:"rootSourceDishId" gorm:"column:root_source_dish_id;type:bigint unsigned;index;default:null;comment:根来源菜品ID;"`                         // 根来源菜品ID
	DirectSourceOfficialDishID *uint   `json:"directSourceOfficialDishId" gorm:"column:direct_source_official_dish_id;type:bigint unsigned;index;default:null;comment:直接来源官方菜品ID;"` // 直接来源官方菜品ID
	RootSourceOfficialDishID   *uint   `json:"rootSourceOfficialDishId" gorm:"column:root_source_official_dish_id;type:bigint unsigned;index;default:null;comment:根来源官方菜品ID;"`      // 根来源官方菜品ID
	OriginalAuthorID           *string `json:"originalAuthorId" gorm:"column:original_author_id;type:varchar(64);index;default:null;comment:原作者用户ID;"`                              // 原作者用户ID
	ChainDepth                 int     `json:"chainDepth" gorm:"column:chain_depth;type:int;not null;default:0;comment:复制链深度;"`                                                     // 复制链深度

	RecommendationCount int `json:"recommendationCount" gorm:"column:recommendation_count;type:int;not null;default:0;comment:推荐引用数;"` // 推荐引用数
	Version             int `json:"version" gorm:"column:version;type:int;not null;default:1;comment:数据版本;"`                           // 数据版本

	DeletedReason        *string `json:"deletedReason" gorm:"column:deleted_reason;type:varchar(600);default:null;comment:删除原因;"`                     // 删除原因
	DeletedViolationType *string `json:"deletedViolationType" gorm:"column:deleted_violation_type;type:varchar(80);default:null;comment:删除违规类型;"`     // 删除违规类型
	DeletedByAdminID     *uint   `json:"deletedByAdminId" gorm:"column:deleted_by_admin_id;type:bigint unsigned;index;default:null;comment:删除管理员ID;"` // 删除管理员ID
	DeletedByUsername    *string `json:"deletedByUsername" gorm:"column:deleted_by_username;type:varchar(120);default:null;comment:删除管理员用户名;"`        // 删除管理员用户名
	DeletedByNickname    *string `json:"deletedByNickname" gorm:"column:deleted_by_nickname;type:varchar(120);default:null;comment:删除管理员昵称;"`         // 删除管理员昵称

	Ingredients []DishIngredient `json:"ingredients" gorm:"foreignKey:DishID"`                                          // 食材列表
	Steps       []DishStep       `json:"steps" gorm:"foreignKey:DishID"`                                                // 步骤列表
	Tags        []ContentTag     `json:"tags" gorm:"many2many:of_dish_tags;joinForeignKey:DishID;joinReferences:TagID"` // 标签关联数据
}

// TableName 指定UserDish对应的数据表名。
func (UserDish) TableName() string { return "of_dishes" }

// DishIngredient 表示菜品食材。
type DishIngredient struct {
	global.GVA_MODEL              // GVA基础模型字段
	DishID           uint         `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`     // 菜品ID
	Name             string       `json:"name" gorm:"column:name;type:varchar(120);not null;comment:名称;"`                     // 名称
	Quantity         *string      `json:"quantity" gorm:"column:quantity;type:varchar(40);default:null;comment:用量;"`          // 用量
	UnitID           *uint        `json:"unitId" gorm:"column:unit_id;type:bigint unsigned;index;default:null;comment:单位ID;"` // 单位ID
	Unit             *ContentUnit `json:"unit" gorm:"foreignKey:UnitID"`                                                      // 单位关联数据
	Note             *string      `json:"note" gorm:"column:note;type:varchar(300);default:null;comment:备注;"`                 // 备注
	SortOrder        int          `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`  // 排序值
}

// TableName 指定DishIngredient对应的数据表名。
func (DishIngredient) TableName() string { return "of_dish_ingredients" }

// DishStep 表示菜品步骤。
type DishStep struct {
	global.GVA_MODEL         // GVA基础模型字段
	DishID           uint    `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`    // 菜品ID
	Description      string  `json:"description" gorm:"column:description;type:varchar(2000);not null;comment:说明;"`     // 说明
	ImageURL         *string `json:"imageUrl" gorm:"column:image_url;type:varchar(512);default:null;comment:图片地址;"`     // 图片地址
	SortOrder        int     `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"` // 排序值
}

// TableName 指定DishStep对应的数据表名。
func (DishStep) TableName() string { return "of_dish_steps" }

// UserDishTag 表示菜品与标签关联。
type UserDishTag struct {
	DishID    uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;primaryKey;not null;comment:菜品ID;"` // 菜品ID
	TagID     uint      `json:"tagId" gorm:"column:tag_id;type:bigint unsigned;primaryKey;not null;comment:标签ID;"`   // 标签ID
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`             // 创建时间
}

// TableName 指定UserDishTag对应的数据表名。
func (UserDishTag) TableName() string { return "of_dish_tags" }

// UserRecipe 表示用户菜谱。
type UserRecipe struct {
	global.GVA_MODEL             // GVA基础模型字段
	PublicID         string      `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"` // 对外公开ID
	OwnerID          string      `json:"ownerId" gorm:"column:owner_id;type:varchar(64);not null;index;comment:所属用户ID;"`         // 所属用户ID
	Owner            MiniAppUser `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`                                          // 所属用户关联数据
	Name             string      `json:"name" gorm:"column:name;type:varchar(120);not null;index;comment:名称;"`                   // 名称
	Note             *string     `json:"note" gorm:"column:note;type:varchar(1000);default:null;comment:备注;"`                    // 备注
	Version          int         `json:"version" gorm:"column:version;type:int;not null;default:1;comment:数据版本;"`                // 数据版本

	DeletedReason        *string `json:"deletedReason" gorm:"column:deleted_reason;type:varchar(600);default:null;comment:删除原因;"`                     // 删除原因
	DeletedViolationType *string `json:"deletedViolationType" gorm:"column:deleted_violation_type;type:varchar(80);default:null;comment:删除违规类型;"`     // 删除违规类型
	DeletedByAdminID     *uint   `json:"deletedByAdminId" gorm:"column:deleted_by_admin_id;type:bigint unsigned;index;default:null;comment:删除管理员ID;"` // 删除管理员ID
	DeletedByUsername    *string `json:"deletedByUsername" gorm:"column:deleted_by_username;type:varchar(120);default:null;comment:删除管理员用户名;"`        // 删除管理员用户名
	DeletedByNickname    *string `json:"deletedByNickname" gorm:"column:deleted_by_nickname;type:varchar(120);default:null;comment:删除管理员昵称;"`         // 删除管理员昵称

	RecipeDishes []RecipeDish `json:"recipeDishes" gorm:"foreignKey:RecipeID"` // 菜谱菜品关联列表
}

// TableName 指定UserRecipe对应的数据表名。
func (UserRecipe) TableName() string { return "of_recipes" }

// RecipeDish 表示菜谱与菜品关联。
type RecipeDish struct {
	global.GVA_MODEL           // GVA基础模型字段
	PublicID         string    `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                         // 对外公开ID
	RecipeID         uint      `json:"recipeId" gorm:"column:recipe_id;type:bigint unsigned;not null;uniqueIndex:idx_recipe_dish;index;comment:菜谱ID;"` // 菜谱ID
	DishID           uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;uniqueIndex:idx_recipe_dish;index;comment:菜品ID;"`     // 菜品ID
	Dish             UserDish  `json:"dish" gorm:"foreignKey:DishID"`                                                                                  // 菜品关联数据
	SortOrder        int       `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`                              // 排序值
	AddedAt          time.Time `json:"addedAt" gorm:"column:added_at;type:datetime;not null;comment:加入菜谱时间;"`                                          // 加入菜谱时间
}

// TableName 指定RecipeDish对应的数据表名。
func (RecipeDish) TableName() string { return "of_recipe_dishes" }

// DishReference 表示菜品业务引用。
type DishReference struct {
	global.GVA_MODEL             // GVA基础模型字段
	PublicID           string    `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                                // 对外公开ID
	DishID             uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`                                        // 菜品ID
	ReferenceType      string    `json:"referenceType" gorm:"column:reference_type;type:varchar(40);not null;index;comment:引用类型;"`                              // 引用类型
	ObjectID           string    `json:"objectId" gorm:"column:object_id;type:varchar(64);not null;index;comment:对象ID;"`                                        // 对象ID
	ObjectLabel        string    `json:"objectLabel" gorm:"column:object_label;type:varchar(160);not null;comment:对象名称;"`                                       // 对象名称
	ObjectStatus       string    `json:"objectStatus" gorm:"column:object_status;type:varchar(40);not null;comment:对象状态;"`                                      // 对象状态
	HistoricalSnapshot bool      `json:"historicalSnapshot" gorm:"column:historical_snapshot;type:tinyint(1) unsigned;not null;default:false;comment:是否为历史快照;"` // 是否为历史快照
	OccurredAt         time.Time `json:"occurredAt" gorm:"column:occurred_at;type:datetime;not null;index;comment:发生时间;"`                                       // 发生时间
}

// TableName 指定DishReference对应的数据表名。
func (DishReference) TableName() string { return "of_dish_refs" }

// GovernanceRecord 表示内容处理记录。
type GovernanceRecord struct {
	global.GVA_MODEL                     // GVA基础模型字段
	PublicID              string         `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                   // 对外公开ID
	TargetType            string         `json:"targetType" gorm:"column:target_type;type:varchar(32);not null;index;comment:目标类型;"`                       // 目标类型
	TargetID              string         `json:"targetId" gorm:"column:target_id;type:varchar(64);not null;index;comment:目标ID;"`                           // 目标ID
	TargetLabel           string         `json:"targetLabel" gorm:"column:target_label;type:varchar(160);not null;comment:目标名称;"`                          // 目标名称
	Actions               datatypes.JSON `json:"actions" gorm:"column:actions;type:json;not null;comment:操作集合JSON;"`                                       // 操作集合JSON
	ViolationType         string         `json:"violationType" gorm:"column:violation_type;type:varchar(80);not null;index;comment:违规类型;"`                 // 违规类型
	Severity              string         `json:"severity" gorm:"column:severity;type:varchar(20);not null;comment:处理严重程度;"`                                // 处理严重程度
	Reason                string         `json:"reason" gorm:"column:reason;type:varchar(600);not null;comment:原因;"`                                       // 原因
	AdministratorID       uint           `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`        // 管理员ID
	AdministratorUsername string         `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);not null;comment:管理员用户名;"`    // 管理员用户名
	AdministratorNickname *string        `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:管理员昵称;"` // 管理员昵称
	JobStatus             *string        `json:"jobStatus" gorm:"column:job_status;type:varchar(32);default:null;comment:异步任务状态;"`                         // 异步任务状态
	ImpactSnapshot        datatypes.JSON `json:"impactSnapshot" gorm:"column:impact_snapshot;type:json;not null;comment:影响范围快照JSON;"`                      // 影响范围快照JSON
	BeforeSummary         datatypes.JSON `json:"beforeSummary" gorm:"column:before_summary;type:json;default:null;comment:变更前摘要JSON;"`                     // 变更前摘要JSON
	AfterSummary          datatypes.JSON `json:"afterSummary" gorm:"column:after_summary;type:json;default:null;comment:变更后摘要JSON;"`                       // 变更后摘要JSON
	NotificationIDs       datatypes.JSON `json:"notificationIds" gorm:"column:notification_i_ds;type:json;not null;comment:通知ID集合JSON;"`                   // 通知ID集合JSON
	RequestID             string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                        // 请求ID
	IdempotencyKey        string         `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(160);not null;index;comment:幂等键;"`               // 幂等键
	AffectedCount         int            `json:"affectedCount" gorm:"column:affected_count;type:int;not null;default:0;comment:受影响数量;"`                    // 受影响数量
}

// TableName 指定GovernanceRecord对应的数据表名。
func (GovernanceRecord) TableName() string { return "of_governance_records" }

// GovernanceJob 表示违规复制链异步处理任务。
type GovernanceJob struct {
	ID               string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:任务ID;"`                           // 任务ID
	RecordID         string     `json:"recordId" gorm:"column:record_id;type:varchar(64);not null;uniqueIndex;comment:违规处理记录ID;"`         // 违规处理记录ID
	Status           string     `json:"status" gorm:"column:status;type:varchar(32);not null;index;comment:任务状态;"`                        // 任务状态
	TotalCount       int        `json:"totalCount" gorm:"column:total_count;type:int;not null;default:0;comment:目标总数;"`                   // 目标总数
	SucceededCount   int        `json:"succeededCount" gorm:"column:succeeded_count;type:int;not null;default:0;comment:成功数量;"`           // 成功数量
	FailedCount      int        `json:"failedCount" gorm:"column:failed_count;type:int;not null;default:0;comment:失败数量;"`                 // 失败数量
	PendingCount     int        `json:"pendingCount" gorm:"column:pending_count;type:int;not null;default:0;comment:待处理数量;"`              // 待处理数量
	LastErrorSummary *string    `json:"lastErrorSummary" gorm:"column:last_error_summary;type:varchar(240);default:null;comment:最近错误摘要;"` // 最近错误摘要
	Version          int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                       // 数据版本
	StartedAt        *time.Time `json:"startedAt" gorm:"column:started_at;type:datetime;default:null;comment:开始时间;"`                      // 开始时间
	FinishedAt       *time.Time `json:"finishedAt" gorm:"column:finished_at;type:datetime;default:null;comment:完成时间;"`                    // 完成时间
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                    // 创建时间
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                    // 更新时间
}

// TableName 指定GovernanceJob对应的数据表名。
func (GovernanceJob) TableName() string { return "of_gov_jobs" }

// GovernanceJobItem 表示违规复制链任务中的单个菜品处理项。
type GovernanceJobItem struct {
	ID               string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:任务项ID;"`                                                // 任务项ID
	JobID            string     `json:"jobId" gorm:"column:job_id;type:varchar(64);not null;uniqueIndex:uk_of_gov_item,priority:1;index;comment:任务ID;"`         // 任务ID
	DishID           uint       `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;uniqueIndex:uk_of_gov_item,priority:2;index;comment:菜品内部ID;"` // 菜品内部ID
	DishPublicID     string     `json:"dishPublicId" gorm:"column:dish_public_id;type:varchar(64);not null;comment:菜品公开ID;"`                                    // 菜品公开ID
	UserID           string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:所属用户ID;"`                                           // 所属用户ID
	Status           string     `json:"status" gorm:"column:status;type:varchar(24);not null;index;comment:任务项状态;"`                                             // 任务项状态
	AttemptCount     int        `json:"attemptCount" gorm:"column:attempt_count;type:int;not null;default:0;comment:执行次数;"`                                     // 执行次数
	LastErrorSummary *string    `json:"lastErrorSummary" gorm:"column:last_error_summary;type:varchar(240);default:null;comment:最近错误摘要;"`                       // 最近错误摘要
	NotificationID   *string    `json:"notificationId" gorm:"column:notification_id;type:varchar(64);default:null;comment:已生成站内通知ID;"`                          // 已生成站内通知ID
	ProcessedAt      *time.Time `json:"processedAt" gorm:"column:processed_at;type:datetime;default:null;comment:最近处理时间;"`                                      // 最近处理时间
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                // 创建时间
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                          // 更新时间
}

// TableName 指定GovernanceJobItem对应的数据表名。
func (GovernanceJobItem) TableName() string { return "of_gov_items" }

// AdminAuditLog 表示管理端业务审计日志。
type AdminAuditLog struct {
	global.GVA_MODEL                     // GVA基础模型字段
	PublicID              string         `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                   // 对外公开ID
	AdministratorID       uint           `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;not null;index;comment:管理员ID;"`        // 管理员ID
	AdministratorUsername string         `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);not null;comment:管理员用户名;"`    // 管理员用户名
	AdministratorNickname *string        `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:管理员昵称;"` // 管理员昵称
	Action                string         `json:"action" gorm:"column:action;type:varchar(80);not null;index;comment:操作类型;"`                                // 操作类型
	TargetType            string         `json:"targetType" gorm:"column:target_type;type:varchar(32);not null;index;comment:目标类型;"`                       // 目标类型
	TargetID              string         `json:"targetId" gorm:"column:target_id;type:varchar(64);not null;index;comment:目标ID;"`                           // 目标ID
	TargetLabel           *string        `json:"targetLabel" gorm:"column:target_label;type:varchar(160);default:null;comment:目标名称;"`                      // 目标名称
	Reason                *string        `json:"reason" gorm:"column:reason;type:varchar(600);default:null;comment:原因;"`                                   // 原因
	BeforeSummary         datatypes.JSON `json:"beforeSummary" gorm:"column:before_summary;type:json;default:null;comment:变更前摘要JSON;"`                     // 变更前摘要JSON
	AfterSummary          datatypes.JSON `json:"afterSummary" gorm:"column:after_summary;type:json;default:null;comment:变更后摘要JSON;"`                       // 变更后摘要JSON
	RequestID             string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;index;comment:请求ID;"`                        // 请求ID
	IdempotencyKey        *string        `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(160);index;default:null;comment:幂等键;"`           // 幂等键
	SourceIPMasked        *string        `json:"sourceIpMasked" gorm:"column:source_ip_masked;type:varchar(80);default:null;comment:脱敏来源IP;"`              // 脱敏来源IP
	UserAgentSummary      *string        `json:"userAgentSummary" gorm:"column:user_agent_summary;type:varchar(240);default:null;comment:客户端信息摘要;"`        // 客户端信息摘要
}

// TableName 指定AdminAuditLog对应的数据表名。
func (AdminAuditLog) TableName() string { return "of_admin_audits" }

// UserNotification 表示用户站内通知。
type UserNotification struct {
	ID                string      `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                               // 主键ID
	UserID            string      `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                           // 用户ID
	User              MiniAppUser `json:"user" gorm:"foreignKey:UserID;references:ID"`                                                                          // 用户关联数据
	Type              string      `json:"type" gorm:"column:type;type:varchar(32);not null;index;comment:类型;"`                                                  // 类型
	Title             string      `json:"title" gorm:"column:title;type:varchar(120);not null;comment:标题;"`                                                     // 标题
	Content           string      `json:"content" gorm:"column:content;type:varchar(1000);not null;comment:内容;"`                                                // 内容
	TargetType        *string     `json:"targetType" gorm:"column:target_type;type:varchar(40);index;default:null;comment:目标类型;"`                               // 目标类型
	TargetID          *string     `json:"targetId" gorm:"column:target_id;type:varchar(64);index;default:null;comment:目标ID;"`                                   // 目标ID
	SubscribeRequired bool        `json:"subscribeRequired" gorm:"column:subscribe_required;type:tinyint(1) unsigned;not null;default:false;comment:是否需要订阅消息;"` // 是否需要订阅消息
	SubscribeLogID    *string     `json:"subscribeLogId" gorm:"column:subscribe_log_id;type:varchar(64);index;default:null;comment:订阅消息日志ID;"`                  // 订阅消息日志ID
	ReadAt            *time.Time  `json:"readAt" gorm:"column:read_at;type:datetime;index;default:null;comment:阅读时间;"`                                          // 阅读时间
	CreatedAt         time.Time   `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                        // 创建时间
}

// TableName 指定UserNotification对应的数据表名。
func (UserNotification) TableName() string { return "of_notifications" }

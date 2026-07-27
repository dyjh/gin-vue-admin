package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FrontMediaAsset 表示小程序媒体文件。
type FrontMediaAsset struct {
	ID                    string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                      // 主键ID
	UserID                string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;default:'';index;comment:用户ID;"`                       // 用户ID
	AdministratorID       *uint      `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;index;default:null;comment:上传管理员ID;"`     // 上传管理员ID
	AdministratorUsername *string    `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);default:null;comment:上传管理员用户名;"` // 上传管理员用户名
	AdministratorNickname *string    `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:上传管理员昵称;"`  // 上传管理员昵称
	FileName              string     `json:"fileName" gorm:"column:file_name;type:varchar(255);not null;default:'';comment:原始文件名;"`                       // 原始文件名
	UploadSource          string     `json:"uploadSource" gorm:"column:upload_source;type:varchar(24);not null;default:user_upload;index;comment:上传来源;"`  // 上传来源
	Scene                 string     `json:"scene" gorm:"column:scene;type:varchar(32);not null;index;comment:审核场景;"`                                     // 审核场景
	ResourceStatus        string     `json:"resourceStatus" gorm:"column:resource_status;type:varchar(24);not null;default:active;index;comment:资源状态;"`   // 资源状态
	URL                   string     `json:"url" gorm:"column:url;type:varchar(512);not null;comment:访问地址;"`                                              // 访问地址
	StoragePath           string     `json:"-" gorm:"column:storage_path;type:varchar(1024);not null;comment:对象存储路径;"`                                    // 对象存储路径
	ContentType           string     `json:"-" gorm:"column:content_type;type:varchar(80);not null;comment:媒体类型;"`                                        // 媒体类型
	Width                 int        `json:"width" gorm:"column:width;type:int;not null;comment:图片宽度;"`                                                   // 图片宽度
	Height                int        `json:"height" gorm:"column:height;type:int;not null;comment:图片高度;"`                                                 // 图片高度
	SizeBytes             int64      `json:"-" gorm:"column:size_bytes;type:bigint;not null;comment:文件字节数;"`                                              // 文件字节数
	Checksum              string     `json:"-" gorm:"column:checksum;type:varchar(64);not null;default:'';comment:文件SHA256摘要;"`                           // 文件SHA256摘要
	ReviewStatus          string     `json:"reviewStatus" gorm:"column:review_status;type:varchar(24);not null;index;comment:图片审核状态;"`                    // 图片审核状态
	RejectReason          *string    `json:"rejectReason" gorm:"column:reject_reason;type:varchar(240);default:null;comment:拒绝原因;"`                       // 拒绝原因
	DeletedAt             *time.Time `json:"deletedAt" gorm:"column:deleted_at;type:datetime;index;default:null;comment:资源删除时间;"`                         // 资源删除时间
	DeleteReason          *string    `json:"deleteReason" gorm:"column:delete_reason;type:varchar(200);default:null;comment:资源删除原因;"`                     // 资源删除原因
	CreatedAt             time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                               // 创建时间
}

// TableName 指定FrontMediaAsset对应的数据表名。
func (FrontMediaAsset) TableName() string { return "of_media" }

// FrontUserActivityDay 表示用户在一个上海自然日内至少完成过一次成功业务请求。
type FrontUserActivityDay struct {
	ID         uint      `json:"id" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                                       // 主键ID
	UserID     string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_user_activity,priority:1;index;comment:用户ID;"`          // 用户ID
	ActiveDate string    `json:"activeDate" gorm:"column:active_date;type:char(10);not null;uniqueIndex:uk_of_user_activity,priority:2;index;comment:上海时区活跃日期;"` // 上海时区活跃日期
	FirstAt    time.Time `json:"firstAt" gorm:"column:first_at;type:datetime;not null;comment:当日首次成功业务请求时间;"`                                                    // 当日首次成功业务请求时间
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                        // 创建时间
}

// TableName 指定FrontUserActivityDay对应的数据表名。
func (FrontUserActivityDay) TableName() string { return "of_user_activity" }

// PlatformRecommendation 表示平台推荐菜。
type PlatformRecommendation struct {
	ID                string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                          // 主键ID
	DishID            *uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;index;default:null;comment:用户菜品内部ID;"`                          // 用户菜品内部ID
	OfficialDishID    *uint      `json:"officialDishId" gorm:"column:official_dish_id;type:bigint unsigned;index;default:null;comment:官方菜品内部ID;"`         // 官方菜品内部ID
	SourceType        string     `json:"sourceType" gorm:"column:source_type;type:varchar(24);not null;index;comment:来源类型;"`                              // 来源类型
	SourceDishID      string     `json:"sourceDishId" gorm:"column:source_dish_id;type:varchar(64);not null;default:'';index;comment:来源菜品公开ID;"`          // 来源菜品公开ID
	Position          string     `json:"position" gorm:"column:position;type:varchar(32);not null;default:home_featured;index;comment:推荐位置;"`             // 推荐位置
	DisplayNote       *string    `json:"displayNote" gorm:"column:display_note;type:varchar(160);default:null;comment:展示说明;"`                             // 展示说明
	Status            string     `json:"status" gorm:"column:status;type:varchar(24);not null;default:draft;index;comment:推荐状态;"`                         // 推荐状态
	Selected          bool       `json:"selected" gorm:"column:selected;type:tinyint(1) unsigned;not null;default:false;index;comment:小程序是否可见;"`          // 小程序是否可见
	SortOrder         int        `json:"sortOrder" gorm:"column:sort_order;type:int;not null;index;comment:排序值;"`                                         // 排序值
	Version           int64      `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                      // 数据版本
	PublishedAt       *time.Time `json:"publishedAt" gorm:"column:published_at;type:datetime;index;default:null;comment:发布时间;"`                           // 发布时间
	OfflineAt         *time.Time `json:"offlineAt" gorm:"column:offline_at;type:datetime;index;default:null;comment:下线时间;"`                               // 下线时间
	OfflineReason     *string    `json:"offlineReason" gorm:"column:offline_reason;type:varchar(200);default:null;comment:下线原因;"`                         // 下线原因
	CreatedByID       uint       `json:"createdById" gorm:"column:created_by_id;type:bigint unsigned;not null;default:0;index;comment:创建管理员ID;"`          // 创建管理员ID
	CreatedByUsername string     `json:"createdByUsername" gorm:"column:created_by_username;type:varchar(120);not null;default:system;comment:创建管理员用户名;"` // 创建管理员用户名
	CreatedByNickname *string    `json:"createdByNickname" gorm:"column:created_by_nickname;type:varchar(120);default:null;comment:创建管理员昵称;"`             // 创建管理员昵称
	UpdatedByID       uint       `json:"updatedById" gorm:"column:updated_by_id;type:bigint unsigned;not null;default:0;index;comment:更新管理员ID;"`          // 更新管理员ID
	UpdatedByUsername string     `json:"updatedByUsername" gorm:"column:updated_by_username;type:varchar(120);not null;default:system;comment:更新管理员用户名;"` // 更新管理员用户名
	UpdatedByNickname *string    `json:"updatedByNickname" gorm:"column:updated_by_nickname;type:varchar(120);default:null;comment:更新管理员昵称;"`             // 更新管理员昵称
	CreatedAt         time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                         // 创建时间
	UpdatedAt         time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`                                   // 更新时间
}

// TableName 指定PlatformRecommendation对应的数据表名。
func (PlatformRecommendation) TableName() string {
	return "of_recommendations"
}

// RecommendationCopy 表示推荐菜复制记录。
type RecommendationCopy struct {
	ID               string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                                     // 主键ID
	RecommendationID string    `json:"recommendationId" gorm:"column:recommendation_id;type:varchar(64);not null;uniqueIndex:uk_of_recommendation_copy,priority:1;comment:推荐位ID;"` // 推荐位ID
	UserID           string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_recommendation_copy,priority:2;comment:用户ID;"`                      // 用户ID
	CopiedDishID     uint      `json:"copiedDishId" gorm:"column:copied_dish_id;type:bigint unsigned;not null;index;comment:复制生成的菜品ID;"`                                           // 复制生成的菜品ID
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                    // 创建时间
}

// TableName 指定RecommendationCopy对应的数据表名。
func (RecommendationCopy) TableName() string {
	return "of_recommendation_copies"
}

// FrontCheckin 表示用户做菜打卡。
type FrontCheckin struct {
	ID                   string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                      // 主键ID
	UserID               string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                  // 用户ID
	DishName             string         `json:"dishName" gorm:"column:dish_name;type:varchar(40);not null;comment:菜品名称;"`                                    // 菜品名称
	ImageFileID          string         `json:"imageFileId" gorm:"column:image_file_id;type:varchar(64);not null;index;comment:图片文件ID;"`                     // 图片文件ID
	ImageURL             string         `json:"imageUrl" gorm:"column:image_url;type:varchar(512);not null;comment:图片地址;"`                                   // 图片地址
	Note                 *string        `json:"note" gorm:"column:note;type:varchar(120);default:null;comment:备注;"`                                          // 备注
	CheckedDate          string         `json:"-" gorm:"column:checked_date;type:varchar(10);not null;index;comment:打卡日期;"`                                  // 打卡日期
	Rewarded             bool           `json:"rewarded" gorm:"column:rewarded;type:tinyint(1) unsigned;not null;comment:是否已发放奖励;"`                          // 是否已发放奖励
	CheckedAt            time.Time      `json:"checkedAt" gorm:"column:checked_at;type:datetime;not null;index;comment:打卡时间;"`                               // 打卡时间
	Version              int            `json:"version" gorm:"column:version;type:int;not null;default:1;comment:数据版本;"`                                     // 数据版本
	DeletedReason        *string        `json:"deletedReason" gorm:"column:deleted_reason;type:varchar(600);default:null;comment:删除原因;"`                     // 删除原因
	DeletedViolationType *string        `json:"deletedViolationType" gorm:"column:deleted_violation_type;type:varchar(80);default:null;comment:删除违规类型;"`     // 删除违规类型
	DeletedByAdminID     *uint          `json:"deletedByAdminId" gorm:"column:deleted_by_admin_id;type:bigint unsigned;index;default:null;comment:删除管理员ID;"` // 删除管理员ID
	DeletedByUsername    *string        `json:"deletedByUsername" gorm:"column:deleted_by_username;type:varchar(120);default:null;comment:删除管理员用户名;"`        // 删除管理员用户名
	DeletedByNickname    *string        `json:"deletedByNickname" gorm:"column:deleted_by_nickname;type:varchar(120);default:null;comment:删除管理员昵称;"`         // 删除管理员昵称
	CreatedAt            time.Time      `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                             // 创建时间
	DeletedAt            gorm.DeletedAt `json:"deletedAt" gorm:"column:deleted_at;type:datetime;index;default:null;comment:软删除时间;"`                          // 软删除时间
}

// TableName 指定FrontCheckin对应的数据表名。
func (FrontCheckin) TableName() string { return "of_checkins" }

// MealStatus 表示饭局从点单到采购完成的生命周期状态。
type MealStatus string

const (
	MealCollecting MealStatus = "collecting"
	MealClosed     MealStatus = "closed"
	MealConfirmed  MealStatus = "confirmed"
	MealCompleted  MealStatus = "completed"
	MealCancelled  MealStatus = "cancelled"
)

// MealCloseSource 表示饭局关闭点单的触发来源。
type MealCloseSource string

const (
	MealCloseSourceCreatorAction MealCloseSource = "creator_action" // 发起人主动关闭
	MealCloseSourceMinuteScan    MealCloseSource = "minute_scan"    // 分钟级截止扫描关闭
	MealCloseSourceServiceGuard  MealCloseSource = "service_guard"  // 业务请求截止守卫关闭
	MealCloseSourceLegacy        MealCloseSource = "legacy"         // 历史数据未记录准确来源
)

// FrontMeal 表示饭局。
type FrontMeal struct {
	ID              string           `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                            // 主键ID
	CreatorID       string           `json:"creatorId" gorm:"column:creator_id;type:varchar(64);not null;index;comment:创建者用户ID;"`               // 创建者用户ID
	Name            string           `json:"name" gorm:"column:name;type:varchar(30);not null;comment:名称;"`                                     // 名称
	Code            string           `json:"code" gorm:"column:code;type:varchar(6);not null;uniqueIndex;comment:饭局邀请码;"`                       // 饭局邀请码
	Status          MealStatus       `json:"status" gorm:"column:status;type:varchar(20);not null;index;comment:状态;"`                           // 状态
	CloseReason     *string          `json:"closeReason" gorm:"column:close_reason;type:varchar(20);default:null;comment:关闭点单原因;"`              // 关闭点单原因
	CloseSource     *MealCloseSource `json:"-" gorm:"column:close_source;type:varchar(24);default:null;comment:关闭点单触发来源;"`                      // 关闭点单触发来源
	ClosedAt        *time.Time       `json:"closedAt" gorm:"column:closed_at;type:datetime;index;default:null;comment:关闭点单时间;"`                 // 关闭点单时间
	DeadlineAt      time.Time        `json:"deadlineAt" gorm:"column:deadline_at;type:datetime;not null;index;comment:点单截止时间;"`                 // 点单截止时间
	SourceRecipeID  *string          `json:"sourceRecipeId" gorm:"column:source_recipe_id;type:varchar(64);default:null;comment:来源菜谱ID;"`       // 来源菜谱ID
	CancelledReason *string          `json:"-" gorm:"column:cancelled_reason;type:varchar(100);default:null;comment:取消原因;"`                     // 取消原因
	CancelledFrom   *MealStatus      `json:"-" gorm:"column:cancelled_from_status;type:varchar(20);default:null;comment:取消前饭局状态;"`              // 取消前饭局状态
	ShoppingListID  *string          `json:"shoppingListId" gorm:"column:shopping_list_id;type:varchar(64);index;default:null;comment:采购清单ID;"` // 采购清单ID
	ConfirmedAt     *time.Time       `json:"confirmedAt" gorm:"column:confirmed_at;type:datetime;default:null;comment:确认时间;"`                   // 确认时间
	CompletedAt     *time.Time       `json:"completedAt" gorm:"column:completed_at;type:datetime;default:null;comment:完成时间;"`                   // 完成时间
	CancelledAt     *time.Time       `json:"cancelledAt" gorm:"column:cancelled_at;type:datetime;default:null;comment:取消时间;"`                   // 取消时间
	CreatedAt       time.Time        `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                     // 创建时间
	UpdatedAt       time.Time        `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                           // 更新时间
}

// TableName 指定FrontMeal对应的数据表名。
func (FrontMeal) TableName() string { return "of_meals" }

// MealParticipant 表示饭局参与人。
type MealParticipant struct {
	ID       string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                   // 主键ID
	MealID   string    `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_participant,priority:1;index;comment:饭局ID;"` // 饭局ID
	UserID   string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_participant,priority:2;index;comment:用户ID;"` // 用户ID
	JoinedAt time.Time `json:"joinedAt" gorm:"column:joined_at;type:datetime;not null;comment:加入时间;"`                                                    // 加入时间
}

// TableName 指定MealParticipant对应的数据表名。
func (MealParticipant) TableName() string { return "of_meal_members" }

// FrontMealCandidate 表示饭局候选菜。
type FrontMealCandidate struct {
	ID                string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                     // 主键ID
	MealID            string    `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_candidate,priority:1;index;comment:饭局ID;"`     // 饭局ID
	DishID            uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;uniqueIndex:uk_of_meal_candidate,priority:2;index;comment:菜品ID;"` // 菜品ID
	Available         bool      `json:"available" gorm:"column:available;type:tinyint(1) unsigned;not null;default:true;index;comment:候选菜是否可选;"`                    // 候选菜是否可选
	UnavailableReason *string   `json:"unavailableReason" gorm:"column:unavailable_reason;type:varchar(24);default:null;comment:不可选原因;"`                            // 不可选原因
	SortOrder         int       `json:"sortOrder" gorm:"column:sort_order;type:int;not null;comment:排序值;"`                                                          // 排序值
	CreatedAt         time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                    // 创建时间
}

// TableName 指定FrontMealCandidate对应的数据表名。
func (FrontMealCandidate) TableName() string { return "of_meal_candidates" }

// FrontMealVote 表示饭局点选记录。
type FrontMealVote struct {
	ID          string    `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                       // 主键ID
	MealID      string    `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_vote,priority:1;index;comment:饭局ID;"`            // 饭局ID
	UserID      string    `json:"userId" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_vote,priority:2;index;comment:用户ID;"`            // 用户ID
	CandidateID string    `json:"candidateId" gorm:"column:candidate_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_vote,priority:3;index;comment:候选菜ID;"` // 候选菜ID
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                      // 创建时间
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                      // 更新时间
}

// TableName 指定FrontMealVote对应的数据表名。
func (FrontMealVote) TableName() string { return "of_meal_votes" }

// FrontMealFinalDish 表示饭局最终菜品快照。
type FrontMealFinalDish struct {
	ID            string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                  // 主键ID
	MealID        string         `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_final,priority:1;index;comment:饭局ID;"`      // 饭局ID
	CandidateID   string         `json:"candidateId" gorm:"column:candidate_id;type:varchar(64);not null;uniqueIndex:uk_of_meal_final,priority:2;comment:候选菜ID;"` // 候选菜ID
	DishID        uint           `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:菜品ID;"`                                          // 菜品ID
	Name          string         `json:"name" gorm:"column:name;type:varchar(120);not null;comment:名称;"`                                                          // 名称
	CoverURL      *string        `json:"coverUrl" gorm:"column:cover_url;type:varchar(512);default:null;comment:封面地址;"`                                           // 封面地址
	FinalServings int            `json:"finalServings" gorm:"column:final_servings;type:int;not null;comment:最终份数;"`                                              // 最终份数
	Ingredients   datatypes.JSON `json:"-" gorm:"column:ingredients;type:json;not null;comment:食材列表;"`                                                            // 食材列表
	Steps         datatypes.JSON `json:"-" gorm:"column:steps_json;type:json;default:null;comment:步骤列表;"`                                                         // 步骤列表
	CreatedAt     time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                 // 创建时间
}

// TableName 指定FrontMealFinalDish对应的数据表名。
func (FrontMealFinalDish) TableName() string { return "of_meal_final_dishes" }

// FrontShoppingList 表示采购清单。
type FrontShoppingList struct {
	ID             string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                 // 主键ID
	MealID         string     `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex;comment:饭局ID;"`       // 饭局ID
	OwnerID        string     `json:"ownerId" gorm:"column:owner_id;type:varchar(64);not null;index;comment:所属用户ID;"`         // 所属用户ID
	ShareTokenHash string     `json:"-" gorm:"column:share_token_hash;type:varchar(64);not null;uniqueIndex;comment:分享令牌摘要;"` // 分享令牌摘要
	ShareToken     string     `json:"shareToken" gorm:"column:share_token;type:varchar(128);not null;comment:分享令牌;"`          // 分享令牌
	ShareExpiresAt *time.Time `json:"-" gorm:"column:share_expires_at;type:datetime;index;default:null;comment:分享过期时间;"`      // 分享过期时间
	ShareRevokedAt *time.Time `json:"-" gorm:"column:share_revoked_at;type:datetime;index;default:null;comment:分享撤销时间;"`      // 分享撤销时间
	CreatedAt      time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                // 创建时间
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                // 更新时间
}

// TableName 指定FrontShoppingList对应的数据表名。
func (FrontShoppingList) TableName() string { return "of_shopping_lists" }

// FrontShoppingItem 表示采购清单项。
type FrontShoppingItem struct {
	ID              string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                       // 主键ID
	ListID          string         `json:"listId" gorm:"column:list_id;type:varchar(64);not null;index;comment:采购清单ID;"`                 // 采购清单ID
	Name            string         `json:"name" gorm:"column:name;type:varchar(40);not null;comment:名称;"`                                // 名称
	Amount          string         `json:"amount" gorm:"column:amount;type:varchar(30);not null;comment:积分或用量变动值;"`                      // 积分或用量变动值
	Note            *string        `json:"note" gorm:"column:note;type:varchar(100);default:null;comment:备注;"`                           // 备注
	Completed       bool           `json:"completed" gorm:"column:completed;type:tinyint(1) unsigned;not null;index;comment:是否已完成;"`     // 是否已完成
	SourceDishNames datatypes.JSON `json:"sourceDishNames" gorm:"column:source_dish_names;type:json;default:null;comment:来源菜品名称集合JSON;"` // 来源菜品名称集合JSON
	SortOrder       int            `json:"sortOrder" gorm:"column:sort_order;type:int;not null;index;comment:排序值;"`                      // 排序值
	CreatedAt       time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                      // 创建时间
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                      // 更新时间
}

// TableName 指定FrontShoppingItem对应的数据表名。
func (FrontShoppingItem) TableName() string { return "of_shopping_items" }

// PointEntryType 表示积分流水的业务变动类型。
type PointEntryType string

const (
	// PointEarned 表示用户获得积分。
	PointEarned PointEntryType = "earned"
	// PointSpent 表示用户消耗积分。
	PointSpent PointEntryType = "spent"
	// PointRefund 表示系统退还已扣积分。
	PointRefund PointEntryType = "refund"
	// PointAdjustment 表示管理员人工调整积分。
	PointAdjustment PointEntryType = "adjustment"
)

const (
	// FeatureExecutionProcessing 表示增强功能正在执行。
	FeatureExecutionProcessing = "processing"
	// FeatureExecutionSucceeded 表示增强功能执行成功。
	FeatureExecutionSucceeded = "succeeded"
	// FeatureExecutionFailed 表示增强功能执行失败。
	FeatureExecutionFailed = "failed"

	// FeatureBillingNotCharged 表示本次调用没有扣除积分。
	FeatureBillingNotCharged = "not_charged"
	// FeatureBillingCharged 表示本次调用已经扣除积分。
	FeatureBillingCharged = "charged"
	// FeatureBillingRefundPending 表示失败调用的积分正在等待退还。
	FeatureBillingRefundPending = "refund_pending"
	// FeatureBillingRefunded 表示失败调用的积分已经退还。
	FeatureBillingRefunded = "refunded"
)

// FrontPointEntry 表示积分流水。
type FrontPointEntry struct {
	ID                    string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                    // 主键ID
	UserID                string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                // 用户ID
	Type                  PointEntryType `json:"type" gorm:"column:type;type:varchar(20);not null;index;comment:类型;"`                                       // 类型
	Scene                 string         `json:"scene" gorm:"column:scene;type:varchar(40);not null;default:'';index;comment:积分业务场景;"`                      // 积分业务场景
	Title                 string         `json:"title" gorm:"column:title;type:varchar(100);not null;comment:标题;"`                                          // 标题
	Description           string         `json:"description" gorm:"column:description;type:varchar(300);not null;comment:说明;"`                              // 说明
	Amount                int64          `json:"amount" gorm:"column:amount;type:bigint;not null;comment:积分或用量变动值;"`                                        // 积分或用量变动值
	BalanceAfter          int64          `json:"balanceAfter" gorm:"column:balance_after;type:bigint;not null;default:0;comment:变动后积分余额;"`                  // 变动后积分余额
	RelatedObjectType     *string        `json:"relatedObjectType" gorm:"column:related_object_type;type:varchar(40);index;default:null;comment:关联对象类型;"`   // 关联对象类型
	RelatedObjectID       *string        `json:"relatedObjectId" gorm:"column:related_object_id;type:varchar(64);index;default:null;comment:关联对象ID;"`       // 关联对象ID
	RelatedEntryID        *string        `json:"relatedEntryId" gorm:"column:related_entry_id;type:varchar(64);index;default:null;comment:关联积分流水ID;"`       // 关联积分流水ID
	RequestID             *string        `json:"requestId" gorm:"column:request_id;type:varchar(128);index;default:null;comment:请求ID;"`                     // 请求ID
	IdempotencyKey        *string        `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(160);index;default:null;comment:幂等键;"`            // 幂等键
	AdministratorID       *uint          `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;index;default:null;comment:操作管理员ID;"`   // 操作管理员ID
	AdministratorUsername *string        `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);default:null;comment:管理员用户名;"` // 管理员用户名
	AdministratorNickname *string        `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:管理员昵称;"`  // 管理员昵称
	CreatedAt             time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                             // 创建时间
}

// TableName 指定FrontPointEntry对应的数据表名。
func (FrontPointEntry) TableName() string { return "of_point_entries" }

// FrontFeatureUsage 表示AI功能调用与计费记录。
type FrontFeatureUsage struct {
	ID                      string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                           // 主键ID
	UserID                  string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                                       // 用户ID
	Feature                 string         `json:"feature" gorm:"column:feature;type:varchar(64);not null;index;comment:客户端功能编码;"`                                                   // 客户端功能编码
	CapabilityCode          string         `json:"capabilityCode" gorm:"column:capability_code;type:varchar(64);not null;default:'';index;comment:实际AI能力编码;"`                        // 实际AI能力编码
	ProviderID              string         `json:"providerId" gorm:"column:provider_id;type:varchar(64);not null;default:'';index;comment:实际供应商ID;"`                                 // 实际供应商ID
	ProviderName            string         `json:"providerName" gorm:"column:provider_name;type:varchar(60);not null;default:'';comment:实际供应商名称快照;"`                                 // 实际供应商名称快照
	ModelID                 string         `json:"modelId" gorm:"column:model_id;type:varchar(64);not null;default:'';index;comment:实际模型ID;"`                                        // 实际模型ID
	ModelName               string         `json:"modelName" gorm:"column:model_name;type:varchar(60);not null;default:'';comment:实际模型名称快照;"`                                        // 实际模型名称快照
	RequestID               string         `json:"requestId" gorm:"column:request_id;type:varchar(128);not null;default:'';index;comment:请求ID;"`                                     // 请求ID
	IdempotencyKey          string         `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(128);not null;default:'';index;comment:请求幂等键;"`                          // 请求幂等键
	PromptMode              string         `json:"promptMode" gorm:"column:prompt_mode;type:varchar(16);not null;default:'';comment:提示词模式;"`                                         // 提示词模式
	PromptHash              string         `json:"promptHash" gorm:"column:prompt_hash;type:varchar(64);not null;default:'';comment:提示词内容摘要;"`                                       // 提示词内容摘要
	PointCost               int            `json:"pointCost" gorm:"column:point_cost;type:int;not null;default:0;comment:积分消耗;"`                                                     // 积分消耗
	ExecutionStatus         string         `json:"executionStatus" gorm:"column:execution_status;type:varchar(20);not null;index;comment:执行状态;"`                                     // 执行状态
	BillingStatus           string         `json:"billingStatus" gorm:"column:billing_status;type:varchar(20);not null;index;comment:计费状态;"`                                         // 计费状态
	DurationMS              *int           `json:"durationMs" gorm:"column:duration_ms;type:int;default:null;comment:调用耗时毫秒数;"`                                                      // 调用耗时毫秒数
	InputTokens             *int           `json:"inputTokens" gorm:"column:input_tokens;type:int;default:null;comment:输入Token数;"`                                                   // 输入Token数
	OutputTokens            *int           `json:"outputTokens" gorm:"column:output_tokens;type:int;default:null;comment:输出Token数;"`                                                 // 输出Token数
	ImageCount              int            `json:"imageCount" gorm:"column:image_count;type:int;not null;default:0;comment:生成或分析图片数;"`                                               // 生成或分析图片数
	EstimatedCostCNY        string         `json:"estimatedCostCny" gorm:"column:estimated_cost_cny;type:decimal(18,6);not null;default:0;comment:估算成本人民币;"`                         // 估算成本人民币
	FailureCategory         *string        `json:"failureCategory" gorm:"column:failure_category;type:varchar(64);default:null;index;comment:失败分类;"`                                 // 失败分类
	FailureSummary          *string        `json:"failureSummary" gorm:"column:failure_summary;type:varchar(240);default:null;comment:安全失败摘要;"`                                      // 安全失败摘要
	OriginalInputJSON       datatypes.JSON `json:"-" gorm:"column:original_input_json;type:json;default:null;comment:原始输入JSON;"`                                                     // 原始输入JSON
	ImageMetadataJSON       datatypes.JSON `json:"-" gorm:"column:image_metadata_json;type:json;default:null;comment:图片元信息JSON;"`                                                    // 图片元信息JSON
	ModelOutputJSON         datatypes.JSON `json:"-" gorm:"column:model_output_json;type:json;default:null;comment:模型输出JSON;"`                                                       // 模型输出JSON
	SensitiveContentCleared bool           `json:"sensitiveContentCleared" gorm:"column:sensitive_cleared;type:tinyint(1) unsigned;not null;default:false;index;comment:敏感内容是否已清除;"` // 敏感内容是否已清除
	ClearedAt               *time.Time     `json:"clearedAt" gorm:"column:cleared_at;type:datetime;default:null;comment:敏感内容清除时间;"`                                                  // 敏感内容清除时间
	ClearedByID             *uint          `json:"-" gorm:"column:cleared_by_id;type:bigint unsigned;default:null;index;comment:清除管理员ID;"`                                           // 清除管理员ID
	ClearedByUsername       *string        `json:"-" gorm:"column:cleared_by_username;type:varchar(120);default:null;comment:清除管理员用户名;"`                                             // 清除管理员用户名
	ClearedByNickname       *string        `json:"-" gorm:"column:cleared_by_nickname;type:varchar(120);default:null;comment:清除管理员昵称;"`                                              // 清除管理员昵称
	Version                 int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                       // 数据版本
	FinishedAt              *time.Time     `json:"finishedAt" gorm:"column:finished_at;type:datetime;default:null;index;comment:调用结束时间;"`                                            // 调用结束时间
	CreatedAt               time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                                    // 创建时间
	UpdatedAt               time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                          // 更新时间
}

// TableName 指定FrontFeatureUsage对应的数据表名。
func (FrontFeatureUsage) TableName() string { return "of_ai_usages" }

// FrontMealSuggestion 表示AI饭局推荐结果。
type FrontMealSuggestion struct {
	ID          string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`           // 主键ID
	UserID      string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`       // 用户ID
	Source      string         `json:"source" gorm:"column:source;type:varchar(20);not null;comment:来源;"`                // 来源
	SourceLabel string         `json:"sourceLabel" gorm:"column:source_label;type:varchar(80);not null;comment:来源说明;"`   // 来源说明
	Reason      string         `json:"reason" gorm:"column:reason;type:varchar(300);not null;comment:原因;"`               // 原因
	People      int            `json:"people" gorm:"column:people;type:int;not null;comment:用餐人数;"`                      // 用餐人数
	DishIDsJSON datatypes.JSON `json:"-" gorm:"column:dish_i_ds_json;type:json;not null;comment:菜品ID集合JSON;"`            // 菜品ID集合JSON
	UsageID     string         `json:"usageId" gorm:"column:usage_id;type:varchar(64);not null;index;comment:AI调用记录ID;"` // AI调用记录ID
	Feedback    *string        `json:"feedback" gorm:"column:feedback;type:varchar(20);default:null;comment:用户反馈;"`      // 用户反馈
	CreatedAt   time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`    // 创建时间
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`          // 更新时间
}

// TableName 指定FrontMealSuggestion对应的数据表名。
func (FrontMealSuggestion) TableName() string { return "of_meal_suggestions" }

// FrontMealSuggestionDish 表示AI饭局推荐菜。
type FrontMealSuggestionDish struct {
	ID                   string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                                                // 主键ID
	SuggestionID         string         `json:"suggestionId" gorm:"column:suggestion_id;type:varchar(64);not null;uniqueIndex:uk_of_suggestion_dish,priority:1;index;comment:推荐结果ID;"` // 推荐结果ID
	SortOrder            int            `json:"sortOrder" gorm:"column:sort_order;type:int;not null;uniqueIndex:uk_of_suggestion_dish,priority:2;comment:排序值;"`                        // 排序值
	Source               string         `json:"source" gorm:"column:source;type:varchar(24);not null;index;comment:来源;"`                                                               // 来源
	SourceLabel          string         `json:"sourceLabel" gorm:"column:source_label;type:varchar(80);not null;comment:来源说明;"`                                                        // 来源说明
	SourceDishID         *uint          `json:"-" gorm:"column:source_dish_id;type:bigint unsigned;index;default:null;comment:来源菜品ID;"`                                                // 来源菜品ID
	SourceOfficialDishID *uint          `json:"-" gorm:"column:source_official_dish_id;type:bigint unsigned;index;default:null;comment:来源官方菜品ID;"`                                     // 来源官方菜品ID
	RecommendationID     *string        `json:"recommendationId" gorm:"column:recommendation_id;type:varchar(64);index;default:null;comment:推荐位ID;"`                                   // 推荐位ID
	Name                 string         `json:"name" gorm:"column:name;type:varchar(120);not null;index;comment:名称;"`                                                                  // 名称
	NormalizedName       string         `json:"-" gorm:"column:normalized_name;type:varchar(120);not null;index;comment:标准化菜品名称;"`                                                     // 标准化菜品名称
	SnapshotJSON         datatypes.JSON `json:"-" gorm:"column:snapshot_json;type:json;not null;comment:菜品快照JSON;"`                                                                    // 菜品快照JSON
	CopiedDishID         *uint          `json:"-" gorm:"column:copied_dish_id;type:bigint unsigned;index;default:null;comment:复制生成的菜品ID;"`                                             // 复制生成的菜品ID
	AdoptedAt            *time.Time     `json:"adoptedAt" gorm:"column:adopted_at;type:datetime;default:null;comment:采纳时间;"`                                                           // 采纳时间
	CreatedAt            time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                                                         // 创建时间
}

// TableName 指定FrontMealSuggestionDish对应的数据表名。
func (FrontMealSuggestionDish) TableName() string {
	return "of_suggestion_dishes"
}

// FrontPrepPlan 表示AI备菜计划。
type FrontPrepPlan struct {
	ID               string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`               // 主键ID
	UserID           string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`           // 用户ID
	MealID           string         `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;index;comment:饭局ID;"`           // 饭局ID
	EstimatedMinutes int            `json:"estimatedMinutes" gorm:"column:estimated_minutes;type:int;not null;comment:预计耗时（分钟）;"` // 预计耗时（分钟）
	StepsJSON        datatypes.JSON `json:"-" gorm:"column:steps_json;type:json;not null;comment:步骤JSON;"`                        // 步骤JSON
	UsageID          string         `json:"usageId" gorm:"column:usage_id;type:varchar(64);not null;index;comment:AI调用记录ID;"`     // AI调用记录ID
	Feedback         *string        `json:"feedback" gorm:"column:feedback;type:varchar(20);default:null;comment:用户反馈;"`          // 用户反馈
	CreatedAt        time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`        // 创建时间
	UpdatedAt        time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`              // 更新时间
}

// TableName 指定FrontPrepPlan对应的数据表名。
func (FrontPrepPlan) TableName() string { return "of_prep_plans" }

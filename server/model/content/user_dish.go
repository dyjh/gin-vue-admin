package content

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	dishModel "github.com/dyjh/order-food-mini-app/server/model/dish"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"time"
)

const (
	DishStatusDraft  = "draft"
	DishStatusUsable = "usable"

	SourceTypeManual              = "manual"
	SourceTypeCreatorCopy         = "creator_copy"
	SourceTypeOfficialCopy        = "official_copy"
	SourceTypeMealSuggestionCopy  = "meal_suggestion_copy"
	SourceTypeGeneratedSuggestion = "generated_suggestion"
)

// UserDish 表示用户菜品。
type UserDish struct {
	global.GVA_MODEL                           // GVA基础模型字段
	PublicID         string                    `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`                           // 对外公开ID
	OwnerID          string                    `json:"ownerId" gorm:"column:owner_id;type:varchar(64);not null;index;comment:所属用户ID;"`                                   // 所属用户ID
	Owner            userModel.MiniAppUser     `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`                                                                    // 所属用户关联数据
	CoverFileID      string                    `json:"coverFileId" gorm:"column:cover_file_id;type:varchar(64);not null;comment:封面文件ID;"`                                // 封面文件ID
	CoverURL         *string                   `json:"coverUrl" gorm:"column:cover_url;type:varchar(512);default:null;comment:封面地址;"`                                    // 封面地址
	Name             string                    `json:"name" gorm:"column:name;type:varchar(120);not null;index;comment:名称;"`                                             // 名称
	CategoryID       uint                      `json:"categoryId" gorm:"column:category_id;type:bigint unsigned;not null;index;comment:分类ID;"`                           // 分类ID
	Category         dishModel.ContentCategory `json:"category" gorm:"foreignKey:CategoryID"`                                                                            // 分类关联数据
	Status           string                    `json:"status" gorm:"column:status;type:varchar(20);not null;index;comment:状态;"`                                          // 状态
	Discoverable     bool                      `json:"discoverable" gorm:"column:discoverable;type:tinyint(1) unsigned;not null;default:false;index;comment:是否允许被发现;"`   // 是否允许被发现
	DiscoverableAt   *time.Time                `json:"discoverableAt" gorm:"column:discoverable_at;type:datetime;default:null;comment:公开时间;"`                            // 公开时间
	SourceType       string                    `json:"sourceType" gorm:"column:source_type;type:varchar(32);not null;index;comment:来源类型;"`                               // 来源类型
	SourceLocked     bool                      `json:"sourceLocked" gorm:"column:source_locked;type:tinyint(1) unsigned;not null;default:false;index;comment:来源字段是否锁定;"` // 来源字段是否锁定
	Description      *string                   `json:"description" gorm:"column:description;type:varchar(1000);default:null;comment:说明;"`                                // 说明
	Serving          int                       `json:"serving" gorm:"column:serving;type:int;not null;default:1;comment:默认份数;"`                                          // 默认份数

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

	Ingredients []DishIngredient       `json:"ingredients" gorm:"foreignKey:DishID"`                                          // 食材列表
	Steps       []DishStep             `json:"steps" gorm:"foreignKey:DishID"`                                                // 步骤列表
	Tags        []dishModel.ContentTag `json:"tags" gorm:"many2many:of_dish_tags;joinForeignKey:DishID;joinReferences:TagID"` // 标签关联数据
}

// TableName 指定UserDish对应的数据表名。
func (UserDish) TableName() string { return "of_dishes" }

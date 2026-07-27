package model

import (
	"time"

	"github.com/dyjh/order-food-mini-app/server/global"
)

// OfficialDish 表示平台维护的官方菜品。
type OfficialDish struct {
	global.GVA_MODEL                             // GVA基础模型字段
	PublicID            string                   `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"`              // 对外公开ID
	Name                string                   `json:"name" gorm:"column:name;type:varchar(40);not null;index;comment:名称;"`                                 // 名称
	CoverFileID         string                   `json:"coverFileId" gorm:"column:cover_file_id;type:varchar(64);not null;index;comment:封面文件ID;"`             // 封面文件ID
	CoverURL            string                   `json:"coverUrl" gorm:"column:cover_url;type:varchar(512);not null;comment:封面地址;"`                           // 封面地址
	CategoryID          uint                     `json:"categoryId" gorm:"column:category_id;type:bigint unsigned;not null;index;comment:分类ID;"`              // 分类ID
	Category            ContentCategory          `json:"category" gorm:"foreignKey:CategoryID"`                                                               // 分类关联数据
	Serving             int                      `json:"serving" gorm:"column:serving;type:int;not null;default:1;comment:默认份数;"`                             // 默认份数
	Description         *string                  `json:"description" gorm:"column:description;type:varchar(180);default:null;comment:说明;"`                    // 说明
	Status              string                   `json:"status" gorm:"column:status;type:varchar(20);not null;default:draft;index;comment:状态;"`               // 状态
	RecommendationCount int                      `json:"recommendationCount" gorm:"column:recommendation_count;type:int;not null;default:0;comment:推荐记录数量;"`  // 推荐记录数量
	Version             int64                    `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                          // 数据版本
	CreatedByID         uint                     `json:"createdById" gorm:"column:created_by_id;type:bigint unsigned;not null;index;comment:创建管理员ID;"`        // 创建管理员ID
	CreatedByUsername   string                   `json:"createdByUsername" gorm:"column:created_by_username;type:varchar(120);not null;comment:创建管理员用户名;"`    // 创建管理员用户名
	CreatedByNickname   *string                  `json:"createdByNickname" gorm:"column:created_by_nickname;type:varchar(120);default:null;comment:创建管理员昵称;"` // 创建管理员昵称
	UpdatedByID         uint                     `json:"updatedById" gorm:"column:updated_by_id;type:bigint unsigned;not null;index;comment:更新管理员ID;"`        // 更新管理员ID
	UpdatedByUsername   string                   `json:"updatedByUsername" gorm:"column:updated_by_username;type:varchar(120);not null;comment:更新管理员用户名;"`    // 更新管理员用户名
	UpdatedByNickname   *string                  `json:"updatedByNickname" gorm:"column:updated_by_nickname;type:varchar(120);default:null;comment:更新管理员昵称;"` // 更新管理员昵称
	DeletedReason       *string                  `json:"deletedReason" gorm:"column:deleted_reason;type:varchar(200);default:null;comment:删除原因;"`             // 删除原因
	DeletedByID         *uint                    `json:"deletedById" gorm:"column:deleted_by_id;type:bigint unsigned;index;default:null;comment:删除管理员ID;"`    // 删除管理员ID
	Ingredients         []OfficialDishIngredient `json:"ingredients" gorm:"foreignKey:DishID"`                                                                // 食材列表
	Steps               []OfficialDishStep       `json:"steps" gorm:"foreignKey:DishID"`                                                                      // 步骤列表
	Tags                []ContentTag             `json:"tags" gorm:"many2many:of_off_tags;joinForeignKey:DishID;joinReferences:TagID"`                        // 标签关联数据
}

// TableName 指定OfficialDish对应的数据表名。
func (OfficialDish) TableName() string { return "of_off_dishes" }

// OfficialDishIngredient 表示官方菜品食材。
type OfficialDishIngredient struct {
	global.GVA_MODEL              // GVA基础模型字段
	DishID           uint         `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:官方菜品ID;"`   // 官方菜品ID
	Name             string       `json:"name" gorm:"column:name;type:varchar(120);not null;comment:名称;"`                     // 名称
	Quantity         *string      `json:"quantity" gorm:"column:quantity;type:varchar(40);default:null;comment:用量;"`          // 用量
	UnitID           *uint        `json:"unitId" gorm:"column:unit_id;type:bigint unsigned;index;default:null;comment:单位ID;"` // 单位ID
	Unit             *ContentUnit `json:"unit" gorm:"foreignKey:UnitID"`                                                      // 单位关联数据
	Note             *string      `json:"note" gorm:"column:note;type:varchar(300);default:null;comment:备注;"`                 // 备注
	SortOrder        int          `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"`  // 排序值
}

// TableName 指定OfficialDishIngredient对应的数据表名。
func (OfficialDishIngredient) TableName() string { return "of_off_ingredients" }

// OfficialDishStep 表示官方菜品步骤。
type OfficialDishStep struct {
	global.GVA_MODEL         // GVA基础模型字段
	DishID           uint    `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;not null;index;comment:官方菜品ID;"`  // 官方菜品ID
	Description      string  `json:"description" gorm:"column:description;type:varchar(2000);not null;comment:说明;"`     // 说明
	ImageURL         *string `json:"imageUrl" gorm:"column:image_url;type:varchar(512);default:null;comment:图片地址;"`     // 图片地址
	SortOrder        int     `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;index;comment:排序值;"` // 排序值
}

// TableName 指定OfficialDishStep对应的数据表名。
func (OfficialDishStep) TableName() string { return "of_off_steps" }

// OfficialDishTag 表示官方菜品与标签关联。
type OfficialDishTag struct {
	DishID    uint      `json:"dishId" gorm:"column:dish_id;type:bigint unsigned;primaryKey;not null;comment:官方菜品ID;"` // 官方菜品ID
	TagID     uint      `json:"tagId" gorm:"column:tag_id;type:bigint unsigned;primaryKey;not null;comment:标签ID;"`     // 标签ID
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`               // 创建时间
}

// TableName 指定OfficialDishTag对应的数据表名。
func (OfficialDishTag) TableName() string { return "of_off_tags" }

package orderfood

import (
	"time"

	"gorm.io/datatypes"
)

// StandardIngredient 表示标准食材索引。
type StandardIngredient struct {
	ID                string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`              // 主键ID
	Name              string         `json:"name" gorm:"column:name;type:varchar(80);not null;uniqueIndex;comment:名称;"`           // 名称
	AliasesJSON       datatypes.JSON `json:"-" gorm:"column:aliases_json;type:json;not null;comment:别名集合JSON;"`                   // 别名集合JSON
	Enabled           bool           `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;index;comment:是否启用;"` // 是否启用
	Version           int64          `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`          // 数据版本
	UpdatedByID       uint           `json:"-" gorm:"column:updated_by_id;type:bigint unsigned;not null;comment:更新管理员ID;"`        // 更新管理员ID
	UpdatedByUsername string         `json:"-" gorm:"column:updated_by_username;type:varchar(80);not null;comment:更新管理员用户名;"`     // 更新管理员用户名
	UpdatedByNickname *string        `json:"-" gorm:"column:updated_by_nickname;type:varchar(80);default:null;comment:更新管理员昵称;"`  // 更新管理员昵称
	LastChangeReason  string         `json:"-" gorm:"column:last_change_reason;type:varchar(200);not null;comment:最近变更原因;"`       // 最近变更原因
	CreatedAt         time.Time      `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`             // 创建时间
	UpdatedAt         time.Time      `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`       // 更新时间
}

// TableName 指定StandardIngredient对应的数据表名。
func (StandardIngredient) TableName() string {
	return "of_std_ingredients"
}

// StandardDishIndex 表示标准菜品索引。
type StandardDishIndex struct {
	ID                string          `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                 // 主键ID
	Name              string          `json:"name" gorm:"column:name;type:varchar(120);not null;uniqueIndex;comment:名称;"`             // 名称
	AliasesJSON       datatypes.JSON  `json:"-" gorm:"column:aliases_json;type:json;not null;comment:别名集合JSON;"`                      // 别名集合JSON
	Cuisine           string          `json:"cuisine" gorm:"column:cuisine;type:varchar(80);not null;index;comment:菜系;"`              // 菜系
	CategoryID        uint            `json:"categoryId" gorm:"column:category_id;type:bigint unsigned;not null;index;comment:分类ID;"` // 分类ID
	Category          ContentCategory `json:"category" gorm:"foreignKey:CategoryID"`                                                  // 分类关联数据
	SourceName        string          `json:"sourceName" gorm:"column:source_name;type:varchar(160);not null;comment:来源名称;"`          // 来源名称
	SourceURL         *string         `json:"sourceUrl" gorm:"column:source_url;type:varchar(500);default:null;comment:来源地址;"`        // 来源地址
	Enabled           bool            `json:"enabled" gorm:"column:enabled;type:tinyint(1) unsigned;not null;index;comment:是否启用;"`    // 是否启用
	Version           int64           `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`             // 数据版本
	UpdatedByID       uint            `json:"-" gorm:"column:updated_by_id;type:bigint unsigned;not null;comment:更新管理员ID;"`           // 更新管理员ID
	UpdatedByUsername string          `json:"-" gorm:"column:updated_by_username;type:varchar(80);not null;comment:更新管理员用户名;"`        // 更新管理员用户名
	UpdatedByNickname *string         `json:"-" gorm:"column:updated_by_nickname;type:varchar(80);default:null;comment:更新管理员昵称;"`     // 更新管理员昵称
	LastChangeReason  string          `json:"-" gorm:"column:last_change_reason;type:varchar(200);not null;comment:最近变更原因;"`          // 最近变更原因
	CreatedAt         time.Time       `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                // 创建时间
	UpdatedAt         time.Time       `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;index;comment:更新时间;"`          // 更新时间

	Ingredients []StandardDishIngredient `json:"ingredients" gorm:"foreignKey:DishID"` // 食材列表
}

// TableName 指定StandardDishIndex对应的数据表名。
func (StandardDishIndex) TableName() string {
	return "of_std_dish_index"
}

// StandardDishIngredient 表示标准菜品与食材关联。
type StandardDishIngredient struct {
	DishID       string             `json:"dishId" gorm:"column:dish_id;type:varchar(64);primaryKey;not null;comment:菜品ID;"`             // 菜品ID
	IngredientID string             `json:"ingredientId" gorm:"column:ingredient_id;type:varchar(64);primaryKey;not null;comment:食材ID;"` // 食材ID
	Ingredient   StandardIngredient `json:"ingredient" gorm:"foreignKey:IngredientID"`                                                   // 食材关联数据
	Required     bool               `json:"required" gorm:"column:required;type:tinyint(1) unsigned;not null;comment:是否必需;"`             // 是否必需
	SortOrder    int                `json:"sortOrder" gorm:"column:sort_order;type:int;not null;default:1;comment:排序值;"`                 // 排序值
	CreatedAt    time.Time          `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                     // 创建时间
}

// TableName 指定StandardDishIngredient对应的数据表名。
func (StandardDishIngredient) TableName() string {
	return "of_std_dish_ingredients"
}

// SuggestionValidationPolicy 表示当前生效的生成结果兜底校验策略。
type SuggestionValidationPolicy struct {
	SingletonKey             string    `json:"-" gorm:"column:singleton_key;type:varchar(32);primaryKey;not null;comment:单例记录键;"`                                         // 单例记录键
	Version                  int64     `json:"version" gorm:"column:version;type:bigint;not null;default:1;comment:数据版本;"`                                                // 数据版本
	CatalogValidationEnabled bool      `json:"catalogValidationEnabled" gorm:"column:catalog_validation_enabled;type:tinyint(1) unsigned;not null;comment:是否启用菜品索引兜底校验;"` // 是否启用菜品索引兜底校验
	AppliedByID              uint      `json:"-" gorm:"column:applied_by_id;type:bigint unsigned;not null;comment:应用管理员ID;"`                                              // 应用管理员ID
	AppliedByUsername        string    `json:"-" gorm:"column:applied_by_username;type:varchar(80);not null;comment:应用管理员用户名;"`                                           // 应用管理员用户名
	AppliedByNickname        *string   `json:"-" gorm:"column:applied_by_nickname;type:varchar(80);default:null;comment:应用管理员昵称;"`                                        // 应用管理员昵称
	AppliedAt                time.Time `json:"appliedAt" gorm:"column:applied_at;type:datetime;not null;index;comment:应用时间;"`                                             // 应用时间
	Reason                   string    `json:"reason" gorm:"column:reason;type:varchar(200);not null;comment:原因;"`                                                        // 原因
}

// TableName 指定SuggestionValidationPolicy对应的数据表名。
func (SuggestionValidationPolicy) TableName() string {
	return "of_suggest_policy"
}

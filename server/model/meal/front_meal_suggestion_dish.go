package meal

import (
	"gorm.io/datatypes"
	"time"
)

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

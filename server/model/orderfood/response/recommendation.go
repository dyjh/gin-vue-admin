package response

import "time"

// DiscoverableDishSummary 表示可发现候选菜品摘要。
type DiscoverableDishSummary struct {
	ID                     string             `json:"id"`                     // 菜品ID
	CoverURL               string             `json:"coverUrl"`               // 封面地址
	Name                   string             `json:"name"`                   // 菜品名称
	Category               CatalogReference   `json:"category"`               // 分类
	Tags                   []CatalogReference `json:"tags"`                   // 标签列表
	Author                 UserReference      `json:"author"`                 // 作者
	DiscoverableAt         time.Time          `json:"discoverableAt"`         // 允许被发现时间
	VoteCount              int64              `json:"voteCount"`              // 饭局点选次数
	CopyCount              int64              `json:"copyCount"`              // 被复制次数
	Selected               bool               `json:"selected"`               // 是否已创建推荐记录
	RecommendationPosition *string            `json:"recommendationPosition"` // 推荐位置
	SourceLocked           bool               `json:"sourceLocked"`           // 来源字段是否锁定
	Version                int                `json:"version"`                // 数据版本
}

// DiscoverableDishDetail 表示可发现候选菜品详情。
type DiscoverableDishDetail struct {
	DiscoverableDishSummary                          // 可发现候选菜品摘要
	Description             *string                  `json:"description"`       // 说明
	Serving                 int                      `json:"serving"`           // 默认份数
	Ingredients             []DishIngredient         `json:"ingredients"`       // 食材列表
	Steps                   []DishStep               `json:"steps"`             // 步骤列表
	MediaReviewStatus       string                   `json:"mediaReviewStatus"` // 图片审核状态
	Recommendation          *RecommendationSummary   `json:"recommendation"`    // 推荐摘要
	ModerationSummary       *ModerationRecordSummary `json:"moderationSummary"` // 图片审核摘要
	UpdatedAt               time.Time                `json:"updatedAt"`         // 更新时间
}

// RecommendationSummary 表示推荐精选摘要。
type RecommendationSummary struct {
	ID                string     `json:"id"`                // 推荐ID
	SourceType        string     `json:"sourceType"`        // 来源类型
	SourceDishID      string     `json:"sourceDishId"`      // 来源菜品ID
	DishName          string     `json:"dishName"`          // 菜品名称
	CoverURL          string     `json:"coverUrl"`          // 封面地址
	SourceAuthorLabel string     `json:"sourceAuthorLabel"` // 来源作者说明
	Position          string     `json:"position"`          // 推荐位置
	SortOrder         int        `json:"sortOrder"`         // 排序值
	Status            string     `json:"status"`            // 推荐状态
	CopyCount         int64      `json:"copyCount"`         // 复制次数
	Version           int64      `json:"version"`           // 数据版本
	PublishedAt       *time.Time `json:"publishedAt"`       // 发布时间
	OfflineAt         *time.Time `json:"offlineAt"`         // 下线时间
	CreatedAt         time.Time  `json:"createdAt"`         // 创建时间
	UpdatedAt         time.Time  `json:"updatedAt"`         // 更新时间
}

// RecommendationDetail 表示推荐精选详情。
type RecommendationDetail struct {
	RecommendationSummary                        // 推荐精选摘要
	DisplayNote             *string              `json:"displayNote"`             // 展示说明
	SourceAvailable         bool                 `json:"sourceAvailable"`         // 来源是否仍可发布
	SourceUnavailableReason *string              `json:"sourceUnavailableReason"` // 来源不可用原因
	Dish                    interface{}          `json:"dish"`                    // 来源菜品详情
	CreatedBy               AdministratorSummary `json:"createdBy"`               // 创建管理员
	UpdatedBy               AdministratorSummary `json:"updatedBy"`               // 更新管理员
}

// RecommendationDeleteResult 表示推荐草稿删除结果。
type RecommendationDeleteResult struct {
	Deleted bool `json:"deleted"` // 是否删除成功
}

// RecommendationSortOrderResult 表示推荐批量排序结果。
type RecommendationSortOrderResult struct {
	Updated int `json:"updated"` // 更新数量
}

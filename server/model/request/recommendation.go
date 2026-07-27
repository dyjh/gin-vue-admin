package request

import "time"

// DiscoverableDishListQuery 表示可发现候选菜品列表查询条件。
type DiscoverableDishListQuery struct {
	Page             int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                  // 页码
	PageSize         int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                // 每页数量
	DishID           string     `form:"dishId" json:"dishId" binding:"omitempty,max=64"`                                             // 菜品ID
	Keyword          string     `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                                           // 菜品名称关键词
	AuthorID         string     `form:"authorId" json:"authorId" binding:"omitempty,max=64"`                                         // 作者用户ID
	AuthorKeyword    string     `form:"authorKeyword" json:"authorKeyword" binding:"omitempty,max=60" checksql:"false"`              // 作者用户ID或昵称关键词
	CategoryID       string     `form:"categoryId" json:"categoryId" binding:"omitempty,max=64"`                                     // 分类ID
	TagIDs           []string   `form:"tagIds" json:"tagIds" binding:"omitempty,max=20,unique,dive,max=64"`                          // 标签ID列表
	Selected         *bool      `form:"selected" json:"selected"`                                                                    // 是否已创建推荐记录
	DiscoverableFrom *time.Time `form:"discoverableFrom" json:"discoverableFrom" time_format:"2006-01-02T15:04:05Z07:00"`            // 允许被发现起始时间
	DiscoverableTo   *time.Time `form:"discoverableTo" json:"discoverableTo" time_format:"2006-01-02T15:04:05Z07:00"`                // 允许被发现结束时间
	SortBy           string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=discoverableAt voteCount copyCount createdAt"` // 排序字段
	SortOrder        string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                               // 排序方向
}

// ApplyDefaults 补齐可发现候选菜品列表的分页和排序默认值。
func (query *DiscoverableDishListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "discoverableAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// DiscoverableDishDetailQuery 表示可发现候选菜品详情查询条件。
type DiscoverableDishDetailQuery struct {
	IncludeModerationSummary bool `form:"includeModerationSummary" json:"includeModerationSummary"` // 是否包含图片审核摘要
}

// RecommendationListQuery 表示推荐精选列表查询条件。
type RecommendationListQuery struct {
	Page         int        `form:"page" json:"page" binding:"omitempty,min=1"`                                     // 页码
	PageSize     int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                   // 每页数量
	Keyword      string     `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                              // 菜品名称关键词
	SourceType   string     `form:"sourceType" json:"sourceType" binding:"omitempty,oneof=official creator"`        // 来源类型
	SourceDishID string     `form:"sourceDishId" json:"sourceDishId" binding:"omitempty,max=64"`                    // 来源菜品ID
	Status       string     `form:"status" json:"status" binding:"omitempty,oneof=draft published offline"`         // 推荐状态
	Position     string     `form:"position" json:"position" binding:"omitempty,oneof=home_featured"`               // 推荐位置
	CreatedFrom  *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`         // 创建起始时间
	CreatedTo    *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`             // 创建结束时间
	SortBy       string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=sortOrder publishedAt createdAt"` // 排序字段
	SortOrder    string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                  // 排序方向
}

// ApplyDefaults 补齐推荐精选列表的分页和排序默认值。
func (query *RecommendationListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "sortOrder"
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
}

// RecommendationCreateInput 表示创建推荐草稿的请求参数。
type RecommendationCreateInput struct {
	SourceType   string  `json:"sourceType" binding:"required,oneof=official creator"` // 来源类型
	SourceDishID string  `json:"sourceDishId" binding:"required,max=64"`               // 来源菜品ID
	Position     string  `json:"position" binding:"required,oneof=home_featured"`      // 推荐位置
	SortOrder    int     `json:"sortOrder" binding:"required,min=1"`                   // 排序值
	DisplayNote  *string `json:"displayNote" binding:"omitempty,max=160"`              // 展示说明
}

// RecommendationUpdateInput 表示编辑推荐信息的请求参数。
type RecommendationUpdateInput struct {
	Position        string  `json:"position" binding:"required,oneof=home_featured"` // 推荐位置
	SortOrder       int     `json:"sortOrder" binding:"required,min=1"`              // 排序值
	DisplayNote     *string `json:"displayNote" binding:"omitempty,max=160"`         // 展示说明
	ExpectedVersion int64   `json:"expectedVersion" binding:"required,min=1"`        // 预期数据版本
}

// RecommendationVersionInput 表示仅包含预期版本的推荐变更参数。
type RecommendationVersionInput struct {
	ExpectedVersion int64 `json:"expectedVersion" binding:"required,min=1"` // 预期数据版本
}

// RecommendationOfflineInput 表示推荐下线参数。
type RecommendationOfflineInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200" checksql:"false"` // 下线原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"`                 // 预期数据版本
}

// RecommendationSortOrderInput 表示推荐批量排序参数。
type RecommendationSortOrderInput struct {
	Items []CatalogSortOrderItem `json:"items" binding:"required,min=1,max=100,dive"` // 推荐排序变更列表
}

// RecommendationIDPath 表示推荐ID路径参数。
type RecommendationIDPath struct {
	RecommendationID string `uri:"recommendationId" json:"recommendationId" binding:"required,max=64"` // 推荐ID
}

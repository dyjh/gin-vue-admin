package request

// CatalogItemListQuery 表示分类、标签或单位的分页查询条件。
type CatalogItemListQuery struct {
	Page      int    `form:"page" json:"page" binding:"omitempty,min=1"`                                   // 页码
	PageSize  int    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                 // 每页数量
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=40"`                            // 名称关键词
	Enabled   *bool  `form:"enabled" json:"enabled"`                                                       // 是否启用
	SortBy    string `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=sortOrder createdAt updatedAt"` // 排序字段
	SortOrder string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                // 排序方向
}

// ApplyDefaults 补齐基础数据列表的分页和排序默认值。
func (query *CatalogItemListQuery) ApplyDefaults() {
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

// CatalogItemCreateInput 表示新增分类、标签或单位的请求参数。
type CatalogItemCreateInput struct {
	Name      string `json:"name" binding:"required,max=40"`     // 名称
	SortOrder int    `json:"sortOrder" binding:"required,min=1"` // 排序值
	Enabled   *bool  `json:"enabled" binding:"required"`         // 是否启用
}

// CatalogItemUpdateInput 表示编辑分类、标签或单位的请求参数。
type CatalogItemUpdateInput struct {
	Name            string `json:"name" binding:"required,max=40"`           // 名称
	SortOrder       int    `json:"sortOrder" binding:"required,min=1"`       // 排序值
	Enabled         *bool  `json:"enabled" binding:"required"`               // 是否启用
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"` // 预期数据版本
}

// CatalogDeleteInput 表示删除分类、标签或单位的请求参数。
type CatalogDeleteInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200" checksql:"false"` // 删除原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"`                 // 预期数据版本
}

// CatalogSortOrderItem 表示一条基础数据排序变更。
type CatalogSortOrderItem struct {
	ID              string `json:"id" binding:"required,max=64"`             // 资源ID
	SortOrder       int    `json:"sortOrder" binding:"required,min=1"`       // 新排序值
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"` // 预期数据版本
}

// CatalogSortOrderInput 表示批量调整基础数据排序的请求参数。
type CatalogSortOrderInput struct {
	Items []CatalogSortOrderItem `json:"items" binding:"required,min=1,max=100,dive"` // 排序变更列表
}

// CategoryIDPath 表示分类ID路径参数。
type CategoryIDPath struct {
	CategoryID string `uri:"categoryId" json:"categoryId" binding:"required,max=64"` // 分类ID
}

// TagIDPath 表示标签ID路径参数。
type TagIDPath struct {
	TagID string `uri:"tagId" json:"tagId" binding:"required,max=64"` // 标签ID
}

// UnitIDPath 表示单位ID路径参数。
type UnitIDPath struct {
	UnitID string `uri:"unitId" json:"unitId" binding:"required,max=64"` // 单位ID
}

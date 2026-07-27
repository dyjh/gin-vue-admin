package request

// SuggestionCatalogEmptyQuery 表示推荐菜索引空查询条件。
type SuggestionCatalogEmptyQuery struct{}

// SuggestionValidationPolicyUpdateInput 表示生成结果校验策略直接保存参数。
type SuggestionValidationPolicyUpdateInput struct {
	CatalogValidationEnabled bool   `json:"catalogValidationEnabled"`                                 // 是否启用标准菜品索引兜底校验
	Reason                   string `json:"reason" binding:"required,min=2,max=200" checksql:"false"` // 修改原因
	ExpectedVersion          int64  `json:"expectedVersion" binding:"required,min=1"`                 // 当前配置版本
}

// StandardCatalogListQuery 表示标准索引列表查询条件。
type StandardCatalogListQuery struct {
	Page      int    `form:"page" json:"page" binding:"omitempty,min=1"`                         // 页码
	PageSize  int    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`       // 每页数量
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=80" checksql:"false"` // 关键词
	Enabled   *bool  `form:"enabled" json:"enabled"`                                             // 是否启用
	SortOrder string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`      // 排序方向
}

// ApplyDefaults 补齐分页、筛选和排序默认值。
func (query *StandardCatalogListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
}

// StandardCatalogIDPath 表示标准索引ID路径参数。
type StandardCatalogIDPath struct {
	ID string `uri:"id" json:"id" binding:"required,max=64"` // ID
}

// StandardIngredientInput 表示标准食材输入参数。
type StandardIngredientInput struct {
	Name            string   `json:"name" binding:"required,max=80" checksql:"false"`          // 名称
	Aliases         []string `json:"aliases" binding:"max=20,dive,max=80" checksql:"false"`    // 别名列表
	Enabled         *bool    `json:"enabled" binding:"required"`                               // 是否启用
	ExpectedVersion *int64   `json:"expectedVersion" binding:"omitempty,min=1"`                // 预期版本
	Reason          string   `json:"reason" binding:"required,min=4,max=200" checksql:"false"` // 原因
}

// StandardDishIngredientInput 表示标准菜品食材输入参数。
type StandardDishIngredientInput struct {
	IngredientID string `json:"ingredientId" binding:"required,max=64"`     // 食材ID
	Required     *bool  `json:"required" binding:"required"`                // 是否必需
	SortOrder    int    `json:"sortOrder" binding:"required,min=1,max=200"` // 排序值
}

// StandardDishInput 表示标准菜品输入参数。
type StandardDishInput struct {
	Name            string                        `json:"name" binding:"required,max=120" checksql:"false"`           // 名称
	Aliases         []string                      `json:"aliases" binding:"max=20,dive,max=120" checksql:"false"`     // 别名列表
	Cuisine         string                        `json:"cuisine" binding:"required,max=80" checksql:"false"`         // 菜系
	CategoryID      string                        `json:"categoryId" binding:"required,max=64"`                       // 分类ID
	SourceName      string                        `json:"sourceName" binding:"required,max=160" checksql:"false"`     // 来源名称
	SourceURL       *string                       `json:"sourceUrl" binding:"omitempty,url,max=500" checksql:"false"` // 来源地址
	Enabled         *bool                         `json:"enabled" binding:"required"`                                 // 是否启用
	Ingredients     []StandardDishIngredientInput `json:"ingredients" binding:"required,min=1,max=100,dive"`          // 食材列表
	ExpectedVersion *int64                        `json:"expectedVersion" binding:"omitempty,min=1"`                  // 预期版本
	Reason          string                        `json:"reason" binding:"required,min=4,max=200" checksql:"false"`   // 原因
}

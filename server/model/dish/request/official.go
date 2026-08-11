package request

import (
	"time"
)

// OfficialDishListQuery 表示官方菜品分页查询条件。
type OfficialDishListQuery struct {
	Page        int        `form:"page" json:"page" binding:"omitempty,min=1"`                              // 页码
	PageSize    int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`            // 每页数量
	Keyword     string     `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                       // 菜品名称关键词
	CategoryID  string     `form:"categoryId" json:"categoryId" binding:"omitempty,max=64"`                 // 分类ID
	TagIDs      []string   `form:"tagIds" json:"tagIds" binding:"omitempty,max=3,unique,dive,max=64"`       // 标签ID列表
	Status      string     `form:"status" json:"status" binding:"omitempty,oneof=draft usable"`             // 菜品状态
	CreatedFrom *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`  // 创建起始时间
	CreatedTo   *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`      // 创建结束时间
	SortBy      string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=updatedAt createdAt name"` // 排序字段
	SortOrder   string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`           // 排序方向
}

// ApplyDefaults 补齐官方菜品列表的分页和排序默认值。
func (query *OfficialDishListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "updatedAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

// OfficialDishIngredientInput 表示官方菜品食材输入参数。
type OfficialDishIngredientInput struct {
	Name      string  `json:"name" binding:"required,max=30" checksql:"false"`      // 食材名称
	Quantity  *string `json:"quantity" binding:"omitempty,max=20" checksql:"false"` // 食材用量
	UnitID    *string `json:"unitId" binding:"omitempty,max=64"`                    // 单位ID
	Note      *string `json:"note" binding:"omitempty,max=80" checksql:"false"`     // 备注
	SortOrder int     `json:"sortOrder" binding:"required,min=1"`                   // 排序值
}

// OfficialDishStepInput 表示官方菜品步骤输入参数。
type OfficialDishStepInput struct {
	Description string  `json:"description" binding:"required,max=500" checksql:"false"` // 步骤说明
	ImageURL    *string `json:"imageUrl" binding:"omitempty,max=512"`                    // 步骤图片地址
	SortOrder   int     `json:"sortOrder" binding:"required,min=1"`                      // 排序值
}

// OfficialDishCreateInput 表示创建官方菜品的请求参数。
type OfficialDishCreateInput struct {
	Name        string                        `json:"name" binding:"required,max=40" checksql:"false"`          // 菜品名称
	CategoryID  string                        `json:"categoryId" binding:"required,max=64"`                     // 分类ID
	TagIDs      []string                      `json:"tagIds" binding:"max=3,unique,dive,max=64"`                // 标签ID列表
	Serving     int                           `json:"serving" binding:"required,min=1,max=20"`                  // 默认份数
	Description *string                       `json:"description" binding:"omitempty,max=180" checksql:"false"` // 菜品说明
	CoverFileID string                        `json:"coverFileId" binding:"required,max=64"`                    // 封面文件ID
	Status      string                        `json:"status" binding:"required,oneof=draft usable"`             // 菜品状态
	Ingredients []OfficialDishIngredientInput `json:"ingredients" binding:"omitempty,max=100,dive"`             // 食材列表，可用状态至少一项
	Steps       []OfficialDishStepInput       `json:"steps" binding:"omitempty,max=100,dive"`                   // 步骤列表，可用状态至少一项
}

// OfficialDishUpdateInput 表示编辑官方菜品的请求参数。
type OfficialDishUpdateInput struct {
	OfficialDishCreateInput       // 官方菜品内容
	ExpectedVersion         int64 `json:"expectedVersion" binding:"required,min=1"` // 预期数据版本
}

// OfficialDishDeleteInput 表示软删除官方菜品的请求参数。
type OfficialDishDeleteInput struct {
	Reason          string `json:"reason" binding:"required,min=4,max=200" checksql:"false"` // 删除原因
	ExpectedVersion int64  `json:"expectedVersion" binding:"required,min=1"`                 // 预期数据版本
}

// OfficialDishIDPath 表示官方菜品ID路径参数。
type OfficialDishIDPath struct {
	DishID string `uri:"dishId" json:"dishId" binding:"required,max=64"` // 官方菜品ID
}

// OfficialDishCoverListQuery 表示官方菜品封面分页查询条件。
type OfficialDishCoverListQuery struct {
	Page      int    `form:"page" json:"page" binding:"omitempty,min=1"`                        // 页码
	PageSize  int    `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`      // 每页数量
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=60"`                 // 原始文件名关键词
	SortBy    string `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt fileSize"` // 排序字段
	SortOrder string `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`     // 排序方向
}

// ApplyDefaults 补齐官方菜品封面列表的分页和排序默认值。
func (query *OfficialDishCoverListQuery) ApplyDefaults() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "createdAt"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}
}

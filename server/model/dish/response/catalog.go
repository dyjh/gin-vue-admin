package response

import (
	"time"
)

// CatalogItem 表示分类、标签或单位的管理端响应数据。
type CatalogItem struct {
	ID             string    `json:"id"`             // 资源ID
	Name           string    `json:"name"`           // 名称
	SortOrder      int       `json:"sortOrder"`      // 排序值
	Enabled        bool      `json:"enabled"`        // 是否启用
	ReferenceCount int64     `json:"referenceCount"` // 业务引用数量
	Version        int64     `json:"version"`        // 数据版本
	CreatedAt      time.Time `json:"createdAt"`      // 创建时间
	UpdatedAt      time.Time `json:"updatedAt"`      // 更新时间
}

// CatalogDeleteResult 表示基础数据删除结果。
type CatalogDeleteResult struct {
	Deleted bool `json:"deleted"` // 是否删除成功
}

// CatalogSortOrderResult 表示基础数据批量排序结果。
type CatalogSortOrderResult struct {
	Updated int `json:"updated"` // 更新数量
}

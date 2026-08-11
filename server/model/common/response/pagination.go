package response

// Page 表示分页响应数据。
type Page[T any] struct {
	Page     int   `json:"page"`     // 页码
	PageSize int   `json:"pageSize"` // 每页数量
	Total    int64 `json:"total"`    // 总数量
	List     []T   `json:"list"`     // 列表
}

// AuditedPage 表示包含访问审计记录ID的分页响应数据。
type AuditedPage[T any] struct {
	Page          int    `json:"page"`          // 页码
	PageSize      int    `json:"pageSize"`      // 每页数量
	Total         int64  `json:"total"`         // 总数量
	List          []T    `json:"list"`          // 列表
	AccessAuditID string `json:"accessAuditId"` // 敏感访问审计记录ID
}

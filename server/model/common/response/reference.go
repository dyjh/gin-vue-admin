package response

// UserReference 表示用户引用响应数据。
type UserReference struct {
	ID        string  `json:"id"`        // ID
	Nickname  string  `json:"nickname"`  // 昵称
	AvatarURL *string `json:"avatarUrl"` // 头像地址
	Status    string  `json:"status"`    // 状态
}

// AdministratorSummary 表示管理员摘要响应数据。
type AdministratorSummary struct {
	ID       string  `json:"id"`       // ID
	Username string  `json:"username"` // 用户名
	Nickname *string `json:"nickname"` // 昵称
}

// CatalogReference 表示索引引用响应数据。
type CatalogReference struct {
	ID      string `json:"id"`      // ID
	Name    string `json:"name"`    // 名称
	Enabled bool   `json:"enabled"` // 是否启用
}

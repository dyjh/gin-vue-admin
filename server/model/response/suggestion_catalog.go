package response

import "time"

// SuggestionValidationPolicyConfig 表示当前生效的推荐菜校验策略。
type SuggestionValidationPolicyConfig struct {
	Version                  int64                `json:"version"`                  // 版本
	CatalogValidationEnabled bool                 `json:"catalogValidationEnabled"` // 是否启用标准菜品索引兜底校验
	RetryCount               int                  `json:"retryCount"`               // 重试数量
	RefundOnFailure          bool                 `json:"refundOnFailure"`          // 全部尝试失败时是否退积分
	UpdatedBy                AdministratorSummary `json:"updatedBy"`                // 更新人
	UpdatedAt                time.Time            `json:"updatedAt"`                // 更新时间
	Reason                   string               `json:"reason"`                   // 原因
}

// SuggestionCatalogWorkspace 表示推荐菜索引配置响应数据。
type SuggestionCatalogWorkspace struct {
	Policy                 SuggestionValidationPolicyConfig `json:"policy"`                 // 当前生效策略
	DishCount              int64                            `json:"dishCount"`              // 菜品数量
	IngredientCount        int64                            `json:"ingredientCount"`        // 食材数量
	EnabledDishCount       int64                            `json:"enabledDishCount"`       // 已启用菜品数量
	EnabledIngredientCount int64                            `json:"enabledIngredientCount"` // 已启用食材数量
	Categories             []CatalogReference               `json:"categories"`             // 分类列表
	Readiness              SuggestionCatalogReadiness       `json:"readiness"`              // 索引就绪状态
}

// SuggestionCatalogReadiness 表示标准菜品索引能否安全开启生成结果校验。
type SuggestionCatalogReadiness struct {
	Ready                    bool     `json:"ready"`                    // 是否满足开启条件
	RequiredMinimumDishCount int64    `json:"requiredMinimumDishCount"` // 最低启用菜品数量
	EnabledDishCount         int64    `json:"enabledDishCount"`         // 已启用菜品数量
	InvalidReferenceCount    int64    `json:"invalidReferenceCount"`    // 无效分类或食材引用数量
	NameCollisionCount       int64    `json:"nameCollisionCount"`       // 名称或别名冲突数量
	Blockers                 []string `json:"blockers"`                 // 阻止开启校验的原因
}

// StandardIngredient 表示标准食材响应数据。
type StandardIngredient struct {
	ID        string               `json:"id"`        // ID
	Name      string               `json:"name"`      // 名称
	Aliases   []string             `json:"aliases"`   // 别名列表
	Enabled   bool                 `json:"enabled"`   // 是否启用
	Version   int64                `json:"version"`   // 版本
	UpdatedBy AdministratorSummary `json:"updatedBy"` // 更新人
	UpdatedAt time.Time            `json:"updatedAt"` // 更新时间
}

// StandardDishIngredient 表示标准菜品食材响应数据。
type StandardDishIngredient struct {
	IngredientID string `json:"ingredientId"` // 食材ID
	Name         string `json:"name"`         // 名称
	Required     bool   `json:"required"`     // 是否必需
	SortOrder    int    `json:"sortOrder"`    // 排序值
}

// StandardDish 表示标准菜品响应数据。
type StandardDish struct {
	ID          string                   `json:"id"`          // ID
	Name        string                   `json:"name"`        // 名称
	Aliases     []string                 `json:"aliases"`     // 别名列表
	Cuisine     string                   `json:"cuisine"`     // 菜系
	CategoryID  string                   `json:"categoryId"`  // 分类ID
	Category    string                   `json:"category"`    // 分类
	SourceName  string                   `json:"sourceName"`  // 来源名称
	SourceURL   *string                  `json:"sourceUrl"`   // 来源地址
	Enabled     bool                     `json:"enabled"`     // 是否启用
	Ingredients []StandardDishIngredient `json:"ingredients"` // 食材列表
	Version     int64                    `json:"version"`     // 版本
	UpdatedBy   AdministratorSummary     `json:"updatedBy"`   // 更新人
	UpdatedAt   time.Time                `json:"updatedAt"`   // 更新时间
}

package response

import "time"

// OfficialDishSummary 表示官方菜品列表摘要。
type OfficialDishSummary struct {
	ID                        string               `json:"id"`                        // 官方菜品ID
	Name                      string               `json:"name"`                      // 菜品名称
	CoverURL                  string               `json:"coverUrl"`                  // 封面地址
	Category                  CatalogReference     `json:"category"`                  // 分类
	Tags                      []CatalogReference   `json:"tags"`                      // 标签列表
	Serving                   int                  `json:"serving"`                   // 默认份数
	Status                    string               `json:"status"`                    // 菜品状态
	RecommendationCount       int                  `json:"recommendationCount"`       // 推荐记录数量
	OnlineRecommendationCount int64                `json:"onlineRecommendationCount"` // 当前在线推荐数量
	CopyCount                 int64                `json:"copyCount"`                 // 用户复制次数
	Version                   int64                `json:"version"`                   // 数据版本
	CreatedAt                 time.Time            `json:"createdAt"`                 // 创建时间
	UpdatedAt                 time.Time            `json:"updatedAt"`                 // 更新时间
	UpdatedBy                 AdministratorSummary `json:"updatedBy"`                 // 最后修改管理员
}

// OfficialDishDetail 表示官方菜品详情。
type OfficialDishDetail struct {
	OfficialDishSummary                      // 官方菜品摘要
	Description         *string              `json:"description"` // 菜品说明
	CoverFileID         string               `json:"coverFileId"` // 封面文件ID
	Ingredients         []DishIngredient     `json:"ingredients"` // 食材列表
	Steps               []DishStep           `json:"steps"`       // 步骤列表
	CreatedBy           AdministratorSummary `json:"createdBy"`   // 创建管理员
}

// OfficialDishMutationResult 表示官方菜品编辑结果及其联动下线数量。
type OfficialDishMutationResult struct {
	Dish                       OfficialDishDetail `json:"dish"`                       // 编辑后的官方菜品
	OfflineRecommendationCount int64              `json:"offlineRecommendationCount"` // 自动下线推荐数量
}

// OfficialDishDeleteResult 表示官方菜品软删除结果。
type OfficialDishDeleteResult struct {
	Deleted                    bool  `json:"deleted"`                    // 是否删除成功
	OfflineRecommendationCount int64 `json:"offlineRecommendationCount"` // 自动下线推荐数量
}

// OfficialDishCoverSummary 表示管理端上传的官方菜品封面摘要。
type OfficialDishCoverSummary struct {
	FileID         string               `json:"fileId"`         // 文件ID
	URL            string               `json:"url"`            // 访问地址
	FileName       string               `json:"fileName"`       // 原始文件名
	FileSize       int64                `json:"fileSize"`       // 文件字节数
	MimeType       string               `json:"mimeType"`       // MIME类型
	Width          int                  `json:"width"`          // 图片宽度
	Height         int                  `json:"height"`         // 图片高度
	UploadSource   string               `json:"uploadSource"`   // 上传来源
	SourceScene    string               `json:"sourceScene"`    // 使用场景
	ReviewStatus   string               `json:"reviewStatus"`   // 审核状态
	ResourceStatus string               `json:"resourceStatus"` // 资源状态
	UploadedBy     AdministratorSummary `json:"uploadedBy"`     // 上传管理员
	CreatedAt      time.Time            `json:"createdAt"`      // 上传时间
}

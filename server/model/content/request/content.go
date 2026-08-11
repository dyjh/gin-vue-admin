package request

import (
	"time"
)

// UserDishSearch 表示用户菜品查询条件。
type UserDishSearch struct {
	Page              int        `json:"page" form:"page" binding:"omitempty,min=1"`                                                                       // 页码
	PageSize          int        `json:"pageSize" form:"pageSize" binding:"omitempty,oneof=20 50 100"`                                                     // 每页数量
	Keyword           string     `json:"keyword" form:"keyword" binding:"omitempty,max=60"`                                                                // 关键词
	UserID            string     `json:"userId" form:"userId" binding:"omitempty,max=64"`                                                                  // 用户ID
	UserKeyword       string     `json:"userKeyword" form:"userKeyword" binding:"omitempty,max=60" checksql:"false"`                                       // 用户ID或昵称关键词
	Status            string     `json:"status" form:"status" binding:"omitempty,oneof=draft usable"`                                                      // 状态
	DeletionStatus    string     `json:"deletionStatus" form:"deletionStatus" binding:"omitempty,oneof=active deleted"`                                    // 删除状态
	Discoverable      *bool      `json:"discoverable" form:"discoverable"`                                                                                 // 菜品是否允许被发现
	SourceType        string     `json:"sourceType" form:"sourceType" binding:"omitempty,oneof=manual creator_copy official_copy meal_suggestion_copy"`    // 来源类型
	SourceLocked      *bool      `json:"sourceLocked" form:"sourceLocked"`                                                                                 // 来源字段是否永久锁定
	CategoryID        string     `json:"categoryId" form:"categoryId" binding:"omitempty,max=64"`                                                          // 分类ID
	TagIDs            []string   `json:"tagIds" form:"tagIds" binding:"omitempty,dive,max=64"`                                                             // 标签ID列表
	MediaReviewStatus string     `json:"mediaReviewStatus" form:"mediaReviewStatus" binding:"omitempty,oneof=pending passed rejected failed not_required"` // 图片审核状态
	RecipeID          string     `json:"recipeId" form:"recipeId" binding:"omitempty,max=64"`                                                              // 菜谱ID
	CreatedFrom       *time.Time `json:"createdFrom" form:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                           // 创建起始时间
	CreatedTo         *time.Time `json:"createdTo" form:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                                               // 创建结束时间
	UpdatedFrom       *time.Time `json:"updatedFrom" form:"updatedFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                           // 更新起始时间
	UpdatedTo         *time.Time `json:"updatedTo" form:"updatedTo" time_format:"2006-01-02T15:04:05Z07:00"`                                               // 更新结束时间
	DeletedFrom       *time.Time `json:"deletedFrom" form:"deletedFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                           // 删除起始时间
	DeletedTo         *time.Time `json:"deletedTo" form:"deletedTo" time_format:"2006-01-02T15:04:05Z07:00"`                                               // 删除结束时间
	SortBy            string     `json:"sortBy" form:"sortBy"`                                                                                             // 排序字段
	SortOrder         string     `json:"sortOrder" form:"sortOrder"`                                                                                       // 排序方向
}

// UserDishReferenceSearch 表示用户菜品引用查询条件。
type UserDishReferenceSearch struct {
	Page          int    `json:"page" form:"page" binding:"omitempty,min=1"`                                                                                     // 页码
	PageSize      int    `json:"pageSize" form:"pageSize" binding:"omitempty,oneof=20 50 100"`                                                                   // 每页数量
	ReferenceType string `json:"referenceType" form:"referenceType" binding:"omitempty,oneof=recipe unconfirmed_meal confirmed_meal_snapshot shopping_snapshot"` // 引用类型
	SortBy        string `json:"sortBy" form:"sortBy" binding:"omitempty,oneof=occurredAt"`                                                                      // 排序字段
	SortOrder     string `json:"sortOrder" form:"sortOrder" binding:"omitempty,oneof=asc desc"`                                                                  // 排序方向
}

// UserRecipeSearch 表示用户菜谱查询条件。
type UserRecipeSearch struct {
	Page           int        `json:"page" form:"page" binding:"omitempty,min=1"`                                                  // 页码
	PageSize       int        `json:"pageSize" form:"pageSize" binding:"omitempty,oneof=20 50 100"`                                // 每页数量
	Keyword        string     `json:"keyword" form:"keyword" binding:"omitempty,max=60"`                                           // 关键词
	UserID         string     `json:"userId" form:"userId" binding:"omitempty,max=64"`                                             // 用户ID
	UserKeyword    string     `json:"userKeyword" form:"userKeyword" binding:"omitempty,max=60" checksql:"false"`                  // 用户ID或昵称关键词
	DishID         string     `json:"dishId" form:"dishId" binding:"omitempty,max=64"`                                             // 菜品ID
	MinDishCount   *int       `json:"minDishCount" form:"minDishCount" binding:"omitempty,min=0"`                                  // 最小菜品数量
	MaxDishCount   *int       `json:"maxDishCount" form:"maxDishCount" binding:"omitempty,min=0"`                                  // 最大菜品数量
	DeletionStatus string     `json:"deletionStatus" form:"deletionStatus" binding:"omitempty,oneof=active deleted"`               // 删除状态
	ContentState   string     `json:"contentState" form:"contentState" binding:"omitempty,oneof=ready empty contains_unavailable"` // 内容状态
	HasNote        *bool      `json:"hasNote" form:"hasNote"`                                                                      // 是否有备注
	CreatedFrom    *time.Time `json:"createdFrom" form:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                      // 创建起始时间
	CreatedTo      *time.Time `json:"createdTo" form:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                          // 创建结束时间
	UpdatedFrom    *time.Time `json:"updatedFrom" form:"updatedFrom" time_format:"2006-01-02T15:04:05Z07:00"`                      // 更新起始时间
	UpdatedTo      *time.Time `json:"updatedTo" form:"updatedTo" time_format:"2006-01-02T15:04:05Z07:00"`                          // 更新结束时间
	SortBy         string     `json:"sortBy" form:"sortBy" binding:"omitempty,oneof=createdAt updatedAt deletedAt name dishCount"` // 排序字段
	SortOrder      string     `json:"sortOrder" form:"sortOrder" binding:"omitempty,oneof=asc desc"`                               // 排序方向
}

// GovernanceActionPreviewInput 表示违规处理操作预览输入参数。
type GovernanceActionPreviewInput struct {
	TargetType            string   `json:"targetType" binding:"required,oneof=dish official_dish recipe checkin"`                                                                                                          // 实际处理的源头类型
	TargetID              string   `json:"targetId" binding:"required,max=64"`                                                                                                                                             // 实际处理的源头ID
	EntryRecommendationID string   `json:"entryRecommendationId" binding:"omitempty,max=64"`                                                                                                                               // 发起处理的推荐ID
	Actions               []string `json:"actions" binding:"required,min=1,unique,dive,oneof=disable_discoverability soft_delete_dish soft_delete_official_dish soft_delete_recipe soft_delete_checkin delete_copy_chain"` // 操作列表
	ViolationType         string   `json:"violationType" binding:"required,max=40"`                                                                                                                                        // 违规类型
	Severity              string   `json:"severity" binding:"required,oneof=normal serious"`                                                                                                                               // 严重程度
	Reason                string   `json:"reason" binding:"required,min=4,max=300"`                                                                                                                                        // 原因
	ExpectedVersion       int      `json:"expectedVersion" binding:"required,min=1"`                                                                                                                                       // 预期版本
}

// GovernanceActionExecuteInput 表示违规处理操作执行输入参数。
type GovernanceActionExecuteInput struct {
	GovernanceActionPreviewInput        // 违规处理预览参数
	PreviewToken                 string `json:"previewToken" binding:"required"`        // 预览Token
	ConfirmText                  string `json:"confirmText" binding:"omitempty,max=20"` // L3复制链处理确认词
}

// UserDishIDPath 表示用户菜品ID路径参数。
type UserDishIDPath struct {
	DishID string `uri:"dishId" json:"dishId" binding:"required,max=64"` // 菜品ID
}

// UserRecipeIDPath 表示用户菜谱ID路径参数。
type UserRecipeIDPath struct {
	RecipeID string `uri:"recipeId" json:"recipeId" binding:"required,max=64"` // 菜谱ID
}

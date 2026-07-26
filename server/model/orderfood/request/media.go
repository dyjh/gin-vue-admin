package request

import "time"

// MediaListQuery 表示图片资源分页查询条件。
type MediaListQuery struct {
	Page                    int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                                                                            // 页码
	PageSize                int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                                                                          // 每页数量
	FileID                  string     `form:"fileId" json:"fileId" binding:"omitempty,max=128"`                                                                                                      // 文件ID
	UploaderUserID          string     `form:"uploaderUserId" json:"uploaderUserId" binding:"omitempty,max=64"`                                                                                       // 上传用户ID
	UploaderAdministratorID string     `form:"uploaderAdministratorId" json:"uploaderAdministratorId" binding:"omitempty,max=32"`                                                                     // 上传管理员ID
	UploadSource            string     `form:"uploadSource" json:"uploadSource" binding:"omitempty,oneof=user_upload admin_upload ai_generated"`                                                      // 上传来源
	SourceScene             string     `form:"sourceScene" json:"sourceScene" binding:"omitempty,oneof=profile_avatar dish_cover dish_step dish_extract checkin generated_cover official_dish_cover"` // 来源场景
	BoundObjectType         string     `form:"boundObjectType" json:"boundObjectType" binding:"omitempty,max=60"`                                                                                     // 绑定对象类型
	BoundObjectID           string     `form:"boundObjectId" json:"boundObjectId" binding:"omitempty,max=80"`                                                                                         // 绑定对象ID
	ReviewStatus            string     `form:"reviewStatus" json:"reviewStatus" binding:"omitempty,oneof=pending passed rejected failed not_required"`                                                // 审核状态
	ResourceStatus          string     `form:"resourceStatus" json:"resourceStatus" binding:"omitempty,oneof=active missing deleted"`                                                                 // 资源状态
	CreatedFrom             *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                                                                // 创建起始时间
	CreatedTo               *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                                                                                    // 创建结束时间
	SortBy                  string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt fileSize"`                                                                                     // 排序字段
	SortOrder               string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                                                                         // 排序方向
}

// ApplyDefaults 补齐图片资源列表的分页和排序默认值。
func (query *MediaListQuery) ApplyDefaults() {
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

// MediaFileIDPath 表示图片文件ID路径参数。
type MediaFileIDPath struct {
	FileID string `uri:"fileId" json:"fileId" binding:"required,max=128"` // 文件ID
}

// ModerationRecordListQuery 表示图片审核记录分页查询条件。
type ModerationRecordListQuery struct {
	Page        int        `form:"page" json:"page" binding:"omitempty,min=1"`                                                                    // 页码
	PageSize    int        `form:"pageSize" json:"pageSize" binding:"omitempty,oneof=20 50 100"`                                                  // 每页数量
	RequestID   string     `form:"requestId" json:"requestId" binding:"omitempty,max=128"`                                                        // 请求ID
	FileID      string     `form:"fileId" json:"fileId" binding:"omitempty,max=128"`                                                              // 文件ID
	UserID      string     `form:"userId" json:"userId" binding:"omitempty,max=64"`                                                               // 用户ID
	ObjectType  string     `form:"objectType" json:"objectType" binding:"omitempty,max=60"`                                                       // 对象类型
	Status      string     `form:"status" json:"status" binding:"omitempty,oneof=pending passed rejected failed not_required rejected_or_failed"` // 审核状态或未通过汇总
	RiskLabel   string     `form:"riskLabel" json:"riskLabel" binding:"omitempty,max=80"`                                                         // 风险标签
	CreatedFrom *time.Time `form:"createdFrom" json:"createdFrom" time_format:"2006-01-02T15:04:05Z07:00"`                                        // 创建起始时间
	CreatedTo   *time.Time `form:"createdTo" json:"createdTo" time_format:"2006-01-02T15:04:05Z07:00"`                                            // 创建结束时间
	SortBy      string     `form:"sortBy" json:"sortBy" binding:"omitempty,oneof=createdAt durationMs"`                                           // 排序字段
	SortOrder   string     `form:"sortOrder" json:"sortOrder" binding:"omitempty,oneof=asc desc"`                                                 // 排序方向
}

// ApplyDefaults 补齐图片审核记录列表的分页和排序默认值。
func (query *ModerationRecordListQuery) ApplyDefaults() {
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

// ModerationRecordIDPath 表示图片审核记录ID路径参数。
type ModerationRecordIDPath struct {
	RecordID string `uri:"recordId" json:"recordId" binding:"required,max=64"` // 审核记录ID
}

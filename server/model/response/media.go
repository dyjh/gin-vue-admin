package response

import "time"

// MediaSummary 表示图片资源摘要。
type MediaSummary struct {
	ID                 string      `json:"id"`                 // 文件ID
	ObjectKeyMasked    string      `json:"objectKeyMasked"`    // 脱敏后的存储对象键
	URL                *string     `json:"url"`                // 可访问地址
	SourceScene        string      `json:"sourceScene"`        // 来源场景
	UploadSource       string      `json:"uploadSource"`       // 上传来源
	Uploader           interface{} `json:"uploader"`           // 上传用户或管理员摘要
	BoundObjectType    *string     `json:"boundObjectType"`    // 绑定对象类型
	BoundObjectID      *string     `json:"boundObjectId"`      // 绑定对象ID
	BoundObjectVersion *int        `json:"boundObjectVersion"` // 绑定对象数据版本
	BoundObjectDeleted bool        `json:"boundObjectDeleted"` // 绑定对象是否已软删除
	ReviewStatus       string      `json:"reviewStatus"`       // 审核状态
	ResourceStatus     string      `json:"resourceStatus"`     // 资源状态
	Width              *int        `json:"width"`              // 图片宽度
	Height             *int        `json:"height"`             // 图片高度
	FileSize           int64       `json:"fileSize"`           // 文件字节数
	MimeType           string      `json:"mimeType"`           // MIME类型
	CreatedAt          time.Time   `json:"createdAt"`          // 创建时间
}

// MediaDetail 表示图片资源详情。
type MediaDetail struct {
	MediaSummary                                    // 图片资源摘要
	Checksum               *string                  `json:"checksum"`               // 文件摘要
	LatestModerationRecord *ModerationRecordSummary `json:"latestModerationRecord"` // 最近审核记录
	DeletedAt              *time.Time               `json:"deletedAt"`              // 资源删除时间
	DeleteReason           *string                  `json:"deleteReason"`           // 资源删除原因
}

// ModerationRecordDetail 表示图片审核记录详情。
type ModerationRecordDetail struct {
	ModerationRecordSummary                        // 图片审核记录摘要
	EndpointLabel           *string                `json:"endpointLabel"`           // 供应商端点说明
	ErrorSummary            *string                `json:"errorSummary"`            // 错误摘要
	ProviderResponseSummary map[string]interface{} `json:"providerResponseSummary"` // 供应商响应脱敏摘要
}

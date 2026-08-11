package content

import (
	"time"
)

// FrontMediaAsset 表示小程序媒体文件。
type FrontMediaAsset struct {
	ID                    string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                      // 主键ID
	UserID                string     `json:"userId" gorm:"column:user_id;type:varchar(64);not null;default:'';index;comment:用户ID;"`                       // 用户ID
	AdministratorID       *uint      `json:"administratorId" gorm:"column:administrator_id;type:bigint unsigned;index;default:null;comment:上传管理员ID;"`     // 上传管理员ID
	AdministratorUsername *string    `json:"administratorUsername" gorm:"column:administrator_username;type:varchar(120);default:null;comment:上传管理员用户名;"` // 上传管理员用户名
	AdministratorNickname *string    `json:"administratorNickname" gorm:"column:administrator_nickname;type:varchar(120);default:null;comment:上传管理员昵称;"`  // 上传管理员昵称
	FileName              string     `json:"fileName" gorm:"column:file_name;type:varchar(255);not null;default:'';comment:原始文件名;"`                       // 原始文件名
	UploadSource          string     `json:"uploadSource" gorm:"column:upload_source;type:varchar(24);not null;default:user_upload;index;comment:上传来源;"`  // 上传来源
	Scene                 string     `json:"scene" gorm:"column:scene;type:varchar(32);not null;index;comment:审核场景;"`                                     // 审核场景
	ResourceStatus        string     `json:"resourceStatus" gorm:"column:resource_status;type:varchar(24);not null;default:active;index;comment:资源状态;"`   // 资源状态
	URL                   string     `json:"url" gorm:"column:url;type:varchar(512);not null;comment:访问地址;"`                                              // 访问地址
	StoragePath           string     `json:"-" gorm:"column:storage_path;type:varchar(1024);not null;comment:对象存储路径;"`                                    // 对象存储路径
	ContentType           string     `json:"-" gorm:"column:content_type;type:varchar(80);not null;comment:媒体类型;"`                                        // 媒体类型
	Width                 int        `json:"width" gorm:"column:width;type:int;not null;comment:图片宽度;"`                                                   // 图片宽度
	Height                int        `json:"height" gorm:"column:height;type:int;not null;comment:图片高度;"`                                                 // 图片高度
	SizeBytes             int64      `json:"-" gorm:"column:size_bytes;type:bigint;not null;comment:文件字节数;"`                                              // 文件字节数
	Checksum              string     `json:"-" gorm:"column:checksum;type:varchar(64);not null;default:'';comment:文件SHA256摘要;"`                           // 文件SHA256摘要
	ReviewStatus          string     `json:"reviewStatus" gorm:"column:review_status;type:varchar(24);not null;index;comment:图片审核状态;"`                    // 图片审核状态
	RejectReason          *string    `json:"rejectReason" gorm:"column:reject_reason;type:varchar(240);default:null;comment:拒绝原因;"`                       // 拒绝原因
	DeletedAt             *time.Time `json:"deletedAt" gorm:"column:deleted_at;type:datetime;index;default:null;comment:资源删除时间;"`                         // 资源删除时间
	DeleteReason          *string    `json:"deleteReason" gorm:"column:delete_reason;type:varchar(200);default:null;comment:资源删除原因;"`                     // 资源删除原因
	CreatedAt             time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;index;comment:创建时间;"`                               // 创建时间
}

// TableName 指定FrontMediaAsset对应的数据表名。
func (FrontMediaAsset) TableName() string { return "of_media" }

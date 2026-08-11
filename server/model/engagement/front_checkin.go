package engagement

import (
	"gorm.io/gorm"
	"time"
)

// FrontCheckin 表示用户做菜打卡。
type FrontCheckin struct {
	ID                   string         `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                                      // 主键ID
	UserID               string         `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`                                  // 用户ID
	DishName             string         `json:"dishName" gorm:"column:dish_name;type:varchar(40);not null;comment:菜品名称;"`                                    // 菜品名称
	ImageFileID          string         `json:"imageFileId" gorm:"column:image_file_id;type:varchar(64);not null;index;comment:图片文件ID;"`                     // 图片文件ID
	ImageURL             string         `json:"imageUrl" gorm:"column:image_url;type:varchar(512);not null;comment:图片地址;"`                                   // 图片地址
	Note                 *string        `json:"note" gorm:"column:note;type:varchar(120);default:null;comment:备注;"`                                          // 备注
	CheckedDate          string         `json:"-" gorm:"column:checked_date;type:varchar(10);not null;index;comment:打卡日期;"`                                  // 打卡日期
	Rewarded             bool           `json:"rewarded" gorm:"column:rewarded;type:tinyint(1) unsigned;not null;comment:是否已发放奖励;"`                          // 是否已发放奖励
	CheckedAt            time.Time      `json:"checkedAt" gorm:"column:checked_at;type:datetime;not null;index;comment:打卡时间;"`                               // 打卡时间
	Version              int            `json:"version" gorm:"column:version;type:int;not null;default:1;comment:数据版本;"`                                     // 数据版本
	DeletedReason        *string        `json:"deletedReason" gorm:"column:deleted_reason;type:varchar(600);default:null;comment:删除原因;"`                     // 删除原因
	DeletedViolationType *string        `json:"deletedViolationType" gorm:"column:deleted_violation_type;type:varchar(80);default:null;comment:删除违规类型;"`     // 删除违规类型
	DeletedByAdminID     *uint          `json:"deletedByAdminId" gorm:"column:deleted_by_admin_id;type:bigint unsigned;index;default:null;comment:删除管理员ID;"` // 删除管理员ID
	DeletedByUsername    *string        `json:"deletedByUsername" gorm:"column:deleted_by_username;type:varchar(120);default:null;comment:删除管理员用户名;"`        // 删除管理员用户名
	DeletedByNickname    *string        `json:"deletedByNickname" gorm:"column:deleted_by_nickname;type:varchar(120);default:null;comment:删除管理员昵称;"`         // 删除管理员昵称
	CreatedAt            time.Time      `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                             // 创建时间
	DeletedAt            gorm.DeletedAt `json:"deletedAt" gorm:"column:deleted_at;type:datetime;index;default:null;comment:软删除时间;"`                          // 软删除时间
}

// TableName 指定FrontCheckin对应的数据表名。
func (FrontCheckin) TableName() string { return "of_checkins" }

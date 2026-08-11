package content

import (
	"github.com/dyjh/order-food-mini-app/server/global"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
)

// UserRecipe 表示用户菜谱。
type UserRecipe struct {
	global.GVA_MODEL                       // GVA基础模型字段
	PublicID         string                `json:"publicId" gorm:"column:public_id;type:varchar(64);uniqueIndex;not null;comment:对外公开ID;"` // 对外公开ID
	OwnerID          string                `json:"ownerId" gorm:"column:owner_id;type:varchar(64);not null;index;comment:所属用户ID;"`         // 所属用户ID
	Owner            userModel.MiniAppUser `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`                                          // 所属用户关联数据
	Name             string                `json:"name" gorm:"column:name;type:varchar(120);not null;index;comment:名称;"`                   // 名称
	Note             *string               `json:"note" gorm:"column:note;type:varchar(1000);default:null;comment:备注;"`                    // 备注
	Version          int                   `json:"version" gorm:"column:version;type:int;not null;default:1;comment:数据版本;"`                // 数据版本

	DeletedReason        *string `json:"deletedReason" gorm:"column:deleted_reason;type:varchar(600);default:null;comment:删除原因;"`                     // 删除原因
	DeletedViolationType *string `json:"deletedViolationType" gorm:"column:deleted_violation_type;type:varchar(80);default:null;comment:删除违规类型;"`     // 删除违规类型
	DeletedByAdminID     *uint   `json:"deletedByAdminId" gorm:"column:deleted_by_admin_id;type:bigint unsigned;index;default:null;comment:删除管理员ID;"` // 删除管理员ID
	DeletedByUsername    *string `json:"deletedByUsername" gorm:"column:deleted_by_username;type:varchar(120);default:null;comment:删除管理员用户名;"`        // 删除管理员用户名
	DeletedByNickname    *string `json:"deletedByNickname" gorm:"column:deleted_by_nickname;type:varchar(120);default:null;comment:删除管理员昵称;"`         // 删除管理员昵称

	RecipeDishes []RecipeDish `json:"recipeDishes" gorm:"foreignKey:RecipeID"` // 菜谱菜品关联列表
}

// TableName 指定UserRecipe对应的数据表名。
func (UserRecipe) TableName() string { return "of_recipes" }

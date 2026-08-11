package meal

import (
	"time"
)

// FrontShoppingList 表示采购清单。
type FrontShoppingList struct {
	ID             string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`                 // 主键ID
	MealID         string     `json:"mealId" gorm:"column:meal_id;type:varchar(64);not null;uniqueIndex;comment:饭局ID;"`       // 饭局ID
	OwnerID        string     `json:"ownerId" gorm:"column:owner_id;type:varchar(64);not null;index;comment:所属用户ID;"`         // 所属用户ID
	ShareTokenHash string     `json:"-" gorm:"column:share_token_hash;type:varchar(64);not null;uniqueIndex;comment:分享令牌摘要;"` // 分享令牌摘要
	ShareToken     string     `json:"shareToken" gorm:"column:share_token;type:varchar(128);not null;comment:分享令牌;"`          // 分享令牌
	ShareExpiresAt *time.Time `json:"-" gorm:"column:share_expires_at;type:datetime;index;default:null;comment:分享过期时间;"`      // 分享过期时间
	ShareRevokedAt *time.Time `json:"-" gorm:"column:share_revoked_at;type:datetime;index;default:null;comment:分享撤销时间;"`      // 分享撤销时间
	CreatedAt      time.Time  `json:"createdAt" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                // 创建时间
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                // 更新时间
}

// TableName 指定FrontShoppingList对应的数据表名。
func (FrontShoppingList) TableName() string { return "of_shopping_lists" }

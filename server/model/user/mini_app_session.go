package user

import (
	"time"
)

// MiniAppSession is an opaque, front-only login session. Only the SHA-256
// digest is persisted; the bearer token itself is returned once and never
// logged or stored.
type MiniAppSession struct {
	ID         string     `json:"-" gorm:"column:id;type:varchar(64);primaryKey;not null;comment:主键ID;"`          // 主键ID
	TokenHash  string     `json:"-" gorm:"column:token_hash;type:varchar(64);not null;uniqueIndex;comment:令牌摘要;"` // 令牌摘要
	UserID     string     `json:"-" gorm:"column:user_id;type:varchar(64);not null;index;comment:用户ID;"`          // 用户ID
	ExpiresAt  time.Time  `json:"-" gorm:"column:expires_at;type:datetime;not null;index;comment:过期时间;"`          // 过期时间
	RevokedAt  *time.Time `json:"-" gorm:"column:revoked_at;type:datetime;index;default:null;comment:撤销时间;"`      // 撤销时间
	LastUsedAt *time.Time `json:"-" gorm:"column:last_used_at;type:datetime;index;default:null;comment:最近使用时间;"`  // 最近使用时间
	CreatedAt  time.Time  `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                // 创建时间
}

// TableName 指定MiniAppSession对应的数据表名。
func (MiniAppSession) TableName() string {
	return "of_user_sessions"
}

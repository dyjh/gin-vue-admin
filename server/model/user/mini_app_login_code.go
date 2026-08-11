package user

import (
	"time"
)

// MiniAppLoginCode stores only a digest of a successfully redeemed one-time
// wx.login code. It provides deterministic replay rejection without retaining
// the credential.
type MiniAppLoginCode struct {
	CodeHash   string    `json:"-" gorm:"column:code_hash;type:varchar(64);primaryKey;not null;comment:登录凭证摘要;"` // 登录凭证摘要
	RedeemedAt time.Time `json:"-" gorm:"column:redeemed_at;type:datetime;not null;comment:核销时间;"`               // 核销时间
	ExpiresAt  time.Time `json:"-" gorm:"column:expires_at;type:datetime;not null;index;comment:过期时间;"`          // 过期时间
}

// TableName 指定MiniAppLoginCode对应的数据表名。
func (MiniAppLoginCode) TableName() string {
	return "of_login_codes"
}

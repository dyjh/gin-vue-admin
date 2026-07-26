package orderfood

import "time"

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

// FrontIdempotencyRecord 表示小程序幂等请求记录。
type FrontIdempotencyRecord struct {
	ID             uint      `json:"-" gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;not null;comment:主键ID;"`                    // 主键ID
	UserID         string    `json:"-" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uk_of_idempotency,priority:1;comment:用户ID;"`   // 用户ID
	Endpoint       string    `json:"-" gorm:"column:endpoint;type:varchar(180);not null;uniqueIndex:uk_of_idempotency,priority:2;comment:服务端点;"` // 服务端点
	Key            string    `json:"-" gorm:"column:key;type:varchar(128);not null;uniqueIndex:uk_of_idempotency,priority:3;comment:幂等键;"`       // 幂等键
	RequestHash    string    `json:"-" gorm:"column:request_hash;type:varchar(64);not null;comment:请求内容摘要;"`                                     // 请求内容摘要
	State          string    `json:"-" gorm:"column:state;type:varchar(20);not null;index;comment:处理状态;"`                                        // 处理状态
	ResponseStatus int       `json:"-" gorm:"column:response_status;type:int;not null;default:0;comment:响应状态码;"`                                 // 响应状态码
	ResponseBody   []byte    `json:"-" gorm:"column:response_body;type:blob;default:null;comment:响应内容;"`                                         // 响应内容
	CreatedAt      time.Time `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                            // 创建时间
	UpdatedAt      time.Time `json:"-" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                            // 更新时间
	ExpiresAt      time.Time `json:"-" gorm:"column:expires_at;type:datetime;not null;index;comment:过期时间;"`                                      // 过期时间
}

// TableName 指定FrontIdempotencyRecord对应的数据表名。
func (FrontIdempotencyRecord) TableName() string {
	return "of_front_idem"
}

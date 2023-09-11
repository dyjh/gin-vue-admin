package business

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type Members struct {
	global.GVA_MODEL
	Nickname string `json:"nickname" gorm:"comment:昵称"` // api路径
	Openid   string `json:"openid" gorm:"comment:openid"` // api中文描述
}

func (Members) TableName() string {
	return "members"
}

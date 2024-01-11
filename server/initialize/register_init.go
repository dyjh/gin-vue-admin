package initialize

import (
	_ "github.com/flipped-aurora/gin-vue-admin/server/source/business"
	_ "github.com/flipped-aurora/gin-vue-admin/server/source/example"
	_ "github.com/flipped-aurora/gin-vue-admin/server/source/system"
)

// 后台初始化数据库增加引用后才能 auto init
func init() {
	// do nothing,only import source package so that inits can be registered
}

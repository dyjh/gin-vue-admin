package request

import "github.com/dyjh/order-food-mini-app/server/model/system"

// AddMenuAuthorityInfo Add menu authority info structure
type AddMenuAuthorityInfo struct {
	Menus       []system.SysBaseMenu `json:"menus"`
	AuthorityId uint                 `json:"authorityId"` // 角色ID
}

// SetMenuAuthorities 通过菜单ID全量覆盖关联角色列表
type SetMenuAuthorities struct {
	MenuId       uint   `json:"menuId" form:"menuId"`             // 菜单ID
	AuthorityIds []uint `json:"authorityIds" form:"authorityIds"` // 角色ID列表
}

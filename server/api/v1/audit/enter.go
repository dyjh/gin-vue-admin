package audit

import "github.com/dyjh/order-food-mini-app/server/service"

// ApiGroup 聚合管理操作审计接口。
type ApiGroup struct {
	AuditApi // 管理操作审计接口
}

var auditService = service.ServiceGroupApp.AuditServiceGroup.Audit

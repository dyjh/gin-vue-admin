package audit

import api "github.com/dyjh/order-food-mini-app/server/api/v1"

type RouterGroup struct {
	AuditRouter
}

var auditApi = api.ApiGroupApp.AuditApiGroup.AuditApi

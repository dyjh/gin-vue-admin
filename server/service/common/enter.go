package common

// ServiceGroup aggregates cross-domain backend services.
type ServiceGroup struct {
	Permission  *PermissionService  // 管理权限服务
	AccessAudit *AccessAuditService // 敏感访问审计服务
	Idempotency *IdempotencyService // 管理操作幂等服务
}

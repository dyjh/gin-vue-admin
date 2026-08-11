package audit

// ServiceGroup aggregates administrator audit services.
type ServiceGroup struct {
	Audit *AuditService // 管理操作审计服务
}

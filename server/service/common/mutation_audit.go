package common

import (
	"context"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	"gorm.io/gorm"
)

// MutationAuditWriter defines the shared mutation-audit capability used by
// backend feature services.
type MutationAuditWriter interface {
	WriteMutationAudit(context.Context, *gorm.DB, *auditModel.AdminAuditLog) error
}

// GormMutationAuditWriter persists mutation audit records with GORM.
type GormMutationAuditWriter struct{}

// WriteMutationAudit writes one mutation audit record.
func (GormMutationAuditWriter) WriteMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	audit *auditModel.AdminAuditLog,
) error {
	if tx == nil || audit == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	if audit.PublicID == "" {
		audit.PublicID = NewPublicID("audit")
	}
	if err := tx.WithContext(ctx).Create(audit).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "create mutation audit log")
	}
	return nil
}

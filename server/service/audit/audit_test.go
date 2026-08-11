package audit

import (
	"context"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"testing"
	"time"

	auditModel "github.com/dyjh/order-food-mini-app/server/model/audit"
	auditRequest "github.com/dyjh/order-food-mini-app/server/model/audit/request"
	"gorm.io/datatypes"
)

func TestAuditServiceListAndDetailUseOperationAudit(t *testing.T) {
	db := newOrderFoodTestDB(t)
	if err := db.AutoMigrate(&auditModel.AdminAuditLog{}); err != nil {
		t.Fatalf("migrate audit table: %v", err)
	}
	targetLabel := "测试用户"
	reason := "违反平台规则"
	idempotencyKey := "disable-user-001"
	ip := "10.20.0.0"
	userAgent := "Mozilla/5.0 test"
	record := auditModel.AdminAuditLog{
		PublicID:              "audit-001",
		AdministratorID:       11,
		AdministratorUsername: "governance",
		AdministratorNickname: stringPointer("治理员"),
		Action:                "update_user_status",
		TargetType:            "user",
		TargetID:              "user-001",
		TargetLabel:           &targetLabel,
		Reason:                &reason,
		BeforeSummary:         datatypes.JSON(`{"status":"normal","version":1}`),
		AfterSummary:          datatypes.JSON(`{"status":"disabled","version":2}`),
		RequestID:             "request-audit-001",
		IdempotencyKey:        &idempotencyKey,
		SourceIPMasked:        &ip,
		UserAgentSummary:      &userAgent,
	}
	record.CreatedAt = time.Date(2026, time.July, 24, 9, 0, 0, 0, time.UTC)
	record.UpdatedAt = record.CreatedAt
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create audit: %v", err)
	}

	service := NewAuditService(db)
	page, err := service.List(context.Background(), auditRequest.AuditLogListQuery{
		AdministratorID: "11",
		TargetType:      "user",
		RequestID:       "request-audit-001",
	})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 {
		t.Fatalf("unexpected audit page: %+v", page)
	}
	item := page.List[0]
	if item.ID != record.PublicID ||
		item.Administrator.Username != record.AdministratorUsername ||
		item.Administrator.Nickname == nil ||
		*item.Administrator.Nickname != "治理员" ||
		item.TargetLabel == nil ||
		*item.TargetLabel != targetLabel ||
		item.Reason == nil ||
		*item.Reason != reason ||
		item.IdempotencyKey == nil ||
		*item.IdempotencyKey != idempotencyKey {
		t.Fatalf("unexpected audit summary: %+v", item)
	}

	detail, err := service.Detail(context.Background(), record.PublicID)
	if err != nil {
		t.Fatalf("audit detail: %v", err)
	}
	if detail.SourceIPMasked == nil || *detail.SourceIPMasked != ip ||
		detail.UserAgentSummary == nil || *detail.UserAgentSummary != userAgent {
		t.Fatalf("unexpected audit detail: %+v", detail)
	}
	before, ok := detail.BeforeSummary.(map[string]interface{})
	if !ok || before["status"] != "normal" || before["version"] != float64(1) {
		t.Fatalf("unexpected before summary: %#v", detail.BeforeSummary)
	}
	after, ok := detail.AfterSummary.(map[string]interface{})
	if !ok || after["status"] != "disabled" || after["version"] != float64(2) {
		t.Fatalf("unexpected after summary: %#v", detail.AfterSummary)
	}
}

func stringPointer(value string) *string {
	return &value
}

func TestMaskIP(t *testing.T) {
	tests := map[string]string{
		"192.168.10.20":      "192.168.0.0",
		"192.168.10.20:8080": "192.168.0.0",
		"2001:db8:abcd::1":   "2001:db8:abcd::",
		"not-an-ip":          "",
	}
	for input, want := range tests {
		if got := serviceCommon.MaskIP(input); got != want {
			t.Fatalf("serviceCommon.MaskIP(%q) = %q, want %q", input, got, want)
		}
	}
}

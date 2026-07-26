package orderfood

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	"gorm.io/gorm"
)

// SensitiveAccess 表示一次敏感数据读取的管理员、权限和请求上下文。
type SensitiveAccess struct {
	AdministratorID       uint    // 管理员ID
	AdministratorUsername string  // 管理员用户名
	AdministratorNickname *string // 管理员昵称
	AuthorityID           uint    // 当前角色ID
	Permission            string  // 本次读取所需权限码
	Action                string  // 审计动作
	TargetType            string  // 目标类型
	TargetID              string  // 目标ID
	TargetLabel           *string // 目标展示名称
	RequestID             string  // 请求ID
	SourceIPMasked        string  // 脱敏来源IP
	UserAgentSummary      string  // 客户端信息摘要
}

// AccessAuditService 提供敏感数据访问审计业务能力。
type AccessAuditService struct {
	DB *gorm.DB // 业务数据库
}

// NewAccessAuditService 创建访问审计服务实例。
func NewAccessAuditService(db *gorm.DB) *AccessAuditService {
	return &AccessAuditService{DB: db}
}

func (service *AccessAuditService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

// Record 在独立事务中记录一次敏感数据访问。
func (service *AccessAuditService) Record(
	ctx context.Context,
	access SensitiveAccess,
) (string, error) {
	return service.RecordWithDB(ctx, service.database(), access)
}

// RecordWithDB 记录配置数据库。
func (service *AccessAuditService) RecordWithDB(
	ctx context.Context,
	db *gorm.DB,
	access SensitiveAccess,
) (string, error) {
	if db == nil {
		return "", appErrors.AdminInternal.DefaultMsg()
	}
	if access.AdministratorID == 0 ||
		access.Permission == "" ||
		access.Action == "" ||
		access.TargetType == "" ||
		access.TargetID == "" ||
		access.RequestID == "" {
		return "", appErrors.AdminBadRequest.DefaultMsg()
	}

	auditID := orderfoodModel.NewID()
	now := time.Now().UTC()
	username := strings.TrimSpace(access.AdministratorUsername)
	if username == "" {
		username = "administrator-" + strconv.FormatUint(uint64(access.AdministratorID), 10)
	}
	var sourceIPMasked *string
	if value := strings.TrimSpace(access.SourceIPMasked); value != "" {
		sourceIPMasked = &value
	}
	var userAgentSummary *string
	if value := truncateRunes(strings.TrimSpace(access.UserAgentSummary), 240); value != "" {
		userAgentSummary = &value
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		accessRecord := orderfoodModel.AdminAccessAudit{
			ID:               auditID,
			AdministratorID:  access.AdministratorID,
			AuthorityID:      access.AuthorityID,
			Permission:       access.Permission,
			Action:           access.Action,
			TargetType:       access.TargetType,
			TargetID:         access.TargetID,
			RequestID:        access.RequestID,
			SourceIPMasked:   sourceIPMasked,
			UserAgentSummary: userAgentSummary,
			CreatedAt:        now,
		}
		if err := tx.Create(&accessRecord).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "write sensitive access audit")
		}
		operationRecord := orderfoodModel.AdminAuditLog{
			PublicID:              auditID,
			AdministratorID:       access.AdministratorID,
			AdministratorUsername: username,
			AdministratorNickname: access.AdministratorNickname,
			Action:                access.Action,
			TargetType:            access.TargetType,
			TargetID:              access.TargetID,
			TargetLabel:           access.TargetLabel,
			RequestID:             access.RequestID,
			SourceIPMasked:        sourceIPMasked,
			UserAgentSummary:      userAgentSummary,
		}
		operationRecord.CreatedAt = now
		operationRecord.UpdatedAt = now
		if err := tx.Create(&operationRecord).Error; err != nil {
			return appErrors.AdminInternal.Wrap(err, "write sensitive operation audit")
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return auditID, nil
}

// MaskIP 对来源IP进行脱敏。
func MaskIP(address string) string {
	host := address
	if parsedHost, _, err := net.SplitHostPort(address); err == nil {
		host = parsedHost
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return ""
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return net.IPv4(ipv4[0], ipv4[1], 0, 0).String()
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return ""
	}
	masked := make(net.IP, net.IPv6len)
	copy(masked[:6], ipv6[:6])
	return masked.String()
}

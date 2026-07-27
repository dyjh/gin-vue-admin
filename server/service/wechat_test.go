package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	"github.com/dyjh/order-food-mini-app/server/model/system"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// setIdentityEncryptionKey 为当前测试注入config.yaml等价的身份字段加密密钥。
func setIdentityEncryptionKey(t *testing.T, key []byte) {
	t.Helper()
	original := global.GVA_CONFIG.OrderFood.IdentityKey
	global.GVA_CONFIG.OrderFood.IdentityKey = base64.StdEncoding.EncodeToString(key)
	t.Cleanup(func() {
		global.GVA_CONFIG.OrderFood.IdentityKey = original
	})
}

// weChatConfigTestService 创建微信配置服务测试所需的MySQL表和权限。
func weChatConfigTestService(t *testing.T) (*WeChatConfigService, *gorm.DB) {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&orderfoodModel.WeChatConfig{},
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAuditLog{},
		&system.SysBaseMenu{},
		&system.SysBaseMenuBtn{},
		&system.SysAuthorityBtn{},
	); err != nil {
		t.Fatal(err)
	}
	menu := system.SysBaseMenu{
		MenuLevel: 0,
		ParentId:  0,
		Path:      "wechat-config",
		Name:      "OrderFoodWeChatConfig",
		Component: "view/orderFood/wechatConfig/index.vue",
	}
	if err := db.Create(&menu).Error; err != nil {
		t.Fatal(err)
	}
	for _, permission := range []string{
		PermissionWeChatConfigRead,
		PermissionWeChatConfigUpdate,
	} {
		button := system.SysBaseMenuBtn{
			Name:          permission,
			Desc:          permission,
			SysBaseMenuID: menu.ID,
		}
		if err := db.Create(&button).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&system.SysAuthorityBtn{
			AuthorityId:      orderfoodModel.AuthorityOrderFoodSuperAdmin,
			SysMenuID:        menu.ID,
			SysBaseMenuBtnID: button.ID,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	permission := NewPermissionService(db)
	service := NewWeChatConfigService(db, permission, NewIdempotencyService(db))
	service.Now = func() time.Time {
		return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	}
	service.Idempotency.Now = service.Now
	service.Random = bytes.NewReader(bytes.Repeat([]byte{0x42}, 256))
	return service, db
}

// TestWeChatConfigServiceEncryptsSecretAndSupportsSafeUpdates 验证Secret加密保存、运行时解密和留空保留。
func TestWeChatConfigServiceEncryptsSecretAndSupportsSafeUpdates(t *testing.T) {
	setIdentityEncryptionKey(t, bytes.Repeat([]byte{0x11}, 32))
	service, db := weChatConfigTestService(t)
	actor := orderfoodRequest.AdminActor{
		AdministratorID: 1,
		AuthorityID:     orderfoodModel.AuthorityOrderFoodSuperAdmin,
		Username:        "admin",
		RequestID:       "request-wechat-config",
	}
	appSecret := "1234567890abcdef1234567890abcdef"
	first, replayed, err := service.UpdateConfig(
		context.Background(),
		actor,
		"wechat-config-first",
		orderfoodRequest.WeChatConfigUpdateInput{
			AppID:           "wx1234567890abcdef",
			AppSecret:       &appSecret,
			Reason:          "首次配置",
			ExpectedVersion: 0,
		},
	)
	if err != nil || replayed {
		t.Fatalf("first update = %#v, replayed=%v, err=%v", first, replayed, err)
	}
	if !first.Configured || !first.AppSecretConfigured || first.Version != 1 {
		t.Fatalf("first response = %#v", first)
	}

	var persisted orderfoodModel.WeChatConfig
	if err := db.First(&persisted, "singleton_key = ?", wechatConfigSingletonKey).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.AppSecretEncrypted == "" ||
		persisted.AppSecretEncrypted == appSecret ||
		strings.Contains(persisted.AppSecretEncrypted, appSecret) {
		t.Fatal("wechat AppSecret was not encrypted before persistence")
	}
	credentials, err := LoadCurrentWeChatCredentials(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if credentials.AppID != first.AppID || credentials.AppSecret != appSecret {
		t.Fatalf("runtime credentials = %#v", credentials)
	}

	second, _, err := service.UpdateConfig(
		context.Background(),
		actor,
		"wechat-config-second",
		orderfoodRequest.WeChatConfigUpdateInput{
			AppID:           first.AppID,
			Reason:          "更新说明",
			ExpectedVersion: first.Version,
		},
	)
	if err != nil || second.Version != 2 {
		t.Fatalf("second update = %#v, err=%v", second, err)
	}
	var updated orderfoodModel.WeChatConfig
	if err := db.First(&updated, "singleton_key = ?", wechatConfigSingletonKey).Error; err != nil {
		t.Fatal(err)
	}
	if updated.AppSecretEncrypted != persisted.AppSecretEncrypted {
		t.Fatal("empty AppSecret unexpectedly replaced the stored secret")
	}

	_, _, err = service.UpdateConfig(
		context.Background(),
		actor,
		"wechat-config-invalid-appid-change",
		orderfoodRequest.WeChatConfigUpdateInput{
			AppID:           "wxabcdef1234567890",
			Reason:          "更换应用",
			ExpectedVersion: second.Version,
		},
	)
	if appErrors.GetType(err) != appErrors.AdminInvalidConfig {
		t.Fatalf("AppID-only change error = %v", err)
	}
}

// TestLoadIdentityEncryptionKeyUsesConfigOnly 验证身份字段密钥读取config.yaml且错误提示可区分。
func TestLoadIdentityEncryptionKeyUsesConfigOnly(t *testing.T) {
	original := global.GVA_CONFIG.OrderFood.IdentityKey
	t.Cleanup(func() {
		global.GVA_CONFIG.OrderFood.IdentityKey = original
	})

	global.GVA_CONFIG.OrderFood.IdentityKey = ""
	if _, err := LoadIdentityEncryptionKey(); appErrors.GetType(err) != appErrors.AdminIdentityKeyMissing {
		t.Fatalf("missing config error = %v", err)
	}

	global.GVA_CONFIG.OrderFood.IdentityKey = "invalid-base64"
	if _, err := LoadIdentityEncryptionKey(); appErrors.GetType(err) != appErrors.AdminIdentityKeyInvalid {
		t.Fatalf("invalid config error = %v", err)
	}

	expected := bytes.Repeat([]byte{0x33}, 32)
	global.GVA_CONFIG.OrderFood.IdentityKey = base64.StdEncoding.EncodeToString(expected)
	actual, err := LoadIdentityEncryptionKey()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("identity key = %x, want %x", actual, expected)
	}
}

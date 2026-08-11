package user

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

type fakeWeChatExchanger struct {
	session WeChatSession
	err     error
}

func (fake fakeWeChatExchanger) Exchange(context.Context, string) (WeChatSession, error) {
	return fake.session, fake.err
}

// frontAuthTestDB 创建小程序登录服务测试所需的独立 MySQL 数据库。
func frontAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&userModel.MiniAppUser{},
		&userModel.MiniAppSession{},
		&userModel.MiniAppLoginCode{},
	); err != nil {
		t.Fatalf("migrate auth models: %v", err)
	}
	return db
}

func TestAuthServiceStoresOnlyTokenAndCredentialDigests(t *testing.T) {
	db := frontAuthTestDB(t)
	now := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	randomBytes := bytes.Repeat([]byte{0x42}, 512)
	service := AuthService{
		DB: db,
		Exchanger: fakeWeChatExchanger{session: WeChatSession{
			OpenID: "openid-sensitive", UnionID: "unionid-sensitive", SessionKey: "session-sensitive",
		}},
		IdentityKey: bytes.Repeat([]byte{0x11}, 32),
		Now:         func() time.Time { return now },
		Random:      bytes.NewReader(randomBytes),
	}
	result, err := service.Login(context.Background(), "wx-code-sensitive")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	var session userModel.MiniAppSession
	if err := db.First(&session).Error; err != nil {
		t.Fatalf("load session: %v", err)
	}
	if session.TokenHash == result.AccessToken || session.TokenHash != digest(result.AccessToken) {
		t.Fatal("session did not persist only the access-token digest")
	}
	var user userModel.MiniAppUser
	if err := db.First(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	for _, persisted := range []string{
		user.OpenIDHash,
		user.OpenIDEncrypted,
		stringValue(user.UnionIDEncrypted),
		stringValue(user.SessionKeyEncrypted),
	} {
		if strings.Contains(persisted, "sensitive") {
			t.Fatal("wechat credential was persisted in plaintext")
		}
	}
	authenticated, err := service.Authenticate(context.Background(), result.AccessToken)
	if err != nil || authenticated.ID != result.User.ID {
		t.Fatalf("authenticate returned wrong user: %v", err)
	}
}

func TestAuthServiceRejectsLoginCodeReplay(t *testing.T) {
	db := frontAuthTestDB(t)
	service := AuthService{
		DB: db,
		Exchanger: fakeWeChatExchanger{session: WeChatSession{
			OpenID: "openid", SessionKey: "session",
		}},
		IdentityKey: bytes.Repeat([]byte{0x22}, 32),
		Random:      bytes.NewReader(bytes.Repeat([]byte{0x33}, 1024)),
	}
	if _, err := service.Login(context.Background(), "one-time-code"); err != nil {
		t.Fatalf("initial login failed: %v", err)
	}
	if _, err := service.Login(context.Background(), "one-time-code"); appErrors.GetType(err) != appErrors.FrontWxLoginInvalid {
		t.Fatalf("replay error = %v, want FrontWxLoginInvalid", err)
	}
}

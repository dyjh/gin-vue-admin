package testutil

import (
	"strings"
	"testing"
)

// TestValidateMySQLTestDatabase 验证测试库名的安全边界。
func TestValidateMySQLTestDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string // 测试场景名称
		databaseName string // 待校验的数据库名
		wantError    bool   // 是否应返回错误
	}{
		{name: "valid", databaseName: "orderfood_test"},
		{name: "uppercase suffix", databaseName: "ORDERFOOD_TEST"},
		{name: "empty", wantError: true},
		{name: "development database", databaseName: "orderfood", wantError: true},
		{name: "unsafe characters", databaseName: "orderfood-test", wantError: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateMySQLTestDatabase(test.databaseName)
			if (err != nil) != test.wantError {
				t.Fatalf(
					"validateMySQLTestDatabase(%q) error = %v, wantError = %v",
					test.databaseName,
					err,
					test.wantError,
				)
			}
		})
	}
}

// TestBuildMySQLTestDatabaseName 验证隔离库名合法、唯一且不超过长度限制。
func TestBuildMySQLTestDatabaseName(t *testing.T) {
	t.Parallel()

	first := buildMySQLTestDatabaseName(
		"orderfood_test",
		strings.Repeat("Very Long Test Name/", 10),
	)
	second := buildMySQLTestDatabaseName("orderfood_test", "same test")
	if first == second {
		t.Fatal("测试数据库名不唯一")
	}
	for _, databaseName := range []string{first, second} {
		if len(databaseName) > maxMySQLDatabaseNameLength {
			t.Fatalf("数据库名长度 = %d，超过限制: %s", len(databaseName), databaseName)
		}
		if !mysqlDatabaseNamePattern.MatchString(databaseName) {
			t.Fatalf("数据库名包含非法字符: %s", databaseName)
		}
	}
}

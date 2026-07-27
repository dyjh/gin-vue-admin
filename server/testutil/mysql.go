// Package testutil 提供仅供自动化测试使用的基础设施工具。
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	driverMySQL "github.com/go-sql-driver/mysql"
	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	// MySQLDSNEnv 是订单模块数据库测试专用的 MySQL DSN 环境变量。
	MySQLDSNEnv = "ORDERFOOD_TEST_MYSQL_DSN"
	// MySQLReuseDatabaseEnv 控制数据库测试是否直接复用 DSN 中的专用空测试库。
	MySQLReuseDatabaseEnv = "ORDERFOOD_TEST_MYSQL_REUSE_DATABASE"

	maxMySQLDatabaseNameLength = 64
)

var (
	mysqlDatabaseNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	mysqlDatabaseCounter     atomic.Uint64
)

// OpenMySQL 为当前测试创建独立的 MySQL 数据库，并在测试结束后自动删除。
func OpenMySQL(t testing.TB) *gorm.DB {
	t.Helper()

	rawDSN := strings.TrimSpace(os.Getenv(MySQLDSNEnv))
	if rawDSN == "" {
		t.Fatalf(
			"%s 未配置；数据库测试必须显式使用名称以 _test 结尾的 MySQL 测试库",
			MySQLDSNEnv,
		)
	}
	baseConfig, err := driverMySQL.ParseDSN(rawDSN)
	if err != nil {
		t.Fatalf("%s 格式无效: %v", MySQLDSNEnv, err)
	}
	if err = validateMySQLTestDatabase(baseConfig.DBName); err != nil {
		t.Fatalf("%s 不安全: %v", MySQLDSNEnv, err)
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv(MySQLReuseDatabaseEnv)), "true") {
		return openReusableMySQL(t, baseConfig)
	}

	// 管理连接不选择具体数据库，只负责创建和回收本次测试的隔离数据库。
	adminConfig := *baseConfig
	adminConfig.DBName = ""
	adminDB, err := sql.Open("mysql", adminConfig.FormatDSN())
	if err != nil {
		t.Fatalf("创建 MySQL 测试管理连接失败: %v", err)
	}
	var testSQLDB *sql.DB
	databaseName := buildMySQLTestDatabaseName(baseConfig.DBName, t.Name())
	t.Cleanup(func() {
		if testSQLDB != nil {
			if closeErr := testSQLDB.Close(); closeErr != nil {
				t.Errorf("关闭 MySQL 测试连接失败: %v", closeErr)
			}
		}

		cleanupContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if _, dropErr := adminDB.ExecContext(
			cleanupContext,
			"DROP DATABASE IF EXISTS "+quoteMySQLIdentifier(databaseName),
		); dropErr != nil {
			t.Errorf("删除 MySQL 测试数据库 %s 失败: %v", databaseName, dropErr)
		}
		if closeErr := adminDB.Close(); closeErr != nil {
			t.Errorf("关闭 MySQL 测试管理连接失败: %v", closeErr)
		}
	})

	setupContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err = adminDB.PingContext(setupContext); err != nil {
		t.Fatalf("连接 MySQL 测试实例失败: %v", err)
	}
	if _, err = adminDB.ExecContext(
		setupContext,
		"CREATE DATABASE "+quoteMySQLIdentifier(databaseName)+
			" CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
	); err != nil {
		t.Fatalf("创建 MySQL 测试数据库 %s 失败: %v", databaseName, err)
	}

	// 统一开启时间解析并使用 UTC，避免不同开发环境的时区设置影响断言。
	testConfig := *baseConfig
	testConfig.DBName = databaseName
	testConfig.ParseTime = true
	testConfig.Loc = time.UTC
	db, err := gorm.Open(
		gormMySQL.Open(testConfig.FormatDSN()),
		&gorm.Config{
			// 与生产初始化保持一致，表依赖由迁移顺序管理，不让 GORM 自动生成外键。
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	)
	if err != nil {
		t.Fatalf("打开 MySQL 测试数据库 %s 失败: %v", databaseName, err)
	}
	testSQLDB, err = db.DB()
	if err != nil {
		t.Fatalf("获取 MySQL 测试连接池失败: %v", err)
	}
	if err = testSQLDB.PingContext(setupContext); err != nil {
		t.Fatalf("校验 MySQL 测试数据库连接失败: %v", err)
	}
	return db
}

// openReusableMySQL 复用显式指定的专用测试库，并在每条用例前后清空库内对象。
func openReusableMySQL(t testing.TB, baseConfig *driverMySQL.Config) *gorm.DB {
	t.Helper()

	testConfig := *baseConfig
	testConfig.ParseTime = true
	testConfig.Loc = time.UTC
	db, err := gorm.Open(
		gormMySQL.Open(testConfig.FormatDSN()),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
	)
	if err != nil {
		t.Fatalf("打开复用 MySQL 测试数据库 %s 失败: %v", testConfig.DBName, err)
	}
	testSQLDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取复用 MySQL 测试连接池失败: %v", err)
	}

	setupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = testSQLDB.PingContext(setupContext); err != nil {
		_ = testSQLDB.Close()
		t.Fatalf("连接复用 MySQL 测试数据库失败: %v", err)
	}
	lockConnection, err := testSQLDB.Conn(setupContext)
	if err != nil {
		_ = testSQLDB.Close()
		t.Fatalf("获取复用 MySQL 测试锁连接失败: %v", err)
	}
	lockName := "orderfood_reuse_" + testConfig.DBName
	var lockResult sql.NullInt64
	if err = lockConnection.QueryRowContext(
		setupContext,
		"SELECT GET_LOCK(?, 30)",
		lockName,
	).Scan(&lockResult); err != nil || !lockResult.Valid || lockResult.Int64 != 1 {
		_ = lockConnection.Close()
		_ = testSQLDB.Close()
		t.Fatalf("锁定复用 MySQL 测试数据库失败: %v", err)
	}
	if err = resetReusableMySQLDatabase(setupContext, lockConnection, testConfig.DBName); err != nil {
		_, _ = lockConnection.ExecContext(setupContext, "SELECT RELEASE_LOCK(?)", lockName)
		_ = lockConnection.Close()
		_ = testSQLDB.Close()
		t.Fatalf("清理复用 MySQL 测试数据库失败: %v", err)
	}

	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if cleanupErr := resetReusableMySQLDatabase(
			cleanupContext,
			lockConnection,
			testConfig.DBName,
		); cleanupErr != nil {
			t.Errorf("回收复用 MySQL 测试数据库失败: %v", cleanupErr)
		}
		if _, releaseErr := lockConnection.ExecContext(
			cleanupContext,
			"SELECT RELEASE_LOCK(?)",
			lockName,
		); releaseErr != nil {
			t.Errorf("释放复用 MySQL 测试数据库锁失败: %v", releaseErr)
		}
		if closeErr := lockConnection.Close(); closeErr != nil {
			t.Errorf("关闭复用 MySQL 测试锁连接失败: %v", closeErr)
		}
		if closeErr := testSQLDB.Close(); closeErr != nil {
			t.Errorf("关闭复用 MySQL 测试连接池失败: %v", closeErr)
		}
	})
	return db
}

// resetReusableMySQLDatabase 删除复用测试库中的视图和数据表，确保用例之间互不污染。
func resetReusableMySQLDatabase(
	ctx context.Context,
	connection *sql.Conn,
	databaseName string,
) error {
	rows, err := connection.QueryContext(
		ctx,
		`SELECT TABLE_NAME, TABLE_TYPE
		 FROM information_schema.TABLES
		 WHERE TABLE_SCHEMA = ?
		 ORDER BY CASE WHEN TABLE_TYPE = 'VIEW' THEN 0 ELSE 1 END, TABLE_NAME`,
		databaseName,
	)
	if err != nil {
		return fmt.Errorf("查询测试库对象失败: %w", err)
	}
	type databaseObject struct {
		name       string
		objectType string
	}
	objects := make([]databaseObject, 0)
	for rows.Next() {
		var object databaseObject
		if scanErr := rows.Scan(&object.name, &object.objectType); scanErr != nil {
			_ = rows.Close()
			return fmt.Errorf("读取测试库对象失败: %w", scanErr)
		}
		objects = append(objects, object)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("遍历测试库对象失败: %w", err)
	}
	if err = rows.Close(); err != nil {
		return fmt.Errorf("关闭测试库对象查询失败: %w", err)
	}

	if _, err = connection.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("关闭测试库外键检查失败: %w", err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(), "SET FOREIGN_KEY_CHECKS = 1")
	}()
	for _, object := range objects {
		statement := "DROP TABLE IF EXISTS "
		if object.objectType == "VIEW" {
			statement = "DROP VIEW IF EXISTS "
		}
		statement += quoteMySQLIdentifier(databaseName) + "." +
			quoteMySQLIdentifier(object.name)
		if _, err = connection.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("删除测试库对象 %s 失败: %w", object.name, err)
		}
	}
	return nil
}

// validateMySQLTestDatabase 校验基础库名，阻止测试误连开发库或生产库。
func validateMySQLTestDatabase(databaseName string) error {
	if databaseName == "" {
		return fmt.Errorf("DSN 必须包含数据库名")
	}
	if len(databaseName) > maxMySQLDatabaseNameLength {
		return fmt.Errorf("数据库名长度不能超过 %d", maxMySQLDatabaseNameLength)
	}
	if !mysqlDatabaseNamePattern.MatchString(databaseName) {
		return fmt.Errorf("数据库名只能包含字母、数字和下划线")
	}
	if !strings.HasSuffix(strings.ToLower(databaseName), "_test") {
		return fmt.Errorf("数据库名必须以 _test 结尾")
	}
	return nil
}

// buildMySQLTestDatabaseName 生成不超过 MySQL 长度限制的独立测试数据库名。
func buildMySQLTestDatabaseName(baseName, testName string) string {
	normalizedTestName := strings.ToLower(testName)
	normalizedTestName = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(
		normalizedTestName,
		"_",
	)
	normalizedTestName = strings.Trim(normalizedTestName, "_")
	if normalizedTestName == "" {
		normalizedTestName = "case"
	}

	uniqueSuffix := fmt.Sprintf(
		"%x_%d",
		time.Now().UnixNano(),
		mysqlDatabaseCounter.Add(1),
	)
	maxPrefixLength := maxMySQLDatabaseNameLength - len(uniqueSuffix) - 2
	prefix := baseName + "_" + normalizedTestName
	if len(prefix) > maxPrefixLength {
		prefix = strings.TrimRight(prefix[:maxPrefixLength], "_")
	}
	return prefix + "__" + uniqueSuffix
}

// quoteMySQLIdentifier 为已通过白名单校验的数据库标识符添加 MySQL 引号。
func quoteMySQLIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	envLoaded bool
	envOnce   sync.Once
)

type DBType string

const (
	DBTypeSQLite   DBType = "sqlite"
	DBTypeMySQL    DBType = "mysql"
	DBTypePostgres DBType = "postgres"
)

func LoadEnv() {
	envOnce.Do(func() {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("当前文件获取失败")
		}
		testUtilDir := filepath.Dir(filename)
		blogDir := filepath.Dir(testUtilDir)
		envfile := filepath.Join(blogDir, ".env")
		if err := godotenv.Load(envfile); err != nil {
			return
		}
		envLoaded = true
	})
}

func GetDBType() DBType {
	LoadEnv()
	dbType := os.Getenv("TEST_DB_TYPE")
	switch dbType {
	case "mysql":
		return DBTypeMySQL
	case "postgres":
		return DBTypePostgres
	case "sqlite":
		return DBTypeSQLite
	default:
		panic("未知的数据库类型: " + dbType)
	}
}

// GetDBDir 获取数据库目录
func GetDBDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("无法获取当前文件")
	}
	testUtilDir := filepath.Dir(filename)
	blogDir := filepath.Dir(testUtilDir)
	dbDir := filepath.Join(blogDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", err
	}
	return dbDir, nil
}

func buildSQLiteFilename(filename string) string {
	if filename == "" {
		return "test_sqlite.db"
	}
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".db"
	}
	base := filename[:len(filename)-len(filepath.Ext(filename))]
	if base == "" {
		base = "test"
	}
	baseLower := strings.ToLower(base)
	if !strings.Contains(baseLower, "sqlite") {
		return base + "_sqlite" + ext
	}
	return base + ext
}

func openSQLiteDB(dbPath string) (*gorm.DB, error) {
	return gorm.Open(
		sqlite.Open(dbPath),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
				TablePrefix:   "",
				NoLowerCase:   true,
			},
		},
	)
}

// OpenDB 根据 .env 配置打开数据库连接（非测试用）
func OpenDB(filename string) (*gorm.DB, error) {
	dbType := GetDBType()
	dbDir, err := GetDBDir()
	if err != nil {
		return nil, err
	}

	switch dbType {
	case DBTypeSQLite:
		dbPath := filepath.Join(dbDir, buildSQLiteFilename(filename))
		return openSQLiteDB(dbPath)
	default:
		panic("未知的数据库类型: " + string(dbType))
	}
}

// NewTestDB 创建测试用数据库连接
func NewTestDB(t *testing.T, filename string) *gorm.DB {
	t.Helper()
	dbType := GetDBType()
	var db *gorm.DB
	var err error

	switch dbType {
	case DBTypeSQLite:
		db, err = newSQLiteDB(filename)
	default:
		panic("未知的数据库类型: " + string(dbType))
	}
	if err != nil {
		t.Fatalf("创建数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetConnMaxLifetime(30 * time.Hour)
	t.Cleanup(func() {
		sqlDB.Close()
	})
	return db
}

func newSQLiteDB(filename string) (*gorm.DB, error) {
	dbDir, err := GetDBDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dbDir, buildSQLiteFilename(filename))
	return openSQLiteDB(dbPath)
}

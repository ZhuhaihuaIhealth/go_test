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
	DB_TYPE_MYSQL      DBType = "mysql"
	DB_TYPE_POSTGRESQL DBType = "postgresql"
	DB_TYPE_SQLITE     DBType = "sqlite"
)

func LoadEnv() {
	envOnce.Do(func() {

		//获取当前文件
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("无法获取当前文件")
		}
		//获取当前文件所在目录
		testUtilDir := filepath.Dir(filename)
		homeworkDir := filepath.Dir(testUtilDir)
		envfile := filepath.Join(homeworkDir, ".env")
		if err := godotenv.Load(envfile); err != nil {
			// panic("无法加载环境变量文件")
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
		return DB_TYPE_MYSQL
	case "postgresql":
		return DB_TYPE_POSTGRESQL
	case "sqlite":
		return DB_TYPE_SQLITE
	default:
		panic("未知的数据库类型: " + dbType)
	}
}

func getDBDir() (string, error) {
	//获取当前文件
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("无法获取当前文件")
	}
	testUtilDir := filepath.Dir(filename)
	homeworkDir := filepath.Dir(testUtilDir)
	dbDir := filepath.Join(homeworkDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", err
	}
	return dbDir, nil
}

func NewTestDB(t *testing.T, filename string) *gorm.DB {
	t.Helper()
	dbType := GetDBType()
	var db *gorm.DB
	var err error
	switch dbType {
	case DB_TYPE_SQLITE:
		db, err = newSQLiteDB(t, filename)
	default:
		panic("unsupported db type")
	}

	if err != nil {
		t.Fatalf("failed to connect database, got error: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB from gorm.DB")
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Hour)
	t.Cleanup(func() {
		sqlDB.Close()
	})
	return db

}
func newSQLiteDB(t *testing.T, filename string) (*gorm.DB,error) {
	dbDir, err := getDBDir()
	if err != nil {
		return nil, err
	}
	if filename == "" {
		filename = "test.sqlite.db"
	} else {
		ext := filepath.Ext(filename)
		if ext == "" {
			ext = ".db"
		}
		base := filename[:len(filename)-len(ext)]
		if base == "" {
			base = "test"
		}
		baseLower := strings.ToLower(base)
		if !strings.Contains(baseLower, "sqlite") {
			filename = base + "_sqlite" + ext
		} else {
			filename = base + ext
		}
	}
	dbPath := filepath.Join(dbDir, filename)
	return gorm.Open(
		sqlite.Open(dbPath),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
				TablePrefix:   "",
				NoLowerCase:   true,
				// NoReplacer:    nil,
			},
		},
	)
}

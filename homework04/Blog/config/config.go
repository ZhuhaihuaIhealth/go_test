package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
		configDir := filepath.Dir(filename)
		blogDir := filepath.Dir(configDir)
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

func GetDBDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("无法获取当前文件")
	}
	configDir := filepath.Dir(filename)
	blogDir := filepath.Dir(configDir)
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

func SetDBPool(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库连接失败: ", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
}

func GetTemplateDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("获取工作目录失败: ", err)
	}
	return filepath.Join(dir, "templates")
}

func GetJWTKey() []byte {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		log.Fatal("JWT_KEY 环境变量未设置")
	}
	return []byte(key)
}

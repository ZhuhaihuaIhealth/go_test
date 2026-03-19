package config

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"
)

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

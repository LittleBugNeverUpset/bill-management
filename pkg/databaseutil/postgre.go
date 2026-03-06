package databaseutil

import (
	"bill-management/pkg/config"
	"fmt"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

func InitPostgreSQL() *gorm.DB {
	// 1. 获取 PostgreSQL DSN
	dsn := config.GetPostgresDSN()

	// 2. 使用 GORM 连接 PostgreSQL 数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("连接 PostgreSQL 数据库失败: %w", err))
	}
	// 3. 返回数据库连接对象

	return db
}

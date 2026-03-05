package util_test

import (
	"bill-management/pkg/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func TestPostgresConnectionWithQuery() error {
	config.InitConfig("config")
	dsn := config.GetPostgresDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("ping数据库失败: %v", err)
	}

	// 测试简单查询
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return fmt.Errorf("查询数据库版本失败: %v", err)
	}

	// 测试当前数据库
	var currentDB string
	err = db.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		return fmt.Errorf("查询当前数据库失败: %v", err)
	}

	log.Printf("PostgreSQL 连接成功!")
	log.Printf("数据库版本: %s", version)
	log.Printf("当前数据库: %s", currentDB)

	return nil
}

package main

import (
	"bill-management/test/util_test"
	"log"
)

func main() {
	if err := util_test.TestPostgresConnectionWithQuery(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
}

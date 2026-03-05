package util_test

import (
	"bill-management/pkg/config"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// 修正：测试函数无返回值，符合 Go testing 规范
func TestPostgresConnectionWithQuery(t *testing.T) {
	// 1. 初始化配置（确保配置文件路径正确，如 config.yaml 在项目根目录）
	// 若配置文件在 test 目录下，需调整路径，如 "../config"
	config.InitConfig("config") // 假设 InitConfig 返回 error（若你的函数不返回，可直接调用）

	// 2. 获取 PostgreSQL DSN
	dsn := config.GetPostgresDSN()
	assert.NotEmpty(t, dsn, "PostgreSQL DSN 不能为空")

	// 3. 打开数据库连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err) // Fatalf 会终止测试，Errorf 仅标记错误继续
	}
	defer db.Close() // 测试结束关闭连接

	// 4. 测试数据库连通性（Ping）
	err = db.Ping()
	assert.NoError(t, err, "Ping 数据库失败")

	// 5. 测试查询数据库版本
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	assert.NoError(t, err, "查询数据库版本失败")
	assert.NotEmpty(t, version, "数据库版本不能为空")
	t.Logf("数据库版本: %s", version) // 打印版本信息，便于调试

	// 6. 测试查询当前数据库名称
	var currentDB string
	err = db.QueryRow("SELECT current_database()").Scan(&currentDB)
	assert.NoError(t, err, "查询当前数据库失败")
	assert.NotEmpty(t, currentDB, "当前数据库名称不能为空")
	t.Logf("当前数据库: %s", currentDB)

	// （可选）验证当前数据库是否符合预期（根据你的配置调整）
	expectedDB := config.GetConfig().Database.Psql.Database // 假设配置结构体有 Postgres.Database 字段
	assert.Equal(t, expectedDB, currentDB, "当前数据库名称与配置不符")
}

package util_test

import (
	"bill-management/pkg/config"
	"bill-management/pkg/logger"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMain 整个测试包的入口，全局初始化一次
func TestMain(m *testing.M) {
	// ========== 第一步：初始化配置和日志（仅执行一次） ==========
	// 注意：测试环境建议用单独的配置文件（如 config.test.yaml）
	config.InitConfig("config") // 加载测试配置
	logger.InitLogger()         // 初始化日志（依赖配置）

	// ========== 第二步：执行所有测试用例 ==========
	exitCode := m.Run() // 运行当前包的所有 TestXxx 函数

	// ========== 第三步：测试后清理（可选） ==========
	// 如关闭数据库连接、清理临时文件等

	// ========== 第四步：退出测试 ==========
	os.Exit(exitCode)
}
func TestConfig(t *testing.T) {
	config.InitConfig("config")
	cfg := config.GetConfig()
	assert.Equal(t, "localhost", cfg.Server.Host, "Server Host 应为 localhost")
	assert.Equal(t, 8080, cfg.Server.Port, "Server Port 应为 8080")
	assert.Equal(t, "debug", cfg.Server.Mode, "Server Mode 应为 debug")
	assert.Equal(t, "127.0.0.1", cfg.Database.Psql.Host, "Database Host 应为 127.0.0.1")
	assert.Equal(t, 5432, cfg.Database.Psql.Port, "Database Port 应为 5432")
	assert.Equal(t, "postgres", cfg.Database.Psql.Username, "Database Username 应为 postgres")
	assert.Equal(t, "postgres", cfg.Database.Psql.Database, "Database Name 应为 bill_management")
}

package util_test

import (
	"bill-management/pkg/config"
	"bill-management/pkg/logger"

	"go.uber.org/zap"
)

func logger_test() {
	config.InitConfig("config")
	logger.InitLogger()
	// 3. 演示全局使用日志
	logger.Info("程序启动", zap.String("app", config.GetConfig().Server.Host), zap.Int("port", config.GetConfig().Server.Port))
	logger.Infof("MySQL 配置：%s:%d", config.GetConfig().Database.Psql.Host, config.GetConfig().Database.Psql.Port)
}

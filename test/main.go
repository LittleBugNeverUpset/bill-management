package main

import (
	"bill-management/pkg/config"
	"bill-management/pkg/logger"
	"bill-management/pkg/redisutil"
	"context"
)

func main() {
	// 这里不需要写任何代码，测试框架会自动调用 TestMain 和 TestXxx 函数
	logger.Info("测试 Redis 连接")
	client := redisutil.NewRedisClient(&config.GetConfig().Database.Redis)
	logger.Infof("Redis 客户端配置: %+v", config.GetConfig().Database.Redis)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic("连接Redis失败: " + err.Error())
	}
	logger.Info("成功连接到 Redis")
}

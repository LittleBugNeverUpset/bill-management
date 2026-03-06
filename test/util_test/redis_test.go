package util_test

import (
	"bill-management/pkg/config"
	"bill-management/pkg/logger"
	"bill-management/pkg/redisutil"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisConnection(t *testing.T) {
	logger.Info("测试 Redis 连接")
	client := redisutil.NewRedisClient(&config.GetConfig().Database.Redis)
	logger.Infof("Redis 客户端配置: %+v", config.GetConfig().Database.Redis)
	assert.NotNil(t, client, "Redis 客户端实例为空")
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		assert.Fail(t, "连接 Redis 失败: "+err.Error())
		panic("连接Redis失败: " + err.Error())
	}
	logger.Info("成功连接到 Redis")

}

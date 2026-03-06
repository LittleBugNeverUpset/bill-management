package redisutil

import (
	"bill-management/pkg/config"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(config *config.RedisConfig) *redis.Client {
	// 这里可以根据 config 创建并返回一个 RedisClient 实例
	if config == nil {
		panic("Redis配置不能为空")
	}
	fmt.Printf("创建 Redis 客户端，配置: %+v\n", config)
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})
	return client
}
